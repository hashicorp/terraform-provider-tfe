// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"log"

	tfe "github.com/hashicorp/go-tfe"
	tfev2 "github.com/hashicorp/go-tfe/v2"
	v2api "github.com/hashicorp/go-tfe/v2/api"
	"github.com/hashicorp/go-tfe/v2/api/models"
	organizationmemberships "github.com/hashicorp/go-tfe/v2/api/organizationmemberships"
	memberitemparams "github.com/hashicorp/go-tfe/v2/api/organizationmemberships/item"
	"github.com/hashicorp/go-tfe/v2/api/organizations"
	orgmembershipparams "github.com/hashicorp/go-tfe/v2/api/organizations/item/organizationmemberships"
)

func fetchOrganizationMembers(client *tfe.Client, orgName string) ([]map[string]string, []map[string]string, error) {
	var members []map[string]string
	var membersWaiting []map[string]string

	options := tfe.OrganizationMembershipListOptions{}
	for {
		organizationMembershipList, err := client.OrganizationMemberships.List(ctx, orgName, &options)
		if err != nil {
			return nil, nil, fmt.Errorf("Error retrieving organization members: %w", err)
		}

		for _, orgMembership := range organizationMembershipList.Items {
			switch orgMembership.Status {
			case tfe.OrganizationMembershipActive:
				member := map[string]string{"user_id": orgMembership.User.ID, "organization_membership_id": orgMembership.ID}
				members = append(members, member)
			case tfe.OrganizationMembershipInvited:
				member := map[string]string{"user_id": orgMembership.User.ID, "organization_membership_id": orgMembership.ID}
				membersWaiting = append(membersWaiting, member)
			default:
				log.Printf("Organization member with unknown status found: %s", orgMembership.Status)
			}
		}
		// Exit the loop when we've seen all pages.
		if organizationMembershipList.CurrentPage >= organizationMembershipList.TotalPages {
			break
		}

		// Update the page number to get the next page.
		options.PageNumber = organizationMembershipList.NextPage
	}

	return members, membersWaiting, nil
}

func fetchOrganizationMemberByNameOrEmail(ctx context.Context, client *tfe.Client, organization, username, email string) (*tfe.OrganizationMembership, error) {
	if email == "" && username == "" {
		return nil, fmt.Errorf("you must specify a username or email")
	}

	options := &tfe.OrganizationMembershipListOptions{
		Include: []tfe.OrgMembershipIncludeOpt{tfe.OrgMembershipUser},
	}

	if email != "" {
		options.Emails = []string{email}
	}

	if username != "" {
		options.Query = username
	}

	oml, err := client.OrganizationMemberships.List(ctx, organization, options)
	if err != nil {
		return nil, fmt.Errorf("failed to list organization memberships: %w", err)
	}

	switch len(oml.Items) {
	case 0:
		return nil, tfe.ErrResourceNotFound
	case 1:
		user := oml.Items[0].User

		// We check this just in case a user's TFE instance only has one organization member
		if user.Email != email && user.Username != username {
			return nil, tfe.ErrResourceNotFound
		}

		return oml.Items[0], nil
	default:
		for {
			for _, member := range oml.Items {
				if (len(email) > 0 && member.User.Email == email) ||
					(len(username) > 0 && member.User.Username == username) {
					return member, nil
				}
			}

			if oml.CurrentPage >= oml.TotalPages {
				break
			}

			options.PageNumber = oml.NextPage

			oml, err = client.OrganizationMemberships.List(ctx, organization, options)
			if err != nil {
				return nil, fmt.Errorf("failed to list organization memberships: %w", err)
			}
		}
	}

	return nil, tfe.ErrResourceNotFound
}

// organizationMembershipUserV2 returns the full user record for an
// organization membership from a JSON:API `included` array populated by
// requesting include=user, or nil when it is not present.
func organizationMembershipUserV2[T includedUserGetter](membership models.OrganizationMembershipsable, included []T) models.Usersable {
	if membership == nil {
		return nil
	}
	relationships := membership.GetRelationships()
	if relationships == nil || relationships.GetUser() == nil || relationships.GetUser().GetData() == nil {
		return nil
	}
	return findIncludedUser(included, valueOrZero(relationships.GetUser().GetData().GetId()))
}

