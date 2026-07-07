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

var (
	_ datasource.DataSource              = &dataSourceTFEWorkspace{}
	_ datasource.DataSourceWithConfigure = &dataSourceTFEWorkspace{}
)

var workspaceVCSRepoObjectType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"identifier":                 types.StringType,
		"branch":                     types.StringType,
		"ingress_submodules":         types.BoolType,
		"oauth_token_id":             types.StringType,
		"tags_regex":                 types.StringType,
		"github_app_installation_id": types.StringType,
	},
}

func NewWorkspaceDataSource() datasource.DataSource {
	return &dataSourceTFEWorkspace{}
}

type dataSourceTFEWorkspace struct {
	config ConfiguredClient
}

type modelDataSourceTFEWorkspaceVCSRepo struct {
	Identifier              types.String `tfsdk:"identifier"`
	Branch                  types.String `tfsdk:"branch"`
	IngressSubmodules       types.Bool   `tfsdk:"ingress_submodules"`
	OAuthTokenID            types.String `tfsdk:"oauth_token_id"`
	TagsRegex               types.String `tfsdk:"tags_regex"`
	GithubAppInstallationID types.String `tfsdk:"github_app_installation_id"`
}

type modelDataSourceTFEWorkspace struct {
	ID                          types.String                         `tfsdk:"id"`
	Name                        types.String                         `tfsdk:"name"`
	Organization                types.String                         `tfsdk:"organization"`
	Description                 types.String                         `tfsdk:"description"`
	AllowDestroyPlan            types.Bool                           `tfsdk:"allow_destroy_plan"`
	AutoApply                   types.Bool                           `tfsdk:"auto_apply"`
	AutoApplyRunTrigger         types.Bool                           `tfsdk:"auto_apply_run_trigger"`
	AutoDestroyAt               types.String                         `tfsdk:"auto_destroy_at"`
	AutoDestroyActivityDuration types.String                         `tfsdk:"auto_destroy_activity_duration"`
	InheritsProjectAutoDestroy  types.Bool                           `tfsdk:"inherits_project_auto_destroy"`
	FileTriggersEnabled         types.Bool                           `tfsdk:"file_triggers_enabled"`
	GlobalRemoteState           types.Bool                           `tfsdk:"global_remote_state"`
	ProjectRemoteState          types.Bool                           `tfsdk:"project_remote_state"`
	RemoteStateConsumerIDs      types.Set                            `tfsdk:"remote_state_consumer_ids"`
	AssessmentsEnabled          types.Bool                           `tfsdk:"assessments_enabled"`
	Operations                  types.Bool                           `tfsdk:"operations"`
	PolicyCheckFailures         types.Int64                          `tfsdk:"policy_check_failures"`
	ProjectID                   types.String                         `tfsdk:"project_id"`
	QueueAllRuns                types.Bool                           `tfsdk:"queue_all_runs"`
	ResourceCount               types.Int64                          `tfsdk:"resource_count"`
	RunFailures                 types.Int64                          `tfsdk:"run_failures"`
	RunsCount                   types.Int64                          `tfsdk:"runs_count"`
	SourceName                  types.String                         `tfsdk:"source_name"`
	SourceURL                   types.String                         `tfsdk:"source_url"`
	SpeculativeEnabled          types.Bool                           `tfsdk:"speculative_enabled"`
	SSHKeyID                    types.String                         `tfsdk:"ssh_key_id"`
	StructuredRunOutputEnabled  types.Bool                           `tfsdk:"structured_run_output_enabled"`
	EffectiveTags               types.Map                            `tfsdk:"effective_tags"`
	TagNames                    types.Set                            `tfsdk:"tag_names"`
	TerraformVersion            types.String                         `tfsdk:"terraform_version"`
	TriggerPrefixes             types.List                           `tfsdk:"trigger_prefixes"`
	TriggerPatterns             types.List                           `tfsdk:"trigger_patterns"`
	WorkingDirectory            types.String                         `tfsdk:"working_directory"`
	ExecutionMode               types.String                         `tfsdk:"execution_mode"`
	VCSRepo                     []modelDataSourceTFEWorkspaceVCSRepo `tfsdk:"vcs_repo"`
	HTMLURL                     types.String                         `tfsdk:"html_url"`
	HYOKEnabled                 types.Bool                           `tfsdk:"hyok_enabled"`
	Locked                      types.Bool                           `tfsdk:"locked"`
	CreatedAt                   types.String                         `tfsdk:"created_at"`
	UpdatedAt                   types.String                         `tfsdk:"updated_at"`
	Environment                 types.String                         `tfsdk:"environment"`
	ApplyDurationAverage        types.Int64                          `tfsdk:"apply_duration_average"`
	PlanDurationAverage         types.Int64                          `tfsdk:"plan_duration_average"`
	Source                      types.String                         `tfsdk:"source"`
	SettingOverwrites           types.Map                            `tfsdk:"setting_overwrites"`
	Permissions                 types.Map                            `tfsdk:"permissions"`
	Actions                     types.Map                            `tfsdk:"actions"`
}

