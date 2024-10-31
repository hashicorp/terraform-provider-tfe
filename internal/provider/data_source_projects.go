// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework/attr"
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
type modelTFEProject struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	Organization types.String `tfsdk:"organization"`
}

// modelFromTFEProject builds a modelTFEProject struct from a
// tfe.Project value.
func modelFromTFEProject(v *tfe.Project) modelTFEProject {
	return modelTFEProject{
		ID:           types.StringValue(v.ID),
		Name:         types.StringValue(v.Name),
		Description:  types.StringValue(v.Description),
		Organization: types.StringValue(v.Organization.Name),
	}
}

// dataSourceTFEProjects is the data source implementation.
type dataSourceTFEProjects struct {
	config ConfiguredClient
}

// modelTFEProjects maps the data source schema data.
type modelTFEProjects struct {
	ID           types.String      `tfsdk:"id"`
	Organization types.String      `tfsdk:"organization"`
	Projects     []modelTFEProject `tfsdk:"projects"`
}

// Metadata returns the data source type name.
func (d *dataSourceTFEProjects) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_projects"
}

// Schema defines the schema for the data source.
func (d *dataSourceTFEProjects) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This data source can be used to retrieve all projects in an organization.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"organization": schema.StringAttribute{
				Description: "Name of the organization. If omitted, organization must be defined in the provider config.",
				Optional:    true,
				Computed:    true,
			},
			"projects": schema.ListAttribute{
				Description: "List of Projects in the organization.",
				Computed:    true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"id":           types.StringType,
						"name":         types.StringType,
						"description":  types.StringType,
						"organization": types.StringType,
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

	options := tfe.ProjectListOptions{
		ListOptions: tfe.ListOptions{
			PageSize: 100,
		},
	}
	tflog.Debug(ctx, "Listing projects")
	projectList, err := d.config.Client.Projects.List(ctx, organization, &options)
	if err != nil {
		resp.Diagnostics.AddError("Unable to list projects", err.Error())
		return
	}

	model.ID = types.StringValue(organization)
	model.Organization = types.StringValue(organization)
	model.Projects = []modelTFEProject{}

	for { // paginate
		for _, project := range projectList.Items {
			model.Projects = append(model.Projects, modelFromTFEProject(project))
		}

		if projectList.CurrentPage >= projectList.TotalPages {
			break
		}
		options.PageNumber = projectList.NextPage

		tflog.Debug(ctx, "Listing projects")
		projectList, err = d.config.Client.Projects.List(ctx, organization, &options)
		if err != nil {
			resp.Diagnostics.AddError("Unable to list projects", err.Error())
			return
		}
	}

	// Save model into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
