// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-tfe/v2/api/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// scimSettingsID is the ID of the SCIM settings singleton resource.
const scimSettingsID string = "scim"

type modelTFESCIMSettings struct {
	ID                          types.String `tfsdk:"id"`
	Enabled                     types.Bool   `tfsdk:"enabled"`
	Paused                      types.Bool   `tfsdk:"paused"`
	SiteAdminGroupSCIMID        types.String `tfsdk:"site_admin_group_scim_id"`
	SiteAdminGroupDisplayName   types.String `tfsdk:"site_admin_group_display_name"`
	SiteAuditorGroupSCIMID      types.String `tfsdk:"site_auditor_group_scim_id"`
	SiteAuditorGroupDisplayName types.String `tfsdk:"site_auditor_group_display_name"`
}

// resourceTFESCIMSettings implements the tfe_scim_settings resource type
type resourceTFESCIMSettings struct {
	config ConfiguredClient
}

// scimSettingsEnvelope wraps a set of SCIM settings attributes in the JSON:API
// document shape the admin SCIM settings endpoint expects.
func scimSettingsEnvelope(attrs models.AdminScimSettings_attributesable) models.AdminScimSettingsEnvelopeable {
	data := models.NewAdminScimSettings()
	data.SetId(ptr(scimSettingsID))
	data.SetTypeEscaped(ptr(models.SCIMSETTINGS_ADMINSCIMSETTINGS_TYPE))
	data.SetAttributes(attrs)

	envelope := models.NewAdminScimSettingsEnvelope()
	envelope.SetData(data)
	return envelope
}

// modelFromV2SCIMSettings builds a modelTFESCIMSettings from an admin SCIM
// settings response.
//
// Every group attribute is read with valueOrZero, so a missing pointer becomes
// "". That is deliberately the same result for both cases that produce one: an
// unlinked group (Terraform Enterprise serializes those as JSON null, because
// AdminSettings::Scim delegates to the group with allow_nil) and a release
// older than minTFEVersionSiteAuditor that does not know the attribute at all.
// Both mean "no group is linked", which is also the schema default, so plan and
// state agree either way. Note this is the opposite of the SAML resource, whose
// Site Auditor attributes have non-empty defaults and therefore need
// stringOrPrior to tell "absent" apart from "empty".
func modelFromV2SCIMSettings(env models.AdminScimSettingsEnvelopeable) (modelTFESCIMSettings, error) {
	if env == nil || env.GetData() == nil || env.GetData().GetAttributes() == nil {
		return modelTFESCIMSettings{}, fmt.Errorf("SCIM settings response did not contain any data")
	}
	data := env.GetData()
	attrs := data.GetAttributes()

	return modelTFESCIMSettings{
		ID:                          types.StringValue(valueOrZero(data.GetId())),
		Enabled:                     types.BoolValue(valueOrZero(attrs.GetEnabled())),
		Paused:                      types.BoolValue(valueOrZero(attrs.GetPaused())),
		SiteAdminGroupSCIMID:        types.StringValue(valueOrZero(attrs.GetSiteAdminGroupScimId())),
		SiteAdminGroupDisplayName:   types.StringValue(valueOrZero(attrs.GetSiteAdminGroupDisplayName())),
		SiteAuditorGroupSCIMID:      types.StringValue(valueOrZero(attrs.GetSiteAuditorGroupScimId())),
		SiteAuditorGroupDisplayName: types.StringValue(valueOrZero(attrs.GetSiteAuditorGroupDisplayName())),
	}, nil
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
	r.config = client
}

// supportsSiteAuditor reports whether the connected Terraform Enterprise
// supports linking a SCIM group to the Site Auditor role. When the practitioner
// has set site_auditor_group_scim_id in configuration against an older release,
// this records an error rather than letting the request silently drop the
// attribute — older releases do not permit it, so the link would never happen
// and Terraform would report no problem at all.
func (r *resourceTFESCIMSettings) supportsSiteAuditor(config modelTFESCIMSettings, d *diag.Diagnostics) bool {
	meets, err := r.config.MeetsMinRemoteTFEVersion(minTFEVersionSiteAuditor)
	if err != nil {
		d.AddError(
			"Error checking minimum Terraform Enterprise version",
			fmt.Sprintf("Could not determine whether Terraform Enterprise version %s meets the minimum required version %s: %v",
				r.config.RemoteTFEVersion(), minTFEVersionSiteAuditor, err),
		)
		return false
	}
	if meets {
		return true
	}

	// Only fail when the practitioner actually asked for a Site Auditor group.
	// The schema default alone must not break existing configurations on older
	// releases.
	if !config.SiteAuditorGroupSCIMID.IsNull() {
		d.AddError(
			"Terraform Enterprise version does not support Site Auditor",
			fmt.Sprintf("The attribute site_auditor_group_scim_id requires Terraform Enterprise %s or later. This instance reports %s. Remove that attribute from your configuration or upgrade Terraform Enterprise.",
				minTFEVersionSiteAuditor, r.config.RemoteTFEVersion()),
		)
	}
	return false
}

