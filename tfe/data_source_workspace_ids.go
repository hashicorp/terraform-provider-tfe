package tfe

import (
	"fmt"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTFEWorkspaceIDs() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTFEWorkspaceIDsRead,

		Schema: map[string]*schema.Schema{
			"names": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
				Required: false,
			},

			"tag_names": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
				Required: false,
			},

			"organization": {
				Type:     schema.TypeString,
				Required: true,
			},

			"ids": {
				Type:     schema.TypeMap,
				Computed: true,
			},

			"full_names": {
				Type:     schema.TypeMap,
				Computed: true,
			},
		},
	}
}

func dataSourceTFEWorkspaceIDsRead(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	// Get the organization.
	organization := d.Get("organization").(string)

	if len(d.Get("names").([]interface{})) == 0 && len(d.Get("tags").([]interface{})) == 0 {
		return fmt.Errorf("Either `names` or `tags` is required")
	}

	// Create a map with all the names we are looking for.
	var id string
	names := make(map[string]bool)
	for _, name := range d.Get("names").([]interface{}) {
		id += name.(string)
		names[name.(string)] = true
	}

	// Create two maps to hold the results.
	fullNames := make(map[string]string, len(names))
	ids := make(map[string]string, len(names))

	options := tfe.WorkspaceListOptions{}

	// Create a search string with all the tags we are looking for.
	var tagSearch string
	for _, tagName := range d.Get("tag_names").([]interface{}) {
		id += tagName.(string) // add to the state id
		tagSearch += fmt.Sprintf("%s,", tagName)
	}
	if tagSearch != "" {
		options.Tags = &tagSearch
	}

	for {
		wl, err := tfeClient.Workspaces.List(ctx, organization, options)
		if err != nil {
			return fmt.Errorf("Error retrieving workspaces: %v", err)
		}

		for _, w := range wl.Items {
			if names["*"] || names[w.Name] {
				fullNames[w.Name] = organization + "/" + w.Name
				ids[w.Name] = w.ID
			}
		}

		// Exit the loop when we've seen all pages.
		if wl.CurrentPage >= wl.TotalPages {
			break
		}

		// Update the page number to get the next page.
		options.PageNumber = wl.NextPage
	}

	d.Set("ids", ids)
	d.Set("full_names", fullNames)
	d.SetId(fmt.Sprintf("%s/%d", organization, schema.HashString(id)))

	return nil
}
