// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTFETeamToken() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFETeamTokenCreate,
		Read:   resourceTFETeamTokenRead,
		Delete: resourceTFETeamTokenDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceTFETeamTokenImporter,
		},

		Schema: map[string]*schema.Schema{
			"team_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"force_regenerate": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},

			"token": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},

			"expired_at": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceTFETeamTokenCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	// Get the team ID.
	teamID := d.Get("team_id").(string)

	log.Printf("[DEBUG] Check if a token already exists for team: %s", teamID)
	_, err := config.Client.TeamTokens.Read(ctx, teamID)
	if err != nil && !errors.Is(err, tfe.ErrResourceNotFound) {
		return fmt.Errorf("error checking if a token exists for team %s: %w", teamID, err)
	}

	// If error is nil, the token already exists.
	if err == nil {
		if !d.Get("force_regenerate").(bool) {
			return fmt.Errorf("a token already exists for team: %s", teamID)
		}
		log.Printf("[DEBUG] Regenerating existing token for team: %s", teamID)
	}

	// Get the token create options.
	options := tfe.TeamTokenCreateOptions{}

	// Check whether the optional expiry was provided.
	expiredAt, expiredAtProvided := d.GetOk("expired_at")

	// If an expiry was provided, parse it and update the options struct.
	if expiredAtProvided {
		expiry, err := time.Parse(time.RFC3339, expiredAt.(string))

		options.ExpiredAt = &expiry

		if err != nil {
			return fmt.Errorf("%s must be a valid date or time, provided in iso8601 format", expiredAt)
		}
	}

	log.Printf("[DEBUG] Create new token for team: %s", teamID)
	token, err := config.Client.TeamTokens.CreateWithOptions(ctx, teamID, options)
	if err != nil {
		return fmt.Errorf(
			"error creating new token for team %s: %w", teamID, err)
	}

	d.SetId(teamID)

	// We need to set this here in the create function as this value will
	// only be returned once during the creation of the token.
	d.Set("token", token.Token)

	return resourceTFETeamTokenRead(d, meta)
}

func resourceTFETeamTokenRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Read the token from team: %s", d.Id())
	_, err := config.Client.TeamTokens.Read(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			log.Printf("[DEBUG] Token for team %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading token from team %s: %w", d.Id(), err)
	}

	return nil
}

func resourceTFETeamTokenDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Delete token from team: %s", d.Id())
	err := config.Client.TeamTokens.Delete(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return nil
		}
		return fmt.Errorf("error deleting token from team %s: %w", d.Id(), err)
	}

	return nil
}

func resourceTFETeamTokenImporter(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// Set the team ID field.
	d.Set("team_id", d.Id())

	return []*schema.ResourceData{d}, nil
}
