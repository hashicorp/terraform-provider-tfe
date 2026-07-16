// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

// NOTE: This is a legacy resource and should be migrated to the Plugin
// Framework if substantial modifications are planned. See
// docs/new-resources.md if planning to use this code as boilerplate for
// a new resource.

package provider

import (
	"fmt"
	"log"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTFEOrganization() *schema.Resource {
	return &schema.Resource{
		Description: "Gets information about an organization.",

		Read: dataSourceTFEOrganizationRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Name of the organization. If omitted, organization must be defined in the provider config.",
				Type:        schema.TypeString,
				Optional:    true,
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
				Description: "Whether or not the cost estimation feature is enabled for all workspaces in the organization. Defaults to true. In a HCP Terraform organization which does not have Teams & Governance features, this value is always false and cannot be changed. In Terraform Enterprise, Cost Estimation must also be enabled in Site Administration.",
				Type:        schema.TypeBool,
				Computed:    true,
			},

			"default_project_id": {
				Description: "ID of the organization's default project. All workspaces created without specifying a project ID are created in this project.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"email": {
				Description: "Admin email address.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"owners_team_saml_role_id": {
				Description: "The name of the \"owners\" team.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"two_factor_conformant": {
				Description: "Whether or not to require two factor authentication for this organization.",
				Type:        schema.TypeBool,
				Computed:    true,
			},

			"send_passing_statuses_for_untriggered_speculative_plans": {
				Description: "Whether or not to send VCS status updates for untriggered speculative plans. This can be useful if large numbers of untriggered workspaces are exhausting request limits for connected version control service providers like GitHub. Defaults to true. In Terraform Enterprise, this setting has no effect and cannot be changed but is also available in Site Administration.",
				Type:        schema.TypeBool,
				Computed:    true,
			},

			"aggregated_commit_status_enabled": {
				Description: "Whether or not to enable Aggregated Status Checks. This can be useful for monorepo repositories with multiple workspaces receiving status checks for events such as a pull request.",
				Type:        schema.TypeBool,
				Computed:    true,
			},

			"assessments_enforced": {
				Description: "(Available only in HCP Terraform) Whether to force health assessments (drift detection) on all eligible workspaces or allow workspaces to set their own preferences.",
				Type:        schema.TypeBool,
				Computed:    true,
			},

			"speculative_plan_management_enabled": {
				Description: "Whether or not to enable Speculative Plan Management. If true, pending VCS-triggered speculative plans from outdated commits will be cancelled if a newer commit is pushed to the same branch.",
				Type:        schema.TypeBool,
				Computed:    true,
			},

			"enforce_hyok": {
				Description: "(Available only in HCP Terraform) Whether HYOK is enforced for all new workspaces in the organization.",
				Type:        schema.TypeBool,
				Computed:    true,
			},

			"max_ttl_enabled": {
				Description: "Whether maximum token TTL policies are enabled for the organization.",
				Type:        schema.TypeBool,
				Computed:    true,
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
	d.Set("name", org.Name)
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
	d.Set("aggregated_commit_status_enabled", org.AggregatedCommitStatusEnabled)
	d.Set("assessments_enforced", org.AssessmentsEnforced)
	d.Set("speculative_plan_management_enabled", org.SpeculativePlanManagementEnabled)
	d.Set("enforce_hyok", org.EnforceHYOK)
	d.Set("max_ttl_enabled", org.MaxTTLEnabled)

	return nil
}
