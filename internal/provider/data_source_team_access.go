// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTFETeamAccess() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTFETeamAccessRead,

		Schema: map[string]*schema.Schema{
			"access": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"permissions": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"runs": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"variables": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"state_versions": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"sentinel_mocks": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"workspace_locking": {
							Type:     schema.TypeBool,
							Computed: true,
						},

						"run_tasks": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},

			"team_id": {
				Type:     schema.TypeString,
				Required: true,
			},

			"workspace_id": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceTFETeamAccessRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	// Get the team ID.
	teamID := d.Get("team_id").(string)

	// Get the workspace
	workspaceID := d.Get("workspace_id").(string)
	ws, err := config.Client.Workspaces.ReadByID(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf(
			"Error retrieving workspace %s: %w", workspaceID, err)
	}

	// Create an options struct.
	options := &tfe.TeamAccessListOptions{
		WorkspaceID: ws.ID,
	}

	for {
		l, err := config.Client.TeamAccess.List(ctx, options)
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

	return fmt.Errorf("could not find team access for %s and workspace %s", teamID, ws.Name)
}
