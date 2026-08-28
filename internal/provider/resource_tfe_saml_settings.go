// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"math"

	"github.com/hashicorp/go-tfe/v2/api/models"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
	samlDefaultAttrSiteAuditor                  string = "SiteAuditor"
	samlDefaultAttrGroups                       string = "MemberOf"
	samlDefaultSiteAdminRole                    string = "site-admins"
	samlDefaultSiteAuditorRole                  string = "site-auditors"
	samlDefaultSSOAPITokenSessionTimeoutSeconds int64  = 1209600 // 14 days

	samlProviderTypeOkta    string = "okta"
	samlProviderTypeEntra   string = "entra"
	samlProviderTypeGeneric string = "saml"
	samlProviderTypeUnknown string = "unknown"

	// samlSettingsID is the ID of the SAML settings singleton resource.
	samlSettingsID string = "saml"

	// minTFEVersionSiteAuditor is the first Terraform Enterprise release that
	// supports provisioning the Site Auditor role through SAML and SCIM. Shared
	// with the tfe_scim_settings resource.
	minTFEVersionSiteAuditor string = "2.1.0"
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
	AttrSiteAuditor           types.String `tfsdk:"attr_site_auditor"`
	SiteAuditorRole           types.String `tfsdk:"site_auditor_role"`
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

// samlSettingsEnvelope wraps a set of SAML settings attributes in the JSON:API
// document shape the admin SAML settings endpoint expects.
func samlSettingsEnvelope(attrs models.AdminSamlSettings_attributesable) models.AdminSamlSettingsEnvelopeable {
	data := models.NewAdminSamlSettings()
	data.SetId(ptr(samlSettingsID))
	data.SetTypeEscaped(ptr(models.SAMLSETTINGS_ADMINSAMLSETTINGS_TYPE))
	data.SetAttributes(attrs)

	envelope := models.NewAdminSamlSettingsEnvelope()
	envelope.SetData(data)
	return envelope
}

// int32SessionTimeout narrows a session timeout to the int32 the API expects.
// The schema validator already constrains the attribute to [0, MaxInt32], so
// out-of-range values are rejected at plan time; this clamp makes the narrowing
// safe by construction for any other caller and keeps a silent wrap-around
// impossible.
func int32SessionTimeout(v int64) int32 {
	if v > math.MaxInt32 {
		return math.MaxInt32
	}
	if v < 0 {
		return 0
	}
	return int32(v)
}

// stringOrPrior returns the server value when the attribute is present in the
// response, and the prior plan/state value otherwise. Terraform Enterprise
// releases older than minTFEVersionSiteAuditor omit the Site Auditor
// attributes entirely; falling back to the prior value keeps plan and state
// consistent instead of collapsing them to "". When there is no prior value —
// on import, or on the first refresh after upgrading a provider whose state
// predates these attributes — the schema default is used, since that is what
// the framework will put in the plan.
func stringOrPrior(v *string, prior types.String, def string) types.String {
	if v != nil {
		return types.StringValue(*v)
	}
	if prior.IsNull() || prior.IsUnknown() {
		return types.StringValue(def)
	}
	return prior
}

