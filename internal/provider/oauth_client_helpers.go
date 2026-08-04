// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/go-tfe/v2/api/organizations"
)

func fetchOAuthClientByNameOrServiceProvider(ctx context.Context, config ConfiguredClient, organization, name string, serviceProvider tfe.ServiceProviderType) (id string, err error) {
	pageSize := int32(100)
	queryParams := &organizations.ItemOauthClientsRequestBuilderGetQueryParameters{
		Pagesize: &pageSize,
	}

	var matched []string
	for {
		list, listErr := config.ClientV2.API.Organizations().ByOrganization_name(organization).OauthClients().Get(ctx, withQueryParams(queryParams))
		if listErr != nil {
			return "", fmt.Errorf("Error retrieving OAuth Clients: %w", listErr)
		}

		for _, item := range list.GetData() {
			attrs := item.GetAttributes()
			if attrs == nil {
				continue
			}
			iName := valueOrZero(attrs.GetName())
			iProvider := valueOrZero(attrs.GetServiceProvider())

			switch {
			case name != "" && serviceProvider != "":
				if iName == name && iProvider == string(serviceProvider) {
					matched = append(matched, valueOrZero(item.GetId()))
				}
			case name != "":
				if iName == name {
					matched = append(matched, valueOrZero(item.GetId()))
				}
			case serviceProvider != "":
				if iProvider == string(serviceProvider) {
					matched = append(matched, valueOrZero(item.GetId()))
				}
			}
		}

		nextPage := nextPageFromMeta(list.GetMeta())
		if nextPage == nil {
			break
		}
		queryParams.Pagenumber = nextPage
	}

	if len(matched) == 0 {
		return "", fmt.Errorf("no OAuthClients found matching the given parameters")
	}
	if len(matched) > 1 {
		return "", fmt.Errorf("too many OAuthClients were found to match the given parameters. Please narrow your search")
	}

	return matched[0], nil
}
