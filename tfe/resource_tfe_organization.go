package tfe

import (
	"fmt"
	"log"
	"strings"

	tfe "github.com/hashicorp/go-tfe"
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
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Name of the organization.",
				Type:        schema.TypeString,
				Required:    true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return strings.EqualFold(old, new)
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
				Description: "The name of the owners team.",
				Type:        schema.TypeString,
				Optional:    true,
			},

			"cost_estimation_enabled": {
				Description: "Whether or not the cost estimation feature is enabled for all workspaces in the organization. Defaults to true. In a Terraform Cloud organization which does not have Teams & Governance features, this value is always false and cannot be changed. In Terraform Enterprise, Cost Estimation must also be enabled in Site Administration.",
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
		},
	}
}

func resourceTFEOrganizationCreate(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	// Get the organization name.
	name := d.Get("name").(string)

	// Create a new options struct.
	options := tfe.OrganizationCreateOptions{
		Name:  tfe.String(name),
		Email: tfe.String(d.Get("email").(string)),
	}

	log.Printf("[DEBUG] Create new organization: %s", name)
	org, err := tfeClient.Organizations.Create(ctx, options)
	if err != nil {
		return fmt.Errorf("Error creating the new organization %s: %w", name, err)
	}

	d.SetId(org.Name)

	return resourceTFEOrganizationUpdate(d, meta)
}

func resourceTFEOrganizationRead(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	log.Printf("[DEBUG] Read configuration of organization: %s", d.Id())
	org, err := tfeClient.Organizations.Read(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			log.Printf("[DEBUG] Organization %s does no longer exist", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	// Update the config.
	d.Set("name", org.Name)
	d.Set("email", org.Email)
	d.Set("session_timeout_minutes", org.SessionTimeout)
	d.Set("session_remember_minutes", org.SessionRemember)
	d.Set("collaborator_auth_policy", org.CollaboratorAuthPolicy)
	d.Set("owners_team_saml_role_id", org.OwnersTeamSAMLRoleID)
	d.Set("cost_estimation_enabled", org.CostEstimationEnabled)
	d.Set("send_passing_statuses_for_untriggered_speculative_plans", org.SendPassingStatusesForUntriggeredSpeculativePlans)

	return nil
}

func resourceTFEOrganizationUpdate(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	// Create a new options struct.
	options := tfe.OrganizationUpdateOptions{
		Name:  tfe.String(d.Get("name").(string)),
		Email: tfe.String(d.Get("email").(string)),
	}

	// If session_timeout is supplied, set it using the options struct.
	if sessionTimeout, ok := d.GetOk("session_timeout_minutes"); ok {
		options.SessionTimeout = tfe.Int(sessionTimeout.(int))
	}

	// If session_remember is supplied, set it using the options struct.
	if sessionRemember, ok := d.GetOk("session_remember_minutes"); ok {
		options.SessionRemember = tfe.Int(sessionRemember.(int))
	}

	// If collaborator_auth_policy is supplied, set it using the options struct.
	if authPolicy, ok := d.GetOk("collaborator_auth_policy"); ok {
		options.CollaboratorAuthPolicy = tfe.AuthPolicy(tfe.AuthPolicyType(authPolicy.(string)))
	}

	// If owners_team_saml_role_id is supplied, set it using the options struct.
	if ownersTeamSAMLRoleID, ok := d.GetOk("owners_team_saml_role_id"); ok {
		options.OwnersTeamSAMLRoleID = tfe.String(ownersTeamSAMLRoleID.(string))
	}

	// If cost_estimation_enabled is supplied, set it using the options struct.
	if costEstimationEnabled, ok := d.GetOkExists("cost_estimation_enabled"); ok {
		options.CostEstimationEnabled = tfe.Bool(costEstimationEnabled.(bool))
	}

	// If send_passing_statuses_for_untriggered_speculative_plans is supplied, set it using the options struct.
	if sendPassingStatusesForUntriggeredSpeculativePlans, ok := d.GetOk("send_passing_statuses_for_untriggered_speculative_plans"); ok {
		options.SendPassingStatusesForUntriggeredSpeculativePlans = tfe.Bool(sendPassingStatusesForUntriggeredSpeculativePlans.(bool))
	}

	log.Printf("[DEBUG] Update configuration of organization: %s", d.Id())
	org, err := tfeClient.Organizations.Update(ctx, d.Id(), options)
	if err != nil {
		return fmt.Errorf("Error updating organization %s: %w", d.Id(), err)
	}

	d.SetId(org.Name)

	return resourceTFEOrganizationRead(d, meta)
}

func resourceTFEOrganizationDelete(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	log.Printf("[DEBUG] Delete organization: %s", d.Id())
	err := tfeClient.Organizations.Delete(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return nil
		}
		return fmt.Errorf("Error deleting organization %s: %w", d.Id(), err)
	}

	return nil
}
