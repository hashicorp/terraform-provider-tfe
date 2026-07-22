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

	tfe "github.com/hashicorp/go-tfe/v2"
	v2api "github.com/hashicorp/go-tfe/v2/api"
	"github.com/hashicorp/go-tfe/v2/api/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTFETeamOrganizationMembers() *schema.Resource {
	return &schema.Resource{
		Description: "Add or remove one or more team members using a [tfe_organization_membership](https://registry.terraform.io/providers/hashicorp/tfe/latest/docs/resources/organization_membership).\n\n" +
			"~> **Note:** Terraform provides four resources for managing team memberships. This resource, along with `tfe_team_organization_member`, is the preferred approach because memberships are managed via organization memberships. `tfe_team_organization_member` manages a single membership, while `tfe_team_organization_members` manages all memberships for a team. These four resources cannot be used for the same team simultaneously.\n\n" +
			"~> **Note:** This resource requires using the provider with HCP Terraform or Terraform Enterprise at least as recent as v202004-1.",

		Create: resourceTFETeamOrganizationMembersCreate,
		Read:   resourceTFETeamOrganizationMembersRead,
		Update: resourceTFETeamOrganizationMembersUpdate,
		Delete: resourceTFETeamOrganizationMembersDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The ID of this resource. Do not rely on this value — use `team_id` instead.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"team_id": {
				Description: "ID of the team.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},

			"organization_membership_ids": {
				Description: "IDs of organization memberships to be added.",
				Type:        schema.TypeSet,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceTFETeamOrganizationMembersCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	// Get the team ID.
	teamID := d.Get("team_id").(string)

	organizationMembershipIDs := schemaSetToStringSlice(d.Get("organization_membership_ids").(*schema.Set))

	log.Printf("[DEBUG] Add organization memberships %v to team: %s", organizationMembershipIDs, teamID)
	err := teamMembersAddOrgMembershipsV2(ctx, config.ClientV2.API, teamID, organizationMembershipIDs)
	if err != nil {
		return fmt.Errorf("Error adding organization memberships %v to team %s: %w", organizationMembershipIDs, teamID, err)
	}

	d.SetId(teamID)

	return nil
}

func resourceTFETeamOrganizationMembersRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Read organization memberships from team: %s", d.Id())
	organizationMemberships, err := teamMembersListOrgMembershipsV2(ctx, config.ClientV2.API, d.Id())
	if err != nil {
		if errors.Is(err, tfe.ErrNotFound) {
			log.Printf("[DEBUG] Organization memberships for team %s no longer exist", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading organization memberships from team %s: %w", d.Id(), err)
	}

	log.Printf("[DEBUG] Read users from team: %s", d.Id())
	nonServiceAccountOrganizationMemberships, err := filterNonServiceAccountOrganizationMembersV2(ctx, config.ClientV2.API, organizationMemberships)
	if err != nil {
		return fmt.Errorf("Error reading users from team %s: %w", d.Id(), err)
	}

	// Get all organization memberships and add them to object
	var organizationMembershipIDs []interface{}
	for _, membership := range nonServiceAccountOrganizationMemberships {
		organizationMembershipIDs = append(organizationMembershipIDs, valueOrZero(membership.GetId()))
	}

	// Check if organization memberships were added at all
	if len(organizationMembershipIDs) > 0 {
		d.Set("team_id", d.Id())
		d.Set("organization_membership_ids", organizationMembershipIDs)
	} else {
		log.Printf("[DEBUG] Organization memberships for team %s no longer exist", d.Id())
		d.SetId("")
	}

	return nil
}

// filterNonServiceAccountOrganizationMembersV2 returns the subset of
// organizationMemberships whose associated user is not a service account.
// This performs one additional read per membership to resolve its user, the
// same shape as go-tfe v1's per-membership
// OrganizationMemberships.ReadWithOptions(ctx, id, include: user) calls.
func filterNonServiceAccountOrganizationMembersV2(ctx context.Context, api *v2api.ApiClient, organizationMemberships []models.OrganizationMembershipsable) ([]models.OrganizationMembershipsable, error) {
	var nonServiceAccountMemberships []models.OrganizationMembershipsable

	for _, om := range organizationMemberships {
		membership, user, err := readOrganizationMembershipUserV2(ctx, api, valueOrZero(om.GetId()))
		if err != nil {
			return nil, fmt.Errorf("failed to fetch organization membership details for membership %s: %w", valueOrZero(om.GetId()), err)
		}

		if user == nil || !valueOrZero(user.GetAttributes().GetIsServiceAccount()) {
			nonServiceAccountMemberships = append(nonServiceAccountMemberships, membership)
		}
	}
	return nonServiceAccountMemberships, nil
}

func fetchExistingTeamMembershipIdsV2(ctx context.Context, api *v2api.ApiClient, teamID string) (map[string]interface{}, error) {
	teamMembers, err := teamMembersListOrgMembershipsV2(ctx, api, teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch existing organization memberships for team %s: %w", teamID, err)
	}

	nonServiceAccountOrganizationMemberships, err := filterNonServiceAccountOrganizationMembersV2(ctx, api, teamMembers)
	if err != nil {
		return nil, err
	}

	teamMembersIDSet := make(map[string]interface{})
	for _, m := range nonServiceAccountOrganizationMemberships {
		teamMembersIDSet[valueOrZero(m.GetId())] = nil
	}

	return teamMembersIDSet, nil
}

func resourceTFETeamOrganizationMembersUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	var membershipIDsToDelete *schema.Set
	var membershipIDsToAdd *schema.Set

	if d.HasChange("organization_membership_ids") {
		oldMembershipIDs, newMembershipIDs := d.GetChange("organization_membership_ids")
		membershipIDsToDelete = oldMembershipIDs.(*schema.Set).Difference(newMembershipIDs.(*schema.Set))
		membershipIDsToAdd = newMembershipIDs.(*schema.Set).Difference(oldMembershipIDs.(*schema.Set))
	}

	// First add the new organization memberships.
	if membershipIDsToAdd.Len() > 0 {
		idsToAdd := schemaSetToStringSlice(membershipIDsToAdd)
		log.Printf("[DEBUG] Add organization memberships %v to team: %s", idsToAdd, d.Id())
		err := teamMembersAddOrgMembershipsV2(ctx, config.ClientV2.API, d.Id(), idsToAdd)
		if err != nil {
			return fmt.Errorf("Error adding organization memberships to team %s: %w", d.Id(), err)
		}
	}

	if membershipIDsToDelete.Len() == 0 {
		return nil
	}

	// Then delete all the old users.
	existingIDs, err := fetchExistingTeamMembershipIdsV2(ctx, config.ClientV2.API, d.Id())
	if err != nil {
		return err
	}

	// Add all the organization memberships that need to be removed, except the
	// ones that already don't exist. Those may have been removed by another
	// destroy operation, such as a membership resource.
	var idsToDelete []string
	for _, id := range membershipIDsToDelete.List() {
		if _, exists := existingIDs[id.(string)]; exists {
			idsToDelete = append(idsToDelete, id.(string))
		}
	}

	if len(idsToDelete) > 0 {
		log.Printf("[DEBUG] Remove organization memberships %v from team: %s", idsToDelete, d.Id())
		err = teamMembersRemoveOrgMembershipsV2(ctx, config.ClientV2.API, d.Id(), idsToDelete)
		if err != nil {
			return fmt.Errorf("Error removing organization memberships from team %s: %w", d.Id(), err)
		}
	} else {
		log.Printf("[DEBUG] All members planned to be removed from this team were already removed from team %s", d.Id())
	}

	return nil
}

func resourceTFETeamOrganizationMembersDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Read organization memberships from team: %s", d.Id())
	organizationMemberships, err := teamMembersListOrgMembershipsV2(ctx, config.ClientV2.API, d.Id())
	if err != nil {
		if errors.Is(err, tfe.ErrNotFound) {
			log.Printf("[DEBUG] Organization memberships for team %s no longer exist", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading organization memberships from team %s: %w", d.Id(), err)
	}

	nonServiceAccountOrganizationMemberships, err := filterNonServiceAccountOrganizationMembersV2(ctx, config.ClientV2.API, organizationMemberships)
	if err != nil {
		return fmt.Errorf("Error fetching account user IDs for team %s: %w", d.Id(), err)
	}

	var idsToRemove []string
	for _, m := range nonServiceAccountOrganizationMemberships {
		idsToRemove = append(idsToRemove, valueOrZero(m.GetId()))
	}

	log.Printf("[DEBUG] Remove organization memberships %v from team: %s", idsToRemove, d.Id())
	err = teamMembersRemoveOrgMembershipsV2(ctx, config.ClientV2.API, d.Id(), idsToRemove)
	if err != nil {
		return fmt.Errorf("Error removing organization membership %v to team %s: %w", idsToRemove, d.Id(), err)
	}

	return nil
}