// modelFromV2SAMLSettings builds a modelTFESAMLSettings from an admin SAML
// settings response.
func modelFromV2SAMLSettings(env models.AdminSamlSettingsEnvelopeable, privateKey types.String, privateKeyWOVersion types.Int64, prior modelTFESAMLSettings) (modelTFESAMLSettings, error) {
	if env == nil || env.GetData() == nil || env.GetData().GetAttributes() == nil {
		return modelTFESAMLSettings{}, fmt.Errorf("SAML settings response did not contain any data")
	}
	data := env.GetData()
	attrs := data.GetAttributes()

	m := modelTFESAMLSettings{
		ID:                        types.StringValue(valueOrZero(data.GetId())),
		Enabled:                   types.BoolValue(valueOrZero(attrs.GetEnabled())),
		Debug:                     types.BoolValue(valueOrZero(attrs.GetDebug())),
		AuthnRequestsSigned:       types.BoolValue(valueOrZero(attrs.GetAuthnRequestsSigned())),
		WantAssertionsSigned:      types.BoolValue(valueOrZero(attrs.GetWantAssertionsSigned())),
		TeamManagementEnabled:     types.BoolValue(valueOrZero(attrs.GetTeamManagementEnabled())),
		OldIDPCert:                types.StringValue(valueOrZero(attrs.GetOldIdpCert())),
		IDPCert:                   types.StringValue(valueOrZero(attrs.GetIdpCert())),
		SLOEndpointURL:            types.StringValue(valueOrZero(attrs.GetSloEndpointUrl())),
		SSOEndpointURL:            types.StringValue(valueOrZero(attrs.GetSsoEndpointUrl())),
		AttrUsername:              types.StringValue(valueOrZero(attrs.GetAttrUsername())),
		AttrGroups:                types.StringValue(valueOrZero(attrs.GetAttrGroups())),
		AttrSiteAdmin:             types.StringValue(valueOrZero(attrs.GetAttrSiteAdmin())),
		SiteAdminRole:             types.StringValue(valueOrZero(attrs.GetSiteAdminRole())),
		AttrSiteAuditor:           stringOrPrior(attrs.GetAttrSiteAuditor(), prior.AttrSiteAuditor, samlDefaultAttrSiteAuditor),
		SiteAuditorRole:           stringOrPrior(attrs.GetSiteAuditorRole(), prior.SiteAuditorRole, samlDefaultSiteAuditorRole),
		SSOAPITokenSessionTimeout: types.Int64Value(int64(valueOrZero(attrs.GetSsoApiTokenSessionTimeout()))),
		ACSConsumerURL:            types.StringValue(valueOrZero(attrs.GetAcsConsumerUrl())),
		MetadataURL:               types.StringValue(valueOrZero(attrs.GetMetadataUrl())),
		Certificate:               types.StringValue(valueOrZero(attrs.GetCertificate())),
		PrivateKey:                types.StringValue(""),
		PrivateKeyWOVersion:       privateKeyWOVersion,
		SignatureSigningMethod:    types.StringValue(enumStringOrEmpty(attrs.GetSignatureSigningMethod())),
		SignatureDigestMethod:     types.StringValue(enumStringOrEmpty(attrs.GetSignatureDigestMethod())),
		ProviderType:              types.StringValue(enumStringOrEmpty(attrs.GetProviderType())),
	}

	// Note: compare against null/unknown explicitly. types.String.String() renders
	// null as "<null>", so a len()>0 check here would treat a null as a real value.
	if !privateKey.IsNull() && !privateKey.IsUnknown() {
		m.PrivateKey = privateKey
	}

	// Don't retrieve values if write-only is being used. Unset the private key field before updating the state.
	isWriteOnlyValue := !privateKeyWOVersion.IsNull()
	if isWriteOnlyValue {
		m.PrivateKey = types.StringValue("")
	}

	return m, nil
}

