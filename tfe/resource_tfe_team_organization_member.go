package tfe

import (
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
			StateContext: schema.ImportStatePassthroughContext,
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
	tfeClient := meta.(*tfe.Client)

	// Get the team ID and username..
	teamID := d.Get("team_id").(string)
	organizationMembershipID := d.Get("organization_membership_id").(string)

	// Create a new options struct.
	options := tfe.TeamMemberAddOptions{
		OrganizationMembershipIDs: []string{organizationMembershipID},
	}

	log.Printf("[DEBUG] Add organization membership %q to team: %s", organizationMembershipID, teamID)
	err := tfeClient.TeamMembers.Add(ctx, teamID, options)
	if err != nil {
		return fmt.Errorf("Error adding organization membership %q to team %s: %w", organizationMembershipID, teamID, err)
	}

	d.SetId(packTeamOrganizationMemberID(teamID, organizationMembershipID))

	return nil
}

func resourceTFETeamOrganizationMemberRead(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	// Get the team ID and organization membership id.
	teamID, organizationMembershipID, err := unpackTeamOrganizationMemberID(d.Id())
	if err != nil {
		return fmt.Errorf("Error unpacking team member ID: %w", err)
	}

	log.Printf("[DEBUG] Read organization membership from team: %s", teamID)
	organizationMemberships, err := tfeClient.TeamMembers.ListOrganizationMemberships(ctx, teamID)
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
	tfeClient := meta.(*tfe.Client)

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
	err = tfeClient.TeamMembers.Remove(ctx, teamID, options)
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
