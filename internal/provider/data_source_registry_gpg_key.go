// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &dataSourceTFERegistryGPGKey{}
	_ datasource.DataSourceWithConfigure = &dataSourceTFERegistryGPGKey{}
)

// NewRegistryGPGKeyDataSource is a helper function to simplify the provider implementation.
func NewRegistryGPGKeyDataSource() datasource.DataSource {
	return &dataSourceTFERegistryGPGKey{}
}

// dataSourceTFERegistryGPGKey is the data source implementation.
type dataSourceTFERegistryGPGKey struct {
	config ConfiguredClient
}

// Metadata returns the data source type name.
func (d *dataSourceTFERegistryGPGKey) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_registry_gpg_key"
}

// Schema defines the schema for the data source.
func (d *dataSourceTFERegistryGPGKey) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This data source can be used to retrieve a private registry GPG key.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required: true,
			},
			"organization": schema.StringAttribute{
				Description: "Name of the organization. If omitted, organization must be defined in the provider config.",
				Optional:    true,
				Computed:    true,
			},
			"ascii_armor": schema.StringAttribute{
				Description: "ASCII-armored representation of the GPG key.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "The time when the GPG key was created.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "The time when the GPG key was last updated.",
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *dataSourceTFERegistryGPGKey) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *dataSourceTFERegistryGPGKey) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data modelTFERegistryGPGKey

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var organization string
	resp.Diagnostics.Append(d.config.dataOrDefaultOrganization(ctx, req.Config, &organization)...)

	if resp.Diagnostics.HasError() {
		return
	}

	keyID := tfe.GPGKeyID{
		RegistryName: "private",
		Namespace:    organization,
		KeyID:        data.ID.ValueString(),
	}

	tflog.Debug(ctx, "Reading private registry GPG key")
	key, err := d.config.Client.GPGKeys.Read(ctx, keyID)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read private registry GPG key", err.Error())
		return
	}

	data = modelFromTFEVGPGKey(key)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
