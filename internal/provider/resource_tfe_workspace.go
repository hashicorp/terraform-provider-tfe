// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/jsonapi"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-tfe/internal/provider/helpers"
)

var (
	_ resource.Resource                 = &resourceTFEWorkspaceFramework{}
	_ resource.ResourceWithConfigure    = &resourceTFEWorkspaceFramework{}
	_ resource.ResourceWithModifyPlan   = &resourceTFEWorkspaceFramework{}
	_ resource.ResourceWithImportState  = &resourceTFEWorkspaceFramework{}
	_ resource.ResourceWithUpgradeState = &resourceTFEWorkspaceFramework{}
)

type resourceTFEWorkspaceFramework struct {
	config ConfiguredClient
}

type modelWorkspace struct {
	ID                          types.String `tfsdk:"id"`
	Name                        types.String `tfsdk:"name"`
	Organization                types.String `tfsdk:"organization"`
	Description                 types.String `tfsdk:"description"`
	AgentPoolID                 types.String `tfsdk:"agent_pool_id"`
	AllowDestroyPlan            types.Bool   `tfsdk:"allow_destroy_plan"`
	AutoApply                   types.Bool   `tfsdk:"auto_apply"`
	AutoApplyRunTrigger         types.Bool   `tfsdk:"auto_apply_run_trigger"`
	AutoDestroyAt               types.String `tfsdk:"auto_destroy_at"`
	AutoDestroyActivityDuration types.String `tfsdk:"auto_destroy_activity_duration"`
	ExecutionMode               types.String `tfsdk:"execution_mode"`
	FileTriggersEnabled         types.Bool   `tfsdk:"file_triggers_enabled"`
	GlobalRemoteState           types.Bool   `tfsdk:"global_remote_state"`
	InheritsProjectAutoDestroy  types.Bool   `tfsdk:"inherits_project_auto_destroy"`
	RemoteStateConsumerIDs      types.Set    `tfsdk:"remote_state_consumer_ids"`
	AssessmentsEnabled          types.Bool   `tfsdk:"assessments_enabled"`
	Operations                  types.Bool   `tfsdk:"operations"`
	ProjectID                   types.String `tfsdk:"project_id"`
	QueueAllRuns                types.Bool   `tfsdk:"queue_all_runs"`
	SourceName                  types.String `tfsdk:"source_name"`
	SourceURL                   types.String `tfsdk:"source_url"`
	SpeculativeEnabled          types.Bool   `tfsdk:"speculative_enabled"`
	SSHKeyID                    types.String `tfsdk:"ssh_key_id"`
	StructuredRunOutputEnabled  types.Bool   `tfsdk:"structured_run_output_enabled"`
	TagNames                    types.Set    `tfsdk:"tag_names"`
	IgnoreAdditionalTagNames    types.Bool   `tfsdk:"ignore_additional_tag_names"`
	Tags                        types.Map    `tfsdk:"tags"`
	IgnoreAdditionalTags        types.Bool   `tfsdk:"ignore_additional_tags"`
	EffectiveTags               types.Map    `tfsdk:"effective_tags"`
	TerraformVersion            types.String `tfsdk:"terraform_version"`
	TriggerPrefixes             types.List   `tfsdk:"trigger_prefixes"`
	TriggerPatterns             types.List   `tfsdk:"trigger_patterns"`
	WorkingDirectory            types.String `tfsdk:"working_directory"`
	VCSRepo                     types.List   `tfsdk:"vcs_repo"`
	ForceDelete                 types.Bool   `tfsdk:"force_delete"`
	ResourceCount               types.Int64  `tfsdk:"resource_count"`
	HTMLURL                     types.String `tfsdk:"html_url"`
	HYOKEnabled                 types.Bool   `tfsdk:"hyok_enabled"`
}

type modelWorkspaceIdentity struct {
	ID       types.String `tfsdk:"id"`
	Hostname types.String `tfsdk:"hostname"`
}

type modelWorkspaceVCSRepo struct {
	Identifier              types.String `tfsdk:"identifier"`
	Branch                  types.String `tfsdk:"branch"`
	IngressSubmodules       types.Bool   `tfsdk:"ingress_submodules"`
	OAuthTokenID            types.String `tfsdk:"oauth_token_id"`
	TagsRegex               types.String `tfsdk:"tags_regex"`
	GithubAppInstallationID types.String `tfsdk:"github_app_installation_id"`
}

func NewWorkspaceResource() resource.Resource {
	return &resourceTFEWorkspaceFramework{}
}

func (r *resourceTFEWorkspaceFramework) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workspace"
}

func (r *resourceTFEWorkspaceFramework) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(ConfiguredClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected resource Configure type",
			fmt.Sprintf("Expected tfe.ConfiguredClient, got %T", req.ProviderData),
		)
		return
	}

	r.config = client
}

