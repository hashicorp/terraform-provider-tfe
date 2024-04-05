// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &dataSourceWorkspaceRunTask{}
	_ datasource.DataSourceWithConfigure = &dataSourceWorkspaceRunTask{}
)

// NewWorkspaceRunTaskDataSource is a helper function to simplify the provider implementation.
func NewWorkspaceRunTaskDataSource() datasource.DataSource {
	return &dataSourceWorkspaceRunTask{}
}

// dataSourceWorkspaceRunTask is the data source implementation.
type dataSourceWorkspaceRunTask struct {
	config ConfiguredClient
}

// Metadata returns the data source type name.
func (d *dataSourceWorkspaceRunTask) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workspace_run_task"
}

func (d *dataSourceWorkspaceRunTask) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Service-generated identifier for the task",
				Computed:    true,
			},
			"workspace_id": schema.StringAttribute{
				Description: "The id of the workspace.",
				Required:    true,
			},
			"task_id": schema.StringAttribute{
				Description: "The id of the run task.",
				Required:    true,
			},
			"enforcement_level": schema.StringAttribute{
				Description: "The enforcement level of the task.",
				Computed:    true,
			},
			"stage": schema.StringAttribute{
				Description: "Which stage the task will run in.",
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *dataSourceWorkspaceRunTask) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *dataSourceWorkspaceRunTask) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data modelTFEWorkspaceRunTaskV0

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	workspaceID := data.WorkspaceID.ValueString()
	taskID := data.TaskID.ValueString()
	var wstask *tfe.WorkspaceRunTask = nil

	// Create an options struct.
	options := &tfe.WorkspaceRunTaskListOptions{}
	for {
		list, err := d.config.Client.WorkspaceRunTasks.List(ctx, workspaceID, options)
		if err != nil {
			resp.Diagnostics.AddError("Error retrieving tasks for workspace",
				fmt.Sprintf("Error retrieving tasks for workspace %s: %s", workspaceID, err.Error()),
			)
			return
		}

		for _, item := range list.Items {
			if item.RunTask.ID == taskID {
				wstask = item
				break
			}
		}
		if wstask != nil {
			break
		}

		// Exit the loop when we've seen all pages.
		if list.CurrentPage >= list.TotalPages {
			break
		}

		// Update the page number to get the next page.
		options.PageNumber = list.NextPage
	}

	if wstask == nil {
		resp.Diagnostics.AddError("Error reading Workspace Run Task",
			fmt.Sprintf("Could not find task %q in workspace %q", taskID, workspaceID),
		)
		return
	}

	result := modelFromTFEWorkspaceRunTask(wstask)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}
