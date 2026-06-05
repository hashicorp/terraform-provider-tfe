// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

// NOTE: This is a legacy resource and should be migrated to the Plugin
// Framework if substantial modifications are planned. See
// docs/new-resources.md if planning to use this code as boilerplate for
// a new resource.

package provider

import (
	"fmt"
	"time"

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
			"scim_linked": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"scim_group_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"scim_sync_paused": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"scim_updated_at": {
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

		setTeamResourceData(d, tl.Items[0])
		return nil
	default:
		options := &tfe.TeamListOptions{}

		for {
			for _, team := range tl.Items {
				if team.Name == name {
					setTeamResourceData(d, team)
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

// setTeamResourceData populates state with the team's attributes. SCIM fields are
// guarded by nil checks so that older TFE instances do not panic.
func setTeamResourceData(d *schema.ResourceData, team *tfe.Team) {
	d.SetId(team.ID)
	d.Set("sso_team_id", team.SSOTeamID)

	if team.SCIMLinked != nil {
		d.Set("scim_linked", *team.SCIMLinked)
	}
	if team.SCIMGroupName != nil {
		d.Set("scim_group_name", *team.SCIMGroupName)
	}
	if team.SCIMSyncPaused != nil {
		d.Set("scim_sync_paused", *team.SCIMSyncPaused)
	}
	if team.SCIMUpdatedAt != nil {
		d.Set("scim_updated_at", team.SCIMUpdatedAt.Format(time.RFC3339))
	}
}
