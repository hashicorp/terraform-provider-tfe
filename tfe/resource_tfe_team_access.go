package tfe

import (
	"fmt"
	"log"
	"strings"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceTFETeamAccess() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFETeamAccessCreate,
		Read:   resourceTFETeamAccessRead,
		Delete: resourceTFETeamAccessDelete,
		Importer: &schema.ResourceImporter{
			State: resourceTFETeamAccessImporter,
		},

		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceTfeTeamAccessResourceV0().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceTfeTeamAccessStateUpgradeV0,
				Version: 0,
			},
		},

		Schema: map[string]*schema.Schema{
			"access": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice(
					[]string{
						string(tfe.AccessAdmin),
						string(tfe.AccessRead),
						string(tfe.AccessPlan),
						string(tfe.AccessWrite),
					},
					false,
				),
			},

			"team_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"workspace_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceTFETeamAccessCreate(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	// Get access and team ID.
	access := d.Get("access").(string)
	teamID := d.Get("team_id").(string)

	// Get the workspace
	workspaceID := d.Get("workspace_id").(string)
	ws, err := tfeClient.Workspaces.ReadByID(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf(
			"Error retrieving workspace %s: %v", workspaceID, err)
	}

	// Get the team.
	tm, err := tfeClient.Teams.Read(ctx, teamID)
	if err != nil {
		return fmt.Errorf("Error retrieving team %s: %v", teamID, err)
	}

	// Create a new options struct.
	options := tfe.TeamAccessAddOptions{
		Access:    tfe.Access(tfe.AccessType(access)),
		Team:      tm,
		Workspace: ws,
	}

	log.Printf("[DEBUG] Give team %s %s access to workspace: %s", tm.Name, access, ws.Name)
	tmAccess, err := tfeClient.TeamAccess.Add(ctx, options)
	if err != nil {
		return fmt.Errorf(
			"Error giving team %s %s access to workspace %s: %v", tm.Name, access, ws.Name, err)
	}

	d.SetId(tmAccess.ID)

	return resourceTFETeamAccessRead(d, meta)
}

func resourceTFETeamAccessRead(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	log.Printf("[DEBUG] Read configuration of team access: %s", d.Id())
	tmAccess, err := tfeClient.TeamAccess.Read(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			log.Printf("[DEBUG] Team access %s does no longer exist", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading configuration of team access %s: %v", d.Id(), err)
	}

	// Update config.
	d.Set("access", string(tmAccess.Access))

	if tmAccess.Team != nil {
		d.Set("team_id", tmAccess.Team.ID)
	} else {
		d.Set("team_id", "")
	}

	return nil
}

func resourceTFETeamAccessDelete(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	log.Printf("[DEBUG] Delete team access: %s", d.Id())
	err := tfeClient.TeamAccess.Remove(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return nil
		}
		return fmt.Errorf("Error deleting team access %s: %v", d.Id(), err)
	}

	return nil
}

//TODO: Gotta fix this too.
func resourceTFETeamAccessImporter(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	tfeClient := meta.(*tfe.Client)

	s := strings.SplitN(d.Id(), "/", 3)
	if len(s) != 3 {
		return nil, fmt.Errorf(
			"invalid team access import format: %s (expected <ORGANIZATION>/<WORKSPACE>/<TEAM ACCESS ID>)",
			d.Id(),
		)
	}

	// Set the fields that are part of the import ID.
	workspace_id, err := fetchWorkspaceExternalID(s[0]+"/"+s[1], tfeClient.Workspaces)
	if err != nil {
		return nil, fmt.Errorf(
			"error retrieving workspace %s from organization %s: %v", s[0], s[1], err)
	}
	d.Set("workspace_id", workspace_id)
	d.SetId(s[2])

	return []*schema.ResourceData{d}, nil
}
