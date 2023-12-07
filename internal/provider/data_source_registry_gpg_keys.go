// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &dataSourceTFERegistryGPGKeys{}
	_ datasource.DataSourceWithConfigure = &dataSourceTFERegistryGPGKeys{}
)

// NewRegistryGPGKeysDataSource is a helper function to simplify the provider implementation.
func NewRegistryGPGKeysDataSource() datasource.DataSource {
	return &dataSourceTFERegistryGPGKeys{}
}

// dataSourceTFERegistryGPGKeys is the data source implementation.
type dataSourceTFERegistryGPGKeys struct {
	config ConfiguredClient
}

// modelTFERegistryGPGKeys maps the data source schema data.
type modelTFERegistryGPGKeys struct {
	ID           types.String             `tfsdk:"id"`
	Organization types.String             `tfsdk:"organization"`
	Keys         []modelTFERegistryGPGKey `tfsdk:"keys"`
}

// Metadata returns the data source type name.
func (d *dataSourceTFERegistryGPGKeys) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_registry_gpg_keys"
}

// Schema defines the schema for the data source.
func (d *dataSourceTFERegistryGPGKeys) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This data source can be used to retrieve all private registry GPG keys of an organization.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"organization": schema.StringAttribute{
				Description: "Name of the organization. If omitted, organization must be defined in the provider config.",
				Optional:    true,
				Computed:    true,
			},
			"keys": schema.ListAttribute{
				Description: "List of GPG keys in the organization.",
				Computed:    true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"id":           types.StringType,
						"organization": types.StringType,
						"ascii_armor":  types.StringType,
						"created_at":   types.StringType,
						"updated_at":   types.StringType,
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *dataSourceTFERegistryGPGKeys) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *dataSourceTFERegistryGPGKeys) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data modelTFERegistryGPGKeys

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

	options := tfe.GPGKeyListOptions{
		Namespaces: []string{organization},
	}
	tflog.Debug(ctx, "Listing private registry GPG keys")
	keyList, err := d.config.Client.GPGKeys.ListPrivate(ctx, options)
	if err != nil {
		resp.Diagnostics.AddError("Unable to list private registry GPG keys", err.Error())
		return
	}

	data.ID = types.StringValue(organization)
	data.Organization = types.StringValue(organization)
	data.Keys = []modelTFERegistryGPGKey{}

	for {
		for _, key := range keyList.Items {
			data.Keys = append(data.Keys, modelFromTFEVGPGKey(key))
		}

		if keyList.CurrentPage >= keyList.TotalPages {
			break
		}
		options.PageNumber = keyList.NextPage

		tflog.Debug(ctx, "Listing private registry GPG keys")
		keyList, err = d.config.Client.GPGKeys.ListPrivate(ctx, options)
		if err != nil {
			resp.Diagnostics.AddError("Unable to list private registry GPG keys", err.Error())
			return
		}
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
