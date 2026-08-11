// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-tfe/v2/api/models"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-tfe/internal/provider/validators"
)

const (
	smtpDefaultPort int64 = 25
)

type modelTFEAdminSMTPSettings struct {
	ID                types.String `tfsdk:"id"`
	Enabled           types.Bool   `tfsdk:"enabled"`
	Host              types.String `tfsdk:"host"`
	Port              types.Int64  `tfsdk:"port"`
	Sender            types.String `tfsdk:"sender"`
	Auth              types.String `tfsdk:"auth"`
	Username          types.String `tfsdk:"username"`
	Password          types.String `tfsdk:"password"`
	PasswordWO        types.String `tfsdk:"password_wo"`
	PasswordWOVersion types.Int64  `tfsdk:"password_wo_version"`
	TestEmailAddress  types.String `tfsdk:"test_email_address"`
}

// resourceTFEAdminSMTPSettings implements the tfe_admin_smtp_settings resource type
type resourceTFEAdminSMTPSettings struct {
	config ConfiguredClient
}

// Configure implements resource.ResourceWithConfigure
func (r *resourceTFEAdminSMTPSettings) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	}
	r.config = client
}

// Metadata implements resource.Resource
func (r *resourceTFEAdminSMTPSettings) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_admin_smtp_settings"
}

// ConfigValidators implements resource.ResourceWithConfigValidators
func (r *resourceTFEAdminSMTPSettings) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.PreferWriteOnlyAttribute(
			path.MatchRoot("password"),
			path.MatchRoot("password_wo"),
		),
	}
}

// Schema implements resource.Resource
func (r *resourceTFEAdminSMTPSettings) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version: 0,
		Description: "(Only for Terraform Enterprise) Creates, updates, and destroys Admin SMTP settings." +
			"\n\nRequires admin token configuration. See example usage for incorporating an admin token in your provider config.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the SMTP settings. Always `smtp`.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether SMTP is enabled. When enabled, all other attributes must have valid values.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"host": schema.StringAttribute{
				Description: "The hostname of the SMTP server.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"port": schema.Int64Attribute{
				Description: "The port of the SMTP server. Defaults to `25`.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(smtpDefaultPort),
			},
			"sender": schema.StringAttribute{
				Description: "The desired sender email address.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"auth": schema.StringAttribute{
				Description: "The authentication type. Valid values are `none`, `plain`, and `login`. Defaults to `none`.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("none"),
				Validators: []validator.String{
					stringvalidator.OneOf(
						"none",
						"plain",
						"login",
					),
				},
			},
			"username": schema.StringAttribute{
				Description: "The username used to authenticate to the SMTP server. Required if auth is `login` or `plain`.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					validators.AttributeValueConflictValidator("auth", []string{"none"}),
				},
			},
			"password": schema.StringAttribute{
				Description: "The password used to authenticate to the SMTP server. Required if auth is `login` or `plain`. Cannot be used with `password_wo`.",
				Optional:    true,
				Sensitive:   true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("password_wo")),
					validators.AttributeValueConflictValidator("auth", []string{"none"}),
				},
			},
			"password_wo": schema.StringAttribute{
				Description: "The password used to authenticate to the SMTP server, guaranteed not to be written to plan or state artifacts. Required if auth is `login` or `plain`. Either `password` or `password_wo` can be provided, but not both. Must be used with `password_wo_version`.",
				Optional:    true,
				Sensitive:   true,
				WriteOnly:   true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("password")),
					validators.AttributeValueConflictValidator("auth", []string{"none"}),
				},
			},
			"password_wo_version": schema.Int64Attribute{
				Optional:    true,
				Description: "Version of the write-only password. Used to trigger updates when the write-only password changes. Must be used with `password_wo`. When `password_wo_version` changes, the write-only password will be updated.",
				Validators: []validator.Int64{
					int64validator.ConflictsWith(path.MatchRoot("password")),
					int64validator.AlsoRequires(path.MatchRoot("password_wo")),
				},
			},
			"test_email_address": schema.StringAttribute{
				Description: "The email address to send a test message to. This value is not persisted and is only used during testing.",
				Optional:    true,
			},
		},
	}
}

// Read implements resource.Resource
func (r *resourceTFEAdminSMTPSettings) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var m modelTFEAdminSMTPSettings
	diags := req.State.Get(ctx, &m)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading Admin SMTP Settings")

	env, err := r.config.ClientV2.API.Admin().SmtpSettings().Get(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error reading Admin SMTP Settings", "Could not read Admin SMTP Settings, unexpected error: "+err.Error())
		return
	}

	// Determine if we should use write-only pattern for password
	isWriteOnly := !m.PasswordWO.IsNull() && !m.PasswordWO.IsUnknown()

	// update state
	result := modelFromTFEAdminSMTPSettingsV2(env.GetData(), m.Password, isWriteOnly)

	// Preserve optional fields from state
	preserveOptionalFields(&result, m)

	diags = resp.State.Set(ctx, &result)
	resp.Diagnostics.Append(diags...)
}

