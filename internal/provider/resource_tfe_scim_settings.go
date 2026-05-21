// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/jsonapi"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type modelTFESCIMSettings struct {
	ID                        types.String `tfsdk:"id"`
	Enabled                   types.Bool   `tfsdk:"enabled"`
	Paused                    types.Bool   `tfsdk:"paused"`
	SiteAdminGroupSCIMID      types.String `tfsdk:"site_admin_group_scim_id"`
	SiteAdminGroupDisplayName types.String `tfsdk:"site_admin_group_display_name"`
}

// resourceTFESCIMSettings implements the tfe_scim_settings resource type
type resourceTFESCIMSettings struct {
	client *tfe.Client
}

// modelFromTFEAdminSCIMSettings builds a modelTFESCIMSettings struct from a tfe.AdminSCIMSetting value
func modelFromTFEAdminSCIMSettings(v tfe.AdminSCIMSetting) modelTFESCIMSettings {
	m := modelTFESCIMSettings{
		ID:                        types.StringValue("scim"),
		Enabled:                   types.BoolValue(v.Enabled),
		Paused:                    types.BoolValue(v.Paused),
		SiteAdminGroupSCIMID:      types.StringValue(v.SiteAdminGroupSCIMID),
		SiteAdminGroupDisplayName: types.StringValue(v.SiteAdminGroupDisplayName),
	}
	return m
}

var (
	_ resource.Resource                = &resourceTFESCIMSettings{}
	_ resource.ResourceWithConfigure   = &resourceTFESCIMSettings{}
	_ resource.ResourceWithImportState = &resourceTFESCIMSettings{}
)

// NewSCIMSettingsResource is a resource function for the framework provider.
func NewSCIMSettingsResource() resource.Resource {
	return &resourceTFESCIMSettings{}
}

// Metadata implements resource.Resource
func (r *resourceTFESCIMSettings) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_scim_settings"
}

// Configure implements resource.ResourceWithConfigure
func (r *resourceTFESCIMSettings) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Early exit if provider is not properly configured (i.e. we're only validating config or something)
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
	r.client = client.Client
}

// Schema implements resource.Resource
func (r *resourceTFESCIMSettings) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages SCIM provisioning settings for the Terraform Enterprise instance.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the SCIM settings. Always `scim`.",
				Computed:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether SCIM provisioning is enabled. Always `true` while this resource exists; use `terraform destroy` to disable. If SCIM is disabled outside of Terraform, the next `terraform plan` will propose re-creating this resource.",
				Computed:    true,
			},
			"paused": schema.BoolAttribute{
				Description: "Whether SCIM provisioning is paused. Defaults to `false`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"site_admin_group_scim_id": schema.StringAttribute{
				Description: "SCIM ID of the group whose members are granted site admin privileges. Defaults to `\"\"` (no group linked).",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"site_admin_group_display_name": schema.StringAttribute{
				Description: "Display name of the group whose members are granted site admin privileges.",
				Computed:    true,
			},
		},
	}
}

// Read implements resource.Resource
func (r *resourceTFESCIMSettings) Read(ctx context.Context, _ resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "Reading SCIM Settings")

	scimSettings, err := r.client.Admin.Settings.SCIM.Read(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading SCIM Settings", "Could not read SCIM Settings, unexpected error: "+err.Error())
		return
	}

	// If SCIM was disabled out-of-band, signal that the resource no longer exists
	// so Terraform will plan a Create to re-enable it on the next apply.
	// API contract: when Enabled=false the remaining fields (Paused,
	// SiteAdminGroupSCIMID, etc.) are always zero-valued, so no valid settings
	// are lost by removing state here.
	if !scimSettings.Enabled {
		resp.State.RemoveResource(ctx)
		return
	}

	// update state with refreshed data
	result := modelFromTFEAdminSCIMSettings(*scimSettings)
	diags := resp.State.Set(ctx, &result)
	resp.Diagnostics.Append(diags...)
}

