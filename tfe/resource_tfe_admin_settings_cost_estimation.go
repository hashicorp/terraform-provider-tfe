package tfe

import (
	"fmt"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTFEAdminSettingsCostEstimation() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFEAdminSettingsCostEstimationUpdate,
		Read:   resourceTFEAdminSettingsCostEstimationRead,
		Update: resourceTFEAdminSettingsCostEstimationUpdate,
		Delete: resourceTFEAdminSettingsCostEstimationDelete,
		Schema: map[string]*schema.Schema{
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"aws_access_key_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"aws_secret_key": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"aws_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"aws_instance_profile_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"gcp_credentials": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"gcp_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"azure_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"azure_client_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"azure_client_secret": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"azure_subscription_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"azure_tenant_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceTFEAdminSettingsCostEstimationUpdate(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	opts := tfe.AdminCostEstimationSettingOptions{
		Enabled: tfe.Bool(d.Get("enabled").(bool)),
	}

	if v, ok := d.GetOk("aws_access_key_id"); ok {
		opts.AWSAccessKeyID = tfe.String(v.(string))
	}
	if v, ok := d.GetOk("aws_secret_key"); ok {
		opts.AWSAccessKey = tfe.String(v.(string))
	}
	if v, ok := d.GetOk("gcp_credentials"); ok {
		opts.GCPCredentials = tfe.String(v.(string))
	}
	if v, ok := d.GetOk("azure_client_id"); ok {
		opts.AzureClientID = tfe.String(v.(string))
	}
	if v, ok := d.GetOk("azure_client_secret"); ok {
		opts.AzureClientSecret = tfe.String(v.(string))
	}
	if v, ok := d.GetOk("azure_subscription_id"); ok {
		opts.AzureSubscriptionID = tfe.String(v.(string))
	}
	if v, ok := d.GetOk("azure_tenant_id"); ok {
		opts.AzureTenantID = tfe.String(v.(string))
	}

	_, err := tfeClient.Admin.Settings.CostEstimation.Update(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to update Cost Estimation settings: %s", err)
	}

	d.SetId("settings-cost-estimation")
	return resourceTFEAdminSettingsCostEstimationRead(d, meta)
}

func resourceTFEAdminSettingsCostEstimationRead(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	settings, err := tfeClient.Admin.Settings.CostEstimation.Read(ctx)
	if err != nil {
		return fmt.Errorf("failed to read Cost Estimation settings: %s", err)
	}

	d.Set("enabled", settings.Enabled)
	d.Set("aws_access_key_id", settings.AWSAccessKeyID)
	d.Set("aws_secret_key", settings.AWSAccessKey)
	d.Set("aws_enabled", settings.AWSEnabled)
	d.Set("aws_instance_profile_enabled", settings.AWSInstanceProfileEnabled)
	d.Set("gcp_credentials", settings.GCPCredentials)
	d.Set("gcp_enabled", settings.GCPEnabled)
	d.Set("azure_enabled", settings.AzureEnabled)
	d.Set("azure_client_id", settings.AzureClientID)
	d.Set("azure_client_secret", settings.AzureClientSecret)
	d.Set("azure_subscription_id", settings.AzureSubscriptionID)
	d.Set("azure_tenant_id", settings.AzureTenantID)

	return nil
}

func resourceTFEAdminSettingsCostEstimationDelete(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	_, err := tfeClient.Admin.Settings.Twilio.Update(ctx, tfe.AdminTwilioSettingsUpdateOptions{
		Enabled: tfe.Bool(false),
	})
	if err != nil {
		return fmt.Errorf("failed to delete Cost Estimation settings: %s", err)
	}

	return nil
}
