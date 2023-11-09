// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-tfe"
)

func fetchTeamByName(ctx context.Context, client *tfe.Client, orgName string, teamName string) (*tfe.Team, error) {
	listOptions := &tfe.TeamListOptions{
		Names: []string{teamName},
	}
	teamList, err := client.Teams.List(ctx, orgName, listOptions)

	for {
		if err != nil {
			return nil, fmt.Errorf("failed to list teams: %w", err)
		}

		for _, team := range teamList.Items {
			if team.Name == teamName {
				return team, nil
			}
		}

		if teamList.CurrentPage >= teamList.TotalPages {
			break
		}

		listOptions.PageNumber = teamList.NextPage

		teamList, err = client.Teams.List(ctx, orgName, listOptions)
	}
	return nil, tfe.ErrResourceNotFound
}
