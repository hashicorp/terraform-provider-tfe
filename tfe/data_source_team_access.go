package tfe

import (
	"fmt"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceTFETeamAccess() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTFETeamAccessRead,

		Schema: map[string]*schema.Schema{
			"access": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"team_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"workspace_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceTFETeamAccessRead(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	// Get the team ID.
	teamID := d.Get("team_id").(string)

	// Get organization and workspace.
	organization, workspace, err := unpackWorkspaceID(d.Get("workspace_id").(string))
	if err != nil {
		return fmt.Errorf("Error unpacking workspace ID: %v", err)
	}

	// Get the workspace.
	ws, err := tfeClient.Workspaces.Read(ctx, organization, workspace)
	if err != nil {
		return fmt.Errorf(
			"Error retrieving workspace %s from organization %s: %v", workspace, organization, err)
	}

	// Create an options struct.
	options := tfe.TeamAccessListOptions{
		WorkspaceID: tfe.String(ws.ID),
	}

	for {
		l, err := tfeClient.TeamAccess.List(ctx, options)
		if err != nil {
			return fmt.Errorf("Error retrieving team access list: %v", err)
		}

		for _, ta := range l.Items {
			if ta.Team.ID == teamID {
				d.SetId(ta.ID)
				return resourceTFETeamAccessRead(d, meta)
			}
		}

		// Exit the loop when we've seen all pages.
		if l.CurrentPage >= l.TotalPages {
			break
		}

		// Update the page number to get the next page.
		options.PageNumber = l.NextPage
	}

	return fmt.Errorf("Could not find team access for %s and workspace %s", teamID, ws.Name)
}