func (r *resourceTFEWorkspaceFramework) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":                             schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"name":                           schema.StringAttribute{Required: true},
		"organization":                   schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
		"description":                    schema.StringAttribute{Optional: true, Computed: true},
		"agent_pool_id":                  schema.StringAttribute{Optional: true, Computed: true, DeprecationMessage: "Use resource tfe_workspace_settings to modify the workspace execution settings. This attribute will be removed in a future release of the provider.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"allow_destroy_plan":             schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true)},
		"auto_apply":                     schema.BoolAttribute{Optional: true, Computed: true},
		"auto_apply_run_trigger":         schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false)},
		"auto_destroy_at":                schema.StringAttribute{Optional: true, Computed: true},
		"auto_destroy_activity_duration": schema.StringAttribute{Optional: true, Computed: true, Validators: []validator.String{stringvalidator.RegexMatches(regexp.MustCompile(`^\d{1,4}[dh]$`), "must be 1-4 digits followed by d or h")}},
		"execution_mode":                 schema.StringAttribute{Optional: true, Computed: true, DeprecationMessage: "Use resource tfe_workspace_settings to modify the workspace execution settings. This attribute will be removed in a future release of the provider.", Validators: []validator.String{stringvalidator.OneOf("agent", "local", "remote")}},
		"file_triggers_enabled":          schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true)},
		"global_remote_state":            schema.BoolAttribute{Optional: true, Computed: true, DeprecationMessage: "Use resource `tfe_workspace_settings` to modify the workspace `global_remote_state`. `global_remote_state` on `tfe_workspace` is no longer validated properly and will be removed in a future release of the provider.", PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}},
		"inherits_project_auto_destroy":  schema.BoolAttribute{Computed: true},
		"remote_state_consumer_ids":      schema.SetAttribute{Optional: true, Computed: true, ElementType: types.StringType, DeprecationMessage: "Use resource `tfe_workspace_settings` to modify the workspace `remote_state_consumer_ids`. This attribute will be removed in a future release of the provider.", PlanModifiers: []planmodifier.Set{setplanmodifier.UseStateForUnknown()}},
		"assessments_enabled":            schema.BoolAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}},
		"operations":                     schema.BoolAttribute{Optional: true, Computed: true, DeprecationMessage: "Use tfe_workspace_settings to modify the workspace execution settings. This attribute will be removed in a future release of the provider."},
		"project_id":                     schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"queue_all_runs":                 schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true)},
		"source_name":                    schema.StringAttribute{Optional: true},
		"source_url":                     schema.StringAttribute{Optional: true, Validators: []validator.String{stringvalidator.RegexMatches(regexp.MustCompile(`^https?://`), "must be a valid URL with http or https scheme")}},
		"speculative_enabled":            schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true)},
		"ssh_key_id":                     schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("")},
		"structured_run_output_enabled":  schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true)},
		"tag_names":                      schema.SetAttribute{Optional: true, Computed: true, ElementType: types.StringType},
		"ignore_additional_tag_names":    schema.BoolAttribute{Optional: true},
		"tags":                           schema.MapAttribute{Optional: true, Computed: true, ElementType: types.StringType, PlanModifiers: []planmodifier.Map{mapplanmodifier.UseStateForUnknown()}},
		"ignore_additional_tags":         schema.BoolAttribute{Optional: true},
		"effective_tags":                 schema.MapAttribute{Computed: true, ElementType: types.StringType},
		"terraform_version":              schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"trigger_prefixes":               schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType, Default: listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{}))},
		"trigger_patterns":               schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType, Default: listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{}))},
		"working_directory":              schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("")},
		"force_delete":                   schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false)},
		"resource_count":                 schema.Int64Attribute{Computed: true, PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()}},
		"html_url":                       schema.StringAttribute{Computed: true},
		"hyok_enabled":                   schema.BoolAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}},
	},
		Blocks: map[string]schema.Block{
			"vcs_repo": schema.ListNestedBlock{
				Validators: []validator.List{
					listvalidator.SizeBetween(1, 1),
				},
				NestedObject: schema.NestedBlockObject{Attributes: map[string]schema.Attribute{
					"identifier":                 schema.StringAttribute{Required: true},
					"branch":                     schema.StringAttribute{Optional: true, Computed: true},
					"ingress_submodules":         schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false)},
					"oauth_token_id":             schema.StringAttribute{Optional: true, Validators: []validator.String{stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("github_app_installation_id"))}},
					"tags_regex":                 schema.StringAttribute{Optional: true, Computed: true, Validators: []validator.String{stringvalidator.ConflictsWith(path.MatchRoot("trigger_patterns"), path.MatchRoot("trigger_prefixes"))}},
					"github_app_installation_id": schema.StringAttribute{Optional: true, Validators: []validator.String{stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("oauth_token_id")), stringvalidator.AtLeastOneOf(path.MatchRelative().AtParent().AtName("oauth_token_id"), path.MatchRelative().AtParent().AtName("github_app_installation_id"))}},
				}},
			},
		},
	}
}

func (r *resourceTFEWorkspaceFramework) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{Attributes: map[string]identityschema.Attribute{
		"id":       identityschema.StringAttribute{RequiredForImport: true},
		"hostname": identityschema.StringAttribute{OptionalForImport: true},
	}}
}

func (r *resourceTFEWorkspaceFramework) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema: &resourceTFEWorkspaceSchemaV0,
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				if req.State == nil {
					resp.Diagnostics.AddError("Error upgrading workspace state", "missing prior state")
					return
				}

				var oldData modelWorkspaceV0
				resp.Diagnostics.Append(req.State.Get(ctx, &oldData)...)
				if resp.Diagnostics.HasError() {
					return
				}

				if oldData.ExternalID.IsNull() || oldData.ExternalID.IsUnknown() || oldData.ExternalID.ValueString() == "" {
					resp.Diagnostics.AddError("Error upgrading workspace state", "missing or invalid external_id in prior state")
					return
				}

				newData := modelWorkspace{
					ID:                     types.StringValue(oldData.ExternalID.ValueString()),
					Name:                   oldData.Name,
					Organization:           oldData.Organization,
					AssessmentsEnabled:     oldData.AssessmentsEnabled,
					AutoApply:              oldData.AutoApply,
					FileTriggersEnabled:    oldData.FileTriggersEnabled,
					Operations:             oldData.Operations,
					QueueAllRuns:           oldData.QueueAllRuns,
					SSHKeyID:               oldData.SSHKeyID,
					TerraformVersion:       oldData.TerraformVersion,
					TriggerPrefixes:        oldData.TriggerPrefixes,
					TriggerPatterns:        types.ListNull(types.StringType),
					WorkingDirectory:       oldData.WorkingDirectory,
					TagNames:               types.SetNull(types.StringType),
					RemoteStateConsumerIDs: types.SetNull(types.StringType),
					Tags:                   types.MapNull(types.StringType),
					EffectiveTags:          types.MapNull(types.StringType),
					VCSRepo: types.ListNull(types.ObjectType{AttrTypes: map[string]attr.Type{
						"identifier":                 types.StringType,
						"branch":                     types.StringType,
						"ingress_submodules":         types.BoolType,
						"oauth_token_id":             types.StringType,
						"tags_regex":                 types.StringType,
						"github_app_installation_id": types.StringType,
					}}),
				}
				resp.Diagnostics.Append(resp.State.Set(ctx, newData)...)
			},
		},
	}
}

