// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/go-tfe"
	tfemocks "github.com/hashicorp/go-tfe/mocks"
	"go.uber.org/mock/gomock"
)

func MockASingleOrgRunTask(t *testing.T, client *tfe.Client, task tfe.RunTask) {
	ctrl := gomock.NewController(t)

	mockRunTaskAPI := tfemocks.NewMockRunTasks(ctrl)
	list := tfe.RunTaskList{
		Items: make([]*tfe.RunTask, 0),
	}
	list.Items = append(list.Items, &task)
	list.Pagination = &tfe.Pagination{
		CurrentPage: 1,
		TotalPages:  1,
		TotalCount:  len(list.Items),
	}

	// Mock a good List response
	mockRunTaskAPI.
		EXPECT().
		List(gomock.Any(), task.Organization.Name, gomock.Any()).
		Return(&list, nil).
		AnyTimes()

	// Mock a bad List response
	mockRunTaskAPI.
		EXPECT().
		List(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, tfe.ErrInvalidOrg).
		AnyTimes()

	// Wire up the mock interfaces
	client.RunTasks = mockRunTaskAPI
}

func MockASingleWorkspaceRunTask(t *testing.T, client *tfe.Client, wsTask tfe.WorkspaceRunTask) {
	ctrl := gomock.NewController(t)

	mockWsRunTaskAPI := tfemocks.NewMockWorkspaceRunTasks(ctrl)
	list := tfe.WorkspaceRunTaskList{
		Items: make([]*tfe.WorkspaceRunTask, 0),
	}
	list.Items = append(list.Items, &wsTask)
	list.Pagination = &tfe.Pagination{
		CurrentPage: 1,
		TotalPages:  1,
		TotalCount:  len(list.Items),
	}

	// Mock a good List response
	mockWsRunTaskAPI.
		EXPECT().
		List(gomock.Any(), wsTask.Workspace.ID, gomock.Any()).
		Return(&list, nil).
		AnyTimes()

	// Mock a bad List response
	mockWsRunTaskAPI.
		EXPECT().
		List(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, tfe.ErrInvalidWorkspaceID).
		AnyTimes()

	// Wire up the mock interfaces
	client.WorkspaceRunTasks = mockWsRunTaskAPI
}

func TestFetchOrganizationRunTask(t *testing.T) {
	orgName := "hashicorp"

	tests := map[string]struct {
		taskName     string
		org          string
		expectExists bool
		err          bool
	}{
		"non exisiting organization": {
			"a-task",
			"not-an-org",
			false,
			true,
		},
		"non exisiting task": {
			"not-a-task",
			orgName,
			false,
			true,
		},
		"existing task": {
			"a-task",
			orgName,
			true,
			false,
		},
	}

	client := testTfeClient(t, testClientOptions{defaultOrganization: orgName})
	// Mock the Task
	task := tfe.RunTask{
		ID:       "task-123",
		Name:     "a-task",
		URL:      runTasksURL(),
		Category: "task",
		Organization: &tfe.Organization{
			Name: orgName,
		},
	}
	MockASingleOrgRunTask(t, client, task)

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := fetchOrganizationRunTask(test.taskName, test.org, client)

			if (err != nil) != test.err {
				t.Fatalf("expected error is %t, got %v", test.err, err)
			}

			if test.expectExists {
				if got == nil || got.Name != test.taskName {
					t.Fatalf("wrong result\ngot: %#v\nwant: %#v", got, nil)
				}
			} else {
				if got != nil {
					t.Fatalf("wrong result\ngot: %#v\nwant: %#v", got, nil)
				}
			}
		})
	}
}

