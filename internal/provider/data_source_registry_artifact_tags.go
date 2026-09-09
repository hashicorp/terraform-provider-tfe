// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
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
	_ datasource.DataSource              = &dataSourceTFERegistryArtifactTags{}
	_ datasource.DataSourceWithConfigure = &dataSourceTFERegistryArtifactTags{}
)

// NewRegistryArtifactTagsDataSource is a helper function to simplify the provider implementation.
func NewRegistryArtifactTagsDataSource() datasource.DataSource {
	return &dataSourceTFERegistryArtifactTags{}
}

// modelRegistryArtifactTagsData maps the data source schema data to a struct.
type modelRegistryArtifactTagsData struct {
	ID       types.String  `tfsdk:"id"`
	Tags     []modelTag    `tfsdk:"tags"`
	Artifact modelArtifact `tfsdk:"artifact"`
}

// dataSourceTFERegistryArtifactTags is the data source implementation.
type dataSourceTFERegistryArtifactTags struct {
	config ConfiguredClient
}

// Metadata returns the data source type name.
func (d *dataSourceTFERegistryArtifactTags) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_registry_artifact_tags"
}

// Schema defines the schema for the data source.
func (d *dataSourceTFERegistryArtifactTags) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve the tags associated with a registry artifact.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of this data source, in the format <artifact_type>/<artifact_id>.",
				Computed:    true,
			},
			"tags": schema.ListAttribute{
				Description: "The list of tags associated with the registry artifact.",
				Computed:    true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"key":   types.StringType,
						"value": types.StringType,
					},
				},
			},
			"artifact": schema.SingleNestedAttribute{
				Required:    true,
				Description: "The registry artifact to retrieve tags for.",
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Description: "The type of the registry artifact.",
						Required:    true,
						Validators: []validator.String{
							stringvalidator.OneOf(
								ArtifactTypeRegistryModule,
								ArtifactTypeRegistryProvider,
								ArtifactTypeRegistryComponent,
							),
						},
					},
					"id": schema.StringAttribute{
						Description: "The external ID of the registry artifact.",
						Required:    true,
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *dataSourceTFERegistryArtifactTags) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *dataSourceTFERegistryArtifactTags) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data modelRegistryArtifactTagsData
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := data.Artifact.ID.ValueString()
	atype := data.Artifact.Type.ValueString()

	tflog.Debug(ctx, fmt.Sprintf("Reading tags for registry artifact type %s with ID %s", atype, id))

	bindings, err := d.listTagBindings(ctx, atype, id)
	if errors.Is(err, tfe.ErrResourceNotFound) {
		resp.Diagnostics.AddError(
			"Registry Artifact Not Found",
			fmt.Sprintf("The registry artifact with ID %q was not found.", id),
		)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Error reading tags for registry artifact", err.Error())
		return
	}

	data.Tags = modelFromTFETagBindings(bindings)
	data.ID = types.StringValue(fmt.Sprintf("%s/%s", atype, id))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// listTagBindings fetches the current tag bindings for a registry artifact
// by dispatching on the artifact type.
func (d *dataSourceTFERegistryArtifactTags) listTagBindings(ctx context.Context, artifactType, artifactID string) ([]*tfe.TagBinding, error) {
	switch artifactType {
	case ArtifactTypeRegistryModule:
		return d.config.Client.RegistryModules.ListTagBindings(ctx, artifactID)
	case ArtifactTypeRegistryProvider:
		return d.config.Client.RegistryProviders.ListTagBindings(ctx, artifactID)
	case ArtifactTypeRegistryComponent:
		return d.config.Client.RegistryComponents.ListTagBindings(ctx, artifactID)
	default:
		return nil, fmt.Errorf("unsupported artifact type: %s", artifactType)
	}
}
