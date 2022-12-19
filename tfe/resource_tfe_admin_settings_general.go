package tfe

import (
	"fmt"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTFEAdminSettingsGeneral() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFEAdminSettingsGeneralUpdate,
		Read:   resourceTFEAdminSettingsGeneralRead,
		Update: resourceTFEAdminSettingsGeneralUpdate,
		Delete: resourceTFEAdminSettingsGeneralDelete,
		Schema: map[string]*schema.Schema{
			"limit_user_organization_creation": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"api_rate_limiting_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"api_rate_limit": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  30,
			},
			"send_passing_statuses_for_untriggered_speculative_plans": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"allow_speculative_plans_on_pull_requests_from_forks": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"default_remote_state_access": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
		},
	}
}

func resourceTFEAdminSettingsGeneralUpdate(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	opts := tfe.AdminGeneralSettingsUpdateOptions{
		LimitUserOrgCreation:              tfe.Bool(d.Get("limit_user_organization_creation").(bool)),
		APIRateLimitingEnabled:            tfe.Bool(d.Get("api_rate_limiting_enabled").(bool)),
		APIRateLimit:                      tfe.Int(d.Get("api_rate_limit").(int)),
		SendPassingStatusUntriggeredPlans: tfe.Bool(d.Get("send_passing_statuses_for_untriggered_speculative_plans").(bool)),
		AllowSpeculativePlansOnPR:         tfe.Bool(d.Get("allow_speculative_plans_on_pull_requests_from_forks").(bool)),
		DefaultRemoteStateAccess:          tfe.Bool(d.Get("default_remote_state_access").(bool)),
	}

	_, err := tfeClient.Admin.Settings.General.Update(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to update general settings: %s", err)
	}

	d.SetId("settings-general")
	return resourceTFEAdminSettingsGeneralRead(d, meta)
}

func resourceTFEAdminSettingsGeneralRead(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	settings, err := tfeClient.Admin.Settings.General.Read(ctx)
	if err != nil {
		return fmt.Errorf("failed to read general settings: %s", err)
	}

	d.Set("limit_user_organization_creation", settings.LimitUserOrganizationCreation)
	d.Set("api_rate_limiting_enabled", settings.APIRateLimitingEnabled)
	d.Set("api_rate_limit", settings.APIRateLimit)
	d.Set("send_passing_statuses_for_untriggered_speculative_plans", settings.SendPassingStatusesEnabled)
	d.Set("allow_speculative_plans_on_pull_requests_from_forks", settings.AllowSpeculativePlansOnPR)
	d.Set("default_remote_state_access", settings.DefaultRemoteStateAccess)

	return nil
}

func resourceTFEAdminSettingsGeneralDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}
