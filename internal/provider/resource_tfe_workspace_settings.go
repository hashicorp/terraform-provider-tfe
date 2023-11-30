// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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
	config ConfiguredClient
}

type modelWorkspaceSettings struct {
	ID            types.String `tfsdk:"id"`
	WorkspaceID   types.String `tfsdk:"workspace_id"`
	ExecutionMode types.String `tfsdk:"execution_mode"`
	AgentPoolID   types.String `tfsdk:"agent_pool_id"`
	Overwrites    types.List   `tfsdk:"overwrites"`
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

var _ planmodifier.String = (*validateAgentExecutionMode)(nil)
var _ planmodifier.List = (*revertOverwritesIfExecutionModeUnset)(nil)
var _ planmodifier.String = (*unknownIfExecutionModeUnset)(nil)

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

	overwritesState := make([]modelOverwrites, 1)
	state.Overwrites.ElementsAs(ctx, &overwritesState, true)

	if configured.ExecutionMode.IsNull() && overwritesState[0].ExecutionMode.ValueBool() {
		resp.PlanValue = types.StringUnknown()
	}
}

func (m unknownIfExecutionModeUnset) Description(_ context.Context) string {
	return "Resets execution_mode to \"remote\" if it is unset"
}

func (m unknownIfExecutionModeUnset) MarkdownDescription(_ context.Context) string {
	return "Resets execution_mode to \"remote\" if it is unset"
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
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						IDPattern("ws"),
						"must be a valid workspace ID (ws-<RANDOM STRING>)",
					),
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
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						IDPattern("apool"),
						"must be a valid workspace ID (apool-<RANDOM STRING>)",
					),
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
		},
	}
}

// workspaceSettingsModelFromTFEWorkspace builds a resource model from the TFE model
func workspaceSettingsModelFromTFEWorkspace(ws *tfe.Workspace) *modelWorkspaceSettings {
	result := modelWorkspaceSettings{
		ID:            types.StringValue(ws.ID),
		WorkspaceID:   types.StringValue(ws.ID),
		ExecutionMode: types.StringValue(ws.ExecutionMode),
	}

	if ws.AgentPool != nil && ws.ExecutionMode == "agent" {
		result.AgentPoolID = types.StringValue(ws.AgentPool.ID)
	}

	settingsModel := modelOverwrites{
		ExecutionMode: types.BoolValue(false),
		AgentPool:     types.BoolValue(false),
	}

	if ws.SettingOverwrites != nil {
		settingsModel = modelOverwrites{
			ExecutionMode: types.BoolValue(*ws.SettingOverwrites.ExecutionMode),
			AgentPool:     types.BoolValue(*ws.SettingOverwrites.AgentPool),
		}
	}

	listOverwrites, diags := types.ListValueFrom(ctx, overwritesElementType, []modelOverwrites{settingsModel})
	if diags.HasError() {
		panic("Could not build list value from slice of models. This should not be possible unless the model breaks reflection rules.")
	}

	result.Overwrites = listOverwrites

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

	return workspaceSettingsModelFromTFEWorkspace(ws), nil
}

func (r *workspaceSettings) updateSettings(ctx context.Context, data *modelWorkspaceSettings, state *tfsdk.State) error {
	workspaceID := data.WorkspaceID.ValueString()

	updateOptions := tfe.WorkspaceUpdateOptions{
		SettingOverwrites: &tfe.WorkspaceSettingOverwritesOptions{
			ExecutionMode: tfe.Bool(false),
			AgentPool:     tfe.Bool(false),
		},
	}

	if executionMode := data.ExecutionMode.ValueString(); executionMode != "" {
		updateOptions.ExecutionMode = tfe.String(executionMode)
		updateOptions.SettingOverwrites.ExecutionMode = tfe.Bool(true)
		updateOptions.SettingOverwrites.AgentPool = tfe.Bool(true)

		agentPoolID := data.AgentPoolID.ValueString() // may be empty
		updateOptions.AgentPoolID = tfe.String(agentPoolID)
	}

	ws, err := r.config.Client.Workspaces.UpdateByID(ctx, workspaceID, updateOptions)
	if err != nil {
		return fmt.Errorf("couldn't update workspace %s: %w", workspaceID, err)
	}

	model, err := r.readSettings(ctx, ws.ID)
	if err != nil {
		return fmt.Errorf("couldn't read workspace %s after update: %w", workspaceID, err)
	}
	state.Set(ctx, model)
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
