package tfe

import (
	"fmt"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTFETeamAccess() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to get information about team permissions for a workspace.",
		Read:        dataSourceTFETeamAccessRead,

		Schema: map[string]*schema.Schema{
			"access": {
				Description: "The type of access granted to the team on the workspace.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"permissions": {
				Description: "The permissions granted to the team on the workspaces for each whatever.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"runs": {
							Description: "The permission granted to runs. Valid values are `read`, `plan`, or `apply`",
							Type:        schema.TypeString,
							Computed:    true,
						},

						"variables": {
							Description: "The permissions granted to variables. Valid values are `none`, `read`, or `write`",
							Type:        schema.TypeString,
							Computed:    true,
						},

						"state_versions": {
							Description: "The permissions granted to state versions. Valid values are `none`, `read-outputs`, `read`, or `write`",
							Type:        schema.TypeString,
							Computed:    true,
						},

						"sentinel_mocks": {
							Description: "The permissions granted to Sentinel mocks. Valid values are `none` or `read`",
							Type:        schema.TypeString,
							Computed:    true,
						},

						"workspace_locking": {
							Description: "Whether permission is granted to manually lock the workspace or not.",
							Type:        schema.TypeBool,
							Computed:    true,
						},

						"run_tasks": {
							Description: "Boolean determining whether or not to grant the team permission to manage workspace run tasks.",
							Type:        schema.TypeBool,
							Computed:    true,
						},
					},
				},
			},

			"team_id": {
				Description: "ID of the team.",
				Type:        schema.TypeString,
				Required:    true,
			},

			"workspace_id": {
				Description: "ID of the workspace.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func dataSourceTFETeamAccessRead(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	// Get the team ID.
	teamID := d.Get("team_id").(string)

	// Get the workspace
	workspaceID := d.Get("workspace_id").(string)
	ws, err := tfeClient.Workspaces.ReadByID(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf(
			"Error retrieving workspace %s: %w", workspaceID, err)
	}

	// Create an options struct.
	options := &tfe.TeamAccessListOptions{
		WorkspaceID: ws.ID,
	}

	for {
		l, err := tfeClient.TeamAccess.List(ctx, options)
		if err != nil {
			return fmt.Errorf("Error retrieving team access list: %w", err)
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
