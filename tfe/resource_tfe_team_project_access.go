// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tfe

import (
	"context"
	"errors"
	"log"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceTFETeamProjectAccess() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTFETeamProjectAccessCreate,
		ReadContext:   resourceTFETeamProjectAccessRead,
		UpdateContext: resourceTFETeamProjectAccessUpdate,
		DeleteContext: resourceTFETeamProjectAccessDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			"access": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice(
					[]string{
						string(tfe.TeamProjectAccessAdmin),
						string(tfe.TeamProjectAccessWrite),
						string(tfe.TeamProjectAccessMaintain),
						string(tfe.TeamProjectAccessRead),
					},
					false,
				),
			},

			"team_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"project_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringMatch(
					projectIDRegexp,
					"must be a valid project ID (prj-<RANDOM STRING>)",
				),
			},
		},
	}
}

func resourceTFETeamProjectAccessCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ConfiguredClient)

	// Get the access level
	access := d.Get("access").(string)

	// Get the project
	projectID := d.Get("project_id").(string)
	proj, err := config.Client.Projects.Read(ctx, projectID)
	if err != nil {
		return diag.Errorf(
			"Error retrieving project %s: %v", projectID, err)
	}

	// Get the team.
	teamID := d.Get("team_id").(string)
	tm, err := config.Client.Teams.Read(ctx, teamID)
	if err != nil {
		return diag.Errorf("Error retrieving team %s: %v", teamID, err)
	}

	// Create a new options struct.
	options := tfe.TeamProjectAccessAddOptions{
		Access:  *tfe.ProjectAccess(tfe.TeamProjectAccessType(access)),
		Team:    tm,
		Project: proj,
	}

	log.Printf("[DEBUG] Give team %s %s access to project: %s", tm.Name, access, proj.Name)
	tmAccess, err := config.Client.TeamProjectAccess.Add(ctx, options)
	if err != nil {
		return diag.Errorf(
			"Error giving team %s %s access to project %s: %v", tm.Name, access, proj.Name, err)
	}

	d.SetId(tmAccess.ID)

	return resourceTFETeamProjectAccessRead(ctx, d, meta)
}

func resourceTFETeamProjectAccessRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Read configuration of team access: %s", d.Id())
	tmAccess, err := config.Client.TeamProjectAccess.Read(ctx, d.Id())
	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
			log.Printf("[DEBUG] Team project access %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		return diag.Errorf("Error reading configuration of team project access %s: %v", d.Id(), err)
	}

	// Update config.
	d.Set("access", string(tmAccess.Access))

	if tmAccess.Team != nil {
		d.Set("team_id", tmAccess.Team.ID)
	} else {
		d.Set("team_id", "")
	}

	if tmAccess.Project != nil {
		d.Set("project_id", tmAccess.Project.ID)
	} else {
		d.Set("project_id", "")
	}

	return nil
}

func resourceTFETeamProjectAccessUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ConfiguredClient)

	// create an options struct
	options := tfe.TeamProjectAccessUpdateOptions{}

	// Set access level
	access := d.Get("access").(string)
	options.Access = tfe.ProjectAccess(tfe.TeamProjectAccessType(access))

	log.Printf("[DEBUG] Update team project access: %s", d.Id())
	_, err := config.Client.TeamProjectAccess.Update(ctx, d.Id(), options)
	if err != nil {
		return diag.Errorf(
			"Error updating team project access %s: %v", d.Id(), err)
	}

	return nil
}

func resourceTFETeamProjectAccessDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Delete team access: %s", d.Id())
	err := config.Client.TeamAccess.Remove(ctx, d.Id())
	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
			return nil
		}
		return diag.Errorf("Error deleting team project access %s: %v", d.Id(), err)
	}

	return nil
}
