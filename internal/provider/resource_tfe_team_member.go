// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"log"
	"strings"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTFETeamMember() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFETeamMemberCreate,
		Read:   resourceTFETeamMemberRead,
		Delete: resourceTFETeamMemberDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"team_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"username": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceTFETeamMemberCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	// Get the team ID and username..
	teamID := d.Get("team_id").(string)
	username := d.Get("username").(string)

	// Create a new options struct.
	options := tfe.TeamMemberAddOptions{
		Usernames: []string{username},
	}

	log.Printf("[DEBUG] Add user %q to team: %s", username, teamID)
	err := config.Client.TeamMembers.Add(ctx, teamID, options)
	if err != nil {
		return fmt.Errorf("Error adding user %q to team %s: %w", username, teamID, err)
	}

	d.SetId(packTeamMemberID(teamID, username))

	return nil
}

func resourceTFETeamMemberRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	// Get the team ID and username.
	teamID, username, err := unpackTeamMemberID(d.Id())
	if err != nil {
		return fmt.Errorf("Error unpacking team member ID: %w", err)
	}

	log.Printf("[DEBUG] Read users from team: %s", teamID)
	users, err := config.Client.TeamMembers.List(ctx, teamID)
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			log.Printf("[DEBUG] User %q no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading users from team %s: %w", teamID, err)
	}

	found := false
	for _, user := range users {
		if user.Username == username {
			d.Set("team_id", teamID)
			d.Set("username", username)

			// We do this here as a means to convert the internal ID,
			// in case anyone still uses the old format.
			d.SetId(packTeamMemberID(teamID, username))
			found = true
			break
		}
	}

	if !found {
		log.Printf("[DEBUG] User %q no longer exists", d.Id())
		d.SetId("")
	}

	return nil
}

func resourceTFETeamMemberDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	// Get the team ID and username.
	teamID, username, err := unpackTeamMemberID(d.Id())
	if err != nil {
		return fmt.Errorf("Error unpacking team member ID: %w", err)
	}

	// Create a new options struct.
	options := tfe.TeamMemberRemoveOptions{
		Usernames: []string{username},
	}

	log.Printf("[DEBUG] Remove user %q from team: %s", username, teamID)
	err = config.Client.TeamMembers.Remove(ctx, teamID, options)
	if err != nil {
		return fmt.Errorf("Error removing user %q to team %s: %w", username, teamID, err)
	}

	return nil
}

func packTeamMemberID(teamID, username string) string {
	return teamID + "/" + username
}

func unpackTeamMemberID(id string) (teamID, username string, err error) {
	// Support the old ID format for backwards compatibitily.
	if s := strings.SplitN(id, "|", 2); len(s) == 2 {
		return s[0], s[1], nil
	}

	s := strings.SplitN(id, "/", 2)
	if len(s) != 2 {
		return "", "", fmt.Errorf(
			"invalid team member ID format: %s (expected <TEAM ID>/<USERNAME>)", id)
	}

	return s[0], s[1], nil
}
