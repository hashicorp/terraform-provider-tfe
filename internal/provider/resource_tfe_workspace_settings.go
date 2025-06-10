// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"log"
	"slices"
	"strings"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// tfe_workspace_settings resource
var _ resource.Resource = &workspaceSettings{}

// overwritesElementType is the object type definition for the
// overwrites field schema.
var overwritesElementType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"execution_mode": types.BoolType,
		"agent_pool":     types.BoolType,
	},
}

type workspaceSettings struct {
	config             ConfiguredClient
	supportsOverwrites bool
}

type modelWorkspaceSettings struct {
	ID                     types.String `tfsdk:"id"`
	WorkspaceID            types.String `tfsdk:"workspace_id"`
	ExecutionMode          types.String `tfsdk:"execution_mode"`
	AgentPoolID            types.String `tfsdk:"agent_pool_id"`
	Overwrites             types.List   `tfsdk:"overwrites"`
	GlobalRemoteState      types.Bool   `tfsdk:"global_remote_state"`
	RemoteStateConsumerIDs types.Set    `tfsdk:"remote_state_consumer_ids"`
	Description            types.String `tfsdk:"description"`
	AutoApply              types.Bool   `tfsdk:"auto_apply"`
	AssessmentsEnabled     types.Bool   `tfsdk:"assessments_enabled"`
}

type modelOverwrites struct {
	ExecutionMode types.Bool `tfsdk:"execution_mode"`
	AgentPool     types.Bool `tfsdk:"agent_pool"`
}

// errWorkspaceNoLongerExists is returned when reading the workspace settings but
// the workspace no longer exists.
var errWorkspaceNoLongerExists = errors.New("workspace no longer exists")

// validateAgentExecutionMode is a PlanModifier that validates that the combination
// of "execution_mode" and "agent_pool_id" is compatible.
type validateAgentExecutionMode struct{}

// revertOverwritesIfExecutionModeUnset is a PlanModifier for "overwrites" that
// sets the values to false if execution_mode is unset. This tells the server to
// compute execution_mode and agent_pool_id if defaults are set. This
// modifier must be used in conjunction with unknownIfExecutionModeUnset plan
// modifier on the execution_mode and agent_pool_id fields.
type revertOverwritesIfExecutionModeUnset struct{}

// unknownIfExecutionModeUnset sets the planned value to (known after apply) if
// execution_mode is unset, avoiding an inconsistent state after the apply. This
// allows the server to compute the new value based on the default. It should be
// applied to both execution_mode and agent_pool_id in conjunction with
// revertOverwritesIfExecutionModeUnset.
type unknownIfExecutionModeUnset struct{}

// validateRemoteStateConsumerIDs validates that if global_remote_state is
// true, remote_state_consumer_ids is not set.
type validateRemoteStateConsumerIDs struct{}

// validateSelfReference validates that the workspace ID is not in the set of
// remote state consumers.
type validateSelfReference struct{}

var _ planmodifier.String = (*validateAgentExecutionMode)(nil)
var _ planmodifier.List = (*revertOverwritesIfExecutionModeUnset)(nil)
var _ planmodifier.String = (*unknownIfExecutionModeUnset)(nil)
var _ planmodifier.Set = (*validateRemoteStateConsumerIDs)(nil)

func (m validateAgentExecutionMode) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// Check if the resource is being created.
	if req.State.Raw.IsNull() {
		return
	}

	configured := modelWorkspaceSettings{}
	resp.Diagnostics.Append(req.Config.Get(ctx, &configured)...)

	if configured.ExecutionMode.ValueString() == "agent" && configured.AgentPoolID.IsNull() {
		resp.Diagnostics.AddError("Invalid agent_pool_id", "If execution mode is \"agent\", \"agent_pool_id\" is required")
	}

	if configured.ExecutionMode.ValueString() != "agent" && !configured.AgentPoolID.IsNull() {
		resp.Diagnostics.AddError("Invalid agent_pool_id", "If execution mode is not \"agent\", \"agent_pool_id\" must not be set")
	}
}

func (m validateAgentExecutionMode) Description(_ context.Context) string {
	return "Validates that configuration values for \"agent_pool_id\" and \"execution_mode\" are compatible"
}

func (m validateAgentExecutionMode) MarkdownDescription(_ context.Context) string {
	return "Validates that configuration values for \"agent_pool_id\" and \"execution_mode\" are compatible"
}

