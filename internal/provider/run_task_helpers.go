// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	tfe "github.com/hashicorp/go-tfe"
	tfev2 "github.com/hashicorp/go-tfe/v2"
	"github.com/hashicorp/go-tfe/v2/api/models"
	"github.com/hashicorp/go-tfe/v2/api/organizations"
	abstractions "github.com/microsoft/kiota-abstractions-go"
)

// fetchOrganizationRunTaskV2 is the go-tfe v2 counterpart of
// fetchOrganizationRunTask. The v1 version remains until the resources that
// use it for imports are migrated.
func fetchOrganizationRunTaskV2(ctx context.Context, name, organization string, client *tfev2.Client) (models.Tasksable, error) {
	tasksBuilder := client.API.Organizations().ByOrganization_name(organization).Tasks()

	pageSize := int32(100)
	queryParams := &organizations.ItemTasksRequestBuilderGetQueryParameters{
		Pagesize: &pageSize,
	}
	for {
		list, err := tasksBuilder.Get(ctx, &abstractions.RequestConfiguration[organizations.ItemTasksRequestBuilderGetQueryParameters]{
			QueryParameters: queryParams,
		})
		if err != nil {
			return nil, fmt.Errorf("Error retrieving organization tasks: %w", err)
		}

		for _, task := range list.GetData() {
			if task == nil {
				continue
			}
			if attributes := task.GetAttributes(); attributes != nil && valueOrZero(attributes.GetName()) == name {
				return task, nil
			}
		}

		// Exit the loop when we've seen all pages.
		var nextPage *int32
		if meta := list.GetMeta(); meta != nil {
			nextPage = nextPageNumber(meta.GetPagination())
		}
		if nextPage == nil {
			break
		}

		// Update the page number to get the next page.
		queryParams.Pagenumber = nextPage
	}

	return nil, fmt.Errorf("could not find organization run task for organization %s and name %s", organization, name)
}

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
