// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

// NOTE: This is a legacy resource and should be migrated to the Plugin
// Framework if substantial modifications are planned. See
// docs/new-resources.md if planning to use this code as boilerplate for
// a new resource.

package provider

import (
	"fmt"
	"log"

	tfe "github.com/hashicorp/go-tfe"
	tfev2 "github.com/hashicorp/go-tfe/v2"
	"github.com/hashicorp/go-tfe/v2/api/organizations"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTFEOrganizations() *schema.Resource {
	return &schema.Resource{
		Description: "Gets a list of organizations and a map of their IDs.",

		Read: dataSourceTFEOrganizationList,

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "Static identifier for this data source. Do not rely on this value.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"names": {
				Description: "A list of names of every organization.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},

			"ids": {
				Description: "A map of organization names and their IDs.",
				Type:        schema.TypeMap,
				Computed:    true,
			},

			"admin": {
				Description: "(Available only in Terraform Enterprise) Determines the list of organizations that should be retrieved. If it is true, then it will retrieve all the organizations for the entire installation. If it is false, then it will retrieve the organizations available as per permissions of the API Token.",
				Type:        schema.TypeBool,
				Default:     false,
				Optional:    true,
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
		names, ids, err = orgsPopulateFields(config.ClientV2)
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

// adminOrgsPopulateFields stays on the v1 client: the pinned go-tfe/v2
// feature-branch client's admin Organizations request builder only exposes
// ByOrganization_name (single-org get/patch), with no Get method for the
// list endpoint (go-tfe/v2 gap).
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

func orgsPopulateFields(client *tfev2.Client) ([]string, map[string]string, error) {
	names := []string{}
	ids := map[string]string{}
	log.Printf("[DEBUG] Listing all organizations (non-admin)")
	pageSize := int32(100)
	queryParams := &organizations.OrganizationsRequestBuilderGetQueryParameters{
		Pagesize: &pageSize,
	}
	for {
		orgList, err := client.API.Organizations().Get(ctx, withQueryParams(queryParams))
		if err != nil {
			return nil, nil, fmt.Errorf("Error retrieving Organizations: %w", err)
		}

		for _, org := range orgList.GetData() {
			name := valueOrZero(org.GetId())
			externalID := ""
			if attrs := org.GetAttributes(); attrs != nil {
				externalID = valueOrZero(attrs.GetExternalId())
			}
			ids[name] = externalID
			names = append(names, name)
		}

		// Exit the loop when we've seen all pages.
		nextPage := nextPageFromMeta(orgList.GetMeta())
		if nextPage == nil {
			break
		}

		// Update the page number to get the next page.
		queryParams = &organizations.OrganizationsRequestBuilderGetQueryParameters{
			Pagesize:   &pageSize,
			Pagenumber: nextPage,
		}
	}

	return names, ids, nil
}

func isAdmin(d *schema.ResourceData) bool {
	return d.Get("admin").(bool)
}
