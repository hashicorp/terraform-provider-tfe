// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// minTFEVersionAdminGeneralSettings is the minimum Terraform Enterprise version
// that supports the admin general settings managed by this resource
// (specifically the enable_public_registration attribute).
const minTFEVersionAdminGeneralSettings = "2.1.0"

// privateKeyOriginalPublicRegistration is the private state key used to record
// the server-side enable_public_registration value observed before this
// resource first managed it. It is restored on delete so that removing this
// resource leaves the instance as it was found rather than forcing a default.
const privateKeyOriginalPublicRegistration = "original_enable_public_registration"

// modelTFEAdminGeneralSettings maps the tfe_admin_general_settings resource
// schema to a Go struct.
type modelTFEAdminGeneralSettings struct {
	ID                       types.String `tfsdk:"id"`
	EnablePublicRegistration types.Bool   `tfsdk:"enable_public_registration"`
}

// resourceTFEAdminGeneralSettings implements the tfe_admin_general_settings
// resource type.
type resourceTFEAdminGeneralSettings struct {
	config ConfiguredClient
}

var (
	_ resource.Resource                = &resourceTFEAdminGeneralSettings{}
	_ resource.ResourceWithConfigure   = &resourceTFEAdminGeneralSettings{}
	_ resource.ResourceWithImportState = &resourceTFEAdminGeneralSettings{}
)

// NewAdminGeneralSettingsResource is a resource function for the framework provider.
func NewAdminGeneralSettingsResource() resource.Resource {
	return &resourceTFEAdminGeneralSettings{}
}

// Configure implements resource.ResourceWithConfigure
func (r *resourceTFEAdminGeneralSettings) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(ConfiguredClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected resource Configure type",
			fmt.Sprintf("Expected tfe.ConfiguredClient, got %T. This is a bug in the tfe provider, so please report it on GitHub.", req.ProviderData),
		)
		return
	}
	r.config = client
}

// Metadata implements resource.Resource
func (r *resourceTFEAdminGeneralSettings) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_admin_general_settings"
}