// fetchOrganizationMemberByNameOrEmailV2 finds an organization membership by
// exact email match or by username (via the "q" search query, verified
// locally for an exact match), following pagination until it is found or the
// pages are exhausted. It mirrors fetchOrganizationMemberByNameOrEmail's
// semantics using the go-tfe v2 generated client.
func fetchOrganizationMemberByNameOrEmailV2(ctx context.Context, api *v2api.ApiClient, organization, username, email string) (models.OrganizationMembershipsable, error) {
	if email == "" && username == "" {
		return nil, fmt.Errorf("you must specify a username or email")
	}

	membershipsBuilder := api.Organizations().ByOrganization_name(organization).OrganizationMemberships()

	include := orgmembershipparams.USER_GETINCLUDEQUERYPARAMETERTYPE
	queryParams := &organizations.ItemOrganizationMembershipsRequestBuilderGetQueryParameters{
		Include: &include,
	}
	if email != "" {
		queryParams.Filteremail = &email
	}
	if username != "" {
		queryParams.Q = &username
	}

	result, err := membershipsBuilder.Get(ctx, withQueryParams(queryParams))
	if err != nil {
		return nil, fmt.Errorf("failed to list organization memberships: %w", err)
	}

	items := result.GetData()
	switch len(items) {
	case 0:
		return nil, tfev2.ErrNotFound
	case 1:
		user := organizationMembershipUserV2(items[0], result.GetIncluded())

		// We check this just in case a user's TFE instance only has one organization member
		if user == nil || (valueOrZero(user.GetAttributes().GetEmail()) != email && valueOrZero(user.GetAttributes().GetUsername()) != username) {
			return nil, tfev2.ErrNotFound
		}

		return items[0], nil
	default:
		for {
			for _, member := range items {
				user := organizationMembershipUserV2(member, result.GetIncluded())
				if user == nil {
					continue
				}
				if (len(email) > 0 && valueOrZero(user.GetAttributes().GetEmail()) == email) ||
					(len(username) > 0 && valueOrZero(user.GetAttributes().GetUsername()) == username) {
					return member, nil
				}
			}

			nextPage := nextPageFromMeta(result.GetMeta())
			if nextPage == nil {
				break
			}

			pagedParams := &organizations.ItemOrganizationMembershipsRequestBuilderGetQueryParameters{
				Include:    &include,
				Pagenumber: nextPage,
			}
			if email != "" {
				pagedParams.Filteremail = &email
			}
			if username != "" {
				pagedParams.Q = &username
			}

			result, err = membershipsBuilder.Get(ctx, withQueryParams(pagedParams))
			if err != nil {
				return nil, fmt.Errorf("failed to list organization memberships: %w", err)
			}
			items = result.GetData()
		}
	}

	return nil, tfev2.ErrNotFound
}

// readOrganizationMembershipUserV2 reads a single organization membership by
// ID, including its associated user, using the go-tfe v2 generated client.
// This mirrors go-tfe v1's
// OrganizationMemberships.ReadWithOptions(ctx, id, OrganizationMembershipReadOptions{Include: []OrgMembershipIncludeOpt{OrgMembershipUser}}).
func readOrganizationMembershipUserV2(ctx context.Context, api *v2api.ApiClient, membershipID string) (models.OrganizationMembershipsable, models.Usersable, error) {
	include := memberitemparams.USER_GETINCLUDEQUERYPARAMETERTYPE
	queryParams := &organizationmemberships.WithOrganization_membership_ItemRequestBuilderGetQueryParameters{
		Include: &include,
	}

	result, err := api.OrganizationMemberships().ByOrganization_membership_id(membershipID).Get(ctx, withQueryParams(queryParams))
	if err != nil {
		if errors.Is(err, tfev2.ErrNotFound) {
			return nil, nil, tfev2.ErrNotFound
		}
		return nil, nil, err
	}
	if result == nil || result.GetData() == nil {
		return nil, nil, tfev2.ErrNotFound
	}

	membership := result.GetData()
	user := organizationMembershipUserV2(membership, result.GetIncluded())

	return membership, user, nil
}
