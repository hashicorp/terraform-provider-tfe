// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource              = &dataSourceTFEProviderSet{}
	_ datasource.DataSourceWithConfigure = &dataSourceTFEProviderSet{}
)

// Schema implements datasource.DataSource.
func (d *dataSourceTFEProviderSet) Schema(
	ctx context.Context,
	req datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "This data source can be used to retrieve a provider set.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "The name of the provider set.",
				Required:    true,
			},
			"id": schema.StringAttribute{
				Description: "The ID of the provider set.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "The description of the provider set.",
				Computed:    true,
			},
			"global": schema.BoolAttribute{
				Description: "Whether the provider set applies globally.",
				Computed:    true,
			},
			"organization": schema.StringAttribute{
				Description: "Name of the organization. If omitted, organization must be defined in the provider config.",
				Computed:    true,
				Optional:    true,
			},
			"workspace_ids": schema.SetAttribute{
				Description: "The workspace IDs attached to the provider set.",
				ElementType: types.StringType,
				Computed:    true,
			},
			"project_ids": schema.SetAttribute{
				Description: "The project IDs attached to the provider set.",
				ElementType: types.StringType,
				Computed:    true,
			},
			"provider_source": schema.StringAttribute{
				Description: "Source address of the provider, e.g. registry.terraform.io/hashicorp/tfe.",
				Computed:    true,
			},
		},
	}
}

type modelDataSourceTFEProviderSet struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	Global         types.Bool   `tfsdk:"global"`
	Organization   types.String `tfsdk:"organization"`
	WorkspaceIDs   types.Set    `tfsdk:"workspace_ids"`
	ProjectIDs     types.Set    `tfsdk:"project_ids"`
	ProviderSource types.String `tfsdk:"provider_source"`
}

// modelDataSourceFromTFEProviderSet builds a modelDataSourceFromTFEProviderSet struct from a tfe.ProviderSet
func modelDataSourceFromTFEProviderSet(
	ctx context.Context,
	v tfe.ProviderSet,
) (m modelDataSourceTFEProviderSet, diags diag.Diagnostics) {
	organization := ""
	if v.Organization != nil {
		organization = v.Organization.Name
	}

	// Initialize all fields from the provided API struct
	m = modelDataSourceTFEProviderSet{
		ID:             types.StringValue(v.ID),
		Name:           types.StringValue(v.Name),
		Description:    types.StringValue(v.Description),
		Global:         types.BoolValue(v.Global),
		Organization:   types.StringValue(organization),
		ProviderSource: types.StringValue(v.ProviderSource),
		ProjectIDs:     types.SetNull(types.StringType),
		WorkspaceIDs:   types.SetNull(types.StringType),
	}

	projectIDs := make([]string, len(v.Projects))
	for i, project := range v.Projects {
		projectIDs[i] = project.ID
	}

	var d diag.Diagnostics

	if len(projectIDs) > 0 {
		m.ProjectIDs, d = types.SetValueFrom(ctx, types.StringType, projectIDs)
		diags.Append(d...)
	}

	workspaceIDs := make([]string, len(v.Workspaces))
	for i, workspace := range v.Workspaces {
		workspaceIDs[i] = workspace.ID
	}
	if len(workspaceIDs) > 0 {
		m.WorkspaceIDs, d = types.SetValueFrom(ctx, types.StringType, workspaceIDs)
		diags.Append(d...)
	}
	return m, diags
}

type dataSourceTFEProviderSet struct {
	config ConfiguredClient
}

// NewProviderSetDataSource is a helper function to simplify the provider implementation.
func NewProviderSetDataSource() datasource.DataSource {
	return &dataSourceTFEProviderSet{}
}

// Configure implements datasource.DataSourceWithConfigure.
func (d *dataSourceTFEProviderSet) Configure(
	ctx context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
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

// Metadata implements datasource.DataSource.
func (d *dataSourceTFEProviderSet) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_provider_set"
}

// Read implements datasource.DataSource.
func (d *dataSourceTFEProviderSet) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var config modelDataSourceTFEProviderSet
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var organization string
	resp.Diagnostics.Append(d.config.dataOrDefaultOrganization(ctx, req.Config, &organization)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Read provider set: %s", config.Name.ValueString()))
	ps, err := d.config.Client.ProviderSets.ReadByName(ctx, organization, config.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving provider set", err.Error())
		return
	}
	m, diags := modelDataSourceFromTFEProviderSet(ctx, *ps)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, m)...)
}
