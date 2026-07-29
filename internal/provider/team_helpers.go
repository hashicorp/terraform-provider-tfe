// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	tfev2 "github.com/hashicorp/go-tfe/v2"
	v2api "github.com/hashicorp/go-tfe/v2/api"
	"github.com/hashicorp/go-tfe/v2/api/models"
	"github.com/hashicorp/go-tfe/v2/api/organizations"
)

// fetchTeamByNameV2 finds a team with an exact-match name in an organization
// using the go-tfe v2 generated client, following pagination until it is
// found or the pages are exhausted. This mirrors go-tfe v1's
// Teams.List(ctx, orgName, &TeamListOptions{Names: []string{teamName}})
// exact-match semantics: the "filter[names]" query parameter is a
// server-side hint that some TFE releases may not support, so results are
// always verified locally.
func fetchTeamByNameV2(ctx context.Context, api *v2api.ApiClient, orgName string, teamName string) (models.Teamsable, error) {
	teamsBuilder := api.Organizations().ByOrganization_name(orgName).Teams()

	filterNames := teamName
	queryParams := &organizations.ItemTeamsRequestBuilderGetQueryParameters{
		Filternames: &filterNames,
	}

	result, err := teamsBuilder.Get(ctx, withQueryParams(queryParams))
	if err != nil {
		return nil, fmt.Errorf("failed to list teams: %w", err)
	}

	for {
		for _, team := range result.GetData() {
			if attrs := team.GetAttributes(); attrs != nil && valueOrZero(attrs.GetName()) == teamName {
				return team, nil
			}
		}

		nextPage := nextPageFromMeta(result.GetMeta())
		if nextPage == nil {
			break
		}

		pagedParams := &organizations.ItemTeamsRequestBuilderGetQueryParameters{
			Filternames: &filterNames,
			Pagenumber:  nextPage,
		}
		result, err = teamsBuilder.Get(ctx, withQueryParams(pagedParams))
		if err != nil {
			return nil, fmt.Errorf("failed to list teams: %w", err)
		}
	}

	return nil, tfev2.ErrNotFound
}
