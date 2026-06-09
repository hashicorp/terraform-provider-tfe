// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/go-tfe"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilterExactSCIMGroups(t *testing.T) {
	groups := []*tfe.AdminSCIMGroup{
		{ID: "sgr-1", Name: "platform-ops-idp"},
		{ID: "sgr-2", Name: "platform-ops-idp-admin-group"},
		{ID: "sgr-3", Name: "platform-ops-idp-eng-group"},
		{ID: "sgr-4", Name: "platform-ops-idp-audit-group"},
	}

	t.Run("case-insensitive exact match", func(t *testing.T) {
		matched := filterExactSCIMGroups(groups, "Platform-Ops-Idp-Admin-Group")
		assert.Len(t, matched, 1)
		assert.Equal(t, "sgr-2", matched[0].ID)
	})

	t.Run("fuzzy substring siblings are rejected", func(t *testing.T) {
		// ?q=platform-ops-idp matches all four as a substring; we keep only
		// the exact name.
		matched := filterExactSCIMGroups(groups, "platform-ops-idp")
		assert.Len(t, matched, 1)
		assert.Equal(t, "sgr-1", matched[0].ID)
	})

	t.Run("no match returns nil", func(t *testing.T) {
		matched := filterExactSCIMGroups(groups, "nonexistent")
		assert.Nil(t, matched)
	})

	t.Run("empty input returns nil", func(t *testing.T) {
		matched := filterExactSCIMGroups(nil, "platform-ops-idp-admin-group")
		assert.Nil(t, matched)
	})

	t.Run("nil entries are skipped", func(t *testing.T) {
		matched := filterExactSCIMGroups([]*tfe.AdminSCIMGroup{nil}, "anything")
		assert.Nil(t, matched)
	})

	t.Run("single exact match", func(t *testing.T) {
		matched := filterExactSCIMGroups(groups, "platform-ops-idp-audit-group")
		assert.Len(t, matched, 1)
		assert.Equal(t, "sgr-4", matched[0].ID)
	})
}

// testSCIMGroupName is the group name searched for throughout the
// findSCIMGroupByName tests.
const testSCIMGroupName = "platform-ops-idp"

// fakeSCIMGroups is a stub tfe.AdminSCIMGroups that returns canned pages and
// records the options passed to each List call so tests can assert on the
// pagination requests findSCIMGroupByName makes.
type fakeSCIMGroups struct {
	pages []*tfe.AdminSCIMGroupList
	// errs, when set, is consulted by call index: a non-nil errs[i] is
	// returned on the (i+1)th List call instead of a page. This lets tests
	// inject a failure on an arbitrary page, not just the first.
	errs  []error
	calls []tfe.AdminSCIMGroupListOptions
}

var _ tfe.AdminSCIMGroups = (*fakeSCIMGroups)(nil)

func (f *fakeSCIMGroups) List(_ context.Context, options *tfe.AdminSCIMGroupListOptions) (*tfe.AdminSCIMGroupList, error) {
	// findSCIMGroupByName mutates and reuses the same options pointer across
	// pages, so snapshot it by value before recording.
	var snapshot tfe.AdminSCIMGroupListOptions
	if options != nil {
		snapshot = *options
	}
	f.calls = append(f.calls, snapshot)

	idx := len(f.calls) - 1
	if idx < len(f.errs) && f.errs[idx] != nil {
		return nil, f.errs[idx]
	}
	if idx >= len(f.pages) {
		return &tfe.AdminSCIMGroupList{Pagination: &tfe.Pagination{}}, nil
	}
	return f.pages[idx], nil
}

// newSCIMGroupsTestClient wires a fake AdminSCIMGroups into the nested client
// path that findSCIMGroupByName reaches through.
func newSCIMGroupsTestClient(groups tfe.AdminSCIMGroups) *tfe.Client {
	return &tfe.Client{
		Admin: tfe.Admin{
			Settings: &tfe.AdminSettings{
				SCIM: &tfe.SCIMResource{
					Groups: groups,
				},
			},
		},
	}
}

