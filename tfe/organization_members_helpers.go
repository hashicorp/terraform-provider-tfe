package tfe

import (
	"fmt"
	"log"

	tfe "github.com/hashicorp/go-tfe"
)

func fetchOrganizationMembers(client *tfe.Client, orgName string, options tfe.OrganizationMembershipListOptions) ([]map[string]string, []map[string]string, error) {
	var members []map[string]string
	var membersWaiting []map[string]string

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
