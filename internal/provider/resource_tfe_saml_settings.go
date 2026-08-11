// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-tfe/v2/api/models"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	samlSignatureMethodSHA1                     string = "SHA1"
	samlSignatureMethodSHA256                   string = "SHA256"
	samlDefaultAttrUsername                     string = "Username"
	samlDefaultAttrSiteAdmin                    string = "SiteAdmin"
	samlDefaultAttrGroups                       string = "MemberOf"
	samlDefaultSiteAdminRole                    string = "site-admins"
	samlDefaultSSOAPITokenSessionTimeoutSeconds int64  = 1209600 // 14 days
	samlProviderTypeOkta                        string = "okta"
	samlProviderTypeEntra                       string = "entra"
	samlProviderTypeGeneric                     string = "saml"
	samlProviderTypeUnknown                     string = "unknown"
)

type modelTFESAMLSettings struct {
	ID                        types.String `tfsdk:"id"`
	Enabled                   types.Bool   `tfsdk:"enabled"`
	Debug                     types.Bool   `tfsdk:"debug"`
	TeamManagementEnabled     types.Bool   `tfsdk:"team_management_enabled"`
	AuthnRequestsSigned       types.Bool   `tfsdk:"authn_requests_signed"`
	WantAssertionsSigned      types.Bool   `tfsdk:"want_assertions_signed"`
	IDPCert                   types.String `tfsdk:"idp_cert"`
	OldIDPCert                types.String `tfsdk:"old_idp_cert"`
	SLOEndpointURL            types.String `tfsdk:"slo_endpoint_url"`
	SSOEndpointURL            types.String `tfsdk:"sso_endpoint_url"`
	AttrUsername              types.String `tfsdk:"attr_username"`
	AttrGroups                types.String `tfsdk:"attr_groups"`
	AttrSiteAdmin             types.String `tfsdk:"attr_site_admin"`
	SiteAdminRole             types.String `tfsdk:"site_admin_role"`
	SSOAPITokenSessionTimeout types.Int64  `tfsdk:"sso_api_token_session_timeout"`
	ACSConsumerURL            types.String `tfsdk:"acs_consumer_url"`
	MetadataURL               types.String `tfsdk:"metadata_url"`
	Certificate               types.String `tfsdk:"certificate"`
	PrivateKey                types.String `tfsdk:"private_key"`
	PrivateKeyWO              types.String `tfsdk:"private_key_wo"`
	PrivateKeyWOVersion       types.Int64  `tfsdk:"private_key_wo_version"`
	SignatureSigningMethod    types.String `tfsdk:"signature_signing_method"`
	SignatureDigestMethod     types.String `tfsdk:"signature_digest_method"`
	ProviderType              types.String `tfsdk:"provider_type"`
}

// resourceTFESAMLSettings implements the tfe_saml_settings resource type
type resourceTFESAMLSettings struct {
	config ConfiguredClient
}

