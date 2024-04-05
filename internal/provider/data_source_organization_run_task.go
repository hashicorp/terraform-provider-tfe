// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &dataSourceOrganizationRunTask{}
	_ datasource.DataSourceWithConfigure = &dataSourceOrganizationRunTask{}
)

// TODO: This model, and the following conversion function, need to be kept the same as the Org. Run Task Resource
// (internal/provider/resource_tfe_organization_run_task.go) but it only differs by the HMAC Key. In the future we
// should put in a change to add HMAC Key into this model and then we can share the struct. And more importantly we
// we can use the later schema changes.
type modelDataTFEOrganizationRunTaskV0 struct {
	Category     types.String `tfsdk:"category"`
	Description  types.String `tfsdk:"description"`
	Enabled      types.Bool   `tfsdk:"enabled"`
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Organization types.String `tfsdk:"organization"`
	URL          types.String `tfsdk:"url"`
}

func dataModelFromTFEOrganizationRunTask(v *tfe.RunTask) modelDataTFEOrganizationRunTaskV0 {
	result := modelDataTFEOrganizationRunTaskV0{
		Category:     types.StringValue(v.Category),
		Description:  types.StringValue(v.Description),
		Enabled:      types.BoolValue(v.Enabled),
		ID:           types.StringValue(v.ID),
		Name:         types.StringValue(v.Name),
		Organization: types.StringValue(v.Organization.Name),
		URL:          types.StringValue(v.URL),
	}

	return result
}

// NewOrganizationRunTaskDataSource is a helper function to simplify the provider implementation.
func NewOrganizationRunTaskDataSource() datasource.DataSource {
	return &dataSourceOrganizationRunTask{}
}

// dataSourceOrganizationRunTask is the data source implementation.
type dataSourceOrganizationRunTask struct {
	config ConfiguredClient
}

// Metadata returns the data source type name.
func (d *dataSourceOrganizationRunTask) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_run_task"
}

func (d *dataSourceOrganizationRunTask) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Service-generated identifier for the task",
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"organization": schema.StringAttribute{
				Optional: true,
			},
			"url": schema.StringAttribute{
				Optional: true,
			},
			"category": schema.StringAttribute{
				Optional: true,
			},
			"enabled": schema.BoolAttribute{
				Optional: true,
			},
			"description": schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *dataSourceOrganizationRunTask) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *dataSourceOrganizationRunTask) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data modelDataTFEOrganizationRunTaskV0

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

	task, err := fetchOrganizationRunTask(data.Name.ValueString(), organization, d.config.Client)
	if err != nil {
		resp.Diagnostics.AddError("Error reading Organization Run Task",
			fmt.Sprintf("Could not read Run Task %q in organization %q, unexpected error: %s", data.Name.String(), organization, err.Error()),
		)
		return
	}

	// We can never read the HMACkey (Write-only) so assume it's the default (empty)
	result := dataModelFromTFEOrganizationRunTask(task)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}
