// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-tfe/v2/api/models"
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
		Description: "This data source can be used to retrieve a provider set. Note that this data source is currently in beta and isn't generally available to all users. It is subject to change or be removed.",
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
			"priority": schema.BoolAttribute{
				Description: "Whether the provider set takes priority over provider sets with more specific scopes.",
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
	Priority       types.Bool   `tfsdk:"priority"`
	Organization   types.String `tfsdk:"organization"`
	WorkspaceIDs   types.Set    `tfsdk:"workspace_ids"`
	ProjectIDs     types.Set    `tfsdk:"project_ids"`
	ProviderSource types.String `tfsdk:"provider_source"`
}

// modelDataSourceFromTFEProviderSet builds a Terraform data source model from
// v2 Provider Set data.
func modelDataSourceFromTFEProviderSet(
	ctx context.Context,
	v models.ProviderSetsable,
	organization string,
) (m modelDataSourceTFEProviderSet, diags diag.Diagnostics) {
	attributes := v.GetAttributes()

	// Initialize all fields from the provided API struct
	m = modelDataSourceTFEProviderSet{
		ID:             types.StringValue(*v.GetId()),
		Name:           types.StringValue(providerSetStringValue(attributes.GetName())),
		Description:    types.StringValue(providerSetStringValue(attributes.GetDescription())),
		Global:         types.BoolValue(providerSetBoolValue(attributes.GetGlobal())),
		Priority:       types.BoolValue(providerSetBoolValue(attributes.GetPriority())),
		Organization:   types.StringValue(organization),
		ProviderSource: types.StringValue(providerSetStringValue(attributes.GetProviderSource())),
		ProjectIDs:     types.SetNull(types.StringType),
		WorkspaceIDs:   types.SetNull(types.StringType),
	}

	projectIDs := providerSetProjectIDs(v)

	var d diag.Diagnostics

	if len(projectIDs) > 0 {
		m.ProjectIDs, d = types.SetValueFrom(ctx, types.StringType, projectIDs)
		diags.Append(d...)
	}

	workspaceIDs := providerSetWorkspaceIDs(v)
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

	providerSetName := config.Name.ValueString()
	if err := providerSetOrganizationValidationError(organization); err != nil {
		resp.Diagnostics.AddError("Error retrieving provider set", err.Error())
		return
	}
	if err := providerSetRequiredNameValidationError(providerSetName); err != nil {
		resp.Diagnostics.AddError("Error retrieving provider set", err.Error())
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Read provider set: %s", providerSetName))
	psEnvelope, err := d.config.ClientV2.API.Organizations().ByOrganization_name(organization).ProviderSets().ByProvider_set_name(providerSetName).Get(ctx, nil)
	if err != nil {
		err = providerSetLegacyError(err)
		resp.Diagnostics.AddError("Error retrieving provider set", err.Error())
		return
	}
	ps, err := providerSetDataFromEnvelope(psEnvelope)
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving provider set", err.Error())
		return
	}
	m, diags := modelDataSourceFromTFEProviderSet(ctx, ps, organization)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, m)...)
}
