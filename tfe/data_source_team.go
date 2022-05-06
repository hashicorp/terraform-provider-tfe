package tfe

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
				Required: true,
			},
			"sso_team_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceTFETeamRead(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	// Get the name and organization.
	name := d.Get("name").(string)
	organization := d.Get("organization").(string)

	tl, err := tfeClient.Teams.List(ctx, organization, &tfe.TeamListOptions{
		Names: []string{name},
	})
	if err != nil {
		return fmt.Errorf("Error retrieving teams: %v", err)
	}

	switch len(tl.Items) {
	case 0:
		return fmt.Errorf("Could not find team %s/%s", organization, name)
	case 1:
		d.SetId(tl.Items[0].ID)
		d.Set("sso_team_id", tl.Items[0].SSOTeamID)
		return nil
	default:
		options := &tfe.TeamListOptions{}

		for {
			for _, team := range tl.Items {
				if team.Name == name {
					d.SetId(tl.Items[0].ID)
					d.Set("sso_team_id", tl.Items[0].SSOTeamID)
					return nil
				}
			}

			if tl.CurrentPage >= tl.TotalPages {
				break
			}

			options.PageNumber = tl.NextPage

			tl, err = tfeClient.Teams.List(ctx, organization, options)
			if err != nil {
				return fmt.Errorf("Error retrieving teams: %v", err)
			}
		}
	}

	return fmt.Errorf("Could not find team %s/%s", organization, name)
}
