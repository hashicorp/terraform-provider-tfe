// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

// NOTE: This is a legacy resource and should be migrated to the Plugin
// Framework if substantial modifications are planned. See
// docs/new-resources.md if planning to use this code as boilerplate for
// a new resource.

package provider

import (
	"context"
	"errors"
	"fmt"

	tfe "github.com/hashicorp/go-tfe"
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

		orgMemberID = orgMember.ID
	}

	d.SetId(orgMemberID)

	options := tfe.OrganizationMembershipReadOptions{
		Include: []tfe.OrgMembershipIncludeOpt{tfe.OrgMembershipUser},
	}

	membership, err := config.Client.OrganizationMemberships.ReadWithOptions(context.Background(), orgMemberID, options)
	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading configuration of membership %s: %w", orgMemberID, err)
	}

	d.Set("email", membership.Email)
	d.Set("organization", membership.Organization.Name)
	d.Set("user_id", membership.User.ID)
	d.Set("username", membership.User.Username)

	return nil
}
