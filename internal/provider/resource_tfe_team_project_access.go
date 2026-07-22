// Copyright IBM Corp. 2018, 2025
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

	tfe "github.com/hashicorp/go-tfe"
	tfev2 "github.com/hashicorp/go-tfe/v2"
	"github.com/hashicorp/go-tfe/v2/api/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceTFETeamProjectAccess() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a team's permissions on a project.",

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
				Description: "Type of fixed access to grant. Valid values are admin, maintain, write, read, or custom.",
				Type:        schema.TypeString,
				Required:    true,
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
				Description: "ID of the team to add to the project.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},

			"project_id": {
				Description: "ID of the project to which the team will be added.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},

			"project_access": {
				Description: "Settings for the team's custom permissions on the project itself. Only used when access is custom.",
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"settings": {
							Description: "The permission to grant for the project's settings. Default: read. Valid strings: read, update, or delete.",
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
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
							Description: "The permission to grant for the project's teams. Default: none. Valid strings: none, read, or manage.",
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							ValidateFunc: validation.StringInSlice(
								[]string{
									string(tfe.ProjectTeamsPermissionNone),
									string(tfe.ProjectTeamsPermissionRead),
									string(tfe.ProjectTeamsPermissionManage),
								},
								false,
							),
						},

						"variable_sets": {
							Description: "The permission to grant for the project's variable sets. Default: none. Valid strings: none, read, or write.",
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							ValidateFunc: validation.StringInSlice(
								[]string{
									string(tfe.ProjectVariableSetsPermissionNone),
									string(tfe.ProjectVariableSetsPermissionRead),
									string(tfe.ProjectVariableSetsPermissionWrite),
								},
								false,
							),
						},
					},
				},
			},

			"workspace_access": {
				Description: "Settings for the team's custom permissions on all workspaces (and future workspaces) in the project. Only used when access is custom.",
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"create": {
							Description: "The permission to create the project's workspaces in the project. Default: false.",
							Type:        schema.TypeBool,
							Optional:    true,
							Computed:    true,
						},

						"locking": {
							Description: "The permission to manually lock or unlock the project's workspaces. Default: false.",
							Type:        schema.TypeBool,
							Optional:    true,
							Computed:    true,
						},

						"move": {
							Description: "The permission to move workspaces into and out of the project. The team must also have permissions to the project(s) receiving the workspace(s). Default: false.",
							Type:        schema.TypeBool,
							Optional:    true,
							Computed:    true,
						},

						"delete": {
							Description: "The permission to delete the project's workspaces. Default: false.",
							Type:        schema.TypeBool,
							Optional:    true,
							Computed:    true,
						},

						"run_tasks": {
							Description: "The permission to manage run tasks within the project's workspaces. Default: false.",
							Type:        schema.TypeBool,
							Optional:    true,
							Computed:    true,
						},

						"policy_overrides": {
							Description: "Allows a team to override soft-mandatory policy evaluations, provided that team has been granted the org level delegate policy overrides permission. Default: false.",
							Type:        schema.TypeBool,
							Optional:    true,
							Computed:    true,
						},

						"runs": {
							Description: "The permission to grant the project's workspaces' runs. Default: read. Valid strings: read, plan, or apply.",
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
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
							Description: "The permission to grant the project's workspaces' Sentinel mocks. Default: none. Valid strings: none, or read.",
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							ValidateFunc: validation.StringInSlice(
								[]string{
									string(tfe.WorkspaceSentinelMocksPermissionNone),
									string(tfe.WorkspaceSentinelMocksPermissionRead),
								},
								false,
							),
						},

						"state_versions": {
							Description: "The permission to grant the project's workspaces' state versions. Default: none. Valid strings: none, read-outputs, read, or write.",
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
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
							Description: "The permission to grant the project's workspaces' variables. Default: none. Valid strings: none, read, or write.",
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
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
	proj, err := config.ClientV2.API.Projects().ByProject_id(projectID).Get(ctx, nil)
	if err != nil {
		return diag.Errorf(
			"Error retrieving project %s: %v", projectID, err)
	}
	if proj == nil || proj.GetData() == nil {
		return diag.Errorf("Error retrieving project %s: no data returned", projectID)
	}

	// Get the team.
	teamID := d.Get("team_id").(string)
	tm, err := config.ClientV2.API.Teams().ById(teamID).Get(ctx, nil)
	if err != nil {
		return diag.Errorf("Error retrieving team %s: %v", teamID, err)
	}
	if tm == nil || tm.GetData() == nil {
		return diag.Errorf("Error retrieving team %s: no data returned", teamID)
	}

	// Create a new attributes struct.
	accessValue, aerr := models.ParseTeamProjects_attributes_access(access)
	if aerr != nil {
		return diag.Errorf("invalid team project access value %q: %v", access, aerr)
	}
	attributes := models.NewTeamProjects_attributes()
	attributes.SetAccess(accessValue.(*models.TeamProjects_attributes_access))

	projectAccess := models.NewTeamProjects_attributes_projectAccess()
	workspaceAccess := models.NewTeamProjects_attributes_workspaceAccess()

	if v, ok := d.GetOk("project_access.0.settings"); ok {
		settingsValue, err := models.ParseTeamProjects_attributes_projectAccess_settings(v.(string))
		if err != nil {
			return diag.Errorf("invalid project_access.settings value %q: %v", v.(string), err)
		}
		projectAccess.SetSettings(settingsValue.(*models.TeamProjects_attributes_projectAccess_settings))
	}

	if v, ok := d.GetOk("project_access.0.teams"); ok {
		teamsValue, err := models.ParseTeamProjects_attributes_projectAccess_teams(v.(string))
		if err != nil {
			return diag.Errorf("invalid project_access.teams value %q: %v", v.(string), err)
		}
		projectAccess.SetTeams(teamsValue.(*models.TeamProjects_attributes_projectAccess_teams))
	}

	if v, ok := d.GetOk("project_access.0.variable_sets"); ok {
		projectAccess.SetVariableSets(ptr(v.(string)))
	}

	if v, ok := d.GetOk("workspace_access.0.state_versions"); ok {
		stateVersionsValue, err := models.ParseTeamProjects_attributes_workspaceAccess_stateVersions(v.(string))
		if err != nil {
			return diag.Errorf("invalid workspace_access.state_versions value %q: %v", v.(string), err)
		}
		workspaceAccess.SetStateVersions(stateVersionsValue.(*models.TeamProjects_attributes_workspaceAccess_stateVersions))
	}

	if v, ok := d.GetOk("workspace_access.0.sentinel_mocks"); ok {
		sentinelMocksValue, err := models.ParseTeamProjects_attributes_workspaceAccess_sentinelMocks(v.(string))
		if err != nil {
			return diag.Errorf("invalid workspace_access.sentinel_mocks value %q: %v", v.(string), err)
		}
		workspaceAccess.SetSentinelMocks(sentinelMocksValue.(*models.TeamProjects_attributes_workspaceAccess_sentinelMocks))
	}

	if v, ok := d.GetOk("workspace_access.0.runs"); ok {
		runsValue, err := models.ParseTeamProjects_attributes_workspaceAccess_runs(v.(string))
		if err != nil {
			return diag.Errorf("invalid workspace_access.runs value %q: %v", v.(string), err)
		}
		workspaceAccess.SetRuns(runsValue.(*models.TeamProjects_attributes_workspaceAccess_runs))
	}

	if v, ok := d.GetOk("workspace_access.0.variables"); ok {
		variablesValue, err := models.ParseTeamProjects_attributes_workspaceAccess_variables(v.(string))
		if err != nil {
			return diag.Errorf("invalid workspace_access.variables value %q: %v", v.(string), err)
		}
		workspaceAccess.SetVariables(variablesValue.(*models.TeamProjects_attributes_workspaceAccess_variables))
	}

	if v, ok := d.GetOk("workspace_access.0.create"); ok {
		workspaceAccess.SetCreate(ptr(v.(bool)))
	}

	if v, ok := d.GetOk("workspace_access.0.locking"); ok {
		workspaceAccess.SetLocking(ptr(v.(bool)))
	}

	if v, ok := d.GetOk("workspace_access.0.move"); ok {
		workspaceAccess.SetMove(ptr(v.(bool)))
	}

	if v, ok := d.GetOk("workspace_access.0.delete"); ok {
		workspaceAccess.SetDelete(ptr(v.(bool)))
	}

	if v, ok := d.GetOk("workspace_access.0.run_tasks"); ok {
		workspaceAccess.SetRunTasks(ptr(v.(bool)))
	}

	if v, ok := d.GetOk("workspace_access.0.policy_overrides"); ok {
		workspaceAccess.SetPolicyOverrides(ptr(v.(bool)))
	}

	attributes.SetProjectAccess(projectAccess)
	attributes.SetWorkspaceAccess(workspaceAccess)

	teamRelationship := models.NewTeamsId()
	teamRelationshipData := models.NewTeamsId_data()
	teamRelationshipData.SetId(ptr(teamID))
	teamRelationshipData.SetTypeEscaped(ptr(models.TEAMS_TEAMSID_DATA_TYPE))
	teamRelationship.SetData(teamRelationshipData)

	projectRelationship := models.NewProjectsId()
	projectRelationshipData := models.NewProjectsId_data()
	projectRelationshipData.SetId(ptr(projectID))
	projectRelationshipData.SetTypeEscaped(ptr(models.PROJECTS_PROJECTSID_DATA_TYPE))
	projectRelationship.SetData(projectRelationshipData)

	relationships := models.NewTeamProjects_relationships()
	relationships.SetTeam(teamRelationship)
	relationships.SetProject(projectRelationship)

	teamProject := models.NewTeamProjects()
	teamProject.SetTypeEscaped(ptr(models.TEAMPROJECTS_TEAMPROJECTS_TYPE))
	teamProject.SetAttributes(attributes)
	teamProject.SetRelationships(relationships)

	envelope := models.NewTeamProjectsEnvelope()
	envelope.SetData(teamProject)

	teamName := valueOrZero(tm.GetData().GetAttributes().GetName())
	projectName := valueOrZero(proj.GetData().GetAttributes().GetName())

	log.Printf("[DEBUG] Give team %s %s access to project: %s", teamName, access, projectName)
	result, err := config.ClientV2.API.TeamProjects().Post(ctx, envelope, nil)
	if err != nil {
		return diag.Errorf(
			"Error giving team %s %s access to project %s: %v", teamName, access, projectName, err)
	}
	if result == nil || result.GetData() == nil {
		return diag.Errorf("Error giving team %s %s access to project %s: no data returned", teamName, access, projectName)
	}

	d.SetId(valueOrZero(result.GetData().GetId()))

	return resourceTFETeamProjectAccessRead(ctx, d, meta)
}

func resourceTFETeamProjectAccessRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Read configuration of team access: %s", d.Id())
	result, err := config.ClientV2.API.TeamProjects().ByTeam_project_id(d.Id()).Get(ctx, nil)
	if err != nil {
		if errors.Is(err, tfev2.ErrNotFound) {
			log.Printf("[DEBUG] Team project access %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		return diag.Errorf("Error reading configuration of team project access %s: %v", d.Id(), err)
	}
	if result == nil || result.GetData() == nil {
		log.Printf("[DEBUG] Team project access %s no longer exists", d.Id())
		d.SetId("")
		return nil
	}
	tmAccess := result.GetData()
	attrs := tmAccess.GetAttributes()

	// Update config.
	d.Set("access", enumStringOrEmpty(attrs.GetAccess()))

	if relationships := tmAccess.GetRelationships(); relationships != nil && relationships.GetTeam() != nil && relationships.GetTeam().GetData() != nil {
		d.Set("team_id", valueOrZero(relationships.GetTeam().GetData().GetId()))
	} else {
		d.Set("team_id", "")
	}

	if relationships := tmAccess.GetRelationships(); relationships != nil && relationships.GetProject() != nil && relationships.GetProject().GetData() != nil {
		d.Set("project_id", valueOrZero(relationships.GetProject().GetData().GetId()))
	} else {
		d.Set("project_id", "")
	}

	if pa := attrs.GetProjectAccess(); pa != nil {
		projectAccess := []map[string]interface{}{{
			"settings":      enumStringOrEmpty(pa.GetSettings()),
			"teams":         enumStringOrEmpty(pa.GetTeams()),
			"variable_sets": valueOrZero(pa.GetVariableSets()),
		}}

		if err := d.Set("project_access", projectAccess); err != nil {
			return diag.Errorf("Error setting configuration of team project access %s: %v", d.Id(), err)
		}
	}

	if wa := attrs.GetWorkspaceAccess(); wa != nil {
		workspaceAccess := []map[string]interface{}{{
			"state_versions":   enumStringOrEmpty(wa.GetStateVersions()),
			"sentinel_mocks":   enumStringOrEmpty(wa.GetSentinelMocks()),
			"runs":             enumStringOrEmpty(wa.GetRuns()),
			"variables":        enumStringOrEmpty(wa.GetVariables()),
			"create":           valueOrZero(wa.GetCreate()),
			"locking":          valueOrZero(wa.GetLocking()),
			"move":             valueOrZero(wa.GetMove()),
			"delete":           valueOrZero(wa.GetDelete()),
			"run_tasks":        valueOrZero(wa.GetRunTasks()),
			"policy_overrides": valueOrZero(wa.GetPolicyOverrides()),
		}}

		if err := d.Set("workspace_access", workspaceAccess); err != nil {
			return diag.Errorf("Error setting configuration of team workspace access %s: %v", d.Id(), err)
		}
	}

	return nil
}

func resourceTFETeamProjectAccessUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ConfiguredClient)

	// create an attributes struct
	attributes := models.NewTeamProjects_attributes()
	projectAccess := models.NewTeamProjects_attributes_projectAccess()
	workspaceAccess := models.NewTeamProjects_attributes_workspaceAccess()

	// Set access level
	access := d.Get("access").(string)
	accessValue, aerr := models.ParseTeamProjects_attributes_access(access)
	if aerr != nil {
		return diag.Errorf("invalid team project access value %q: %v", access, aerr)
	}
	attributes.SetAccess(accessValue.(*models.TeamProjects_attributes_access))

	if d.HasChange("project_access.0.settings") {
		if settings, ok := d.GetOk("project_access.0.settings"); ok {
			settingsValue, err := models.ParseTeamProjects_attributes_projectAccess_settings(settings.(string))
			if err != nil {
				return diag.Errorf("invalid project_access.settings value %q: %v", settings.(string), err)
			}
			projectAccess.SetSettings(settingsValue.(*models.TeamProjects_attributes_projectAccess_settings))
		}
	}

	if d.HasChange("project_access.0.teams") {
		if teams, ok := d.GetOk("project_access.0.teams"); ok {
			teamsValue, err := models.ParseTeamProjects_attributes_projectAccess_teams(teams.(string))
			if err != nil {
				return diag.Errorf("invalid project_access.teams value %q: %v", teams.(string), err)
			}
			projectAccess.SetTeams(teamsValue.(*models.TeamProjects_attributes_projectAccess_teams))
		}
	}

	if d.HasChange("project_access.0.variable_sets") {
		if variableSets, ok := d.GetOk("project_access.0.variable_sets"); ok {
			projectAccess.SetVariableSets(ptr(variableSets.(string)))
		}
	}

	if d.HasChange("workspace_access.0.state_versions") {
		if stateVersions, ok := d.GetOk("workspace_access.0.state_versions"); ok {
			stateVersionsValue, err := models.ParseTeamProjects_attributes_workspaceAccess_stateVersions(stateVersions.(string))
			if err != nil {
				return diag.Errorf("invalid workspace_access.state_versions value %q: %v", stateVersions.(string), err)
			}
			workspaceAccess.SetStateVersions(stateVersionsValue.(*models.TeamProjects_attributes_workspaceAccess_stateVersions))
		}
	}

	if d.HasChange("workspace_access.0.sentinel_mocks") {
		if sentinelMocks, ok := d.GetOk("workspace_access.0.sentinel_mocks"); ok {
			sentinelMocksValue, err := models.ParseTeamProjects_attributes_workspaceAccess_sentinelMocks(sentinelMocks.(string))
			if err != nil {
				return diag.Errorf("invalid workspace_access.sentinel_mocks value %q: %v", sentinelMocks.(string), err)
			}
			workspaceAccess.SetSentinelMocks(sentinelMocksValue.(*models.TeamProjects_attributes_workspaceAccess_sentinelMocks))
		}
	}

	if d.HasChange("workspace_access.0.runs") {
		if runs, ok := d.GetOk("workspace_access.0.runs"); ok {
			runsValue, err := models.ParseTeamProjects_attributes_workspaceAccess_runs(runs.(string))
			if err != nil {
				return diag.Errorf("invalid workspace_access.runs value %q: %v", runs.(string), err)
			}
			workspaceAccess.SetRuns(runsValue.(*models.TeamProjects_attributes_workspaceAccess_runs))
		}
	}

	if d.HasChange("workspace_access.0.variables") {
		if variables, ok := d.GetOk("workspace_access.0.variables"); ok {
			variablesValue, err := models.ParseTeamProjects_attributes_workspaceAccess_variables(variables.(string))
			if err != nil {
				return diag.Errorf("invalid workspace_access.variables value %q: %v", variables.(string), err)
			}
			workspaceAccess.SetVariables(variablesValue.(*models.TeamProjects_attributes_workspaceAccess_variables))
		}
	}

	if d.HasChange("workspace_access.0.create") {
		if create, ok := d.GetOkExists("workspace_access.0.create"); ok {
			workspaceAccess.SetCreate(ptr(create.(bool)))
		}
	}

	if d.HasChange("workspace_access.0.locking") {
		if locking, ok := d.GetOkExists("workspace_access.0.locking"); ok {
			workspaceAccess.SetLocking(ptr(locking.(bool)))
		}
	}

	if d.HasChange("workspace_access.0.move") {
		if move, ok := d.GetOkExists("workspace_access.0.move"); ok {
			workspaceAccess.SetMove(ptr(move.(bool)))
		}
	}

	if d.HasChange("workspace_access.0.delete") {
		if deleteAttr, ok := d.GetOkExists("workspace_access.0.delete"); ok {
			workspaceAccess.SetDelete(ptr(deleteAttr.(bool)))
		}
	}

	if d.HasChange("workspace_access.0.run_tasks") {
		if runTasks, ok := d.GetOkExists("workspace_access.0.run_tasks"); ok {
			workspaceAccess.SetRunTasks(ptr(runTasks.(bool)))
		}
	}

	if d.HasChange("workspace_access.0.policy_overrides") {
		if v, ok := d.GetOkExists("workspace_access.0.policy_overrides"); ok {
			workspaceAccess.SetPolicyOverrides(ptr(v.(bool)))
		}
	}

	attributes.SetProjectAccess(projectAccess)
	attributes.SetWorkspaceAccess(workspaceAccess)

	teamProject := models.NewTeamProjects()
	teamProject.SetTypeEscaped(ptr(models.TEAMPROJECTS_TEAMPROJECTS_TYPE))
	teamProject.SetId(ptr(d.Id()))
	teamProject.SetAttributes(attributes)

	envelope := models.NewTeamProjectsEnvelope()
	envelope.SetData(teamProject)

	log.Printf("[DEBUG] Update team project access: %s", d.Id())
	_, err := config.ClientV2.API.TeamProjects().ByTeam_project_id(d.Id()).Patch(ctx, envelope, nil)
	if err != nil {
		return diag.Errorf(
			"Error updating team project access %s: %v", d.Id(), err)
	}

	return resourceTFETeamProjectAccessRead(ctx, d, meta)
}

func resourceTFETeamProjectAccessDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Delete team access: %s", d.Id())
	err := config.ClientV2.API.TeamProjects().ByTeam_project_id(d.Id()).Delete(ctx, nil)
	if err != nil {
		if errors.Is(err, tfev2.ErrNotFound) {
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
		projectAccess := d.GetRawConfig().GetAttr("project_access").AsValueSet().Values()
		if len(projectAccess) != 0 {
			return fmt.Errorf("you can only set project_access permissions with access level custom")
		}

		// is an empty [] if project_access is not in the config
		workspaceAccess := d.GetRawConfig().GetAttr("workspace_access").AsValueSet().Values()
		if len(workspaceAccess) != 0 {
			return fmt.Errorf("you can only set workspace_access permissions with access level custom")
		}
	}

	return nil
}
