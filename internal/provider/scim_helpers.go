// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"

	tfe "github.com/hashicorp/go-tfe"
)

// filterExactSCIMGroups returns the groups whose name matches the given name,
// case-insensitively. The List API only does substring matching, so we filter
// for exact matches here. Safe to call per page while paginating.
func filterExactSCIMGroups(groups []*tfe.AdminSCIMGroup, name string) []*tfe.AdminSCIMGroup {
	var matched []*tfe.AdminSCIMGroup
	for _, g := range groups {
		if g == nil {
			continue
		}
		if strings.EqualFold(g.Name, name) {
			matched = append(matched, g)
		}
	}
	return matched
}

// findSCIMGroupByName returns the SCIM group whose name matches exactly
// (case-insensitive), paging as needed. Returns (nil, nil) if none match, so
// the caller can craft its own "not found" message. ?q= only prefilters on the
// server; filterExactSCIMGroups does the real matching.
func findSCIMGroupByName(ctx context.Context, client *tfe.Client, name string) (*tfe.AdminSCIMGroup, error) {
	options := &tfe.AdminSCIMGroupListOptions{
		Query: name,
	}

	for {
		list, err := client.Admin.Settings.SCIM.Groups.List(ctx, options)
		if err != nil {
			return nil, fmt.Errorf("unable to list SCIM groups: %w", err)
		}

		if matched := filterExactSCIMGroups(list.Items, name); len(matched) > 0 {
			return matched[0], nil
		}

		if list.Pagination == nil || list.CurrentPage >= list.TotalPages {
			break
		}
		// Guard against a malformed response (NextPage not advancing past the
		// current page) that would otherwise re-fetch the same page forever.
		if list.NextPage <= list.CurrentPage {
			break
		}
		options.PageNumber = list.NextPage
	}

	return nil, nil
}
