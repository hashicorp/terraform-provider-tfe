// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTFEOrganizationTags() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTFEOrganizationTagsRead,

		Schema: map[string]*schema.Schema{
			"organization": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"tags": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"workspace_count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceTFEOrganizationTagsRead(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(ConfiguredClient)

	organizationName, err := tfeClient.schemaOrDefaultOrganization(d)
	if err != nil {
		return err
	}

	var tags []map[string]interface{}

	options := tfe.OrganizationTagsListOptions{}
	for {
		organizationTagsList, err := tfeClient.Client.OrganizationTags.List(ctx, organizationName, &options)
		if err != nil {
			return fmt.Errorf("Error retrieving organization tags: %w", err)
		}

		for _, orgTag := range organizationTagsList.Items {
			tag := map[string]interface{}{
				"id":              orgTag.ID,
				"name":            orgTag.Name,
				"workspace_count": orgTag.InstanceCount,
			}
			tags = append(tags, tag)
		}

		// Exit the loop when we've seen all pages.
		if organizationTagsList.CurrentPage >= organizationTagsList.TotalPages {
			break
		}

		// Update the page number to get the next page.
		options.PageNumber = organizationTagsList.NextPage
	}

	d.Set("tags", tags)
	d.SetId(organizationName)

	return nil
}
