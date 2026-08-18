// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-tfe/v2/api/models"
	organizationsapi "github.com/hashicorp/go-tfe/v2/api/organizations"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &dataSourceTFEProjects{}
	_ datasource.DataSourceWithConfigure = &dataSourceTFEProjects{}
)

// NewProjectsDataSource is a helper function to simplify the provider implementation.
func NewProjectsDataSource() datasource.DataSource {
	return &dataSourceTFEProjects{}
}

// modelTFEProject maps the resource or data source schema data to a
// struct.
type modelTFEProjectsProject struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	Organization types.String `tfsdk:"organization"`
}

// modelFromTFEProjectsProject builds a modelTFEProjectsProject struct from a v2 project resource.
func modelFromTFEProjectsProject(v models.Projectsable) modelTFEProjectsProject {
	result := modelTFEProjectsProject{
		ID:           types.StringValue(valueOrZero(v.GetId())),
		Organization: types.StringValue(projectOrganizationID(v.GetRelationships())),
	}
	if attrs := v.GetAttributes(); attrs != nil {
		result.Name = types.StringValue(valueOrZero(attrs.GetName()))
		result.Description = types.StringValue(valueOrZero(attrs.GetDescription()))
	}
	return result
}

// projectOrganizationID extracts the ID of a project's organization from its organization relationship.
func projectOrganizationID(relationships models.Projects_relationshipsable) string {
	if relationships == nil {
		return ""
	}
	org := relationships.GetOrganization()
	if org == nil || org.GetData() == nil {
		return ""
	}
	return valueOrZero(org.GetData().GetId())
}

// dataSourceTFEProjects is the data source implementation.
type dataSourceTFEProjects struct {
	config ConfiguredClient
}

// modelTFEProjects maps the data source schema data.
type modelTFEProjects struct {
	ID           types.String              `tfsdk:"id"`
	Organization types.String              `tfsdk:"organization"`
	Projects     []modelTFEProjectsProject `tfsdk:"projects"`
}

// Metadata returns the data source type name.
func (d *dataSourceTFEProjects) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_projects"
}

// Schema defines the schema for the data source.
func (d *dataSourceTFEProjects) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Gets information on all projects in an organization.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Name of the organization for use as an ID.",
				Computed:    true,
			},
			"organization": schema.StringAttribute{
				Description: "Name of the organization. If omitted, organization must be defined in the provider config.",
				Optional:    true,
				Computed:    true,
			},
			"projects": schema.ListNestedAttribute{
				Description: "List of projects in the organization.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "ID of the project.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Name of the project.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "Description of the project.",
							Computed:    true,
						},
						"organization": schema.StringAttribute{
							Description: "Name of the organization. If omitted, organization must be defined in the provider config.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *dataSourceTFEProjects) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *dataSourceTFEProjects) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model modelTFEProjects // The model is what we save to the state

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var organization string
	resp.Diagnostics.Append(d.config.dataOrDefaultOrganization(ctx, req.Config, &organization)...)

	if resp.Diagnostics.HasError() {
		return
	}

	model.ID = types.StringValue(organization)
	model.Organization = types.StringValue(organization)
	model.Projects = []modelTFEProjectsProject{}

	pageSize := int32(100)
	pageNumber := int32(1)
	for { // paginate
		query := &organizationsapi.ItemProjectsRequestBuilderGetQueryParameters{
			Pagesize:   &pageSize,
			Pagenumber: &pageNumber,
		}

		tflog.Debug(ctx, "Listing projects")
		projectList, err := d.config.ClientV2.API.Organizations().ByOrganization_name(organization).Projects().Get(ctx, withQueryParams(query))
		if err != nil {
			resp.Diagnostics.AddError("Unable to list projects", err.Error())
			return
		}
		if projectList == nil {
			break
		}

		for _, project := range projectList.GetData() {
			if project == nil {
				continue
			}
			model.Projects = append(model.Projects, modelFromTFEProjectsProject(project))
		}

		nextPage := nextPageNumber(projectList.GetMeta())
		if nextPage == nil {
			break
		}
		pageNumber = *nextPage
	}

	// Save model into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
