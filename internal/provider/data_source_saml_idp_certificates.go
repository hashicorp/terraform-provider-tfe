// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/go-tfe/v2/api/models"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource              = &dataSourceTFESAMLIDPCertificates{}
	_ datasource.DataSourceWithConfigure = &dataSourceTFESAMLIDPCertificates{}
)

// samlIDPCertificatesID is the static ID of the SAML IdP certificates data
// source, matching the ID of the SAML settings singleton.
const samlIDPCertificatesID = "saml"

// NewSAMLIDPCertificatesDataSource returns the SAML IdP certificates data source.
func NewSAMLIDPCertificatesDataSource() datasource.DataSource {
	return &dataSourceTFESAMLIDPCertificates{}
}

type dataSourceTFESAMLIDPCertificates struct {
	config ConfiguredClient
}

type modelDataTFESAMLIDPCertificates struct {
	ID               types.String                 `tfsdk:"id"`
	DisplayNameMatch types.String                 `tfsdk:"display_name_match"`
	Certificates     []modelTFESAMLIDPCertificate `tfsdk:"certificates"`
}

// Configure adds the provider configured client to the data source.
func (d *dataSourceTFESAMLIDPCertificates) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *dataSourceTFESAMLIDPCertificates) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_saml_idp_certificates"
}

// Schema defines the schema for the data source.
func (d *dataSourceTFESAMLIDPCertificates) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	certAttrs := samlIDPCertificateComputedAttributes()
	certAttrs["id"] = schema.StringAttribute{
		Description: "The ID of the IdP certificate (idpc-*).",
		Computed:    true,
	}

	resp.Schema = schema.Schema{
		Description: "(Only for Terraform Enterprise) Gets the list of trusted SAML Identity Provider certificates. Requires Terraform Enterprise v2.1.0 or later." +
			"\n\nThis requires admin token configuration. See example usage for incorporating an admin token in your provider config.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "ID of the data source. It is always `saml`.",
				Computed:            true,
			},
			"display_name_match": schema.StringAttribute{
				Description: "A case-insensitive substring to match against certificate display names. When omitted, all certificates are returned.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"certificates": schema.ListNestedAttribute{
				MarkdownDescription: "The trusted SAML IdP certificates, including legacy certificates, in the order returned by Terraform Enterprise: " +
					"the `legacy_primary` certificate first, then `legacy_old`, then `managed` certificates by least recently updated. " +
					"Updating a certificate, including renaming it, moves it to the end of the list, so prefer filtering by `id` or `display_name` over indexing.",
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: certAttrs,
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *dataSourceTFESAMLIDPCertificates) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config modelDataTFESAMLIDPCertificates
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	checkSAMLIDPCertificateDataSourcesSupported(d.config, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// The list endpoint isn't paginated, so one call returns every
	// certificate and we filter on our side.
	tflog.Debug(ctx, "List SAML IdP certificates")
	list, err := d.config.ClientV2.API.Admin().SamlSettings().IdpCertificates().Get(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error listing SAML IdP certificates", apiErrorDetail(err))
		return
	}

	// A nil response means the instance has no certificates; return an
	// empty list rather than an error.
	var data []models.SamlIdpCertificatesable
	if list != nil {
		data = list.GetData()
	}

	all := make([]modelTFESAMLIDPCertificate, 0, len(data))
	for i, item := range data {
		cert, err := modelFromSAMLIDPCertificate(item)
		if err != nil {
			resp.Diagnostics.AddError("Error listing SAML IdP certificates", fmt.Sprintf("Certificate at index %d: %s", i, err))
			return
		}
		all = append(all, cert)
	}

	config.ID = types.StringValue(samlIDPCertificatesID)
	config.Certificates = filterSAMLIDPCertificatesByDisplayName(all, config.DisplayNameMatch.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

// filterSAMLIDPCertificatesByDisplayName returns the certificates whose display
// name contains match, case-insensitively, preserving their order. An empty
// match returns certs unchanged.
func filterSAMLIDPCertificatesByDisplayName(certs []modelTFESAMLIDPCertificate, match string) []modelTFESAMLIDPCertificate {
	if match == "" {
		return certs
	}
	match = strings.ToLower(match)
	out := make([]modelTFESAMLIDPCertificate, 0, len(certs))
	for i := range certs {
		if strings.Contains(strings.ToLower(certs[i].DisplayName.ValueString()), match) {
			out = append(out, certs[i])
		}
	}
	return out
}
