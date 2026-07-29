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
)

func dataSourceTFETeams() *schema.Resource {
	return &schema.Resource{
		Description: "Gets information on teams.",

		Read: dataSourceTFETeamsRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "Name of the organization.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"organization": {
				Description: "Name of the organization.",
				Type:        schema.TypeString,
				Optional:    true,
			},

			"names": {
				Description: "A list of team names in an organization.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},

			"ids": {
				Description: "A map of team names in an organization and their IDs.",
				Type:        schema.TypeMap,
				Computed:    true,
			},
		},
	}
}

func dataSourceTFETeamsRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)
	organization, err := config.schemaOrDefaultOrganization(d)
	if err != nil {
		return err
	}

	teamsBuilder := config.ClientV2.API.Organizations().ByOrganization_name(organization).Teams()

	pageSize := int32(100)
	queryParams := &organizations.ItemTeamsRequestBuilderGetQueryParameters{
		Pagesize: &pageSize,
	}

	result, err := teamsBuilder.Get(ctx, withQueryParams(queryParams))
	if err != nil {
		return fmt.Errorf("Error retrieving teams: %w", err)
	}

	items := result.GetData()
	if len(items) == 0 {
		return fmt.Errorf("could not find teams in %q", organization)
	}

	names := []string{}
	ids := map[string]string{}
	for {
		for _, team := range items {
			attrs := team.GetAttributes()
			if attrs == nil {
				continue
			}
			name := valueOrZero(attrs.GetName())
			names = append(names, name)
			ids[name] = valueOrZero(team.GetId())
		}

		nextPage := nextPageFromMeta(result.GetMeta())
		if nextPage == nil {
			break
		}

		queryParams = &organizations.ItemTeamsRequestBuilderGetQueryParameters{
			Pagesize:   &pageSize,
			Pagenumber: nextPage,
		}
		result, err = teamsBuilder.Get(ctx, withQueryParams(queryParams))
		if err != nil {
			return fmt.Errorf("Error retrieving teams: %w", err)
		}
		items = result.GetData()
	}

	d.SetId(organization)
	d.Set("names", names)
	d.Set("ids", ids)

	return nil
}
