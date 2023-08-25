// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/golang/mock/gomock"
	tfe "github.com/hashicorp/go-tfe"
	tfemocks "github.com/hashicorp/go-tfe/mocks"
)

func MockRunsListForWorkspaceQueue(t *testing.T, client *tfe.Client, workspaceIDWithExpectedRun string, workspaceIDWithUnexpectedRun string) {
	ctrl := gomock.NewController(t)
	mockRunsAPI := tfemocks.NewMockRuns(ctrl)

	runListWithExpectedIDNotIncluded := tfe.RunList{
		Items: []*tfe.Run{
			{
				ID:     "run-01",
				Status: tfe.RunPending,
			},
		},
		Pagination: &tfe.Pagination{
			CurrentPage: 1,
			TotalPages:  1,
			TotalCount:  1,
		},
	}

	runListWithExpectedIDIncluded := tfe.RunList{
		Items: []*tfe.Run{
			{
				ID:     "run-01",
				Status: tfe.RunPending,
			},
			{
				ID:     "run-02",
				Status: tfe.RunPending,
			},
			{
				ID:     "run-03",
				Status: tfe.RunApplied,
			},
			{
				ID:     "run-04",
				Status: tfe.RunPending,
			},
			{
				ID:     "run-05",
				Status: tfe.RunApplying,
			},
		},
		Pagination: &tfe.Pagination{
			CurrentPage: 1,
			TotalPages:  1,
			TotalCount:  4,
		},
	}

	mockRunsAPI.
		EXPECT().
		List(gomock.Any(), workspaceIDWithExpectedRun, gomock.Any()).
		Return(&runListWithExpectedIDIncluded, nil).
		AnyTimes()

	mockRunsAPI.
		EXPECT().
		List(gomock.Any(), workspaceIDWithUnexpectedRun, gomock.Any()).
		Return(&runListWithExpectedIDNotIncluded, nil).
		AnyTimes()

	mockRunsAPI.
		EXPECT().
		List(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, tfe.ErrInvalidOrg).
		AnyTimes()

	client.Runs = mockRunsAPI
}

func TestReadRunPositionInWorkspaceQueue(t *testing.T) {
	client := testTfeClient(t, testClientOptions{})
	MockRunsListForWorkspaceQueue(t, client, "ws-1", "ws-2")

	testCases := map[string]struct {
		currentRunID string
		workspace    string
		err          bool
		returnVal    int
	}{
		"when fetching run list returns error": {
			"run-02",
			"ws-unknown",
			true,
			0,
		},
		"when runID is found in the workspace queue": {
			"run-02",
			"ws-1",
			false,
			2,
		},
		"when runID is not found in the workspace queue": {
			"run-02",
			"ws-2",
			false,
			0,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			position, err := readRunPositionInWorkspaceQueue(
				client,
				testCase.currentRunID,
				testCase.workspace,
				false,
				&tfe.Run{
					ID:     "run-02",
					Status: tfe.RunApplying,
				})

			if (err != nil) != testCase.err {
				t.Fatalf("expected error is %t, got %v", testCase.err, err)
			}

			if position != testCase.returnVal {
				t.Fatalf("expected returned value is %d, got %v", testCase.returnVal, position)
			}
		})
	}
}

func MockRunsQueueForOrg(t *testing.T, client *tfe.Client, orgName string, orgNameWithRun string) {
	ctrl := gomock.NewController(t)
	mockRunQueueAPI := tfemocks.NewMockOrganizations(ctrl)

	runQueueWithExpectedIDNotIncluded := tfe.RunQueue{
		Items: []*tfe.Run{
			{
				ID:              "run-01",
				Status:          tfe.RunPending,
				PositionInQueue: 0,
			},
		},
		Pagination: &tfe.Pagination{
			CurrentPage: 1,
			TotalPages:  1,
			TotalCount:  1,
		},
	}

	runQueueWithExpectedIDIncluded := tfe.RunQueue{
		Items: []*tfe.Run{
			{
				ID:              "run-01",
				Status:          tfe.RunPending,
				PositionInQueue: 0,
			},
			{
				ID:              "run-02",
				Status:          tfe.RunPending,
				PositionInQueue: 1,
			},
		},
		Pagination: &tfe.Pagination{
			CurrentPage: 1,
			TotalPages:  1,
			TotalCount:  2,
		},
	}

	mockRunQueueAPI.
		EXPECT().
		ReadRunQueue(gomock.Any(), orgNameWithRun, gomock.Any()).
		Return(&runQueueWithExpectedIDIncluded, nil).
		AnyTimes()

	mockRunQueueAPI.
		EXPECT().
		ReadRunQueue(gomock.Any(), orgName, gomock.Any()).
		Return(&runQueueWithExpectedIDNotIncluded, nil).
		AnyTimes()

	mockRunQueueAPI.
		EXPECT().
		ReadRunQueue(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, tfe.ErrInvalidOrg).
		AnyTimes()

	client.Organizations = mockRunQueueAPI
}

func TestReadRunPositionInOrgQueue(t *testing.T) {
	defaultOrganization := "my-org"
	client := testTfeClient(t, testClientOptions{})
	MockRunsQueueForOrg(t, client, "another-org", defaultOrganization)

	testCases := map[string]struct {
		orgName   string
		err       bool
		returnVal int
	}{
		"when fetching organization run queue returns error": {
			"unknown-org",
			true,
			0,
		},
		"when run is found in organization run queue": {
			defaultOrganization,
			false,
			1,
		},
		"when run is not found in organization run queue": {
			"another-org",
			false,
			0,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			position, err := readRunPositionInOrgQueue(
				client,
				"run-02",
				testCase.orgName,
			)

			if (err != nil) != testCase.err {
				t.Fatalf("expected error is %t, got %v", testCase.err, err)
			}

			if position != testCase.returnVal {
				t.Fatalf("expected returned value is %d, got %v", testCase.returnVal, position)
			}
		})
	}
}
