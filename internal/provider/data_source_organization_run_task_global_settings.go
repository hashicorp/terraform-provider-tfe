// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-tfe/v2/api/models"
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
				MarkdownDescription: "The enforcement level of the global task. Valid values are `advisory` and `mandatory`.",
				Optional:            true,
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Service-generated identifier for the task settings.",
			},
			"stages": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Which stages the task will run in. Valid values are one or more of `pre_plan`, `post_plan`, `pre_apply` and `post_apply`.",
				Optional:            true,
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

	taskEnvelope, err := d.config.ClientV2.API.Tasks().ById(taskID).Get(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving task",
			fmt.Sprintf("Error retrieving task %s: %s", taskID, err.Error()),
		)
		return
	}

	var task models.Tasksable
	if taskEnvelope != nil {
		task = taskEnvelope.GetData()
	}
	if task == nil {
		resp.Diagnostics.AddError("Error retrieving task",
			fmt.Sprintf("Error retrieving task %s", taskID),
		)
		return
	}

	if taskGlobalConfiguration(task) == nil {
		resp.Diagnostics.AddWarning("Error retrieving task",
			fmt.Sprintf("The task %s exists however it does not support global run tasks.", taskID),
		)
		return
	}

	result := dataModelFromTFEOrganizationRunTaskGlobalSettingsV2(task)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

// taskGlobalConfiguration returns the task's global configuration attribute,
// or nil when it is not present in the response.
func taskGlobalConfiguration(task models.Tasksable) models.Tasks_attributes_globalConfigurationable {
	if attributes := task.GetAttributes(); attributes != nil {
		return attributes.GetGlobalConfiguration()
	}
	return nil
}

// dataModelFromTFEOrganizationRunTaskGlobalSettingsV2 is the go-tfe v2
// counterpart of dataModelFromTFEOrganizationRunTaskGlobalSettings. The v1
// version remains until the tfe_organization_run_task_global_settings
// resource is migrated.
func dataModelFromTFEOrganizationRunTaskGlobalSettingsV2(v models.Tasksable) modelDataTFEOrganizationRunTaskGlobalSettings {
	result := modelDataTFEOrganizationRunTaskGlobalSettings{
		Enabled:          types.BoolNull(),
		ID:               types.StringValue(valueOrZero(v.GetId())),
		TaskID:           types.StringValue(valueOrZero(v.GetId())),
		EnforcementLevel: types.StringNull(),
		Stages:           types.ListNull(types.StringType),
	}

	global := taskGlobalConfiguration(v)
	if global == nil {
		return result
	}

	result.Enabled = types.BoolValue(valueOrZero(global.GetEnabled()))
	result.EnforcementLevel = types.StringValue(valueOrZero(global.GetEnforcementLevel()))
	if stages, err := types.ListValueFrom(ctx, types.StringType, global.GetStages()); err == nil {
		result.Stages = stages
	}

	return result
}
