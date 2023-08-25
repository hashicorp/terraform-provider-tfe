// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"log"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTFETeamMembers() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFETeamMembersCreate,
		Read:   resourceTFETeamMembersRead,
		Update: resourceTFETeamMembersUpdate,
		Delete: resourceTFETeamMembersDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceTFETeamMembersImporter,
		},

		Schema: map[string]*schema.Schema{
			"team_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"usernames": {
				Type:     schema.TypeSet,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceTFETeamMembersCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	// Get the team ID.
	teamID := d.Get("team_id").(string)

	// Create a new options struct.
	options := tfe.TeamMemberAddOptions{}

	// Add all the users that need to be added.
	for _, username := range d.Get("usernames").(*schema.Set).List() {
		options.Usernames = append(options.Usernames, username.(string))
	}

	log.Printf("[DEBUG] Add users to team: %s", teamID)
	err := config.Client.TeamMembers.Add(ctx, teamID, options)
	if err != nil {
		return fmt.Errorf("Error adding users to team %s: %w", teamID, err)
	}

	d.SetId(teamID)

	return nil
}

func resourceTFETeamMembersRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Read users from team: %s", d.Id())
	users, err := config.Client.TeamMembers.List(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			log.Printf("[DEBUG] Users no longer exist")
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading users from team %s: %w", d.Id(), err)
	}

	var usernames []interface{}
	for _, user := range users {
		usernames = append(usernames, user.Username)
	}

	if len(usernames) > 0 {
		d.Set("usernames", usernames)
	} else {
		log.Printf("[DEBUG] Users no longer exist")
		d.SetId("")
	}

	return nil
}

func resourceTFETeamMembersUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	if d.HasChange("usernames") {
		oldUsernames, newUsernames := d.GetChange("usernames")
		oldUsers := oldUsernames.(*schema.Set).Difference(newUsernames.(*schema.Set))
		newUsers := newUsernames.(*schema.Set).Difference(oldUsernames.(*schema.Set))

		// First add the new users.
		if newUsers.Len() > 0 {
			// Create a new options struct.
			options := tfe.TeamMemberAddOptions{}

			// Add all the users that need to be added.
			for _, username := range newUsers.List() {
				options.Usernames = append(options.Usernames, username.(string))
			}

			log.Printf("[DEBUG] Add users to team: %s", d.Id())
			err := config.Client.TeamMembers.Add(ctx, d.Id(), options)
			if err != nil {
				return fmt.Errorf("Error adding users to team %s: %w", d.Id(), err)
			}
		}

		// Then delete all the old users.
		if oldUsers.Len() > 0 {
			// Create a new options struct.
			options := tfe.TeamMemberRemoveOptions{}

			// Add all the users that need to be added.
			for _, username := range oldUsers.List() {
				options.Usernames = append(options.Usernames, username.(string))
			}

			log.Printf("[DEBUG] Remove users from team: %s", d.Id())
			err := config.Client.TeamMembers.Remove(ctx, d.Id(), options)
			if err != nil {
				return fmt.Errorf("Error removing users to team %s: %w", d.Id(), err)
			}
		}
	}

	return nil
}

func resourceTFETeamMembersDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Retrieve users to remove from team: %s", d.Id())
	users, err := config.Client.TeamMembers.List(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return nil
		}
		return fmt.Errorf("Error retrieving users to remove from team %s: %w", d.Id(), err)
	}

	// Create a new options struct.
	options := tfe.TeamMemberRemoveOptions{}

	// Add all the users that need to be removed.
	for _, user := range users {
		options.Usernames = append(options.Usernames, user.Username)
	}

	log.Printf("[DEBUG] Remove users from team: %s", d.Id())
	err = config.Client.TeamMembers.Remove(ctx, d.Id(), options)
	if err != nil {
		return fmt.Errorf("Error removing users to team %s: %w", d.Id(), err)
	}

	return nil
}

func resourceTFETeamMembersImporter(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// Set the team ID field.
	d.Set("team_id", d.Id())

	return []*schema.ResourceData{d}, nil
}
