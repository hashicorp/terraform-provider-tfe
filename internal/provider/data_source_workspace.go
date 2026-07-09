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
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-tfe/internal/provider/helpers"
)

var (
	_ datasource.DataSource              = &dataSourceTFEWorkspace{}
	_ datasource.DataSourceWithConfigure = &dataSourceTFEWorkspace{}
)

func NewWorkspaceDataSource() datasource.DataSource {
	return &dataSourceTFEWorkspace{}
}

type dataSourceTFEWorkspace struct {
	config ConfiguredClient
}

// model TFEWorkspace maps the data source schema data to a struct.
type modelTFEWorkspace struct {
	ID                          types.String                        `tfsdk:"id"`
	Name                        types.String                        `tfsdk:"name"`
	Organization                types.String                        `tfsdk:"organization"`
	Description                 types.String                        `tfsdk:"description"`
	AllowDestroyPlan            types.Bool                          `tfsdk:"allow_destroy_plan"`
	AutoApply                   types.Bool                          `tfsdk:"auto_apply"`
	AutoApplyRunTrigger         types.Bool                          `tfsdk:"auto_apply_run_trigger"`
	AutoDestroyAt               types.String                        `tfsdk:"auto_destroy_at"`
	AutoDestroyActivityDuration types.String                        `tfsdk:"auto_destroy_activity_duration"`
	InheritsProjectAutoDestroy  types.Bool                          `tfsdk:"inherits_project_auto_destroy"`
	FileTriggersEnabled         types.Bool                          `tfsdk:"file_triggers_enabled"`
	GlobalRemoteState           types.Bool                          `tfsdk:"global_remote_state"`
	ProjectRemoteState          types.Bool                          `tfsdk:"project_remote_state"`
	RemoteStateConsumerIDs      types.Set                           `tfsdk:"remote_state_consumer_ids"`
	AssessmentsEnabled          types.Bool                          `tfsdk:"assessments_enabled"`
	Operations                  types.Bool                          `tfsdk:"operations"`
	PolicyCheckFailures         types.Int64                         `tfsdk:"policy_check_failures"`
	ProjectID                   types.String                        `tfsdk:"project_id"`
	QueueAllRuns                types.Bool                          `tfsdk:"queue_all_runs"`
	ResourceCount               types.Int64                         `tfsdk:"resource_count"`
	RunFailures                 types.Int64                         `tfsdk:"run_failures"`
	RunsCount                   types.Int64                         `tfsdk:"runs_count"`
	SourceName                  types.String                        `tfsdk:"source_name"`
	SourceURL                   types.String                        `tfsdk:"source_url"`
	SpeculativeEnabled          types.Bool                          `tfsdk:"speculative_enabled"`
	SSHKeyID                    types.String                        `tfsdk:"ssh_key_id"`
	StructuredRunOutputEnabled  types.Bool                          `tfsdk:"structured_run_output_enabled"`
	EffectiveTags               types.Map                           `tfsdk:"effective_tags"`
	TagNames                    types.Set                           `tfsdk:"tag_names"`
	TerraformVersion            types.String                        `tfsdk:"terraform_version"`
	TriggerPrefixes             types.List                          `tfsdk:"trigger_prefixes"`
	TriggerPatterns             types.List                          `tfsdk:"trigger_patterns"`
	WorkingDirectory            types.String                        `tfsdk:"working_directory"`
	ExecutionMode               types.String                        `tfsdk:"execution_mode"`
	VCSRepo                     *modelTFEVCSRepo                    `tfsdk:"vcs_repo"`
	HTMLURL                     types.String                        `tfsdk:"html_url"`
	HYOKEnabled                 types.Bool                          `tfsdk:"hyok_enabled"`
	Locked                      types.Bool                          `tfsdk:"locked"`
	CreatedAt                   types.String                        `tfsdk:"created_at"`
	UpdatedAt                   types.String                        `tfsdk:"updated_at"`
	Environment                 types.String                        `tfsdk:"environment"`
	ApplyDurationAverage        types.Int64                         `tfsdk:"apply_duration_average"`
	PlanDurationAverage         types.Int64                         `tfsdk:"plan_duration_average"`
	Source                      types.String                        `tfsdk:"source"`
	SettingOverwrites           *modelTFEWorkspaceSettingOverwrites `tfsdk:"setting_overwrites"`
	Permissions                 *modelTFEWorkspacePermissions       `tfsdk:"permissions"`
	Actions                     *modelTFEWorkspaceActions           `tfsdk:"actions"`
}

