// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource                   = &dataSourceTFERegistryProvider{}
	_ datasource.DataSourceWithConfigure      = &dataSourceTFERegistryProvider{}
	_ datasource.DataSourceWithValidateConfig = &dataSourceTFERegistryProvider{}
)

// NewRegistryProviderDataSource is a helper function to simplify the provider implementation.
func NewRegistryProviderDataSource() datasource.DataSource {
	return &dataSourceTFERegistryProvider{}
}

// dataSourceTFERegistryProvider is the data source implementation.
type dataSourceTFERegistryProvider struct {
	config ConfiguredClient
}

// Metadata returns the data source type name.
func (d *dataSourceTFERegistryProvider) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_registry_provider"
}

// Schema defines the schema for the data source.
func (d *dataSourceTFERegistryProvider) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This data source can be used to retrieve a public or private provider from the private registry.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID of the provider.",
				Computed:    true,
			},
			"organization": schema.StringAttribute{
				Description: "Name of the organization. If omitted, organization must be defined in the provider config.",
				Optional:    true,
				Computed:    true,
			},
			"registry_name": schema.StringAttribute{
				Description: "Whether this is a publicly maintained provider or private. Must be either `public` or `private`.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(tfe.PrivateRegistry),
						string(tfe.PublicRegistry),
					),
				},
			},
			"namespace": schema.StringAttribute{
				Description: "The namespace of the provider. For private providers this is the same as the oraganization.",
				Optional:    true,
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the provider.",
				Required:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "The time when the provider was created.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "The time when the provider was last updated.",
				Computed:    true,
			},
		},
	}
}

func (r *dataSourceTFERegistryProvider) ValidateConfig(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	var config modelTFERegistryProvider

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if config.RegistryName.ValueString() == "public" && config.Namespace.IsNull() {
		resp.Diagnostics.AddAttributeError(
			path.Root("namespace"),
			"Missing Attribute Configuration",
			"Expected namespace to be configured when registry_name is \"public\".",
		)
	} else if (config.RegistryName.IsNull() || config.RegistryName.ValueString() == "private") && !config.Namespace.IsNull() && !config.Namespace.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("namespace"),
			"Invalid Attribute Combination",
			"The namespace attribute cannot be configured when registry_name is \"private\".",
		)
	}
}

// Configure adds the provider configured client to the data source.
func (d *dataSourceTFERegistryProvider) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *dataSourceTFERegistryProvider) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data modelTFERegistryProvider

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

	var registryName string
	if data.RegistryName.IsNull() {
		registryName = "private"
	} else {
		registryName = data.RegistryName.ValueString()
	}

	var namespace string
	if registryName == "private" {
		namespace = organization
	} else {
		namespace = data.Namespace.ValueString()
	}

	providerID := tfe.RegistryProviderID{
		OrganizationName: organization,
		RegistryName:     tfe.RegistryName(registryName),
		Namespace:        namespace,
		Name:             data.Name.ValueString(),
	}

	options := tfe.RegistryProviderReadOptions{}

	tflog.Debug(ctx, "Reading private registry provider")
	provider, err := d.config.Client.RegistryProviders.Read(ctx, providerID, &options)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read private registry provider", err.Error())
		return
	}

	data = modelFromTFERegistryProvider(provider)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