// modelFromTFEAdminSAMLSettingsV2 builds a modelTFESAMLSettings struct from a v2 AdminSamlSettingsable value
func modelFromTFEAdminSAMLSettingsV2(v models.AdminSamlSettingsable, privateKey types.String, privateKeyWOVersion types.Int64) modelTFESAMLSettings {
	m := modelTFESAMLSettings{
		PrivateKey:          types.StringValue(""),
		PrivateKeyWOVersion: privateKeyWOVersion,
	}

	if id := v.GetId(); id != nil {
		m.ID = types.StringValue(id.String())
	}

	if attrs := v.GetAttributes(); attrs != nil {
		m.Enabled = types.BoolValue(valueOrZero(attrs.GetEnabled()))
		m.Debug = types.BoolValue(valueOrZero(attrs.GetDebug()))
		m.AuthnRequestsSigned = types.BoolValue(valueOrZero(attrs.GetAuthnRequestsSigned()))
		m.WantAssertionsSigned = types.BoolValue(valueOrZero(attrs.GetWantAssertionsSigned()))
		m.TeamManagementEnabled = types.BoolValue(valueOrZero(attrs.GetTeamManagementEnabled()))
		m.OldIDPCert = types.StringValue(valueOrZero(attrs.GetOldIdpCert()))
		m.IDPCert = types.StringValue(valueOrZero(attrs.GetIdpCert()))
		m.SLOEndpointURL = types.StringValue(valueOrZero(attrs.GetSloEndpointUrl()))
		m.SSOEndpointURL = types.StringValue(valueOrZero(attrs.GetSsoEndpointUrl()))
		m.AttrUsername = types.StringValue(valueOrZero(attrs.GetAttrUsername()))
		m.AttrGroups = types.StringValue(valueOrZero(attrs.GetAttrGroups()))
		m.AttrSiteAdmin = types.StringValue(valueOrZero(attrs.GetAttrSiteAdmin()))
		m.SiteAdminRole = types.StringValue(valueOrZero(attrs.GetSiteAdminRole()))
		m.SSOAPITokenSessionTimeout = types.Int64Value(int64(valueOrZero(attrs.GetSsoApiTokenSessionTimeout())))
		m.ACSConsumerURL = types.StringValue(valueOrZero(attrs.GetAcsConsumerUrl()))
		m.MetadataURL = types.StringValue(valueOrZero(attrs.GetMetadataUrl()))
		m.Certificate = types.StringValue(valueOrZero(attrs.GetCertificate()))
		m.SignatureSigningMethod = types.StringValue(enumStringOrEmpty(attrs.GetSignatureSigningMethod()))
		m.SignatureDigestMethod = types.StringValue(enumStringOrEmpty(attrs.GetSignatureDigestMethod()))
		m.ProviderType = types.StringValue(enumStringOrEmpty(attrs.GetProviderType()))
	}

	if len(privateKey.String()) > 0 {
		m.PrivateKey = privateKey
	}

	// Don't retrieve values if write-only is being used. Unset the private key field before updating the state.
	isWriteOnlyValue := !privateKeyWOVersion.IsNull()
	if isWriteOnlyValue {
		m.PrivateKey = types.StringValue("")
	}

	return m
}

// Configure implements resource.ResourceWithConfigure
func (r *resourceTFESAMLSettings) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *resourceTFESAMLSettings) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_saml_settings"
}

