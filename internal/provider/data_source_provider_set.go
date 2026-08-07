// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"

	tfe "github.com/hashicorp/go-tfe/v2"
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
		Description: "Gets a provider set by name." +
			"\n\n~> **Warning:** This data source is currently in beta and isn't generally available to all users. It is subject to change or be removed.",
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
				Description: "The IDs of the workspaces attached to the provider set.",
				ElementType: types.StringType,
				Computed:    true,
			},
			"project_ids": schema.SetAttribute{
				Description: "The IDs of the projects attached to the provider set.",
				ElementType: types.StringType,
				Computed:    true,
			},
			"provider_source": schema.StringAttribute{
				Description: "Source address of the provider, e.g. `registry.terraform.io/hashicorp/tfe`.",
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

// modelDataSourceFromTFEProviderSet builds a modelDataSourceTFEProviderSet struct from a v2 provider set resource.
func modelDataSourceFromTFEProviderSet(
	ctx context.Context,
	v models.ProviderSetsable,
) (m modelDataSourceTFEProviderSet, diags diag.Diagnostics) {
	m = modelDataSourceTFEProviderSet{
		ProjectIDs:   types.SetNull(types.StringType),
		WorkspaceIDs: types.SetNull(types.StringType),
	}

	if id := v.GetId(); id != nil {
		m.ID = types.StringValue(*id)
	}

	if attrs := v.GetAttributes(); attrs != nil {
		if name := attrs.GetName(); name != nil {
			m.Name = types.StringValue(*name)
		}

		description := ""
		if d := attrs.GetDescription(); d != nil {
			description = *d
		}
		m.Description = types.StringValue(description)

		var global bool
		if g := attrs.GetGlobal(); g != nil {
			global = *g
		}
		m.Global = types.BoolValue(global)

		m.Priority = providerSetPriorityValue(attrs.GetPriority())

		if source := attrs.GetProviderSource(); source != nil {
			m.ProviderSource = types.StringValue(*source)
		}
	}

	relationships := v.GetRelationships()

	m.Organization = types.StringValue(providerSetOrganizationID(relationships))

	if relationships == nil {
		return m, diags
	}

	var d diag.Diagnostics

	if projectsRel := relationships.GetProjects(); projectsRel != nil {
		projectData := projectsRel.GetData()
		projectIDs := make([]string, 0, len(projectData))
		for _, p := range projectData {
			if p == nil {
				continue
			}
			if id := p.GetId(); id != nil {
				projectIDs = append(projectIDs, *id)
			}
		}
		if len(projectIDs) > 0 {
			m.ProjectIDs, d = types.SetValueFrom(ctx, types.StringType, projectIDs)
			diags.Append(d...)
		}
	}

	if workspacesRel := relationships.GetWorkspaces(); workspacesRel != nil {
		workspaceData := workspacesRel.GetData()
		workspaceIDs := make([]string, 0, len(workspaceData))
		for _, w := range workspaceData {
			if w == nil {
				continue
			}
			if id := w.GetId(); id != nil {
				workspaceIDs = append(workspaceIDs, *id)
			}
		}
		if len(workspaceIDs) > 0 {
			m.WorkspaceIDs, d = types.SetValueFrom(ctx, types.StringType, workspaceIDs)
			diags.Append(d...)
		}
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
	psEnvelope, err := d.config.ClientV2.API.Organizations().ByOrganization_name(organization).ProviderSets().ByProvider_set_name(config.Name.ValueString()).Get(ctx, nil)
	if err != nil {
		// Preserve the v1 client's "resource not found" wording for not-found
		// errors, since go-tfe/v2 returns "404 Not Found" instead.
		if errors.Is(err, tfe.ErrNotFound) {
			resp.Diagnostics.AddError("Error retrieving provider set", "resource not found")
			return
		}
		resp.Diagnostics.AddError("Error retrieving provider set", err.Error())
		return
	}
	if psEnvelope == nil || psEnvelope.GetData() == nil {
		resp.Diagnostics.AddError("Error retrieving provider set", "no data was returned by the API")
		return
	}

	m, diags := modelDataSourceFromTFEProviderSet(ctx, psEnvelope.GetData())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, m)...)
}
