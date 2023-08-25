// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-tfe"
)

func fetchOAuthClientByNameOrServiceProvider(ctx context.Context, tfeClient *tfe.Client, organization, name string, serviceProvider tfe.ServiceProviderType) (*tfe.OAuthClient, error) {
	// Paginate through all OAuthClients in the organization; if multiple pages
	// of results are returned by the API, use the options variable to increment
	// the page number until all results have been retrieved.
	//
	// Within the pagination loop, loop again through each result on each page.
	// If 'name' was set, then match against the 'Name' field. If 'service_provider'
	// was set, then match against the 'ServiceProvider' field. If both are set,
	// then both must match. All matches are added to the ocMatches slice.
	//
	// At the end of the loop, if zero or more than one matches were found, an
	// error is returned. Otherwise, only one match was found, and that match is
	// returned.
	//
	var ocMatches []*tfe.OAuthClient
	options := &tfe.OAuthClientListOptions{}
	for {
		ocList, err := tfeClient.OAuthClients.List(ctx, organization, options)
		if err != nil {
			return nil, fmt.Errorf("Error retrieving OAuth Clients: %w", err)
		}

		for _, item := range ocList.Items {
			switch {
			case name != "" && serviceProvider != "":
				if item.Name != nil && *item.Name == name && item.ServiceProvider == serviceProvider {
					ocMatches = append(ocMatches, item)
				}
			case name != "":
				if item.Name != nil && *item.Name == name {
					ocMatches = append(ocMatches, item)
				}
			case serviceProvider != "":
				if item.ServiceProvider == serviceProvider {
					ocMatches = append(ocMatches, item)
				}
			}
		}

		// Exit the loop when we've seen all pages.
		if ocList.CurrentPage >= ocList.TotalPages {
			break
		}

		// Update the page number to get the next page.
		options.PageNumber = ocList.NextPage
	}
	if len(ocMatches) == 0 {
		return nil, fmt.Errorf("no OAuthClients found matching the given parameters")
	}
	if len(ocMatches) > 1 {
		return nil, fmt.Errorf("too many OAuthClients were found to match the given parameters. Please narrow your search")
	}

	return ocMatches[0], nil
}
