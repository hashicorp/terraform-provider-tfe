// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &dataSourceTFESAMLSettings{}
	_ datasource.DataSourceWithConfigure = &dataSourceTFESAMLSettings{}
)

// NewSAMLSettingsDataSource is a helper function to simplify the provider implementation.
func NewSAMLSettingsDataSource() datasource.DataSource {
	return &dataSourceTFESAMLSettings{}
}

// dataSourceTFESAMLSettings is the data source implementation.
type dataSourceTFESAMLSettings struct {
	config ConfiguredClient
}

// modelDataTFESAMLSettings maps the data source schema data.
type modelDataTFESAMLSettings struct {
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
	SignatureSigningMethod    types.String `tfsdk:"signature_signing_method"`
	SignatureDigestMethod     types.String `tfsdk:"signature_digest_method"`
	ProviderType              types.String `tfsdk:"provider_type"`
}

// Metadata returns the data source type name.
func (d *dataSourceTFESAMLSettings) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_saml_settings"
}

// Schema defines the schema for the data source.
func (d *dataSourceTFESAMLSettings) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "(Only for Terraform Enterprise) Gets information on SAML settings." +
			"\n\nThis requires admin token configuration. See example usage for incorporating an admin token in your provider config.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "ID of the SAML settings. It is always `saml`.",
				Computed:            true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether SAML single sign-on is enabled.",
				Computed:    true,
			},
			"debug": schema.BoolAttribute{
				Description: "Whether debug mode is enabled, which means that the SAMLResponse XML will be displayed on the login page.",
				Computed:    true,
			},
			"team_management_enabled": schema.BoolAttribute{
				Description: "Whether Terraform Enterprise is set to manage team membership.",
				Computed:    true,
			},
			"authn_requests_signed": schema.BoolAttribute{
				MarkdownDescription: "Whether `<samlp:AuthnRequest>` messages are signed.",
				Computed:            true,
			},
			"want_assertions_signed": schema.BoolAttribute{
				MarkdownDescription: "Whether `<saml:Assertion>` elements are signed.",
				Computed:            true,
			},
			"idp_cert": schema.StringAttribute{
				Description: "PEM encoded X.509 Certificate as provided by the IdP configuration.",
				Computed:    true,
			},
			"old_idp_cert": schema.StringAttribute{
				Description: "Previous version of the PEM encoded X.509 Certificate as provided by the IdP configuration.",
				Computed:    true,
			},
			"slo_endpoint_url": schema.StringAttribute{
				Description: "Single Log Out URL.",
				Computed:    true,
			},
			"sso_endpoint_url": schema.StringAttribute{
				Description: "Single Sign On URL.",
				Computed:    true,
			},
			"attr_username": schema.StringAttribute{
				Description: "Name of the SAML attribute that determines the user's username.",
				Computed:    true,
			},
			"attr_groups": schema.StringAttribute{
				Description: "Name of the SAML attribute that determines team membership.",
				Computed:    true,
			},
			"attr_site_admin": schema.StringAttribute{
				Description: "Site admin access role.",
				Computed:    true,
			},
			"site_admin_role": schema.StringAttribute{
				Description: "Site admin access role.",
				Computed:    true,
			},
			"attr_site_auditor": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf("Name of the SAML attribute that determines site auditor access. Empty on Terraform Enterprise releases older than %s.", minTFEVersionSiteAuditor),
				Computed:            true,
			},
			"site_auditor_role": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf("Site auditor access role. Empty on Terraform Enterprise releases older than %s.", minTFEVersionSiteAuditor),
				Computed:            true,
			},
			"sso_api_token_session_timeout": schema.Int64Attribute{
				Description: "Single Sign On session timeout in seconds.",
				Computed:    true,
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
				Description: "Request and assertion signing certificate.",
				Computed:    true,
			},
			"private_key": schema.StringAttribute{
				Description: "The private key used for request and assertion signing.",
				Computed:    true,
				Sensitive:   true,
			},
			"signature_signing_method": schema.StringAttribute{
				Description: "Signature Signing Method.",
				Computed:    true,
			},
			"signature_digest_method": schema.StringAttribute{
				Description: "Signature Digest Method.",
				Computed:    true,
			},
			"provider_type": schema.StringAttribute{
				MarkdownDescription: "The type of identity provider used. One of `okta`, `entra`, `saml`, or `unknown`.",
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *dataSourceTFESAMLSettings) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(ConfiguredClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected tfe.ConfiguredClient, got %T. This is a bug in the tfe provider, so please report it on GitHub.", req.ProviderData),
		)

		return
	}
	d.config = client
}

// Read refreshes the Terraform state with the latest data.
func (d *dataSourceTFESAMLSettings) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	env, err := d.config.ClientV2.API.Admin().SamlSettings().Get(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read SAML settings", apiErrorDetail(err))
		return
	}
	if env == nil || env.GetData() == nil || env.GetData().GetAttributes() == nil {
		resp.Diagnostics.AddError("Unable to read SAML settings", "SAML settings response did not contain any data")
		return
	}
	s := env.GetData().GetAttributes()

	// Set state
	diags := resp.State.Set(ctx, &modelDataTFESAMLSettings{
		ID:                        types.StringValue(valueOrZero(env.GetData().GetId())),
		Enabled:                   types.BoolValue(valueOrZero(s.GetEnabled())),
		Debug:                     types.BoolValue(valueOrZero(s.GetDebug())),
		TeamManagementEnabled:     types.BoolValue(valueOrZero(s.GetTeamManagementEnabled())),
		AuthnRequestsSigned:       types.BoolValue(valueOrZero(s.GetAuthnRequestsSigned())),
		WantAssertionsSigned:      types.BoolValue(valueOrZero(s.GetWantAssertionsSigned())),
		IDPCert:                   types.StringValue(valueOrZero(s.GetIdpCert())),
		OldIDPCert:                types.StringValue(valueOrZero(s.GetOldIdpCert())),
		SLOEndpointURL:            types.StringValue(valueOrZero(s.GetSloEndpointUrl())),
		SSOEndpointURL:            types.StringValue(valueOrZero(s.GetSsoEndpointUrl())),
		AttrUsername:              types.StringValue(valueOrZero(s.GetAttrUsername())),
		AttrGroups:                types.StringValue(valueOrZero(s.GetAttrGroups())),
		AttrSiteAdmin:             types.StringValue(valueOrZero(s.GetAttrSiteAdmin())),
		SiteAdminRole:             types.StringValue(valueOrZero(s.GetSiteAdminRole())),
		AttrSiteAuditor:           types.StringValue(valueOrZero(s.GetAttrSiteAuditor())),
		SiteAuditorRole:           types.StringValue(valueOrZero(s.GetSiteAuditorRole())),
		SSOAPITokenSessionTimeout: types.Int64Value(int64(valueOrZero(s.GetSsoApiTokenSessionTimeout()))),
		ACSConsumerURL:            types.StringValue(valueOrZero(s.GetAcsConsumerUrl())),
		MetadataURL:               types.StringValue(valueOrZero(s.GetMetadataUrl())),
		Certificate:               types.StringValue(valueOrZero(s.GetCertificate())),
		PrivateKey:                types.StringValue(valueOrZero(s.GetPrivateKey())),
		SignatureSigningMethod:    types.StringValue(enumStringOrEmpty(s.GetSignatureSigningMethod())),
		SignatureDigestMethod:     types.StringValue(enumStringOrEmpty(s.GetSignatureDigestMethod())),
		ProviderType:              types.StringValue(enumStringOrEmpty(s.GetProviderType())),
	})
	resp.Diagnostics.Append(diags...)
}
