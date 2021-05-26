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

			"owners_team_saml_role_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"saml_enalbed": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"two_factor_conformant": {
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
	d.SetId(org.ExternalID)
	d.Set("name", org.Name)
	d.Set("external_id", org.ExternalID)
	d.Set("collaborator_auth_policy", org.CollaboratorAuthPolicy)
	d.Set("cost_estimation_enabled", org.CostEstimationEnabled)
	d.Set("email", org.Email)
	d.Set("owners_team_saml_role_id", org.OwnersTeamSAMLRoleID)
	d.Set("saml_enalbed", org.SAMLEnabled)
	d.Set("two_factor_conformant", org.TwoFactorConformant)

	return nil
}
