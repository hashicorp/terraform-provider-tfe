// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tfe

import (
	"context"
	"fmt"

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
				Optional: true,
			},

			"user_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceTFEOrganizationMembershipRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	// Get the user email and organization.
	email := d.Get("email").(string)
	username := d.Get("username").(string)

	organization, err := config.schemaOrDefaultOrganization(d)
	if err != nil {
		return err
	}

	orgMember, err := fetchOrganizationMemberByNameOrEmail(context.Background(), config.Client, organization, username, email)
	if err != nil {
		return fmt.Errorf("could not find organization membership for organization %s: %w", organization, err)
	}

	d.SetId(orgMember.ID)
	return resourceTFEOrganizationMembershipRead(d, meta)
}
