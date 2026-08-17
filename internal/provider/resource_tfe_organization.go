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
	"strings"

	tfe "github.com/hashicorp/go-tfe"
	tfev2 "github.com/hashicorp/go-tfe/v2"
	"github.com/hashicorp/go-tfe/v2/api/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceTFEOrganization() *schema.Resource {
	return &schema.Resource{
		Description: "Manages organizations.",

		Create: resourceTFEOrganizationCreate,
		Read:   resourceTFEOrganizationRead,
		Update: resourceTFEOrganizationUpdate,
		Delete: resourceTFEOrganizationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The name of the organization.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"name": {
				Description: "Name of the organization.",
				Type:        schema.TypeString,
				Required:    true,
				DiffSuppressFunc: func(k, old, current string, d *schema.ResourceData) bool {
					return strings.EqualFold(old, current)
				},
			},

			"email": {
				Description: "Admin email address.",
				Type:        schema.TypeString,
				Required:    true,
			},

			"session_timeout_minutes": {
				Description: "Session timeout after inactivity. Defaults to `20160`.",
				Type:        schema.TypeInt,
				Optional:    true,
			},

			"session_remember_minutes": {
				Description: "Session expiration. Defaults to `20160`.",
				Type:        schema.TypeInt,
				Optional:    true,
			},

			"collaborator_auth_policy": {
				Description: "Authentication policy (`password` or `two_factor_mandatory`). Defaults to `password`.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     string(tfe.AuthPolicyPassword),
				ValidateFunc: validation.StringInSlice(
					[]string{
						string(tfe.AuthPolicyPassword),
						string(tfe.AuthPolicyTwoFactor),
					},
					false,
				),
			},

			"owners_team_saml_role_id": {
				Description: "A SAML attribute value used to identify members of the \"owners\" team. When SAML SSO is enabled, users whose SAML role attribute matches this value will be added to the \"owners\" team.",
				Type:        schema.TypeString,
				Optional:    true,
			},

			"cost_estimation_enabled": {
				Description: "Whether or not the cost estimation feature is enabled for all workspaces in the organization. Defaults to true. In a HCP Terraform organization which does not have Teams & Governance features, this value is always false and cannot be changed. In Terraform Enterprise, Cost Estimation must also be enabled in Site Administration.",
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
			},

			"send_passing_statuses_for_untriggered_speculative_plans": {
				Description: "Whether or not to send VCS status updates for untriggered speculative plans. This can be useful if large numbers of untriggered workspaces are exhausting request limits for connected version control service providers like GitHub. Defaults to false. In Terraform Enterprise, this setting has no effect and cannot be changed but is also available in Site Administration.",
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
			},

			"aggregated_commit_status_enabled": {
				Description: "Whether or not to enable Aggregated Status Checks. This can be useful for monorepo repositories with multiple workspaces receiving status checks for events such as a pull request. If enabled, `send_passing_statuses_for_untriggered_speculative_plans` needs to be `false`.",
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
			},

			"assessments_enforced": {
				Description: "(Available only in HCP Terraform) Whether to force health assessments (drift detection) on all eligible workspaces or allow workspaces to set their own preferences.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"allow_force_delete_workspaces": {
				Description: "Whether workspace administrators are permitted to delete workspaces with resources under management. If false, only organization owners may delete these workspaces. Defaults to `false`.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},

			"default_project_id": {
				Description: "The ID of the organization's default project.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"speculative_plan_management_enabled": {
				Description: "Whether or not to enable Speculative Plan Management. If true, pending VCS-triggered speculative plans from outdated commits will be cancelled if a newer commit is pushed to the same branch.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
			"user_tokens_enabled": {
				Description: "Whether user tokens can be used to read or update the organization. Defaults to true.",
				Type:        schema.TypeBool,
				Default:     true,
				Optional:    true,
			},

			"enforce_hyok": {
				Description: "(Available only in HCP Terraform) Whether HYOK is enabled for all new workspaces in the organization. Defaults to `false`.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"stacks_enabled": {
				Description: "Whether the creation of Stacks are enabled in this organization or not.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"max_ttl_enabled": {
				Description: "Whether maximum token TTL policies are enabled for the organization. When enabled, you can configure maximum TTL values for different token types using the `tfe_org_max_token_ttl_policy` resource. Defaults to `false`.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
		},
	}
}

func resourceTFEOrganizationCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	// Get the organization name.
	name := d.Get("name").(string)
	email := d.Get("email").(string)

	body := models.NewOrganizationsEnvelope()
	org := models.NewOrganizations()
	orgType := models.ORGANIZATIONS_ORGANIZATIONS_TYPE
	org.SetTypeEscaped(&orgType)
	attrs := models.NewOrganizations_attributes()
	attrs.SetName(&name)
	attrs.SetEmail(&email)
	org.SetAttributes(attrs)
	body.SetData(org)

	log.Printf("[DEBUG] Create new organization: %s", name)
	env, err := config.ClientV2.API.Organizations().Post(ctx, body, nil)
	if err != nil {
		return fmt.Errorf("Error creating the new organization %s: %w", name, err)
	}

	orgData := env.GetData()
	if orgData == nil {
		return fmt.Errorf("Error creating the new organization %s: API returned no data", name)
	}

	d.SetId(valueOrZero(orgData.GetId()))

	return resourceTFEOrganizationUpdate(d, meta)
}

func resourceTFEOrganizationRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Read configuration of organization: %s", d.Id())
	env, err := config.ClientV2.API.Organizations().ByOrganization_name(d.Id()).Get(ctx, nil)
	if err != nil {
		if errors.Is(err, tfev2.ErrNotFound) {
			log.Printf("[DEBUG] Organization %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	orgData := env.GetData()
	if orgData == nil {
		log.Printf("[DEBUG] Organization %s no longer exists", d.Id())
		d.SetId("")
		return nil
	}

	// Update the config.
	if attrs := orgData.GetAttributes(); attrs != nil {
		d.Set("name", valueOrZero(attrs.GetName()))
		d.Set("email", valueOrZero(attrs.GetEmail()))
		d.Set("session_timeout_minutes", int(valueOrZero(attrs.GetSessionTimeout())))
		d.Set("session_remember_minutes", int(valueOrZero(attrs.GetSessionRemember())))
		d.Set("collaborator_auth_policy", enumStringOrEmpty(attrs.GetCollaboratorAuthPolicy()))
		d.Set("owners_team_saml_role_id", valueOrZero(attrs.GetOwnersTeamSamlRoleId()))
		d.Set("cost_estimation_enabled", valueOrZero(attrs.GetCostEstimationEnabled()))
		d.Set("send_passing_statuses_for_untriggered_speculative_plans", valueOrZero(attrs.GetSendPassingStatusesForUntriggeredSpeculativePlans()))
		d.Set("aggregated_commit_status_enabled", valueOrZero(attrs.GetAggregatedCommitStatusEnabled()))
		// TFE (onprem) does not currently have this feature and this value won't be returned in those cases.
		// AssessmentsEnforced will default to false
		d.Set("assessments_enforced", valueOrZero(attrs.GetAssessmentsEnforced()))
		d.Set("allow_force_delete_workspaces", valueOrZero(attrs.GetAllowForceDeleteWorkspaces()))
		d.Set("speculative_plan_management_enabled", valueOrZero(attrs.GetSpeculativePlanManagementEnabled()))
		d.Set("stacks_enabled", valueOrZero(attrs.GetStacksEnabled()))
	}

	if rel := orgData.GetRelationships(); rel != nil {
		if dp := rel.GetDefaultProject(); dp != nil && dp.GetData() != nil {
			d.Set("default_project_id", valueOrZero(dp.GetData().GetId()))
		}
	}

	// enforce_hyok, max_ttl_enabled, and user_tokens_enabled have no v2
	// generated getter (go-tfe/v2 gap); backfill via a narrow v1 read.
	if err := readOrganizationEnterpriseFieldsV1(d, config); err != nil {
		return err
	}

	return nil
}

// readOrganizationEnterpriseFieldsV1 backfills enforce_hyok, max_ttl_enabled,
// and user_tokens_enabled from the v1 client. Organizations_attributes in the
// pinned go-tfe/v2 generated client has no getter for these three fields.
func readOrganizationEnterpriseFieldsV1(d *schema.ResourceData, config ConfiguredClient) error {
	org, err := config.Client.Organizations.Read(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return nil
		}
		return err
	}

	d.Set("enforce_hyok", org.EnforceHYOK)
	d.Set("max_ttl_enabled", org.MaxTTLEnabled)
	if org.UserTokensEnabled != nil {
		d.Set("user_tokens_enabled", org.UserTokensEnabled)
	}

	return nil
}

func resourceTFEOrganizationUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	name := d.Get("name").(string)
	email := d.Get("email").(string)

	attrs := models.NewOrganizations_attributes()
	attrs.SetName(&name)
	attrs.SetEmail(&email)

	// If session_timeout is supplied, set it using the attributes.
	if sessionTimeout, ok := d.GetOk("session_timeout_minutes"); ok {
		attrs.SetSessionTimeout(ptr(int32(sessionTimeout.(int)))) //nolint:gosec // session_timeout_minutes is a small user-supplied duration, never near int32 overflow
	}

	// If session_remember is supplied, set it using the attributes.
	if sessionRemember, ok := d.GetOk("session_remember_minutes"); ok {
		attrs.SetSessionRemember(ptr(int32(sessionRemember.(int)))) //nolint:gosec // session_remember_minutes is a small user-supplied duration, never near int32 overflow
	}

	// If collaborator_auth_policy is supplied, set it using the attributes.
	if authPolicy, ok := d.GetOk("collaborator_auth_policy"); ok {
		parsed, err := models.ParseOrganizations_attributes_collaboratorAuthPolicy(authPolicy.(string))
		if err != nil || parsed == nil {
			return fmt.Errorf("Error parsing collaborator_auth_policy %q: %w", authPolicy.(string), err)
		}
		policy := parsed.(*models.Organizations_attributes_collaboratorAuthPolicy)
		attrs.SetCollaboratorAuthPolicy(policy)
	}

	// If owners_team_saml_role_id is supplied, set it using the attributes.
	if ownersTeamSAMLRoleID, ok := d.GetOk("owners_team_saml_role_id"); ok {
		attrs.SetOwnersTeamSamlRoleId(ptr(ownersTeamSAMLRoleID.(string)))
	}

	// If cost_estimation_enabled is supplied, set it using the attributes.
	if costEstimationEnabled, ok := d.GetOkExists("cost_estimation_enabled"); ok {
		attrs.SetCostEstimationEnabled(ptr(costEstimationEnabled.(bool)))
	}

	// If send_passing_statuses_for_untriggered_speculative_plans is supplied, set it using the attributes.
	if d.HasChange("send_passing_statuses_for_untriggered_speculative_plans") {
		_, newVal := d.GetChange("send_passing_statuses_for_untriggered_speculative_plans")
		attrs.SetSendPassingStatusesForUntriggeredSpeculativePlans(ptr(newVal.(bool)))
	}

	// If aggregated_commit_status_enabled is supplied, set it using the attributes.
	if d.HasChange("aggregated_commit_status_enabled") {
		_, newVal := d.GetChange("aggregated_commit_status_enabled")
		attrs.SetAggregatedCommitStatusEnabled(ptr(newVal.(bool)))
	}

	// If assessments_enforced is supplied, set it using the attributes.
	if assessmentsEnforced, ok := d.GetOkExists("assessments_enforced"); ok {
		attrs.SetAssessmentsEnforced(ptr(assessmentsEnforced.(bool)))
	}

	// If allow_force_delete_workspaces is supplied, set it using the attributes.
	if allowForceDeleteWorkspaces, ok := d.GetOkExists("allow_force_delete_workspaces"); ok {
		attrs.SetAllowForceDeleteWorkspaces(ptr(allowForceDeleteWorkspaces.(bool)))
	}

	// If speculative_plan_management_enabled is supplied, set it using the attributes.
	if speculativePlanManagementEnabled, ok := d.GetOkExists("speculative_plan_management_enabled"); ok {
		attrs.SetSpeculativePlanManagementEnabled(ptr(speculativePlanManagementEnabled.(bool)))
	}

	// If speculative_plan_management_enabled is supplied, set it using the attributes.
	if stacksEnabled, ok := d.GetOkExists("stacks_enabled"); ok {
		attrs.SetStacksEnabled(ptr(stacksEnabled.(bool)))
	}

	orgType := models.ORGANIZATIONS_ORGANIZATIONS_TYPE
	org := models.NewOrganizations()
	org.SetTypeEscaped(&orgType)
	org.SetAttributes(attrs)
	body := models.NewOrganizationsEnvelope()
	body.SetData(org)

	log.Printf("[DEBUG] Update configuration of organization: %s", d.Id())
	env, err := config.ClientV2.API.Organizations().ByOrganization_name(d.Id()).Patch(ctx, body, nil)
	if err != nil {
		return fmt.Errorf("Error updating organization %s: %w", d.Id(), err)
	}

	orgData := env.GetData()
	if orgData == nil {
		return fmt.Errorf("Error updating organization %s: API returned no data", d.Id())
	}
	d.SetId(valueOrZero(orgData.GetId()))

	// enforce_hyok, max_ttl_enabled, and user_tokens_enabled have no v2
	// generated setter (go-tfe/v2 gap); update via a narrow v1 call.
	if err := updateOrganizationEnterpriseFieldsV1(d, config); err != nil {
		return err
	}

	return resourceTFEOrganizationRead(d, meta)
}

