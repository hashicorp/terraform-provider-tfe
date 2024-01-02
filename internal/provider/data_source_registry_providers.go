// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &dataSourceTFERegistryProviders{}
	_ datasource.DataSourceWithConfigure = &dataSourceTFERegistryProviders{}
)

// NewRegistryProvidersDataSource is a helper function to simplify the provider implementation.
func NewRegistryProvidersDataSource() datasource.DataSource {
	return &dataSourceTFERegistryProviders{}
}

// dataSourceTFERegistryProviders is the data source implementation.
type dataSourceTFERegistryProviders struct {
	config ConfiguredClient
}

// modelTFERegistryProviders maps the data source schema data.
type modelTFERegistryProviders struct {
	ID           types.String               `tfsdk:"id"`
	Organization types.String               `tfsdk:"organization"`
	RegistryName types.String               `tfsdk:"registry_name"`
	Search       types.String               `tfsdk:"search"`
	Providers    []modelTFERegistryProvider `tfsdk:"providers"`
}

// Metadata returns the data source type name.
func (d *dataSourceTFERegistryProviders) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_registry_providers"
}

// Schema defines the schema for the data source.
func (d *dataSourceTFERegistryProviders) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This data source can be used to retrieve public and private providers from the private registry of an organization.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"organization": schema.StringAttribute{
				Description: "Name of the organization. If omitted, organization must be defined in the provider config.",
				Optional:    true,
				Computed:    true,
			},
			"registry_name": schema.StringAttribute{
				Description: "Whether to list only public or private providers. Must be either `public` or `private`.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(tfe.PrivateRegistry),
						string(tfe.PublicRegistry),
					),
				},
			},
			"search": schema.StringAttribute{
				Description: "A query string to do a fuzzy search on provider name and namespace.",
				Optional:    true,
			},
			"providers": schema.ListAttribute{
				Description: "List of providers in the organization.",
				Computed:    true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"id":            types.StringType,
						"organization":  types.StringType,
						"registry_name": types.StringType,
						"namespace":     types.StringType,
						"name":          types.StringType,
						"created_at":    types.StringType,
						"updated_at":    types.StringType,
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *dataSourceTFERegistryProviders) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *dataSourceTFERegistryProviders) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data modelTFERegistryProviders

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

	var registryName tfe.RegistryName
	if !data.RegistryName.IsNull() {
		registryName = tfe.RegistryName(data.RegistryName.ValueString())
	}

	options := tfe.RegistryProviderListOptions{
		RegistryName: registryName,
		Search:       data.Search.ValueString(),
	}

	tflog.Debug(ctx, "Listing private registry providers")
	providerList, err := d.config.Client.RegistryProviders.List(ctx, organization, &options)
	if err != nil {
		resp.Diagnostics.AddError("Unable to list private registry providers", err.Error())
		return
	}

	data.ID = types.StringValue(organization)
	data.Organization = types.StringValue(organization)
	data.Providers = []modelTFERegistryProvider{}

	for {
		for _, provider := range providerList.Items {
			data.Providers = append(data.Providers, modelFromTFERegistryProvider(provider))
		}

		if providerList.CurrentPage >= providerList.TotalPages {
			break
		}
		options.PageNumber = providerList.NextPage

		tflog.Debug(ctx, "Listing private registry providers")
		providerList, err = d.config.Client.RegistryProviders.List(ctx, organization, &options)
		if err != nil {
			resp.Diagnostics.AddError("Unable to list private registry providers", err.Error())
			return
		}
	}
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
