// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"

	"github.com/hashicorp/go-tfe/v2/api/models"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	customValidators "github.com/hashicorp/terraform-provider-tfe/internal/provider/validators"
)

// minTFEVersionAdminCustomizationSettings is the minimum Terraform Enterprise
// version that supports the admin customization settings managed by this
// resource (internal URL overrides and avatar source configuration).
const minTFEVersionAdminCustomizationSettings = "2.1.0"

const internalURLMaxLength = 2048

var (
	// supportEmailRegexp matches email addresses including localhost/internal domains,
	// aligned with Atlas's email validation (URI::MailTo::EMAIL_REGEXP).
	supportEmailRegexp = regexp.MustCompile(`^[a-zA-Z0-9.!\#$%&'*+/=?^_\x60{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)
	// noURLSchemeRegexp rejects values containing a scheme such as "https://".
	noURLSchemeRegexp = regexp.MustCompile(`^[A-Za-z0-9.-]+(?::\d+)?(?:/\S*)?$`)
)

// useStateForUnknownString keeps Optional+Computed string attributes stable in
// plans when they are not configured.
func useStateForUnknownString() []planmodifier.String {
	return []planmodifier.String{stringplanmodifier.UseStateForUnknown()}
}

// privateKeyOriginalCustomization is the private state key used to record the
// server-side customization values observed before this resource first managed
// them. They are restored on delete so that removing this resource leaves the
// instance as it was found rather than wiping every setting to empty.
const privateKeyOriginalCustomization = "original_customization_settings"

// originalCustomizationSnapshot captures the customization values that existed
// before this resource took ownership of configured attributes, so delete can
// restore them without overwriting unmanaged attributes.
type originalCustomizationSnapshot struct {
	SupportEmail             *string             `json:"support_email,omitempty"`
	SupportURLAddress        *string             `json:"support_url_address,omitempty"`
	LoginHelp                *string             `json:"login_help,omitempty"`
	Footer                   *string             `json:"footer,omitempty"`
	Error                    *string             `json:"error,omitempty"`
	NewUser                  *string             `json:"new_user,omitempty"`
	InternalSupportURL       *string             `json:"internal_support_url,omitempty"`
	InternalDocumentationURL *string             `json:"internal_documentation_url,omitempty"`
	InternalTutorialsURL     *string             `json:"internal_tutorials_url,omitempty"`
	AvatarSources            []adminAvatarSource `json:"avatar_sources,omitempty"`
}

func (s originalCustomizationSnapshot) isEmpty() bool {
	return s.SupportEmail == nil &&
		s.SupportURLAddress == nil &&
		s.LoginHelp == nil &&
		s.Footer == nil &&
		s.Error == nil &&
		s.NewUser == nil &&
		s.InternalSupportURL == nil &&
		s.InternalDocumentationURL == nil &&
		s.InternalTutorialsURL == nil &&
		s.AvatarSources == nil
}

// modelTFEAdminCustomizationSettings maps the tfe_admin_customization_settings
// resource schema to a Go struct.
type modelTFEAdminCustomizationSettings struct {
	ID                       types.String `tfsdk:"id"`
	SupportEmail             types.String `tfsdk:"support_email"`
	SupportURLAddress        types.String `tfsdk:"support_url_address"`
	InternalSupportURL       types.String `tfsdk:"internal_support_url"`
	InternalDocumentationURL types.String `tfsdk:"internal_documentation_url"`
	InternalTutorialsURL     types.String `tfsdk:"internal_tutorials_url"`
	LoginHelp                types.String `tfsdk:"login_help"`
	Footer                   types.String `tfsdk:"footer"`
	Error                    types.String `tfsdk:"error"`
	NewUser                  types.String `tfsdk:"new_user"`
	// AvatarSources maps an avatar source name to whether it is enabled. Using a
	// map (rather than an ordered list) keeps the attribute stable regardless of
	// the order or number of sources Terraform Enterprise reports.
	AvatarSources types.Map `tfsdk:"avatar_sources"`
}

// resourceTFEAdminCustomizationSettings implements the
// tfe_admin_customization_settings resource type.
type resourceTFEAdminCustomizationSettings struct {
	config ConfiguredClient
}

var (
	_ resource.Resource                = &resourceTFEAdminCustomizationSettings{}
	_ resource.ResourceWithConfigure   = &resourceTFEAdminCustomizationSettings{}
	_ resource.ResourceWithImportState = &resourceTFEAdminCustomizationSettings{}
)

// NewAdminCustomizationSettingsResource is a resource function for the framework provider.
func NewAdminCustomizationSettingsResource() resource.Resource {
	return &resourceTFEAdminCustomizationSettings{}
}

// Configure implements resource.ResourceWithConfigure
func (r *resourceTFEAdminCustomizationSettings) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *resourceTFEAdminCustomizationSettings) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_admin_customization_settings"
}

// Schema implements resource.Resource
func (r *resourceTFEAdminCustomizationSettings) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version: 0,
		Description: "(Only for Terraform Enterprise) Manages the admin customization settings, including support and " +
			"documentation link routing, avatar source configuration, and login/footer/error/new-user messaging." +
			"\n\nRequires admin token configuration. See example usage for incorporating an admin token in your provider config." +
			"\n\nRequires Terraform Enterprise v" + minTFEVersionAdminCustomizationSettings + " or later." +
			"\n\nThis is a singleton resource: there is only ever one set of customization settings for a TFE instance. " +
			"Destroying this resource restores managed attributes to the values that were in effect before Terraform managed them. " +
			"If an imported resource is destroyed without having modified any settings, Terraform Enterprise settings remain unchanged." +
			"\n\nAll attributes are optional; any attribute omitted from configuration keeps its current server-side value. " +
			"Removing an attribute from configuration does not reset it — set it explicitly (for example, to an empty string) instead.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the customization settings.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"support_email": schema.StringAttribute{
				Description: "The support email address presented to users. Set to an empty string to clear it, " +
					"in which case `support_url_address` is used instead.",
				Optional:      true,
				Computed:      true,
				PlanModifiers: useStateForUnknownString(),
				Validators: []validator.String{
					stringvalidator.Any(
						stringvalidator.LengthAtMost(0),
						stringvalidator.RegexMatches(supportEmailRegexp, "must be a valid email address"),
					),
				},
			},
			"support_url_address": schema.StringAttribute{
				Description: "The support hostname presented to users, without a scheme (for example, `support.example.com`). " +
					"Terraform Enterprise adds the scheme when rendering the link. Used as the fallback when `support_email` is not set. " +
					"Terraform Enterprise requires this value, so it cannot be empty.",
				Optional:      true,
				Computed:      true,
				PlanModifiers: useStateForUnknownString(),
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					stringvalidator.RegexMatches(noURLSchemeRegexp, "must be a hostname without a scheme (for example, support.example.com)"),
				},
			},
			"internal_support_url": schema.StringAttribute{
				Description: "An internal support URL that overrides the default HashiCorp support link. " +
					"Accepts an `https://`, `http://`, or `mailto:` value. " +
					"Set to an empty string to clear the override and restore the default link.",
				Optional:      true,
				Computed:      true,
				PlanModifiers: useStateForUnknownString(),
				Validators: []validator.String{
					stringvalidator.LengthAtMost(internalURLMaxLength),
					customValidators.URLWithSchemes("https", "http", "mailto"),
				},
			},
			"internal_documentation_url": schema.StringAttribute{
				Description: "An internal documentation URL that overrides the default HashiCorp documentation link. " +
					"Accepts an `https://` or `http://` value. " +
					"Set to an empty string to clear the override and restore the default link.",
				Optional:      true,
				Computed:      true,
				PlanModifiers: useStateForUnknownString(),
				Validators: []validator.String{
					stringvalidator.LengthAtMost(internalURLMaxLength),
					customValidators.URLWithSchemes("https", "http"),
				},
			},
			"internal_tutorials_url": schema.StringAttribute{
				Description: "An internal tutorials URL that overrides the default HashiCorp tutorials link. " +
					"Accepts an `https://` or `http://` value. " +
					"Set to an empty string to clear the override and restore the default link.",
				Optional:      true,
				Computed:      true,
				PlanModifiers: useStateForUnknownString(),
				Validators: []validator.String{
					stringvalidator.LengthAtMost(internalURLMaxLength),
					customValidators.URLWithSchemes("https", "http"),
				},
			},
			"login_help": schema.StringAttribute{
				Description:   "Custom help text shown on the login page. Set to an empty string to clear it.",
				Optional:      true,
				Computed:      true,
				PlanModifiers: useStateForUnknownString(),
			},
			"footer": schema.StringAttribute{
				Description:   "Custom footer content shown throughout the application. Set to an empty string to clear it.",
				Optional:      true,
				Computed:      true,
				PlanModifiers: useStateForUnknownString(),
			},
			"error": schema.StringAttribute{
				Description:   "Custom instructions shown to users when an error occurs. Set to an empty string to clear it.",
				Optional:      true,
				Computed:      true,
				PlanModifiers: useStateForUnknownString(),
			},
			"new_user": schema.StringAttribute{
				Description:   "Custom instructions shown to new users. Set to an empty string to clear it.",
				Optional:      true,
				Computed:      true,
				PlanModifiers: useStateForUnknownString(),
			},
			"avatar_sources": schema.MapAttribute{
				Description: "A map of avatar source name to whether it is enabled (for example " +
					"`{ gravatar = true }`). Source names are validated against the avatar sources known to the " +
					"provider's API client (currently `gravatar`); new sources require a provider upgrade. Only the sources " +
					"you specify are managed by Terraform; any others reported by Terraform Enterprise are left " +
					"untouched. If omitted, the current configuration is preserved.",
				Optional:    true,
				Computed:    true,
				ElementType: types.BoolType,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Map{
					mapvalidator.KeysAre(
						customValidators.EnumMember(
							models.ParseAdminCustomizationSettings_attributes_avatarSources_source,
							"avatar source",
						),
					),
				},
			},
		},
	}
}

