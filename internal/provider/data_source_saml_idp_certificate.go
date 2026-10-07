// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	tfe "github.com/hashicorp/go-tfe/v2"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-tfe/internal/provider/customtypes"
)

var (
	_ datasource.DataSource              = &dataSourceTFESAMLIDPCertificate{}
	_ datasource.DataSourceWithConfigure = &dataSourceTFESAMLIDPCertificate{}
)

// NewSAMLIDPCertificateDataSource returns the SAML IdP certificate data source.
func NewSAMLIDPCertificateDataSource() datasource.DataSource {
	return &dataSourceTFESAMLIDPCertificate{}
}

type dataSourceTFESAMLIDPCertificate struct {
	config ConfiguredClient
}

// samlIDPCertificateComputedAttributes returns the computed attributes shared
// by both SAML IdP certificate data sources. "id" is not included.
func samlIDPCertificateComputedAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"display_name": schema.StringAttribute{
			Description: "A human-readable label for the certificate.",
			Computed:    true,
		},
		"cert": schema.StringAttribute{
			Description: "The PEM encoded X.509 certificate as provided by the IdP.",
			Computed:    true,
			CustomType:  customtypes.PEMCertificateType{},
		},
		"fingerprint": schema.StringAttribute{
			Description: "Colon-separated uppercase SHA-256 fingerprint of the certificate.",
			Computed:    true,
		},
		"expires_at": schema.StringAttribute{
			Description: "The time the certificate expires, in RFC 3339 format.",
			Computed:    true,
		},
		"created_at": schema.StringAttribute{
			Description: "The time the certificate was added, in RFC 3339 format.",
			Computed:    true,
		},
		"issuer": schema.StringAttribute{
			Description: "The issuer of the certificate, derived from its O and CN fields.",
			Computed:    true,
		},
		"cert_role": schema.StringAttribute{
			MarkdownDescription: "The role of the certificate: `managed`, `legacy_primary` or `legacy_old`.",
			Computed:            true,
		},
	}
}

// checkSAMLIDPCertificateDataSourcesSupported records an error when the
// connected Terraform Enterprise predates the idp-certificates API. Shared by
// both SAML IdP certificate data sources.
func checkSAMLIDPCertificateDataSourcesSupported(c ConfiguredClient, d *diag.Diagnostics) {
	meets, err := c.MeetsMinRemoteTFEVersion(minTFEVersionSAMLIDPCertificates)
	if err != nil {
		d.AddError(
			"Error checking minimum Terraform Enterprise version",
			fmt.Sprintf("Could not determine whether Terraform Enterprise version %s meets the minimum required version %s: %v",
				c.RemoteTFEVersion(), minTFEVersionSAMLIDPCertificates, err),
		)
		return
	}
	if !meets {
		d.AddError(
			"Terraform Enterprise version does not support SAML IdP certificates",
			fmt.Sprintf("Reading SAML IdP certificates requires Terraform Enterprise %s or later. This instance reports %s. Upgrade Terraform Enterprise or use the idp_cert attribute of the tfe_saml_settings data source.",
				minTFEVersionSAMLIDPCertificates, c.RemoteTFEVersion()),
		)
	}
}

// Configure adds the provider configured client to the data source.
func (d *dataSourceTFESAMLIDPCertificate) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Metadata returns the data source type name.
func (d *dataSourceTFESAMLIDPCertificate) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_saml_idp_certificate"
}

// Schema defines the schema for the data source.
func (d *dataSourceTFESAMLIDPCertificate) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attrs := samlIDPCertificateComputedAttributes()
	attrs["id"] = schema.StringAttribute{
		Description: "The ID of the IdP certificate (idpc-*).",
		Required:    true,
		Validators: []validator.String{
			stringvalidator.RegexMatches(
				regexp.MustCompile(`^idpc-.+`),
				"must be a SAML IdP certificate ID of the form idpc-<id>",
			),
		},
	}
	resp.Schema = schema.Schema{
		Description: "(Only for Terraform Enterprise) Gets information on a trusted SAML Identity Provider certificate. Requires Terraform Enterprise v2.1.0 or later." +
			"\n\nThis requires admin token configuration. See example usage for incorporating an admin token in your provider config.",
		Attributes: attrs,
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *dataSourceTFESAMLIDPCertificate) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config modelTFESAMLIDPCertificate
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	checkSAMLIDPCertificateDataSourcesSupported(d.config, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	id := config.ID.ValueString()
	tflog.Debug(ctx, fmt.Sprintf("Read SAML IdP certificate %s", id))
	env, err := d.config.ClientV2.API.Admin().SamlSettings().IdpCertificates().ByExternal_id(id).Get(ctx, nil)
	if err != nil {
		if errors.Is(err, tfe.ErrNotFound) {
			resp.Diagnostics.AddError(
				"SAML IdP certificate not found",
				fmt.Sprintf("Could not find SAML IdP certificate %s.", id),
			)
			return
		}
		resp.Diagnostics.AddError(fmt.Sprintf("Error reading SAML IdP certificate %s", id), apiErrorDetail(err))
		return
	}

	result, err := modelFromSAMLIDPCertificateEnvelope(env)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Error reading SAML IdP certificate %s", id), err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}
