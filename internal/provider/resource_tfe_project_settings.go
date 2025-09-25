// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"

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
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// tfe_project_settings resource
var _ resource.Resource = &projectSettings{}

// projectOverwritesElementType is the object type definition for the
// overwrites field schema.
var projectOverwritesElementType = map[string]attr.Type{
	"default_execution_mode": types.BoolType,
	"default_agent_pool_id":  types.BoolType,
}

type projectSettings struct {
	config ConfiguredClient
}

type modelProjectSettings struct {
	ID                   types.String `tfsdk:"id"`
	ProjectID            types.String `tfsdk:"project_id"`
	DefaultExecutionMode types.String `tfsdk:"default_execution_mode"`
	DefaultAgentPoolID   types.String `tfsdk:"default_agent_pool_id"`
	Overwrites           types.Object `tfsdk:"overwrites"`
}

type projectOverwrites struct {
	DefaultExecutionMode types.Bool `tfsdk:"default_execution_mode"`
	DefaultAgentPoolID   types.Bool `tfsdk:"default_agent_pool_id"`
}

// errProjectNoLongerExists is returned when reading the project settings but
// the project no longer exists.
var errProjectNoLongerExists = errors.New("project no longer exists")

// validateProjectDefaultAgentExecutionMode is a PlanModifier that validates that the combination
// of "default_execution_mode" and "default_agent_pool_id" is compatible.
type validateProjectDefaultAgentExecutionMode struct{}

// revertOverwritesIfDefaultExecutionModeUnset is a PlanModifier for "overwrites" that
// sets the values to false if default_execution_mode is unset. This tells the server to
// compute default_execution_mode and default_agent_pool_id if defaults are set. This
// modifier must be used in conjunction with unknownIfDefaultExecutionModeUnset plan
// modifier on the default_execution_mode and default_agent_pool_id fields.
type revertOverwritesIfDefaultExecutionModeUnset struct{}

// unknownIfDefaultExecutionModeUnset sets the planned value to (known after apply) if
// default_execution_mode is unset, avoiding an inconsistent state after the apply. This
// allows the server to compute the new value based on the default. It should be
// applied to both default_execution_mode and default_agent_pool_id in conjunction with
// revertOverwritesIfDefaultExecutionModeUnset.
type unknownIfDefaultExecutionModeUnset struct{}

// overwriteExecutionModeIfSpecified is a PlanModifier that forces the value of
// default_execution_mode to be set if it is explicitly specified in the configuration,
// even if it matches the organization default. This ensures that a value is always
// sent to the API when the user has specified one.
type overwriteExecutionModeIfSpecified struct{}

var _ planmodifier.String = (*validateProjectDefaultAgentExecutionMode)(nil)
var _ planmodifier.Object = (*revertOverwritesIfDefaultExecutionModeUnset)(nil)
var _ planmodifier.String = (*unknownIfDefaultExecutionModeUnset)(nil)
var _ planmodifier.Object = (*overwriteExecutionModeIfSpecified)(nil)

func (m validateProjectDefaultAgentExecutionMode) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	configured := modelProjectSettings{}
	resp.Diagnostics.Append(req.Config.Get(ctx, &configured)...)

	if configured.DefaultExecutionMode.ValueString() == "agent" && configured.DefaultAgentPoolID.IsNull() {
		resp.Diagnostics.AddError("Invalid default_agent_pool_id", "If default execution mode is \"agent\", \"default_agent_pool_id\" is required")
	}

	if configured.DefaultExecutionMode.ValueString() != "agent" && !configured.DefaultAgentPoolID.IsNull() {
		resp.Diagnostics.AddError("Invalid default_agent_pool_id", "If default execution mode is not \"agent\", \"default_agent_pool_id\" must not be set")
	}
}

func (m validateProjectDefaultAgentExecutionMode) Description(_ context.Context) string {
	return "Validates that configuration values for \"default_agent_pool_id\" and \"default_execution_mode\" are compatible"
}

func (m validateProjectDefaultAgentExecutionMode) MarkdownDescription(_ context.Context) string {
	return "Validates that configuration values for \"default_agent_pool_id\" and \"default_execution_mode\" are compatible"
}