// Schema implements resource.Resource
func (r *resourceTFESAMLSettings) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "(Only for Terraform Enterprise) Creates, updates, and destroys SAML settings." +
			"\n\nRequires admin token configuration. See example usage for incorporating an admin token in your provider config.",
		Version: 1,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the SAML settings. Always `saml`.",
				Computed:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether or not SAML single sign-on is enabled.",
				Computed:    true,
			},
			"debug": schema.BoolAttribute{
				Description: "When sign-on fails and this is enabled, the SAMLResponse XML will be displayed on the login page.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"authn_requests_signed": schema.BoolAttribute{
				MarkdownDescription: "Ensure that `<samlp:AuthnRequest>` messages are signed.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"want_assertions_signed": schema.BoolAttribute{
				MarkdownDescription: "Ensure that `<saml:Assertion>` elements are signed.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"team_management_enabled": schema.BoolAttribute{
				Description: "Whether Terraform Enterprise manages team membership via SAML. Set to false if you would rather use Terraform Enterprise to manage team membership.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"old_idp_cert": schema.StringAttribute{
				Description: "The previous identity provider certificate, kept after the IdP certificate is updated.",
				Computed:    true,
			},
			"idp_cert": schema.StringAttribute{
				Description: "Identity Provider Certificate specifies the PEM encoded X.509 Certificate as provided by the IdP configuration.",
				Required:    true,
			},
			"slo_endpoint_url": schema.StringAttribute{
				Description: "Single Log Out URL specifies the HTTPS endpoint on your IdP for single logout requests. This value is provided by the IdP configuration.",
				Required:    true,
			},
			"sso_endpoint_url": schema.StringAttribute{
				Description: "Single Sign On URL specifies the HTTPS endpoint on your IdP for single sign-on requests. This value is provided by the IdP configuration.",
				Required:    true,
			},
			"attr_username": schema.StringAttribute{
				Description: "Username Attribute Name specifies the name of the SAML attribute that determines the user's username.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(samlDefaultAttrUsername),
			},
			"attr_site_admin": schema.StringAttribute{
				Description: "Specifies the role for site admin access. Overrides the \"Site Admin Role\" method.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(samlDefaultAttrSiteAdmin),
			},
			"attr_groups": schema.StringAttribute{
				Description: "Team Attribute Name specifies the name of the SAML attribute that determines team membership.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(samlDefaultAttrGroups),
			},
			"site_admin_role": schema.StringAttribute{
				Description: "Specifies the role for site admin access, provided in the list of roles sent in the Team Attribute Name attribute.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(samlDefaultSiteAdminRole),
			},
			"sso_api_token_session_timeout": schema.Int64Attribute{
				Description: "Specifies the Single Sign On session timeout in seconds. Defaults to 14 days.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(samlDefaultSSOAPITokenSessionTimeoutSeconds),
			},
			"acs_consumer_url": schema.StringAttribute{
				Description: "ACS Consumer (Recipient) URL.",
				Computed:    true,
			},
			"metadata_url": schema.StringAttribute{
				Description: "Metadata (Audience) URL.",
				Computed:    true,
			},
			"certificate": schema.StringAttribute{
				Description: "The certificate used for request and assertion signing.",
				Optional:    true,
				Computed:    true,
			},
			"private_key": schema.StringAttribute{
				Description: "The private key used for request and assertion signing.",
				Default:     stringdefault.StaticString(""),
				Optional:    true,
				Computed:    true,
				Sensitive:   true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("private_key_wo")),
				},
			},
			// since the private_key_wo write-only values are not saved to state, they will not trigger updates on their own.
			// Instead the private_key_wo_version responsibility is to trigger updates to the private_key_wo attribute when version number changes.
			"private_key_wo": schema.StringAttribute{
				Description: "The private key in write-only mode used for request and assertion signing. Guaranteed not to be written to plan or state artifacts. Either `private_key` or `private_key_wo` can be provided, but not both. Must be used with `private_key_wo_version`.",
				Optional:    true,
				Sensitive:   true,
				WriteOnly:   true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("private_key")),
					stringvalidator.AlsoRequires(path.MatchRoot("private_key_wo_version")),
				},
			},

			"private_key_wo_version": schema.Int64Attribute{
				Optional:    true,
				Description: "Version of the write-only private key. This field is used to trigger updates when the write-only private key changes. Must be used with `private_key_wo`. When `private_key_wo_version` changes, the write-only private key will be updated.",
				Validators: []validator.Int64{
					int64validator.ConflictsWith(path.MatchRoot("private_key")),
					int64validator.AlsoRequires(path.MatchRoot("private_key_wo")),
				},
			},
			"signature_signing_method": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf("Signature Signing Method. Must be either `%s` or `%s`. Defaults to `%s`.", samlSignatureMethodSHA1, samlSignatureMethodSHA256, samlSignatureMethodSHA256),
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(samlSignatureMethodSHA256),
				Validators: []validator.String{
					stringvalidator.OneOf(
						samlSignatureMethodSHA1,
						samlSignatureMethodSHA256,
					),
				},
			},
			"signature_digest_method": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf("Signature Digest Method. Must be either `%s` or `%s`. Defaults to `%s`.", samlSignatureMethodSHA1, samlSignatureMethodSHA256, samlSignatureMethodSHA256),
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(samlSignatureMethodSHA256),
				Validators: []validator.String{
					stringvalidator.OneOf(
						samlSignatureMethodSHA1,
						samlSignatureMethodSHA256,
					),
				},
			},
			"provider_type": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf("The type of identity provider used. Valid values are `%s`, `%s`, `%s`, and `%s`. Defaults to `%s`.", samlProviderTypeOkta, samlProviderTypeEntra, samlProviderTypeGeneric, samlProviderTypeUnknown, samlProviderTypeUnknown),
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(samlProviderTypeUnknown),
				Validators: []validator.String{
					stringvalidator.OneOf(
						samlProviderTypeOkta,
						samlProviderTypeEntra,
						// `samlProviderTypeGeneric` is the string literal "saml", and is shown as `SAML` in the TFE UI.
						samlProviderTypeGeneric,
						samlProviderTypeUnknown,
					),
				},
			},
		},
	}
}

