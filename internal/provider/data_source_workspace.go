// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

// NOTE: This is a legacy resource and should be migrated to the Plugin
// Framework if substantial modifications are planned. See
// docs/new-resources.md if planning to use this code as boilerplate for
// a new resource.

package provider

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-tfe/internal/provider/helpers"
)

func dataSourceTFEWorkspace() *schema.Resource {
	return &schema.Resource{
		Description: "Gets information about a workspace.\n\n" +
			"~> **Note:** Using `global_remote_state` or `remote_state_consumer_ids` requires using the provider with HCP Terraform or an instance of Terraform Enterprise at least as recent as v202104-1.",

		Read: dataSourceTFEWorkspaceRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The workspace ID.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"name": {
				Description: "Name of the workspace.",
				Type:        schema.TypeString,
				Required:    true,
			},

			"organization": {
				Description: "Name of the organization.",
				Type:        schema.TypeString,
				Optional:    true,
			},

			"description": {
				Description: "Description of the workspace.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"allow_destroy_plan": {
				Description: "Indicates whether destroy plans can be queued on the workspace.",
				Type:        schema.TypeBool,
				Computed:    true,
			},

			"auto_apply": {
				Description: "Indicates whether to automatically apply changes when a Terraform plan is successful.",
				Type:        schema.TypeBool,
				Computed:    true,
			},

			"auto_apply_run_trigger": {
				Description: "Whether the workspace will automatically apply changes for runs that were created by run triggers from another workspace.",
				Type:        schema.TypeBool,
				Computed:    true,
			},

			"auto_destroy_at": {
				Description: "Future date/time string at which point all resources in a workspace will be scheduled to be deleted.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"auto_destroy_activity_duration": {
				Description: "A duration string representing time after workspace activity when an auto-destroy run will be triggered.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"inherits_project_auto_destroy": {
				Description: "Indicates whether this workspace inherits project auto destroy settings.",
				Type:        schema.TypeBool,
				Computed:    true,
			},

			"file_triggers_enabled": {
				Description: "Indicates whether runs are triggered based on the changed files in a VCS push (if `true`) or always triggered on every push (if `false`).",
				Type:        schema.TypeBool,
				Computed:    true,
			},

			"global_remote_state": {
				Description: "Whether the workspace should allow all workspaces in the organization to access its state data during runs. If false, then only specifically approved workspaces can access its state (determined by the `remote_state_consumer_ids` argument). Cannot be true if `project_remote_state` is true.",
				Type:        schema.TypeBool,
				Computed:    true,
			},

			"project_remote_state": {
				Description: "Whether the workspace should allow all workspaces in the project to access its state data during runs. If false, then only specifically approved workspaces can access its state (determined by the `remote_state_consumer_ids` argument). Cannot be true if `global_remote_state` is true.",
				Type:        schema.TypeBool,
				Computed:    true,
			},

			"remote_state_consumer_ids": {
				Description: "A set of workspace IDs that will be set as the remote state consumers for the given workspace. Cannot be used if `global_remote_state` or `project_remote_state` is set to `true`.",
				Type:        schema.TypeSet,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},

			"assessments_enabled": {
				Description: "(Available only in HCP Terraform) Indicates whether health assessments such as drift detection are enabled for the workspace.",
				Type:        schema.TypeBool,
				Computed:    true,
			},

			"operations": {
				Description: "Indicates whether the workspace is using remote execution mode. Set to `false` to switch execution mode to local. `true` by default.",
				Type:        schema.TypeBool,
				Computed:    true,
			},

			"policy_check_failures": {
				Description: "The number of policy check failures from the latest run.",
				Type:        schema.TypeInt,
				Computed:    true,
			},

			"project_id": {
				Description: "ID of the workspace's project.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"queue_all_runs": {
				Description: "Indicates whether the workspace will automatically perform runs in response to webhooks immediately after its creation. If `false`, an initial run must be manually queued to enable future automatic runs.",
				Type:        schema.TypeBool,
				Computed:    true,
			},

			"resource_count": {
				Description: "The number of resources managed by the workspace.",
				Type:        schema.TypeInt,
				Computed:    true,
			},

			"run_failures": {
				Description: "The number of run failures on the workspace.",
				Type:        schema.TypeInt,
				Computed:    true,
			},

			"runs_count": {
				Description: "The number of runs on the workspace.",
				Type:        schema.TypeInt,
				Computed:    true,
			},

			"source_name": {
				Description: "The name of the workspace creation source, if set.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"source_url": {
				Description: "The URL of the workspace creation source, if set.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"speculative_enabled": {
				Description: "Indicates whether this workspace allows speculative plans.",
				Type:        schema.TypeBool,
				Computed:    true,
			},

			"ssh_key_id": {
				Description: "The ID of an SSH key assigned to the workspace.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"structured_run_output_enabled": {
				Description: "Indicates whether runs in this workspace use the enhanced apply UI.",
				Type:        schema.TypeBool,
				Computed:    true,
			},

			"effective_tags": {
				Description: "A map of key-value tags associated with the workspace, including any inherited tags from the parent project.",
				Type:        schema.TypeMap,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},

			"tag_names": {
				Description: "The names of tags added to this workspace.",
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},

			"terraform_version": {
				Description: "The version (or version constraint) of Terraform used for this workspace.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"trigger_prefixes": {
				Description: "List of trigger prefixes that describe the paths HCP Terraform monitors for changes, in addition to the working directory. Trigger prefixes are always appended to the root directory of the repository.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},

			"trigger_patterns": {
				Description: "List of [glob patterns](https://developer.hashicorp.com/terraform/cloud-docs/workspaces/settings/vcs#glob-patterns-for-automatic-run-triggering) that describe the files HCP Terraform monitors for changes. Trigger patterns are always appended to the root directory of the repository.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},

			"working_directory": {
				Description: "A relative path that Terraform will execute within.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"execution_mode": {
				Description: "Indicates the [execution mode](https://developer.hashicorp.com/terraform/cloud-docs/workspaces/settings#execution-mode) of the workspace. **Note:** This value might be derived from an organization-level default or set on the workspace itself; see the [`tfe_workspace_settings` resource](tfe_workspace_settings) for details.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"vcs_repo": {
				Description: "Settings for the workspace's VCS repository.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"identifier": {
							Description: "A reference to your VCS repository in the format `<vcs organization>/<repository>` where `<vcs organization>` and `<repository>` refer to the organization and repository in your VCS provider.",
							Type:        schema.TypeString,
							Computed:    true,
						},

						"branch": {
							Description: "The repository branch that Terraform will execute from.",
							Type:        schema.TypeString,
							Computed:    true,
						},

						"ingress_submodules": {
							Description: "Indicates whether submodules should be fetched when cloning the VCS repository.",
							Type:        schema.TypeBool,
							Computed:    true,
						},

						"oauth_token_id": {
							Description: "OAuth token ID of the configured VCS connection.",
							Type:        schema.TypeString,
							Computed:    true,
						},

						"tags_regex": {
							Description: "A regular expression used to trigger a Workspace run for matching Git tags.",
							Type:        schema.TypeString,
							Computed:    true,
						},

						"github_app_installation_id": {
							Description: "The installation ID of the GitHub App.",
							Type:        schema.TypeString,
							Computed:    true,
						},
					},
				},
			},
			"html_url": {
				Description: "The URL to the browsable HTML overview of the workspace.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"hyok_enabled": {
				Description: "Whether HYOK is enabled for the workspace.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"locked": {
				Description: "Indicates whether the workspace is locked.",
				Type:        schema.TypeBool,
				Computed:    true,
			},

			"created_at": {
				Description: "The time when the workspace was created.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"updated_at": {
				Description: "The time when the workspace was last updated.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"environment": {
				Description: "The environment of the workspace.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"apply_duration_average": {
				Description: "The average duration of applies for this workspace.",
				Type:        schema.TypeInt,
				Computed:    true,
			},

			"plan_duration_average": {
				Description: "The average duration of plans for this workspace.",
				Type:        schema.TypeInt,
				Computed:    true,
			},

			"source": {
				Description: "The source of the workspace.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"setting_overwrites": {
				Description: "Settings that are overwritten for this workspace. Contains: `is-destroyable` - Whether the workspace can be destroyed.", // On migration, ideally reformat into non-inline descriptions
				Type:        schema.TypeMap,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeBool},
			},

			"permissions": {
				Description: "The permissions for the current user on this workspace. Contains: `can-update` - Can update the workspace. `can-destroy` - Can destroy the workspace. `can-queue-run` - Can queue runs. `can-queue-apply` - Can queue apply. `can-queue-destroy` - Can queue destroy. `can-lock` - Can lock the workspace. `can-unlock` - Can unlock the workspace. `can-force-unlock` - Can force unlock the workspace. `can-read-settings` - Can read workspace settings. `can-update-variable` - Can update variables. `can-manage-run-tasks` - Can manage run tasks. `can-force-delete` - Can force delete the workspace.", // On migration, ideally reformat into non-inline descriptions
				Type:        schema.TypeMap,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeBool},
			},

			"actions": {
				Description: "Actions that can be performed on this workspace. Contains: `execution-mode` - Whether execution mode is overwritten at the workspace level. `agent-pool` - Whether agent pool is overwritten at the workspace level.", // On migration, ideally reformat into non-inline descriptions
				Type:        schema.TypeMap,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeBool},
			},
		},
	}
}

