package tfe

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTFETeamProjectAccess() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceTFETeamProjectAccessRead,

		Schema: map[string]*schema.Schema{
			"access": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"team_id": {
				Type:     schema.TypeString,
				Required: true,
			},

			"project_id": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceTFETeamProjectAccessRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ConfiguredClient)

	// Get the team ID.
	teamID := d.Get("team_id").(string)

	// Get the project
	projectID := d.Get("project_id").(string)
	proj, err := config.Client.Projects.Read(ctx, projectID)
	if err != nil {
		return diag.Errorf(
			"Error retrieving project %s: %v", projectID, err)
	}

	options := tfe.TeamProjectAccessListOptions{
		ProjectID: proj.ID,
	}

	for {
		l, err := config.Client.TeamProjectAccess.List(ctx, options)
		if err != nil {
			return diag.Errorf("Error retrieving team access list: %v", err)
		}

		for _, ta := range l.Items {
			if ta.Team.ID == teamID {
				d.SetId(ta.ID)
				return resourceTFETeamProjectAccessRead(ctx, d, meta)
			}
		}

		// Exit the loop when we've seen all pages.
		if l.CurrentPage >= l.TotalPages {
			break
		}

		// Update the page number to get the next page.
		options.PageNumber = l.NextPage
	}

	return diag.Errorf("could not find team project access for %s and project %s", teamID, proj.Name)
}