func (m validateRemoteStateConsumerIDs) PlanModifySet(_ context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
	var remoteStateConsumerIDs types.Set
	diags := req.Config.GetAttribute(ctx, path.Root("remote_state_consumer_ids"), &remoteStateConsumerIDs)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	if remoteStateConsumerIDs.IsNull() || len(remoteStateConsumerIDs.Elements()) == 0 {
		return
	}

	// This situation is invalid if global_remote_state is true
	var globalRemoteState types.Bool
	diags = req.Config.GetAttribute(ctx, path.Root("global_remote_state"), &globalRemoteState)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	log.Printf("[DEBUG] planned global_remote_state: %v", globalRemoteState.ValueBool())

	if !globalRemoteState.IsNull() && globalRemoteState.ValueBool() {
		resp.Diagnostics.AddError("Invalid remote_state_consumer_ids", "If global_remote_state is true, remote_state_consumer_ids must not be set")
	}
}

func (m validateRemoteStateConsumerIDs) Description(_ context.Context) string {
	return "Validates that configuration values for \"global_remote_state\" and \"remote_state_consumer_ids\" are compatible"
}

func (m validateRemoteStateConsumerIDs) MarkdownDescription(_ context.Context) string {
	return "Validates that configuration values for \"global_remote_state\" and \"remote_state_consumer_ids\" are compatible"
}

func (m validateSelfReference) PlanModifySet(_ context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
	var remoteStateConsumerIDSet types.Set
	diags := req.Plan.GetAttribute(ctx, path.Root("remote_state_consumer_ids"), &remoteStateConsumerIDSet)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	if remoteStateConsumerIDSet.IsNull() || len(remoteStateConsumerIDSet.Elements()) == 0 {
		return
	}

	remoteStateConsumerIDs := make([]string, 0)
	diags = remoteStateConsumerIDSet.ElementsAs(ctx, &remoteStateConsumerIDs, true)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	var workspaceID types.String
	diags = req.Config.GetAttribute(ctx, path.Root("workspace_id"), &workspaceID)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Check if the workspace ID is in the set
	if !workspaceID.IsUnknown() && slices.Contains(remoteStateConsumerIDs, workspaceID.ValueString()) {
		resp.Diagnostics.AddError("Invalid remote_state_consumer_ids", "workspace_id cannot be in the set of remote_state_consumer_ids")
	}
}

func (m validateSelfReference) Description(_ context.Context) string {
	return "Validates that configuration values for \"remote_state_consumer_ids\" does not include the workspace ID"
}

func (m validateSelfReference) MarkdownDescription(_ context.Context) string {
	return "Validates that configuration values for \"remote_state_consumer_ids\" does not include the workspace ID"
}

