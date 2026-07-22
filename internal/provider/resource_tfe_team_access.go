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
	"strings"

	tfe "github.com/hashicorp/go-tfe"
	tfev2 "github.com/hashicorp/go-tfe/v2"
	"github.com/hashicorp/go-tfe/v2/api/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	// Schema field names for tfe_team_access resource
	teamAccessAccessKey      = "access"
	teamAccessPermissionsKey = "permissions"
	teamAccessTeamIDKey      = "team_id"
	teamAccessWorkspaceIDKey = "workspace_id"

	// Permission field names
	permissionsRunsKey             = "runs"
	permissionsVariablesKey        = "variables"
	permissionsStateVersionsKey    = "state_versions"
	permissionsSentinelMocksKey    = "sentinel_mocks"
	permissionsWorkspaceLockingKey = "workspace_locking"
	permissionsRunTasksKey         = "run_tasks"
	permissionsPolicyOverridesKey  = "policy_overrides"
)

func resourceTFETeamAccess() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a team's permissions on a workspace.\n\n" +
			"-> **Note:** At least one of `access` or `permissions` must be provided, but not both. Whichever is omitted will automatically reflect the state of the other.",

		Create: resourceTFETeamAccessCreate,
		Read:   resourceTFETeamAccessRead,
		Update: resourceTFETeamAccessUpdate,
		Delete: resourceTFETeamAccessDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceTFETeamAccessImporter,
		},

		CustomizeDiff: setCustomOrComputedPermissions,
		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceTfeTeamAccessResourceV0().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceTfeTeamAccessStateUpgradeV0,
				Version: 0,
			},
		},

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The team access ID.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			teamAccessAccessKey: {
				Description: "Type of fixed access to grant. Valid values are admin, read, plan, or write. To use custom permissions, use a permissions block instead. This value must not be provided if permissions is provided.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				// This should be moved to the Resource level when possible:
				// https://github.com/hashicorp/terraform-plugin-sdk/issues/470
				ExactlyOneOf: []string{teamAccessAccessKey, teamAccessPermissionsKey},
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

			teamAccessPermissionsKey: {
				Description: "Permissions to grant using [custom workspace permissions](https://developer.hashicorp.com/terraform/cloud-docs/users-teams-organizations/permissions#custom-workspace-permissions). This value must not be provided if access is provided.",
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						permissionsRunsKey: {
							Description: "The permission to grant the team on the workspace's runs. Valid values are read, plan, or apply.",
							Type:        schema.TypeString,
							Required:    true,
							ValidateFunc: validation.StringInSlice(
								[]string{
									string(tfe.RunsPermissionRead),
									string(tfe.RunsPermissionPlan),
									string(tfe.RunsPermissionApply),
								},
								false,
							),
						},

						permissionsVariablesKey: {
							Description: "The permission to grant the team on the workspace's variables. Valid values are none, read, or write.",
							Type:        schema.TypeString,
							Required:    true,
							ValidateFunc: validation.StringInSlice(
								[]string{
									string(tfe.VariablesPermissionNone),
									string(tfe.VariablesPermissionRead),
									string(tfe.VariablesPermissionWrite),
								},
								false,
							),
						},

						permissionsStateVersionsKey: {
							Description: "The permission to grant the team on the workspace's state versions. Valid values are none, read, read-outputs, or write.",
							Type:        schema.TypeString,
							Required:    true,
							ValidateFunc: validation.StringInSlice(
								[]string{
									string(tfe.StateVersionsPermissionNone),
									string(tfe.StateVersionsPermissionReadOutputs),
									string(tfe.StateVersionsPermissionRead),
									string(tfe.StateVersionsPermissionWrite),
								},
								false,
							),
						},

						permissionsSentinelMocksKey: {
							Description: "The permission to grant the team on the workspace's generated Sentinel mocks. Valid values are none or read.",
							Type:        schema.TypeString,
							Required:    true,
							ValidateFunc: validation.StringInSlice(
								[]string{
									string(tfe.SentinelMocksPermissionNone),
									string(tfe.SentinelMocksPermissionRead),
								},
								false,
							),
						},

						permissionsWorkspaceLockingKey: {
							Description: "Whether or not to grant the team permission to manually lock/unlock the workspace.",
							Type:        schema.TypeBool,
							Required:    true,
						},

						permissionsRunTasksKey: {
							Description: "Whether or not to grant the team permission to manage workspace run tasks.",
							Type:        schema.TypeBool,
							Required:    true,
						},

						"policy_overrides": {
							Description: "Allows a team to override soft-mandatory policy evaluations, provided that team has been granted the org level delegate policy overrides permission.",
							Type:        schema.TypeBool,
							Optional:    true,
							Computed:    true,
						},
					},
				},
			},

			teamAccessTeamIDKey: {
				Description: "ID of the team to add to the workspace.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},

			teamAccessWorkspaceIDKey: {
				Description: "ID of the workspace to which the team will be added.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				ValidateFunc: validation.StringMatch(
					workspaceIDRegexp,
					"must be a valid workspace ID (ws-<RANDOM STRING>)",
				),
			},
		},
	}
}

