package tfe

import (
	"fmt"
	"log"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTFEOrganization() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTFEOrganizationRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"external_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"collaborator_auth_policy": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"cost_estimation_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"email": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"enterprise_plan": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"owners_team_saml_role_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"permissions": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"can_create_team": {
							Type:     schema.TypeBool,
							Computed: true,
						},

						"can_create_workspace": {
							Type:     schema.TypeBool,
							Computed: true,
						},

						"can_create_workspace_migration": {
							Type:     schema.TypeBool,
							Computed: true,
						},

						"can_destroy": {
							Type:     schema.TypeBool,
							Computed: true,
						},

						"can_traverse": {
							Type:     schema.TypeBool,
							Computed: true,
						},

						"can_update": {
							Type:     schema.TypeBool,
							Computed: true,
						},

						"can_update_api_token": {
							Type:     schema.TypeBool,
							Computed: true,
						},

						"can_update_oauth": {
							Type:     schema.TypeBool,
							Computed: true,
						},

						"can_update_sentinel": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},

			"saml_enalbed": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"session_remember": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"session_timeout": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"two_factor_confrmant": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataSourceTFEOrganizationRead(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	name := d.Get("name").(string)
	log.Printf("[DEBUG] Read configuration for Organization: %s", name)
	org, err := tfeClient.Organizations.Read(ctx, name)
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return fmt.Errorf("Could not read organization '%s'", name)
		}
		return fmt.Errorf("Error retrieving organization: %v", err)
	}

	log.Printf("[DEBUG] Setting Organization Attributes")
	d.Set("name", org.Name)
	d.SetId(org.Name)
	d.Set("external_id", org.ExternalID)
	d.Set("collaborator_auth_policy", org.CollaboratorAuthPolicy)
	d.Set("cost_estimation_enabled", org.CostEstimationEnabled)
	d.Set("email", org.Email)
	d.Set("enterprise_plan", org.EnterprisePlan)
	var permissions []interface{}
	permissions = append(permissions, map[string]interface{}{
		"can_create_team":                org.Permissions.CanCreateTeam,
		"can_create_workspace":           org.Permissions.CanCreateWorkspace,
		"can_create_workspace_migration": org.Permissions.CanCreateWorkspaceMigration,
		"can_destroy":                    org.Permissions.CanDestroy,
		"can_traverse":                   org.Permissions.CanTraverse,
		"can_update":                     org.Permissions.CanUpdate,
		"can_update_api_token":           org.Permissions.CanUpdateAPIToken,
		"can_update_oauth":               org.Permissions.CanUpdateOAuth,
		"can_update_sentinel":            org.Permissions.CanUpdateSentinel,
	})
	d.Set("permissions", permissions)
	d.Set("owners_team_saml_role_id", org.OwnersTeamSAMLRoleID)
	d.Set("saml_enalbed", org.SAMLEnabled)
	d.Set("session_remember", org.SessionRemember)
	d.Set("session_timeout", org.SessionTimeout)
	d.Set("two_factor_confrmant", org.TwoFactorConformant)

	return nil
}
