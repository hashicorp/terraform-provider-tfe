package tfe

import (
	"fmt"
	"strconv"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTFEOrganizationWorkspaceTags() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTFEOrganizationWorkspaceTagsRead,

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

						"instance_count": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceTFEOrganizationWorkspaceTagsRead(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(ConfiguredClient)

	organizationName, err := tfeClient.schemaOrDefaultOrganization(d)
	if err != nil {
		return err
	}

	var tags []map[string]string

	options := tfe.OrganizationTagsListOptions{}
	for {
		organizationTagsList, err := tfeClient.Client.OrganizationTags.List(ctx, organizationName, &options)
		if err != nil {
			return fmt.Errorf("Error retrieving organization workspace tags: %w", err)
		}

		for _, orgTag := range organizationTagsList.Items {
			tag := map[string]string{"id": orgTag.ID, "name": orgTag.Name, "instance_count": strconv.Itoa(orgTag.InstanceCount)}
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
