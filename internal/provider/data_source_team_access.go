// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

// NOTE: This is a legacy resource and should be migrated to the Plugin
// Framework if substantial modifications are planned. See
// docs/new-resources.md if planning to use this code as boilerplate for
// a new resource.

package provider

import (
	"fmt"

	"github.com/hashicorp/go-tfe/v2/api/teamworkspaces"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTFETeamAccess() *schema.Resource {
	return &schema.Resource{
		Description: "Gets information on team permissions on a workspace.",

		Read: dataSourceTFETeamAccessRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The team access ID.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"access": {
				Description: "The type of access granted to the team on the workspace.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"permissions": {
				Description: "The custom permissions granted to the team on the workspace.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"runs": {
							Description: "The permission granted to runs. Valid values are `read`, `plan`, or `apply`.",
							Type:        schema.TypeString,
							Computed:    true,
						},

						"variables": {
							Description: "The permission granted to variables. Valid values are `none`, `read`, or `write`.",
							Type:        schema.TypeString,
							Computed:    true,
						},

						"state_versions": {
							Description: "The permission granted to state versions. Valid values are `none`, `read-outputs`, `read`, or `write`.",
							Type:        schema.TypeString,
							Computed:    true,
						},

						"sentinel_mocks": {
							Description: "The permission granted to Sentinel mocks. Valid values are `none` or `read`.",
							Type:        schema.TypeString,
							Computed:    true,
						},

						"workspace_locking": {
							Description: "Whether the team can manually lock or unlock the workspace.",
							Type:        schema.TypeBool,
							Computed:    true,
						},

						"run_tasks": {
							Description: "Whether the team can manage workspace run tasks.",
							Type:        schema.TypeBool,
							Computed:    true,
						},

						"policy_overrides": {
							Description: "This permission allows a team to override soft-mandatory policy evaluations, provided that team has been granted the org level 'delegate policy overrides' permission.",
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
	config := meta.(ConfiguredClient)

	// Get the team ID.
	teamID := d.Get("team_id").(string)

	// Get the workspace
	workspaceID := d.Get("workspace_id").(string)
	ws, err := config.ClientV2.API.Workspaces().ByWorkspace_id(workspaceID).Get(ctx, nil)
	if err != nil {
		return fmt.Errorf(
			"Error retrieving workspace %s: %w", workspaceID, err)
	}
	if ws == nil || ws.GetData() == nil {
		return fmt.Errorf("Error retrieving workspace %s: no data returned", workspaceID)
	}

	// Filter directly by workspace and team, which uniquely identify at
	// most one team-workspace access relationship.
	teamWorkspacesBuilder := config.ClientV2.API.TeamWorkspaces()
	queryParams := &teamworkspaces.TeamWorkspacesRequestBuilderGetQueryParameters{
		Filterworkspaceid: &workspaceID,
		Filterteamid:      &teamID,
	}

	result, err := teamWorkspacesBuilder.Get(ctx, withQueryParams(queryParams))
	if err != nil {
		return fmt.Errorf("Error retrieving team access list: %w", err)
	}

	items := result.GetData()
	for {
		for _, ta := range items {
			relationships := ta.GetRelationships()
			if relationships == nil || relationships.GetTeam() == nil || relationships.GetTeam().GetData() == nil {
				continue
			}
			if valueOrZero(relationships.GetTeam().GetData().GetId()) == teamID {
				d.SetId(valueOrZero(ta.GetId()))
				return resourceTFETeamAccessRead(d, meta)
			}
		}

		nextPage := nextPageFromMeta(result.GetMeta())
		if nextPage == nil {
			break
		}

		queryParams = &teamworkspaces.TeamWorkspacesRequestBuilderGetQueryParameters{
			Filterworkspaceid: &workspaceID,
			Filterteamid:      &teamID,
			Pagenumber:        nextPage,
		}
		result, err = teamWorkspacesBuilder.Get(ctx, withQueryParams(queryParams))
		if err != nil {
			return fmt.Errorf("Error retrieving team access list: %w", err)
		}
		items = result.GetData()
	}

	return fmt.Errorf("could not find team access for %s and workspace %s", teamID, valueOrZero(ws.GetData().GetAttributes().GetName()))
}
