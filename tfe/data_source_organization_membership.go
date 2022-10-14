package tfe

import (
	"context"
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
				Optional: true,
			},

			"username": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"organization": {
				Type:     schema.TypeString,
				Required: true,
			},

			"user_id": {
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
	username := d.Get("username").(string)
	organization := d.Get("organization").(string)

	orgMember, err := fetchOrganizationMemberByNameOrEmail(context.Background(), tfeClient, organization, username, email)
	if err != nil {
		return fmt.Errorf("Could not find organization membership for organization %s: %w", organization, err)
	}

	d.SetId(orgMember.ID)
	return resourceTFEOrganizationMembershipRead(d, meta)
}
