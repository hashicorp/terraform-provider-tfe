// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"net/http"
	"testing"
)

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