// Read implements resource.Resource
func (r *resourceTFESAMLSettings) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var m modelTFESAMLSettings
	diags := req.State.Get(ctx, &m)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading SAML Settings")

	env, err := r.config.ClientV2.API.Admin().SamlSettings().Get(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error reading SAML Settings", "Could not read SAML Settings, unexpected error: "+err.Error())
		return
	}

	// update state
	result := modelFromTFEAdminSAMLSettingsV2(env.GetData(), m.PrivateKey, m.PrivateKeyWOVersion)
	diags = resp.State.Set(ctx, &result)
	resp.Diagnostics.Append(diags...)
}

// Create implements resource.Resource
func (r *resourceTFESAMLSettings) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var m modelTFESAMLSettings
	diags := req.Plan.Get(ctx, &m)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var config modelTFESAMLSettings
	diags = req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !config.PrivateKeyWO.IsNull() {
		m.PrivateKey = config.PrivateKeyWO
	}

	tflog.Debug(ctx, "Create SAML Settings")
	data, err := r.updateSAMLSettings(ctx, m)
	if err != nil {
		resp.Diagnostics.AddError("Error creating SAML Settings", "Could not set SAML Settings, unexpected error: "+err.Error())
		return
	}

	result := modelFromTFEAdminSAMLSettingsV2(data, m.PrivateKey, config.PrivateKeyWOVersion)
	diags = resp.State.Set(ctx, &result)
	resp.Diagnostics.Append(diags...)
}

// Update implements resource.Resource
func (r *resourceTFESAMLSettings) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var m modelTFESAMLSettings
	diags := req.Plan.Get(ctx, &m)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var config modelTFESAMLSettings
	diags = req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state modelTFESAMLSettings
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if privateKey := r.determinePrivateKeyForUpdate(m, state, config); privateKey != nil {
		m.PrivateKey = types.StringValue(*privateKey)
	} else {
		m.PrivateKey = types.StringNull()
	}

	tflog.Debug(ctx, "Update SAML Settings")
	data, err := r.updateSAMLSettings(ctx, m)
	if err != nil {
		resp.Diagnostics.AddError("Error updating SAML Settings", "Could not set SAML Settings, unexpected error: "+err.Error())
		return
	}

	result := modelFromTFEAdminSAMLSettingsV2(data, m.PrivateKey, config.PrivateKeyWOVersion)
	// Save data into Terraform state
	diags = resp.State.Set(ctx, &result)
	resp.Diagnostics.Append(diags...)
}

// Delete disables the SAML Settings and then removes the resource from the state file. You cannot delete TFE SAML Settings, only disable them
func (r *resourceTFESAMLSettings) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var m modelTFESAMLSettings
	diags := req.State.Get(ctx, &m)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Delete SAML Settings")

	attrs := models.NewAdminSamlSettings_attributes()
	attrs.SetEnabled(ptr(false))
	attrs.SetDebug(ptr(false))
	attrs.SetAuthnRequestsSigned(ptr(false))
	attrs.SetWantAssertionsSigned(ptr(false))
	attrs.SetTeamManagementEnabled(ptr(false))
	attrs.SetIdpCert(ptr(""))
	attrs.SetSloEndpointUrl(ptr(""))
	attrs.SetSsoEndpointUrl(ptr(""))
	attrs.SetAttrUsername(ptr(samlDefaultAttrUsername))
	attrs.SetAttrSiteAdmin(ptr(samlDefaultAttrSiteAdmin))
	attrs.SetAttrGroups(ptr(samlDefaultAttrGroups))
	attrs.SetSiteAdminRole(ptr(samlDefaultSiteAdminRole))
	attrs.SetSsoApiTokenSessionTimeout(ptr(int32(samlDefaultSSOAPITokenSessionTimeoutSeconds)))
	attrs.SetCertificate(ptr(""))
	attrs.SetPrivateKey(ptr(""))
	if signingMethod, err := parseSAMLSignatureSigningMethod(samlSignatureMethodSHA256); err == nil {
		attrs.SetSignatureSigningMethod(signingMethod)
	}
	if digestMethod, err := parseSAMLSignatureDigestMethod(samlSignatureMethodSHA256); err == nil {
		attrs.SetSignatureDigestMethod(digestMethod)
	}
	if providerType, err := parseSAMLProviderType(samlProviderTypeUnknown); err == nil {
		attrs.SetProviderType(providerType)
	}

	_, err := r.config.ClientV2.API.Admin().SamlSettings().Patch(ctx, buildSAMLSettingsEnvelope(attrs), nil)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting SAML Settings", "Could not disable SAML Settings, unexpected error: "+err.Error())
		return
	}
}

