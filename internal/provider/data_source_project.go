// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
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

func modelDataSourceFromTFEProject(p *tfe.Project, workspaceIDs, workspaceNames []string, effectiveTags []*tfe.EffectiveTagBinding) (modelDataSourceTFEProject, diag.Diagnostics) {
	m := modelDataSourceTFEProject{
		ID:           types.StringValue(p.ID),
		Name:         types.StringValue(p.Name),
		Description:  types.StringValue(p.Description),
		Organization: types.StringValue(p.Organization.Name),
	}

	var wids, wnames []attr.Value
	for w := range workspaceIDs {
		wids = append(wids, types.StringValue(workspaceIDs[w]))
		wnames = append(wnames, types.StringValue(workspaceNames[w]))
	}
	m.WorkspaceIDs = types.SetValueMust(types.StringType, wids)
	m.WorkspaceNames = types.SetValueMust(types.StringType, wnames)

	var diags diag.Diagnostics
	if p.AutoDestroyActivityDuration.IsSpecified() {
		autoDestroyDuration, err := p.AutoDestroyActivityDuration.Get()
		if err != nil {
			diags.AddAttributeError(path.Root("auto_destroy_activity_duration"), "Error reading auto destroy activity duration", err.Error())
			return m, diags
		}

		m.AutoDestroyActivityDuration = types.StringValue(autoDestroyDuration)
	}

	tagElems := make(map[string]attr.Value)
	for _, binding := range effectiveTags {
		tagElems[binding.Key] = types.StringValue(binding.Value)
	}
	m.EffectiveTags = types.MapValueMust(types.StringType, tagElems)

	return m, diags
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
		Description: "This data source can be used to retrieve a project in an organization.",
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

	// Create an options struct.
	name := config.Name.ValueString()
	options := &tfe.ProjectListOptions{
		Name: name,
	}

	tflog.Debug(ctx, fmt.Sprintf("Read project: %s", name))
	l, err := d.config.Client.Projects.List(ctx, organization, options)
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving projects", err.Error())
		return
	}

	for _, proj := range l.Items {
		// Case-insensitive uniqueness is enforced in TFC
		if !strings.EqualFold(proj.Name, name) {
			continue
		}

		// Only now include workspaces to cut down on request load.
		readOptions := &tfe.WorkspaceListOptions{
			ProjectID: proj.ID,
		}

		var workspaceIDs []string
		var workspaceNames []string
		for {
			wl, err := d.config.Client.Workspaces.List(ctx, organization, readOptions)
			if err != nil {
				resp.Diagnostics.AddError("Error retrieving workspaces", err.Error())
				return
			}

			for _, workspace := range wl.Items {
				workspaceIDs = append(workspaceIDs, workspace.ID)
				workspaceNames = append(workspaceNames, workspace.Name)
			}

			// Exit the loop when we've seen all pages.
			if wl.CurrentPage >= wl.TotalPages {
				break
			}

			// Update the page number to get the next page.
			readOptions.PageNumber = wl.NextPage
		}

		effectiveBindings, err := d.config.Client.Projects.ListEffectiveTagBindings(ctx, proj.ID)
		if err != nil && !errors.Is(err, tfe.ErrResourceNotFound) {
			resp.Diagnostics.AddError(fmt.Sprintf("Error retrieving effective tag bindings for project %s", proj.Name), err.Error())
			return
		}
		if err != nil {
			// This endpoint may not be supported against a given TFE instance.
			// Initialize to empty slice to avoid ranging over nil
			effectiveBindings = []*tfe.EffectiveTagBinding{}
		}

		m, diags := modelDataSourceFromTFEProject(proj, workspaceIDs, workspaceNames, effectiveBindings)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		resp.Diagnostics.Append(resp.State.Set(ctx, m)...)
		// Update state
		return
	}

	resp.Diagnostics.AddError("Could not find project", fmt.Sprintf("Project %s/%s not found", organization, name))
}