func (m revertOverwritesIfDefaultExecutionModeUnset) PlanModifyObject(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
	// Check if the resource is being created.
	if req.State.Raw.IsNull() {
		return
	}

	// Determine if configured default_execution_mode is being unset
	state := modelProjectSettings{}
	configured := modelProjectSettings{}

	resp.Diagnostics.Append(req.Config.Get(ctx, &configured)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	// Check if overwrites are supported by the platform
	if state.Overwrites.IsNull() {
		return
	}

	overwritesState := projectOverwrites{}

	state.Overwrites.As(ctx, &overwritesState, basetypes.ObjectAsOptions{})

	// if there is a default execution mode set in state, but not one configured, then set the overwrites to false
	if configured.DefaultExecutionMode.IsNull() && overwritesState.DefaultExecutionMode.ValueBool() {
		overwritesState.DefaultAgentPoolID = types.BoolValue(false)
		overwritesState.DefaultExecutionMode = types.BoolValue(false)

		newProjOverwrites, diags := types.ObjectValueFrom(ctx, projectOverwritesElementType, overwritesState)
		resp.Diagnostics.Append(diags...)

		resp.PlanValue = newProjOverwrites
	}
}

func (m revertOverwritesIfDefaultExecutionModeUnset) Description(_ context.Context) string {
	return "Reverts to computed defaults if settings are unset"
}

func (m revertOverwritesIfDefaultExecutionModeUnset) MarkdownDescription(_ context.Context) string {
	return "Reverts to computed defaults if settings are unset"
}

func (m unknownIfDefaultExecutionModeUnset) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// Check if the resource is being created.
	if req.State.Raw.IsNull() {
		return
	}

	// Determine if configured default_execution_mode is being unset
	state := modelProjectSettings{}
	configured := modelProjectSettings{}

	resp.Diagnostics.Append(req.Config.Get(ctx, &configured)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if !state.Overwrites.IsNull() {
		overwritesState := projectOverwrites{}
		state.Overwrites.As(ctx, &overwritesState, basetypes.ObjectAsOptions{})

		// if there is a default execution mode set in state, but not one configured, then set the planned value for the default execution mode and agent pool to unknown
		if configured.DefaultExecutionMode.IsNull() && overwritesState.DefaultExecutionMode.ValueBool() {
			resp.PlanValue = types.StringUnknown()
		}
	}
}

func (m unknownIfDefaultExecutionModeUnset) Description(_ context.Context) string {
	return "Resets default_execution_mode to an unknown value if it is unset"
}

func (m unknownIfDefaultExecutionModeUnset) MarkdownDescription(_ context.Context) string {
	return "Resets default_execution_mode to an unknown value if it is unset"
}

func (m overwriteExecutionModeIfSpecified) PlanModifyObject(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
	// Check if the resource is being created.
	if req.State.Raw.IsNull() {
		return
	}

	state := modelProjectSettings{}
	configured := modelProjectSettings{}

	resp.Diagnostics.Append(req.Config.Get(ctx, &configured)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	overwritesState := projectOverwrites{}
	state.Overwrites.As(ctx, &overwritesState, basetypes.ObjectAsOptions{})

	if !state.Overwrites.IsNull() {
		// if an execution mode is configured, ensure that the overwrites are set to true
		if !configured.DefaultExecutionMode.IsNull() {
			overwritesState.DefaultAgentPoolID = types.BoolValue(true)
			overwritesState.DefaultExecutionMode = types.BoolValue(true)

			newList, diags := types.ObjectValueFrom(ctx, projectOverwritesElementType, overwritesState)
			resp.Diagnostics.Append(diags...)

			resp.PlanValue = newList
		}
	}
}

func (m overwriteExecutionModeIfSpecified) Description(_ context.Context) string {
	return "Overwrites default_execution_mode if it is explicitly specified in the configuration, even if it matches the organization default"
}

func (m overwriteExecutionModeIfSpecified) MarkdownDescription(_ context.Context) string {
	return "Overwrites default_execution_mode if it is explicitly specified in the configuration, even if it matches the organization default"
}

func (r *projectSettings) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:        "Additional Project settings, which may override organization defaults",
		DeprecationMessage: "",
		Version:            1,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Service-generated identifier for the resource",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"project_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"default_execution_mode": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					unknownIfDefaultExecutionModeUnset{},
				},
				Validators: []validator.String{
					stringvalidator.OneOf("agent", "local", "remote"),
				},
			},

			"default_agent_pool_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					unknownIfDefaultExecutionModeUnset{},
					validateProjectDefaultAgentExecutionMode{},
				},
			},
			"overwrites": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Describes which settings are being overwritten from the organization defaults",
				Attributes: map[string]schema.Attribute{
					"default_execution_mode": schema.BoolAttribute{
						Computed:    true,
						Description: "Whether the default_execution_mode is being overwritten from the organization default",
					},
					"default_agent_pool_id": schema.BoolAttribute{
						Computed:    true,
						Description: "Whether the default_agent_pool_id is being overwritten from the organization default",
					},
				},
				PlanModifiers: []planmodifier.Object{
					revertOverwritesIfDefaultExecutionModeUnset{},
					overwriteExecutionModeIfSpecified{},
				},
			},
		},
	}
}