// ImportState implements resource.ResourceWithImportState
func (r *resourceTFESAMLSettings) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	env, err := r.config.ClientV2.API.Admin().SamlSettings().Get(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error importing SAML Settings", "Could not retrieve SAML Settings, unexpected error: "+err.Error())
		return
	}

	result := modelFromTFEAdminSAMLSettingsV2(env.GetData(), types.StringValue(""), types.Int64Null())
	diags := resp.State.Set(ctx, &result)
	resp.Diagnostics.Append(diags...)
}

var (
	_ resource.Resource                = &resourceTFESAMLSettings{}
	_ resource.ResourceWithConfigure   = &resourceTFESAMLSettings{}
	_ resource.ResourceWithImportState = &resourceTFESAMLSettings{}
)

// NewSAMLSettingsResource is a resource function for the framework provider.
func NewSAMLSettingsResource() resource.Resource {
	return &resourceTFESAMLSettings{}
}

// determinePrivateKeyForUpdate is invoked only after terraform determines that an attribute update is needed.
// note that the update can be triggered by other attributes outside of the private_key/private_key_wo attributes.
// this function compares the PrivateKeyWOVersion vs PrivateKey to ensure that during api update call, private_key is not mistakenly unset.
// Returns nil if no key update is needed.
func (r *resourceTFESAMLSettings) determinePrivateKeyForUpdate(plan, state, config modelTFESAMLSettings) *string {
	// Determine if we're using write-only private key in plan vs state
	usingWriteOnlyInPlan := !plan.PrivateKeyWOVersion.IsNull()
	usingWriteOnlyInState := !state.PrivateKeyWOVersion.IsNull()

	// Case 1: Switching FROM private_key TO private_key_wo
	if !usingWriteOnlyInState && usingWriteOnlyInPlan && !config.PrivateKeyWO.IsNull() {
		return config.PrivateKeyWO.ValueStringPointer()
	}
	// Case 2: Switching FROM private_key_wo TO private_key
	if usingWriteOnlyInState && !usingWriteOnlyInPlan && !plan.PrivateKey.IsNull() {
		return plan.PrivateKey.ValueStringPointer()
	}
	// Case 3: private_key_wo version changed in plan
	if usingWriteOnlyInPlan && plan.PrivateKeyWOVersion.ValueInt64() != state.PrivateKeyWOVersion.ValueInt64() && !config.PrivateKeyWO.IsNull() {
		return config.PrivateKeyWO.ValueStringPointer()
	}
	// Case 4: Regular private_key changed. Only set PrivateKey if our planned value would be a CHANGE from
	// the prior state. This prevents accidentally resetting the private key on unrelated changes.
	if state.PrivateKey.ValueString() != plan.PrivateKey.ValueString() {
		return plan.PrivateKey.ValueStringPointer()
	}
	return nil
}

// parseSAMLSignatureSigningMethod converts a signature method string to the v2 generated enum.
func parseSAMLSignatureSigningMethod(v string) (*models.AdminSamlSettings_attributes_signatureSigningMethod, error) {
	parsed, err := models.ParseAdminSamlSettings_attributes_signatureSigningMethod(v)
	if err != nil || parsed == nil {
		return nil, fmt.Errorf("invalid signature_signing_method %q: %w", v, err)
	}
	method := parsed.(*models.AdminSamlSettings_attributes_signatureSigningMethod)
	return method, nil
}

// parseSAMLSignatureDigestMethod converts a signature method string to the v2 generated enum.
func parseSAMLSignatureDigestMethod(v string) (*models.AdminSamlSettings_attributes_signatureDigestMethod, error) {
	parsed, err := models.ParseAdminSamlSettings_attributes_signatureDigestMethod(v)
	if err != nil || parsed == nil {
		return nil, fmt.Errorf("invalid signature_digest_method %q: %w", v, err)
	}
	method := parsed.(*models.AdminSamlSettings_attributes_signatureDigestMethod)
	return method, nil
}