// Schema implements resource.Resource
func (r *resourceTFESCIMSettings) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "(Only for Terraform Enterprise) Manages SCIM provisioning settings for the Terraform Enterprise instance." +
			"\n\nRequires admin token configuration. See example usage for incorporating an admin token in your provider config." +
			"\n\n-> **Note:** SCIM requires SAML to be configured first, so the examples below depend on a `tfe_saml_settings` resource. While this resource exists, SCIM is always `enabled = true`; running `terraform destroy` disables SCIM." +
			"\n\n-> **Note:** `paused`, `site_admin_group_scim_id` and `site_auditor_group_scim_id` are the only mutable arguments. To fully disable SCIM you must run `terraform destroy` on this resource; there is no argument to disable it in-place.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the SCIM settings. Always `scim`.",
				Computed:            true,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether SCIM provisioning is enabled. Always `true` while this resource exists; use `terraform destroy` to disable. If SCIM is disabled outside of Terraform, the next `terraform plan` will propose re-creating this resource.",
				Computed:            true,
			},
			"paused": schema.BoolAttribute{
				MarkdownDescription: "Whether SCIM provisioning is paused. Defaults to `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"site_admin_group_scim_id": schema.StringAttribute{
				MarkdownDescription: "SCIM ID of the group whose members are granted site admin privileges. Defaults to `\"\"` (unlinked).",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"site_admin_group_display_name": schema.StringAttribute{
				Description: "Display name of the group whose members are granted site admin privileges. Empty when no group is linked.",
				Computed:    true,
			},
			"site_auditor_group_scim_id": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf("SCIM ID of the group whose members are granted site auditor privileges. Defaults to `\"\"` (unlinked). Requires Terraform Enterprise %s or later.", minTFEVersionSiteAuditor),
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"site_auditor_group_display_name": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf("Display name of the group whose members are granted site auditor privileges. Empty when no group is linked, and on Terraform Enterprise releases older than %s.", minTFEVersionSiteAuditor),
				Computed:            true,
			},
		},
	}
}

// Read implements resource.Resource
func (r *resourceTFESCIMSettings) Read(ctx context.Context, _ resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "Reading SCIM Settings")

	scimSettings, err := r.config.ClientV2.API.Admin().ScimSettings().Get(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error reading SCIM Settings", "Could not read SCIM Settings, unexpected error: "+apiErrorDetail(err))
		return
	}

	result, err := modelFromV2SCIMSettings(scimSettings)
	if err != nil {
		resp.Diagnostics.AddError("Error reading SCIM Settings", err.Error())
		return
	}

	// If SCIM was disabled out-of-band, signal that the resource no longer exists
	// so Terraform will plan a Create to re-enable it on the next apply.
	// API contract: when Enabled=false the remaining fields (Paused,
	// SiteAdminGroupSCIMID, etc.) are always zero-valued, so no valid settings
	// are lost by removing state here.
	if !result.Enabled.ValueBool() {
		resp.State.RemoveResource(ctx)
		return
	}

	// update state with refreshed data
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

	var config modelTFESCIMSettings
	diags = req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	withSiteAuditor := r.supportsSiteAuditor(config, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating SCIM Settings")
	scimSettings, err := r.updateSCIMSettings(ctx, m, withSiteAuditor)
	if err != nil {
		resp.Diagnostics.AddError("Error creating SCIM Settings", "Could not set SCIM Settings, unexpected error: "+apiErrorDetail(err))
		return
	}

	result, err := modelFromV2SCIMSettings(scimSettings)
	if err != nil {
		resp.Diagnostics.AddError("Error creating SCIM Settings", err.Error())
		return
	}
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

	var config modelTFESCIMSettings
	diags = req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	withSiteAuditor := r.supportsSiteAuditor(config, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Update SCIM Settings")
	scimSettings, err := r.updateSCIMSettings(ctx, m, withSiteAuditor)
	if err != nil {
		resp.Diagnostics.AddError("Error updating SCIM Settings", "Could not set SCIM Settings, unexpected error: "+apiErrorDetail(err))
		return
	}

	result, err := modelFromV2SCIMSettings(scimSettings)
	if err != nil {
		resp.Diagnostics.AddError("Error updating SCIM Settings", err.Error())
		return
	}
	diags = resp.State.Set(ctx, &result)
	resp.Diagnostics.Append(diags...)
}

func (r *resourceTFESCIMSettings) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Debug(ctx, "Delete SCIM Settings")

	// Deleting resets every SCIM setting server-side, including both group
	// links, so there is nothing version-specific to send here.
	if _, err := r.config.ClientV2.API.Admin().ScimSettings().Delete(ctx, nil); err != nil {
		resp.Diagnostics.AddError("Error deleting SCIM Settings", "Could not disable SCIM Settings, unexpected error: "+apiErrorDetail(err))
		return
	}
}