func (d *dataSourceTFEWorkspace) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workspace"
}

func (d *dataSourceTFEWorkspace) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":                             schema.StringAttribute{Computed: true},
			"name":                           schema.StringAttribute{Required: true},
			"organization":                   schema.StringAttribute{Optional: true, Computed: true},
			"description":                    schema.StringAttribute{Computed: true},
			"allow_destroy_plan":             schema.BoolAttribute{Computed: true},
			"auto_apply":                     schema.BoolAttribute{Computed: true},
			"auto_apply_run_trigger":         schema.BoolAttribute{Computed: true},
			"auto_destroy_at":                schema.StringAttribute{Computed: true},
			"auto_destroy_activity_duration": schema.StringAttribute{Computed: true},
			"inherits_project_auto_destroy":  schema.BoolAttribute{Computed: true},
			"file_triggers_enabled":          schema.BoolAttribute{Computed: true},
			"global_remote_state":            schema.BoolAttribute{Computed: true},
			"project_remote_state":           schema.BoolAttribute{Computed: true},
			"remote_state_consumer_ids":      schema.SetAttribute{Computed: true, ElementType: types.StringType},
			"assessments_enabled":            schema.BoolAttribute{Computed: true},
			"operations":                     schema.BoolAttribute{Computed: true},
			"policy_check_failures":          schema.Int64Attribute{Computed: true},
			"project_id":                     schema.StringAttribute{Computed: true},
			"queue_all_runs":                 schema.BoolAttribute{Computed: true},
			"resource_count":                 schema.Int64Attribute{Computed: true},
			"run_failures":                   schema.Int64Attribute{Computed: true},
			"runs_count":                     schema.Int64Attribute{Computed: true},
			"source_name":                    schema.StringAttribute{Computed: true},
			"source_url":                     schema.StringAttribute{Computed: true},
			"speculative_enabled":            schema.BoolAttribute{Computed: true},
			"ssh_key_id":                     schema.StringAttribute{Computed: true},
			"structured_run_output_enabled":  schema.BoolAttribute{Computed: true},
			"effective_tags":                 schema.MapAttribute{Computed: true, ElementType: types.StringType},
			"tag_names":                      schema.SetAttribute{Optional: true, Computed: true, ElementType: types.StringType},
			"terraform_version":              schema.StringAttribute{Computed: true},
			"trigger_prefixes":               schema.ListAttribute{Computed: true, ElementType: types.StringType},
			"trigger_patterns":               schema.ListAttribute{Computed: true, ElementType: types.StringType},
			"working_directory":              schema.StringAttribute{Computed: true},
			"execution_mode":                 schema.StringAttribute{Computed: true},
			"vcs_repo":                       schema.ListAttribute{Computed: true, ElementType: workspaceVCSRepoObjectType},
			"html_url":                       schema.StringAttribute{Computed: true},
			"hyok_enabled":                   schema.BoolAttribute{Computed: true},
			"locked":                         schema.BoolAttribute{Computed: true},
			"created_at":                     schema.StringAttribute{Computed: true},
			"updated_at":                     schema.StringAttribute{Computed: true},
			"environment":                    schema.StringAttribute{Computed: true},
			"apply_duration_average":         schema.Int64Attribute{Computed: true},
			"plan_duration_average":          schema.Int64Attribute{Computed: true},
			"source":                         schema.StringAttribute{Computed: true},
			"setting_overwrites":             schema.MapAttribute{Computed: true, ElementType: types.BoolType},
			"permissions":                    schema.MapAttribute{Computed: true, ElementType: types.BoolType},
			"actions":                        schema.MapAttribute{Computed: true, ElementType: types.BoolType},
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