// Create implements resource.Resource
func (r *resourceTFESCIMSettings) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var m modelTFESCIMSettings
	diags := req.Plan.Get(ctx, &m)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating SCIM Settings")
	scimSettings, err := r.updateSCIMSettings(ctx, m)
	if err != nil {
		resp.Diagnostics.AddError("Error creating SCIM Settings", "Could not set SCIM Settings, unexpected error: "+err.Error())
		return
	}

	result := modelFromTFEAdminSCIMSettings(*scimSettings)
	diags = resp.State.Set(ctx, &result)
	resp.Diagnostics.Append(diags...)
}

func (r *resourceTFESCIMSettings) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var m modelTFESCIMSettings
	diags := req.Plan.Get(ctx, &m)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Update SCIM Settings")
	scimSettings, err := r.updateSCIMSettings(ctx, m)
	if err != nil {
		resp.Diagnostics.AddError("Error updating SCIM Settings", "Could not set SCIM Settings, unexpected error: "+err.Error())
		return
	}
	result := modelFromTFEAdminSCIMSettings(*scimSettings)
	diags = resp.State.Set(ctx, &result)
	resp.Diagnostics.Append(diags...)
}

func (r *resourceTFESCIMSettings) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Debug(ctx, "Delete SCIM Settings")
	err := r.client.Admin.Settings.SCIM.Delete(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting SCIM Settings", "Could not disable SCIM Settings, unexpected error: "+err.Error())
		return
	}
}

func (r *resourceTFESCIMSettings) ImportState(ctx context.Context, _ resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	scimSettings, err := r.client.Admin.Settings.SCIM.Read(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error importing SCIM Settings", "Could not retrieve SCIM Settings, unexpected error: "+err.Error())
		return
	}

	if !scimSettings.Enabled {
		resp.Diagnostics.AddError(
			"Cannot import disabled SCIM Settings",
			"SCIM provisioning is currently disabled. Enable SCIM before importing, or use 'terraform apply' to enable it via this resource.",
		)
		return
	}

	result := modelFromTFEAdminSCIMSettings(*scimSettings)
	diags := resp.State.Set(ctx, &result)
	resp.Diagnostics.Append(diags...)
}

// updateSCIMSettings applies the SCIM settings for Create and Update. The plan
// is the source of truth: every field is sent on every call (schema defaults
// populate fields the user omits). site_admin_group_scim_id sends JSON null
// when empty (unlinks the group) and the raw value otherwise (links it).
func (r *resourceTFESCIMSettings) updateSCIMSettings(ctx context.Context, m modelTFESCIMSettings) (*tfe.AdminSCIMSetting, error) {
	var siteAdminGroupSCIMID jsonapi.NullableAttr[string]
	switch {
	case m.SiteAdminGroupSCIMID.IsUnknown():
		// Can't distinguish "not yet resolved" from an intentional unlink; fail loudly.
		return nil, fmt.Errorf("site_admin_group_scim_id is not yet known; ensure the value is resolved before applying")
	case m.SiteAdminGroupSCIMID.IsNull() || m.SiteAdminGroupSCIMID.ValueString() == "":
		// Empty/null → unlink the site admin group.
		siteAdminGroupSCIMID = tfe.NullString()
	default:
		siteAdminGroupSCIMID = tfe.NullableString(m.SiteAdminGroupSCIMID.ValueString())
	}

	s, err := r.client.Admin.Settings.SCIM.Update(ctx, tfe.AdminSCIMSettingUpdateOptions{
		Enabled:              tfe.Bool(true),
		Paused:               m.Paused.ValueBoolPointer(),
		SiteAdminGroupSCIMID: siteAdminGroupSCIMID,
	})

	if err != nil {
		return s, fmt.Errorf("failed to set SCIM Settings: %w", err)
	}
	return s, nil
}