// parseSAMLProviderType converts a provider type string to the v2 generated enum.
func parseSAMLProviderType(v string) (*models.AdminSamlSettings_attributes_providerType, error) {
	parsed, err := models.ParseAdminSamlSettings_attributes_providerType(v)
	if err != nil || parsed == nil {
		return nil, fmt.Errorf("invalid provider_type %q: %w", v, err)
	}
	providerType := parsed.(*models.AdminSamlSettings_attributes_providerType)
	return providerType, nil
}

// buildSAMLSettingsEnvelope wraps attrs in an AdminSamlSettingsEnvelope ready
// to send to the admin SAML settings PATCH endpoint.
func buildSAMLSettingsEnvelope(attrs models.AdminSamlSettings_attributesable) *models.AdminSamlSettingsEnvelope {
	settingsType := models.SAMLSETTINGS_ADMINSAMLSETTINGS_TYPE
	settingsID := models.SAML_ADMINSAMLSETTINGS_ID
	data := models.NewAdminSamlSettings()
	data.SetTypeEscaped(&settingsType)
	data.SetId(&settingsID)
	data.SetAttributes(attrs)

	envelope := models.NewAdminSamlSettingsEnvelope()
	envelope.SetData(data)
	return envelope
}

// updateSAMLSettings was created to keep the code DRY. It is used in both Create and Update functions
func (r *resourceTFESAMLSettings) updateSAMLSettings(ctx context.Context, m modelTFESAMLSettings) (models.AdminSamlSettingsable, error) {
	attrs := models.NewAdminSamlSettings_attributes()
	attrs.SetEnabled(ptr(true))
	attrs.SetDebug(m.Debug.ValueBoolPointer())
	attrs.SetIdpCert(m.IDPCert.ValueStringPointer())
	attrs.SetCertificate(m.Certificate.ValueStringPointer())
	attrs.SetPrivateKey(m.PrivateKey.ValueStringPointer())
	attrs.SetSloEndpointUrl(m.SLOEndpointURL.ValueStringPointer())
	attrs.SetSsoEndpointUrl(m.SSOEndpointURL.ValueStringPointer())
	attrs.SetAttrUsername(m.AttrUsername.ValueStringPointer())
	attrs.SetAttrGroups(m.AttrGroups.ValueStringPointer())
	attrs.SetAttrSiteAdmin(m.AttrSiteAdmin.ValueStringPointer())
	attrs.SetSiteAdminRole(m.SiteAdminRole.ValueStringPointer())
	attrs.SetSsoApiTokenSessionTimeout(ptr(int32(m.SSOAPITokenSessionTimeout.ValueInt64()))) //nolint:gosec // sso_api_token_session_timeout is a small user-supplied duration, never near int32 overflow
	attrs.SetTeamManagementEnabled(m.TeamManagementEnabled.ValueBoolPointer())
	attrs.SetAuthnRequestsSigned(m.AuthnRequestsSigned.ValueBoolPointer())
	attrs.SetWantAssertionsSigned(m.WantAssertionsSigned.ValueBoolPointer())

	signingMethod, err := parseSAMLSignatureSigningMethod(m.SignatureSigningMethod.ValueString())
	if err != nil {
		return nil, err
	}
	attrs.SetSignatureSigningMethod(signingMethod)

	digestMethod, err := parseSAMLSignatureDigestMethod(m.SignatureDigestMethod.ValueString())
	if err != nil {
		return nil, err
	}
	attrs.SetSignatureDigestMethod(digestMethod)

	providerType, err := parseSAMLProviderType(m.ProviderType.ValueString())
	if err != nil {
		return nil, err
	}
	attrs.SetProviderType(providerType)

	env, err := r.config.ClientV2.API.Admin().SamlSettings().Patch(ctx, buildSAMLSettingsEnvelope(attrs), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to update SAML Settings: %w", err)
	}
	if env == nil || env.GetData() == nil {
		return nil, fmt.Errorf("failed to update SAML Settings: API returned no data")
	}
	return env.GetData(), nil
}
