// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &dataSourceOrganizationRunTask{}
	_ datasource.DataSourceWithConfigure = &dataSourceOrganizationRunTask{}
)

func NewOrganizationRunTaskGlobalSettingsDataSource() datasource.DataSource {
	return &dataSourceOrganizationRunTaskGlobalSettings{}
}

type dataSourceOrganizationRunTaskGlobalSettings struct {
	config ConfiguredClient
}

func (d *dataSourceOrganizationRunTaskGlobalSettings) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_run_task_global_settings"
}

func (d *dataSourceOrganizationRunTaskGlobalSettings) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				Description: "Whether the run task will be applied globally",
				Optional:    true,
			},
			"enforcement_level": schema.StringAttribute{
				Description: "The enforcement level of the global task.",
				Optional:    true,
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Service-generated identifier for the task settings",
			},
			"stages": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "Which stages the task will run in.",
				Optional:    true,
			},
			"task_id": schema.StringAttribute{
				Description: "The id of the run task.",
				Required:    true,
			},
		},
	}
}

func (d *dataSourceOrganizationRunTaskGlobalSettings) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *dataSourceOrganizationRunTaskGlobalSettings) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data modelDataTFEOrganizationRunTaskGlobalSettings

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	taskID := data.TaskID.ValueString()

	task, err := d.config.Client.RunTasks.Read(ctx, taskID)
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving task",
			fmt.Sprintf("Error retrieving task %s: %s", taskID, err.Error()),
		)
		return
	}

	if task == nil {
		resp.Diagnostics.AddError("Error retrieving task",
			fmt.Sprintf("Error retrieving task %s", taskID),
		)
		return
	}

	if task.Global == nil {
		resp.Diagnostics.AddWarning("Error retrieving task",
			fmt.Sprintf("The task %s exists however it does not support global run tasks.", taskID),
		)
		return
	}

	result := dataModelFromTFEOrganizationRunTaskGlobalSettings(*task)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}
