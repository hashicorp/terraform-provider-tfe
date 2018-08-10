package tfe

import (
	"fmt"
	"log"
	"strings"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceTFETeamMember() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFETeamMemberCreate,
		Read:   resourceTFETeamMemberRead,
		Delete: resourceTFETeamMemberDelete,

		Schema: map[string]*schema.Schema{
			"team_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"username": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceTFETeamMemberCreate(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	// Get the team ID and username..
	teamID := d.Get("team_id").(string)
	username := d.Get("username").(string)

	// Create a new options struct.
	options := tfe.TeamMemberAddOptions{
		Usernames: []string{username},
	}

	log.Printf("[DEBUG] Add user %q to team: %s", username, teamID)
	err := tfeClient.TeamMembers.Add(ctx, teamID, options)
	if err != nil {
		return fmt.Errorf("Error adding user %q to team %s: %v", username, teamID, err)
	}

	d.SetId(packTeamMemberID(teamID, username))

	return nil
}

func resourceTFETeamMemberRead(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	// Get the team ID and username..
	teamID, username := unpackTeamMemberID(d.Id())

	log.Printf("[DEBUG] Read users from team: %s", teamID)
	users, err := tfeClient.TeamMembers.List(ctx, teamID)
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			log.Printf("[DEBUG] User %q does no longer exist", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading users from team %s: %v", teamID, err)
	}

	found := false
	for _, user := range users {
		if user.Username == username {
			found = true
			break
		}
	}

	if !found {
		log.Printf("[DEBUG] User %q does no longer exist", d.Id())
		d.SetId("")
	}

	return nil
}

func resourceTFETeamMemberDelete(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	// Get the team ID and username..
	teamID, username := unpackTeamMemberID(d.Id())

	// Create a new options struct.
	options := tfe.TeamMemberRemoveOptions{
		Usernames: []string{username},
	}

	log.Printf("[DEBUG] Remove user %q from team: %s", username, teamID)
	err := tfeClient.TeamMembers.Remove(ctx, teamID, options)
	if err != nil {
		return fmt.Errorf("Error removing user %q to team %s: %v", username, teamID, err)
	}

	return nil
}

func packTeamMemberID(teamID, username string) string {
	return teamID + "|" + username
}

func unpackTeamMemberID(id string) (teamID, username string) {
	s := strings.SplitN(id, "|", 2)
	return s[0], s[1]
}
