// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	tfev2 "github.com/hashicorp/go-tfe/v2"
	"github.com/hashicorp/go-tfe/v2/api/models"
	organizationsapi "github.com/hashicorp/go-tfe/v2/api/organizations"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource              = &dataSourceTFEProject{}
	_ datasource.DataSourceWithConfigure = &dataSourceTFEProject{}
)

func NewProjectDataSource() datasource.DataSource {
	return &dataSourceTFEProject{}
}

type modelDataSourceTFEProject struct {
	ID                          types.String `tfsdk:"id"`
	Name                        types.String `tfsdk:"name"`
	Description                 types.String `tfsdk:"description"`
	Organization                types.String `tfsdk:"organization"`
	AutoDestroyActivityDuration types.String `tfsdk:"auto_destroy_activity_duration"`
	WorkspaceIDs                types.Set    `tfsdk:"workspace_ids"`
	WorkspaceNames              types.Set    `tfsdk:"workspace_names"`
	EffectiveTags               types.Map    `tfsdk:"effective_tags"`
}

// modelDataSourceFromTFEProject builds a modelDataSourceTFEProject struct from a v2 project resource.
func modelDataSourceFromTFEProject(p models.Projectsable, workspaces map[string]string, effectiveTags []models.EffectiveTagBindingsable) modelDataSourceTFEProject {
	m := modelDataSourceTFEProject{
		ID:           types.StringValue(valueOrZero(p.GetId())),
		Organization: types.StringValue(projectOrganizationID(p.GetRelationships())),
	}

	if attrs := p.GetAttributes(); attrs != nil {
		m.Name = types.StringValue(valueOrZero(attrs.GetName()))
		m.Description = types.StringValue(valueOrZero(attrs.GetDescription()))
		if duration := attrs.GetAutoDestroyActivityDuration(); duration != nil {
			m.AutoDestroyActivityDuration = types.StringValue(*duration)
		}
	}

	var wids, wnames []attr.Value
	for workspaceID, workspaceName := range workspaces {
		wids = append(wids, types.StringValue(workspaceID))
		wnames = append(wnames, types.StringValue(workspaceName))
	}
	m.WorkspaceIDs = types.SetValueMust(types.StringType, wids)
	m.WorkspaceNames = types.SetValueMust(types.StringType, wnames)

	tagElems := make(map[string]attr.Value)
	for _, binding := range effectiveTags {
		if binding == nil || binding.GetAttributes() == nil {
			continue
		}
		tagElems[valueOrZero(binding.GetAttributes().GetKey())] = types.StringValue(valueOrZero(binding.GetAttributes().GetValue()))
	}
	m.EffectiveTags = types.MapValueMust(types.StringType, tagElems)

	return m
}

type dataSourceTFEProject struct {
	config ConfiguredClient
}

// Metadata returns the data source type name.
func (d *dataSourceTFEProject) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

// Schema defines the schema for the data source.
func (d *dataSourceTFEProject) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Gets information on a project." +
			"\n\n~> **Note:** The `workspace_ids` and `workspace_names` attributes are not guaranteed to return values in the same order, so they cannot be reliably mapped to one another. To map workspace names to IDs reliably, pass those names to the `tfe_workspace_ids` data source.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The system-generated ID of the project.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the project.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "The description of the project.",
				Computed:    true,
			},
			"organization": schema.StringAttribute{
				Description: "The name of the organization that the project belongs to.",
				Optional:    true,
				Computed:    true,
			},
			"auto_destroy_activity_duration": schema.StringAttribute{
				Description: "The duration after which the project will be auto-destroyed.",
				Computed:    true,
			},
			"workspace_ids": schema.SetAttribute{
				Description: "The IDs of the workspaces associated with the project.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"workspace_names": schema.SetAttribute{
				Description: "The names of the workspaces associated with the project.",
				Computed:    true,
				ElementType: types.StringType,
			},

			"effective_tags": schema.MapAttribute{
				Description: "A map of key-value tags associated with the project.",
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *dataSourceTFEProject) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read implements datasource.DataSource.
func (d *dataSourceTFEProject) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Read Terraform configuration data into the model
	var config modelDataSourceTFEProject
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var organization string
	resp.Diagnostics.Append(d.config.dataOrDefaultOrganization(ctx, req.Config, &organization)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := config.Name.ValueString()

	tflog.Debug(ctx, fmt.Sprintf("Read project: %s", name))

	pageSize := int32(100)
	pageNumber := int32(1)
	for {
		query := &organizationsapi.ItemProjectsRequestBuilderGetQueryParameters{
			Filternames: &name,
			Pagesize:    &pageSize,
			Pagenumber:  &pageNumber,
		}
		projectList, err := d.config.ClientV2.API.Organizations().ByOrganization_name(organization).Projects().Get(ctx, withQueryParams(query))
		if err != nil {
			resp.Diagnostics.AddError("Error retrieving projects", err.Error())
			return
		}
		if projectList == nil {
			break
		}

		for _, proj := range projectList.GetData() {
			if proj == nil || proj.GetAttributes() == nil {
				continue
			}
			// Case-insensitive uniqueness is enforced in TFC
			if !strings.EqualFold(valueOrZero(proj.GetAttributes().GetName()), name) {
				continue
			}

			projID := valueOrZero(proj.GetId())

			// Only now include workspaces to cut down on request load.
			// Store GET /workspaces response in a map to ensure uniqueness
			// key: workspaceID, value: workspaceName
			workspaces := make(map[string]string)
			wsPageNumber := int32(1)
			for {
				wsQuery := &organizationsapi.ItemWorkspacesRequestBuilderGetQueryParameters{
					Filterprojectid: &projID,
					Pagesize:        &pageSize,
					Pagenumber:      &wsPageNumber,
				}
				wl, err := d.config.ClientV2.API.Organizations().ByOrganization_name(organization).Workspaces().Get(ctx, withQueryParams(wsQuery))
				if err != nil {
					resp.Diagnostics.AddError("Error retrieving workspaces", err.Error())
					return
				}
				if wl == nil {
					break
				}

				for _, workspace := range wl.GetData() {
					if workspace == nil || workspace.GetAttributes() == nil {
						continue
					}
					workspaces[valueOrZero(workspace.GetId())] = valueOrZero(workspace.GetAttributes().GetName())
				}

				nextPage := nextPageFromMeta(wl.GetMeta())
				if nextPage == nil {
					break
				}
				wsPageNumber = *nextPage
			}

			bindingsColl, err := d.config.ClientV2.API.Projects().ByProject_id(projID).EffectiveTagBindings().Get(ctx, nil)
			if err != nil && !errors.Is(err, tfev2.ErrNotFound) {
				resp.Diagnostics.AddError(fmt.Sprintf("Error retrieving effective tag bindings for project %s", name), err.Error())
				return
			}

			var effectiveBindings []models.EffectiveTagBindingsable
			if bindingsColl != nil {
				effectiveBindings = bindingsColl.GetData()
			}

			m := modelDataSourceFromTFEProject(proj, workspaces, effectiveBindings)

			resp.Diagnostics.Append(resp.State.Set(ctx, m)...)
			// Update state
			return
		}

		nextPage := nextPageNumber(projectList.GetMeta())
		if nextPage == nil {
			break
		}
		pageNumber = *nextPage
	}

	resp.Diagnostics.AddError("Could not find project", fmt.Sprintf("Project %s/%s not found", organization, name))
}
