// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"log"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTFEOrganizations() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTFEOrganizationList,

		Schema: map[string]*schema.Schema{
			"names": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"ids": {
				Type:     schema.TypeMap,
				Computed: true,
			},

			"admin": {
				Type:     schema.TypeBool,
				Default:  false,
				Optional: true,
			},
		},
	}
}

func dataSourceTFEOrganizationList(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	var names []string
	var ids map[string]string
	var err error

	if isAdmin(d) {
		names, ids, err = adminOrgsPopulateFields(config.Client)
	} else {
		names, ids, err = orgsPopulateFields(config.Client)
	}

	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Setting Organizations Attributes")
	d.SetId("organizations")
	d.Set("names", names)
	d.Set("ids", ids)

	return nil
}

func adminOrgsPopulateFields(client *tfe.Client) ([]string, map[string]string, error) {
	names := []string{}
	ids := map[string]string{}
	log.Printf("[DEBUG] Listing all organizations (admin)")
	options := &tfe.AdminOrganizationListOptions{
		ListOptions: tfe.ListOptions{
			PageSize: 100,
		},
	}
	for {
		orgList, err := client.Admin.Organizations.List(ctx, options)
		if err != nil {
			return nil, nil, fmt.Errorf("Error retrieving Admin Organizations: %w", err)
		}

		for _, org := range orgList.Items {
			ids[org.Name] = org.ExternalID
			names = append(names, org.Name)
		}

		// Exit the loop when we've seen all pages.
		if orgList.CurrentPage >= orgList.TotalPages {
			break
		}

		// Update the page number to get the next page.
		options.PageNumber = orgList.NextPage
	}

	return names, ids, nil
}

func orgsPopulateFields(client *tfe.Client) ([]string, map[string]string, error) {
	names := []string{}
	ids := map[string]string{}
	log.Printf("[DEBUG] Listing all organizations (non-admin)")
	options := &tfe.OrganizationListOptions{
		ListOptions: tfe.ListOptions{
			PageSize: 100,
		},
	}
	for {
		orgList, err := client.Organizations.List(ctx, options)
		if err != nil {
			return nil, nil, fmt.Errorf("Error retrieving Organizations: %w", err)
		}

		for _, org := range orgList.Items {
			ids[org.Name] = org.ExternalID
			names = append(names, org.Name)
		}

		// Exit the loop when we've seen all pages.
		if orgList.CurrentPage >= orgList.TotalPages {
			break
		}

		// Update the page number to get the next page.
		options.PageNumber = orgList.NextPage
	}

	return names, ids, nil
}

func isAdmin(d *schema.ResourceData) bool {
	return d.Get("admin").(bool)
}
