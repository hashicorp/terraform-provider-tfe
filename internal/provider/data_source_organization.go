// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

// NOTE: This is a legacy resource and should be migrated to the Plugin
// Framework if substantial modifications are planned. See
// docs/new-resources.md if planning to use this code as boilerplate for
// a new resource.

package provider

import (
	"errors"
	"fmt"
	"log"

	tfev2 "github.com/hashicorp/go-tfe/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTFEOrganization() *schema.Resource {
	return &schema.Resource{
		Description: "Gets information about an organization.",

		Read: dataSourceTFEOrganizationRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The external ID of the organization. Do not rely on this value — use `external_id` instead.",
				Type:        schema.TypeString,
				Computed:    true,
			},

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
				Description: "A SAML attribute value used to identify members of the Owners team. When SAML SSO is enabled, users whose SAML role attribute matches this value will be added to the \"owners\" team.",
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
	env, err := config.ClientV2.API.Organizations().ByOrganization_name(name).Get(ctx, nil)
	if err != nil {
		if errors.Is(err, tfev2.ErrNotFound) {
			return fmt.Errorf("could not read organization '%s'", name)
		}
		return fmt.Errorf("Error retrieving organization: %w", err)
	}

	orgData := env.GetData()
	if orgData == nil {
		return fmt.Errorf("could not read organization '%s'", name)
	}

	log.Printf("[DEBUG] Setting Organization Attributes")
	externalID := valueOrZero(orgData.GetId())
	d.SetId(externalID)
	d.Set("external_id", externalID)

	if attrs := orgData.GetAttributes(); attrs != nil {
		d.Set("name", valueOrZero(attrs.GetName()))
		d.Set("collaborator_auth_policy", enumStringOrEmpty(attrs.GetCollaboratorAuthPolicy()))
		d.Set("cost_estimation_enabled", valueOrZero(attrs.GetCostEstimationEnabled()))
		d.Set("email", valueOrZero(attrs.GetEmail()))
		d.Set("owners_team_saml_role_id", valueOrZero(attrs.GetOwnersTeamSamlRoleId()))
		d.Set("two_factor_conformant", valueOrZero(attrs.GetTwoFactorConformant()))
		d.Set("send_passing_statuses_for_untriggered_speculative_plans", valueOrZero(attrs.GetSendPassingStatusesForUntriggeredSpeculativePlans()))
		d.Set("aggregated_commit_status_enabled", valueOrZero(attrs.GetAggregatedCommitStatusEnabled()))
		d.Set("assessments_enforced", valueOrZero(attrs.GetAssessmentsEnforced()))
		d.Set("speculative_plan_management_enabled", valueOrZero(attrs.GetSpeculativePlanManagementEnabled()))
	}

	if rel := orgData.GetRelationships(); rel != nil {
		if dp := rel.GetDefaultProject(); dp != nil && dp.GetData() != nil {
			d.Set("default_project_id", valueOrZero(dp.GetData().GetId()))
		}
	}

	// enforce_hyok and max_ttl_enabled have no v2 generated getter (go-tfe/v2
	// gap); backfill via a narrow v1 read.
	if err := readOrganizationHYOKAndMaxTTLV1(d, config, name); err != nil {
		return err
	}

	return nil
}

// readOrganizationHYOKAndMaxTTLV1 backfills enforce_hyok and max_ttl_enabled
// from the v1 client. Organizations_attributes in the pinned go-tfe/v2
// generated client has no getter for these two fields. Unlike
// resource_tfe_organization.go's equivalent helper, this data source's
// schema has no user_tokens_enabled attribute, so that field is not read.
func readOrganizationHYOKAndMaxTTLV1(d *schema.ResourceData, config ConfiguredClient, name string) error {
	org, err := config.Client.Organizations.Read(ctx, name)
	if err != nil {
		return fmt.Errorf("Error retrieving organization: %w", err)
	}

	d.Set("enforce_hyok", org.EnforceHYOK)
	d.Set("max_ttl_enabled", org.MaxTTLEnabled)

	return nil
}
