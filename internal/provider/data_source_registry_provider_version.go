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
	_ datasource.DataSource              = &dataSourceTFERegistryProviderVersion{}
	_ datasource.DataSourceWithConfigure = &dataSourceTFERegistryProviderVersion{}
)

// NewRegistryProviderVersionDataSource is a helper function to simplify the provider implementation.
func NewRegistryProviderVersionDataSource() datasource.DataSource {
	return &dataSourceTFERegistryProviderVersion{}
}

// dataSourceTFERegistryProviderVersion is the data source implementation.
type dataSourceTFERegistryProviderVersion struct {
	config ConfiguredClient
}

// modelTFERegistryProviderVersionDataSource is the data source model with platforms
type modelTFERegistryProviderVersionDataSource struct {
	ID                 types.String                              `tfsdk:"id"`
	Organization       types.String                              `tfsdk:"organization"`
	RegistryName       types.String                              `tfsdk:"registry_name"`
	Namespace          types.String                              `tfsdk:"namespace"`
	ProviderName       types.String                              `tfsdk:"name"`
	Version            types.String                              `tfsdk:"version"`
	KeyID              types.String                              `tfsdk:"key_id"`
	Protocols          types.List                                `tfsdk:"protocols"`
	ShasumsUploaded    types.Bool                                `tfsdk:"shasums_uploaded"`
	ShasumsSigUploaded types.Bool                                `tfsdk:"shasums_sig_uploaded"`
	Platforms          []modelTFERegistryProviderVersionPlatform `tfsdk:"platforms"`
	CreatedAt          types.String                              `tfsdk:"created_at"`
	UpdatedAt          types.String                              `tfsdk:"updated_at"`
}

// Metadata returns the data source type name.
func (d *dataSourceTFERegistryProviderVersion) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_registry_provider_version"
}

// Schema defines the schema for the data source.
func (d *dataSourceTFERegistryProviderVersion) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This data source can be used to retrieve a provider version from the private registry.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID of the provider version.",
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
				Description: "The version of the provider.",
				Required:    true,
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
			"platforms": schema.ListNestedAttribute{
				Description: "List of platform release files for this provider version.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "ID of the provider version platform.",
							Computed:    true,
						},
						"os_arch": schema.StringAttribute{
							Description: "A valid operating system string and architecture string, separated by an underscore (e.g., 'linux_amd64').",
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
func (d *dataSourceTFERegistryProviderVersion) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *dataSourceTFERegistryProviderVersion) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data modelTFERegistryProviderVersionDataSource

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

	tflog.Debug(ctx, "Reading registry provider version")
	providerVersion, err := d.config.Client.RegistryProviderVersions.Read(ctx, versionID)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read registry provider version", err.Error())
		return
	}

	baseResult := modelFromTFERegistryProviderVersion(providerVersion, organization, registryName, namespace, data.ProviderName.ValueString())

	// Create data source result with platforms
	result := modelTFERegistryProviderVersionDataSource{
		ID:                 baseResult.ID,
		Organization:       baseResult.Organization,
		RegistryName:       baseResult.RegistryName,
		Namespace:          baseResult.Namespace,
		ProviderName:       baseResult.ProviderName,
		Version:            baseResult.Version,
		KeyID:              baseResult.KeyID,
		Protocols:          baseResult.Protocols,
		ShasumsUploaded:    baseResult.ShasumsUploaded,
		ShasumsSigUploaded: baseResult.ShasumsSigUploaded,
		CreatedAt:          baseResult.CreatedAt,
		UpdatedAt:          baseResult.UpdatedAt,
	}

	// Fetch platforms for this version
	tflog.Debug(ctx, "Listing registry provider version platforms")
	platforms, err := d.config.Client.RegistryProviderPlatforms.List(ctx, versionID, nil)
	if err != nil {
		resp.Diagnostics.AddError("Unable to list registry provider version platforms", err.Error())
		return
	}

	// Convert platforms to model
	result.Platforms = make([]modelTFERegistryProviderVersionPlatform, 0, len(platforms.Items))
	for _, platform := range platforms.Items {
		result.Platforms = append(result.Platforms, modelFromTFERegistryProviderVersionPlatform(
			platform,
			organization,
			registryName,
			namespace,
			data.ProviderName.ValueString(),
			data.Version.ValueString(),
		))
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}
