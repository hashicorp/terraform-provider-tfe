// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/hashicorp/go-tfe/v2/api/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func scimGroupResource(id, name string) string {
	return fmt.Sprintf(`{"id": %q, "type": "scim-groups", "attributes": {"name": %q}}`, id, name)
}

func TestFilterExactSCIMGroups(t *testing.T) {
	groups := []models.AdminScimGroupsable{
		mustParseSCIMGroup(t, "sgr-1", "platform-ops-idp"),
		mustParseSCIMGroup(t, "sgr-2", "platform-ops-idp-admin-group"),
		mustParseSCIMGroup(t, "sgr-3", "platform-ops-idp-eng-group"),
		mustParseSCIMGroup(t, "sgr-4", "platform-ops-idp-audit-group"),
	}

	t.Run("case-insensitive exact match", func(t *testing.T) {
		matched := filterExactSCIMGroups(groups, "Platform-Ops-Idp-Admin-Group")
		assert.Len(t, matched, 1)
		assert.Equal(t, "sgr-2", valueOrZero(matched[0].GetId()))
	})

	t.Run("fuzzy substring siblings are rejected", func(t *testing.T) {
		// ?q=platform-ops-idp matches all four as a substring; we keep only
		// the exact name.
		matched := filterExactSCIMGroups(groups, "platform-ops-idp")
		assert.Len(t, matched, 1)
		assert.Equal(t, "sgr-1", valueOrZero(matched[0].GetId()))
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
		matched := filterExactSCIMGroups([]models.AdminScimGroupsable{nil}, "anything")
		assert.Nil(t, matched)
	})

	t.Run("single exact match", func(t *testing.T) {
		matched := filterExactSCIMGroups(groups, "platform-ops-idp-audit-group")
		assert.Len(t, matched, 1)
		assert.Equal(t, "sgr-4", valueOrZero(matched[0].GetId()))
	})
}

// mustParseSCIMGroup builds a models.AdminScimGroupsable directly, without
// going through JSON, for use in in-memory filterExactSCIMGroups tests.
func mustParseSCIMGroup(t *testing.T, id, name string) models.AdminScimGroupsable {
	t.Helper()
	g := models.NewAdminScimGroups()
	g.SetId(&id)
	attrs := models.NewAdminScimGroups_attributes()
	attrs.SetName(&name)
	g.SetAttributes(attrs)
	return g
}

// testSCIMGroupName is the group name searched for throughout the
// findSCIMGroupByName tests.
const testSCIMGroupName = "platform-ops-idp"

// scimGroupsHandler serves GET /api/v2/admin/scim-groups with the given
// JSON:API pages (served in order by the page[number] query parameter), and
// records every query it receives so tests can assert on the requests
// findSCIMGroupByName makes.
func scimGroupsHandler(pages map[string]string, calls *[]string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v2/admin/scim-groups", func(w http.ResponseWriter, r *http.Request) {
		*calls = append(*calls, r.URL.RawQuery)

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
	return mux
}

