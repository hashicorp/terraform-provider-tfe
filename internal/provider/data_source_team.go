// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

// NOTE: This is a legacy resource and should be migrated to the Plugin
// Framework if substantial modifications are planned. See
// docs/new-resources.md if planning to use this code as boilerplate for
// a new resource.

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
			"organization_access": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"manage_policies": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Allow members to create, edit, read, list and delete the organization's policies.",
						},
						"manage_policy_overrides": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Allow members to override soft-mandatory policy checks.",
						},
						"delegate_policy_overrides": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "When this setting is enabled for a team, its members can override failed policy evaluations on projects and workspaces they manage.",
						},
						"manage_workspaces": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Grants members the ability to view, edit, delete, and assign team access to all workspaces in this organization, as well as the ability to create new workspaces in the default project.",
						},
						"manage_vcs_settings": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Allow members to manage the organization's VCS providers and SSH keys.",
						},
						"manage_providers": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Allow members to publish and delete providers in the organization's private registry.",
						},
						"manage_modules": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Allow members to publish and delete modules in the organization's private registry.",
						},
						"manage_run_tasks": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Allow members to create, update, and delete run tasks on an organization.",
						},
						"manage_projects": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Grants members the ability to view, edit, delete, and assign team access to all projects in this organization, as well as the ability to create new workspaces in any project.",
						},
						"read_workspaces": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Allow this team to view all workspaces in this organization.",
						},
						"read_projects": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Allow this team to view all projects in this organization.",
						},
						"manage_membership": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Allow members to add and remove users from the organization, and to manage the membership of teams. This permission allows members to assign themselves to other teams.",
						},
						"manage_teams": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Grant members the ability to manage membership, as well as to create and delete teams and team tokens. This permission allows members to manage all teams, including those that they are not a part of.",
						},
						"manage_organization_access": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Grant members the ability to manage team memberships, permissions, and organization access.",
						},
						"access_secret_teams": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Allow members to access secret teams. Members will be able to view all secret teams and potentially manage them depending on their level of team permissions.",
						},
						"manage_agent_pools": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Allow members to create, update, and delete the organization's agent pools.",
						},
					},
				},
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
		if err := d.Set("organization_access", flattenTeamOrganizationAccess(tl.Items[0].OrganizationAccess)); err != nil {
			return fmt.Errorf("error setting organization access for team %s: %w", tl.Items[0].ID, err)
		}
		d.Set("sso_team_id", tl.Items[0].SSOTeamID)

		return nil
	default:
		options := &tfe.TeamListOptions{}

		for {
			for _, team := range tl.Items {
				if team.Name == name {
					d.SetId(team.ID)
					if err := d.Set("organization_access", flattenTeamOrganizationAccess(team.OrganizationAccess)); err != nil {
						return fmt.Errorf("error setting organization access for team %s: %w", team.ID, err)
					}
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

func flattenTeamOrganizationAccess(organizationAccess *tfe.OrganizationAccess) []map[string]bool {
	if organizationAccess == nil {
		return nil
	}

	return []map[string]bool{{
		"manage_policies":            organizationAccess.ManagePolicies,
		"manage_policy_overrides":    organizationAccess.ManagePolicyOverrides,
		"delegate_policy_overrides":  organizationAccess.DelegatePolicyOverrides,
		"manage_workspaces":          organizationAccess.ManageWorkspaces,
		"manage_vcs_settings":        organizationAccess.ManageVCSSettings,
		"manage_providers":           organizationAccess.ManageProviders,
		"manage_modules":             organizationAccess.ManageModules,
		"manage_run_tasks":           organizationAccess.ManageRunTasks,
		"manage_projects":            organizationAccess.ManageProjects,
		"read_projects":              organizationAccess.ReadProjects,
		"read_workspaces":            organizationAccess.ReadWorkspaces,
		"manage_membership":          organizationAccess.ManageMembership,
		"manage_teams":               organizationAccess.ManageTeams,
		"manage_organization_access": organizationAccess.ManageOrganizationAccess,
		"access_secret_teams":        organizationAccess.AccessSecretTeams,
		"manage_agent_pools":         organizationAccess.ManageAgentPools,
	}}
}
