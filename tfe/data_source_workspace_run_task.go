package tfe

import (
	"fmt"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTFEWorkspaceRunTask() *schema.Resource {
	return &schema.Resource{
		Description: "[Run tasks](https://www.terraform.io/cloud-docs/workspaces/settings/run-tasks) allow Terraform Cloud to interact with external systems at specific points in the Terraform Cloud run lifecycle. Run tasks are reusable configurations that you can attach to any workspace in an organization." +
			"\n\nUse this data source to get information about a [Workspace Run tasks](https://www.terraform.io/cloud-docs/workspaces/settings/run-tasks#associating-run-tasks-with-a-workspace).",

		Read: dataSourceTFEWorkspaceRunTaskRead,

		Schema: map[string]*schema.Schema{
			"workspace_id": {
				Description: "The id of the workspace.",
				Type:        schema.TypeString,
				Required:    true,
			},

			"task_id": {
				Description: "The id of the Run task.",
				Type:        schema.TypeString,
				Required:    true,
			},

			"enforcement_level": {
				Description: "The enforcement level of the task.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourceTFEWorkspaceRunTaskRead(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	workspaceID := d.Get("workspace_id").(string)
	taskID := d.Get("task_id").(string)

	// Create an options struct.
	options := &tfe.WorkspaceRunTaskListOptions{}
	for {
		list, err := tfeClient.WorkspaceRunTasks.List(ctx, workspaceID, options)
		if err != nil {
			return fmt.Errorf("Error retrieving tasks for workspace %s: %w", workspaceID, err)
		}

		for _, wstask := range list.Items {
			if wstask.RunTask.ID == taskID {
				d.Set("enforcement_level", string(wstask.EnforcementLevel))
				d.SetId(wstask.ID)
				return nil
			}
		}

		// Exit the loop when we've seen all pages.
		if list.CurrentPage >= list.TotalPages {
			break
		}

		// Update the page number to get the next page.
		options.PageNumber = list.NextPage
	}

	return fmt.Errorf("Could not find workspace run task %s in workspace %s", taskID, workspaceID)
}
