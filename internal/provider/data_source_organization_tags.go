// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

// NOTE: This is a legacy resource and should be migrated to the Plugin
// Framework if substantial modifications are planned. See
// docs/new-resources.md if planning to use this code as boilerplate for
// a new resource.

package provider

import (
	"fmt"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTFEOrganizationTags() *schema.Resource {
	return &schema.Resource{
		Description: "Gets information about the workspace tags for a given organization.",

		Read: dataSourceTFEOrganizationTagsRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The ID of the workspace tag.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"organization": {
				Description: "Name of the organization.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"tags": {
				Description: "A list of workspace tags within the organization.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Description: "The name of the workspace tag.",
							Type:        schema.TypeString,
							Computed:    true,
						},

						"id": {
							Description: "The ID of the workspace tag.",
							Type:        schema.TypeString,
							Computed:    true,
						},

						"workspace_count": {
							Description: "The number of workspaces the tag is associated with.",
							Type:        schema.TypeInt,
							Computed:    true,
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
