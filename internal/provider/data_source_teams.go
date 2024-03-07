// Copyright (c) HashiCorp, Inc.
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

func dataSourceTFETeams() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTFETeamsRead,

		Schema: map[string]*schema.Schema{
			"organization": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"names": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"ids": {
				Type:     schema.TypeMap,
				Computed: true,
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

	teams, err := config.Client.Teams.List(ctx, organization, &tfe.TeamListOptions{})
	if err != nil {
		return fmt.Errorf("Error retrieving teams: %w", err)
	}

	if len(teams.Items) == 0 {
		return fmt.Errorf("could not find teams in %q", organization)
	}

	options := &tfe.TeamListOptions{}
	names := []string{}
	ids := map[string]string{}
	for {
		for _, team := range teams.Items {
			names = append(names, team.Name)
			ids[team.Name] = team.ID
		}

		if teams.CurrentPage >= teams.TotalPages {
			break
		}

		options.PageNumber = teams.NextPage

		teams, err = config.Client.Teams.List(ctx, organization, options)
		if err != nil {
			return fmt.Errorf("Error retrieving teams: %w", err)
		}
	}

	d.SetId(organization)
	d.Set("names", names)
	d.Set("ids", ids)

	return nil
}
