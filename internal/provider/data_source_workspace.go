// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-tfe/internal/provider/helpers"
)

func dataSourceTFEWorkspace() *schema.Resource {
	return &schema.Resource{
		Description: "Gets information about a workspace. Note that using `global_remote_state` or `remote_state_consumer_ids` requires using the provider with HCP Terraform or an instance of Terraform Enterprise at least as recent as v202104-1.",

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
							Type:     schema.TypeString,
							Computed: true,
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

func fallbackWorkspaceRead(ctx context.Context, config ConfiguredClient, organization, name string) (*tfe.Workspace, error) {
	tflog.Debug(ctx, "Workspace read failed due to unsupported include; retrying without include", map[string]any{"workspace_name": name})
	workspace, err := config.Client.Workspaces.Read(ctx, organization, name)
	if err != nil && errors.Is(err, tfe.ErrResourceNotFound) {
		return nil, fmt.Errorf("could not find workspace %s/%s", organization, name)
	} else if err != nil {
		return nil, fmt.Errorf("error reading workspace %s without include: %w", name, err)
	}

	return workspace, err
}

func (d *dataSourceTFEWorkspace) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(ConfiguredClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected tfe.ConfiguredClient, got %T. This is a bug in the tfe provider, so please report it on GitHub.", req.ProviderData),
		)
		return
	}

	d.config = client
}