// Schema implements resource.Resource
func (r *resourceTFEAdminGeneralSettings) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version: 0,
		Description: "(Only for Terraform Enterprise) Manages the admin general settings." +
			"\n\nRequires admin token configuration. See example usage for incorporating an admin token in your provider config." +
			"\n\nRequires Terraform Enterprise v" + minTFEVersionAdminGeneralSettings + " or later." +
			"\n\nThis is a singleton resource: there is only ever one set of general settings for a TFE instance. " +
			"Destroying this resource restores managed attributes to the values that were in effect before Terraform managed them. " +
			"If an imported resource is destroyed without having modified any settings, Terraform Enterprise settings remain unchanged.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the general settings.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enable_public_registration": schema.BoolAttribute{
				Description: "Whether public registration (self sign-up) is enabled. When omitted, the current " +
					"server-side value is preserved. Note that the underlying setting defaults to `false` on " +
					"Terraform Enterprise, so public registration is disabled unless an administrator has explicitly " +
					"enabled it. Removing this attribute from configuration does not reset it; set it explicitly instead.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Read implements resource.Resource
func (r *resourceTFEAdminGeneralSettings) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state modelTFEAdminGeneralSettings
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading Admin General Settings")
	settings, err := readAdminGeneralSettings(ctx, r.config.ClientV2)
	if err != nil {
		resp.Diagnostics.AddError("Error reading Admin General Settings", "Could not read Admin General Settings, unexpected error: "+apiErrorDetail(err))
		return
	}

	result := modelFromTFEAdminGeneralSettings(settings)
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

// Create implements resource.Resource
func (r *resourceTFEAdminGeneralSettings) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan modelTFEAdminGeneralSettings
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var config modelTFEAdminGeneralSettings
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !r.checkMinTFEVersion(&resp.Diagnostics) {
		return
	}

	if !config.EnablePublicRegistration.IsNull() && !config.EnablePublicRegistration.IsUnknown() {
		// Capture the value that existed before this resource took ownership so it
		// can be restored on delete, mirroring how a boolean flag would be left
		// untouched if it had never been managed.
		tflog.Debug(ctx, "Reading Admin General Settings before create")
		current, err := readAdminGeneralSettings(ctx, r.config.ClientV2)
		if err != nil {
			resp.Diagnostics.AddError("Error creating Admin General Settings", "Could not read existing Admin General Settings, unexpected error: "+apiErrorDetail(err))
			return
		}
		resp.Diagnostics.Append(setOriginalPublicRegistration(ctx, resp.Private, current.EnablePublicRegistration)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	tflog.Debug(ctx, "Creating Admin General Settings")
	result, err := r.updateAdminGeneralSettings(ctx, plan, config)
	if err != nil {
		resp.Diagnostics.AddError("Error creating Admin General Settings", "Could not set Admin General Settings, unexpected error: "+apiErrorDetail(err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

// Update implements resource.Resource
func (r *resourceTFEAdminGeneralSettings) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan modelTFEAdminGeneralSettings
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var config modelTFEAdminGeneralSettings
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !r.checkMinTFEVersion(&resp.Diagnostics) {
		return
	}

	if !config.EnablePublicRegistration.IsNull() && !config.EnablePublicRegistration.IsUnknown() {
		resp.Diagnostics.Append(r.recordOriginalPublicRegistrationIfNeeded(ctx, req.Private, resp.Private)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	tflog.Debug(ctx, "Updating Admin General Settings")
	result, err := r.updateAdminGeneralSettings(ctx, plan, config)
	if err != nil {
		resp.Diagnostics.AddError("Error updating Admin General Settings", "Could not set Admin General Settings, unexpected error: "+apiErrorDetail(err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

// Delete restores the enable_public_registration value that existed before this
// resource managed it, then removes the resource from state. There is no API
// delete endpoint; the value is restored via a PATCH.
func (r *resourceTFEAdminGeneralSettings) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Debug(ctx, "Deleting Admin General Settings")

	original, ok, diags := getOriginalPublicRegistration(ctx, req.Private)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Without a recorded original value (for example, an imported resource that
	// never ran through Create), leave the server-side setting untouched.
	if !ok {
		resp.Diagnostics.AddWarning(
			"Admin General Settings left unchanged",
			"No pre-management value was recorded for this resource (for example, because it was imported), "+
				"so enable_public_registration was removed from state without being modified in Terraform Enterprise.",
		)
		return
	}

	_, err := patchAdminGeneralSettings(ctx, r.config.ClientV2, &original)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting Admin General Settings", "Could not restore Admin General Settings, unexpected error: "+apiErrorDetail(err))
		return
	}
}

// ImportState implements resource.ResourceWithImportState
func (r *resourceTFEAdminGeneralSettings) ImportState(ctx context.Context, _ resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	settings, err := readAdminGeneralSettings(ctx, r.config.ClientV2)
	if err != nil {
		resp.Diagnostics.AddError("Error importing Admin General Settings", "Could not retrieve Admin General Settings, unexpected error: "+apiErrorDetail(err))
		return
	}

	result := modelFromTFEAdminGeneralSettings(settings)
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

// updateAdminGeneralSettings builds the update options from the plan and config,
// issues the PATCH (or reads if unconfigured), and maps the result back into
// the resource model.
func (r *resourceTFEAdminGeneralSettings) updateAdminGeneralSettings(ctx context.Context, plan modelTFEAdminGeneralSettings, config modelTFEAdminGeneralSettings) (modelTFEAdminGeneralSettings, error) {
	var enablePublicRegistration *bool
	if !config.EnablePublicRegistration.IsNull() && !config.EnablePublicRegistration.IsUnknown() {
		enablePublicRegistration = plan.EnablePublicRegistration.ValueBoolPointer()
	}
	settings, err := patchAdminGeneralSettings(ctx, r.config.ClientV2, enablePublicRegistration)
	if err != nil {
		return modelTFEAdminGeneralSettings{}, err
	}
	return modelFromTFEAdminGeneralSettings(settings), nil
}

// modelFromTFEAdminGeneralSettings maps a go-tfe general setting to the resource model.
func modelFromTFEAdminGeneralSettings(v *adminGeneralSettings) modelTFEAdminGeneralSettings {
	return modelTFEAdminGeneralSettings{
		ID:                       types.StringValue(v.ID),
		EnablePublicRegistration: types.BoolValue(v.EnablePublicRegistration),
	}
}

// checkMinTFEVersion verifies the connected Terraform Enterprise meets the
// minimum version that supports admin general settings. It records an error
// diagnostic and returns false when the requirement is not met, so that a write
// against HCP Terraform or an older release fails loudly instead of silently doing nothing.
func (r *resourceTFEAdminGeneralSettings) checkMinTFEVersion(diags *diag.Diagnostics) bool {
	meets, err := r.config.MeetsMinRemoteTFEOnlyVersion(minTFEVersionAdminGeneralSettings)
	if err != nil {
		diags.AddError(
			"Error checking minimum Terraform Enterprise version",
			fmt.Sprintf("Could not determine whether Terraform Enterprise version %s meets the minimum required version %s: %v",
				r.config.RemoteTFEVersion(), minTFEVersionAdminGeneralSettings, err),
		)
		return false
	}
	if !meets {
		if r.config.Client.IsCloud() {
			diags.AddError(
				"Unsupported platform",
				"tfe_admin_general_settings is only supported on Terraform Enterprise and cannot be used with HCP Terraform.",
			)
		} else {
			diags.AddError(
				"Unsupported Terraform Enterprise version",
				fmt.Sprintf("tfe_admin_general_settings requires Terraform Enterprise version %s or later, but the connected instance is running version %s.",
					minTFEVersionAdminGeneralSettings, r.config.RemoteTFEVersion()),
			)
		}
		return false
	}
	return true
}

// privateStateWriter is the subset of the framework's private state that this
// resource uses to record the pre-management value.
type privateStateWriter interface {
	SetKey(ctx context.Context, key string, value []byte) diag.Diagnostics
}

// privateStateReader is the subset of the framework's private state that this
// resource uses to recover the pre-management value.
type privateStateReader interface {
	GetKey(ctx context.Context, key string) ([]byte, diag.Diagnostics)
}

// setOriginalPublicRegistration records the enable_public_registration value
// observed before this resource managed it, so delete can restore it.
func setOriginalPublicRegistration(ctx context.Context, private privateStateWriter, value bool) diag.Diagnostics {
	encoded, err := json.Marshal(value)
	if err != nil {
		var diags diag.Diagnostics
		diags.AddError("Error storing Admin General Settings state", "Could not encode the original enable_public_registration value: "+err.Error())
		return diags
	}
	return private.SetKey(ctx, privateKeyOriginalPublicRegistration, encoded)
}

// getOriginalPublicRegistration returns the recorded pre-management value. The
// second return value is false when no value was ever recorded (for example,
// after an import that never ran through Create).
func getOriginalPublicRegistration(ctx context.Context, private privateStateReader) (bool, bool, diag.Diagnostics) {
	encoded, diags := private.GetKey(ctx, privateKeyOriginalPublicRegistration)
	if diags.HasError() || len(encoded) == 0 {
		return false, false, diags
	}

	var value bool
	if err := json.Unmarshal(encoded, &value); err != nil {
		diags.AddError("Error reading Admin General Settings state", "Could not decode the original enable_public_registration value: "+err.Error())
		return false, false, diags
	}
	return value, true, diags
}

// recordOriginalPublicRegistrationIfNeeded ensures a pre-management value is captured
// when an attribute is first configured during an update.
func (r *resourceTFEAdminGeneralSettings) recordOriginalPublicRegistrationIfNeeded(ctx context.Context, reader privateStateReader, writer privateStateWriter) diag.Diagnostics {
	var diags diag.Diagnostics
	_, ok, d := getOriginalPublicRegistration(ctx, reader)
	diags.Append(d...)
	if diags.HasError() || ok {
		return diags
	}

	current, err := readAdminGeneralSettings(ctx, r.config.ClientV2)
	if err != nil {
		diags.AddError("Error updating Admin General Settings", "Could not read existing Admin General Settings, unexpected error: "+apiErrorDetail(err))
		return diags
	}
	diags.Append(setOriginalPublicRegistration(ctx, writer, current.EnablePublicRegistration)...)
	return diags
}
