// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

// NOTE: This is a legacy resource and should be migrated to the Plugin
// Framework if substantial modifications are planned. See
// docs/new-resources.md if planning to use this code as boilerplate for
// a new resource.

// go-tfe v2 migration exception: this data source intentionally remains on
// the go-tfe v1 client. It depends on `include=user` side-loaded user records
// (for the `username` attribute and name/email lookups), but the generated v2
// client cannot deserialize JSON:API `included` arrays: the composed-type
// factories (e.g. CreateOrganizationsFromDiscriminatorValue) do not
// discriminate on the JSON:API `type` field, so the first composed type
// always wins and GetUsers() is always nil. Migrate once the upstream
// generated client discriminates included records by type.

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
		Description: "Gets information about an organization membership. Requires using the provider with HCP Terraform or an instance of Terraform Enterprise at least as recent as v202004-1. Note that if a user updates their email address, configurations using the email address should be updated manually.",

		Read: dataSourceTFEOrganizationMembershipRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The organization membership ID.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"email": {
				Description: "Email of the user.",
				Type:        schema.TypeString,
				Optional:    true,
			},

			"username": {
				Description: "The username of the user. Although both are option, at least one of `email` and `username` is required.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},

			"organization": {
				Description: "Name of the organization.",
				Type:        schema.TypeString,
				Optional:    true,
			},

			"user_id": {
				Description: "The ID of the user associated with the organization membership.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"organization_membership_id": {
				Description:  "ID belonging to the organization membership.",
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