func (d *dataSourceTFEWorkspace) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var configModel modelDataSourceTFEWorkspace
	resp.Diagnostics.Append(req.Config.Get(ctx, &configModel)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var organization string
	resp.Diagnostics.Append(d.config.dataOrDefaultOrganization(ctx, req.Config, &organization)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := configModel.Name.ValueString()
	tflog.Debug(ctx, "Reading workspace", map[string]any{"workspace_name": name, "organization": organization})

	workspace, err := d.config.Client.Workspaces.ReadWithOptions(ctx, organization, name, &tfe.WorkspaceReadOptions{
		Include: []tfe.WSIncludeOpt{tfe.WSEffectiveTagBindings},
	})
	if err != nil && errors.Is(err, tfe.ErrResourceNotFound) {
		resp.Diagnostics.AddError("Could not find workspace", fmt.Sprintf("Workspace %s/%s not found", organization, name))
		return
	}
	if err != nil && errors.Is(err, tfe.ErrInvalidIncludeValue) {
		workspace, err = fallbackWorkspaceRead(ctx, d.config, organization, name)
		if err != nil {
			resp.Diagnostics.AddError("Error retrieving workspace", err.Error())
			return
		}
	}
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving workspace", err.Error())
		return
	}

	autoDestroyAt := ""
	flattenedAutoDestroyAt, err := flattenAutoDestroyAt(workspace.AutoDestroyAt)
	if err != nil {
		resp.Diagnostics.AddError("Error flattening auto destroy timestamp", err.Error())
		return
	}
	if flattenedAutoDestroyAt != nil {
		autoDestroyAt = *flattenedAutoDestroyAt
	}

	autoDestroyDuration := ""
	if workspace.AutoDestroyActivityDuration.IsSpecified() {
		autoDestroyDuration, err = workspace.AutoDestroyActivityDuration.Get()
		if err != nil {
			resp.Diagnostics.AddError("Error reading auto destroy activity duration", err.Error())
			return
		}
	}

	htmlURL := ""
	if workspace.Links["self-html"] != nil {
		baseAPI := d.config.Client.BaseURL()
		href := url.URL{
			Scheme: baseAPI.Scheme,
			Host:   baseAPI.Host,
			Path:   workspace.Links["self-html"].(string),
		}
		htmlURL = href.String()
	}

	globalRemoteState := workspace.GlobalRemoteState
	projectRemoteState := workspace.ProjectRemoteState
	remoteStateConsumerIDs := []string{}

	if !globalRemoteState && !projectRemoteState {
		legacyGlobalState, consumers, err := readWorkspaceStateConsumers(workspace.ID, d.config.Client)
		if err != nil {
			resp.Diagnostics.AddError("Error reading remote state consumers", fmt.Sprintf("Error reading remote state consumers for workspace %s: %s", workspace.ID, err.Error()))
			return
		}

		if legacyGlobalState {
			globalRemoteState = true
		}
		remoteStateConsumerIDs = consumers
	}

	remoteStateConsumersValue, diags := types.SetValueFrom(ctx, types.StringType, remoteStateConsumerIDs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tagInfo := helpers.NewTagInfo(nil, workspace.EffectiveTagBindings, false)
	effectiveTags := map[string]attr.Value{}
	for key, value := range tagInfo.EffectiveTags {
		effectiveTags[key] = types.StringValue(value.(string))
	}
	effectiveTagsValue := types.MapValueMust(types.StringType, effectiveTags)

	tagNamesValue, diags := types.SetValueFrom(ctx, types.StringType, workspace.TagNames)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	triggerPrefixesValue, diags := types.ListValueFrom(ctx, types.StringType, workspace.TriggerPrefixes)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	triggerPatternsValue, diags := types.ListValueFrom(ctx, types.StringType, workspace.TriggerPatterns)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	settingOverwritesValue := types.MapNull(types.BoolType)
	if workspace.SettingOverwrites != nil {
		settingOverwrites := map[string]attr.Value{}
		if workspace.SettingOverwrites.ExecutionMode != nil {
			settingOverwrites["execution-mode"] = types.BoolValue(*workspace.SettingOverwrites.ExecutionMode)
		}
		if workspace.SettingOverwrites.AgentPool != nil {
			settingOverwrites["agent-pool"] = types.BoolValue(*workspace.SettingOverwrites.AgentPool)
		}
		settingOverwritesValue = types.MapValueMust(types.BoolType, settingOverwrites)
	}

	permissionsValue := types.MapNull(types.BoolType)
	if workspace.Permissions != nil {
		permissions := map[string]attr.Value{
			"can-update":           types.BoolValue(workspace.Permissions.CanUpdate),
			"can-destroy":          types.BoolValue(workspace.Permissions.CanDestroy),
			"can-queue-run":        types.BoolValue(workspace.Permissions.CanQueueRun),
			"can-queue-apply":      types.BoolValue(workspace.Permissions.CanQueueApply),
			"can-queue-destroy":    types.BoolValue(workspace.Permissions.CanQueueDestroy),
			"can-lock":             types.BoolValue(workspace.Permissions.CanLock),
			"can-unlock":           types.BoolValue(workspace.Permissions.CanUnlock),
			"can-force-unlock":     types.BoolValue(workspace.Permissions.CanForceUnlock),
			"can-read-settings":    types.BoolValue(workspace.Permissions.CanReadSettings),
			"can-update-variable":  types.BoolValue(workspace.Permissions.CanUpdateVariable),
			"can-manage-run-tasks": types.BoolValue(workspace.Permissions.CanManageRunTasks),
		}
		if workspace.Permissions.CanForceDelete != nil {
			permissions["can-force-delete"] = types.BoolValue(*workspace.Permissions.CanForceDelete)
		}
		permissionsValue = types.MapValueMust(types.BoolType, permissions)
	}

	actionsValue := types.MapNull(types.BoolType)
	if workspace.Actions != nil {
		actionsValue = types.MapValueMust(types.BoolType, map[string]attr.Value{
			"is-destroyable": types.BoolValue(workspace.Actions.IsDestroyable),
		})
	}

	projectID := ""
	if workspace.Project != nil {
		projectID = workspace.Project.ID
	}

	sshKeyID := ""
	if workspace.SSHKey != nil {
		sshKeyID = workspace.SSHKey.ID
	}

	hyokEnabled := types.BoolNull()
	if workspace.HYOKEnabled != nil {
		hyokEnabled = types.BoolValue(*workspace.HYOKEnabled)
	}

	vcsRepo := []modelDataSourceTFEWorkspaceVCSRepo{}
	if workspace.VCSRepo != nil {
		vcsRepo = append(vcsRepo, modelDataSourceTFEWorkspaceVCSRepo{
			Identifier:              types.StringValue(workspace.VCSRepo.Identifier),
			Branch:                  types.StringValue(workspace.VCSRepo.Branch),
			IngressSubmodules:       types.BoolValue(workspace.VCSRepo.IngressSubmodules),
			OAuthTokenID:            types.StringValue(workspace.VCSRepo.OAuthTokenID),
			TagsRegex:               types.StringValue(workspace.VCSRepo.TagsRegex),
			GithubAppInstallationID: types.StringValue(workspace.VCSRepo.GHAInstallationID),
		})
	}

	result := modelDataSourceTFEWorkspace{
		ID:                          types.StringValue(workspace.ID),
		Name:                        types.StringValue(workspace.Name),
		Organization:                types.StringValue(organization),
		Description:                 types.StringValue(workspace.Description),
		AllowDestroyPlan:            types.BoolValue(workspace.AllowDestroyPlan),
		AutoApply:                   types.BoolValue(workspace.AutoApply),
		AutoApplyRunTrigger:         types.BoolValue(workspace.AutoApplyRunTrigger),
		AutoDestroyAt:               types.StringValue(autoDestroyAt),
		AutoDestroyActivityDuration: types.StringValue(autoDestroyDuration),
		InheritsProjectAutoDestroy:  types.BoolValue(workspace.InheritsProjectAutoDestroy),
		FileTriggersEnabled:         types.BoolValue(workspace.FileTriggersEnabled),
		GlobalRemoteState:           types.BoolValue(globalRemoteState),
		ProjectRemoteState:          types.BoolValue(projectRemoteState),
		RemoteStateConsumerIDs:      remoteStateConsumersValue,
		AssessmentsEnabled:          types.BoolValue(workspace.AssessmentsEnabled),
		Operations:                  types.BoolValue(workspace.Operations),
		PolicyCheckFailures:         types.Int64Value(int64(workspace.PolicyCheckFailures)),
		ProjectID:                   types.StringValue(projectID),
		QueueAllRuns:                types.BoolValue(workspace.QueueAllRuns),
		ResourceCount:               types.Int64Value(int64(workspace.ResourceCount)),
		RunFailures:                 types.Int64Value(int64(workspace.RunFailures)),
		RunsCount:                   types.Int64Value(int64(workspace.RunsCount)),
		SourceName:                  types.StringValue(workspace.SourceName),
		SourceURL:                   types.StringValue(workspace.SourceURL),
		SpeculativeEnabled:          types.BoolValue(workspace.SpeculativeEnabled),
		SSHKeyID:                    types.StringValue(sshKeyID),
		StructuredRunOutputEnabled:  types.BoolValue(workspace.StructuredRunOutputEnabled),
		EffectiveTags:               effectiveTagsValue,
		TagNames:                    tagNamesValue,
		TerraformVersion:            types.StringValue(workspace.TerraformVersion),
		TriggerPrefixes:             triggerPrefixesValue,
		TriggerPatterns:             triggerPatternsValue,
		WorkingDirectory:            types.StringValue(workspace.WorkingDirectory),
		ExecutionMode:               types.StringValue(workspace.ExecutionMode),
		VCSRepo:                     vcsRepo,
		HTMLURL:                     types.StringValue(htmlURL),
		HYOKEnabled:                 hyokEnabled,
		Locked:                      types.BoolValue(workspace.Locked),
		CreatedAt:                   types.StringValue(workspace.CreatedAt.Format(time.RFC3339)),
		UpdatedAt:                   types.StringValue(workspace.UpdatedAt.Format(time.RFC3339)),
		Environment:                 types.StringValue(workspace.Environment),
		ApplyDurationAverage:        types.Int64Value(workspace.ApplyDurationAverage.Milliseconds()),
		PlanDurationAverage:         types.Int64Value(workspace.PlanDurationAverage.Milliseconds()),
		Source:                      types.StringValue(string(workspace.Source)),
		SettingOverwrites:           settingOverwritesValue,
		Permissions:                 permissionsValue,
		Actions:                     actionsValue,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}
