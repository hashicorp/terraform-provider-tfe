// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"

	tfe "github.com/hashicorp/go-tfe"
)

// fetchOrganizationRunTask returns the task in an organization by name
func fetchOrganizationRunTask(name, organization string, client *tfe.Client) (*tfe.RunTask, error) {
	options := &tfe.RunTaskListOptions{}
	for {
		list, err := client.RunTasks.List(ctx, organization, options)
		if err != nil {
			return nil, fmt.Errorf("Error retrieving organization tasks: %w", err)
		}

		for _, task := range list.Items {
			if task != nil && task.Name == name {
				return task, nil
			}
		}

		// Exit the loop when we've seen all pages.
		if list.CurrentPage >= list.TotalPages {
			break
		}

		// Update the page number to get the next page.
		options.PageNumber = list.NextPage
	}

	return nil, fmt.Errorf("could not find organization run task for organization %s and name %s", organization, name)
}

// fetchWorkspaceRunTask returns the task association in a workspace by name
func fetchWorkspaceRunTask(name, workspace, organization string, client *tfe.Client) (*tfe.WorkspaceRunTask, error) {
	task, err := fetchOrganizationRunTask(name, organization, client)
	if err != nil {
		return nil, fmt.Errorf("Error reading configuration of task %s in organization %s: %w", name, organization, err)
	}

	ws, err := client.Workspaces.Read(ctx, organization, workspace)
	if err != nil {
		return nil, fmt.Errorf("Error reading configuration of workspace %s in organization %s: %w", workspace, organization, err)
	}

	options := &tfe.WorkspaceRunTaskListOptions{}
	for {
		list, err := client.WorkspaceRunTasks.List(ctx, ws.ID, options)
		if err != nil {
			return nil, fmt.Errorf("Error retrieving workspace run tasks: %w", err)
		}
		for _, wstask := range list.Items {
			if wstask != nil && wstask.RunTask.ID == task.ID {
				return wstask, nil
			}
		}

		// Exit the loop when we've seen all pages.
		if list.CurrentPage >= list.TotalPages {
			break
		}

		// Update the page number to get the next page.
		options.PageNumber = list.NextPage
	}

	return nil, fmt.Errorf("could not find organization run task %s for workspace %s in organization %s", name, workspace, organization)
}
