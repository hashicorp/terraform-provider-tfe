// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"

	tfev2 "github.com/hashicorp/go-tfe/v2"
	"github.com/hashicorp/go-tfe/v2/api/admin"
	"github.com/hashicorp/go-tfe/v2/api/models"
)

// filterExactSCIMGroups returns the groups whose name matches the given name,
// case-insensitively. The List API only does substring matching, so we filter
// for exact matches here. Safe to call per page while paginating.
func filterExactSCIMGroups(groups []models.AdminScimGroupsable, name string) []models.AdminScimGroupsable {
	var matched []models.AdminScimGroupsable
	for _, g := range groups {
		if g == nil {
			continue
		}
		attrs := g.GetAttributes()
		if attrs == nil {
			continue
		}
		if strings.EqualFold(valueOrZero(attrs.GetName()), name) {
			matched = append(matched, g)
		}
	}
	return matched
}

// findSCIMGroupByName returns the SCIM group whose name matches exactly
// (case-insensitive), paging as needed. Returns (nil, nil) if none match, so
// the caller can craft its own "not found" message. ?q= only prefilters on the
// server; filterExactSCIMGroups does the real matching.
func findSCIMGroupByName(ctx context.Context, client *tfev2.Client, name string) (models.AdminScimGroupsable, error) {
	queryParams := &admin.ScimGroupsRequestBuilderGetQueryParameters{Q: &name}

	for {
		result, err := client.API.Admin().ScimGroups().Get(ctx, withQueryParams(queryParams))
		if err != nil {
			return nil, fmt.Errorf("unable to list SCIM groups: %w", err)
		}

		if matched := filterExactSCIMGroups(result.GetData(), name); len(matched) > 0 {
			return matched[0], nil
		}

		nextPage := nextPageFromMeta(result.GetMeta())
		if nextPage == nil {
			break
		}
		queryParams = &admin.ScimGroupsRequestBuilderGetQueryParameters{Q: &name, Pagenumber: nextPage}
	}

	return nil, nil
}