// modelVCSRepo and modelFromTFEVCSRepo derives from data_source_registry_module which partially derives from go-tfe/v1 workspace.go

type modelTFEWorkspaceSettingOverwrites struct {
	ExecutionMode types.Bool `tfsdk:"execution_mode"`
	AgentPool     types.Bool `tfsdk:"agent_pool"`
}

type modelTFEWorkspacePermissions struct {
	CanUpdate         types.Bool `tfsdk:"can_update"`
	CanDestroy        types.Bool `tfsdk:"can_destroy"`
	CanQueueRun       types.Bool `tfsdk:"can_queue_run"`
	CanQueueApply     types.Bool `tfsdk:"can_queue_apply"`
	CanQueueDestroy   types.Bool `tfsdk:"can_queue_destroy"`
	CanLock           types.Bool `tfsdk:"can_lock"`
	CanUnlock         types.Bool `tfsdk:"can_unlock"`
	CanForceUnlock    types.Bool `tfsdk:"can_force_unlock"`
	CanReadSettings   types.Bool `tfsdk:"can_read_settings"`
	CanUpdateVariable types.Bool `tfsdk:"can_update_variable"`
	CanManageRunTasks types.Bool `tfsdk:"can_manage_run_tasks"`
	CanForceDelete    types.Bool `tfsdk:"can_force_delete"`
}

type modelTFEWorkspaceActions struct {
	IsDestroyable types.Bool `tfsdk:"is_destroyable"`
}

