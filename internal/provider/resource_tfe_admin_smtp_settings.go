// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	tfe "github.com/hashicorp/go-tfe"
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
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
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
	client *tfe.Client
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
	r.client = client.Client
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
		Version:     0,
		Description: "Manages SMTP settings for Terraform Enterprise.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the SMTP settings. Always 'smtp'.",
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
				Description: "The port of the SMTP server.",
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
				Description: "The authentication type. Valid values are 'none', 'plain', and 'login'.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(string(tfe.SMTPAuthNone)),
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(tfe.SMTPAuthNone),
						string(tfe.SMTPAuthPlain),
						string(tfe.SMTPAuthLogin),
					),
				},
			},
			"username": schema.StringAttribute{
				Description: "The username used to authenticate to the SMTP server. Required if auth is 'login' or 'plain'.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					validators.AttributeValueConflictValidator("auth", []string{string(tfe.SMTPAuthNone)}),
				},
			},
			"password": schema.StringAttribute{
				Description: "The password used to authenticate to the SMTP server. Required if auth is 'login' or 'plain'.",
				Optional:    true,
				Sensitive:   true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("password_wo")),
					validators.AttributeValueConflictValidator("auth", []string{string(tfe.SMTPAuthNone)}),
				},
			},
			"password_wo": schema.StringAttribute{
				Description: "The password in write only used to authenticate to the SMTP server. Required if auth is 'login' or 'plain'.",
				Optional:    true,
				Sensitive:   true,
				WriteOnly:   true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("password")),
					validators.AttributeValueConflictValidator("auth", []string{string(tfe.SMTPAuthNone)}),
				},
			},
			"password_wo_version": schema.Int64Attribute{
				Optional:    true,
				Description: "Version of the write-only private key to trigger updates",
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

	smtpSettings, err := r.client.Admin.Settings.SMTP.Read(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading Admin SMTP Settings", "Could not read Admin SMTP Settings, unexpected error: "+err.Error())
		return
	}

	// Determine if we should use write-only pattern for password
	isWriteOnly := !m.PasswordWO.IsNull() && !m.PasswordWO.IsUnknown()

	// update state
	result := modelFromTFEAdminSMTPSettings(smtpSettings, m.Password, isWriteOnly)

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
	smtpSettings, err := r.updateAdminSMTPSettings(ctx, m, config)
	if err != nil {
		resp.Diagnostics.AddError("Error creating AdminSMTP Settings", "Could not set Admin SMTP Settings, unexpected error: "+err.Error())
		return
	}

	result := modelFromTFEAdminSMTPSettings(smtpSettings, m.Password, isWriteOnly)

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
	smtpSettings, err := r.updateAdminSMTPSettings(ctx, m, config)
	if err != nil {
		resp.Diagnostics.AddError("Error updating Admin SMTP Settings", "Could not set Admin SMTP Settings, unexpected error: "+err.Error())
		return
	}

	result := modelFromTFEAdminSMTPSettings(smtpSettings, m.Password, isWriteOnly)

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
	_, err := r.client.Admin.Settings.SMTP.Update(ctx, tfe.AdminSMTPSettingsUpdateOptions{
		Enabled:          basetypes.NewBoolValue(false).ValueBoolPointer(),
		Host:             basetypes.NewStringValue("").ValueStringPointer(),
		Port:             tfe.Int(int(smtpDefaultPort)),
		Sender:           basetypes.NewStringValue("").ValueStringPointer(),
		Auth:             (*tfe.SMTPAuthType)(m.Auth.ValueStringPointer()),
		Username:         basetypes.NewStringValue("").ValueStringPointer(),
		Password:         basetypes.NewStringValue("").ValueStringPointer(),
		TestEmailAddress: basetypes.NewStringValue("").ValueStringPointer(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error deleting SMTP Settings", "Could not disable SMTP Settings, unexpected error: "+err.Error())
		return
	}
}

// ImportState implements resource.ResourceWithImportState
func (r *resourceTFEAdminSMTPSettings) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	smtpSettings, err := r.client.Admin.Settings.SMTP.Read(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error importing Admin SMTP Settings", "Could not retrieve Admin SMTP Settings, unexpected error: "+err.Error())
		return
	}

	result := modelFromTFEAdminSMTPSettings(smtpSettings, types.StringValue(""), false)
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

// updateSMTPSettings was created to keep the code DRY. It is used in both Create and Update functions
func (r *resourceTFEAdminSMTPSettings) updateAdminSMTPSettings(ctx context.Context, m modelTFEAdminSMTPSettings, config modelTFEAdminSMTPSettings) (*tfe.AdminSMTPSetting, error) {
	// Use password from config since write-only attributes aren't in the plan
	cur_pass := config.Password
	if !config.PasswordWO.IsNull() && !config.PasswordWO.IsUnknown() {
		cur_pass = config.PasswordWO
	}

	s, err := r.client.Admin.Settings.SMTP.Update(ctx, tfe.AdminSMTPSettingsUpdateOptions{
		Enabled:          m.Enabled.ValueBoolPointer(),
		Host:             m.Host.ValueStringPointer(),
		Port:             tfe.Int(int(m.Port.ValueInt64())),
		Sender:           m.Sender.ValueStringPointer(),
		Auth:             (*tfe.SMTPAuthType)(m.Auth.ValueStringPointer()),
		Username:         m.Username.ValueStringPointer(),
		Password:         cur_pass.ValueStringPointer(),
		TestEmailAddress: m.TestEmailAddress.ValueStringPointer(),
	})
	if err != nil {
		return s, fmt.Errorf("failed to update Admin SMTP Settings: %w", err)
	}
	return s, nil
}

// modelFromTFEAdminSMTPSettings builds a modelTFEAdminSMTPSettings struct from a tfe.AdminSMTPSetting value
func modelFromTFEAdminSMTPSettings(v *tfe.AdminSMTPSetting, password types.String, isWriteOnly bool) modelTFEAdminSMTPSettings {
	m := modelTFEAdminSMTPSettings{
		ID:       types.StringValue(v.ID),
		Enabled:  types.BoolValue(v.Enabled),
		Host:     types.StringValue(v.Host),
		Port:     types.Int64Value(int64(v.Port)),
		Sender:   types.StringValue(v.Sender),
		Auth:     types.StringValue(string(v.Auth)),
		Username: types.StringValue(v.Username),
		Password: types.StringValue(""),
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
