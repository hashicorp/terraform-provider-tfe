// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
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

		CustomizeDiff: checkForCustomPermissions,
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
						string(tfe.TeamProjectAccessCustom),
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

			"project_access": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"settings": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ValidateFunc: validation.StringInSlice(
								[]string{
									string(tfe.ProjectSettingsPermissionRead),
									string(tfe.ProjectSettingsPermissionUpdate),
									string(tfe.ProjectSettingsPermissionDelete),
								},
								false,
							),
						},

						"teams": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ValidateFunc: validation.StringInSlice(
								[]string{
									string(tfe.ProjectTeamsPermissionNone),
									string(tfe.ProjectTeamsPermissionRead),
									string(tfe.ProjectTeamsPermissionManage),
								},
								false,
							),
						},
					},
				},
			},

			"workspace_access": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"create": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},

						"locking": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},

						"move": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},

						"delete": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},

						"run_tasks": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},

						"runs": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ValidateFunc: validation.StringInSlice(
								[]string{
									string(tfe.WorkspaceRunsPermissionRead),
									string(tfe.WorkspaceRunsPermissionPlan),
									string(tfe.WorkspaceRunsPermissionApply),
								},
								false,
							),
						},

						"sentinel_mocks": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ValidateFunc: validation.StringInSlice(
								[]string{
									string(tfe.WorkspaceSentinelMocksPermissionNone),
									string(tfe.WorkspaceSentinelMocksPermissionRead),
								},
								false,
							),
						},

						"state_versions": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ValidateFunc: validation.StringInSlice(
								[]string{
									string(tfe.WorkspaceStateVersionsPermissionNone),
									string(tfe.WorkspaceStateVersionsPermissionReadOutputs),
									string(tfe.WorkspaceStateVersionsPermissionRead),
									string(tfe.WorkspaceStateVersionsPermissionWrite),
								},
								false,
							),
						},

						"variables": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ValidateFunc: validation.StringInSlice(
								[]string{
									string(tfe.WorkspaceVariablesPermissionNone),
									string(tfe.WorkspaceVariablesPermissionRead),
									string(tfe.WorkspaceVariablesPermissionWrite),
								},
								false,
							),
						},
					},
				},
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
		Access:          *tfe.ProjectAccess(tfe.TeamProjectAccessType(access)),
		Team:            tm,
		Project:         proj,
		ProjectAccess:   &tfe.TeamProjectAccessProjectPermissionsOptions{},
		WorkspaceAccess: &tfe.TeamProjectAccessWorkspacePermissionsOptions{},
	}

	if v, ok := d.GetOk("project_access.0.settings"); ok {
		options.ProjectAccess.Settings = tfe.ProjectSettingsPermission(tfe.ProjectSettingsPermissionType(v.(string)))
	}

	if v, ok := d.GetOk("project_access.0.teams"); ok {
		options.ProjectAccess.Teams = tfe.ProjectTeamsPermission(tfe.ProjectTeamsPermissionType(v.(string)))
	}

	if v, ok := d.GetOk("workspace_access.0.state_versions"); ok {
		options.WorkspaceAccess.StateVersions = tfe.WorkspaceStateVersionsPermission(tfe.WorkspaceStateVersionsPermissionType(v.(string)))
	}

	if v, ok := d.GetOk("workspace_access.0.sentinel_mocks"); ok {
		options.WorkspaceAccess.SentinelMocks = tfe.WorkspaceSentinelMocksPermission(tfe.WorkspaceSentinelMocksPermissionType(v.(string)))
	}

	if v, ok := d.GetOk("workspace_access.0.runs"); ok {
		options.WorkspaceAccess.Runs = tfe.WorkspaceRunsPermission(tfe.WorkspaceRunsPermissionType(v.(string)))
	}

	if v, ok := d.GetOk("workspace_access.0.variables"); ok {
		options.WorkspaceAccess.Variables = tfe.WorkspaceVariablesPermission(tfe.WorkspaceVariablesPermissionType(v.(string)))
	}

	if v, ok := d.GetOk("workspace_access.0.create"); ok {
		options.WorkspaceAccess.Create = tfe.Bool(v.(bool))
	}

	if v, ok := d.GetOk("workspace_access.0.locking"); ok {
		options.WorkspaceAccess.Locking = tfe.Bool(v.(bool))
	}

	if v, ok := d.GetOk("workspace_access.0.move"); ok {
		options.WorkspaceAccess.Move = tfe.Bool(v.(bool))
	}

	if v, ok := d.GetOk("workspace_access.0.delete"); ok {
		options.WorkspaceAccess.Delete = tfe.Bool(v.(bool))
	}

	if v, ok := d.GetOk("workspace_access.0.run_tasks"); ok {
		options.WorkspaceAccess.RunTasks = tfe.Bool(v.(bool))
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

	// These two fields are only available in TFC and TFE v202308-1 and later
	if tmAccess.ProjectAccess != nil {
		project_access := []map[string]interface{}{{
			"settings": tmAccess.ProjectAccess.ProjectSettingsPermission,
			"teams":    tmAccess.ProjectAccess.ProjectTeamsPermission,
		}}

		if err := d.Set("project_access", project_access); err != nil {
			return diag.Errorf("Error setting configuration of team project access %s: %v", d.Id(), err)
		}
	}

	if tmAccess.WorkspaceAccess != nil {
		workspace_access := []map[string]interface{}{{
			"state_versions": tmAccess.WorkspaceAccess.WorkspaceStateVersionsPermission,
			"sentinel_mocks": tmAccess.WorkspaceAccess.WorkspaceSentinelMocksPermission,
			"runs":           tmAccess.WorkspaceAccess.WorkspaceRunsPermission,
			"variables":      tmAccess.WorkspaceAccess.WorkspaceVariablesPermission,
			"create":         tmAccess.WorkspaceAccess.WorkspaceCreatePermission,
			"locking":        tmAccess.WorkspaceAccess.WorkspaceLockingPermission,
			"move":           tmAccess.WorkspaceAccess.WorkspaceMovePermission,
			"delete":         tmAccess.WorkspaceAccess.WorkspaceDeletePermission,
			"run_tasks":      tmAccess.WorkspaceAccess.WorkspaceRunTasksPermission,
		}}

		if err := d.Set("workspace_access", workspace_access); err != nil {
			return diag.Errorf("Error setting configuration of team workspace access %s: %v", d.Id(), err)
		}
	}

	return nil
}

func resourceTFETeamProjectAccessUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ConfiguredClient)

	// create an options struct
	options := tfe.TeamProjectAccessUpdateOptions{
		ProjectAccess:   &tfe.TeamProjectAccessProjectPermissionsOptions{},
		WorkspaceAccess: &tfe.TeamProjectAccessWorkspacePermissionsOptions{},
	}

	// Set access level
	access := d.Get("access").(string)
	options.Access = tfe.ProjectAccess(tfe.TeamProjectAccessType(access))

	if d.HasChange("project_access.0.settings") {
		if settings, ok := d.GetOk("project_access.0.settings"); ok {
			projectSettingsPermissionType := tfe.ProjectSettingsPermissionType(settings.(string))
			options.ProjectAccess.Settings = &projectSettingsPermissionType
		}
	}

	if d.HasChange("project_access.0.teams") {
		if teams, ok := d.GetOk("project_access.0.teams"); ok {
			projectTeamsPermissionType := tfe.ProjectTeamsPermissionType(teams.(string))
			options.ProjectAccess.Teams = &projectTeamsPermissionType
		}
	}

	if d.HasChange("workspace_access.0.state_versions") {
		if state_versions, ok := d.GetOk("workspace_access.0.state_versions"); ok {
			workspaceStateVersionsPermissionType := tfe.WorkspaceStateVersionsPermissionType(state_versions.(string))
			options.WorkspaceAccess.StateVersions = &workspaceStateVersionsPermissionType
		}
	}

	if d.HasChange("workspace_access.0.sentinel_mocks") {
		if sentinel_mocks, ok := d.GetOk("workspace_access.0.sentinel_mocks"); ok {
			workspaceSentinelMocksPermissionType := tfe.WorkspaceSentinelMocksPermissionType(sentinel_mocks.(string))
			options.WorkspaceAccess.SentinelMocks = &workspaceSentinelMocksPermissionType
		}
	}

	if d.HasChange("workspace_access.0.runs") {
		if runs, ok := d.GetOk("workspace_access.0.runs"); ok {
			workspaceRunsPermissionType := tfe.WorkspaceRunsPermissionType(runs.(string))
			options.WorkspaceAccess.Runs = &workspaceRunsPermissionType
		}
	}

	if d.HasChange("workspace_access.0.variables") {
		if variables, ok := d.GetOk("workspace_access.0.variables"); ok {
			workspaceVariablesPermissionType := tfe.WorkspaceVariablesPermissionType(variables.(string))
			options.WorkspaceAccess.Variables = &workspaceVariablesPermissionType
		}
	}

	if d.HasChange("workspace_access.0.create") {
		if create, ok := d.GetOkExists("workspace_access.0.create"); ok {
			create := tfe.Bool(create.(bool))
			options.WorkspaceAccess.Create = create
		}
	}

	if d.HasChange("workspace_access.0.locking") {
		if locking, ok := d.GetOkExists("workspace_access.0.locking"); ok {
			options.WorkspaceAccess.Locking = tfe.Bool(locking.(bool))
		}
	}

	if d.HasChange("workspace_access.0.move") {
		if move, ok := d.GetOkExists("workspace_access.0.move"); ok {
			options.WorkspaceAccess.Move = tfe.Bool(move.(bool))
		}
	}

	if d.HasChange("workspace_access.0.delete") {
		if delete, ok := d.GetOkExists("workspace_access.0.delete"); ok {
			options.WorkspaceAccess.Delete = tfe.Bool(delete.(bool))
		}
	}

	if d.HasChange("workspace_access.0.run_tasks") {
		if run_tasks, ok := d.GetOkExists("workspace_access.0.run_tasks"); ok {
			options.WorkspaceAccess.RunTasks = tfe.Bool(run_tasks.(bool))
		}
	}

	log.Printf("[DEBUG] Update team project access: %s", d.Id())
	_, err := config.Client.TeamProjectAccess.Update(ctx, d.Id(), options)
	if err != nil {
		return diag.Errorf(
			"Error updating team project access %s: %v", d.Id(), err)
	}

	return resourceTFETeamProjectAccessRead(ctx, d, meta)
}

func resourceTFETeamProjectAccessDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Delete team access: %s", d.Id())
	err := config.Client.TeamProjectAccess.Remove(ctx, d.Id())
	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
			return nil
		}
		return diag.Errorf("Error deleting team project access %s: %v", d.Id(), err)
	}

	return nil
}

// You cannot set custom permissions when access level is not "custom"
func checkForCustomPermissions(_ context.Context, d *schema.ResourceDiff, meta interface{}) error {

	if access, ok := d.GetOk("access"); ok && access != "custom" {
		// is an empty [] if project_access is not in the config
		project_access := d.GetRawConfig().GetAttr("project_access").AsValueSet().Values()
		if len(project_access) != 0 {
			return fmt.Errorf("you can only set project_access permissions with access level custom")
		}

		// is an empty [] if project_access is not in the config
		workspace_access := d.GetRawConfig().GetAttr("workspace_access").AsValueSet().Values()
		if len(workspace_access) != 0 {
			return fmt.Errorf("you can only set workspace_access permissions with access level custom")
		}
	}

	return nil
}
