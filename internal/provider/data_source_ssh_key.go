// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

// NOTE: This is a legacy resource and should be migrated to the Plugin
// Framework if substantial modifications are planned. See
// docs/new-resources.md if planning to use this code as boilerplate for
// a new resource.

package provider

import (
	"fmt"

	"github.com/hashicorp/go-tfe/v2/api/organizations"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	abstractions "github.com/microsoft/kiota-abstractions-go"
)

func dataSourceTFESSHKey() *schema.Resource {
	return &schema.Resource{
		Description: "Get information on an SSH key.",

		Read: dataSourceTFESSHKeyRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The ID of the SSH key.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"name": {
				Description: "Name of the SSH key.",
				Type:        schema.TypeString,
				Required:    true,
			},

			"organization": {
				Description: "Name of the organization. If omitted, organization must be defined in the provider config.",
				Type:        schema.TypeString,
				Optional:    true,
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

	// Create the query parameters.
	pageSize := int32(100)
	queryParams := &organizations.ItemSshKeysRequestBuilderGetQueryParameters{
		Pagesize: &pageSize,
	}

	for {
		l, err := config.ClientV2.API.Organizations().ByOrganization_name(organization).SshKeys().Get(ctx, &abstractions.RequestConfiguration[organizations.ItemSshKeysRequestBuilderGetQueryParameters]{
			QueryParameters: queryParams,
		})
		if err != nil {
			return fmt.Errorf("Error retrieving SSH keys: %w", err)
		}

		for _, k := range l.GetData() {
			if attributes := k.GetAttributes(); attributes != nil && valueOrZero(attributes.GetName()) == name {
				d.SetId(valueOrZero(k.GetId()))
				return nil
			}
		}

		// Exit the loop when we've seen all pages.
		var nextPage *int32
		if meta := l.GetMeta(); meta != nil {
			nextPage = nextPageNumber(meta.GetPagination())
		}
		if nextPage == nil {
			break
		}

		// Update the page number to get the next page.
		queryParams.Pagenumber = nextPage
	}

	return fmt.Errorf("could not find SSH key %s/%s", organization, name)
}