func (r *resourceTFEWorkspaceFramework) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}

	modifyPlanForDefaultOrganizationChange(ctx, r.config.Organization, req.State, req.Config, req.Plan, resp)

	var plan modelWorkspace
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var configSourceName types.String
	var configSourceURL types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("source_name"), &configSourceName)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("source_url"), &configSourceURL)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !configSourceURL.IsNull() && !configSourceURL.IsUnknown() && (configSourceName.IsNull() || configSourceName.IsUnknown()) {
		resp.Diagnostics.AddError("Missing required argument", "The argument \"source_name\" is required when \"source_url\" is set")
	}
	if !configSourceName.IsNull() && !configSourceName.IsUnknown() && (configSourceURL.IsNull() || configSourceURL.IsUnknown()) {
		resp.Diagnostics.AddError("Missing required argument", "The argument \"source_url\" is required when \"source_name\" is set")
	}

	var configTriggerPrefixes types.List
	var configTriggerPatterns types.List
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("trigger_prefixes"), &configTriggerPrefixes)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("trigger_patterns"), &configTriggerPatterns)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if hasConfiguredTriggerConflict(configTriggerPrefixes, configTriggerPatterns) {
		resp.Diagnostics.AddError("Conflicting configuration", "Only one of trigger_prefixes or trigger_patterns can be configured")
	}

	if !plan.TagNames.IsNull() && !plan.TagNames.IsUnknown() {
		var tagNames []string
		resp.Diagnostics.Append(plan.TagNames.ElementsAs(ctx, &tagNames, false)...)
		for _, tagName := range tagNames {
			if !validTagName(tagName) {
				resp.Diagnostics.AddError("Invalid tag_names value", fmt.Sprintf("%q is not a valid tag name. Tag must begin and end with alphanumeric lowercase characters", tagName))
			}
		}
	}
}