func modelFromTFEWorkspace(ctx context.Context, i *tfe.Workspace, autoDestroyAt *string, autoDestroyDuration string, htmlURLStringPointer *string, derivedGlobalRemoteState bool, remoteStateConsumerIDs []string) (modelTFEWorkspace, diag.Diagnostics) {
	var diags diag.Diagnostics

	model := modelTFEWorkspace{
		ID:                  types.StringValue(i.ID),
		Name:                types.StringValue(i.Name),
		Organization:        types.StringValue(i.Organization.Name),
		Description:         types.StringValue(i.Description),
		AllowDestroyPlan:    types.BoolValue(i.AllowDestroyPlan),
		AutoApply:           types.BoolValue(i.AutoApply),
		AutoApplyRunTrigger: types.BoolValue(i.AutoApplyRunTrigger),
		AutoDestroyAt: types.StringValue(func() string {
			if autoDestroyAt == nil {
				return ""
			}
			return *autoDestroyAt
		}()),
		AutoDestroyActivityDuration: types.StringValue(autoDestroyDuration),
		InheritsProjectAutoDestroy:  types.BoolValue(i.InheritsProjectAutoDestroy),
		FileTriggersEnabled:         types.BoolValue(i.FileTriggersEnabled),
		GlobalRemoteState:           types.BoolValue(derivedGlobalRemoteState),
		ProjectRemoteState:          types.BoolValue(i.ProjectRemoteState),
		RemoteStateConsumerIDs:      types.SetNull(types.StringType),
		AssessmentsEnabled:          types.BoolValue(i.AssessmentsEnabled),
		Operations:                  types.BoolValue(i.Operations),
		PolicyCheckFailures:         types.Int64Value(int64(i.PolicyCheckFailures)),
		ProjectID:                   types.StringNull(),
		QueueAllRuns:                types.BoolValue(i.QueueAllRuns),
		ResourceCount:               types.Int64Value(int64(i.ResourceCount)),
		RunFailures:                 types.Int64Value(int64(i.RunFailures)),
		RunsCount:                   types.Int64Value(int64(i.RunsCount)),
		SourceName:                  types.StringValue(i.SourceName),
		SourceURL:                   types.StringValue(i.SourceURL),
		SpeculativeEnabled:          types.BoolValue(i.SpeculativeEnabled),
		SSHKeyID:                    types.StringNull(),
		StructuredRunOutputEnabled:  types.BoolValue(i.StructuredRunOutputEnabled),
		EffectiveTags:               types.MapNull(types.StringType),
		TagNames:                    types.SetNull(types.StringType),
		TerraformVersion:            types.StringValue(i.TerraformVersion),
		TriggerPrefixes:             types.ListNull(types.StringType),
		TriggerPatterns:             types.ListNull(types.StringType),
		WorkingDirectory:            types.StringValue(i.WorkingDirectory),
		ExecutionMode:               types.StringValue(i.ExecutionMode),
		HTMLURL:                     types.StringNull(),
		HYOKEnabled: func() types.Bool {
			if i.HYOKEnabled == nil {
				return types.BoolNull()
			}
			return types.BoolValue(*i.HYOKEnabled)
		}(),
		Locked:               types.BoolValue(i.Locked),
		CreatedAt:            types.StringValue(i.CreatedAt.Format(time.RFC3339)),
		UpdatedAt:            types.StringValue(i.UpdatedAt.Format(time.RFC3339)),
		Environment:          types.StringValue(i.Environment),
		ApplyDurationAverage: types.Int64Value(i.ApplyDurationAverage.Milliseconds()),
		PlanDurationAverage:  types.Int64Value(i.PlanDurationAverage.Milliseconds()),
		Source:               types.StringValue(string(i.Source)),
		// VCSRepo, SettingOverwrites, Permissions, and Actions handled elsewhere
	}

	// pulling list and set items
	triggerPrefixes, prefixesConversionDiags := types.ListValueFrom(ctx, types.StringType, i.TriggerPrefixes)
	diags.Append(prefixesConversionDiags...)
	if diags.HasError() {
		return model, diags
	}
	model.TriggerPrefixes = triggerPrefixes

	triggerPatterns, patternsConversionDiags := types.ListValueFrom(ctx, types.StringType, i.TriggerPatterns)
	diags.Append(patternsConversionDiags...)
	if diags.HasError() {
		return model, diags
	}
	model.TriggerPatterns = triggerPatterns

	// If target tfe instance predates projects, then workspace.Project will be nil
	if i.Project != nil {
		model.ProjectID = types.StringValue(i.Project.ID)
	}

	// Setting overwrites
	model.SettingOverwrites = modelWorkspaceSettingOverwritesFromTFEWorkspace(i)

	// Set permissions
	model.Permissions = modelWorkspacePermissionsFromTFEWorkspace(i)

	// Set actions
	model.Actions = modelWorkspaceActionsFromTFEWorkspace(i)

	// html url
	if htmlURLStringPointer != nil {
		model.HTMLURL = types.StringValue(*htmlURLStringPointer)
	}

	// ssh key id
	if i.SSHKey != nil {
		model.SSHKeyID = types.StringValue(i.SSHKey.ID)
	}

	// Pull in remote state consumer ID information
	if remoteStateConsumerIDs != nil {
		convertedRemoteStateConsumerIDs, remoteStateConsumerDiags := types.SetValueFrom(ctx, types.StringType, remoteStateConsumerIDs)
		diags.Append(remoteStateConsumerDiags...)
		if diags.HasError() {
			return model, diags
		}
		model.RemoteStateConsumerIDs = convertedRemoteStateConsumerIDs
	}

	// Include tags
	tagInfo := helpers.NewTagInfo(nil, i.EffectiveTagBindings, false)
	effectiveTags, effectiveTagsDiags := types.MapValueFrom(ctx, types.StringType, tagInfo.EffectiveTags)
	diags.Append(effectiveTagsDiags...)
	if diags.HasError() {
		return model, diags
	}
	model.EffectiveTags = effectiveTags

	// Update the tag names
	tagNames, tagNamesDiags := types.SetValueFrom(ctx, types.StringType, i.TagNames)
	diags.Append(tagNamesDiags...)
	if diags.HasError() {
		return model, diags
	}
	model.TagNames = tagNames

	// Set VCSRepo
	if i.VCSRepo != nil {
		VCSRepoStruct := modelFromTFEVCSRepo(i.VCSRepo)
		model.VCSRepo = &VCSRepoStruct
	}

	return model, diags
}

