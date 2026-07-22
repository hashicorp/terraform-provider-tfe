// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"log"

	tfe "github.com/hashicorp/go-tfe"
	tfev2 "github.com/hashicorp/go-tfe/v2"
	"github.com/hashicorp/go-tfe/v2/api/models"
	"github.com/hashicorp/go-tfe/v2/api/organizations"
	orgmembershipsitem "github.com/hashicorp/go-tfe/v2/api/organizations/item/organizationmemberships"
)

func fetchOrganizationMembers(client *tfev2.Client, orgName string) ([]map[string]string, []map[string]string, error) {
	var members []map[string]string
	var membersWaiting []map[string]string

	pageSize := int32(100)
	queryParams := &organizations.ItemOrganizationMembershipsRequestBuilderGetQueryParameters{
		Pagesize: &pageSize,
	}
	for {
		organizationMembershipList, err := client.API.Organizations().ByOrganization_name(orgName).OrganizationMemberships().Get(ctx, withQueryParams(queryParams))
		if err != nil {
			return nil, nil, fmt.Errorf("Error retrieving organization members: %w", err)
		}

		for _, orgMembership := range organizationMembershipList.GetData() {
			member := map[string]string{
				"user_id":                    organizationMembershipUserID(orgMembership),
				"organization_membership_id": valueOrZero(orgMembership.GetId()),
			}

			var status *models.OrganizationMemberships_attributes_status
			if attributes := orgMembership.GetAttributes(); attributes != nil {
				status = attributes.GetStatus()
			}
			switch {
			case status != nil && *status == models.ACTIVE_ORGANIZATIONMEMBERSHIPS_ATTRIBUTES_STATUS:
				members = append(members, member)
			case status != nil && *status == models.INVITED_ORGANIZATIONMEMBERSHIPS_ATTRIBUTES_STATUS:
				membersWaiting = append(membersWaiting, member)
			default:
				statusString := ""
				if status != nil {
					statusString = status.String()
				}
				log.Printf("Organization member with unknown status found: %s", statusString)
			}
		}

		// Exit the loop when we've seen all pages.
		nextPage := nextPageFromMeta(organizationMembershipList.GetMeta())
		if nextPage == nil {
			break
		}

		// Update the page number to get the next page.
		queryParams.Pagenumber = nextPage
	}

	return members, membersWaiting, nil
}

// organizationMembershipUserID returns the ID of the user related to the
// given organization membership, or an empty string when the relationship is
// not present in the response.
func organizationMembershipUserID(membership models.OrganizationMembershipsable) string {
	relationships := membership.GetRelationships()
	if relationships == nil {
		return ""
	}
	user := relationships.GetUser()
	if user == nil || user.GetData() == nil {
		return ""
	}
	return valueOrZero(user.GetData().GetId())
}

// userEmailAndUsername returns the email and username attributes of a user
// record, or empty strings when they are not present.
func userEmailAndUsername(user models.Usersable) (string, string) {
	if user == nil {
		return "", ""
	}
	attributes := user.GetAttributes()
	if attributes == nil {
		return "", ""
	}
	return valueOrZero(attributes.GetEmail()), valueOrZero(attributes.GetUsername())
}

// fetchOrganizationMemberByNameOrEmailV2 is the go-tfe v2 counterpart of
// fetchOrganizationMemberByNameOrEmail. The v1 version remains until the
// resources that use it for imports are migrated.
func fetchOrganizationMemberByNameOrEmailV2(ctx context.Context, client *tfev2.Client, organization, username, email string) (models.OrganizationMembershipsable, error) {
	if email == "" && username == "" {
		return nil, fmt.Errorf("you must specify a username or email")
	}

	includeUser := orgmembershipsitem.USER_GETINCLUDEQUERYPARAMETERTYPE
	queryParams := &organizations.ItemOrganizationMembershipsRequestBuilderGetQueryParameters{
		Include: &includeUser,
	}

	if email != "" {
		queryParams.Filteremail = &email
	}

	if username != "" {
		queryParams.Q = &username
	}

	membershipsBuilder := client.API.Organizations().ByOrganization_name(organization).OrganizationMemberships()

	oml, err := membershipsBuilder.Get(ctx, withQueryParams(queryParams))
	if err != nil {
		return nil, fmt.Errorf("failed to list organization memberships: %w", err)
	}

	items := oml.GetData()
	switch len(items) {
	case 0:
		return nil, tfev2.ErrNotFound
	case 1:
		userEmail, userName := userEmailAndUsername(findIncludedUser(oml.GetIncluded(), organizationMembershipUserID(items[0])))

		// We check this just in case a user's TFE instance only has one organization member
		if userEmail != email && userName != username {
			return nil, tfev2.ErrNotFound
		}

		return items[0], nil
	default:
		for {
			for _, member := range items {
				userEmail, userName := userEmailAndUsername(findIncludedUser(oml.GetIncluded(), organizationMembershipUserID(member)))
				if (len(email) > 0 && userEmail == email) ||
					(len(username) > 0 && userName == username) {
					return member, nil
				}
			}

			nextPage := nextPageFromMeta(oml.GetMeta())
			if nextPage == nil {
				break
			}

			queryParams.Pagenumber = nextPage

			oml, err = membershipsBuilder.Get(ctx, withQueryParams(queryParams))
			if err != nil {
				return nil, fmt.Errorf("failed to list organization memberships: %w", err)
			}
			items = oml.GetData()
		}
	}

	return nil, tfev2.ErrNotFound
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