// applyTeamWorkspacePermissionAttrs reads the seven custom workspace
// permission fields from Terraform state and writes them onto attrs.
//
// When checkChanges is false (Create path), all fields present in config are
// applied unconditionally. d.HasChange always returns false during Create for
// bool fields whose config value equals their zero value (false), so using it
// there would silently skip sending those fields to the API.
//
// When checkChanges is true (Update path), each field is only applied when it
// has actually changed, preserving the existing partial-update semantics.
func applyTeamWorkspacePermissionAttrs(d *schema.ResourceData, attrs models.TeamWorkspaces_attributesable, checkChanges bool) error {
	apply := func(path string) bool {
		return !checkChanges || d.HasChange(path)
	}

	runsPath := fmt.Sprintf("%s.0.%s", teamAccessPermissionsKey, permissionsRunsKey)
	if apply(runsPath) {
		if v, ok := d.GetOk(runsPath); ok {
			runsValue, err := models.ParseTeamWorkspaces_attributes_runs(v.(string))
			if err != nil {
				return fmt.Errorf("invalid runs permission %q: %w", v.(string), err)
			}
			attrs.SetRuns(runsValue.(*models.TeamWorkspaces_attributes_runs))
		}
	}

	variablesPath := fmt.Sprintf("%s.0.%s", teamAccessPermissionsKey, permissionsVariablesKey)
	if apply(variablesPath) {
		if v, ok := d.GetOk(variablesPath); ok {
			variablesValue, err := models.ParseTeamWorkspaces_attributes_variables(v.(string))
			if err != nil {
				return fmt.Errorf("invalid variables permission %q: %w", v.(string), err)
			}
			attrs.SetVariables(variablesValue.(*models.TeamWorkspaces_attributes_variables))
		}
	}

	stateVersionsPath := fmt.Sprintf("%s.0.%s", teamAccessPermissionsKey, permissionsStateVersionsKey)
	if apply(stateVersionsPath) {
		if v, ok := d.GetOk(stateVersionsPath); ok {
			stateVersionsValue, err := models.ParseTeamWorkspaces_attributes_stateVersions(v.(string))
			if err != nil {
				return fmt.Errorf("invalid state_versions permission %q: %w", v.(string), err)
			}
			attrs.SetStateVersions(stateVersionsValue.(*models.TeamWorkspaces_attributes_stateVersions))
		}
	}

	sentinelMocksPath := fmt.Sprintf("%s.0.%s", teamAccessPermissionsKey, permissionsSentinelMocksKey)
	if apply(sentinelMocksPath) {
		if v, ok := d.GetOk(sentinelMocksPath); ok {
			sentinelMocksValue, err := models.ParseTeamWorkspaces_attributes_sentinelMocks(v.(string))
			if err != nil {
				return fmt.Errorf("invalid sentinel_mocks permission %q: %w", v.(string), err)
			}
			attrs.SetSentinelMocks(sentinelMocksValue.(*models.TeamWorkspaces_attributes_sentinelMocks))
		}
	}

	workspaceLockingPath := fmt.Sprintf("%s.0.%s", teamAccessPermissionsKey, permissionsWorkspaceLockingKey)
	if apply(workspaceLockingPath) {
		if v, ok := d.GetOkExists(workspaceLockingPath); ok { //nolint:staticcheck
			attrs.SetWorkspaceLocking(ptr(v.(bool)))
		}
	}

	runTasksPath := fmt.Sprintf("%s.0.%s", teamAccessPermissionsKey, permissionsRunTasksKey)
	if apply(runTasksPath) {
		if v, ok := d.GetOkExists(runTasksPath); ok { //nolint:staticcheck
			attrs.SetRunTasks(ptr(v.(bool)))
		}
	}

	policyOverridesPath := fmt.Sprintf("%s.0.%s", teamAccessPermissionsKey, permissionsPolicyOverridesKey)
	if apply(policyOverridesPath) {
		if v, ok := d.GetOkExists(policyOverridesPath); ok { //nolint:staticcheck
			attrs.SetPolicyOverrides(ptr(v.(bool)))
		}
	}

	return nil
}

func resourceTFETeamAccessCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	// Get the access level
	access := d.Get(teamAccessAccessKey).(string)

	// Get the workspace
	workspaceID := d.Get(teamAccessWorkspaceIDKey).(string)
	ws, err := config.ClientV2.API.Workspaces().ByWorkspace_id(workspaceID).Get(ctx, nil)
	if err != nil {
		return fmt.Errorf(
			"Error retrieving workspace %s: %w", workspaceID, err)
	}
	if ws == nil || ws.GetData() == nil {
		return fmt.Errorf("Error retrieving workspace %s: no data returned", workspaceID)
	}

	// Get the team.
	teamID := d.Get(teamAccessTeamIDKey).(string)
	tm, err := config.ClientV2.API.Teams().ById(teamID).Get(ctx, nil)
	if err != nil {
		return fmt.Errorf("Error retrieving team %s: %w", teamID, err)
	}
	if tm == nil || tm.GetData() == nil {
		return fmt.Errorf("Error retrieving team %s: no data returned", teamID)
	}

	// Create a new attributes struct.
	accessValue, aerr := models.ParseTeamWorkspaces_attributes_access(access)
	if aerr != nil {
		return fmt.Errorf("invalid team access value %q: %w", access, aerr)
	}
	attributes := models.NewTeamWorkspaces_attributes()
	attributes.SetAccess(accessValue.(*models.TeamWorkspaces_attributes_access))

	if err := applyTeamWorkspacePermissionAttrs(d, attributes, false); err != nil {
		return err
	}

	teamRelationship := models.NewTeamsId()
	teamRelationshipData := models.NewTeamsId_data()
	teamRelationshipData.SetId(ptr(teamID))
	teamRelationshipData.SetTypeEscaped(ptr(models.TEAMS_TEAMSID_DATA_TYPE))
	teamRelationship.SetData(teamRelationshipData)

	workspaceRelationship := models.NewWorkspacesId()
	workspaceRelationshipData := models.NewWorkspacesId_data()
	workspaceRelationshipData.SetId(ptr(workspaceID))
	workspaceRelationship.SetData(workspaceRelationshipData)

	relationships := models.NewTeamWorkspaces_relationships()
	relationships.SetTeam(teamRelationship)
	relationships.SetWorkspace(workspaceRelationship)

	teamWorkspace := models.NewTeamWorkspaces()
	teamWorkspace.SetTypeEscaped(ptr(models.TEAMWORKSPACES_TEAMWORKSPACES_TYPE))
	teamWorkspace.SetAttributes(attributes)
	teamWorkspace.SetRelationships(relationships)

	envelope := models.NewTeamWorkspacesEnvelope()
	envelope.SetData(teamWorkspace)

	teamName := valueOrZero(tm.GetData().GetAttributes().GetName())
	workspaceName := valueOrZero(ws.GetData().GetAttributes().GetName())

	log.Printf("[DEBUG] Give team %s %s access to workspace: %s", teamName, access, workspaceName)
	result, err := config.ClientV2.API.TeamWorkspaces().Post(ctx, envelope, nil)
	if err != nil {
		return fmt.Errorf(
			"Error giving team %s %s access to workspace %s: %w", teamName, access, workspaceName, err)
	}
	if result == nil || result.GetData() == nil {
		return fmt.Errorf("Error giving team %s %s access to workspace %s: no data returned", teamName, access, workspaceName)
	}

	d.SetId(valueOrZero(result.GetData().GetId()))

	return resourceTFETeamAccessRead(d, meta)
}

func resourceTFETeamAccessRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Read configuration of team access: %s", d.Id())
	result, err := config.ClientV2.API.TeamWorkspaces().ByTeam_workspace_id(d.Id()).Get(ctx, nil)
	if err != nil {
		if errors.Is(err, tfev2.ErrNotFound) {
			log.Printf("[DEBUG] Team access %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading configuration of team access %s: %w", d.Id(), err)
	}
	if result == nil || result.GetData() == nil {
		log.Printf("[DEBUG] Team access %s no longer exists", d.Id())
		d.SetId("")
		return nil
	}
	tmAccess := result.GetData()
	attrs := tmAccess.GetAttributes()

	// Update config.
	d.Set(teamAccessAccessKey, enumStringOrEmpty(attrs.GetAccess()))
	permissions := []map[string]interface{}{{
		permissionsRunsKey:             enumStringOrEmpty(attrs.GetRuns()),
		permissionsVariablesKey:        enumStringOrEmpty(attrs.GetVariables()),
		permissionsStateVersionsKey:    enumStringOrEmpty(attrs.GetStateVersions()),
		permissionsSentinelMocksKey:    enumStringOrEmpty(attrs.GetSentinelMocks()),
		permissionsWorkspaceLockingKey: valueOrZero(attrs.GetWorkspaceLocking()),
		permissionsRunTasksKey:         valueOrZero(attrs.GetRunTasks()),
		permissionsPolicyOverridesKey:  valueOrZero(attrs.GetPolicyOverrides()),
	}}
	if err := d.Set(teamAccessPermissionsKey, permissions); err != nil {
		return fmt.Errorf("error setting permissions for team access %s: %w", d.Id(), err)
	}

	if relationships := tmAccess.GetRelationships(); relationships != nil && relationships.GetTeam() != nil && relationships.GetTeam().GetData() != nil {
		d.Set(teamAccessTeamIDKey, valueOrZero(relationships.GetTeam().GetData().GetId()))
	} else {
		d.Set(teamAccessTeamIDKey, "")
	}

	return nil
}

func resourceTFETeamAccessUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	// Set access level
	access := d.Get(teamAccessAccessKey).(string)
	accessValue, aerr := models.ParseTeamWorkspaces_attributes_access(access)
	if aerr != nil {
		return fmt.Errorf("invalid team access value %q: %w", access, aerr)
	}
	attributes := models.NewTeamWorkspaces_attributes()
	attributes.SetAccess(accessValue.(*models.TeamWorkspaces_attributes_access))

	if err := applyTeamWorkspacePermissionAttrs(d, attributes, true); err != nil {
		return err
	}

	teamWorkspace := models.NewTeamWorkspaces()
	teamWorkspace.SetTypeEscaped(ptr(models.TEAMWORKSPACES_TEAMWORKSPACES_TYPE))
	teamWorkspace.SetId(ptr(d.Id()))
	teamWorkspace.SetAttributes(attributes)

	envelope := models.NewTeamWorkspacesEnvelope()
	envelope.SetData(teamWorkspace)

	log.Printf("[DEBUG] Update team access: %s", d.Id())
	result, err := config.ClientV2.API.TeamWorkspaces().ByTeam_workspace_id(d.Id()).Patch(ctx, envelope, nil)
	if err != nil {
		return fmt.Errorf(
			"Error updating team access %s: %w", d.Id(), err)
	}
	if result == nil || result.GetData() == nil {
		return fmt.Errorf("Error updating team access %s: no data returned", d.Id())
	}
	updatedAttrs := result.GetData().GetAttributes()

	// Update permissions, in the case that they were marked to be recomputed.
	permissions := []map[string]interface{}{{
		permissionsRunsKey:             enumStringOrEmpty(updatedAttrs.GetRuns()),
		permissionsVariablesKey:        enumStringOrEmpty(updatedAttrs.GetVariables()),
		permissionsStateVersionsKey:    enumStringOrEmpty(updatedAttrs.GetStateVersions()),
		permissionsSentinelMocksKey:    enumStringOrEmpty(updatedAttrs.GetSentinelMocks()),
		permissionsWorkspaceLockingKey: valueOrZero(updatedAttrs.GetWorkspaceLocking()),
		permissionsRunTasksKey:         valueOrZero(updatedAttrs.GetRunTasks()),
		permissionsPolicyOverridesKey:  valueOrZero(updatedAttrs.GetPolicyOverrides()),
	}}
	if err := d.Set(teamAccessPermissionsKey, permissions); err != nil {
		return fmt.Errorf("error setting permissions for team access %s: %w", d.Id(), err)
	}

	return nil
}

func resourceTFETeamAccessDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Delete team access: %s", d.Id())
	err := config.ClientV2.API.TeamWorkspaces().ByTeam_workspace_id(d.Id()).Delete(ctx, nil)
	if err != nil {
		if errors.Is(err, tfev2.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("Error deleting team access %s: %w", d.Id(), err)
	}

	return nil
}

