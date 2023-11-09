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

func resourceTFETeamOrganizationMember() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFETeamOrganizationMemberCreate,
		Read:   resourceTFETeamOrganizationMemberRead,
		Delete: resourceTFETeamOrganizationMemberDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceTFETeamOrganizationMemberImporter,
		},

		Schema: map[string]*schema.Schema{
			"team_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"organization_membership_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceTFETeamOrganizationMemberCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	// Get the team ID and username..
	teamID := d.Get("team_id").(string)
	organizationMembershipID := d.Get("organization_membership_id").(string)

	// Create a new options struct.
	options := tfe.TeamMemberAddOptions{
		OrganizationMembershipIDs: []string{organizationMembershipID},
	}

	log.Printf("[DEBUG] Add organization membership %q to team: %s", organizationMembershipID, teamID)
	err := config.Client.TeamMembers.Add(ctx, teamID, options)
	if err != nil {
		return fmt.Errorf("Error adding organization membership %q to team %s: %w", organizationMembershipID, teamID, err)
	}

	d.SetId(packTeamOrganizationMemberID(teamID, organizationMembershipID))

	return nil
}

func resourceTFETeamOrganizationMemberRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	// Get the team ID and organization membership id.
	teamID, organizationMembershipID, err := unpackTeamOrganizationMemberID(d.Id())
	if err != nil {
		return fmt.Errorf("Error unpacking team member ID: %w", err)
	}

	log.Printf("[DEBUG] Read organization membership from team: %s", teamID)
	organizationMemberships, err := config.Client.TeamMembers.ListOrganizationMemberships(ctx, teamID)
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			log.Printf("[DEBUG] Organization membership %q no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading organization memberships from team %s: %w", teamID, err)
	}

	found := false
	for _, organizationMembership := range organizationMemberships {
		if organizationMembership.ID == organizationMembershipID {
			d.Set("team_id", teamID)
			d.Set("organization_membership_id", organizationMembershipID)

			found = true
			break
		}
	}

	if !found {
		log.Printf("[DEBUG] Organization membership %q no longer exists", d.Id())
		d.SetId("")
	}

	return nil
}

func resourceTFETeamOrganizationMemberDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	// Get the team ID and organization membership id.
	teamID, organizationMembershipID, err := unpackTeamOrganizationMemberID(d.Id())
	if err != nil {
		return fmt.Errorf("Error unpacking team member ID: %w", err)
	}

	// Create a new options struct.
	options := tfe.TeamMemberRemoveOptions{
		OrganizationMembershipIDs: []string{organizationMembershipID},
	}

	log.Printf("[DEBUG] Remove organization membership %q from team: %s", organizationMembershipID, teamID)
	err = config.Client.TeamMembers.Remove(ctx, teamID, options)
	if err != nil {
		return fmt.Errorf("Error removing organization membership %q to team %s: %w", organizationMembershipID, teamID, err)
	}

	return nil
}

func packTeamOrganizationMemberID(teamID, organizationMembershipID string) string {
	return teamID + "/" + organizationMembershipID
}

func unpackTeamOrganizationMemberID(id string) (teamID, organizationMembershipID string, err error) {
	s := strings.SplitN(id, "/", 2)
	if len(s) != 2 {
		return "", "", fmt.Errorf(
			"invalid team organization member ID format: %s (expected <TEAM ID>/<ORGANIZATION MEMBERSHIP ID>)", id)
	}

	return s[0], s[1], nil
}

func resourceTFETeamOrganizationMemberImporter(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	config := meta.(ConfiguredClient)

	// Import formats:
	//  - <TEAM ID>/<ORGANIZATION MEMBERSHIP ID>
	//  - <ORGANIZATION NAME>/<USER EMAIL>/<TEAM NAME>
	s := strings.SplitN(d.Id(), "/", 3)

	if len(s) == 2 {
		// the <TEAM ID>/<ORGANIZATION MEMBERSHIP ID> is the default ID, so pass it on through
		return []*schema.ResourceData{d}, nil
	} else if len(s) == 3 {
		// the ID we want to construct is <TEAM ID>/<ORGANIZATION MEMBERSHIP ID>
		// we can use org and email to get the org membership ID, and find the team based on org and team name
		org := s[0]
		email := s[1]
		teamName := s[2]
		orgMembership, err := fetchOrganizationMemberByNameOrEmail(ctx, config.Client, org, "", email)
		if err != nil {
			return nil, fmt.Errorf(
				"error retrieving user with email %s from organization %s: %w", email, org, err)
		}
		team, err := fetchTeamByName(ctx, config.Client, org, teamName)
		if err != nil {
			return nil, fmt.Errorf(
				"error retrieving team with name %s from organization %s: %w", teamName, org, err)
		}

		d.SetId(fmt.Sprintf("%s/%s", team.ID, orgMembership.ID))
		return []*schema.ResourceData{d}, nil
	}
	return nil, fmt.Errorf(
		"invalid organization membership input format: %s (expected <TEAM ID>/<ORGANIZATION MEMBERSHIP ID> or <ORGANIZATION NAME>/<TEAM NAME>/<USER EMAIL>)",
		d.Id(),
	)
}