func (m revertOverwritesIfExecutionModeUnset) PlanModifyList(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	// Check if the resource is being created.
	if req.State.Raw.IsNull() {
		return
	}

	// Determine if configured execution_mode is being unset
	state := modelWorkspaceSettings{}
	configured := modelWorkspaceSettings{}

	resp.Diagnostics.Append(req.Config.Get(ctx, &configured)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	// Check if overwrites are supported by the platform
	if state.Overwrites.IsNull() {
		return
	}

	overwritesState := make([]modelOverwrites, 1)
	state.Overwrites.ElementsAs(ctx, &overwritesState, true)

	if configured.ExecutionMode.IsNull() && overwritesState[0].ExecutionMode.ValueBool() {
		overwritesState[0].AgentPool = types.BoolValue(false)
		overwritesState[0].ExecutionMode = types.BoolValue(false)

		newList, diags := types.ListValueFrom(ctx, overwritesElementType, overwritesState)
		resp.Diagnostics.Append(diags...)

		resp.PlanValue = newList
	}
}

func (m revertOverwritesIfExecutionModeUnset) Description(_ context.Context) string {
	return "Reverts to computed defaults if settings are unset"
}

func (m revertOverwritesIfExecutionModeUnset) MarkdownDescription(_ context.Context) string {
	return "Reverts to computed defaults if settings are unset"
}

func (m unknownIfExecutionModeUnset) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// Check if the resource is being created.
	if req.State.Raw.IsNull() {
		return
	}

	// Determine if configured execution_mode is being unset
	state := modelWorkspaceSettings{}
	configured := modelWorkspaceSettings{}

	resp.Diagnostics.Append(req.Config.Get(ctx, &configured)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if !state.Overwrites.IsNull() {
		// Normal operation
		overwritesState := make([]modelOverwrites, 1)
		state.Overwrites.ElementsAs(ctx, &overwritesState, true)

		if configured.ExecutionMode.IsNull() && overwritesState[0].ExecutionMode.ValueBool() {
			resp.PlanValue = types.StringUnknown()
		}
	} else if configured.ExecutionMode.IsNull() && req.Path.Equal(path.Root("execution_mode")) {
		// TFE does not support overwrites so default the execution mode to "remote"
		resp.PlanValue = types.StringValue("remote")
	} else if configured.AgentPoolID.IsNull() && req.Path.Equal(path.Root("agent_pool_id")) {
		resp.PlanValue = types.StringNull()
	}
}

func (m unknownIfExecutionModeUnset) Description(_ context.Context) string {
	return "Resets execution_mode to an unknown value if it is unset"
}

func (m unknownIfExecutionModeUnset) MarkdownDescription(_ context.Context) string {
	return "Resets execution_mode to an unknown value if it is unset"
}

func (r *workspaceSettings) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:        "Additional Workspace settings that override organization defaults",
		DeprecationMessage: "",
		Version:            1,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Service-generated identifier for the variable",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"workspace_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"execution_mode": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					unknownIfExecutionModeUnset{},
				},
				Validators: []validator.String{
					stringvalidator.OneOf("agent", "local", "remote"),
				},
			},

			"agent_pool_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					unknownIfExecutionModeUnset{},
					validateAgentExecutionMode{},
				},
			},

			// ListAttribute was required here because we are still using plugin protocol v5.
			// Once compatibility is broken for v1, and we convert all
			// providers to protocol v6, this can become a single nested object.
			"overwrites": schema.ListAttribute{
				Computed:    true,
				ElementType: overwritesElementType,
				PlanModifiers: []planmodifier.List{
					revertOverwritesIfExecutionModeUnset{},
				},
			},

			"global_remote_state": schema.BoolAttribute{
				Description: "Whether the workspace allows all workspaces in the organization to access its state data during runs. If false, then only workspaces defined in `remote_state_consumer_ids` can access its state.",
				Optional:    true,
				Computed:    true,
			},

			"remote_state_consumer_ids": schema.SetAttribute{
				Description: "The set of workspace IDs set as explicit remote state consumers for the given workspace.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.Set{
					validateRemoteStateConsumerIDs{},
					validateSelfReference{},
				},
			},

			"description": schema.StringAttribute{
				Optional:    true,
				Description: "A description of the workspace.",
			},

			"auto_apply": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "If set to false a human will have to manually confirm a plan in HCP Terraform's UI to start an apply. If set to true, this resource will be automatically applied.",
				Default:     booldefault.StaticBool(false),
			},

			"assessments_enabled": schema.BoolAttribute{
				Description: "If set to true, assessments will be enabled for the workspace. This includes drift and continuous validation checks.",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

// workspaceSettingsModelFromTFEWorkspace builds a resource model from the TFE model
func (r *workspaceSettings) workspaceSettingsModelFromTFEWorkspace(ws *tfe.Workspace) *modelWorkspaceSettings {
	result := modelWorkspaceSettings{
		ID:                 types.StringValue(ws.ID),
		WorkspaceID:        types.StringValue(ws.ID),
		ExecutionMode:      types.StringValue(ws.ExecutionMode),
		GlobalRemoteState:  types.BoolValue(ws.GlobalRemoteState),
		Description:        types.StringValue(ws.Description),
		AutoApply:          types.BoolValue(ws.AutoApply),
		AssessmentsEnabled: types.BoolValue(ws.AssessmentsEnabled),
	}

	if ws.AgentPool != nil && ws.ExecutionMode == "agent" {
		result.AgentPoolID = types.StringValue(ws.AgentPool.ID)
	}

	result.RemoteStateConsumerIDs = types.SetValueMust(types.StringType, []attr.Value{})

	if !ws.GlobalRemoteState {
		_, remoteStateConsumerIDs, err := readWorkspaceStateConsumers(ws.ID, r.config.Client)
		if err != nil {
			log.Printf("[ERROR] Error reading remote state consumers for workspace %s: %s", ws.ID, err)
			return nil
		}

		remoteStateConsumerIDValues, diags := types.SetValueFrom(ctx, types.StringType, remoteStateConsumerIDs)
		if diags.HasError() {
			log.Printf("[ERROR] Error reading remote state consumers for workspace %s: %v", ws.ID, diags)
			return nil
		}
		result.RemoteStateConsumerIDs = remoteStateConsumerIDValues
	}

	result.Overwrites = types.ListNull(overwritesElementType)
	if r.supportsOverwrites = ws.SettingOverwrites != nil; r.supportsOverwrites {
		settingsModel := modelOverwrites{
			ExecutionMode: types.BoolValue(*ws.SettingOverwrites.ExecutionMode),
			AgentPool:     types.BoolValue(*ws.SettingOverwrites.AgentPool),
		}

		listOverwrites, diags := types.ListValueFrom(ctx, overwritesElementType, []modelOverwrites{settingsModel})
		if diags.HasError() {
			panic("Could not build list value from slice of models. This should not be possible unless the model breaks reflection rules.")
		}

		result.Overwrites = listOverwrites
	}

	return &result
}

func (r *workspaceSettings) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data modelWorkspaceSettings
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	model, err := r.readSettings(ctx, data.WorkspaceID.ValueString())
	if errors.Is(err, errWorkspaceNoLongerExists) {
		resp.State.RemoveResource(ctx)
		return
	} else if err != nil {
		resp.Diagnostics.AddError("Error reading workspace", err.Error())
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

func (r *workspaceSettings) readSettings(ctx context.Context, workspaceID string) (*modelWorkspaceSettings, error) {
	ws, err := r.config.Client.Workspaces.ReadByID(ctx, workspaceID)
	if err != nil {
		// If it's gone: that's not an error, but we are done.
		if errors.Is(err, tfe.ErrResourceNotFound) {
			log.Printf("[DEBUG] Workspace %s no longer exists", workspaceID)
			return nil, errWorkspaceNoLongerExists
		}
		return nil, fmt.Errorf("couldn't read workspace %s: %s", workspaceID, err.Error())
	}

	return r.workspaceSettingsModelFromTFEWorkspace(ws), nil
}

func (r *workspaceSettings) updateSettings(ctx context.Context, data *modelWorkspaceSettings, state *tfsdk.State) error {
	workspaceID := data.WorkspaceID.ValueString()

	updateOptions := tfe.WorkspaceUpdateOptions{
		GlobalRemoteState: tfe.Bool(data.GlobalRemoteState.ValueBool()),
		SettingOverwrites: &tfe.WorkspaceSettingOverwritesOptions{
			ExecutionMode: tfe.Bool(false),
			AgentPool:     tfe.Bool(false),
		},
		Description:        tfe.String(data.Description.ValueString()),
		AutoApply:          tfe.Bool(data.AutoApply.ValueBool()),
		AssessmentsEnabled: tfe.Bool(data.AssessmentsEnabled.ValueBool()),
	}

	executionMode := data.ExecutionMode.ValueString()
	if executionMode != "" {
		updateOptions.ExecutionMode = tfe.String(executionMode)
		updateOptions.SettingOverwrites.ExecutionMode = tfe.Bool(true)
		updateOptions.SettingOverwrites.AgentPool = tfe.Bool(true)

		agentPoolID := data.AgentPoolID.ValueString() // may be empty
		updateOptions.AgentPoolID = tfe.String(agentPoolID)
	} else if executionMode == "" && data.Overwrites.IsNull() {
		// Not supported by TFE
		updateOptions.ExecutionMode = tfe.String("remote")
	}

	ws, err := r.config.Client.Workspaces.UpdateByID(ctx, workspaceID, updateOptions)
	if err != nil {
		return fmt.Errorf("couldn't update workspace %s: %w", workspaceID, err)
	}

	if !data.GlobalRemoteState.ValueBool() {
		r.addAndRemoveRemoteStateConsumers(workspaceID, data.RemoteStateConsumerIDs, state)
	}

	model, err := r.readSettings(ctx, ws.ID)
	if err != nil {
		return fmt.Errorf("couldn't read workspace %s after update: %w", workspaceID, err)
	}
	state.Set(ctx, model)
	return nil
}

func (r *workspaceSettings) addAndRemoveRemoteStateConsumers(workspaceID string, newWorkspaceIDsSet types.Set, state *tfsdk.State) error {
	var oldWorkspaceIDsSet types.Set
	diags := state.GetAttribute(ctx, path.Root("remote_state_consumer_ids"), &oldWorkspaceIDsSet)
	if diags.HasError() {
		return fmt.Errorf("error comparing remote state consumer IDs: %s", diags.Errors())
	}

	var oldWorkspaceIDs []string
	if !oldWorkspaceIDsSet.IsNull() {
		diags = oldWorkspaceIDsSet.ElementsAs(ctx, &oldWorkspaceIDs, true)
		if diags.HasError() {
			return fmt.Errorf("error comparing remote state consumer IDs: %s", diags.Errors())
		}
	}

	var newWorkspaceIDs []string
	if !newWorkspaceIDsSet.IsNull() {
		diags = newWorkspaceIDsSet.ElementsAs(ctx, &newWorkspaceIDs, true)
		if diags.HasError() {
			return fmt.Errorf("error comparing remote state consumer IDs: %s", diags.Errors())
		}
	}

	var workspaceIDsToRemove []string
	for _, id := range oldWorkspaceIDs {
		if !slices.Contains(newWorkspaceIDs, id) {
			workspaceIDsToRemove = append(workspaceIDsToRemove, id)
		}
	}

	var workspaceIDsToAdd []string
	for _, id := range newWorkspaceIDs {
		if !slices.Contains(oldWorkspaceIDs, id) {
			workspaceIDsToAdd = append(workspaceIDsToAdd, id)
		}
	}

	// First add the new consumerss
	if len(workspaceIDsToAdd) > 0 {
		options := tfe.WorkspaceAddRemoteStateConsumersOptions{}

		for _, wsID := range newWorkspaceIDs {
			options.Workspaces = append(options.Workspaces, &tfe.Workspace{ID: wsID})
		}

		log.Printf("[DEBUG] Adding remote state consumers %v to workspace: %s", workspaceIDsToAdd, workspaceID)
		err := r.config.Client.Workspaces.AddRemoteStateConsumers(ctx, workspaceID, options)
		if err != nil {
			return fmt.Errorf("Error adding remote state consumers to workspace %s: %w", workspaceID, err)
		}
	}

	// Then remove all the old consumers.
	if len(workspaceIDsToRemove) > 0 {
		options := tfe.WorkspaceRemoveRemoteStateConsumersOptions{}

		for _, wsID := range workspaceIDsToRemove {
			options.Workspaces = append(options.Workspaces, &tfe.Workspace{ID: wsID})
		}

		log.Printf("[DEBUG] Removing remote state consumers %v from workspace: %s", workspaceIDsToRemove, workspaceID)
		err := r.config.Client.Workspaces.RemoveRemoteStateConsumers(ctx, workspaceID, options)
		if err != nil {
			return fmt.Errorf("Error removing remote state consumers from workspace %s: %w", workspaceID, err)
		}
	}

	return nil
}

func (r *workspaceSettings) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data modelWorkspaceSettings
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if err := r.updateSettings(ctx, &data, &resp.State); err != nil {
		resp.Diagnostics.AddError("Error updating workspace", err.Error())
	}
}

func (r *workspaceSettings) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data modelWorkspaceSettings
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if err := r.updateSettings(ctx, &data, &resp.State); err != nil {
		resp.Diagnostics.AddError("Error updating workspace", err.Error())
	}
}

