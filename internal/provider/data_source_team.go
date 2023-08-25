// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTFETeam() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTFETeamRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"organization": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"sso_team_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceTFETeamRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	// Get the name and organization.
	name := d.Get("name").(string)
	organization, err := config.schemaOrDefaultOrganization(d)
	if err != nil {
		return err
	}

	tl, err := config.Client.Teams.List(ctx, organization, &tfe.TeamListOptions{
		Names: []string{name},
	})
	if err != nil {
		return fmt.Errorf("Error retrieving teams: %w", err)
	}

	switch len(tl.Items) {
	case 0:
		return fmt.Errorf("could not find team %s/%s", organization, name)
	case 1:
		// We check this just in case a user's TFE instance only has one team
		// and doesn't support the filter query param
		if tl.Items[0].Name != name {
			return fmt.Errorf("could not find team %s/%s", organization, name)
		}

		d.SetId(tl.Items[0].ID)
		d.Set("sso_team_id", tl.Items[0].SSOTeamID)

		return nil
	default:
		options := &tfe.TeamListOptions{}

		for {
			for _, team := range tl.Items {
				if team.Name == name {
					d.SetId(team.ID)
					d.Set("sso_team_id", team.SSOTeamID)
					return nil
				}
			}

			if tl.CurrentPage >= tl.TotalPages {
				break
			}

			options.PageNumber = tl.NextPage

			tl, err = config.Client.Teams.List(ctx, organization, options)
			if err != nil {
				return fmt.Errorf("Error retrieving teams: %w", err)
			}
		}
	}

	return fmt.Errorf("could not find team %s/%s", organization, name)
}
