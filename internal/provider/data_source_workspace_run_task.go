// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTFEWorkspaceRunTask() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTFEWorkspaceRunTaskRead,

		Schema: map[string]*schema.Schema{
			"workspace_id": {
				Description: "The id of the workspace.",
				Type:        schema.TypeString,
				Required:    true,
			},

			"task_id": {
				Description: "The id of the run task.",
				Type:        schema.TypeString,
				Required:    true,
			},

			"enforcement_level": {
				Description: "The enforcement level of the task.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"stage": {
				Description: "Which stage the task will run in.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourceTFEWorkspaceRunTaskRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	workspaceID := d.Get("workspace_id").(string)
	taskID := d.Get("task_id").(string)

	// Create an options struct.
	options := &tfe.WorkspaceRunTaskListOptions{}
	for {
		list, err := config.Client.WorkspaceRunTasks.List(ctx, workspaceID, options)
		if err != nil {
			return fmt.Errorf("Error retrieving tasks for workspace %s: %w", workspaceID, err)
		}

		for _, wstask := range list.Items {
			if wstask.RunTask.ID == taskID {
				d.Set("enforcement_level", string(wstask.EnforcementLevel))
				d.Set("stage", string(wstask.Stage))
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

	return fmt.Errorf("could not find workspace run task %s in workspace %s", taskID, workspaceID)
}
