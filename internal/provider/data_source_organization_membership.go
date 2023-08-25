// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

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

			"organization_membership_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: []string{"email", "username"},
			},
		},
	}
}

func dataSourceTFEOrganizationMembershipRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	// Get the user email and organization.
	email := d.Get("email").(string)
	username := d.Get("username").(string)
	orgMemberID := d.Get("organization_membership_id").(string)

	organization, err := config.schemaOrDefaultOrganization(d)
	if err != nil {
		return err
	}

	if orgMemberID == "" {
		orgMember, err := fetchOrganizationMemberByNameOrEmail(context.Background(), config.Client, organization, username, email)
		if err != nil {
			return fmt.Errorf("could not find organization membership for organization %s: %w", organization, err)
		}

		d.SetId(orgMember.ID)
	} else {
		d.SetId(orgMemberID)
	}

	return resourceTFEOrganizationMembershipRead(d, meta)
}
