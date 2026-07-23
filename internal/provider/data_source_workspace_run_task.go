// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-tfe/v2/api/models"
	workspacesapi "github.com/hashicorp/go-tfe/v2/api/workspaces"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	kiota "github.com/microsoft/kiota-abstractions-go"
)

var (
	_ datasource.DataSource              = &dataSourceWorkspaceRunTask{}
	_ datasource.DataSourceWithConfigure = &dataSourceWorkspaceRunTask{}
)

func NewWorkspaceRunTaskDataSource() datasource.DataSource {
	return &dataSourceWorkspaceRunTask{}
}

type dataSourceWorkspaceRunTask struct {
	config ConfiguredClient
}

func (d *dataSourceWorkspaceRunTask) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workspace_run_task"
}

func (d *dataSourceWorkspaceRunTask) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Gets information about a [Workspace Run task](https://developer.hashicorp.com/terraform/cloud-docs/workspaces/settings/run-tasks#associating-run-tasks-with-a-workspace).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Service-generated identifier for the task.",
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
				DeprecationMessage: "The `stage` attribute is deprecated. Use `stages` instead.",
				Description:        "Which stage the task will run in.",
				Computed:           true,
			},
			"stages": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "Which stages the task will run in.",
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

func (d *dataSourceWorkspaceRunTask) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data modelTFEWorkspaceRunTaskV1

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	workspaceID := data.WorkspaceID.ValueString()
	taskID := data.TaskID.ValueString()
	var wstask models.WorkspaceTasksable

	pageNumber := int32(1)
	for {
		query := &workspacesapi.ItemTasksRequestBuilderGetQueryParameters{
			Pagenumber: &pageNumber,
		}
		requestConfig := &kiota.RequestConfiguration[workspacesapi.ItemTasksRequestBuilderGetQueryParameters]{
			QueryParameters: query,
		}

		list, err := d.config.ClientV2.API.Workspaces().ByWorkspace_id(workspaceID).Tasks().Get(ctx, requestConfig)
		if err != nil {
			resp.Diagnostics.AddError("Error retrieving tasks for workspace",
				fmt.Sprintf("Error retrieving tasks for workspace %s: %s", workspaceID, err.Error()),
			)
			return
		}
		if list == nil {
			break
		}

		for _, item := range list.GetData() {
			if item == nil || item.GetRelationships() == nil || item.GetRelationships().GetTask() == nil {
				continue
			}

			taskData := item.GetRelationships().GetTask().GetData()
			if taskData != nil && taskData.GetId() != nil && *taskData.GetId() == taskID {
				wstask = item
				break
			}
		}
		if wstask != nil {
			break
		}
		if list.GetMeta() == nil {
			break
		}

		nextPage, hasNextPage := nextPageFromPagination(list.GetMeta().GetPagination())
		if !hasNextPage {
			break
		}
		pageNumber = nextPage
	}

	if wstask == nil {
		resp.Diagnostics.AddError("Error reading Workspace Run Task",
			fmt.Sprintf("Could not find task %q in workspace %q", taskID, workspaceID),
		)
		return
	}

	result := modelFromTFEWorkspaceRunTaskV2(wstask)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}