func fallbackWorkspaceRead(config ConfiguredClient, organization, name string) (*tfe.Workspace, error) {
	log.Printf("[DEBUG] Workspace %s read failed due to unsupported Include; retrying without it", name)
	workspace, err := config.Client.Workspaces.Read(ctx, organization, name)
	if err != nil && errors.Is(err, tfe.ErrResourceNotFound) {
		return nil, fmt.Errorf("could not find workspace %s/%s", organization, name)
	} else if err != nil {
		return nil, fmt.Errorf("error reading workspace %s without include: %w", name, err)
	}

	return workspace, err
}

func dataSourceTFEWorkspaceRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	// Get the name and organization.
	name := d.Get("name").(string)
	organization, err := config.schemaOrDefaultOrganization(d)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Read configuration of workspace: %s", name)
	workspace, err := config.Client.Workspaces.ReadWithOptions(ctx, organization, name, &tfe.WorkspaceReadOptions{
		Include: []tfe.WSIncludeOpt{tfe.WSEffectiveTagBindings},
	})
	if err != nil && errors.Is(err, tfe.ErrResourceNotFound) {
		return fmt.Errorf("could not find workspace %s/%s", organization, name)
	}
	if err != nil && errors.Is(err, tfe.ErrInvalidIncludeValue) {
		workspace, err = fallbackWorkspaceRead(config, organization, name)
		if err != nil {
			return err
		}
	}
	if err != nil {
		return fmt.Errorf("Error retrieving workspace: %w", err)
	}
	// Update the config.
	d.Set("allow_destroy_plan", workspace.AllowDestroyPlan)
	d.Set("auto_apply", workspace.AutoApply)
	d.Set("auto_apply_run_trigger", workspace.AutoApplyRunTrigger)
	d.Set("description", workspace.Description)
	d.Set("assessments_enabled", workspace.AssessmentsEnabled)
	d.Set("file_triggers_enabled", workspace.FileTriggersEnabled)
	d.Set("operations", workspace.Operations)
	d.Set("policy_check_failures", workspace.PolicyCheckFailures)
	d.Set("inherits_project_auto_destroy", workspace.InheritsProjectAutoDestroy)

	autoDestroyAt, err := flattenAutoDestroyAt(workspace.AutoDestroyAt)
	if err != nil {
		return fmt.Errorf("Error flattening auto destroy during read: %w", err)
	}
	d.Set("auto_destroy_at", autoDestroyAt)

	var autoDestroyDuration string
	if workspace.AutoDestroyActivityDuration.IsSpecified() {
		autoDestroyDuration, err = workspace.AutoDestroyActivityDuration.Get()
		if err != nil {
			return fmt.Errorf("Error reading auto destroy activity duration: %w", err)
		}
	}
	d.Set("auto_destroy_activity_duration", autoDestroyDuration)

	// If target tfe instance predates projects, then workspace.Project will be nil
	if workspace.Project != nil {
		d.Set("project_id", workspace.Project.ID)
	}

	d.Set("queue_all_runs", workspace.QueueAllRuns)
	d.Set("resource_count", workspace.ResourceCount)
	d.Set("run_failures", workspace.RunFailures)
	d.Set("runs_count", workspace.RunsCount)
	d.Set("source_name", workspace.SourceName)
	d.Set("source_url", workspace.SourceURL)
	d.Set("speculative_enabled", workspace.SpeculativeEnabled)
	d.Set("structured_run_output_enabled", workspace.StructuredRunOutputEnabled)
	d.Set("terraform_version", workspace.TerraformVersion)
	d.Set("trigger_prefixes", workspace.TriggerPrefixes)
	d.Set("trigger_patterns", workspace.TriggerPatterns)
	d.Set("working_directory", workspace.WorkingDirectory)
	d.Set("execution_mode", workspace.ExecutionMode)
	d.Set("hyok_enabled", workspace.HYOKEnabled)
	d.Set("locked", workspace.Locked)
	d.Set("created_at", workspace.CreatedAt.Format(time.RFC3339))
	d.Set("updated_at", workspace.UpdatedAt.Format(time.RFC3339))
	d.Set("environment", workspace.Environment)
	d.Set("source", string(workspace.Source))
	d.Set("apply_duration_average", int(workspace.ApplyDurationAverage.Milliseconds()))
	d.Set("plan_duration_average", int(workspace.PlanDurationAverage.Milliseconds()))

	// Set setting overwrites
	if workspace.SettingOverwrites != nil {
		settingOverwrites := make(map[string]interface{})
		if workspace.SettingOverwrites.ExecutionMode != nil {
			settingOverwrites["execution-mode"] = *workspace.SettingOverwrites.ExecutionMode
		}
		if workspace.SettingOverwrites.AgentPool != nil {
			settingOverwrites["agent-pool"] = *workspace.SettingOverwrites.AgentPool
		}
		d.Set("setting_overwrites", settingOverwrites)
	}

	// Set permissions
	if workspace.Permissions != nil {
		permissions := map[string]interface{}{
			"can-update":           workspace.Permissions.CanUpdate,
			"can-destroy":          workspace.Permissions.CanDestroy,
			"can-queue-run":        workspace.Permissions.CanQueueRun,
			"can-queue-apply":      workspace.Permissions.CanQueueApply,
			"can-queue-destroy":    workspace.Permissions.CanQueueDestroy,
			"can-lock":             workspace.Permissions.CanLock,
			"can-unlock":           workspace.Permissions.CanUnlock,
			"can-force-unlock":     workspace.Permissions.CanForceUnlock,
			"can-read-settings":    workspace.Permissions.CanReadSettings,
			"can-update-variable":  workspace.Permissions.CanUpdateVariable,
			"can-manage-run-tasks": workspace.Permissions.CanManageRunTasks,
		}
		if workspace.Permissions.CanForceDelete != nil {
			permissions["can-force-delete"] = *workspace.Permissions.CanForceDelete
		}
		d.Set("permissions", permissions)
	}

	// Set actions
	if workspace.Actions != nil {
		actions := map[string]interface{}{
			"is-destroyable": workspace.Actions.IsDestroyable,
		}
		d.Set("actions", actions)
	}

	if workspace.Links["self-html"] != nil {
		baseAPI := config.Client.BaseURL()
		htmlURL := url.URL{
			Scheme: baseAPI.Scheme,
			Host:   baseAPI.Host,
			Path:   workspace.Links["self-html"].(string),
		}

		d.Set("html_url", htmlURL.String())
	}

	// Set remote_state_consumer_ids if global_remote_state and project_remote_state are false
	globalRemoteState := workspace.GlobalRemoteState
	projectRemoteState := workspace.ProjectRemoteState
	if globalRemoteState || projectRemoteState {
		if err := d.Set("remote_state_consumer_ids", []string{}); err != nil {
			return err
		}
	} else {
		legacyGlobalState, remoteStateConsumerIDs, err := readWorkspaceStateConsumers(workspace.ID, config.Client)

		if err != nil {
			return fmt.Errorf(
				"Error reading remote state consumers for workspace %s: %w", workspace.ID, err)
		}

		if legacyGlobalState {
			globalRemoteState = true
		}
		d.Set("remote_state_consumer_ids", remoteStateConsumerIDs)
	}
	d.Set("global_remote_state", globalRemoteState)
	d.Set("project_remote_state", projectRemoteState)

	if workspace.SSHKey != nil {
		d.Set("ssh_key_id", workspace.SSHKey.ID)
	}

	tagInfo := helpers.NewTagInfo(nil, workspace.EffectiveTagBindings, false)
	d.Set("effective_tags", tagInfo.EffectiveTags)

	// Update the tag names
	var tagNames []interface{}
	for _, tagName := range workspace.TagNames {
		tagNames = append(tagNames, tagName)
	}
	d.Set("tag_names", tagNames)

	var vcsRepo []interface{}
	if workspace.VCSRepo != nil {
		vcsConfig := map[string]interface{}{
			"identifier":                 workspace.VCSRepo.Identifier,
			"branch":                     workspace.VCSRepo.Branch,
			"ingress_submodules":         workspace.VCSRepo.IngressSubmodules,
			"oauth_token_id":             workspace.VCSRepo.OAuthTokenID,
			"tags_regex":                 workspace.VCSRepo.TagsRegex,
			"github_app_installation_id": workspace.VCSRepo.GHAInstallationID,
		}
		vcsRepo = append(vcsRepo, vcsConfig)
	}
	d.Set("vcs_repo", vcsRepo)

	d.SetId(workspace.ID)

	return nil
}
