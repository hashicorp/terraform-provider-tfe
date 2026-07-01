// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-tfe"
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
	client *tfe.Client
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
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID of the SAML settings. It is always `saml`.",
				Computed:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whetther SAML single sign-on is enabled.",
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
				Description: "Whether `<samlp:AuthnRequest>` messages are signed.",
				Computed:    true,
			},
			"want_assertions_signed": schema.BoolAttribute{
				Description: "Whether `<saml:Assertion>` elements are signed.",
				Computed:    true,
			},
			"idp_cert": schema.StringAttribute{
				Description: "EM encoded X.509 Certificate as provided by the IdP configuration.",
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
				Description: "The type of identity provider used. One of `okta`, `entra`, `saml`, or `unknown`.",
				Computed:    true,
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
	d.client = client.Client
}

// Read refreshes the Terraform state with the latest data.
func (d *dataSourceTFESAMLSettings) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	s, err := d.client.Admin.Settings.SAML.Read(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read SAML settings", err.Error())
		return
	}

	// Set state
	diags := resp.State.Set(ctx, &modelDataTFESAMLSettings{
		ID:                        types.StringValue(s.ID),
		Enabled:                   types.BoolValue(s.Enabled),
		Debug:                     types.BoolValue(s.Debug),
		TeamManagementEnabled:     types.BoolValue(s.TeamManagementEnabled),
		AuthnRequestsSigned:       types.BoolValue(s.AuthnRequestsSigned),
		WantAssertionsSigned:      types.BoolValue(s.WantAssertionsSigned),
		IDPCert:                   types.StringValue(s.IDPCert),
		OldIDPCert:                types.StringValue(s.OldIDPCert),
		SLOEndpointURL:            types.StringValue(s.SLOEndpointURL),
		SSOEndpointURL:            types.StringValue(s.SSOEndpointURL),
		AttrUsername:              types.StringValue(s.AttrUsername),
		AttrGroups:                types.StringValue(s.AttrGroups),
		AttrSiteAdmin:             types.StringValue(s.AttrSiteAdmin),
		SiteAdminRole:             types.StringValue(s.SiteAdminRole),
		SSOAPITokenSessionTimeout: types.Int64Value(int64(s.SSOAPITokenSessionTimeout)),
		ACSConsumerURL:            types.StringValue(s.ACSConsumerURL),
		MetadataURL:               types.StringValue(s.MetadataURL),
		Certificate:               types.StringValue(s.Certificate),
		PrivateKey:                types.StringValue(s.PrivateKey),
		SignatureSigningMethod:    types.StringValue(s.SignatureSigningMethod),
		SignatureDigestMethod:     types.StringValue(s.SignatureDigestMethod),
		ProviderType:              types.StringValue(string(s.ProviderType)),
	})
	resp.Diagnostics.Append(diags...)
}
