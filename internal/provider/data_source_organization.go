// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

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

			"default_project_id": {
				Type:     schema.TypeString,
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

			"two_factor_conformant": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"send_passing_statuses_for_untriggered_speculative_plans": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"assessments_enforced": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataSourceTFEOrganizationRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	name, err := config.schemaOrDefaultOrganizationKey(d, "name")
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Read configuration for Organization: %s", name)
	org, err := config.Client.Organizations.Read(ctx, name)
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return fmt.Errorf("could not read organization '%s'", name)
		}
		return fmt.Errorf("Error retrieving organization: %w", err)
	}

	log.Printf("[DEBUG] Setting Organization Attributes")
	d.SetId(org.ExternalID)
	d.Set("external_id", org.ExternalID)
	d.Set("collaborator_auth_policy", org.CollaboratorAuthPolicy)
	d.Set("cost_estimation_enabled", org.CostEstimationEnabled)

	if org.DefaultProject != nil {
		d.Set("default_project_id", org.DefaultProject.ID)
	}

	d.Set("email", org.Email)
	d.Set("owners_team_saml_role_id", org.OwnersTeamSAMLRoleID)
	d.Set("two_factor_conformant", org.TwoFactorConformant)
	d.Set("send_passing_statuses_for_untriggered_speculative_plans", org.SendPassingStatusesForUntriggeredSpeculativePlans)
	d.Set("assessments_enforced", org.AssessmentsEnforced)

	return nil
}
