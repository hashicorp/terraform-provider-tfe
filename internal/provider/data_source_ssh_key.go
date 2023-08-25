// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTFESSHKey() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTFESSHKeyRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"organization": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func dataSourceTFESSHKeyRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	// Get the name and organization.
	name := d.Get("name").(string)
	organization, err := config.schemaOrDefaultOrganization(d)
	if err != nil {
		return err
	}

	// Create an options struct.
	options := &tfe.SSHKeyListOptions{}

	for {
		l, err := config.Client.SSHKeys.List(ctx, organization, options)
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

	return fmt.Errorf("could not find SSH key %s/%s", organization, name)
}