// projectSettingsModelFromTFEProject builds a resource model from the TFE model
func (r *projectSettings) projectSettingsModelFromTFEProject(proj *tfe.Project) *modelProjectSettings {
	result := modelProjectSettings{
		ID:                   types.StringValue(proj.ID),
		ProjectID:            types.StringValue(proj.ID),
		DefaultExecutionMode: types.StringValue(proj.DefaultExecutionMode),
	}

	if proj.DefaultAgentPool != nil && proj.DefaultExecutionMode == "agent" {
		result.DefaultAgentPoolID = types.StringValue(proj.DefaultAgentPool.ID)
	}

	result.Overwrites = types.ObjectNull(projectOverwritesElementType)
	if proj.SettingOverwrites != nil {
		settingsModel := projectOverwrites{
			DefaultExecutionMode: types.BoolValue(*proj.SettingOverwrites.ExecutionMode),
			DefaultAgentPoolID:   types.BoolValue(*proj.SettingOverwrites.AgentPool),
		}

		objectOverwrites, diags := types.ObjectValueFrom(ctx, projectOverwritesElementType, settingsModel)
		if diags.HasError() {
			panic("Could not build object value from model. This should not be possible unless the model breaks reflection rules.")
		}

		result.Overwrites = objectOverwrites
	}

	return &result
}

func (r *projectSettings) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data modelProjectSettings
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	model, err := r.readSettings(ctx, data.ProjectID.ValueString())
	if errors.Is(err, errProjectNoLongerExists) {
		resp.State.RemoveResource(ctx)
		return
	} else if err != nil {
		resp.Diagnostics.AddError("Error reading project", err.Error())
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

func (r *projectSettings) readSettings(ctx context.Context, projectID string) (*modelProjectSettings, error) {
	proj, err := r.config.Client.Projects.Read(ctx, projectID)
	if errors.Is(err, tfe.ErrResourceNotFound) {
		return nil, errProjectNoLongerExists
	}

	if err != nil {
		return nil, fmt.Errorf("error reading configuration of project %s: %w", projectID, err)
	}

	return r.projectSettingsModelFromTFEProject(proj), nil
}

func (r *projectSettings) updateSettings(ctx context.Context, data *modelProjectSettings, state *tfsdk.State) error {
	projectID := data.ProjectID.ValueString()

	updateOptions := tfe.ProjectUpdateOptions{
		SettingOverwrites: &tfe.ProjectSettingOverwrites{
			ExecutionMode: tfe.Bool(false),
			AgentPool:     tfe.Bool(false),
		},
	}

	defaultExecutionMode := data.DefaultExecutionMode.ValueString()
	if defaultExecutionMode != "" {
		updateOptions.DefaultExecutionMode = tfe.String(defaultExecutionMode)
		updateOptions.SettingOverwrites.ExecutionMode = tfe.Bool(true)
		updateOptions.SettingOverwrites.AgentPool = tfe.Bool(true)

		defaultAgentPoolID := data.DefaultAgentPoolID.ValueString() // may be empty
		updateOptions.DefaultAgentPoolID = tfe.String(defaultAgentPoolID)
	}

	proj, err := r.config.Client.Projects.Update(ctx, projectID, updateOptions)
	if err != nil {
		return fmt.Errorf("couldn't update project %s: %w", projectID, err)
	}

	model, err := r.readSettings(ctx, proj.ID)
	if errors.Is(err, errProjectNoLongerExists) {
		state.RemoveResource(ctx)
	}

	if err == nil {
		state.Set(ctx, model)
	}

	return err
}

func (r *projectSettings) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var projectID string
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("project_id"), &projectID)...)
	if resp.Diagnostics.HasError() {
		return
	}

	planned := modelProjectSettings{}
	resp.Diagnostics.Append(req.Config.Get(ctx, &planned)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.updateSettings(ctx, &planned, &resp.State); err != nil {
		resp.Diagnostics.AddError("Error updating project", err.Error())
	}
}

func (r *projectSettings) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data modelProjectSettings
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if err := r.updateSettings(ctx, &data, &resp.State); err != nil {
		resp.Diagnostics.AddError("Error updating project", err.Error())
	}
}

func (r *projectSettings) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data modelProjectSettings
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	noneModel := modelProjectSettings{
		ID:        data.ID,
		ProjectID: data.ID,
	}

	if err := r.updateSettings(ctx, &noneModel, &resp.State); err == nil {
		resp.State.RemoveResource(ctx)
	}
}

func (r *projectSettings) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "tfe_project_settings"
}

func (r *projectSettings) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Early exit if provider is un-configured (i.e. we're only validating config or something)
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

func (r *projectSettings) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.State.SetAttribute(ctx, path.Root("project_id"), req.ID)
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func NewProjectSettingsResource() resource.Resource {
	return &projectSettings{}
}