func (r *resourceTFEWorkspaceFramework) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan modelWorkspace
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var orgName string
	resp.Diagnostics.Append(r.config.dataOrDefaultOrganization(ctx, req.Config, &orgName)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := tfe.WorkspaceCreateOptions{
		Name:                       tfe.String(plan.Name.ValueString()),
		AllowDestroyPlan:           tfe.Bool(boolValueOrDefault(plan.AllowDestroyPlan, true)),
		AutoApplyRunTrigger:        tfe.Bool(boolValueOrDefault(plan.AutoApplyRunTrigger, false)),
		FileTriggersEnabled:        tfe.Bool(boolValueOrDefault(plan.FileTriggersEnabled, true)),
		QueueAllRuns:               tfe.Bool(boolValueOrDefault(plan.QueueAllRuns, true)),
		SpeculativeEnabled:         tfe.Bool(boolValueOrDefault(plan.SpeculativeEnabled, true)),
		StructuredRunOutputEnabled: tfe.Bool(boolValueOrDefault(plan.StructuredRunOutputEnabled, true)),
		WorkingDirectory:           tfe.String(stringValueOrDefault(plan.WorkingDirectory, "")),
	}

	r.applyWorkspaceOptionsFromModel(ctx, &plan, &opts, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	ws, err := r.config.Client.Workspaces.Create(ctx, orgName, opts)
	if err != nil {
		resp.Diagnostics.AddError("Error creating workspace", fmt.Sprintf("Error creating workspace %s for organization %s: %v", plan.Name.ValueString(), orgName, err))
		return
	}

	if !plan.SSHKeyID.IsNull() && plan.SSHKeyID.ValueString() != "" {
		_, err = r.config.Client.Workspaces.AssignSSHKey(ctx, ws.ID, tfe.WorkspaceAssignSSHKeyOptions{SSHKeyID: tfe.String(plan.SSHKeyID.ValueString())})
		if err != nil {
			resp.Diagnostics.AddError("Error assigning SSH key", fmt.Sprintf("Error assigning SSH key to workspace %s: %v", plan.Name.ValueString(), err))
			return
		}
	}

	if !plan.GlobalRemoteState.IsNull() && !plan.GlobalRemoteState.IsUnknown() && !plan.GlobalRemoteState.ValueBool() && !plan.RemoteStateConsumerIDs.IsNull() && !plan.RemoteStateConsumerIDs.IsUnknown() {
		consumerIDs := []string{}
		resp.Diagnostics.Append(plan.RemoteStateConsumerIDs.ElementsAs(ctx, &consumerIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		if len(consumerIDs) > 0 {
			addOptions := tfe.WorkspaceAddRemoteStateConsumersOptions{}
			for _, id := range consumerIDs {
				addOptions.Workspaces = append(addOptions.Workspaces, &tfe.Workspace{ID: id})
			}
			if err := r.config.Client.Workspaces.AddRemoteStateConsumers(ctx, ws.ID, addOptions); err != nil {
				resp.Diagnostics.AddError("Error adding remote state consumers", fmt.Sprintf("Error adding remote state consumers to workspace %s: %v", plan.Name.ValueString(), err))
				return
			}
		}
	}

	r.readByIDIntoState(ctx, ws.ID, &plan, &resp.State, resp.Identity, &resp.Diagnostics)
}

func (r *resourceTFEWorkspaceFramework) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state modelWorkspace
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.readByIDIntoState(ctx, state.ID.ValueString(), &state, &resp.State, resp.Identity, &resp.Diagnostics)
}

func (r *resourceTFEWorkspaceFramework) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan modelWorkspace
	var state modelWorkspace
	var configAutoDestroyAt types.String
	var configAutoDestroyActivityDuration types.String
	var configExecutionMode types.String
	var configAgentPoolID types.String
	var configOperations types.Bool
	var configTriggerPrefixes types.List
	var configTriggerPatterns types.List
	var configTags types.Map
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("auto_destroy_at"), &configAutoDestroyAt)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("auto_destroy_activity_duration"), &configAutoDestroyActivityDuration)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("execution_mode"), &configExecutionMode)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("agent_pool_id"), &configAgentPoolID)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("operations"), &configOperations)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("trigger_prefixes"), &configTriggerPrefixes)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("trigger_patterns"), &configTriggerPatterns)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("tags"), &configTags)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := tfe.WorkspaceUpdateOptions{
		Name:                       tfe.String(plan.Name.ValueString()),
		AllowDestroyPlan:           tfe.Bool(boolValueOrDefault(plan.AllowDestroyPlan, true)),
		AutoApplyRunTrigger:        tfe.Bool(boolValueOrDefault(plan.AutoApplyRunTrigger, false)),
		FileTriggersEnabled:        tfe.Bool(boolValueOrDefault(plan.FileTriggersEnabled, true)),
		QueueAllRuns:               tfe.Bool(boolValueOrDefault(plan.QueueAllRuns, true)),
		SpeculativeEnabled:         tfe.Bool(boolValueOrDefault(plan.SpeculativeEnabled, true)),
		StructuredRunOutputEnabled: tfe.Bool(boolValueOrDefault(plan.StructuredRunOutputEnabled, true)),
		WorkingDirectory:           tfe.String(stringValueOrDefault(plan.WorkingDirectory, "")),
	}

	if !plan.GlobalRemoteState.IsNull() && !plan.GlobalRemoteState.IsUnknown() {
		opts.GlobalRemoteState = tfe.Bool(plan.GlobalRemoteState.ValueBool())
	}

	r.applyWorkspaceUpdateOptionsFromModel(ctx, &plan, &opts, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if configTriggerPatterns.IsNull() {
		opts.TriggerPatterns = nil
	} else if configTriggerPrefixes.IsNull() {
		opts.TriggerPrefixes = nil
	}

	if !configTags.IsNull() && !configTags.IsUnknown() && len(configTags.Elements()) == 0 && !boolValueOrDefault(plan.IgnoreAdditionalTags, false) {
		if err := r.config.Client.Workspaces.DeleteAllTagBindings(ctx, state.ID.ValueString()); err != nil {
			resp.Diagnostics.AddError("Error removing tag bindings", fmt.Sprintf("Error removing tag bindings from workspace %s: %v", state.ID.ValueString(), err))
			return
		}
	}

	if configExecutionMode.IsNull() || configExecutionMode.IsUnknown() {
		opts.ExecutionMode = nil
		if opts.SettingOverwrites != nil {
			opts.SettingOverwrites.ExecutionMode = nil
		}
	}
	if configAgentPoolID.IsNull() || configAgentPoolID.IsUnknown() || configAgentPoolID.ValueString() == "" {
		opts.AgentPoolID = nil
		if opts.SettingOverwrites != nil {
			opts.SettingOverwrites.AgentPool = nil
		}
	}
	if configOperations.IsNull() || configOperations.IsUnknown() {
		opts.Operations = nil
	}
	if opts.SettingOverwrites != nil && opts.SettingOverwrites.ExecutionMode == nil && opts.SettingOverwrites.AgentPool == nil {
		opts.SettingOverwrites = nil
	}

	if configAutoDestroyAt.IsNull() || configAutoDestroyAt.IsUnknown() || configAutoDestroyAt.ValueString() == "" {
		opts.AutoDestroyAt = jsonapi.NewNullNullableAttr[time.Time]()
	} else {
		t, err := time.Parse(time.RFC3339, configAutoDestroyAt.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error expanding auto destroy", err.Error())
			return
		}
		opts.AutoDestroyAt = jsonapi.NewNullableAttrWithValue(t)
	}

	if configAutoDestroyActivityDuration.IsNull() || configAutoDestroyActivityDuration.IsUnknown() || configAutoDestroyActivityDuration.ValueString() == "" {
		opts.AutoDestroyActivityDuration = jsonapi.NewNullNullableAttr[string]()
	} else {
		opts.AutoDestroyActivityDuration = jsonapi.NewNullableAttrWithValue(configAutoDestroyActivityDuration.ValueString())
	}

	if _, err := r.config.Client.Workspaces.UpdateByID(ctx, state.ID.ValueString(), opts); err != nil {
		resp.Diagnostics.AddError("Error updating workspace", fmt.Sprintf("Error updating workspace %s: %v", state.ID.ValueString(), err))
		return
	}

	if !plan.SSHKeyID.Equal(state.SSHKeyID) {
		if !plan.SSHKeyID.IsNull() && plan.SSHKeyID.ValueString() != "" {
			_, err := r.config.Client.Workspaces.AssignSSHKey(ctx, state.ID.ValueString(), tfe.WorkspaceAssignSSHKeyOptions{SSHKeyID: tfe.String(plan.SSHKeyID.ValueString())})
			if err != nil {
				resp.Diagnostics.AddError("Error assigning SSH key", err.Error())
				return
			}
		} else {
			_, err := r.config.Client.Workspaces.UnassignSSHKey(ctx, state.ID.ValueString())
			if err != nil {
				resp.Diagnostics.AddError("Error unassigning SSH key", err.Error())
				return
			}
		}
	}

	r.syncRemoteStateConsumers(ctx, state.ID.ValueString(), state.RemoteStateConsumerIDs, plan.GlobalRemoteState, plan.RemoteStateConsumerIDs, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	r.readByIDIntoState(ctx, state.ID.ValueString(), &plan, &resp.State, resp.Identity, &resp.Diagnostics)
}

func (r *resourceTFEWorkspaceFramework) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state modelWorkspace
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	ws, err := r.config.Client.Workspaces.ReadByID(ctx, id)
	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
			return
		}
		resp.Diagnostics.AddError("Error reading workspace", fmt.Sprintf("Error reading workspace %s: %v", id, err))
		return
	}

	forceDelete := !state.ForceDelete.IsNull() && state.ForceDelete.ValueBool()
	if ws.Permissions.CanForceDelete == nil {
		if forceDelete {
			err = r.config.Client.Workspaces.DeleteByID(ctx, id)
		} else {
			resp.Diagnostics.AddError("Error deleting workspace", fmt.Sprintf("This version of Terraform Enterprise does not support workspace safe-delete. Workspaces must be force deleted by setting force_delete=true (workspace %s)", id))
			return
		}
	} else if *ws.Permissions.CanForceDelete {
		if forceDelete {
			err = r.config.Client.Workspaces.DeleteByID(ctx, id)
		} else {
			err = errWorkspaceResourceCountCheck(id, ws.ResourceCount)
			if err == nil {
				err = errWorkspaceSafeDeleteWithPermission(id, safeWorkspaceDelete(ctx, r.config, id))
			}
		}
	} else {
		if forceDelete {
			resp.Diagnostics.AddError("Error deleting workspace", fmt.Sprintf("missing required permissions to set force delete workspaces in the organization for workspace %s", id))
			return
		}
		err = errWorkspaceResourceCountCheck(id, ws.ResourceCount)
		if err == nil {
			err = safeWorkspaceDelete(ctx, r.config, id)
		}
	}

	if err != nil && !errors.Is(err, tfe.ErrResourceNotFound) {
		resp.Diagnostics.AddError("Error deleting workspace", err.Error())
	}
}

