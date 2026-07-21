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
	abstractions "github.com/microsoft/kiota-abstractions-go"
)

func fetchOrganizationMembers(client *tfev2.Client, orgName string) ([]map[string]string, []map[string]string, error) {
	var members []map[string]string
	var membersWaiting []map[string]string

	pageSize := int32(100)
	queryParams := &organizations.ItemOrganizationMembershipsRequestBuilderGetQueryParameters{
		Pagesize: &pageSize,
	}
	for {
		organizationMembershipList, err := client.API.Organizations().ByOrganization_name(orgName).OrganizationMemberships().Get(ctx, &abstractions.RequestConfiguration[organizations.ItemOrganizationMembershipsRequestBuilderGetQueryParameters]{
			QueryParameters: queryParams,
		})
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
		var nextPage *int32
		if meta := organizationMembershipList.GetMeta(); meta != nil {
			nextPage = nextPageNumber(meta.GetPagination())
		}
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
