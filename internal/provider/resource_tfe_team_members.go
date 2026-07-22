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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTFETeamMembers() *schema.Resource {
	return &schema.Resource{
		Description: "Manages users in a team.",

		Create: resourceTFETeamMembersCreate,
		Read:   resourceTFETeamMembersRead,
		Update: resourceTFETeamMembersUpdate,
		Delete: resourceTFETeamMembersDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceTFETeamMembersImporter,
		},

		Schema: map[string]*schema.Schema{
			"team_id": {
				Description: "ID of the team.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},

			"usernames": {
				Description: "Names of the users to add.",
				Type:        schema.TypeSet,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceTFETeamMembersCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	// Get the team ID.
	teamID := d.Get("team_id").(string)

	// Collect all the usernames that need to be added.
	var usernames []string
	for _, username := range d.Get("usernames").(*schema.Set).List() {
		usernames = append(usernames, username.(string))
	}

	log.Printf("[DEBUG] Add users to team: %s", teamID)
	err := teamMembersAddUsersV2(ctx, config.ClientV2.API, teamID, usernames)
	if err != nil {
		return fmt.Errorf("Error adding users to team %s: %w", teamID, err)
	}

	d.SetId(teamID)

	return nil
}

func resourceTFETeamMembersRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Read users from team: %s", d.Id())
	users, err := teamMembersListUsersV2(ctx, config.ClientV2.API, d.Id())
	if err != nil {
		if errors.Is(err, tfe.ErrNotFound) {
			log.Printf("[DEBUG] Users no longer exist")
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading users from team %s: %w", d.Id(), err)
	}

	var usernames []interface{}
	for _, user := range users {
		usernames = append(usernames, valueOrZero(user.GetAttributes().GetUsername()))
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
			var usernames []string
			for _, username := range newUsers.List() {
				usernames = append(usernames, username.(string))
			}

			log.Printf("[DEBUG] Add users to team: %s", d.Id())
			err := teamMembersAddUsersV2(ctx, config.ClientV2.API, d.Id(), usernames)
			if err != nil {
				return fmt.Errorf("Error adding users to team %s: %w", d.Id(), err)
			}
		}

		// Then delete all the old users.
		if oldUsers.Len() > 0 {
			var usernames []string
			for _, username := range oldUsers.List() {
				usernames = append(usernames, username.(string))
			}

			log.Printf("[DEBUG] Remove users from team: %s", d.Id())
			err := teamMembersRemoveUsersV2(ctx, config.ClientV2.API, d.Id(), usernames)
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
	users, err := teamMembersListUsersV2(ctx, config.ClientV2.API, d.Id())
	if err != nil {
		if errors.Is(err, tfe.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("Error retrieving users to remove from team %s: %w", d.Id(), err)
	}

	var usernames []string
	for _, user := range users {
		usernames = append(usernames, valueOrZero(user.GetAttributes().GetUsername()))
	}

	log.Printf("[DEBUG] Remove users from team: %s", d.Id())
	err = teamMembersRemoveUsersV2(ctx, config.ClientV2.API, d.Id(), usernames)
	if err != nil {
		return fmt.Errorf("Error removing users from team %s: %w", d.Id(), err)
	}

	return nil
}

func resourceTFETeamMembersImporter(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// Set the team ID field.
	d.Set("team_id", d.Id())

	return []*schema.ResourceData{d}, nil
}