func (r *workspaceSettings) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data modelWorkspaceSettings
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	noneModel := modelWorkspaceSettings{
		ID:          data.ID,
		WorkspaceID: data.ID,
	}

	if err := r.updateSettings(ctx, &noneModel, &resp.State); err == nil {
		resp.State.RemoveResource(ctx)
	}
}

func (r *workspaceSettings) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "tfe_workspace_settings"
}

func (r *workspaceSettings) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Early exit if provider is unconfigured (i.e. we're only validating config or something)
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(ConfiguredClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected resource Configure type",
			fmt.Sprintf("Expected tfe.ConfiguredClient, got %T. This is a bug in the tfe provider, so please report it on GitHub.", req.ProviderData),
		)
	}
	r.config = client
}

func (r *workspaceSettings) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	s := strings.Split(req.ID, "/")
	if len(s) >= 3 {
		resp.Diagnostics.AddError("Error importing workspace settings", fmt.Sprintf(
			"invalid workspace input format: %s (expected <ORGANIZATION>/<WORKSPACE NAME> or <WORKSPACE ID>)",
			req.ID,
		))
	} else if len(s) == 2 {
		workspaceID, err := fetchWorkspaceExternalID(s[0]+"/"+s[1], r.config.Client)
		if err != nil {
			resp.Diagnostics.AddError("Error importing workspace settings", fmt.Sprintf(
				"error retrieving workspace with name %s from organization %s: %s", s[1], s[0], err.Error(),
			))
		}

		req.ID = workspaceID
	}

	resp.State.SetAttribute(ctx, path.Root("workspace_id"), req.ID)
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func NewResourceWorkspaceSettings() resource.Resource {
	return &workspaceSettings{}
}