// updateOrganizationEnterpriseFieldsV1 updates enforce_hyok, max_ttl_enabled,
// and user_tokens_enabled via the v1 client. Organizations_attributes in the
// pinned go-tfe/v2 generated client has no setter for these three fields.
func updateOrganizationEnterpriseFieldsV1(d *schema.ResourceData, config ConfiguredClient) error {
	options := tfe.OrganizationUpdateOptions{}

	if userTokensEnabled, ok := d.GetOkExists("user_tokens_enabled"); ok {
		options.UserTokensEnabled = tfe.Bool(userTokensEnabled.(bool))
	}
	if enforceHYOK, ok := d.GetOkExists("enforce_hyok"); ok {
		options.EnforceHYOK = tfe.Bool(enforceHYOK.(bool))
	}
	if maxTTLEnabled, ok := d.GetOkExists("max_ttl_enabled"); ok {
		options.MaxTTLEnabled = tfe.Bool(maxTTLEnabled.(bool))
	}

	_, err := config.Client.Organizations.Update(ctx, d.Id(), options)
	if err != nil {
		return fmt.Errorf("Error updating organization %s: %w", d.Id(), err)
	}

	return nil
}

func resourceTFEOrganizationDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Delete organization: %s", d.Id())
	err := config.ClientV2.API.Organizations().ByOrganization_name(d.Id()).Delete(ctx, nil)
	if err != nil {
		if errors.Is(err, tfev2.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("Error deleting organization %s: %w", d.Id(), err)
	}

	return nil
}
