// Copyright IBM Corp. 2018, 2026
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
	"log"
	"strings"

	tfev2 "github.com/hashicorp/go-tfe/v2"
	"github.com/hashicorp/go-tfe/v2/api/models"
	"github.com/hashicorp/go-tfe/v2/api/organizationmemberships"
	membershipitem "github.com/hashicorp/go-tfe/v2/api/organizationmemberships/item"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-tfe/internal/provider/helpers"
)

func resourceTFEOrganizationMembership() *schema.Resource {
	return &schema.Resource{
		Description: "Adds or removes a user from an organization." +
			"\n\n~> **Note:** This resource requires using the provider with HCP Terraform or Terraform Enterprise at least as recent as v202004-1." +
			"\n\n~> **Note:** This resource cannot be used to update an existing user's email address since users themselves are the only ones permitted to update their email address. If a user updates their email address, configurations using the email address should be updated manually.",

		Create: resourceTFEOrganizationMembershipCreate,
		Read:   resourceTFEOrganizationMembershipRead,
		Delete: resourceTFEOrganizationMembershipDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceTFEOrganizationMembershipImporter,
		},

		CustomizeDiff: customizeDiffIfProviderDefaultOrganizationChanged,

		Identity: &schema.ResourceIdentity{
			SchemaFunc: func() map[string]*schema.Schema {
				return map[string]*schema.Schema{
					"id": {
						Type:              schema.TypeString,
						RequiredForImport: true,
					},
					"hostname": {
						Type:              schema.TypeString,
						OptionalForImport: true,
					},
				}
			},
		},

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The organization membership ID.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"email": {
				Description: "Email of the user to add.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},

			"organization": {
				Description: "Name of the organization. If omitted, organization must be defined in the provider config.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
			},

			"user_id": {
				Description: "The ID of the user associated with the organization membership.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"username": {
				Description: "The username of the user associated with the organization membership.",
				Type:        schema.TypeString,
				Computed:    true,
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

	membershipType := models.ORGANIZATIONMEMBERSHIPS_ORGANIZATIONMEMBERSHIPS_TYPE
	membership := models.NewOrganizationMemberships()
	membership.SetTypeEscaped(&membershipType)
	attrs := models.NewOrganizationMemberships_attributes()
	attrs.SetEmail(&email)
	membership.SetAttributes(attrs)
	body := models.NewOrganizationMembershipsEnvelope()
	body.SetData(membership)

	log.Printf("[DEBUG] Create membership %s for organization: %s", email, organization)
	env, err := config.ClientV2.API.Organizations().ByOrganization_name(organization).OrganizationMemberships().Post(ctx, body, nil)
	if err != nil {
		return fmt.Errorf(
			"Error creating membership %s for organization %s: %w", email, organization, err)
	}

	data := env.GetData()
	if data == nil {
		return fmt.Errorf(
			"Error creating membership %s for organization %s: API returned no data", email, organization)
	}

	membershipID := valueOrZero(data.GetId())
	d.SetId(membershipID)

	err = helpers.WriteTFEIdentity(d, membershipID, config.Client.BaseURL().Host)
	if err != nil {
		return err
	}

	return resourceTFEOrganizationMembershipRead(d, meta)
}

func resourceTFEOrganizationMembershipRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Read configuration of membership: %s", d.Id())
	env, err := config.ClientV2.API.OrganizationMemberships().ByOrganization_membership_id(d.Id()).Get(ctx, withQueryParams(&organizationmemberships.WithOrganization_membership_ItemRequestBuilderGetQueryParameters{
		Include: []membershipitem.GetIncludeQueryParameterType{membershipitem.USER_GETINCLUDEQUERYPARAMETERTYPE},
	}))
	if err != nil {
		if errors.Is(err, tfev2.ErrNotFound) {
			log.Printf("[DEBUG] Membership %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading configuration of membership %s: %w", d.Id(), err)
	}

	membership := env.GetData()
	if membership == nil {
		log.Printf("[DEBUG] Membership %s no longer exists", d.Id())
		d.SetId("")
		return nil
	}

	if attrs := membership.GetAttributes(); attrs != nil {
		d.Set("email", valueOrZero(attrs.GetEmail()))
	}

	if rel := membership.GetRelationships(); rel != nil {
		if org := rel.GetOrganization(); org != nil && org.GetData() != nil {
			d.Set("organization", valueOrZero(org.GetData().GetId()))
		}
	}

	userID := organizationMembershipUserID(membership)
	d.Set("user_id", userID)
	if user := findIncludedUser(env.GetIncluded(), userID); user != nil {
		d.Set("username", userUsername(user))
	}

	err = helpers.WriteTFEIdentity(d, d.Id(), config.Client.BaseURL().Host)
	if err != nil {
		return err
	}

	return nil
}

func resourceTFEOrganizationMembershipDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Delete membership: %s", d.Id())
	err := config.ClientV2.API.OrganizationMemberships().ByOrganization_membership_id(d.Id()).Delete(ctx, nil)
	if err != nil {
		if errors.Is(err, tfev2.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("Error deleting membership %s: %w", d.Id(), err)
	}

	return nil
}

func resourceTFEOrganizationMembershipImporter(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	config := meta.(ConfiguredClient)

	// First we'll check for an identity
	identity, err := d.Identity()
	if err != nil {
		return nil, fmt.Errorf("error reading organization membership identity: %w", err)
	}

	if externalID := identity.Get("id").(string); externalID != "" {
		// We are importing by identity
		// This only supported when using an import block, since import blocks
		// are the only way to specify an identity. Importing via TF CLI does
		// not support specifying an identity.
		d.SetId(externalID)

		// Exit early
		return []*schema.ResourceData{d}, nil
	}

	// Import formats:
	//  - <ORGANIZATION MEMBERSHIP ID>
	//  - <organization name>/<user email>
	s := strings.SplitN(d.Id(), "/", 2)
	if len(s) == 2 {
		org := s[0]
		email := s[1]
		orgMembership, err := fetchOrganizationMemberByNameOrEmailV2(ctx, config.ClientV2, org, "", email)
		if err != nil {
			return nil, fmt.Errorf(
				"error retrieving user with email %s from organization %s: %w", email, org, err)
		}

		d.SetId(valueOrZero(orgMembership.GetId()))
	}

	return []*schema.ResourceData{d}, nil
}
