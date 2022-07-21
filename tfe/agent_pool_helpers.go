package tfe

import (
	"fmt"

	tfe "github.com/hashicorp/go-tfe"
)

func fetchtAgentPoolID(orgName string, poolName string, client *tfe.Client) (string, error) {
	// to reduce the number of pages returned, search based on the name. TFE instances which
	// do not support agent pool search will just ignore the query parameter
	options := tfe.AgentPoolListOptions{
		Query: poolName,
	}

	for {
		l, err := client.AgentPools.List(ctx, orgName, &options)
		if err != nil {
			return "", fmt.Errorf("Error retrieving agent pools: %w", err)
		}

		for _, k := range l.Items {
			if k.Name == poolName {
				return k.ID, nil
			}
		}

		// Exit the loop when we've seen all pages.
		if l.CurrentPage >= l.TotalPages {
			break
		}

		// Update the page number to get the next page.
		options.PageNumber = l.NextPage
	}

	return "", fmt.Errorf("Could not find agent pool %s/%s", orgName, poolName)
}
