// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

// NOTE: This is a legacy resource and should be migrated to the Plugin
// Framework if substantial modifications are planned. See
// docs/new-resources.md if planning to use this code as boilerplate for
// a new resource.

package provider

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	tfe "github.com/hashicorp/go-tfe/v2"
	"github.com/hashicorp/go-tfe/v2/api/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-tfe/internal/provider/helpers"
)

func resourceTFETeam() *schema.Resource {
	return &schema.Resource{
		Description: "Manages teams.",

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
			"id": {
				Description: "The ID of the team.",
				Type:        schema.TypeString,
				Computed:    true,
			},
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
				Description: "Settings for the team's [organization access](https://developer.hashicorp.com/terraform/cloud-docs/users-teams-organizations/permissions#organization-permissions).",
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
				Description: "Unique identifier to control [team membership](https://developer.hashicorp.com/terraform/cloud-docs/users-teams-organizations/single-sign-on#team-names-and-sso-team-ids) via SAML. Defaults to null.",
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

// teamOrganizationAccessAttributesV2 builds the go-tfe v2 organization-access
// attributes from the resource's "organization_access" block. Returns nil
// when the block is not set, matching go-tfe v1's optional-pointer behavior.
func teamOrganizationAccessAttributesV2(d *schema.ResourceData) models.Teams_attributes_organizationAccessable {
	v, ok := d.GetOk("organization_access")
	if !ok {
		return nil
	}
	organizationAccess := v.([]interface{})[0].(map[string]interface{})

	access := models.NewTeams_attributes_organizationAccess()
	access.SetManagePolicies(ptr(organizationAccess["manage_policies"].(bool)))
	access.SetManagePolicyOverrides(ptr(organizationAccess["manage_policy_overrides"].(bool)))
	access.SetDelegatePolicyOverrides(ptr(organizationAccess["delegate_policy_overrides"].(bool)))
	access.SetManageWorkspaces(ptr(organizationAccess["manage_workspaces"].(bool)))
	access.SetManageVcsSettings(ptr(organizationAccess["manage_vcs_settings"].(bool)))
	access.SetManageProviders(ptr(organizationAccess["manage_providers"].(bool)))
	access.SetManageModules(ptr(organizationAccess["manage_modules"].(bool)))
	access.SetManageRunTasks(ptr(organizationAccess["manage_run_tasks"].(bool)))
	access.SetManageProjects(ptr(organizationAccess["manage_projects"].(bool)))
	access.SetReadWorkspaces(ptr(organizationAccess["read_workspaces"].(bool)))
	access.SetReadProjects(ptr(organizationAccess["read_projects"].(bool)))
	access.SetManageMembership(ptr(organizationAccess["manage_membership"].(bool)))
	access.SetManageTeams(ptr(organizationAccess["manage_teams"].(bool)))
	access.SetManageOrganizationAccess(ptr(organizationAccess["manage_organization_access"].(bool)))
	access.SetAccessSecretTeams(ptr(organizationAccess["access_secret_teams"].(bool)))
	access.SetManageAgentPools(ptr(organizationAccess["manage_agent_pools"].(bool)))

	return access
}

func resourceTFETeamCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	// Get team attributes.
	name := d.Get("name").(string)
	organization, err := config.schemaOrDefaultOrganization(d)
	if err != nil {
		return err
	}

	attributes := models.NewTeams_attributes()
	attributes.SetName(ptr(name))
	attributes.SetOrganizationAccess(teamOrganizationAccessAttributesV2(d))

	if v, ok := d.GetOk("visibility"); ok {
		visibility, verr := models.ParseTeams_attributes_visibility(v.(string))
		if verr != nil {
			return fmt.Errorf("invalid team visibility %q: %w", v.(string), verr)
		}
		attributes.SetVisibility(visibility.(*models.Teams_attributes_visibility))
	}

	if v, ok := d.GetOk("sso_team_id"); ok {
		attributes.SetSsoTeamId(ptr(v.(string)))
	}

	attributes.SetAllowMemberTokenManagement(ptr(d.Get("allow_member_token_management").(bool)))

	team := models.NewTeams()
	team.SetTypeEscaped(ptr(models.TEAMS_TEAMS_TYPE))
	team.SetAttributes(attributes)

	envelope := models.NewTeamsEnvelope()
	envelope.SetData(team)

	log.Printf("[DEBUG] Create team %s for organization: %s", name, organization)
	result, err := config.ClientV2.API.Organizations().ByOrganization_name(organization).Teams().Post(ctx, envelope, nil)
	if err != nil {
		if errors.Is(err, tfe.ErrNotFound) {
			// go-tfe v2 does not generate a route for organization
			// entitlement sets (no /organizations/{name}/entitlement-set
			// request builder exists in the generated client). This call
			// remains on go-tfe v1 solely to enrich the error message below
			// when team creation 404s; it is not part of the primary create
			// call path.
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
	if result == nil || result.GetData() == nil {
		return fmt.Errorf("Error creating team %s for organization %s: no data returned", name, organization)
	}

	teamID := valueOrZero(result.GetData().GetId())
	d.SetId(teamID)

	err = helpers.WriteTFEIdentityWithOrg(d, teamID, organization, config.Client.BaseURL().Host)
	if err != nil {
		return err
	}

	return resourceTFETeamRead(d, meta)
}

func resourceTFETeamRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Read configuration of team: %s", d.Id())
	result, err := config.ClientV2.API.Teams().ById(d.Id()).Get(ctx, nil)
	if err != nil {
		if errors.Is(err, tfe.ErrNotFound) {
			log.Printf("[DEBUG] Team %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading configuration of team %s: %w", d.Id(), err)
	}
	if result == nil || result.GetData() == nil {
		log.Printf("[DEBUG] Team %s no longer exists", d.Id())
		d.SetId("")
		return nil
	}
	team := result.GetData()

	organization, err := config.schemaOrDefaultOrganization(d)
	if err != nil {
		return err
	}

	err = helpers.WriteTFEIdentityWithOrg(d, valueOrZero(team.GetId()), organization, config.Client.BaseURL().Host)
	if err != nil {
		return err
	}

	attrs := team.GetAttributes()

	// Update the config.
	d.Set("name", valueOrZero(attrs.GetName()))
	if organizationAccess := attrs.GetOrganizationAccess(); organizationAccess != nil {
		organizationAccessData := []map[string]bool{{
			"manage_policies":            valueOrZero(organizationAccess.GetManagePolicies()),
			"manage_policy_overrides":    valueOrZero(organizationAccess.GetManagePolicyOverrides()),
			"delegate_policy_overrides":  valueOrZero(organizationAccess.GetDelegatePolicyOverrides()),
			"manage_workspaces":          valueOrZero(organizationAccess.GetManageWorkspaces()),
			"manage_vcs_settings":        valueOrZero(organizationAccess.GetManageVcsSettings()),
			"manage_providers":           valueOrZero(organizationAccess.GetManageProviders()),
			"manage_modules":             valueOrZero(organizationAccess.GetManageModules()),
			"manage_run_tasks":           valueOrZero(organizationAccess.GetManageRunTasks()),
			"manage_projects":            valueOrZero(organizationAccess.GetManageProjects()),
			"read_projects":              valueOrZero(organizationAccess.GetReadProjects()),
			"read_workspaces":            valueOrZero(organizationAccess.GetReadWorkspaces()),
			"manage_membership":          valueOrZero(organizationAccess.GetManageMembership()),
			"manage_teams":               valueOrZero(organizationAccess.GetManageTeams()),
			"manage_organization_access": valueOrZero(organizationAccess.GetManageOrganizationAccess()),
			"access_secret_teams":        valueOrZero(organizationAccess.GetAccessSecretTeams()),
			"manage_agent_pools":         valueOrZero(organizationAccess.GetManageAgentPools()),
		}}
		if err := d.Set("organization_access", organizationAccessData); err != nil {
			return fmt.Errorf("error setting organization access for team %s: %w", d.Id(), err)
		}
	}
	if visibility := attrs.GetVisibility(); visibility != nil {
		d.Set("visibility", visibility.String())
	} else {
		d.Set("visibility", "")
	}
	d.Set("sso_team_id", valueOrZero(attrs.GetSsoTeamId()))
	d.Set("allow_member_token_management", valueOrZero(attrs.GetAllowMemberTokenManagement()))

	return nil
}

func resourceTFETeamUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	// Get the name.
	name := d.Get("name").(string)

	attributes := models.NewTeams_attributes()
	attributes.SetName(ptr(name))
	attributes.SetOrganizationAccess(teamOrganizationAccessAttributesV2(d))

	if v, ok := d.GetOk("visibility"); ok {
		visibility, verr := models.ParseTeams_attributes_visibility(v.(string))
		if verr != nil {
			return fmt.Errorf("invalid team visibility %q: %w", v.(string), verr)
		}
		attributes.SetVisibility(visibility.(*models.Teams_attributes_visibility))
	}

	if v, ok := d.GetOk("sso_team_id"); ok {
		attributes.SetSsoTeamId(ptr(v.(string)))
	} else {
		attributes.SetSsoTeamId(ptr(""))
	}

	attributes.SetAllowMemberTokenManagement(ptr(d.Get("allow_member_token_management").(bool)))

	team := models.NewTeams()
	team.SetTypeEscaped(ptr(models.TEAMS_TEAMS_TYPE))
	team.SetId(ptr(d.Id()))
	team.SetAttributes(attributes)

	envelope := models.NewTeamsEnvelope()
	envelope.SetData(team)

	log.Printf("[DEBUG] Update team: %s", d.Id())
	_, err := config.ClientV2.API.Teams().ById(d.Id()).Patch(ctx, envelope, nil)
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
	err := config.ClientV2.API.Teams().ById(d.Id()).Delete(ctx, nil)
	if err != nil {
		if errors.Is(err, tfe.ErrNotFound) {
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
		result, err := config.ClientV2.API.Teams().ById(teamNameOrID).Get(ctx, nil)
		if err == nil && result != nil && result.GetData() != nil {
			d.SetId(valueOrZero(result.GetData().GetId()))
			return []*schema.ResourceData{d}, nil
		}
	}

	// a team does not exist (or cannot be found) with the ID s[1]...check if it is the team name instead
	team, err := fetchTeamByNameV2(ctx, config.ClientV2.API, orgName, teamNameOrID)
	if err != nil {
		return nil, fmt.Errorf("no team found with name or ID %s in organization %s: %w", teamNameOrID, orgName, err)
	}
	d.SetId(valueOrZero(team.GetId()))
	return []*schema.ResourceData{d}, nil
}
