// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

// NOTE: This is a legacy resource and should be migrated to the Plugin
// Framework if substantial modifications are planned. See
// docs/new-resources.md if planning to use this code as boilerplate for
// a new resource.

package provider

import (
	"context"
	"fmt"
	"log"
	"strings"

	tfe "github.com/hashicorp/go-tfe"
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
)

func resourceTFETeamAccess() *schema.Resource {
	return &schema.Resource{
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
			teamAccessAccessKey: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
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
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						permissionsRunsKey: {
							Type:     schema.TypeString,
							Required: true,
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
							Type:     schema.TypeString,
							Required: true,
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
							Type:     schema.TypeString,
							Required: true,
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
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice(
								[]string{
									string(tfe.SentinelMocksPermissionNone),
									string(tfe.SentinelMocksPermissionRead),
								},
								false,
							),
						},

						permissionsWorkspaceLockingKey: {
							Type:     schema.TypeBool,
							Required: true,
						},

						permissionsRunTasksKey: {
							Type:     schema.TypeBool,
							Required: true,
						},
					},
				},
			},

			teamAccessTeamIDKey: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			teamAccessWorkspaceIDKey: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringMatch(
					workspaceIDRegexp,
					"must be a valid workspace ID (ws-<RANDOM STRING>)",
				),
			},
		},
	}
}

func resourceTFETeamAccessCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	// Get the access level
	access := d.Get(teamAccessAccessKey).(string)

	// Get the workspace
	workspaceID := d.Get(teamAccessWorkspaceIDKey).(string)
	ws, err := config.Client.Workspaces.ReadByID(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf(
			"Error retrieving workspace %s: %w", workspaceID, err)
	}

	// Get the team.
	teamID := d.Get(teamAccessTeamIDKey).(string)
	tm, err := config.Client.Teams.Read(ctx, teamID)
	if err != nil {
		return fmt.Errorf("Error retrieving team %s: %w", teamID, err)
	}

	// Create a new options struct.
	options := tfe.TeamAccessAddOptions{
		Access:    tfe.Access(tfe.AccessType(access)),
		Team:      tm,
		Workspace: ws,
	}

	permissionsRunsPath := fmt.Sprintf("%s.0.%s", teamAccessPermissionsKey, permissionsRunsKey)
	if d.HasChange(permissionsRunsPath) {
		if v, ok := d.GetOk(permissionsRunsPath); ok {
			options.Runs = tfe.RunsPermission(tfe.RunsPermissionType(v.(string)))
		}
	}

	permissionsVariablesPath := fmt.Sprintf("%s.0.%s", teamAccessPermissionsKey, permissionsVariablesKey)
	if d.HasChange(permissionsVariablesPath) {
		if v, ok := d.GetOk(permissionsVariablesPath); ok {
			options.Variables = tfe.VariablesPermission(tfe.VariablesPermissionType(v.(string)))
		}
	}

	permissionsStateVersionsPath := fmt.Sprintf("%s.0.%s", teamAccessPermissionsKey, permissionsStateVersionsKey)
	if d.HasChange(permissionsStateVersionsPath) {
		if v, ok := d.GetOk(permissionsStateVersionsPath); ok {
			options.StateVersions = tfe.StateVersionsPermission(tfe.StateVersionsPermissionType(v.(string)))
		}
	}

	permissionsSentinelMocksPath := fmt.Sprintf("%s.0.%s", teamAccessPermissionsKey, permissionsSentinelMocksKey)
	if d.HasChange(permissionsSentinelMocksPath) {
		if v, ok := d.GetOk(permissionsSentinelMocksPath); ok {
			options.SentinelMocks = tfe.SentinelMocksPermission(tfe.SentinelMocksPermissionType(v.(string)))
		}
	}

	permissionsWorkspaceLockingPath := fmt.Sprintf("%s.0.%s", teamAccessPermissionsKey, permissionsWorkspaceLockingKey)
	if d.HasChange(permissionsWorkspaceLockingPath) {
		if v, ok := d.GetOkExists(permissionsWorkspaceLockingPath); ok {
			options.WorkspaceLocking = tfe.Bool(v.(bool))
		}
	}

	permissionsRunTasksPath := fmt.Sprintf("%s.0.%s", teamAccessPermissionsKey, permissionsRunTasksKey)
	if d.HasChange(permissionsRunTasksPath) {
		if v, ok := d.GetOkExists(permissionsRunTasksPath); ok {
			options.RunTasks = tfe.Bool(v.(bool))
		}
	}

	log.Printf("[DEBUG] Give team %s %s access to workspace: %s", tm.Name, access, ws.Name)
	tmAccess, err := config.Client.TeamAccess.Add(ctx, options)
	if err != nil {
		return fmt.Errorf(
			"Error giving team %s %s access to workspace %s: %w", tm.Name, access, ws.Name, err)
	}

	d.SetId(tmAccess.ID)

	return resourceTFETeamAccessRead(d, meta)
}

func resourceTFETeamAccessRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Read configuration of team access: %s", d.Id())
	tmAccess, err := config.Client.TeamAccess.Read(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			log.Printf("[DEBUG] Team access %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading configuration of team access %s: %w", d.Id(), err)
	}

	// Update config.
	d.Set(teamAccessAccessKey, string(tmAccess.Access))
	permissions := []map[string]interface{}{{
		permissionsRunsKey:             tmAccess.Runs,
		permissionsVariablesKey:        tmAccess.Variables,
		permissionsStateVersionsKey:    tmAccess.StateVersions,
		permissionsSentinelMocksKey:    tmAccess.SentinelMocks,
		permissionsWorkspaceLockingKey: tmAccess.WorkspaceLocking,
		permissionsRunTasksKey:         tmAccess.RunTasks,
	}}
	if err := d.Set(teamAccessPermissionsKey, permissions); err != nil {
		return fmt.Errorf("error setting permissions for team access %s: %w", d.Id(), err)
	}

	if tmAccess.Team != nil {
		d.Set(teamAccessTeamIDKey, tmAccess.Team.ID)
	} else {
		d.Set(teamAccessTeamIDKey, "")
	}

	return nil
}

func resourceTFETeamAccessUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	// create an options struct
	options := tfe.TeamAccessUpdateOptions{}

	// Set access level
	access := d.Get(teamAccessAccessKey).(string)
	options.Access = tfe.Access(tfe.AccessType(access))

	permissionsRunsPath := fmt.Sprintf("%s.0.%s", teamAccessPermissionsKey, permissionsRunsKey)
	if d.HasChange(permissionsRunsPath) {
		if v, ok := d.GetOk(permissionsRunsPath); ok {
			options.Runs = tfe.RunsPermission(tfe.RunsPermissionType(v.(string)))
		}
	}

	permissionsVariablesPath := fmt.Sprintf("%s.0.%s", teamAccessPermissionsKey, permissionsVariablesKey)
	if d.HasChange(permissionsVariablesPath) {
		if v, ok := d.GetOk(permissionsVariablesPath); ok {
			options.Variables = tfe.VariablesPermission(tfe.VariablesPermissionType(v.(string)))
		}
	}

	permissionsStateVersionsPath := fmt.Sprintf("%s.0.%s", teamAccessPermissionsKey, permissionsStateVersionsKey)
	if d.HasChange(permissionsStateVersionsPath) {
		if v, ok := d.GetOk(permissionsStateVersionsPath); ok {
			options.StateVersions = tfe.StateVersionsPermission(tfe.StateVersionsPermissionType(v.(string)))
		}
	}

	permissionsSentinelMocksPath := fmt.Sprintf("%s.0.%s", teamAccessPermissionsKey, permissionsSentinelMocksKey)
	if d.HasChange(permissionsSentinelMocksPath) {
		if v, ok := d.GetOk(permissionsSentinelMocksPath); ok {
			options.SentinelMocks = tfe.SentinelMocksPermission(tfe.SentinelMocksPermissionType(v.(string)))
		}
	}

	permissionsWorkspaceLockingPath := fmt.Sprintf("%s.0.%s", teamAccessPermissionsKey, permissionsWorkspaceLockingKey)
	if d.HasChange(permissionsWorkspaceLockingPath) {
		if v, ok := d.GetOkExists(permissionsWorkspaceLockingPath); ok {
			options.WorkspaceLocking = tfe.Bool(v.(bool))
		}
	}

	permissionsRunTasksPath := fmt.Sprintf("%s.0.%s", teamAccessPermissionsKey, permissionsRunTasksKey)
	if d.HasChange(permissionsRunTasksPath) {
		if v, ok := d.GetOkExists(permissionsRunTasksPath); ok {
			options.RunTasks = tfe.Bool(v.(bool))
		}
	}

	log.Printf("[DEBUG] Update team access: %s", d.Id())
	tmAccess, err := config.Client.TeamAccess.Update(ctx, d.Id(), options)
	if err != nil {
		return fmt.Errorf(
			"Error updating team access %s: %w", d.Id(), err)
	}

	// Update permissions, in the case that they were marked to be recomputed.
	permissions := []map[string]interface{}{{
		permissionsRunsKey:             tmAccess.Runs,
		permissionsVariablesKey:        tmAccess.Variables,
		permissionsStateVersionsKey:    tmAccess.StateVersions,
		permissionsSentinelMocksKey:    tmAccess.SentinelMocks,
		permissionsWorkspaceLockingKey: tmAccess.WorkspaceLocking,
		permissionsRunTasksKey:         tmAccess.RunTasks,
	}}
	if err := d.Set(teamAccessPermissionsKey, permissions); err != nil {
		return fmt.Errorf("error setting permissions for team access %s: %w", d.Id(), err)
	}

	return nil
}

func resourceTFETeamAccessDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Delete team access: %s", d.Id())
	err := config.Client.TeamAccess.Remove(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
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
	workspaceID, err := fetchWorkspaceExternalID(s[0]+"/"+s[1], config.Client)
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
	// Check if access is being removed from config (exists in old but not in new config)
	oldAccess, _ := d.GetChange(teamAccessAccessKey)
	accessRemovedFromConfig := oldAccess != nil && oldAccess.(string) != "" && !d.NewValueKnown(teamAccessAccessKey)

	// If access was removed from config and permissions block is present, set to custom
	if accessRemovedFromConfig {
		if _, ok := d.GetOk(teamAccessPermissionsKey); ok {
			if err := setCustomAccess(d); err != nil {
				return err
			}
			return nil
		}
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
