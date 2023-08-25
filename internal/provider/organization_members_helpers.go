// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"log"

	tfe "github.com/hashicorp/go-tfe"
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
			if orgMembership.Status == tfe.OrganizationMembershipActive {
				member := map[string]string{"user_id": orgMembership.User.ID, "organization_membership_id": orgMembership.ID}
				members = append(members, member)
			} else if orgMembership.Status == tfe.OrganizationMembershipInvited {
				member := map[string]string{"user_id": orgMembership.User.ID, "organization_membership_id": orgMembership.ID}
				membersWaiting = append(membersWaiting, member)
			} else {
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
