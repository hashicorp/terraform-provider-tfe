// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

// NOTE: This is a legacy resource and should be migrated to the Plugin
// Framework if substantial modifications are planned. See
// docs/new-resources.md if planning to use this code as boilerplate for
// a new resource.

package provider

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-tfe/internal/provider/helpers"

	"errors"

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

		Identity: &schema.ResourceIdentity{
			SchemaFunc: func() map[string]*schema.Schema {
				return map[string]*schema.Schema{
					"id": {
						Type:              schema.TypeString,
						RequiredForImport: true,
					},
					"hostname": {
						Type:              schema.TypeString,
						OptionalForImport: true,
					},
					"organization": {
						Type:              schema.TypeString,
						RequiredForImport: true,
					},
				}
			},
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Name of the team.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"organization": {
				Description: "Name of the organization. If omitted, organization must be defined in the provider config.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
			},
			"organization_access": {
				Description: "Settings for the team's organization access.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"manage_policies": {
							Description: "Allows members to create, edit, and delete the organization's Sentinel policies.",
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
						},
						"manage_policy_overrides": {
							Description: "Allows members to override soft-mandatory policy checks.",
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
						},
						"delegate_policy_overrides": {
							Description: "When this setting is enabled for a team, its members can override failed policy evaluations on projects and workspaces they manage.",
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
						},
						"manage_workspaces": {
							Description: "Allows members to create and administrate all workspaces within the organization.",
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
						},
						"manage_vcs_settings": {
							Description: "Allows members to manage the organization's VCS Providers and SSH keys.",
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
						},
						"manage_providers": {
							Description: "Allow members to publish and delete providers in the organization's private registry.",
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
						},
						"manage_modules": {
							Description: "Allow members to publish and delete modules in the organization's private registry.",
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
						},
						"manage_run_tasks": {
							Description: "Allow members to create, edit, and delete the organization's run tasks.",
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
						},
						"manage_projects": {
							Description: "Allow members to create and administrate all projects within the organization. Requires manage_workspaces to be set to true.",
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
						},
						"read_workspaces": {
							Description: "Allow members to view all workspaces in this organization.",
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
						},
						"read_projects": {
							Description: "Allow members to view all projects within the organization. Requires read_workspaces to be set to true.",
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
						},
						"manage_membership": {
							Description: "Allow members to add/remove users from the organization, and to add/remove users from visible teams.",
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
						},
						"manage_teams": {
							Description: "Allow members to create, update, and delete teams.",
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
						},
						"manage_organization_access": {
							Description: "Allow members to update the organization access settings of teams.",
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
						},
						"access_secret_teams": {
							Description: "Allow members access to secret teams up to the level of permissions granted by their team permissions setting.",
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
						},
						"manage_agent_pools": {
							Description: "Allow members to create, edit, and delete agent pools within their organization.",
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
						},
					},
				},
			},
			"visibility": {
				Description: "The visibility of the team (\"secret\" or \"organization\").",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ValidateFunc: validation.StringInSlice([]string{
					"secret",
					"organization",
				}, false),
			},
			"sso_team_id": {
				Description: "Unique identifier to control team membership via SAML. Defaults to null.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"allow_member_token_management": {
				Description: "Whether team members can manage team tokens. Used by Owners and users with Manage Teams permissions. Defaults to true.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
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
			ManagePolicies:           tfe.Bool(organizationAccess["manage_policies"].(bool)),
			ManagePolicyOverrides:    tfe.Bool(organizationAccess["manage_policy_overrides"].(bool)),
			DelegatePolicyOverrides:  tfe.Bool(organizationAccess["delegate_policy_overrides"].(bool)),
			ManageWorkspaces:         tfe.Bool(organizationAccess["manage_workspaces"].(bool)),
			ManageVCSSettings:        tfe.Bool(organizationAccess["manage_vcs_settings"].(bool)),
			ManageProviders:          tfe.Bool(organizationAccess["manage_providers"].(bool)),
			ManageModules:            tfe.Bool(organizationAccess["manage_modules"].(bool)),
			ManageRunTasks:           tfe.Bool(organizationAccess["manage_run_tasks"].(bool)),
			ManageProjects:           tfe.Bool(organizationAccess["manage_projects"].(bool)),
			ReadWorkspaces:           tfe.Bool(organizationAccess["read_workspaces"].(bool)),
			ReadProjects:             tfe.Bool(organizationAccess["read_projects"].(bool)),
			ManageMembership:         tfe.Bool(organizationAccess["manage_membership"].(bool)),
			ManageTeams:              tfe.Bool(organizationAccess["manage_teams"].(bool)),
			ManageOrganizationAccess: tfe.Bool(organizationAccess["manage_organization_access"].(bool)),
			AccessSecretTeams:        tfe.Bool(organizationAccess["access_secret_teams"].(bool)),
			ManageAgentPools:         tfe.Bool(organizationAccess["manage_agent_pools"].(bool)),
		}
	}

	if v, ok := d.GetOk("visibility"); ok {
		options.Visibility = tfe.String(v.(string))
	}

	if v, ok := d.GetOk("sso_team_id"); ok {
		options.SSOTeamID = tfe.String(v.(string))
	}

	options.AllowMemberTokenManagement = tfe.Bool(d.Get("allow_member_token_management").(bool))

	log.Printf("[DEBUG] Create team %s for organization: %s", name, organization)
	team, err := config.Client.Teams.Create(ctx, organization, options)
	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
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

	err = helpers.WriteTFEIdentityWithOrg(d, team.ID, organization, config.Client.BaseURL().Host)
	if err != nil {
		return err
	}

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

	organization, err := config.schemaOrDefaultOrganization(d)
	if err != nil {
		return err
	}

	err = helpers.WriteTFEIdentityWithOrg(d, team.ID, organization, config.Client.BaseURL().Host)
	if err != nil {
		return err
	}

	// Update the config.
	d.Set("name", team.Name)
	if team.OrganizationAccess != nil {
		organizationAccess := []map[string]bool{{
			"manage_policies":            team.OrganizationAccess.ManagePolicies,
			"manage_policy_overrides":    team.OrganizationAccess.ManagePolicyOverrides,
			"delegate_policy_overrides":  team.OrganizationAccess.DelegatePolicyOverrides,
			"manage_workspaces":          team.OrganizationAccess.ManageWorkspaces,
			"manage_vcs_settings":        team.OrganizationAccess.ManageVCSSettings,
			"manage_providers":           team.OrganizationAccess.ManageProviders,
			"manage_modules":             team.OrganizationAccess.ManageModules,
			"manage_run_tasks":           team.OrganizationAccess.ManageRunTasks,
			"manage_projects":            team.OrganizationAccess.ManageProjects,
			"read_projects":              team.OrganizationAccess.ReadProjects,
			"read_workspaces":            team.OrganizationAccess.ReadWorkspaces,
			"manage_membership":          team.OrganizationAccess.ManageMembership,
			"manage_teams":               team.OrganizationAccess.ManageTeams,
			"manage_organization_access": team.OrganizationAccess.ManageOrganizationAccess,
			"access_secret_teams":        team.OrganizationAccess.AccessSecretTeams,
			"manage_agent_pools":         team.OrganizationAccess.ManageAgentPools,
		}}
		if err := d.Set("organization_access", organizationAccess); err != nil {
			return fmt.Errorf("error setting organization access for team %s: %w", d.Id(), err)
		}
	}
	d.Set("visibility", team.Visibility)
	d.Set("sso_team_id", team.SSOTeamID)
	d.Set("allow_member_token_management", team.AllowMemberTokenManagement)

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
			ManagePolicies:           tfe.Bool(organizationAccess["manage_policies"].(bool)),
			ManagePolicyOverrides:    tfe.Bool(organizationAccess["manage_policy_overrides"].(bool)),
			DelegatePolicyOverrides:  tfe.Bool(organizationAccess["delegate_policy_overrides"].(bool)),
			ManageWorkspaces:         tfe.Bool(organizationAccess["manage_workspaces"].(bool)),
			ManageVCSSettings:        tfe.Bool(organizationAccess["manage_vcs_settings"].(bool)),
			ManageProviders:          tfe.Bool(organizationAccess["manage_providers"].(bool)),
			ManageModules:            tfe.Bool(organizationAccess["manage_modules"].(bool)),
			ManageRunTasks:           tfe.Bool(organizationAccess["manage_run_tasks"].(bool)),
			ManageProjects:           tfe.Bool(organizationAccess["manage_projects"].(bool)),
			ReadProjects:             tfe.Bool(organizationAccess["read_projects"].(bool)),
			ReadWorkspaces:           tfe.Bool(organizationAccess["read_workspaces"].(bool)),
			ManageMembership:         tfe.Bool(organizationAccess["manage_membership"].(bool)),
			ManageTeams:              tfe.Bool(organizationAccess["manage_teams"].(bool)),
			ManageOrganizationAccess: tfe.Bool(organizationAccess["manage_organization_access"].(bool)),
			AccessSecretTeams:        tfe.Bool(organizationAccess["access_secret_teams"].(bool)),
			ManageAgentPools:         tfe.Bool(organizationAccess["manage_agent_pools"].(bool)),
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

	options.AllowMemberTokenManagement = tfe.Bool(d.Get("allow_member_token_management").(bool))

	log.Printf("[DEBUG] Update team: %s", d.Id())
	_, err := config.Client.Teams.Update(ctx, d.Id(), options)
	if err != nil {
		return fmt.Errorf(
			"Error updating team %s: %w", d.Id(), err)
	}

	organization, err := config.schemaOrDefaultOrganization(d)
	if err != nil {
		return err
	}

	err = helpers.WriteTFEIdentityWithOrg(d, d.Id(), organization, config.Client.BaseURL().Host)
	if err != nil {
		return err
	}

	return nil
}

func resourceTFETeamDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Delete team: %s", d.Id())
	err := config.Client.Teams.Delete(ctx, d.Id())
	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
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

	// First we'll check for an identity
	identity, err := d.Identity()
	if err != nil {
		return nil, fmt.Errorf("error reading team identity: %w", err)
	}

	if externalID := identity.Get("id").(string); externalID != "" {
		// We are importing by identity
		// This only supported when using an import block, since import blocks
		// are the only way to specify an identity. Importing via TF CLI does
		// not support specifying an identity.
		d.SetId(externalID)
		orgName := identity.Get("organization").(string)
		err = d.Set("organization", orgName)
		if err != nil {
			return nil, fmt.Errorf("could not set organization name %s on resource: %w", orgName, err)
		}

		// Exit early
		return []*schema.ResourceData{d}, nil
	}

	s := strings.SplitN(d.Id(), "/", 2)
	if len(s) != 2 {
		return nil, fmt.Errorf(
			"invalid team import format: %s (expected <ORGANIZATION>/<TEAM ID> or <ORGANIZATION>/<TEAM NAME>)",
			d.Id(),
		)
	}

	orgName := s[0]
	teamNameOrID := s[1]

	err = d.Set("organization", orgName)
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