// Create implements resource.Resource
func (r *resourceTFEAdminSMTPSettings) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var m modelTFEAdminSMTPSettings
	diags := req.Plan.Get(ctx, &m)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var config modelTFEAdminSMTPSettings
	diags = req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Create Admin SMTP Settings")
	// Check config for write-only password since plan may not have it populated
	isWriteOnly := !config.PasswordWO.IsNull() && !config.PasswordWO.IsUnknown()
	data, err := r.updateAdminSMTPSettings(ctx, m, config)
	if err != nil {
		resp.Diagnostics.AddError("Error creating AdminSMTP Settings", "Could not set Admin SMTP Settings, unexpected error: "+err.Error())
		return
	}

	result := modelFromTFEAdminSMTPSettingsV2(data, m.Password, isWriteOnly)

	// Preserve optional fields from config
	preserveOptionalFields(&result, config)

	diags = resp.State.Set(ctx, &result)
	resp.Diagnostics.Append(diags...)
}

// Update implements resource.Resource
func (r *resourceTFEAdminSMTPSettings) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var m modelTFEAdminSMTPSettings
	diags := req.Plan.Get(ctx, &m)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var config modelTFEAdminSMTPSettings
	diags = req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state modelTFEAdminSMTPSettings
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Update Admin SMTP Settings")
	// Check config for write-only password since plan may not have it populated
	isWriteOnly := !config.PasswordWO.IsNull() && !config.PasswordWO.IsUnknown()
	data, err := r.updateAdminSMTPSettings(ctx, m, config)
	if err != nil {
		resp.Diagnostics.AddError("Error updating Admin SMTP Settings", "Could not set Admin SMTP Settings, unexpected error: "+err.Error())
		return
	}

	result := modelFromTFEAdminSMTPSettingsV2(data, m.Password, isWriteOnly)

	// Preserve optional fields from config
	preserveOptionalFields(&result, config)

	diags = resp.State.Set(ctx, &result)
	resp.Diagnostics.Append(diags...)
}

// Delete disables the SMTP Settings and then removes the resource from the state file. You cannot delete TFE SMTP Settings, only disable them
func (r *resourceTFEAdminSMTPSettings) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var m modelTFEAdminSMTPSettings
	diags := req.State.Get(ctx, &m)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Delete Admin SMTP Settings")

	attrs := models.NewAdminSmtpSettings_attributes()
	attrs.SetEnabled(ptr(false))
	attrs.SetHost(ptr(""))
	attrs.SetPort(ptr(int32(smtpDefaultPort)))
	attrs.SetSender(ptr(""))
	if authEnum, err := parseAdminSMTPAuth(m.Auth.ValueString()); err == nil && authEnum != nil {
		attrs.SetAuth(authEnum)
	}
	attrs.SetUsername(ptr(""))
	attrs.SetPassword(ptr(""))
	attrs.SetTestEmailAddress(ptr(""))

	_, err := r.config.ClientV2.API.Admin().SmtpSettings().Patch(ctx, buildAdminSMTPSettingsEnvelope(attrs), nil)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting SMTP Settings", "Could not disable SMTP Settings, unexpected error: "+err.Error())
		return
	}
}

// ImportState implements resource.ResourceWithImportState
func (r *resourceTFEAdminSMTPSettings) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	env, err := r.config.ClientV2.API.Admin().SmtpSettings().Get(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error importing Admin SMTP Settings", "Could not retrieve Admin SMTP Settings, unexpected error: "+err.Error())
		return
	}

	result := modelFromTFEAdminSMTPSettingsV2(env.GetData(), types.StringValue(""), false)
	diags := resp.State.Set(ctx, &result)
	resp.Diagnostics.Append(diags...)
}

var (
	_ resource.Resource                = &resourceTFEAdminSMTPSettings{}
	_ resource.ResourceWithConfigure   = &resourceTFEAdminSMTPSettings{}
	_ resource.ResourceWithImportState = &resourceTFEAdminSMTPSettings{}
)

// NewSMTPSettingsResource is a resource function for the framework provider.
func NewAdminSMTPSettingsResource() resource.Resource {
	return &resourceTFEAdminSMTPSettings{}
}

// parseAdminSMTPAuth converts an auth string to the v2 generated enum,
// returning (nil, nil) for an empty string (no auth configured).
func parseAdminSMTPAuth(auth string) (*models.AdminSmtpSettings_attributes_auth, error) {
	if auth == "" {
		return nil, nil
	}
	parsed, err := models.ParseAdminSmtpSettings_attributes_auth(auth)
	if err != nil {
		return nil, err
	}
	if parsed == nil {
		return nil, nil
	}
	authEnum := parsed.(*models.AdminSmtpSettings_attributes_auth)
	return authEnum, nil
}