func (r *resourceTFESCIMSettings) ImportState(ctx context.Context, _ resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	scimSettings, err := r.config.ClientV2.API.Admin().ScimSettings().Get(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error importing SCIM Settings", "Could not retrieve SCIM Settings, unexpected error: "+apiErrorDetail(err))
		return
	}

	result, err := modelFromV2SCIMSettings(scimSettings)
	if err != nil {
		resp.Diagnostics.AddError("Error importing SCIM Settings", err.Error())
		return
	}

	if !result.Enabled.ValueBool() {
		resp.Diagnostics.AddError(
			"Cannot import disabled SCIM Settings",
			"SCIM provisioning is currently disabled. Enable SCIM before importing, or use 'terraform apply' to enable it via this resource.",
		)
		return
	}

	diags := resp.State.Set(ctx, &result)
	resp.Diagnostics.Append(diags...)
}

// updateSCIMSettings applies the SCIM settings for Create and Update. The plan
// is the source of truth: every field is sent on every call (schema defaults
// populate fields the user omits). A group ID is sent as the raw value to link
// a group and as "" to unlink one — Terraform Enterprise branches on Ruby's
// .present?, and the generated client omits a nil pointer from the request
// body entirely, which the server reads as "leave this link unchanged".
//
// withSiteAuditor controls whether the Site Auditor group is sent at all:
// releases older than minTFEVersionSiteAuditor drop the attribute from their
// permitted parameters, so sending it there is silently ignored.
func (r *resourceTFESCIMSettings) updateSCIMSettings(ctx context.Context, m modelTFESCIMSettings, withSiteAuditor bool) (models.AdminScimSettingsEnvelopeable, error) {
	siteAdminGroupSCIMID, err := groupSCIMIDForRequest(m.SiteAdminGroupSCIMID, "site_admin_group_scim_id")
	if err != nil {
		return nil, err
	}

	attrs := models.NewAdminScimSettings_attributes()
	attrs.SetEnabled(ptr(true))
	attrs.SetPaused(m.Paused.ValueBoolPointer())
	attrs.SetSiteAdminGroupScimId(siteAdminGroupSCIMID)

	if withSiteAuditor {
		siteAuditorGroupSCIMID, err := groupSCIMIDForRequest(m.SiteAuditorGroupSCIMID, "site_auditor_group_scim_id")
		if err != nil {
			return nil, err
		}
		attrs.SetSiteAuditorGroupScimId(siteAuditorGroupSCIMID)
	}

	s, err := r.config.ClientV2.API.Admin().ScimSettings().Patch(ctx, scimSettingsEnvelope(attrs), nil)
	if err != nil {
		return s, fmt.Errorf("failed to set SCIM Settings: %w", err)
	}
	return s, nil
}

// groupSCIMIDForRequest converts a planned group SCIM ID into the value to send
// on the wire. Null and "" both mean "unlink", and both are sent as "" rather
// than omitted, because an omitted attribute leaves the existing link in place.
func groupSCIMIDForRequest(v types.String, attribute string) (*string, error) {
	if v.IsUnknown() {
		// Can't distinguish "not yet resolved" from an intentional unlink; fail loudly.
		return nil, fmt.Errorf("%s is not yet known; ensure the value is resolved before applying", attribute)
	}
	if v.IsNull() {
		return ptr(""), nil
	}
	return ptr(v.ValueString()), nil
}