func TestFindSCIMGroupByName(t *testing.T) {
	ctx := context.Background()

	t.Run("single page exact match", func(t *testing.T) {
		fake := &fakeSCIMGroups{
			pages: []*tfe.AdminSCIMGroupList{
				{
					Pagination: &tfe.Pagination{CurrentPage: 1, TotalPages: 1},
					Items: []*tfe.AdminSCIMGroup{
						{ID: "sgr-1", Name: "platform-ops-idp"},
					},
				},
			},
		}

		group, err := findSCIMGroupByName(ctx, newSCIMGroupsTestClient(fake), testSCIMGroupName)
		require.NoError(t, err)
		require.NotNil(t, group)
		assert.Equal(t, "sgr-1", group.ID)

		// One request, and the name is forwarded as the server-side prefilter.
		require.Len(t, fake.calls, 1)
		assert.Equal(t, testSCIMGroupName, fake.calls[0].Query)
	})

	t.Run("case-insensitive match", func(t *testing.T) {
		fake := &fakeSCIMGroups{
			pages: []*tfe.AdminSCIMGroupList{
				{
					Pagination: &tfe.Pagination{CurrentPage: 1, TotalPages: 1},
					Items: []*tfe.AdminSCIMGroup{
						{ID: "sgr-9", Name: "Platform-Ops-IDP"},
					},
				},
			},
		}

		group, err := findSCIMGroupByName(ctx, newSCIMGroupsTestClient(fake), testSCIMGroupName)
		require.NoError(t, err)
		require.NotNil(t, group)
		assert.Equal(t, "sgr-9", group.ID)
	})

	t.Run("paginates until the match is found", func(t *testing.T) {
		fake := &fakeSCIMGroups{
			pages: []*tfe.AdminSCIMGroupList{
				{
					Pagination: &tfe.Pagination{CurrentPage: 1, TotalPages: 2, NextPage: 2},
					Items: []*tfe.AdminSCIMGroup{
						{ID: "sgr-1", Name: "platform-ops-idp-other"},
					},
				},
				{
					Pagination: &tfe.Pagination{CurrentPage: 2, TotalPages: 2},
					Items: []*tfe.AdminSCIMGroup{
						{ID: "sgr-2", Name: "platform-ops-idp"},
					},
				},
			},
		}

		group, err := findSCIMGroupByName(ctx, newSCIMGroupsTestClient(fake), testSCIMGroupName)
		require.NoError(t, err)
		require.NotNil(t, group)
		assert.Equal(t, "sgr-2", group.ID)

		// Two requests; the second one asks for the next page and still carries
		// the server-side prefilter so it isn't dropped mid-pagination.
		require.Len(t, fake.calls, 2)
		assert.Equal(t, 2, fake.calls[1].PageNumber)
		assert.Equal(t, testSCIMGroupName, fake.calls[1].Query)
	})

	t.Run("stops paginating once a match is found", func(t *testing.T) {
		fake := &fakeSCIMGroups{
			pages: []*tfe.AdminSCIMGroupList{
				{
					// Pagination claims more pages exist, but the match is on
					// this page so no further request should be made.
					Pagination: &tfe.Pagination{CurrentPage: 1, TotalPages: 5, NextPage: 2},
					Items: []*tfe.AdminSCIMGroup{
						{ID: "sgr-1", Name: "platform-ops-idp"},
					},
				},
			},
		}

		group, err := findSCIMGroupByName(ctx, newSCIMGroupsTestClient(fake), testSCIMGroupName)
		require.NoError(t, err)
		require.NotNil(t, group)
		assert.Equal(t, "sgr-1", group.ID)
		assert.Len(t, fake.calls, 1)
	})

	t.Run("no match across all pages returns nil", func(t *testing.T) {
		fake := &fakeSCIMGroups{
			pages: []*tfe.AdminSCIMGroupList{
				{
					Pagination: &tfe.Pagination{CurrentPage: 1, TotalPages: 2, NextPage: 2},
					Items: []*tfe.AdminSCIMGroup{
						{ID: "sgr-1", Name: "platform-ops-idp-bar"},
					},
				},
				{
					Pagination: &tfe.Pagination{CurrentPage: 2, TotalPages: 2},
					Items: []*tfe.AdminSCIMGroup{
						{ID: "sgr-2", Name: "platform-ops-idp-baz"},
					},
				},
			},
		}

		group, err := findSCIMGroupByName(ctx, newSCIMGroupsTestClient(fake), testSCIMGroupName)
		require.NoError(t, err)
		assert.Nil(t, group)
		assert.Len(t, fake.calls, 2)
	})

	t.Run("empty name forwards an empty query and matches nothing named", func(t *testing.T) {
		fake := &fakeSCIMGroups{
			pages: []*tfe.AdminSCIMGroupList{
				{
					Pagination: &tfe.Pagination{CurrentPage: 1, TotalPages: 1},
					Items: []*tfe.AdminSCIMGroup{
						{ID: "sgr-1", Name: "platform-ops-idp"},
					},
				},
			},
		}

		group, err := findSCIMGroupByName(ctx, newSCIMGroupsTestClient(fake), "")
		require.NoError(t, err)
		assert.Nil(t, group)
		require.Len(t, fake.calls, 1)
		assert.Empty(t, fake.calls[0].Query)
	})

	t.Run("nil pagination stops after one page", func(t *testing.T) {
		fake := &fakeSCIMGroups{
			pages: []*tfe.AdminSCIMGroupList{
				{
					Pagination: nil,
					Items: []*tfe.AdminSCIMGroup{
						{ID: "sgr-1", Name: "platform-ops-idp-bar"},
					},
				},
			},
		}

		group, err := findSCIMGroupByName(ctx, newSCIMGroupsTestClient(fake), testSCIMGroupName)
		require.NoError(t, err)
		assert.Nil(t, group)
		assert.Len(t, fake.calls, 1)
	})

	t.Run("list error is wrapped", func(t *testing.T) {
		fake := &fakeSCIMGroups{errs: []error{errors.New("foo")}}

		group, err := findSCIMGroupByName(ctx, newSCIMGroupsTestClient(fake), testSCIMGroupName)
		require.Error(t, err)
		assert.Nil(t, group)
		assert.Contains(t, err.Error(), "unable to list SCIM groups")
		assert.Contains(t, err.Error(), "foo")
	})

	t.Run("error on a later page is wrapped", func(t *testing.T) {
		fake := &fakeSCIMGroups{
			pages: []*tfe.AdminSCIMGroupList{
				{
					Pagination: &tfe.Pagination{CurrentPage: 1, TotalPages: 2, NextPage: 2},
					Items: []*tfe.AdminSCIMGroup{
						{ID: "sgr-1", Name: "platform-ops-idp-bar"},
					},
				},
			},
			// First page succeeds, the second fails.
			errs: []error{nil, errors.New("bar")},
		}

		group, err := findSCIMGroupByName(ctx, newSCIMGroupsTestClient(fake), testSCIMGroupName)
		require.Error(t, err)
		assert.Nil(t, group)
		assert.Contains(t, err.Error(), "unable to list SCIM groups")
		assert.Contains(t, err.Error(), "bar")
		assert.Len(t, fake.calls, 2)
	})

	t.Run("non-advancing pagination errors instead of looping", func(t *testing.T) {
		fake := &fakeSCIMGroups{
			pages: []*tfe.AdminSCIMGroupList{
				{
					// Pagination claims another page exists but NextPage does
					// not advance past CurrentPage, which would otherwise
					// re-fetch the same page forever.
					Pagination: &tfe.Pagination{CurrentPage: 1, TotalPages: 2, NextPage: 1},
					Items: []*tfe.AdminSCIMGroup{
						{ID: "sgr-1", Name: "platform-ops-idp-bar"},
					},
				},
			},
		}

		group, err := findSCIMGroupByName(ctx, newSCIMGroupsTestClient(fake), testSCIMGroupName)
		require.Error(t, err)
		assert.Nil(t, group)
		assert.Contains(t, err.Error(), "pagination did not advance")
		// Only one request was made; the loop bailed instead of repeating.
		assert.Len(t, fake.calls, 1)
	})
}
