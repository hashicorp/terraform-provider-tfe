// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-tfe"
)

func fetchGithubAppInstallationByNameOrGHID(ctx context.Context, tfeClient *tfe.Client, name string, installationID int) (*tfe.GHAInstallation, error) {
	// Paginate through all GithubAppInstallation; if multiple pages
	// of results are returned by the API, use the options variable to increment
	// the page number until all results have been retrieved.
	//
	// Within the pagination loop, loop again through each result on each page.
	// If 'name' was set, then match against the 'Name' field. If 'installation_id'
	// was set, then match against the 'installation_id' field. If both are set,
	// then both must match. All matches are added to the ocMatches slice.
	//
	// At the end of the loop, if zero or more than one matches were found, an
	// error is returned. Otherwise, only one match was found, and that match is
	// returned.
	//
	if name == "" && installationID == 0 {
		return nil, fmt.Errorf("invalid parameters, either name or installation id must have a value")
	}
	var ghaInstallation *tfe.GHAInstallation
	options := &tfe.GHAInstallationListOptions{}
	for {
		ghaInstList, err := tfeClient.GHAInstallations.List(ctx, options)
		if err != nil {
			return nil, fmt.Errorf("error retrieving Github App Installations: %w", err)
		}
		for _, item := range ghaInstList.Items {
			switch {
			case name != "" && installationID != 0:
				if item.Name != nil && *item.Name == name && item.InstallationID != nil && *item.InstallationID == installationID {
					ghaInstallation = item
				}
			case name != "":
				if item.Name != nil && *item.Name == name {
					ghaInstallation = item
				}
			case installationID != 0:
				if *item.InstallationID == installationID {
					ghaInstallation = item
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
	if ghaInstallation == nil {
		return nil, fmt.Errorf("no Github App Installation found matching the given parameters")
	}
	return ghaInstallation, nil
}
