package tfe

import (
	"fmt"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTFESSHKey() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to get information about a SSH key.",
		Read:        dataSourceTFESSHKeyRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Name of the SSH key.",
				Type:        schema.TypeString,
				Required:    true,
			},

			"organization": {
				Description: "Name of the organization.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func dataSourceTFESSHKeyRead(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	// Get the name and organization.
	name := d.Get("name").(string)
	organization := d.Get("organization").(string)

	// Create an options struct.
	options := &tfe.SSHKeyListOptions{}

	for {
		l, err := tfeClient.SSHKeys.List(ctx, organization, options)
		if err != nil {
			return fmt.Errorf("Error retrieving SSH keys: %w", err)
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

	return fmt.Errorf("Could not find SSH key %s/%s", organization, name)
}
