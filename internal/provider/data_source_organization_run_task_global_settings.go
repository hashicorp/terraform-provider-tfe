// Copyright IBM Corp. 2018, 2025
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
		MarkdownDescription: "[Run tasks](https://developer.hashicorp.com/terraform/cloud-docs/workspaces/settings/run-tasks) allow HCP Terraform to interact with external systems at specific points in the HCP Terraform run lifecycle. Run tasks are reusable configurations that you can attach to any workspace in an organization. <br><br> The tfe_organization_run_task_global_settings resource creates, updates and destroys the [global settings](https://developer.hashicorp.com/terraform/cloud-docs/workspaces/settings/run-tasks#global-run-tasks) for an [Organization Run task](https://developer.hashicorp.com/terraform/cloud-docs/workspaces/settings/run-tasks#creating-a-run-task). Your organization must have the `global-run-task` [entitlement](https://developer.hashicorp.com/terraform/cloud-docs/api-docs#feature-entitlements) to use global run tasks.",

		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				Description: "Whether the run task will be applied globally.",
				Optional:    true,
			},
			"enforcement_level": schema.StringAttribute{
				Description: "The enforcement level of the global task. Valid values are `advisory` and `mandatory`",
				Optional:    true,
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Service-generated identifier for the task settings.",
			},
			"stages": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "Which stages the task will run in. Valid values are one or more of `pre_plan`, `post_plan`, `pre_apply` and `post_apply`.",
				Optional:    true,
			},
			"task_id": schema.StringAttribute{
				Description: "The id of the Run task with the global settings.",
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