// Read implements resource.Resource
func (r *resourceTFEAdminCustomizationSettings) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state modelTFEAdminCustomizationSettings
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading Admin Customization Settings")
	settings, err := readAdminCustomizationSettings(ctx, r.config.ClientV2)
	if err != nil {
		resp.Diagnostics.AddError("Error reading Admin Customization Settings", "Could not read Admin Customization Settings, unexpected error: "+err.Error())
		return
	}

	result, diags := modelFromTFEAdminCustomizationSettings(ctx, settings, state.AvatarSources)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

// Create implements resource.Resource
func (r *resourceTFEAdminCustomizationSettings) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan modelTFEAdminCustomizationSettings
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var config modelTFEAdminCustomizationSettings
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating Admin Customization Settings")
	if !r.checkMinTFEVersion(&resp.Diagnostics) {
		return
	}

	// Capture the values that existed before this resource took ownership of
	// the configured attributes, so they can be restored on delete.
	tflog.Debug(ctx, "Reading Admin Customization Settings before create")
	current, err := readAdminCustomizationSettings(ctx, r.config.ClientV2)
	if err != nil {
		resp.Diagnostics.AddError("Error creating Admin Customization Settings", "Could not read existing Admin Customization Settings, unexpected error: "+err.Error())
		return
	}
	resp.Diagnostics.Append(setOriginalCustomization(ctx, resp.Private, current, config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, diags := r.updateAdminCustomizationSettings(ctx, plan, config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

// Update implements resource.Resource
func (r *resourceTFEAdminCustomizationSettings) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan modelTFEAdminCustomizationSettings
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var config modelTFEAdminCustomizationSettings
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating Admin Customization Settings")
	if !r.checkMinTFEVersion(&resp.Diagnostics) {
		return
	}

	// If new attributes are now configured that were not previously recorded
	// (including on an imported resource), capture their pre-management values.
	existingSnapshot, _, diags := getOriginalCustomization(ctx, req.Private)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	current, err := readAdminCustomizationSettings(ctx, r.config.ClientV2)
	if err != nil {
		resp.Diagnostics.AddError("Error updating Admin Customization Settings", "Could not read existing Admin Customization Settings, unexpected error: "+err.Error())
		return
	}
	updatedSnapshot := mergeSnapshot(existingSnapshot, current, config)
	resp.Diagnostics.Append(saveOriginalCustomizationSnapshot(ctx, resp.Private, updatedSnapshot)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, diags := r.updateAdminCustomizationSettings(ctx, plan, config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

// Delete restores the customization values that existed before this resource
// managed them, then removes the resource from state. Terraform Enterprise has
// no delete endpoint for this singleton, so the values are restored via a PATCH.
func (r *resourceTFEAdminCustomizationSettings) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Debug(ctx, "Deleting Admin Customization Settings")

	// Delete intentionally skips the minimum version check: the resource could
	// only have been created against a supported version, and destroy must not
	// be blocked by a transient version-detection failure.
	original, ok, diags := getOriginalCustomization(ctx, req.Private)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Without a recorded snapshot (for example, an imported resource that never
	// ran through Create, or a resource that managed no attributes), leave the
	// server-side settings untouched.
	if !ok || original.isEmpty() {
		resp.Diagnostics.AddWarning(
			"Admin Customization Settings left unchanged",
			"No pre-management snapshot was recorded for this resource (for example, because it was imported or unconfigured), "+
				"so the customization settings were removed from state without being modified in Terraform Enterprise.",
		)
		return
	}

	var avatarSourcesToRestore []adminAvatarSource
	if len(original.AvatarSources) > 0 {
		current, err := readAdminCustomizationSettings(ctx, r.config.ClientV2)
		if err != nil {
			resp.Diagnostics.AddError("Error deleting Admin Customization Settings", "Could not read current avatar sources, unexpected error: "+err.Error())
			return
		}
		avatarSourcesToRestore = mergeAvatarSources(current.AvatarSources, original.AvatarSources)
	}

	_, err := patchAdminCustomizationSettings(ctx, r.config.ClientV2, adminCustomizationSettingsUpdate{
		SupportEmail:             original.SupportEmail,
		SupportURLAddress:        original.SupportURLAddress,
		LoginHelp:                original.LoginHelp,
		Footer:                   original.Footer,
		Error:                    original.Error,
		NewUser:                  original.NewUser,
		InternalSupportURL:       original.InternalSupportURL,
		InternalDocumentationURL: original.InternalDocumentationURL,
		InternalTutorialsURL:     original.InternalTutorialsURL,
		AvatarSources:            avatarSourcesToRestore,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error deleting Admin Customization Settings", "Could not restore Admin Customization Settings, unexpected error: "+err.Error())
		return
	}
}

// ImportState implements resource.ResourceWithImportState
func (r *resourceTFEAdminCustomizationSettings) ImportState(ctx context.Context, _ resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	settings, err := readAdminCustomizationSettings(ctx, r.config.ClientV2)
	if err != nil {
		resp.Diagnostics.AddError("Error importing Admin Customization Settings", "Could not retrieve Admin Customization Settings, unexpected error: "+err.Error())
		return
	}

	result, diags := modelFromTFEAdminCustomizationSettings(ctx, settings, types.MapNull(types.BoolType))
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

// updateAdminCustomizationSettings builds the update options from the plan and
// config, issues the PATCH (omitting unconfigured fields), and maps the result
// back into the resource model.
func (r *resourceTFEAdminCustomizationSettings) updateAdminCustomizationSettings(ctx context.Context, plan modelTFEAdminCustomizationSettings, config modelTFEAdminCustomizationSettings) (modelTFEAdminCustomizationSettings, diag.Diagnostics) {
	var diags diag.Diagnostics

	options := adminCustomizationSettingsUpdate{
		SupportEmail:             configuredStringPointer(config.SupportEmail, plan.SupportEmail),
		SupportURLAddress:        configuredStringPointer(config.SupportURLAddress, plan.SupportURLAddress),
		LoginHelp:                configuredStringPointer(config.LoginHelp, plan.LoginHelp),
		Footer:                   configuredStringPointer(config.Footer, plan.Footer),
		Error:                    configuredStringPointer(config.Error, plan.Error),
		NewUser:                  configuredStringPointer(config.NewUser, plan.NewUser),
		InternalSupportURL:       configuredStringPointer(config.InternalSupportURL, plan.InternalSupportURL),
		InternalDocumentationURL: configuredStringPointer(config.InternalDocumentationURL, plan.InternalDocumentationURL),
		InternalTutorialsURL:     configuredStringPointer(config.InternalTutorialsURL, plan.InternalTutorialsURL),
	}

	if isMapConfigured(config.AvatarSources) {
		configured, d := avatarSourcesFromMap(ctx, plan.AvatarSources)
		diags.Append(d...)
		if diags.HasError() {
			return modelTFEAdminCustomizationSettings{}, diags
		}

		// Terraform Enterprise treats the PATCHed avatar_sources as the complete
		// set of enabled sources: any source omitted from the request is
		// disabled. Merge the configured sources over the current server-side
		// catalog so sources the practitioner does not manage are left untouched.
		current, err := readAdminCustomizationSettings(ctx, r.config.ClientV2)
		if err != nil {
			diags.AddError("Error updating Admin Customization Settings", "Could not read current avatar sources, unexpected error: "+err.Error())
			return modelTFEAdminCustomizationSettings{}, diags
		}
		options.AvatarSources = mergeAvatarSources(current.AvatarSources, configured)
	}

	settings, err := patchAdminCustomizationSettings(ctx, r.config.ClientV2, options)
	if err != nil {
		diags.AddError("Error updating Admin Customization Settings", "Could not update Admin Customization Settings, unexpected error: "+err.Error())
		return modelTFEAdminCustomizationSettings{}, diags
	}

	result, d := modelFromTFEAdminCustomizationSettings(ctx, settings, plan.AvatarSources)
	diags.Append(d...)
	return result, diags
}

// modelFromTFEAdminCustomizationSettings maps a go-tfe customization setting to
// the resource model.
//
// priorSources is the avatar_sources value from the plan/prior state and
// determines which source keys the practitioner is managing. Terraform
// Enterprise always returns its full catalog, but we only project the subset of
// keys the practitioner configured (or that already existed in state) back into
// the attribute. This keeps the attribute stable — and free of "inconsistent
// result after apply" errors — even if the server-side catalog is expanded to
// include new sources the provider does not yet know about.
func modelFromTFEAdminCustomizationSettings(ctx context.Context, v *adminCustomizationSettings, priorSources types.Map) (modelTFEAdminCustomizationSettings, diag.Diagnostics) {
	var diags diag.Diagnostics

	avatarMap, d := avatarSourcesToMap(ctx, v.AvatarSources, priorSources)
	diags.Append(d...)

	m := modelTFEAdminCustomizationSettings{
		ID:                       types.StringValue(v.ID),
		SupportEmail:             types.StringValue(v.SupportEmail),
		SupportURLAddress:        types.StringValue(v.SupportURLAddress),
		InternalSupportURL:       types.StringValue(v.InternalSupportURL),
		InternalDocumentationURL: types.StringValue(v.InternalDocumentationURL),
		InternalTutorialsURL:     types.StringValue(v.InternalTutorialsURL),
		LoginHelp:                types.StringValue(v.LoginHelp),
		Footer:                   types.StringValue(v.Footer),
		Error:                    types.StringValue(v.Error),
		NewUser:                  types.StringValue(v.NewUser),
		AvatarSources:            avatarMap,
	}
	return m, diags
}

// checkMinTFEVersion verifies the connected Terraform Enterprise meets the
// minimum version that supports the admin customization settings. It records an
// error diagnostic and returns false when the requirement is not met, so a
// write against HCP Terraform or an older release fails loudly instead of silently dropping
// unsupported attributes.
func (r *resourceTFEAdminCustomizationSettings) checkMinTFEVersion(diags *diag.Diagnostics) bool {
	meets, err := r.config.MeetsMinRemoteTFEOnlyVersion(minTFEVersionAdminCustomizationSettings)
	if err != nil {
		diags.AddError(
			"Error checking minimum Terraform Enterprise version",
			fmt.Sprintf("Could not determine whether Terraform Enterprise version %s meets the minimum required version %s: %v",
				r.config.RemoteTFEVersion(), minTFEVersionAdminCustomizationSettings, err),
		)
		return false
	}
	if !meets {
		if r.config.Client.IsCloud() {
			diags.AddError(
				"Unsupported platform",
				"tfe_admin_customization_settings is only supported on Terraform Enterprise and cannot be used with HCP Terraform.",
			)
		} else {
			diags.AddError(
				"Unsupported Terraform Enterprise version",
				fmt.Sprintf("tfe_admin_customization_settings requires Terraform Enterprise version %s or later, but the connected instance is running version %s.",
					minTFEVersionAdminCustomizationSettings, r.config.RemoteTFEVersion()),
			)
		}
		return false
	}
	return true
}

// configuredStringPointer returns a pointer to the planned string value when the
// attribute is present in configuration (non-null and known), or nil when omitted.
// An empty string in configuration is preserved so callers can clear an override.
func configuredStringPointer(configVal, planVal types.String) *string {
	if configVal.IsNull() || configVal.IsUnknown() {
		return nil
	}
	return planVal.ValueStringPointer()
}

func isConfigured(s types.String) bool {
	return !s.IsNull() && !s.IsUnknown()
}

func isMapConfigured(m types.Map) bool {
	return !m.IsNull() && !m.IsUnknown()
}

func buildSnapshot(current *adminCustomizationSettings, config modelTFEAdminCustomizationSettings) originalCustomizationSnapshot {
	return mergeSnapshot(originalCustomizationSnapshot{}, current, config)
}

func mergeSnapshot(existing originalCustomizationSnapshot, current *adminCustomizationSettings, config modelTFEAdminCustomizationSettings) originalCustomizationSnapshot {
	s := existing
	if s.SupportEmail == nil && isConfigured(config.SupportEmail) {
		s.SupportEmail = ptr(current.SupportEmail)
	}
	if s.SupportURLAddress == nil && isConfigured(config.SupportURLAddress) {
		s.SupportURLAddress = ptr(current.SupportURLAddress)
	}
	if s.LoginHelp == nil && isConfigured(config.LoginHelp) {
		s.LoginHelp = ptr(current.LoginHelp)
	}
	if s.Footer == nil && isConfigured(config.Footer) {
		s.Footer = ptr(current.Footer)
	}
	if s.Error == nil && isConfigured(config.Error) {
		s.Error = ptr(current.Error)
	}
	if s.NewUser == nil && isConfigured(config.NewUser) {
		s.NewUser = ptr(current.NewUser)
	}
	if s.InternalSupportURL == nil && isConfigured(config.InternalSupportURL) {
		s.InternalSupportURL = ptr(current.InternalSupportURL)
	}
	if s.InternalDocumentationURL == nil && isConfigured(config.InternalDocumentationURL) {
		s.InternalDocumentationURL = ptr(current.InternalDocumentationURL)
	}
	if s.InternalTutorialsURL == nil && isConfigured(config.InternalTutorialsURL) {
		s.InternalTutorialsURL = ptr(current.InternalTutorialsURL)
	}
	if isMapConfigured(config.AvatarSources) {
		configuredKeys := make(map[string]bool)
		for k := range config.AvatarSources.Elements() {
			configuredKeys[k] = true
		}
		existingKeys := make(map[string]bool, len(s.AvatarSources))
		for _, src := range s.AvatarSources {
			existingKeys[src.Source] = true
		}
		for _, src := range current.AvatarSources {
			if configuredKeys[src.Source] && !existingKeys[src.Source] {
				s.AvatarSources = append(s.AvatarSources, src)
			}
		}
	}
	return s
}

func saveOriginalCustomizationSnapshot(ctx context.Context, private privateStateWriter, snapshot originalCustomizationSnapshot) diag.Diagnostics {
	encoded, err := json.Marshal(snapshot)
	if err != nil {
		var diags diag.Diagnostics
		diags.AddError("Error storing Admin Customization Settings state", "Could not encode the original customization values: "+err.Error())
		return diags
	}
	return private.SetKey(ctx, privateKeyOriginalCustomization, encoded)
}

// avatarSourcesFromMap converts the resource map attribute (source name =>
// enabled) into the go-tfe avatar source slice, in a deterministic order.
func avatarSourcesFromMap(ctx context.Context, m types.Map) ([]adminAvatarSource, diag.Diagnostics) {
	var diags diag.Diagnostics
	if m.IsNull() || m.IsUnknown() {
		return nil, diags
	}

	enabledBySource := make(map[string]bool, len(m.Elements()))
	diags.Append(m.ElementsAs(ctx, &enabledBySource, false)...)
	if diags.HasError() {
		return nil, diags
	}

	// Sort keys so the request body is stable across runs.
	names := make([]string, 0, len(enabledBySource))
	for name := range enabledBySource {
		names = append(names, name)
	}
	sort.Strings(names)

	sources := make([]adminAvatarSource, 0, len(names))
	for _, name := range names {
		sources = append(sources, adminAvatarSource{
			Source:  name,
			Enabled: enabledBySource[name],
		})
	}
	return sources, diags
}

// avatarSourcesToMap converts a go-tfe avatar source slice into the resource map
// attribute. Only the source keys present in priorSources are projected back
// into state; any other sources reported by Terraform Enterprise are ignored so
// that a server-side catalog expansion cannot introduce drift. When priorSources
// is null/unknown (for example on import), the full catalog reported by the API
// is projected instead.
func avatarSourcesToMap(ctx context.Context, sources []adminAvatarSource, priorSources types.Map) (types.Map, diag.Diagnostics) {
	enabledBySource := make(map[string]bool, len(sources))
	for _, s := range sources {
		enabledBySource[s.Source] = s.Enabled
	}

	// Determine which keys to keep. Default to the full catalog (import/read
	// with no prior config), otherwise restrict to the managed keys.
	var keys []string
	if !priorSources.IsNull() && !priorSources.IsUnknown() {
		for name := range priorSources.Elements() {
			keys = append(keys, name)
		}
	} else {
		for name := range enabledBySource {
			keys = append(keys, name)
		}
	}

	projected := make(map[string]bool, len(keys))
	for _, name := range keys {
		// A managed source that the API did not report is treated as disabled.
		// Terraform Enterprise always reports its full catalog, and plan-time
		// validation only admits catalog members, so this only happens if the
		// server stops recognising a source.
		projected[name] = enabledBySource[name]
	}

	return types.MapValueFrom(ctx, types.BoolType, projected)
}

// mergeAvatarSources overlays the configured avatar sources on the current
// server-side catalog. Terraform Enterprise persists only the sources marked
// enabled in a PATCH, so sending the configured subset alone would disable any
// unmanaged source. The result preserves catalog order, then appends any
// configured sources the catalog did not report.
func mergeAvatarSources(current, configured []adminAvatarSource) []adminAvatarSource {
	desired := make(map[string]bool, len(configured))
	for _, s := range configured {
		desired[s.Source] = s.Enabled
	}

	merged := make([]adminAvatarSource, 0, len(current)+len(configured))
	seen := make(map[string]bool, len(current))
	for _, s := range current {
		if enabled, ok := desired[s.Source]; ok {
			s.Enabled = enabled
		}
		merged = append(merged, s)
		seen[s.Source] = true
	}
	for _, s := range configured {
		if !seen[s.Source] {
			merged = append(merged, s)
		}
	}
	return merged
}

// setOriginalCustomization records the pre-management customization values for
// attributes configured in this resource, so delete can restore them without
// clobbering unmanaged attributes.
func setOriginalCustomization(ctx context.Context, private privateStateWriter, v *adminCustomizationSettings, config modelTFEAdminCustomizationSettings) diag.Diagnostics {
	snapshot := buildSnapshot(v, config)
	return saveOriginalCustomizationSnapshot(ctx, private, snapshot)
}

// getOriginalCustomization returns the recorded pre-management snapshot. The
// second return value is false when no snapshot was ever recorded (for example,
// after an import that never ran through Create).
func getOriginalCustomization(ctx context.Context, private privateStateReader) (originalCustomizationSnapshot, bool, diag.Diagnostics) {
	encoded, diags := private.GetKey(ctx, privateKeyOriginalCustomization)
	if diags.HasError() || len(encoded) == 0 {
		return originalCustomizationSnapshot{}, false, diags
	}

	var snapshot originalCustomizationSnapshot
	if err := json.Unmarshal(encoded, &snapshot); err != nil {
		diags.AddError("Error reading Admin Customization Settings state", "Could not decode the original customization values: "+err.Error())
		return originalCustomizationSnapshot{}, false, diags
	}
	return snapshot, true, diags
}