func resourceTFETeamAccessImporter(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	config := meta.(ConfiguredClient)

	s := strings.SplitN(d.Id(), "/", 3)
	if len(s) != 3 {
		return nil, fmt.Errorf(
			"invalid team access import format: %s (expected <ORGANIZATION>/<WORKSPACE>/<TEAM ACCESS ID>)",
			d.Id(),
		)
	}

	// Set the fields that are part of the import ID.
	workspaceID, err := fetchWorkspaceExternalIDV2(s[0]+"/"+s[1], config.ClientV2)
	if err != nil {
		return nil, fmt.Errorf(
			"error retrieving workspace %s from organization %s: %w", s[1], s[0], err)
	}
	d.Set(teamAccessWorkspaceIDKey, workspaceID)
	d.SetId(s[2])

	return []*schema.ResourceData{d}, nil
}

// The Team Access API and behavior for 'custom' access is very hard for the current SDK to model.
//
//   - Schema validations are limited to the single attribute they are defined on; you cannot validate something with the
//     additional context of another attribute's value in the resource.
//   - The SDK cannot discern between something defined only in state, or only in configuration. Some assumptions can be
//     made (and are made in these changes) via GetChange(), but it's hacky at best.
//
// This CustomizeDiff function is what allows the provider resource to model the right API behavior with these
// limitations, rooting out the user's intentions to figure out when to automatically assign 'access' to custom and/or
// recompute 'permissions'.
func setCustomOrComputedPermissions(_ context.Context, d *schema.ResourceDiff, meta interface{}) error {
	// Check if permissions block is present in new config
	_, hasPermissionsInNew := d.GetOk(teamAccessPermissionsKey)

	// Check if access is in the raw config as state may have both access and permissions
	rawConfig := d.GetRawConfig()
	accessInConfig := !rawConfig.GetAttr(teamAccessAccessKey).IsNull()

	// User removed 'access' from config and has 'permissions' block in new config
	if !accessInConfig && hasPermissionsInNew {
		if err := setCustomAccess(d); err != nil {
			return err
		}
		return nil
	}

	if _, ok := d.GetOk(teamAccessAccessKey); ok {
		if d.HasChange(teamAccessAccessKey) {
			// If access is being added or changed to a known value, all permissions
			// will be read-only and computed by the API (access is never marked as 'custom' in the
			// configuration).
			d.SetNewComputed(teamAccessPermissionsKey)
		} else if d.HasChange(fmt.Sprintf("%s.0", teamAccessPermissionsKey)) {
			// If access is present, not being explicitly changed, but permissions are being
			// changed, the user might be switching from using a fixed access level
			// (read/plan/write/admin) to a permissions block ('custom' access).
			// Set the access to custom.
			if err := setCustomAccess(d); err != nil {
				return err
			}
		}
	} else if !d.NewValueKnown(teamAccessAccessKey) {
		if d.Id() != "" {
			// If the value for access isn't known on an existing resource, the user must have set the
			// access attribute to an interpolated value not known at plan time.
			// Set permissions as computed.
			d.SetNewComputed(teamAccessPermissionsKey)
		} else if _, ok := d.GetOk(teamAccessPermissionsKey); ok {
			// If the resource is new, the value for access isn't known, and permissions are
			// present, the user must be creating a new resource with custom access.
			// Set access to custom.
			if err := setCustomAccess(d); err != nil {
				return err
			}
		}
	}

	return nil
}

func setCustomAccess(d *schema.ResourceDiff) error {
	// If a change in permissions contains a value not known at plan time, error.
	// Interpolated values not known at plan time are not allowed because we cannot re-check
	// for a change in permissions later - when the plan is expanded for new values learned during
	// an apply. This creates an inconsistent final plan and causes an error.
	for _, permission := range []string{
		"permissions.0.runs",
		"permissions.0.variables",
		"permissions.0.state_versions",
		"permissions.0.sentinel_mocks",
		"permissions.0.workspace_locking",
		"permissions.0.run_tasks",
	} {
		if !d.NewValueKnown(permission) {
			return fmt.Errorf("'%q' cannot be derived from a value that is unknown during planning", permission)
		}
	}

	d.SetNew("access", tfe.AccessCustom)

	return nil
}