func (r *resourceTFEWorkspaceFramework) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var importID string

	if req.ID != "" {
		s := strings.Split(req.ID, "/")
		if len(s) >= 3 {
			resp.Diagnostics.AddError("Invalid import format", fmt.Sprintf("invalid workspace input format: %s (expected <ORGANIZATION>/<WORKSPACE NAME> or <WORKSPACE ID>)", req.ID))
			return
		}
		if len(s) == 2 {
			workspaceID, err := fetchWorkspaceExternalID(s[0]+"/"+s[1], r.config.Client)
			if err != nil {
				resp.Diagnostics.AddError("Error importing workspace", fmt.Sprintf("error retrieving workspace with name %s from organization %s: %v", s[1], s[0], err))
				return
			}
			importID = workspaceID
		} else {
			importID = req.ID
		}
	} else {
		if req.Identity == nil {
			resp.Diagnostics.AddError("Error importing workspace", "missing identity for import")
			return
		}
		var identity modelWorkspaceIdentity
		resp.Diagnostics.Append(req.Identity.Get(ctx, &identity)...)
		if resp.Diagnostics.HasError() {
			return
		}
		if identity.ID.IsNull() || identity.ID.IsUnknown() || identity.ID.ValueString() == "" {
			resp.Diagnostics.AddError("Error importing workspace", "identity.id must be set for workspace import")
			return
		}
		importID = identity.ID.ValueString()
	}

	prior := modelWorkspace{}
	r.readByIDIntoState(ctx, importID, &prior, &resp.State, nil, &resp.Diagnostics)
}

