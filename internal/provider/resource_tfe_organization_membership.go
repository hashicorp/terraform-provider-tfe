// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"log"
	"strings"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTFEOrganizationMembership() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFEOrganizationMembershipCreate,
		Read:   resourceTFEOrganizationMembershipRead,
		Delete: resourceTFEOrganizationMembershipDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceTFEOrganizationMembershipImporter,
		},

		Schema: map[string]*schema.Schema{
			"email": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"organization": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
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

func resourceTFEOrganizationMembershipCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	// Get the email and organization.
	email := d.Get("email").(string)
	organization, err := config.schemaOrDefaultOrganization(d)
	if err != nil {
		return err
	}

	// Create a new options struct.
	options := tfe.OrganizationMembershipCreateOptions{
		Email: tfe.String(email),
	}

	log.Printf("[DEBUG] Create membership %s for organization: %s", email, organization)
	membership, err := config.Client.OrganizationMemberships.Create(ctx, organization, options)
	if err != nil {
		return fmt.Errorf(
			"Error creating membership %s for organization %s: %w", email, organization, err)
	}

	d.SetId(membership.ID)

	return resourceTFEOrganizationMembershipRead(d, meta)
}

func resourceTFEOrganizationMembershipRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	options := tfe.OrganizationMembershipReadOptions{
		Include: []tfe.OrgMembershipIncludeOpt{tfe.OrgMembershipUser},
	}

	log.Printf("[DEBUG] Read configuration of membership: %s", d.Id())
	membership, err := config.Client.OrganizationMemberships.ReadWithOptions(ctx, d.Id(), options)

	if err != nil {
		if err == tfe.ErrResourceNotFound {
			log.Printf("[DEBUG] Membership %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading configuration of membership %s: %w", d.Id(), err)
	}

	d.Set("email", membership.Email)
	d.Set("organization", membership.Organization.Name)
	d.Set("user_id", membership.User.ID)
	d.Set("username", membership.User.Username)

	return nil
}

func resourceTFEOrganizationMembershipDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Delete membership: %s", d.Id())
	err := config.Client.OrganizationMemberships.Delete(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return nil
		}
		return fmt.Errorf("Error deleting membership %s: %w", d.Id(), err)
	}

	return nil
}

func resourceTFEOrganizationMembershipImporter(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	config := meta.(ConfiguredClient)

	// Import formats:
	//  - <ORGANIZATION MEMBERSHIP ID>
	//  - <organization name>/<user email>
	s := strings.SplitN(d.Id(), "/", 2)
	if len(s) == 2 {
		org := s[0]
		email := s[1]
		orgMembership, err := fetchOrganizationMemberByNameOrEmail(ctx, config.Client, org, "", email)
		if err != nil {
			return nil, fmt.Errorf(
				"error retrieving user with email %s from organization %s: %w", email, org, err)
		}

		d.SetId(orgMembership.ID)
	}

	return []*schema.ResourceData{d}, nil
}
