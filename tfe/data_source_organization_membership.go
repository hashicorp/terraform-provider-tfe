package tfe

import (
	"fmt"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTFEOrganizationMembership() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to get information about an organization membership." +
			"\n\n ~> **NOTE:** This data source requires using the provider with Terraform Cloud or an instance of Terraform Enterprise at least as recent as v202004-1." +
			"\n\n ~> **NOTE** If a user updates their email address, configurations using the email address should be updated manually.",
		Read: dataSourceTFEOrganizationMembershipRead,

		Schema: map[string]*schema.Schema{
			"email": {
				Description: "Email of the user.",
				Type:        schema.TypeString,
				Required:    true,
			},

			"organization": {
				Description: "Name of the organization.",
				Type:        schema.TypeString,
				Required:    true,
			},

			"user_id": {
				Description: "The ID of the user associated with the organization membership.",
				Type:        schema.TypeString,
				Computed:    true,
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
	options := &tfe.OrganizationMembershipListOptions{
		Include: []tfe.OrgMembershipIncludeOpt{tfe.OrgMembershipUser},
		Emails:  []string{email},
	}

	oml, err := tfeClient.OrganizationMemberships.List(ctx, organization, options)
	if err != nil {
		return fmt.Errorf("Error retrieving organization memberships: %w", err)
	}

	switch len(oml.Items) {
	case 0:
		return fmt.Errorf("Could not find organization membership for organization %s and email %s", organization, email)
	case 1:
		// We check this just in case a user's TFE instance only has one organization member
		// and doesn't support the filter query param
		if oml.Items[0].User.Email != email {
			return fmt.Errorf("Could not find organization membership for organization %s and email %s", organization, email)
		}

		d.SetId(oml.Items[0].ID)
		return resourceTFEOrganizationMembershipRead(d, meta)
	default:
		options = &tfe.OrganizationMembershipListOptions{
			Include: []tfe.OrgMembershipIncludeOpt{tfe.OrgMembershipUser},
		}

		for {
			for _, member := range oml.Items {
				if member.User.Email == email {
					d.SetId(member.ID)
					return resourceTFEOrganizationMembershipRead(d, meta)
				}
			}

			if oml.CurrentPage >= oml.TotalPages {
				break
			}

			options.PageNumber = oml.NextPage

			oml, err = tfeClient.OrganizationMemberships.List(ctx, organization, options)
			if err != nil {
				return fmt.Errorf("Error retrieving organization memberships: %w", err)
			}
		}
	}

	return fmt.Errorf("Could not find organization membership for organization %s and email %s", organization, email)
}
