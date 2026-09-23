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
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &dataSourceTFERegistryProviderVersionPlatforms{}
	_ datasource.DataSourceWithConfigure = &dataSourceTFERegistryProviderVersionPlatforms{}
)

// NewRegistryProviderVersionPlatformsDataSource is a helper function to simplify the provider implementation.
func NewRegistryProviderVersionPlatformsDataSource() datasource.DataSource {
	return &dataSourceTFERegistryProviderVersionPlatforms{}
}

// dataSourceTFERegistryProviderVersionPlatforms is the data source implementation.
type dataSourceTFERegistryProviderVersionPlatforms struct {
	config ConfiguredClient
}

// modelTFERegistryProviderVersionPlatforms maps the data source schema data.
type modelTFERegistryProviderVersionPlatforms struct {
	ID           types.String                              `tfsdk:"id"`
	Organization types.String                              `tfsdk:"organization"`
	RegistryName types.String                              `tfsdk:"registry_name"`
	Namespace    types.String                              `tfsdk:"namespace"`
	ProviderName types.String                              `tfsdk:"name"`
	Version      types.String                              `tfsdk:"version"`
	Platforms    []modelTFERegistryProviderVersionPlatform `tfsdk:"platforms"`
}

// Metadata returns the data source type name.
func (d *dataSourceTFERegistryProviderVersionPlatforms) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_registry_provider_version_platforms"
}

// Schema defines the schema for the data source.
func (d *dataSourceTFERegistryProviderVersionPlatforms) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This data source can be used to retrieve all platforms for a provider version in the private registry.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID of the data source (computed from organization, registry_name, namespace, provider name, and version).",
				Computed:    true,
			},
			"organization": schema.StringAttribute{
				Description: "Name of the organization. If omitted, organization must be defined in the provider config.",
				Optional:    true,
				Computed:    true,
			},
			"registry_name": schema.StringAttribute{
				Description: "Whether this is a publicly maintained provider or private. Must be either `public` or `private`. Defaults to `private`.",
				Optional:    true,
				Computed:    true,
			},
			"namespace": schema.StringAttribute{
				Description: "The namespace of the provider. For private providers this is the same as the organization.",
				Optional:    true,
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the provider.",
				Required:    true,
			},
			"version": schema.StringAttribute{
				Description: "Version of the provider.",
				Required:    true,
			},
			"platforms": schema.ListNestedAttribute{
				Description: "List of provider version platforms.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "ID of the provider version platform.",
							Computed:    true,
						},
						"registry_provider_id": schema.StringAttribute{
							Description: "ID of the registry provider version this platform belongs to.",
							Computed:    true,
						},
						"organization": schema.StringAttribute{
							Description: "Name of the organization.",
							Computed:    true,
						},
						"registry_name": schema.StringAttribute{
							Description: "Whether this is a publicly maintained provider or private.",
							Computed:    true,
						},
						"namespace": schema.StringAttribute{
							Description: "The namespace of the provider.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Name of the provider.",
							Computed:    true,
						},
						"version": schema.StringAttribute{
							Description: "Version of the provider.",
							Computed:    true,
						},
						"os_arch": schema.StringAttribute{
							Description: "A valid operating system string and architecture string, separated by an underscore.",
							Computed:    true,
						},
						"filename": schema.StringAttribute{
							Description: "The filename of the provider binary.",
							Computed:    true,
						},
						"shasum": schema.StringAttribute{
							Description: "The SHA256 checksum of the provider binary.",
							Computed:    true,
						},
						"provider_binary_uploaded": schema.BoolAttribute{
							Description: "Indicates whether the provider binary has been uploaded.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *dataSourceTFERegistryProviderVersionPlatforms) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *dataSourceTFERegistryProviderVersionPlatforms) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data modelTFERegistryProviderVersionPlatforms

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

	// Determine registry name and namespace
	var registryName string
	if data.RegistryName.IsNull() {
		registryName = string(tfe.PrivateRegistry)
	} else {
		registryName = data.RegistryName.ValueString()
	}

	var namespace string
	if registryName == string(tfe.PrivateRegistry) {
		namespace = organization
	} else {
		namespace = data.Namespace.ValueString()
	}

	versionID := tfe.RegistryProviderVersionID{
		RegistryProviderID: tfe.RegistryProviderID{
			OrganizationName: organization,
			RegistryName:     tfe.RegistryName(registryName),
			Namespace:        namespace,
			Name:             data.ProviderName.ValueString(),
		},
		Version: data.Version.ValueString(),
	}

	options := &tfe.RegistryProviderPlatformListOptions{}

	tflog.Debug(ctx, "Listing registry provider version platforms")
	platformList, err := d.config.Client.RegistryProviderPlatforms.List(ctx, versionID, options)
	if err != nil {
		resp.Diagnostics.AddError("Unable to list registry provider version platforms", err.Error())
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("%s/%s/%s/%s/%s", organization, registryName, namespace, data.ProviderName.ValueString(), data.Version.ValueString()))
	data.Organization = types.StringValue(organization)
	data.RegistryName = types.StringValue(registryName)
	data.Namespace = types.StringValue(namespace)
	data.Platforms = []modelTFERegistryProviderVersionPlatform{}

	if platformList != nil {
		for _, platform := range platformList.Items {
			data.Platforms = append(data.Platforms, modelFromTFERegistryProviderVersionPlatform(platform, organization, registryName, namespace, data.ProviderName.ValueString(), data.Version.ValueString()))
		}
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
