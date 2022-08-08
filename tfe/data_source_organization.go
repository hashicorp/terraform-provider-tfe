package tfe

import (
	"fmt"
	"log"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTFEOrganization() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to get information about an organization.",
		Read:        dataSourceTFEOrganizationRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Name of the organization.",
				Type:        schema.TypeString,
				Required:    true,
			},

			"external_id": {
				Description: "An identifier for the organization.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"collaborator_auth_policy": {
				Description: "Authentication policy (`password` or `two_factor_mandatory`). Defaults to `password`.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"cost_estimation_enabled": {
				Description: "Whether or not the cost estimation feature is enabled for all workspaces in the organization. Defaults to true. In a Terraform Cloud organization which does not have Teams & Governance features, this value is always false and cannot be changed. In Terraform Enterprise, Cost Estimation must also be enabled in Site Administration.",
				Type:        schema.TypeBool,
				Computed:    true,
			},

			"email": {
				Description: "Admin email address.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"owners_team_saml_role_id": {
				Description: "The name of the owners team.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"two_factor_conformant": {
				Description: "",
				Type:        schema.TypeBool,
				Computed:    true,
			},

			"send_passing_statuses_for_untriggered_speculative_plans": {
				Description: "Whether or not to send VCS status updates for untriggered speculative plans. This can be useful if large numbers of untriggered workspaces are exhausting request limits for connected version control service providers like GitHub. Defaults to true. In Terraform Enterprise, this setting has no effect and cannot be changed but is also available in Site Administration.",
				Type:        schema.TypeBool,
				Computed:    true,
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
		return fmt.Errorf("Error retrieving organization: %w", err)
	}

	log.Printf("[DEBUG] Setting Organization Attributes")
	d.SetId(org.ExternalID)
	d.Set("name", org.Name)
	d.Set("external_id", org.ExternalID)
	d.Set("collaborator_auth_policy", org.CollaboratorAuthPolicy)
	d.Set("cost_estimation_enabled", org.CostEstimationEnabled)
	d.Set("email", org.Email)
	d.Set("owners_team_saml_role_id", org.OwnersTeamSAMLRoleID)
	d.Set("two_factor_conformant", org.TwoFactorConformant)
	d.Set("send_passing_statuses_for_untriggered_speculative_plans", org.SendPassingStatusesForUntriggeredSpeculativePlans)

	return nil
}
