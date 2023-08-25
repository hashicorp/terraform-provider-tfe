// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
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
)

// resourceTFESAMLSettings implements the tfe_saml_settings resource type
type resourceTFESAMLSettings struct {
	client *tfe.Client
}

// modelFromTFEAdminSAMLSettings builds a modelTFESAMLSettings struct from a tfe.AdminSAMLSetting value
func modelFromTFEAdminSAMLSettings(v tfe.AdminSAMLSetting, privateKey types.String) modelTFESAMLSettings {
	m := modelTFESAMLSettings{
		ID:                        types.StringValue(v.ID),
		Enabled:                   types.BoolValue(v.Enabled),
		Debug:                     types.BoolValue(v.Debug),
		AuthnRequestsSigned:       types.BoolValue(v.AuthnRequestsSigned),
		WantAssertionsSigned:      types.BoolValue(v.WantAssertionsSigned),
		TeamManagementEnabled:     types.BoolValue(v.TeamManagementEnabled),
		OldIDPCert:                types.StringValue(v.OldIDPCert),
		IDPCert:                   types.StringValue(v.IDPCert),
		SLOEndpointURL:            types.StringValue(v.SLOEndpointURL),
		SSOEndpointURL:            types.StringValue(v.SSOEndpointURL),
		AttrUsername:              types.StringValue(v.AttrUsername),
		AttrGroups:                types.StringValue(v.AttrGroups),
		AttrSiteAdmin:             types.StringValue(v.AttrSiteAdmin),
		SiteAdminRole:             types.StringValue(v.SiteAdminRole),
		SSOAPITokenSessionTimeout: types.Int64Value(int64(v.SSOAPITokenSessionTimeout)),
		ACSConsumerURL:            types.StringValue(v.ACSConsumerURL),
		MetadataURL:               types.StringValue(v.MetadataURL),
		Certificate:               types.StringValue(v.Certificate),
		PrivateKey:                types.StringValue(""),
		SignatureSigningMethod:    types.StringValue(v.SignatureSigningMethod),
		SignatureDigestMethod:     types.StringValue(v.SignatureDigestMethod),
	}
	if len(privateKey.String()) > 0 {
		m.PrivateKey = privateKey
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
	r.client = client.Client
}

// Metadata implements resource.Resource
func (r *resourceTFESAMLSettings) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_saml_settings"
}

