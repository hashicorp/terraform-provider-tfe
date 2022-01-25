package tfe

import (
	"fmt"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTFEOrganizationMembership() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTFEOrganizationMembershipRead,

		Schema: map[string]*schema.Schema{
			"email": {
				Type:     schema.TypeString,
				Required: true,
			},

			"organization": {
				Type:     schema.TypeString,
				Required: true,
			},

			"user_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"username": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceTFEOrganizationMembershipRead(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	// Get the user email and organization.
	email := d.Get("email").(string)
	organization := d.Get("organization").(string)

	// Create an options struct.
	options := tfe.OrganizationMembershipListOptions{
		Include: "user",
	}

	for {
		organizationMembershipsList, err := tfeClient.OrganizationMemberships.List(ctx, organization, options)
		if err != nil {
			return fmt.Errorf("Error retrieving organization memberships: %v", err)
		}

		for _, organizationMembership := range organizationMembershipsList.Items {
			if organizationMembership.User.Email == email {
				d.SetId(organizationMembership.ID)
				return resourceTFEOrganizationMembershipRead(d, meta)
			}
		}

		// Exit the loop when we've seen all pages.
		if organizationMembershipsList.CurrentPage >= organizationMembershipsList.TotalPages {
			break
		}

		// Update the page number to get the next page.
		options.PageNumber = organizationMembershipsList.NextPage
	}

	return fmt.Errorf("Could not find organization membership for organization %s and email %s", organization, email)
}