func (r *resourceTFEWorkspaceFramework) readByIDIntoState(ctx context.Context, id string, prior *modelWorkspace, state *tfsdk.State, identity *tfsdk.ResourceIdentity, diags *diag.Diagnostics) {
	tflog.Debug(ctx, "Read configuration of workspace", map[string]any{"id": id})

	workspace, err := r.config.Client.Workspaces.ReadByIDWithOptions(ctx, id, &tfe.WorkspaceReadOptions{Include: []tfe.WSIncludeOpt{tfe.WSEffectiveTagBindings}})
	if err != nil && errors.Is(err, tfe.ErrInvalidIncludeValue) {
		workspace, err = r.config.Client.Workspaces.ReadByID(ctx, id)
	}
	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
			state.RemoveResource(ctx)
			return
		}
		diags.AddError("Error reading workspace", fmt.Sprintf("Error reading configuration of workspace %s: %v", id, err))
		return
	}

	model := *prior
	model.ID = types.StringValue(workspace.ID)
	model.Name = types.StringValue(workspace.Name)
	model.AllowDestroyPlan = types.BoolValue(workspace.AllowDestroyPlan)
	model.AssessmentsEnabled = types.BoolValue(workspace.AssessmentsEnabled)
	model.AutoApply = types.BoolValue(workspace.AutoApply)
	model.AutoApplyRunTrigger = types.BoolValue(workspace.AutoApplyRunTrigger)
	model.Description = stringToFramework(workspace.Description)
	model.FileTriggersEnabled = types.BoolValue(workspace.FileTriggersEnabled)
	model.Operations = types.BoolValue(workspace.Operations)
	model.ExecutionMode = stringToFramework(workspace.ExecutionMode)
	model.QueueAllRuns = types.BoolValue(workspace.QueueAllRuns)
	model.SourceName = stringToFramework(workspace.SourceName)
	model.SourceURL = stringToFramework(workspace.SourceURL)
	model.SpeculativeEnabled = types.BoolValue(workspace.SpeculativeEnabled)
	model.StructuredRunOutputEnabled = types.BoolValue(workspace.StructuredRunOutputEnabled)
	model.TerraformVersion = stringToFramework(workspace.TerraformVersion)
	model.WorkingDirectory = types.StringValue(workspace.WorkingDirectory)
	model.Organization = types.StringValue(workspace.Organization.Name)
	model.ResourceCount = types.Int64Value(int64(workspace.ResourceCount))
	model.InheritsProjectAutoDestroy = types.BoolValue(workspace.InheritsProjectAutoDestroy)
	if workspace.HYOKEnabled == nil {
		model.HYOKEnabled = types.BoolNull()
	} else {
		model.HYOKEnabled = types.BoolValue(*workspace.HYOKEnabled)
	}
	model.TriggerPrefixes = stringSliceToList(workspace.TriggerPrefixes)
	model.TriggerPatterns = stringSliceToList(workspace.TriggerPatterns)

	if workspace.Project != nil {
		model.ProjectID = types.StringValue(workspace.Project.ID)
	} else {
		model.ProjectID = types.StringNull()
	}

	if workspace.SSHKey != nil {
		model.SSHKeyID = types.StringValue(workspace.SSHKey.ID)
	} else {
		model.SSHKeyID = types.StringValue("")
	}

	if workspace.AgentPool != nil {
		model.AgentPoolID = types.StringValue(workspace.AgentPool.ID)
	} else {
		model.AgentPoolID = types.StringValue("")
	}

	if workspace.InheritsProjectAutoDestroy {
		model.AutoDestroyAt = types.StringValue("")
		model.AutoDestroyActivityDuration = types.StringValue("")
	} else {
		autoDestroyAt, err := flattenAutoDestroyAt(workspace.AutoDestroyAt)
		if err != nil {
			diags.AddError("Error flattening auto destroy", err.Error())
			return
		}
		if autoDestroyAt == nil {
			model.AutoDestroyAt = types.StringValue("")
		} else {
			model.AutoDestroyAt = types.StringValue(*autoDestroyAt)
		}

		if workspace.AutoDestroyActivityDuration.IsSpecified() {
			v, err := workspace.AutoDestroyActivityDuration.Get()
			if err != nil {
				diags.AddError("Error reading auto destroy activity duration", err.Error())
				return
			}
			model.AutoDestroyActivityDuration = types.StringValue(v)
		} else {
			model.AutoDestroyActivityDuration = types.StringValue("")
		}
	}

	managedTagNames := map[string]struct{}{}
	if !prior.TagNames.IsNull() && !prior.TagNames.IsUnknown() {
		var current []string
		diags.Append(prior.TagNames.ElementsAs(ctx, &current, false)...)
		for _, t := range current {
			managedTagNames[t] = struct{}{}
		}
	}

	tagNames := make([]attr.Value, 0, len(workspace.TagNames))
	ignoreAdditionalTagNames := !prior.IgnoreAdditionalTagNames.IsNull() && !prior.IgnoreAdditionalTagNames.IsUnknown() && prior.IgnoreAdditionalTagNames.ValueBool()
	for _, tagName := range workspace.TagNames {
		_, isManaged := managedTagNames[tagName]
		if isManaged || !ignoreAdditionalTagNames {
			tagNames = append(tagNames, types.StringValue(tagName))
		}
	}
	if !prior.TagNames.IsNull() && !prior.TagNames.IsUnknown() {
		model.TagNames = prior.TagNames
	} else {
		model.TagNames = types.SetValueMust(types.StringType, tagNames)
	}

	if workspace.VCSRepo != nil {
		vcsRepoVal, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: map[string]attr.Type{
			"identifier":                 types.StringType,
			"branch":                     types.StringType,
			"ingress_submodules":         types.BoolType,
			"oauth_token_id":             types.StringType,
			"tags_regex":                 types.StringType,
			"github_app_installation_id": types.StringType,
		}}, []modelWorkspaceVCSRepo{{
			Identifier:              stringToFramework(workspace.VCSRepo.Identifier),
			Branch:                  types.StringValue(workspace.VCSRepo.Branch),
			IngressSubmodules:       types.BoolValue(workspace.VCSRepo.IngressSubmodules),
			OAuthTokenID:            types.StringValue(workspace.VCSRepo.OAuthTokenID),
			TagsRegex:               types.StringValue(workspace.VCSRepo.TagsRegex),
			GithubAppInstallationID: types.StringValue(workspace.VCSRepo.GHAInstallationID),
		}})
		diags.Append(d...)
		model.VCSRepo = vcsRepoVal
	} else {
		model.VCSRepo = types.ListNull(types.ObjectType{AttrTypes: map[string]attr.Type{
			"identifier":                 types.StringType,
			"branch":                     types.StringType,
			"ingress_submodules":         types.BoolType,
			"oauth_token_id":             types.StringType,
			"tags_regex":                 types.StringType,
			"github_app_installation_id": types.StringType,
		}})
	}

	tagInfo := helpers.NewTagInfo(mapFromStringMapType(prior.Tags), workspace.EffectiveTagBindings, !prior.IgnoreAdditionalTags.IsNull() && !prior.IgnoreAdditionalTags.IsUnknown() && prior.IgnoreAdditionalTags.ValueBool())
	model.Tags = mapTypeFromStringMap(tagInfo.SelfTags)
	model.EffectiveTags = mapTypeFromStringMap(tagInfo.EffectiveTags)

	if workspace.GlobalRemoteState {
		model.GlobalRemoteState = types.BoolValue(true)
		model.RemoteStateConsumerIDs = types.SetValueMust(types.StringType, []attr.Value{})
	} else {
		globalRemoteState, remoteStateConsumerIDs, err := readWorkspaceStateConsumers(id, r.config.Client)
		if err != nil {
			diags.AddError("Error reading remote state consumers", fmt.Sprintf("Error reading remote state consumers for workspace %s: %v", id, err))
			return
		}
		model.GlobalRemoteState = types.BoolValue(globalRemoteState)
		model.RemoteStateConsumerIDs = stringSliceToSet(remoteStateConsumerIDs)
	}

	if workspace.Links["self-html"] != nil {
		baseAPI := r.config.Client.BaseURL()
		htmlURL := url.URL{Scheme: baseAPI.Scheme, Host: baseAPI.Host, Path: workspace.Links["self-html"].(string)}
		model.HTMLURL = types.StringValue(htmlURL.String())
	}

	if prior.ForceDelete.IsNull() || prior.ForceDelete.IsUnknown() {
		model.ForceDelete = types.BoolValue(false)
	} else {
		model.ForceDelete = prior.ForceDelete
	}

	diags.Append(state.Set(ctx, &model)...)

	if identity != nil {
		identityModel := modelWorkspaceIdentity{ID: model.ID, Hostname: types.StringValue(r.config.Client.BaseURL().Host)}
		diags.Append(identity.Set(ctx, &identityModel)...)
	}
}

