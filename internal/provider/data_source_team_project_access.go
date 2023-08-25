// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

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

			"project_access": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"settings": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"teams": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"workspace_access": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"create": {
							Type:     schema.TypeBool,
							Computed: true,
						},

						"locking": {
							Type:     schema.TypeBool,
							Computed: true,
						},

						"move": {
							Type:     schema.TypeBool,
							Computed: true,
						},

						"delete": {
							Type:     schema.TypeBool,
							Computed: true,
						},

						"run_tasks": {
							Type:     schema.TypeBool,
							Computed: true,
						},

						"runs": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"sentinel_mocks": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"state_versions": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"variables": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
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
