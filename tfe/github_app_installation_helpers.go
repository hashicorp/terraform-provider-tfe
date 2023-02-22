// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tfe

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-tfe"
)

func fetchGithubAppInstallationByNameOrGHID(ctx context.Context, tfeClient *tfe.Client, name string, GHInstallationIDS int) (*tfe.GHAInstallation, error) {
	// Paginate through all GithubAppInstallation; if multiple pages
	// of results are returned by the API, use the options variable to increment
	// the page number until all results have been retrieved.
	//
	// Within the pagination loop, loop again through each result on each page.
	// If 'name' was set, then match against the 'Name' field. If 'github_installation_id'
	// was set, then match against the 'github_installation_id' field. If both are set,
	// then both must match. All matches are added to the ocMatches slice.
	//
	// At the end of the loop, if zero or more than one matches were found, an
	// error is returned. Otherwise, only one match was found, and that match is
	// returned.
	//
	var ghainsMatches []*tfe.GHAInstallation
	options := &tfe.GHAInstallationListOptions{}
	for {
		ghaInstList, err := tfeClient.GHAInstallations.List(ctx, options)
		if err != nil {
			return nil, fmt.Errorf("Error retrieving Github App Installations: %w", err)
		}

		for _, item := range ghaInstList.Items {
			switch {
			case name != "" && GHInstallationIDS != 0:
				if item.Name != "" && item.Name == name && item.GHInstallationId == GHInstallationIDS {
					ghainsMatches = append(ghainsMatches, item)
				}
			case name != "":
				if item.Name != "" && item.Name == name {
					ghainsMatches = append(ghainsMatches, item)
				}
			case GHInstallationIDS != 0:
				if item.GHInstallationId == GHInstallationIDS {
					ghainsMatches = append(ghainsMatches, item)
				}
			}
		}

		// Exit the loop when we've seen all pages.
		if ghaInstList.CurrentPage >= ghaInstList.TotalPages {
			break
		}

		// Update the page number to get the next page.
		options.PageNumber = ghaInstList.NextPage
	}
	if len(ghainsMatches) == 0 {
		return nil, fmt.Errorf("no Github App Installation found matching the given parameters")
	}
	if len(ghainsMatches) > 1 {
		return nil, fmt.Errorf("too many Github App Installation were found to match the given parameters. Please narrow your search")
	}

	return ghainsMatches[0], nil
}