func modelWorkspaceSettingOverwritesFromTFEWorkspace(i *tfe.Workspace) *modelTFEWorkspaceSettingOverwrites {
	model := &modelTFEWorkspaceSettingOverwrites{
		ExecutionMode: types.BoolNull(),
		AgentPool:     types.BoolNull(),
	}

	if i.SettingOverwrites == nil {
		return model
	}

	if i.SettingOverwrites.ExecutionMode != nil {
		model.ExecutionMode = types.BoolValue(*i.SettingOverwrites.ExecutionMode)
	}
	if i.SettingOverwrites.AgentPool != nil {
		model.AgentPool = types.BoolValue(*i.SettingOverwrites.AgentPool)
	}

	return model
}

func modelWorkspacePermissionsFromTFEWorkspace(i *tfe.Workspace) *modelTFEWorkspacePermissions {
	if i.Permissions == nil {
		return &modelTFEWorkspacePermissions{
			CanUpdate:         types.BoolNull(),
			CanDestroy:        types.BoolNull(),
			CanQueueRun:       types.BoolNull(),
			CanQueueApply:     types.BoolNull(),
			CanQueueDestroy:   types.BoolNull(),
			CanLock:           types.BoolNull(),
			CanUnlock:         types.BoolNull(),
			CanForceUnlock:    types.BoolNull(),
			CanReadSettings:   types.BoolNull(),
			CanUpdateVariable: types.BoolNull(),
			CanManageRunTasks: types.BoolNull(),
			CanForceDelete:    types.BoolNull(),
		}
	}

	model := &modelTFEWorkspacePermissions{
		CanUpdate:         types.BoolValue(i.Permissions.CanUpdate),
		CanDestroy:        types.BoolValue(i.Permissions.CanDestroy),
		CanQueueRun:       types.BoolValue(i.Permissions.CanQueueRun),
		CanQueueApply:     types.BoolValue(i.Permissions.CanQueueApply),
		CanQueueDestroy:   types.BoolValue(i.Permissions.CanQueueDestroy),
		CanLock:           types.BoolValue(i.Permissions.CanLock),
		CanUnlock:         types.BoolValue(i.Permissions.CanUnlock),
		CanForceUnlock:    types.BoolValue(i.Permissions.CanForceUnlock),
		CanReadSettings:   types.BoolValue(i.Permissions.CanReadSettings),
		CanUpdateVariable: types.BoolValue(i.Permissions.CanUpdateVariable),
		CanManageRunTasks: types.BoolValue(i.Permissions.CanManageRunTasks),
		CanForceDelete:    types.BoolNull(),
	}

	if i.Permissions.CanForceDelete != nil {
		model.CanForceDelete = types.BoolValue(*i.Permissions.CanForceDelete)
	}

	return model
}

func modelWorkspaceActionsFromTFEWorkspace(i *tfe.Workspace) *modelTFEWorkspaceActions {
	if i.Actions == nil {
		return &modelTFEWorkspaceActions{
			IsDestroyable: types.BoolNull(),
		}
	}

	model := &modelTFEWorkspaceActions{
		IsDestroyable: types.BoolValue(i.Actions.IsDestroyable),
	}

	return model
}

func (d *dataSourceTFEWorkspace) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workspace"
}