func (r *resourceTFEWorkspaceFramework) applyWorkspaceOptionsFromModel(ctx context.Context, model *modelWorkspace, options *tfe.WorkspaceCreateOptions, diags *diag.Diagnostics) {
	if !model.GlobalRemoteState.IsNull() && !model.GlobalRemoteState.IsUnknown() {
		options.GlobalRemoteState = tfe.Bool(model.GlobalRemoteState.ValueBool())
	}
	if !model.AutoApply.IsNull() && !model.AutoApply.IsUnknown() {
		options.AutoApply = tfe.Bool(model.AutoApply.ValueBool())
	}
	if !model.AssessmentsEnabled.IsNull() && !model.AssessmentsEnabled.IsUnknown() {
		options.AssessmentsEnabled = tfe.Bool(model.AssessmentsEnabled.ValueBool())
	}
	if !model.Description.IsNull() && !model.Description.IsUnknown() {
		options.Description = tfe.String(model.Description.ValueString())
	}
	if !model.AgentPoolID.IsNull() && !model.AgentPoolID.IsUnknown() && model.AgentPoolID.ValueString() != "" {
		options.AgentPoolID = tfe.String(model.AgentPoolID.ValueString())
		options.SettingOverwrites = &tfe.WorkspaceSettingOverwritesOptions{ExecutionMode: tfe.Bool(true), AgentPool: tfe.Bool(true)}
	}
	if !model.AutoDestroyAt.IsNull() && !model.AutoDestroyAt.IsUnknown() && model.AutoDestroyAt.ValueString() != "" {
		t, err := time.Parse(time.RFC3339, model.AutoDestroyAt.ValueString())
		if err != nil {
			diags.AddError("Error expanding auto destroy", err.Error())
			return
		}
		options.AutoDestroyAt = jsonapi.NewNullableAttrWithValue(t)
	}
	if !model.AutoDestroyActivityDuration.IsNull() && !model.AutoDestroyActivityDuration.IsUnknown() {
		options.AutoDestroyActivityDuration = jsonapi.NewNullableAttrWithValue(model.AutoDestroyActivityDuration.ValueString())
	}
	if !model.ExecutionMode.IsNull() && !model.ExecutionMode.IsUnknown() {
		options.ExecutionMode = tfe.String(model.ExecutionMode.ValueString())
		options.SettingOverwrites = &tfe.WorkspaceSettingOverwritesOptions{ExecutionMode: tfe.Bool(true), AgentPool: tfe.Bool(true)}
	}
	if !model.Operations.IsNull() && !model.Operations.IsUnknown() {
		options.Operations = tfe.Bool(model.Operations.ValueBool())
		options.SettingOverwrites = &tfe.WorkspaceSettingOverwritesOptions{ExecutionMode: tfe.Bool(true), AgentPool: tfe.Bool(true)}
	}
	if options.SettingOverwrites == nil {
		options.SettingOverwrites = &tfe.WorkspaceSettingOverwritesOptions{ExecutionMode: tfe.Bool(false), AgentPool: tfe.Bool(false)}
	}
	if !model.SourceURL.IsNull() && !model.SourceURL.IsUnknown() {
		options.SourceURL = tfe.String(model.SourceURL.ValueString())
	}
	if !model.SourceName.IsNull() && !model.SourceName.IsUnknown() {
		options.SourceName = tfe.String(model.SourceName.ValueString())
	}

	if !model.Tags.IsNull() && !model.Tags.IsUnknown() {
		options.TagBindings = []*tfe.TagBinding{}
		options.TagBindings = append(options.TagBindings, expandWorkspaceTagBindings(model.Tags)...)
	}

	if !model.TerraformVersion.IsNull() && !model.TerraformVersion.IsUnknown() {
		options.TerraformVersion = tfe.String(model.TerraformVersion.ValueString())
	}
	vcsConfigured := !model.VCSRepo.IsNull() && !model.VCSRepo.IsUnknown()
	if prefixes, ok := expandWorkspaceStringList(ctx, model.TriggerPrefixes, diags); ok && (vcsConfigured || len(prefixes) > 0) {
		options.TriggerPrefixes = prefixes
	}
	if patterns, ok := expandWorkspaceStringList(ctx, model.TriggerPatterns, diags); ok && (vcsConfigured || len(patterns) > 0) {
		options.TriggerPatterns = patterns
	}
	if !model.ProjectID.IsNull() && !model.ProjectID.IsUnknown() && model.ProjectID.ValueString() != "" {
		options.Project = &tfe.Project{ID: model.ProjectID.ValueString()}
	}
	options.VCSRepo = expandWorkspaceVCSRepoOptions(ctx, model.VCSRepo, diags, false)
	if tags, ok := expandWorkspaceTagNames(ctx, model.TagNames, diags); ok {
		options.Tags = append(options.Tags, tags...)
	}
}

