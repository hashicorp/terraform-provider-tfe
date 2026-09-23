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
	_ datasource.DataSource              = &dataSourceTFERegistryProviderVersions{}
	_ datasource.DataSourceWithConfigure = &dataSourceTFERegistryProviderVersions{}
)

// NewRegistryProviderVersionsDataSource is a helper function to simplify the provider implementation.
func NewRegistryProviderVersionsDataSource() datasource.DataSource {
	return &dataSourceTFERegistryProviderVersions{}
}

// dataSourceTFERegistryProviderVersions is the data source implementation.
type dataSourceTFERegistryProviderVersions struct {
	config ConfiguredClient
}

// modelTFERegistryProviderVersionReadOnly is a read-only view of a provider version
// used in list data sources. It omits write-only resource inputs (shasums_file, shasums_sig_file).
type modelTFERegistryProviderVersionReadOnly struct {
	ID                 types.String `tfsdk:"id"`
	Organization       types.String `tfsdk:"organization"`
	RegistryName       types.String `tfsdk:"registry_name"`
	Namespace          types.String `tfsdk:"namespace"`
	ProviderName       types.String `tfsdk:"name"`
	Version            types.String `tfsdk:"version"`
	KeyID              types.String `tfsdk:"key_id"`
	Protocols          types.List   `tfsdk:"protocols"`
	ShasumsUploaded    types.Bool   `tfsdk:"shasums_uploaded"`
	ShasumsSigUploaded types.Bool   `tfsdk:"shasums_sig_uploaded"`
	CreatedAt          types.String `tfsdk:"created_at"`
	UpdatedAt          types.String `tfsdk:"updated_at"`
}

// modelTFERegistryProviderVersions maps the data source schema data.
type modelTFERegistryProviderVersions struct {
	ID           types.String                              `tfsdk:"id"`
	Organization types.String                              `tfsdk:"organization"`
	RegistryName types.String                              `tfsdk:"registry_name"`
	Namespace    types.String                              `tfsdk:"namespace"`
	ProviderName types.String                              `tfsdk:"name"`
	Versions     []modelTFERegistryProviderVersionReadOnly `tfsdk:"versions"`
}

// Metadata returns the data source type name.
func (d *dataSourceTFERegistryProviderVersions) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_registry_provider_versions"
}

// Schema defines the schema for the data source.
func (d *dataSourceTFERegistryProviderVersions) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This data source can be used to retrieve all versions for a provider in the private registry.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID of the data source (computed from organization, registry_name, namespace, and provider name).",
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
			"versions": schema.ListNestedAttribute{
				Description: "List of provider versions.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "ID of the provider version.",
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
							Description: "The version string.",
							Computed:    true,
						},
						"key_id": schema.StringAttribute{
							Description: "The GPG key ID used to sign the provider version.",
							Computed:    true,
						},
						"protocols": schema.ListAttribute{
							Description: "An array of Terraform provider API versions that this version supports.",
							Computed:    true,
							ElementType: types.StringType,
						},
						"shasums_uploaded": schema.BoolAttribute{
							Description: "Indicates whether the SHASUMS file has been uploaded.",
							Computed:    true,
						},
						"shasums_sig_uploaded": schema.BoolAttribute{
							Description: "Indicates whether the SHASUMS signature file has been uploaded.",
							Computed:    true,
						},
						"created_at": schema.StringAttribute{
							Description: "The time when the provider version was created.",
							Computed:    true,
						},
						"updated_at": schema.StringAttribute{
							Description: "The time when the provider version was last updated.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *dataSourceTFERegistryProviderVersions) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *dataSourceTFERegistryProviderVersions) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data modelTFERegistryProviderVersions

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

	providerID := tfe.RegistryProviderID{
		OrganizationName: organization,
		RegistryName:     tfe.RegistryName(registryName),
		Namespace:        namespace,
		Name:             data.ProviderName.ValueString(),
	}

	options := &tfe.RegistryProviderVersionListOptions{}

	tflog.Debug(ctx, "Listing registry provider versions")
	versionList, err := d.config.Client.RegistryProviderVersions.List(ctx, providerID, options)
	if err != nil {
		resp.Diagnostics.AddError("Unable to list registry provider versions", err.Error())
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("%s/%s/%s/%s", organization, registryName, namespace, data.ProviderName.ValueString()))
	data.Organization = types.StringValue(organization)
	data.RegistryName = types.StringValue(registryName)
	data.Namespace = types.StringValue(namespace)
	data.Versions = []modelTFERegistryProviderVersionReadOnly{}

	if versionList != nil {
		for _, version := range versionList.Items {
			protocols, _ := types.ListValueFrom(context.Background(), types.StringType, version.Protocols)
			data.Versions = append(data.Versions, modelTFERegistryProviderVersionReadOnly{
				ID:                 types.StringValue(version.ID),
				Organization:       types.StringValue(organization),
				RegistryName:       types.StringValue(registryName),
				Namespace:          types.StringValue(namespace),
				ProviderName:       types.StringValue(data.ProviderName.ValueString()),
				Version:            types.StringValue(version.Version),
				KeyID:              types.StringValue(version.KeyID),
				Protocols:          protocols,
				ShasumsUploaded:    types.BoolValue(version.ShasumsUploaded),
				ShasumsSigUploaded: types.BoolValue(version.ShasumsSigUploaded),
				CreatedAt:          types.StringValue(version.CreatedAt),
				UpdatedAt:          types.StringValue(version.UpdatedAt),
			})
		}
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