func (d *dataSourceTFEWorkspace) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Gets information about a workspace. Note that using `global_remote_state` or `remote_state_consumer_ids` requires using the provider with HCP Terraform or an instance of Terraform Enterprise at least as recent as v202104-1.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The workspace ID.",
				Computed:    true,
			},

			"name": schema.StringAttribute{
				Description: "Name of the workspace.",
				Required:    true,
			},

			"organization": schema.StringAttribute{
				Description: "Name of the organization.",
				Optional:    true,
			},

			"description": schema.StringAttribute{
				Description: "Description of the workspace.",
				Computed:    true,
			},

			"allow_destroy_plan": schema.BoolAttribute{
				Description: "Indicates whether destroy plans can be queued on the workspace.",
				Computed:    true,
			},

			"auto_apply": schema.BoolAttribute{
				Description: "Indicates whether to automatically apply changes when a Terraform plan is successful.",
				Computed:    true,
			},

			"auto_apply_run_trigger": schema.BoolAttribute{
				Description: "Whether the workspace will automatically apply changes for runs that were created by run triggers from another workspace.",
				Computed:    true,
			},

			"auto_destroy_at": schema.StringAttribute{
				Description: "Future date/time string at which point all resources in a workspace will be scheduled to be deleted.",
				Computed:    true,
			},

			"auto_destroy_activity_duration": schema.StringAttribute{
				Description: "A duration string representing time after workspace activity when an auto-destroy run will be triggered.",
				Computed:    true,
			},

			"inherits_project_auto_destroy": schema.BoolAttribute{
				Description: "Indicates whether this workspace inherits project auto destroy settings.",
				Computed:    true,
			},

			"file_triggers_enabled": schema.BoolAttribute{
				Description: "Indicates whether runs are triggered based on the changed files in a VCS push (if `true`) or always triggered on every push (if `false`).",
				Computed:    true,
			},

			"global_remote_state": schema.BoolAttribute{
				Description: "Whether the workspace should allow all workspaces in the organization to access its state data during runs. If false, then only specifically approved workspaces can access its state (determined by the `remote_state_consumer_ids` argument). Cannot be true if `project_remote_state` is true.",
				Computed:    true,
			},

			"project_remote_state": schema.BoolAttribute{
				Description: "Whether the workspace should allow all workspaces in the project to access its state data during runs. If false, then only specifically approved workspaces can access its state (determined by the `remote_state_consumer_ids` argument). Cannot be true if `global_remote_state` is true.",
				Computed:    true,
			},

			"remote_state_consumer_ids": schema.SetAttribute{
				Description: "A set of workspace IDs that will be set as the remote state consumers for the given workspace. Cannot be used if `global_remote_state` or `project_remote_state` is set to `true`.",
				Computed:    true,
				ElementType: types.StringType,
			},

			"assessments_enabled": schema.BoolAttribute{
				Description: "(Available only in HCP Terraform) Indicates whether health assessments such as drift detection are enabled for the workspace.",
				Computed:    true,
			},

			"operations": schema.BoolAttribute{
				Description: "Indicates whether the workspace is using remote execution mode. Set to `false` to switch execution mode to local. `true` by default.",
				Computed:    true,
			},

			"policy_check_failures": schema.Int64Attribute{
				Description: "The number of policy check failures from the latest run.",
				Computed:    true,
			},

			"project_id": schema.StringAttribute{
				Description: "ID of the workspace's project.",
				Computed:    true,
			},

			"queue_all_runs": schema.BoolAttribute{
				Description: "Indicates whether the workspace will automatically perform runs in response to webhooks immediately after its creation. If `false`, an initial run must be manually queued to enable future automatic runs.",
				Computed:    true,
			},

			"resource_count": schema.Int64Attribute{
				Description: "The number of resources managed by the workspace.",
				Computed:    true,
			},

			"run_failures": schema.Int64Attribute{
				Description: "The number of run failures on the workspace.",
				Computed:    true,
			},

			"runs_count": schema.Int64Attribute{
				Description: "The number of runs on the workspace.",
				Computed:    true,
			},

			"source_name": schema.StringAttribute{
				Description: "The name of the workspace creation source, if set.",
				Computed:    true,
			},

			"source_url": schema.StringAttribute{
				Description: "The URL of the workspace creation source, if set.",
				Computed:    true,
			},

			"speculative_enabled": schema.BoolAttribute{
				Description: "Indicates whether this workspace allows speculative plans.",
				Computed:    true,
			},

			"ssh_key_id": schema.StringAttribute{
				Description: "The ID of an SSH key assigned to the workspace.",
				Computed:    true,
			},

			"structured_run_output_enabled": schema.BoolAttribute{
				Description: "Indicates whether runs in this workspace use the enhanced apply UI.",
				Computed:    true,
			},

			"effective_tags": schema.MapAttribute{
				Description: "A map of key-value tags associated with the workspace, including any inherited tags from the parent project.",
				Computed:    true,
				ElementType: types.StringType,
			},

			"tag_names": schema.SetAttribute{
				Description: "The names of tags added to this workspace.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
			},

			"terraform_version": schema.StringAttribute{
				Description: "The version (or version constraint) of Terraform used for this workspace.",
				Computed:    true,
			},

			"trigger_prefixes": schema.ListAttribute{
				Description: "List of trigger prefixes that describe the paths HCP Terraform monitors for changes, in addition to the working directory. Trigger prefixes are always appended to the root directory of the repository.",
				Computed:    true,
				ElementType: types.StringType,
			},

			"trigger_patterns": schema.ListAttribute{
				MarkdownDescription: "List of [glob patterns](https://developer.hashicorp.com/terraform/cloud-docs/workspaces/settings/vcs#glob-patterns-for-automatic-run-triggering) that describe the files HCP Terraform monitors for changes. Trigger patterns are always appended to the root directory of the repository.",
				Computed:            true,
				ElementType:         types.StringType,
			},

			"working_directory": schema.StringAttribute{
				Description: "A relative path that Terraform will execute within.",
				Computed:    true,
			},

			"execution_mode": schema.StringAttribute{
				MarkdownDescription: "Indicates the [execution mode](https://developer.hashicorp.com/terraform/cloud-docs/workspaces/settings#execution-mode) of the workspace. **Note:** This value might be derived from an organization-level default or set on the workspace itself; see the [`tfe_workspace_settings` resource](tfe_workspace_settings) for details.",
				Computed:            true,
			},
			"html_url": schema.StringAttribute{
				Description: "The URL to the browsable HTML overview of the workspace.",
				Computed:    true,
			},
			"hyok_enabled": schema.BoolAttribute{
				Description: "Whether HYOK is enbabled for the workspace.",
				Computed:    true,
			},
			"locked": schema.BoolAttribute{
				Description: "Indicates whether the workspace is locked.",
				Computed:    true,
			},

			"created_at": schema.StringAttribute{
				Description: "The time when the workspace was created.",
				Computed:    true,
			},

			"updated_at": schema.StringAttribute{
				Description: "The time when the workspace was last updated.",
				Computed:    true,
			},

			"environment": schema.StringAttribute{
				Description: "The environment of the workspace.",
				Computed:    true,
			},

			"apply_duration_average": schema.Int64Attribute{
				Description: "The average duration of applies for this workspace.",
				Computed:    true,
			},

			"plan_duration_average": schema.Int64Attribute{
				Description: "The average duration of plans for this workspace.",
				Computed:    true,
			},

			"source": schema.StringAttribute{
				Description: "The source of the workspace.",
				Computed:    true,
			},
		},

		// !!!: beware these modifications have been made
		//
		// vcs_repo[0].identifier
		// -->
		// vcs_repo.identifier
		//
		// and for the other three blocks
		//
		// actions['is-destroyable']
		// -->
		// actions.is_destroyable
		//

		Blocks: map[string]schema.Block{
			"vcs_repo": schema.SingleNestedBlock{
				Description: "Settings for the workspace's VCS repository.",
				Attributes: map[string]schema.Attribute{
					"identifier": schema.StringAttribute{
						Description: "A reference to your VCS repository in the format `<vcs organization>/<repository>` where `<vcs organization>` and `<repository>` refer to the organization and repository in your VCS provider.",
						Computed:    true,
					},
					"branch": schema.StringAttribute{
						Description: "The repository branch that Terraform will execute from.",
						Computed:    true,
					},
					"ingress_submodules": schema.BoolAttribute{
						Description: "Indicates whether submodules should be fetched when cloning the VCS repository.",
						Computed:    true,
					},
					"oauth_token_id": schema.StringAttribute{
						Description: "OAuth token ID of the configured VCS connection.",
						Computed:    true,
					},
					"github_app_installation_id": schema.StringAttribute{
						Description: "The installation ID of the GitHub App. Conflicts with `oauth_token_id`.",
						Computed:    true,
					},
					"tags_regex": schema.StringAttribute{
						Description: "A regular expression used to trigger a workspace run for matching Git tags.",
						Computed:    true,
					},
				},
			},
			"setting_overwrites": schema.SingleNestedBlock{
				Description: "Settings that are overwritten for this workspace.",
				Attributes: map[string]schema.Attribute{
					"execution_mode": schema.BoolAttribute{
						Description: "Whether execution mode is overwritten at the workspace level.",
						Computed:    true,
					},
					"agent_pool": schema.BoolAttribute{
						Description: "Whether agent pool is overwritten at the workspace level.",
						Computed:    true,
					},
				},
			},
			"actions": schema.SingleNestedBlock{
				Description: "Actions that can be performed on this workspace.",
				Attributes: map[string]schema.Attribute{
					"is_destroyable": schema.BoolAttribute{
						Description: "Whether the workspace can be destroyed.",
						Computed:    true,
					},
				},
			},
			"permissions": schema.SingleNestedBlock{
				Description: "The permissions for the current user on this workspace.",
				Attributes: map[string]schema.Attribute{
					"can_update": schema.BoolAttribute{
						Description: "Can update the workspace.",
						Computed:    true,
					},
					"can_destroy": schema.BoolAttribute{
						Description: "Can destroy the workspace.",
						Computed:    true,
					},
					"can_queue_run": schema.BoolAttribute{
						Description: "Can queue runs.",
						Computed:    true,
					},
					"can_queue_apply": schema.BoolAttribute{
						Description: "Can queue apply.",
						Computed:    true,
					},
					"can_queue_destroy": schema.BoolAttribute{
						Description: "Can queue destroy.",
						Computed:    true,
					},
					"can_lock": schema.BoolAttribute{
						Description: "Can lock the workspace.",
						Computed:    true,
					},
					"can_unlock": schema.BoolAttribute{
						Description: "Can unlock the workspace.",
						Computed:    true,
					},
					"can_force_unlock": schema.BoolAttribute{
						Description: "Can force unlock the workspace.",
						Computed:    true,
					},
					"can_read_settings": schema.BoolAttribute{
						Description: "Can read workspace settings.",
						Computed:    true,
					},
					"can_update_variable": schema.BoolAttribute{
						Description: "Can update variables.",
						Computed:    true,
					},
					"can_manage_run_tasks": schema.BoolAttribute{
						Description: "Can manage run tasks.",
						Computed:    true,
					},
					"can_force_delete": schema.BoolAttribute{
						Description: "Can force delete the workspace.",
						Computed:    true,
					},
				},
			},
		},
	}
}