func TestFindSCIMGroupByName(t *testing.T) {
	ctx := context.Background()

	t.Run("single page exact match", func(t *testing.T) {
		var calls []string
		pages := map[string]string{
			"1": fmt.Sprintf(`{"data": [%s], "meta": %s}`, scimGroupResource("sgr-1", "platform-ops-idp"), paginationMeta(1, "null", "1")),
		}

		client := testTfeClientV2(t, scimGroupsHandler(pages, &calls))
		group, err := findSCIMGroupByName(ctx, client, testSCIMGroupName)
		require.NoError(t, err)
		require.NotNil(t, group)
		assert.Equal(t, "sgr-1", valueOrZero(group.GetId()))

		// One request, and the name is forwarded as the server-side prefilter.
		require.Len(t, calls, 1)
		assert.Contains(t, calls[0], "q="+testSCIMGroupName)
	})

	t.Run("case-insensitive match", func(t *testing.T) {
		var calls []string
		pages := map[string]string{
			"1": fmt.Sprintf(`{"data": [%s], "meta": %s}`, scimGroupResource("sgr-9", "Platform-Ops-IDP"), paginationMeta(1, "null", "1")),
		}

		client := testTfeClientV2(t, scimGroupsHandler(pages, &calls))
		group, err := findSCIMGroupByName(ctx, client, testSCIMGroupName)
		require.NoError(t, err)
		require.NotNil(t, group)
		assert.Equal(t, "sgr-9", valueOrZero(group.GetId()))
	})

	t.Run("paginates until the match is found", func(t *testing.T) {
		var calls []string
		pages := map[string]string{
			"1": fmt.Sprintf(`{"data": [%s], "meta": %s}`, scimGroupResource("sgr-1", "platform-ops-idp-other"), paginationMeta(1, "2", "2")),
			"2": fmt.Sprintf(`{"data": [%s], "meta": %s}`, scimGroupResource("sgr-2", "platform-ops-idp"), paginationMeta(2, "null", "2")),
		}

		client := testTfeClientV2(t, scimGroupsHandler(pages, &calls))
		group, err := findSCIMGroupByName(ctx, client, testSCIMGroupName)
		require.NoError(t, err)
		require.NotNil(t, group)
		assert.Equal(t, "sgr-2", valueOrZero(group.GetId()))

		// Two requests; the second one asks for the next page and still carries
		// the server-side prefilter so it isn't dropped mid-pagination.
		require.Len(t, calls, 2)
		assert.Contains(t, calls[1], "page%5Bnumber%5D=2")
		assert.Contains(t, calls[1], "q="+testSCIMGroupName)
	})

	t.Run("stops paginating once a match is found", func(t *testing.T) {
		var calls []string
		pages := map[string]string{
			// Pagination claims more pages exist, but the match is on this page
			// so no further request should be made.
			"1": fmt.Sprintf(`{"data": [%s], "meta": %s}`, scimGroupResource("sgr-1", "platform-ops-idp"), paginationMeta(1, "2", "5")),
		}

		client := testTfeClientV2(t, scimGroupsHandler(pages, &calls))
		group, err := findSCIMGroupByName(ctx, client, testSCIMGroupName)
		require.NoError(t, err)
		require.NotNil(t, group)
		assert.Equal(t, "sgr-1", valueOrZero(group.GetId()))
		assert.Len(t, calls, 1)
	})

	t.Run("no match across all pages returns nil", func(t *testing.T) {
		var calls []string
		pages := map[string]string{
			"1": fmt.Sprintf(`{"data": [%s], "meta": %s}`, scimGroupResource("sgr-1", "platform-ops-idp-bar"), paginationMeta(1, "2", "2")),
			"2": fmt.Sprintf(`{"data": [%s], "meta": %s}`, scimGroupResource("sgr-2", "platform-ops-idp-baz"), paginationMeta(2, "null", "2")),
		}

		client := testTfeClientV2(t, scimGroupsHandler(pages, &calls))
		group, err := findSCIMGroupByName(ctx, client, testSCIMGroupName)
		require.NoError(t, err)
		assert.Nil(t, group)
		assert.Len(t, calls, 2)
	})

	t.Run("empty name forwards an empty query and matches nothing named", func(t *testing.T) {
		var calls []string
		pages := map[string]string{
			"1": fmt.Sprintf(`{"data": [%s], "meta": %s}`, scimGroupResource("sgr-1", "platform-ops-idp"), paginationMeta(1, "null", "1")),
		}

		client := testTfeClientV2(t, scimGroupsHandler(pages, &calls))
		group, err := findSCIMGroupByName(ctx, client, "")
		require.NoError(t, err)
		assert.Nil(t, group)
		require.Len(t, calls, 1)
		query, err := url.ParseQuery(calls[0])
		require.NoError(t, err)
		assert.Empty(t, query.Get("q"))
	})

	t.Run("nil pagination meta stops after one page", func(t *testing.T) {
		var calls []string
		pages := map[string]string{
			"1": fmt.Sprintf(`{"data": [%s]}`, scimGroupResource("sgr-1", "platform-ops-idp-bar")),
		}

		client := testTfeClientV2(t, scimGroupsHandler(pages, &calls))
		group, err := findSCIMGroupByName(ctx, client, testSCIMGroupName)
		require.NoError(t, err)
		assert.Nil(t, group)
		assert.Len(t, calls, 1)
	})

	t.Run("list error is wrapped", func(t *testing.T) {
		var calls []string
		client := testTfeClientV2(t, scimGroupsHandler(map[string]string{}, &calls))
		group, err := findSCIMGroupByName(ctx, client, testSCIMGroupName)
		require.Error(t, err)
		assert.Nil(t, group)
		assert.Contains(t, err.Error(), "unable to list SCIM groups")
	})

	t.Run("error on a later page is wrapped", func(t *testing.T) {
		var calls []string
		pages := map[string]string{
			"1": fmt.Sprintf(`{"data": [%s], "meta": %s}`, scimGroupResource("sgr-1", "platform-ops-idp-bar"), paginationMeta(1, "2", "2")),
			// Page 2 deliberately absent, so the handler 404s it.
		}

		client := testTfeClientV2(t, scimGroupsHandler(pages, &calls))
		group, err := findSCIMGroupByName(ctx, client, testSCIMGroupName)
		require.Error(t, err)
		assert.Nil(t, group)
		assert.Contains(t, err.Error(), "unable to list SCIM groups")
		assert.Len(t, calls, 2)
	})

	t.Run("non-advancing pagination errors instead of looping", func(t *testing.T) {
		var calls []string
		pages := map[string]string{
			"1": fmt.Sprintf(`{"data": [%s], "meta": %s}`, scimGroupResource("sgr-1", "platform-ops-idp-bar"), paginationMeta(1, "1", "2")),
		}

		client := testTfeClientV2(t, scimGroupsHandler(pages, &calls))
		group, err := findSCIMGroupByName(ctx, client, testSCIMGroupName)
		require.Error(t, err)
		assert.Nil(t, group)
		assert.Contains(t, err.Error(), "pagination did not advance")
		assert.Len(t, calls, 1)
	})
}
