// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"

	tfev2 "github.com/hashicorp/go-tfe/v2"
	"github.com/hashicorp/go-tfe/v2/api/models"
	"github.com/hashicorp/go-tfe/v2/api/organizations"
)

func fetchAgentPool(orgName string, poolName string, client *tfev2.Client) (models.AgentPoolsable, error) {
	// to reduce the number of pages returned, search based on the name. TFE instances which
	// do not support agent pool search will just ignore the query parameter
	pageSize := int32(100)
	queryParams := &organizations.ItemAgentPoolsRequestBuilderGetQueryParameters{
		Q:        &poolName,
		Pagesize: &pageSize,
	}

	for {
		response, err := client.API.Organizations().ByOrganization_name(orgName).AgentPools().Get(ctx, withQueryParams(queryParams))
		if err != nil {
			return nil, fmt.Errorf("Error retrieving agent pools: %w", err)
		}

		for _, pool := range response.GetData() {
			if pool == nil {
				continue
			}
			attrs := pool.GetAttributes()
			if attrs != nil && valueOrZero(attrs.GetName()) == poolName {
				return pool, nil
			}
		}

		// Exit the loop when we've seen all pages.
		nextPage := nextPageFromMeta(response.GetMeta())
		if nextPage == nil {
			break
		}

		// Update the page number to get the next page.
		queryParams.Pagenumber = nextPage
	}

	return nil, tfev2.ErrNotFound
}