func TestFetchOrganizationRunTaskV2(t *testing.T) {
	orgName := "hashicorp"

	taskResource := func(taskID, name string) string {
		return fmt.Sprintf(`{
			"id": %q,
			"type": "tasks",
			"attributes": {"name": %q, "url": "https://example.com", "category": "task", "enabled": true},
			"relationships": {"organization": {"data": {"id": %q, "type": "organizations"}}}
		}`, taskID, name, orgName)
	}

	pages := map[string]string{
		"1": fmt.Sprintf(`{"data": [%s], "meta": {"pagination": {"current-page": 1, "next-page": 2, "total-pages": 2}}}`,
			taskResource("task-123", "a-task"),
		),
		"2": fmt.Sprintf(`{"data": [%s], "meta": {"pagination": {"current-page": 2, "next-page": null, "total-pages": 2}}}`,
			taskResource("task-456", "b-task"),
		),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v2/organizations/"+orgName+"/tasks", func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page[number]")
		if page == "" {
			page = "1"
		}
		body, ok := pages[page]
		if !ok {
			http.Error(w, `{"errors":[{"status":"404","title":"not found"}]}`, http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/vnd.api+json")
		fmt.Fprint(w, body)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"errors":[{"status":"404","title":"not found"}]}`, http.StatusNotFound)
	})

	tests := map[string]struct {
		taskName     string
		org          string
		expectExists bool
		expectedID   string
		err          bool
	}{
		"non existing organization": {
			"a-task",
			"not-an-org",
			false,
			"",
			true,
		},
		"non existing task": {
			"not-a-task",
			orgName,
			false,
			"",
			true,
		},
		"existing task on the first page": {
			"a-task",
			orgName,
			true,
			"task-123",
			false,
		},
		"existing task on a later page": {
			"b-task",
			orgName,
			true,
			"task-456",
			false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			client := testTfeClientV2(t, mux)

			got, err := fetchOrganizationRunTaskV2(test.taskName, test.org, client)

			if (err != nil) != test.err {
				t.Fatalf("expected error is %t, got %v", test.err, err)
			}

			if test.expectExists {
				if got == nil || valueOrZero(got.GetId()) != test.expectedID {
					t.Fatalf("wrong result\ngot: %#v\nwant task with ID %q", got, test.expectedID)
				}
			} else if got != nil {
				t.Fatalf("wrong result\ngot: %#v\nwant: nil", got)
			}
		})
	}
}

func TestFetchWorkspaceRunTask(t *testing.T) {
	orgName := "hashicorp"
	workspaceName := "a-workspace"
	taskname := "a-task"

	tests := map[string]struct {
		org          string
		workspace    string
		taskName     string
		expectExists bool
		err          bool
	}{
		"non exisiting organization": {
			"not-an-org",
			workspaceName,
			taskname,
			false,
			true,
		},
		"non exisiting workspace": {
			orgName,
			"not-a-workspace",
			taskname,
			false,
			true,
		},
		"non exisiting run task": {
			orgName,
			workspaceName,
			"not-a-task",
			false,
			true,
		},
		"an existing workspace run task": {
			orgName,
			workspaceName,
			taskname,
			true,
			false,
		},
	}

	client := testTfeClient(t, testClientOptions{defaultOrganization: orgName})
	// TODO : Use gomocks for the workspace
	ws, _ := client.Workspaces.Create(context.TODO(), orgName, tfe.WorkspaceCreateOptions{
		Name: &workspaceName,
	})

	// Mock the Task
	task := tfe.RunTask{
		ID:       "task-123",
		Name:     taskname,
		URL:      runTasksURL(),
		Category: "task",
		Organization: &tfe.Organization{
			Name: orgName,
		},
	}
	MockASingleOrgRunTask(t, client, task)

	// Mock the Workspace Task
	wsTask := tfe.WorkspaceRunTask{
		ID:               "wstask-123",
		EnforcementLevel: tfe.Mandatory,
		RunTask:          &task,
		Workspace:        ws,
	}
	MockASingleWorkspaceRunTask(t, client, wsTask)

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := fetchWorkspaceRunTask(test.taskName, test.workspace, test.org, client)

			if (err != nil) != test.err {
				t.Fatalf("expected error is %t, got %v", test.err, err)
			}

			if test.expectExists {
				if got == nil || got.RunTask.Name != test.taskName {
					t.Fatalf("wrong result\ngot: %#v\nwant: %#v", got, nil)
				}
			} else {
				if got != nil {
					t.Fatalf("wrong result\ngot: %#v\nwant: %#v", got, nil)
				}
			}
		})
	}
}