func (r *resourceTFEWorkspaceFramework) applyWorkspaceUpdateOptionsFromModel(ctx context.Context, model *modelWorkspace, options *tfe.WorkspaceUpdateOptions, diags *diag.Diagnostics) {
	if !model.ProjectID.IsNull() && !model.ProjectID.IsUnknown() && model.ProjectID.ValueString() != "" {
		options.Project = &tfe.Project{ID: model.ProjectID.ValueString()}
	}
	if !model.AssessmentsEnabled.IsNull() && !model.AssessmentsEnabled.IsUnknown() {
		options.AssessmentsEnabled = tfe.Bool(model.AssessmentsEnabled.ValueBool())
	}
	if !model.AutoApply.IsNull() && !model.AutoApply.IsUnknown() {
		options.AutoApply = tfe.Bool(model.AutoApply.ValueBool())
	}
	if !model.Description.IsNull() && !model.Description.IsUnknown() {
		options.Description = tfe.String(model.Description.ValueString())
	}
	if !model.AgentPoolID.IsNull() && !model.AgentPoolID.IsUnknown() && model.AgentPoolID.ValueString() != "" {
		options.AgentPoolID = tfe.String(model.AgentPoolID.ValueString())
		options.SettingOverwrites = &tfe.WorkspaceSettingOverwritesOptions{AgentPool: tfe.Bool(true)}
	}
	if !model.AutoDestroyAt.IsNull() && !model.AutoDestroyAt.IsUnknown() {
		if model.AutoDestroyAt.ValueString() == "" {
			options.AutoDestroyAt = jsonapi.NewNullNullableAttr[time.Time]()
		} else {
			t, err := time.Parse(time.RFC3339, model.AutoDestroyAt.ValueString())
			if err != nil {
				diags.AddError("Error expanding auto destroy", err.Error())
				return
			}
			options.AutoDestroyAt = jsonapi.NewNullableAttrWithValue(t)
		}
	}
	if model.AutoDestroyActivityDuration.IsNull() || model.AutoDestroyActivityDuration.IsUnknown() || model.AutoDestroyActivityDuration.ValueString() == "" {
		options.AutoDestroyActivityDuration = jsonapi.NewNullNullableAttr[string]()
	} else {
		options.AutoDestroyActivityDuration = jsonapi.NewNullableAttrWithValue(model.AutoDestroyActivityDuration.ValueString())
	}
	if !model.ExecutionMode.IsNull() && !model.ExecutionMode.IsUnknown() {
		options.ExecutionMode = tfe.String(model.ExecutionMode.ValueString())
		options.SettingOverwrites = &tfe.WorkspaceSettingOverwritesOptions{ExecutionMode: tfe.Bool(true)}
	}
	if !model.Operations.IsNull() && !model.Operations.IsUnknown() {
		options.Operations = tfe.Bool(model.Operations.ValueBool())
	}

	if !model.Tags.IsNull() && !model.Tags.IsUnknown() {
		options.TagBindings = []*tfe.TagBinding{}
		options.TagBindings = append(options.TagBindings, expandWorkspaceTagBindings(model.Tags)...)
	}

	if !model.TerraformVersion.IsNull() && !model.TerraformVersion.IsUnknown() {
		options.TerraformVersion = tfe.String(model.TerraformVersion.ValueString())
	}
	vcsConfigured := !model.VCSRepo.IsNull() && !model.VCSRepo.IsUnknown()
	if prefixes, ok := expandWorkspaceStringList(ctx, model.TriggerPrefixes, diags); ok && (vcsConfigured || len(prefixes) > 0) {
		options.TriggerPrefixes = prefixes
	}
	if patterns, ok := expandWorkspaceStringList(ctx, model.TriggerPatterns, diags); ok && (vcsConfigured || len(patterns) > 0) {
		options.TriggerPatterns = patterns
	}
	if !model.WorkingDirectory.IsNull() && !model.WorkingDirectory.IsUnknown() {
		options.WorkingDirectory = tfe.String(model.WorkingDirectory.ValueString())
	}
	options.VCSRepo = expandWorkspaceVCSRepoOptions(ctx, model.VCSRepo, diags, true)
}

func (r *resourceTFEWorkspaceFramework) syncRemoteStateConsumers(ctx context.Context, workspaceID string, oldSet types.Set, newGlobal types.Bool, newSet types.Set, diags *diag.Diagnostics) {
	if !newGlobal.IsNull() && newGlobal.ValueBool() {
		return
	}

	oldIDs := []string{}
	newIDs := []string{}
	if !oldSet.IsNull() && !oldSet.IsUnknown() {
		diags.Append(oldSet.ElementsAs(ctx, &oldIDs, false)...)
	}
	if !newSet.IsNull() && !newSet.IsUnknown() {
		diags.Append(newSet.ElementsAs(ctx, &newIDs, false)...)
	}
	if diags.HasError() {
		return
	}

	oldMap := map[string]struct{}{}
	newMap := map[string]struct{}{}
	for _, id := range oldIDs {
		oldMap[id] = struct{}{}
	}
	for _, id := range newIDs {
		newMap[id] = struct{}{}
	}

	toAdd := make([]string, 0)
	toRemove := make([]string, 0)
	for id := range newMap {
		if _, ok := oldMap[id]; !ok {
			toAdd = append(toAdd, id)
		}
	}
	for id := range oldMap {
		if _, ok := newMap[id]; !ok {
			toRemove = append(toRemove, id)
		}
	}

	if len(toAdd) > 0 {
		opts := tfe.WorkspaceAddRemoteStateConsumersOptions{}
		for _, id := range toAdd {
			opts.Workspaces = append(opts.Workspaces, &tfe.Workspace{ID: id})
		}
		if err := r.config.Client.Workspaces.AddRemoteStateConsumers(ctx, workspaceID, opts); err != nil {
			diags.AddError("Error adding remote state consumers", err.Error())
			return
		}
	}

	if len(toRemove) > 0 {
		opts := tfe.WorkspaceRemoveRemoteStateConsumersOptions{}
		for _, id := range toRemove {
			opts.Workspaces = append(opts.Workspaces, &tfe.Workspace{ID: id})
		}
		if err := r.config.Client.Workspaces.RemoveRemoteStateConsumers(ctx, workspaceID, opts); err != nil {
			diags.AddError("Error removing remote state consumers", err.Error())
			return
		}
	}
}
