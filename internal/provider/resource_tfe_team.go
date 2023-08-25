// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTFETeam() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFETeamCreate,
		Read:   resourceTFETeamRead,
		Update: resourceTFETeamUpdate,
		Delete: resourceTFETeamDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceTFETeamImporter,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"organization": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"organization_access": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"manage_policies": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"manage_policy_overrides": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"manage_workspaces": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"manage_vcs_settings": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"manage_providers": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"manage_modules": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"manage_run_tasks": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"manage_projects": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"read_workspaces": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"read_projects": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"manage_membership": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},
			"visibility": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "secret",
				ValidateFunc: validation.StringInSlice([]string{
					"secret",
					"organization",
				}, false),
			},
			"sso_team_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceTFETeamCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	// Get team attributes.
	name := d.Get("name").(string)
	organization, err := config.schemaOrDefaultOrganization(d)
	if err != nil {
		return err
	}

	// Create a new options struct.
	options := tfe.TeamCreateOptions{
		Name: tfe.String(name),
	}

	if v, ok := d.GetOk("organization_access"); ok {
		organizationAccess := v.([]interface{})[0].(map[string]interface{})

		options.OrganizationAccess = &tfe.OrganizationAccessOptions{
			ManagePolicies:        tfe.Bool(organizationAccess["manage_policies"].(bool)),
			ManagePolicyOverrides: tfe.Bool(organizationAccess["manage_policy_overrides"].(bool)),
			ManageWorkspaces:      tfe.Bool(organizationAccess["manage_workspaces"].(bool)),
			ManageVCSSettings:     tfe.Bool(organizationAccess["manage_vcs_settings"].(bool)),
			ManageProviders:       tfe.Bool(organizationAccess["manage_providers"].(bool)),
			ManageModules:         tfe.Bool(organizationAccess["manage_modules"].(bool)),
			ManageRunTasks:        tfe.Bool(organizationAccess["manage_run_tasks"].(bool)),
			ManageProjects:        tfe.Bool(organizationAccess["manage_projects"].(bool)),
			ReadWorkspaces:        tfe.Bool(organizationAccess["read_workspaces"].(bool)),
			ReadProjects:          tfe.Bool(organizationAccess["read_projects"].(bool)),
			ManageMembership:      tfe.Bool(organizationAccess["manage_membership"].(bool)),
		}
	}

	if v, ok := d.GetOk("visibility"); ok {
		options.Visibility = tfe.String(v.(string))
	}

	if v, ok := d.GetOk("sso_team_id"); ok {
		options.SSOTeamID = tfe.String(v.(string))
	}

	log.Printf("[DEBUG] Create team %s for organization: %s", name, organization)
	team, err := config.Client.Teams.Create(ctx, organization, options)
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			entitlements, _ := config.Client.Organizations.ReadEntitlements(ctx, organization)
			if entitlements == nil {
				return fmt.Errorf("Error creating team %s for organization %s: %w", name, organization, err)
			}
			if !entitlements.Teams {
				return fmt.Errorf("Error creating team %s for organization %s: missing entitlements to create teams", name, organization)
			}
		}
		return fmt.Errorf("Error creating team %s for organization %s: %w", name, organization, err)
	}

	d.SetId(team.ID)

	return resourceTFETeamRead(d, meta)
}

func resourceTFETeamRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Read configuration of team: %s", d.Id())
	team, err := config.Client.Teams.Read(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			log.Printf("[DEBUG] Team %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading configuration of team %s: %w", d.Id(), err)
	}

	// Update the config.
	d.Set("name", team.Name)
	if team.OrganizationAccess != nil {
		organizationAccess := []map[string]bool{{
			"manage_policies":         team.OrganizationAccess.ManagePolicies,
			"manage_policy_overrides": team.OrganizationAccess.ManagePolicyOverrides,
			"manage_workspaces":       team.OrganizationAccess.ManageWorkspaces,
			"manage_vcs_settings":     team.OrganizationAccess.ManageVCSSettings,
			"manage_providers":        team.OrganizationAccess.ManageProviders,
			"manage_modules":          team.OrganizationAccess.ManageModules,
			"manage_run_tasks":        team.OrganizationAccess.ManageRunTasks,
			"manage_projects":         team.OrganizationAccess.ManageProjects,
			"read_projects":           team.OrganizationAccess.ReadProjects,
			"read_workspaces":         team.OrganizationAccess.ReadWorkspaces,
			"manage_membership":       team.OrganizationAccess.ManageMembership,
		}}
		if err := d.Set("organization_access", organizationAccess); err != nil {
			return fmt.Errorf("error setting organization access for team %s: %w", d.Id(), err)
		}
	}
	d.Set("visibility", team.Visibility)
	d.Set("sso_team_id", team.SSOTeamID)

	return nil
}

func resourceTFETeamUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	// Get the name and organization.
	name := d.Get("name").(string)

	// create an options struct
	options := tfe.TeamUpdateOptions{
		Name: tfe.String(name),
	}

	if v, ok := d.GetOk("organization_access"); ok {
		organizationAccess := v.([]interface{})[0].(map[string]interface{})

		options.OrganizationAccess = &tfe.OrganizationAccessOptions{
			ManagePolicies:        tfe.Bool(organizationAccess["manage_policies"].(bool)),
			ManagePolicyOverrides: tfe.Bool(organizationAccess["manage_policy_overrides"].(bool)),
			ManageWorkspaces:      tfe.Bool(organizationAccess["manage_workspaces"].(bool)),
			ManageVCSSettings:     tfe.Bool(organizationAccess["manage_vcs_settings"].(bool)),
			ManageProviders:       tfe.Bool(organizationAccess["manage_providers"].(bool)),
			ManageModules:         tfe.Bool(organizationAccess["manage_modules"].(bool)),
			ManageRunTasks:        tfe.Bool(organizationAccess["manage_run_tasks"].(bool)),
			ManageProjects:        tfe.Bool(organizationAccess["manage_projects"].(bool)),
			ReadProjects:          tfe.Bool(organizationAccess["read_projects"].(bool)),
			ReadWorkspaces:        tfe.Bool(organizationAccess["read_workspaces"].(bool)),
			ManageMembership:      tfe.Bool(organizationAccess["manage_membership"].(bool)),
		}
	}

	if v, ok := d.GetOk("visibility"); ok {
		options.Visibility = tfe.String(v.(string))
	}

	if v, ok := d.GetOk("sso_team_id"); ok {
		options.SSOTeamID = tfe.String(v.(string))
	} else {
		options.SSOTeamID = tfe.String("")
	}

	log.Printf("[DEBUG] Update team: %s", d.Id())
	_, err := config.Client.Teams.Update(ctx, d.Id(), options)
	if err != nil {
		return fmt.Errorf(
			"Error updating team %s: %w", d.Id(), err)
	}

	return nil
}

func resourceTFETeamDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Delete team: %s", d.Id())
	err := config.Client.Teams.Delete(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return nil
		}
		return fmt.Errorf("Error deleting team %s: %w", d.Id(), err)
	}

	return nil
}

func resourceTFETeamImporter(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// Import formats:
	//  - <ORGANIZATION NAME>/<TEAM ID>
	//  - <ORGANIZATION NAME>/<TEAM NAME>
	s := strings.SplitN(d.Id(), "/", 2)
	if len(s) != 2 {
		return nil, fmt.Errorf(
			"invalid team import format: %s (expected <ORGANIZATION>/<TEAM ID> or <ORGANIZATION>/<TEAM NAME>)",
			d.Id(),
		)
	}

	orgName := s[0]
	teamNameOrID := s[1]

	err := d.Set("organization", orgName)
	if err != nil {
		return nil, fmt.Errorf("could not set organization name %s on resource: %w", orgName, err)
	}

	// we don't know if the second piece of the import is an ID or a team name. If it is an ID we should be able to read
	// the team by that ID
	config := meta.(ConfiguredClient)
	if isResourceIDFormat("team", teamNameOrID) {
		team, err := config.Client.Teams.Read(ctx, teamNameOrID)
		if err == nil {
			d.SetId(team.ID)
			return []*schema.ResourceData{d}, nil
		}
	}

	// a team does not exist (or cannot be found) with the ID s[1]...check if it is the team name instead
	team, err := fetchTeamByName(ctx, config.Client, orgName, teamNameOrID)
	if err != nil {
		return nil, fmt.Errorf("no team found with name or ID %s in organization %s: %w", teamNameOrID, orgName, err)
	}
	d.SetId(team.ID)
	return []*schema.ResourceData{d}, nil
}