// buildAdminSMTPSettingsEnvelope wraps attrs in an AdminSmtpSettingsEnvelope
// ready to send to the admin SMTP settings PATCH endpoint.
func buildAdminSMTPSettingsEnvelope(attrs models.AdminSmtpSettings_attributesable) *models.AdminSmtpSettingsEnvelope {
	settingsType := models.SMTPSETTINGS_ADMINSMTPSETTINGS_TYPE
	settingsID := models.SMTP_ADMINSMTPSETTINGS_ID
	data := models.NewAdminSmtpSettings()
	data.SetTypeEscaped(&settingsType)
	data.SetId(&settingsID)
	data.SetAttributes(attrs)

	envelope := models.NewAdminSmtpSettingsEnvelope()
	envelope.SetData(data)
	return envelope
}

// updateSMTPSettings was created to keep the code DRY. It is used in both Create and Update functions
func (r *resourceTFEAdminSMTPSettings) updateAdminSMTPSettings(ctx context.Context, m modelTFEAdminSMTPSettings, config modelTFEAdminSMTPSettings) (models.AdminSmtpSettingsable, error) {
	// Use password from config since write-only attributes aren't in the plan
	curPass := config.Password
	if !config.PasswordWO.IsNull() && !config.PasswordWO.IsUnknown() {
		curPass = config.PasswordWO
	}

	attrs := models.NewAdminSmtpSettings_attributes()
	attrs.SetEnabled(m.Enabled.ValueBoolPointer())
	attrs.SetHost(m.Host.ValueStringPointer())
	attrs.SetPort(ptr(int32(m.Port.ValueInt64()))) //nolint:gosec // port is a network port number, always well within int32 range
	attrs.SetSender(m.Sender.ValueStringPointer())
	authEnum, err := parseAdminSMTPAuth(m.Auth.ValueString())
	if err != nil {
		return nil, fmt.Errorf("failed to parse auth %q: %w", m.Auth.ValueString(), err)
	}
	attrs.SetAuth(authEnum)
	attrs.SetUsername(m.Username.ValueStringPointer())
	attrs.SetPassword(curPass.ValueStringPointer())
	attrs.SetTestEmailAddress(m.TestEmailAddress.ValueStringPointer())

	env, err := r.config.ClientV2.API.Admin().SmtpSettings().Patch(ctx, buildAdminSMTPSettingsEnvelope(attrs), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to update Admin SMTP Settings: %w", err)
	}
	if env == nil || env.GetData() == nil {
		return nil, fmt.Errorf("failed to update Admin SMTP Settings: API returned no data")
	}
	return env.GetData(), nil
}

// modelFromTFEAdminSMTPSettingsV2 builds a modelTFEAdminSMTPSettings struct from a v2 AdminSmtpSettingsable value
func modelFromTFEAdminSMTPSettingsV2(v models.AdminSmtpSettingsable, password types.String, isWriteOnly bool) modelTFEAdminSMTPSettings {
	m := modelTFEAdminSMTPSettings{
		Password: types.StringValue(""),
	}

	if id := v.GetId(); id != nil {
		m.ID = types.StringValue(id.String())
	}

	if attrs := v.GetAttributes(); attrs != nil {
		m.Enabled = types.BoolValue(valueOrZero(attrs.GetEnabled()))
		m.Host = types.StringValue(valueOrZero(attrs.GetHost()))
		m.Port = types.Int64Value(int64(valueOrZero(attrs.GetPort())))
		m.Sender = types.StringValue(valueOrZero(attrs.GetSender()))
		m.Auth = types.StringValue(enumStringOrEmpty(attrs.GetAuth()))
		m.Username = types.StringValue(valueOrZero(attrs.GetUsername()))
	}

	if len(password.ValueString()) > 0 {
		m.Password = password
	}

	// Don't retrieve values if write-only is being used. Unset the password field before updating the state.
	if isWriteOnly {
		m.Password = types.StringValue("")
	}

	return m
}

// preserveOptionalFields updates the result model with preserved values from source model
func preserveOptionalFields(result *modelTFEAdminSMTPSettings, source modelTFEAdminSMTPSettings) {
	// Preserve null values for optional fields
	if source.Host.IsNull() {
		result.Host = types.StringNull()
	}
	if source.Sender.IsNull() {
		result.Sender = types.StringNull()
	}
	if source.Username.IsNull() {
		result.Username = types.StringNull()
	}
	if source.Password.IsNull() {
		result.Password = types.StringNull()
	}
	// Preserve password_wo_version
	if !source.PasswordWOVersion.IsNull() {
		result.PasswordWOVersion = source.PasswordWOVersion
	}
	// Preserve test_email_address since API doesn't return it
	if !source.TestEmailAddress.IsNull() {
		result.TestEmailAddress = source.TestEmailAddress
	}
}