func fallbackWorkspaceRead(ctx context.Context, config ConfiguredClient, organization string, name string) (*tfe.Workspace, diag.Diagnostics) {
	var diags diag.Diagnostics

	tflog.Debug(ctx, fmt.Sprintf("Workspace %s read failed due to unsupported Include; retrying without it", name))

	workspace, err := config.Client.Workspaces.Read(ctx, organization, name)
	if err != nil && errors.Is(err, tfe.ErrResourceNotFound) {
		diags.AddError(fmt.Sprintf("Could not find workspace %s/%s", organization, name), err.Error())
		return nil, diags
	} else if err != nil {
		diags.AddError(fmt.Sprintf("Error reading workspace %s without include", name), err.Error())
		return nil, diags
	}

	return workspace, diags
}

func (d *dataSourceTFEWorkspace) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data modelTFEWorkspace
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var organization string
	resp.Diagnostics.Append(d.config.dataOrDefaultOrganization(ctx, req.Config, &organization)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.ValueString()
	tflog.Debug(ctx, fmt.Sprintf("Requesting configuration of workspace: %s", name))

	workspace, err := d.config.Client.Workspaces.ReadWithOptions(ctx, organization, name, &tfe.WorkspaceReadOptions{
		Include: []tfe.WSIncludeOpt{tfe.WSEffectiveTagBindings}, // ??? What is this doing, exactly
	})
	if err != nil && errors.Is(err, tfe.ErrResourceNotFound) {
		resp.Diagnostics.AddError("Unable to read workspace", err.Error())
		return
	}
	if err != nil && errors.Is(err, tfe.ErrInvalidIncludeValue) {
		var fallbackDiags diag.Diagnostics
		workspace, fallbackDiags = fallbackWorkspaceRead(ctx, d.config, organization, name)
		resp.Diagnostics.Append(fallbackDiags...)
		if resp.Diagnostics.HasError() {
			return
		}
	} else if err != nil { // any new err created by the previous block will be captured by the previous block's post-action checks
		resp.Diagnostics.AddError("Unable to read workspace", err.Error())
		return
	}

	// Collect auto destroy information
	autoDestroyAt, err := flattenAutoDestroyAt(workspace.AutoDestroyAt)
	if err != nil {
		resp.Diagnostics.AddError("Error flattening auto destroy during read", err.Error())
		return
	}
	var autoDestroyDuration string
	if workspace.AutoDestroyActivityDuration.IsSpecified() {
		autoDestroyDuration, err = workspace.AutoDestroyActivityDuration.Get()
		if err != nil {
			resp.Diagnostics.AddError("Error reading auto destroy activity duration", err.Error())
			return
		}
	}

	// html url prep
	var htmlURLStringPointer *string = nil // ??? I hope there's a better way to do this
	if workspace.Links["self-html"] != nil {
		baseAPI := d.config.Client.BaseURL()
		htmlURL := url.URL{
			Scheme: baseAPI.Scheme,
			Host:   baseAPI.Host,
			Path:   workspace.Links["self-html"].(string),
		}
		htmlURLString := htmlURL.String()
		htmlURLStringPointer = &htmlURLString
	}

	// Set remote_state_consumer_ids if global_remote_state and project_remote_state are false
	var remoteStateConsumerIDs []string
	globalRemoteState := workspace.GlobalRemoteState
	derivedGlobalRemoteState := globalRemoteState
	projectRemoteState := workspace.ProjectRemoteState
	if globalRemoteState || projectRemoteState {
		remoteStateConsumerIDs = []string{} // make non-nil
	} else {
		legacyGlobalState, resultRemoteStateConsumerIDs, err := readWorkspaceStateConsumers(workspace.ID, d.config.Client)

		if err != nil {
			resp.Diagnostics.AddError(fmt.Sprintf("Error reading remote state consumers for workspace %s", workspace.ID), err.Error())
		}

		if legacyGlobalState {
			derivedGlobalRemoteState = true
		}
		remoteStateConsumerIDs = resultRemoteStateConsumerIDs
	}

	// Pipe everything in and build the model
	result, diags := modelFromTFEWorkspace(ctx, workspace, autoDestroyAt, autoDestroyDuration, htmlURLStringPointer, derivedGlobalRemoteState, remoteStateConsumerIDs)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (d *dataSourceTFEWorkspace) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Early exit if provider is unconfigured (i.e. we're only validating config or something)
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(ConfiguredClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure type",
			fmt.Sprintf("Expected tfe.ConfiguredClient, got %T. This is a bug in the tfe provider, so please report it on GitHub.", req.ProviderData),
		)
	}
	d.config = client
}
