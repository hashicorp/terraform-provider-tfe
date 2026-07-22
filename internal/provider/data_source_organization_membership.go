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

	tfev2 "github.com/hashicorp/go-tfe/v2"
	"github.com/hashicorp/go-tfe/v2/api/organizationmemberships"
	membershipitem "github.com/hashicorp/go-tfe/v2/api/organizationmemberships/item"
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
		orgMember, err := fetchOrganizationMemberByNameOrEmailV2(context.Background(), config.ClientV2, organization, username, email)
		if err != nil {
			return fmt.Errorf("could not find organization membership for organization %s: %w", organization, err)
		}

		orgMemberID = valueOrZero(orgMember.GetId())
	}

	d.SetId(orgMemberID)

	includeUser := membershipitem.USER_GETINCLUDEQUERYPARAMETERTYPE

	membershipResponse, err := config.ClientV2.API.OrganizationMemberships().ByOrganization_membership_id(orgMemberID).Get(context.Background(), withQueryParams(&organizationmemberships.WithOrganization_membership_ItemRequestBuilderGetQueryParameters{
		Include: &includeUser,
	}))
	if err != nil {
		if errors.Is(err, tfev2.ErrNotFound) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading configuration of membership %s: %w", orgMemberID, err)
	}

	membership := membershipResponse.GetData()
	if membership == nil {
		return fmt.Errorf("error reading configuration of membership %s: response contained no membership data", orgMemberID)
	}

	var membershipEmail string
	if attributes := membership.GetAttributes(); attributes != nil {
		membershipEmail = valueOrZero(attributes.GetEmail())
	}

	var organizationName string
	userID := organizationMembershipUserID(membership)
	if relationships := membership.GetRelationships(); relationships != nil {
		if org := relationships.GetOrganization(); org != nil && org.GetData() != nil {
			organizationName = valueOrZero(org.GetData().GetId())
		}
	}

	_, membershipUsername := userEmailAndUsername(findIncludedUser(membershipResponse.GetIncluded(), userID))

	d.Set("email", membershipEmail)
	d.Set("organization", organizationName)
	d.Set("user_id", userID)
	d.Set("username", membershipUsername)

	return nil
}
