package tfe

import (
	"fmt"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTFEAgentPool() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTFEAgentPoolRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"organization": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceTFEAgentPoolRead(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	// Get the name and organization.
	name := d.Get("name").(string)
	organization := d.Get("organization").(string)

	// Create an options struct.
	// to reduce the number of pages returned, search based on the name. TFE instances which
	// do not support agent pool seach will just ignore the query parameter
	options := tfe.AgentPoolListOptions{
		Query: name,
	}

	for {
		l, err := tfeClient.AgentPools.List(ctx, organization, &options)
		if err != nil {
			return fmt.Errorf("Error retrieving agent pools: %w", err)
		}

		for _, k := range l.Items {
			if k.Name == name {
				d.SetId(k.ID)
				return nil
			}
		}

		// Exit the loop when we've seen all pages.
		if l.CurrentPage >= l.TotalPages {
			break
		}

		// Update the page number to get the next page.
		options.PageNumber = l.NextPage
	}

	return fmt.Errorf("Could not find agent pool %s/%s", organization, name)
}