// supportsSiteAuditor reports whether the connected Terraform Enterprise
// supports the Site Auditor SAML attributes. When the practitioner has set
// either attribute in configuration against an older release, this records an
// error rather than letting the request silently drop them, which would surface
// as an opaque "Provider produced inconsistent result after apply".
func (r *resourceTFESAMLSettings) supportsSiteAuditor(config modelTFESAMLSettings, d *diag.Diagnostics) bool {
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

	// Only fail when the practitioner actually asked for Site Auditor. Schema
	// defaults alone must not break existing configurations on older releases.
	if !config.AttrSiteAuditor.IsNull() || !config.SiteAuditorRole.IsNull() {
		d.AddError(
			"Terraform Enterprise version does not support Site Auditor",
			fmt.Sprintf("The attributes attr_site_auditor and site_auditor_role require Terraform Enterprise %s or later. This instance reports %s. Remove those attributes from your configuration or upgrade Terraform Enterprise.",
				minTFEVersionSiteAuditor, r.config.RemoteTFEVersion()),
		)
	}
	return false
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
		return
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
			"attr_site_auditor": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf("Specifies the role for site auditor access. Overrides the \"Site Auditor Role\" method. Requires Terraform Enterprise %s or later.", minTFEVersionSiteAuditor),
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(samlDefaultAttrSiteAuditor),
			},
			"site_auditor_role": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf("Specifies the role for site auditor access, provided in the list of roles sent in the Team Attribute Name attribute. Requires Terraform Enterprise %s or later.", minTFEVersionSiteAuditor),
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(samlDefaultSiteAuditorRole),
			},
			"sso_api_token_session_timeout": schema.Int64Attribute{
				Description: "Specifies the Single Sign On session timeout in seconds. Defaults to 14 days.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(samlDefaultSSOAPITokenSessionTimeoutSeconds),
				// The API field is a 32-bit integer; reject values that would
				// silently wrap rather than sending a negative timeout.
				Validators: []validator.Int64{
					int64validator.Between(0, math.MaxInt32),
				},
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

	samlSettings, err := r.config.ClientV2.API.Admin().SamlSettings().Get(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error reading SAML Settings", "Could not read SAML Settings, unexpected error: "+apiErrorDetail(err))
		return
	}

	// update state
	result, err := modelFromV2SAMLSettings(samlSettings, m.PrivateKey, m.PrivateKeyWOVersion, m)
	if err != nil {
		resp.Diagnostics.AddError("Error reading SAML Settings", err.Error())
		return
	}
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

	withSiteAuditor := r.supportsSiteAuditor(config, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Create SAML Settings")
	samlSettings, err := r.updateSAMLSettings(ctx, m, withSiteAuditor)
	if err != nil {
		resp.Diagnostics.AddError("Error creating SAML Settings", "Could not set SAML Settings, unexpected error: "+apiErrorDetail(err))
		return
	}

	result, err := modelFromV2SAMLSettings(samlSettings, m.PrivateKey, config.PrivateKeyWOVersion, m)
	if err != nil {
		resp.Diagnostics.AddError("Error creating SAML Settings", err.Error())
		return
	}
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

	// m.PrivateKey does double duty: it is the request payload, where a null
	// means "omit private-key from the PATCH so the stored key is left alone",
	// and it would otherwise also become the state value. Those must not be the
	// same thing — writing null to state while the plan holds "" (the schema
	// default) or the configured key makes Terraform reject the apply with
	// "Provider produced inconsistent result after apply". Keep the planned
	// value for state and null out only the copy used to build the request.
	stateKey := m.PrivateKey
	if privateKey := r.determinePrivateKeyForUpdate(m, state, config); privateKey != nil {
		m.PrivateKey = types.StringValue(*privateKey)
		stateKey = m.PrivateKey
	} else {
		m.PrivateKey = types.StringNull()
	}

	withSiteAuditor := r.supportsSiteAuditor(config, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Update SAML Settings")
	samlSettings, err := r.updateSAMLSettings(ctx, m, withSiteAuditor)
	if err != nil {
		resp.Diagnostics.AddError("Error updating SAML Settings", "Could not set SAML Settings, unexpected error: "+apiErrorDetail(err))
		return
	}

	result, err := modelFromV2SAMLSettings(samlSettings, stateKey, config.PrivateKeyWOVersion, m)
	if err != nil {
		resp.Diagnostics.AddError("Error updating SAML Settings", err.Error())
		return
	}
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

	// Site Auditor attributes are only reset on releases that know about them.
	// An empty config is passed because destroying is never an explicit request
	// to manage Site Auditor, so an older release must not raise an error here.
	withSiteAuditor := r.supportsSiteAuditor(modelTFESAMLSettings{}, &resp.Diagnostics)
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
	attrs.SetSignatureSigningMethod(parseEnumPtr[models.AdminSamlSettings_attributes_signatureSigningMethod](models.ParseAdminSamlSettings_attributes_signatureSigningMethod, samlSignatureMethodSHA256))
	attrs.SetSignatureDigestMethod(parseEnumPtr[models.AdminSamlSettings_attributes_signatureDigestMethod](models.ParseAdminSamlSettings_attributes_signatureDigestMethod, samlSignatureMethodSHA256))
	attrs.SetProviderType(parseEnumPtr[models.AdminSamlSettings_attributes_providerType](models.ParseAdminSamlSettings_attributes_providerType, samlProviderTypeUnknown))
	if withSiteAuditor {
		attrs.SetAttrSiteAuditor(ptr(samlDefaultAttrSiteAuditor))
		attrs.SetSiteAuditorRole(ptr(samlDefaultSiteAuditorRole))
	}

	if _, err := r.config.ClientV2.API.Admin().SamlSettings().Patch(ctx, samlSettingsEnvelope(attrs), nil); err != nil {
		resp.Diagnostics.AddError("Error deleting SAML Settings", "Could not disable SAML Settings, unexpected error: "+apiErrorDetail(err))
		return
	}
}

// ImportState implements resource.ResourceWithImportState
func (r *resourceTFESAMLSettings) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	samlSettings, err := r.config.ClientV2.API.Admin().SamlSettings().Get(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error importing SAML Settings", "Could not retrieve SAML Settings, unexpected error: "+apiErrorDetail(err))
		return
	}

	// An imported resource has no prior state; stringOrPrior falls back to the
	// schema defaults for any attribute the release does not return.
	result, err := modelFromV2SAMLSettings(samlSettings, types.StringValue(""), types.Int64Null(), modelTFESAMLSettings{})
	if err != nil {
		resp.Diagnostics.AddError("Error importing SAML Settings", err.Error())
		return
	}
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

// updateSAMLSettings was created to keep the code DRY. It is used in both Create and Update functions.
// withSiteAuditor controls whether the Site Auditor attributes are sent: older
// Terraform Enterprise releases ignore unknown attributes, so sending them
// there would leave plan and state inconsistent.
func (r *resourceTFESAMLSettings) updateSAMLSettings(ctx context.Context, m modelTFESAMLSettings, withSiteAuditor bool) (models.AdminSamlSettingsEnvelopeable, error) {
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
	attrs.SetSsoApiTokenSessionTimeout(ptr(int32SessionTimeout(m.SSOAPITokenSessionTimeout.ValueInt64())))
	attrs.SetTeamManagementEnabled(m.TeamManagementEnabled.ValueBoolPointer())
	attrs.SetAuthnRequestsSigned(m.AuthnRequestsSigned.ValueBoolPointer())
	attrs.SetWantAssertionsSigned(m.WantAssertionsSigned.ValueBoolPointer())
	attrs.SetSignatureSigningMethod(parseEnumPtr[models.AdminSamlSettings_attributes_signatureSigningMethod](models.ParseAdminSamlSettings_attributes_signatureSigningMethod, m.SignatureSigningMethod.ValueString()))
	attrs.SetSignatureDigestMethod(parseEnumPtr[models.AdminSamlSettings_attributes_signatureDigestMethod](models.ParseAdminSamlSettings_attributes_signatureDigestMethod, m.SignatureDigestMethod.ValueString()))
	attrs.SetProviderType(parseEnumPtr[models.AdminSamlSettings_attributes_providerType](models.ParseAdminSamlSettings_attributes_providerType, m.ProviderType.ValueString()))
	if withSiteAuditor {
		attrs.SetAttrSiteAuditor(m.AttrSiteAuditor.ValueStringPointer())
		attrs.SetSiteAuditorRole(m.SiteAuditorRole.ValueStringPointer())
	}

	s, err := r.config.ClientV2.API.Admin().SamlSettings().Patch(ctx, samlSettingsEnvelope(attrs), nil)
	if err != nil {
		return s, fmt.Errorf("failed to update SAML Settings: %w", err)
	}
	return s, nil
}
