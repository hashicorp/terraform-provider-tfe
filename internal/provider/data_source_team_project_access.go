// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

// NOTE: This is a legacy resource and should be migrated to the Plugin
// Framework if substantial modifications are planned. See
// docs/new-resources.md if planning to use this code as boilerplate for
// a new resource.

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTFETeamProjectAccess() *schema.Resource {
	return &schema.Resource{
		Description: "Gets information on team permissions on a project.",

		ReadContext: dataSourceTFETeamProjectAccessRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The team project access ID.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"access": {
				Description: "The type of access granted to the team on the project.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"team_id": {
				Description: "ID of the team.",
				Type:        schema.TypeString,
				Required:    true,
			},

			"project_id": {
				Description: "ID of the project.",
				Type:        schema.TypeString,
				Required:    true,
			},

			"project_access": {
				Description: "The permissions granted to the team on the project itself.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"settings": {
							Description: "The permission granted to the project's settings. Valid values are `read`, `update`, or `delete`.",
							Type:        schema.TypeString,
							Computed:    true,
						},

						"teams": {
							Description: "The permission granted to the project's teams. Valid values are `none`, `read`, or `manage`.",
							Type:        schema.TypeString,
							Computed:    true,
						},

						"variable_sets": {
							Description: "The permission granted to the project's variable sets. Valid values are `none`, `read`, or `write`.",
							Type:        schema.TypeString,
							Computed:    true,
						},
					},
				},
			},

			"workspace_access": {
				Description: "The permissions granted to the team across all workspaces in the project.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"create": {
							Description: "Whether the team can create workspaces in the project.",
							Type:        schema.TypeBool,
							Computed:    true,
						},

						"locking": {
							Description: "Whether the team can manually lock or unlock workspaces in the project.",
							Type:        schema.TypeBool,
							Computed:    true,
						},

						"move": {
							Description: "Whether the team can move workspaces into and out of the project.",
							Type:        schema.TypeBool,
							Computed:    true,
						},

						"delete": {
							Description: "Whether the team can delete workspaces in the project.",
							Type:        schema.TypeBool,
							Computed:    true,
						},

						"run_tasks": {
							Description: "Whether the team can manage run tasks in the project's workspaces.",
							Type:        schema.TypeBool,
							Computed:    true,
						},

						"policy_overrides": {
							Description: "This permission allows a team to override soft-mandatory policy evaluations, provided that team has been granted the org level 'delegate policy overrides' permission.",
							Type:        schema.TypeBool,
							Computed:    true,
						},

						"runs": {
							Description: "The permission granted to runs. Valid values are `read`, `plan`, or `apply`.",
							Type:        schema.TypeString,
							Computed:    true,
						},

						"sentinel_mocks": {
							Description: "The permission granted to Sentinel mocks. Valid values are `none` or `read`.",
							Type:        schema.TypeString,
							Computed:    true,
						},

						"state_versions": {
							Description: "The permission granted to state versions. Valid values are `none`, `read-outputs`, `read`, or `write`.",
							Type:        schema.TypeString,
							Computed:    true,
						},

						"variables": {
							Description: "The permission granted to variables. Valid values are `none`, `read`, or `write`.",
							Type:        schema.TypeString,
							Computed:    true,
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