// Schema implements resource.Resource
func (r *resourceTFESAMLSettings) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version: 1,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether or not SAML single sign-on is enabled",
				Computed:    true,
			},
			"debug": schema.BoolAttribute{
				Description: "When sign-on fails and this is enabled, the SAMLResponse XML will be displayed on the login page",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"authn_requests_signed": schema.BoolAttribute{
				Description: "Ensure that <samlp:AuthnRequest> messages are signed",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"want_assertions_signed": schema.BoolAttribute{
				Description: "Ensure that <saml:Assertion> elements are signed",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"team_management_enabled": schema.BoolAttribute{
				Description: "Set it to false if you would rather use Terraform Enterprise to manage team membership",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"old_idp_cert": schema.StringAttribute{
				Computed: true,
			},
			"idp_cert": schema.StringAttribute{
				Description: "Identity Provider Certificate specifies the PEM encoded X.509 Certificate as provided by the IdP configuration",
				Required:    true,
			},
			"slo_endpoint_url": schema.StringAttribute{
				Description: "Single Log Out URL specifies the HTTPS endpoint on your IdP for single logout requests. This value is provided by the IdP configuration",
				Required:    true,
			},
			"sso_endpoint_url": schema.StringAttribute{
				Description: "Single Sign On URL specifies the HTTPS endpoint on your IdP for single sign-on requests. This value is provided by the IdP configuration",
				Required:    true,
			},
			"attr_username": schema.StringAttribute{
				Description: "Username Attribute Name specifies the name of the SAML attribute that determines the user's username",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(samlDefaultAttrUsername),
			},
			"attr_site_admin": schema.StringAttribute{
				Description: "Specifies the role for site admin access. Overrides the \"Site Admin Role\" method",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(samlDefaultAttrSiteAdmin),
			},
			"attr_groups": schema.StringAttribute{
				Description: "Team Attribute Name specifies the name of the SAML attribute that determines team membership",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(samlDefaultAttrGroups),
			},
			"site_admin_role": schema.StringAttribute{
				Description: "Specifies the role for site admin access, provided in the list of roles sent in the Team Attribute Name attribute",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(samlDefaultSiteAdminRole),
			},
			"sso_api_token_session_timeout": schema.Int64Attribute{
				Description: "Specifies the Single Sign On session timeout in seconds. Defaults to 14 days",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(samlDefaultSSOAPITokenSessionTimeoutSeconds),
			},
			"acs_consumer_url": schema.StringAttribute{
				Description: "ACS Consumer (Recipient) URL",
				Computed:    true,
			},
			"metadata_url": schema.StringAttribute{
				Description: "Metadata (Audience) URL",
				Computed:    true,
			},
			"certificate": schema.StringAttribute{
				Description: "The certificate used for request and assertion signing",
				Optional:    true,
				Computed:    true,
			},
			"private_key": schema.StringAttribute{
				Description: "The private key used for request and assertion signing",
				Default:     stringdefault.StaticString(""),
				Optional:    true,
				Computed:    true,
				Sensitive:   true,
			},
			"signature_signing_method": schema.StringAttribute{
				Description: fmt.Sprintf("Signature Signing Method. Must be either `%s` or `%s`. Defaults to `%s`", samlSignatureMethodSHA1, samlSignatureMethodSHA256, samlSignatureMethodSHA256),
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(samlSignatureMethodSHA256),
				Validators: []validator.String{
					stringvalidator.OneOf(
						samlSignatureMethodSHA1,
						samlSignatureMethodSHA256,
					),
				},
			},
			"signature_digest_method": schema.StringAttribute{
				Description: fmt.Sprintf("Signature Digest Method. Must be either `%s` or `%s`. Defaults to `%s`", samlSignatureMethodSHA1, samlSignatureMethodSHA256, samlSignatureMethodSHA256),
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(samlSignatureMethodSHA256),
				Validators: []validator.String{
					stringvalidator.OneOf(
						samlSignatureMethodSHA1,
						samlSignatureMethodSHA256,
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

	samlSettings, err := r.client.Admin.Settings.SAML.Read(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading SAML Settings", "Could not read SAML Settings, unexpected error: "+err.Error())
		return
	}

	result := modelFromTFEAdminSAMLSettings(*samlSettings, m.PrivateKey)
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

	tflog.Debug(ctx, "Create SAML Settings")
	samlSettings, err := r.updateSAMLSettings(ctx, m)
	if err != nil {
		resp.Diagnostics.AddError("Error creating SAML Settings", "Could not set SAML Settings, unexpected error: "+err.Error())
		return
	}

	result := modelFromTFEAdminSAMLSettings(*samlSettings, m.PrivateKey)
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

	tflog.Debug(ctx, "Update SAML Settings")
	samlSettings, err := r.updateSAMLSettings(ctx, m)
	if err != nil {
		resp.Diagnostics.AddError("Error updating SAML Settings", "Could not set SAML Settings, unexpected error: "+err.Error())
		return
	}

	result := modelFromTFEAdminSAMLSettings(*samlSettings, m.PrivateKey)
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
	_, err := r.client.Admin.Settings.SAML.Update(ctx, tfe.AdminSAMLSettingsUpdateOptions{
		Enabled:                   basetypes.NewBoolValue(false).ValueBoolPointer(),
		Debug:                     basetypes.NewBoolValue(false).ValueBoolPointer(),
		AuthnRequestsSigned:       basetypes.NewBoolValue(false).ValueBoolPointer(),
		WantAssertionsSigned:      basetypes.NewBoolValue(false).ValueBoolPointer(),
		TeamManagementEnabled:     basetypes.NewBoolValue(false).ValueBoolPointer(),
		IDPCert:                   basetypes.NewStringValue("").ValueStringPointer(),
		SLOEndpointURL:            basetypes.NewStringValue("").ValueStringPointer(),
		SSOEndpointURL:            basetypes.NewStringValue("").ValueStringPointer(),
		AttrUsername:              basetypes.NewStringValue(samlDefaultAttrUsername).ValueStringPointer(),
		AttrSiteAdmin:             basetypes.NewStringValue(samlDefaultAttrSiteAdmin).ValueStringPointer(),
		AttrGroups:                basetypes.NewStringValue(samlDefaultAttrGroups).ValueStringPointer(),
		SiteAdminRole:             basetypes.NewStringValue(samlDefaultSiteAdminRole).ValueStringPointer(),
		SSOAPITokenSessionTimeout: tfe.Int(int(samlDefaultSSOAPITokenSessionTimeoutSeconds)),
		Certificate:               basetypes.NewStringValue("").ValueStringPointer(),
		PrivateKey:                basetypes.NewStringValue("").ValueStringPointer(),
		SignatureSigningMethod:    basetypes.NewStringValue(samlSignatureMethodSHA256).ValueStringPointer(),
		SignatureDigestMethod:     basetypes.NewStringValue(samlSignatureMethodSHA256).ValueStringPointer(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error deleting SAML Settings", "Could not disable SAML Settings, unexpected error: "+err.Error())
		return
	}
}

// ImportState implements resource.ResourceWithImportState
func (r *resourceTFESAMLSettings) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	samlSettings, err := r.client.Admin.Settings.SAML.Read(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error importing SAML Settings", "Could not retrieve SAML Settings, unexpected error: "+err.Error())
		return
	}

	result := modelFromTFEAdminSAMLSettings(*samlSettings, types.StringValue(""))
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

// updateSAMLSettings was created to keep the code DRY. It is used in both Create and Update functions
func (r *resourceTFESAMLSettings) updateSAMLSettings(ctx context.Context, m modelTFESAMLSettings) (*tfe.AdminSAMLSetting, error) {
	s, err := r.client.Admin.Settings.SAML.Update(ctx, tfe.AdminSAMLSettingsUpdateOptions{
		Enabled:                   basetypes.NewBoolValue(true).ValueBoolPointer(),
		Debug:                     m.Debug.ValueBoolPointer(),
		IDPCert:                   m.IDPCert.ValueStringPointer(),
		Certificate:               m.Certificate.ValueStringPointer(),
		PrivateKey:                m.PrivateKey.ValueStringPointer(),
		SLOEndpointURL:            m.SLOEndpointURL.ValueStringPointer(),
		SSOEndpointURL:            m.SSOEndpointURL.ValueStringPointer(),
		AttrUsername:              m.AttrUsername.ValueStringPointer(),
		AttrGroups:                m.AttrGroups.ValueStringPointer(),
		AttrSiteAdmin:             m.AttrSiteAdmin.ValueStringPointer(),
		SiteAdminRole:             m.SiteAdminRole.ValueStringPointer(),
		SSOAPITokenSessionTimeout: tfe.Int(int(m.SSOAPITokenSessionTimeout.ValueInt64())),
		TeamManagementEnabled:     m.TeamManagementEnabled.ValueBoolPointer(),
		AuthnRequestsSigned:       m.AuthnRequestsSigned.ValueBoolPointer(),
		WantAssertionsSigned:      m.WantAssertionsSigned.ValueBoolPointer(),
		SignatureSigningMethod:    m.SignatureSigningMethod.ValueStringPointer(),
		SignatureDigestMethod:     m.SignatureDigestMethod.ValueStringPointer(),
	})
	if err != nil {
		return s, fmt.Errorf("failed to update SAML Settings %v", err)
	}
	return s, err
}
