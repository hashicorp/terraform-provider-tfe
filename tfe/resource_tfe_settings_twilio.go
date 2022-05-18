package tfe

import (
	"fmt"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTFESettingsTwilio() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFESettingsTwilioUpdate,
		Read:   resourceTFESettingsTwilioRead,
		Update: resourceTFESettingsTwilioUpdate,
		Delete: resourceTFESettingsTwilioDelete,
		Schema: map[string]*schema.Schema{
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"account_sid": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"from_number": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"auth_token": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceTFESettingsTwilioUpdate(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	opts := tfe.AdminTwilioSettingsUpdateOptions{
		Enabled: tfe.Bool(d.Get("enabled").(bool)),
	}

	if v, ok := d.GetOk("account_sid"); ok {
		opts.AccountSid = tfe.String(v.(string))
	}
	if v, ok := d.GetOk("from_number"); ok {
		opts.FromNumber = tfe.String(v.(string))
	}
	if v, ok := d.GetOk("auth_token"); ok {
		opts.AuthToken = tfe.String(v.(string))
	}

	_, err := tfeClient.Admin.Settings.Twilio.Update(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to update Twilio settings: %s", err)
	}

	d.SetId("settings-twilio")
	return resourceTFESettingsTwilioRead(d, meta)
}

func resourceTFESettingsTwilioRead(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	settings, err := tfeClient.Admin.Settings.Twilio.Read(ctx)
	if err != nil {
		return fmt.Errorf("failed to read Twilio settings: %s", err)
	}

	d.Set("enabled", settings.Enabled)
	d.Set("account_sid", settings.AccountSid)
	d.Set("from_number", settings.FromNumber)

	return nil
}

func resourceTFESettingsTwilioDelete(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	_, err := tfeClient.Admin.Settings.Twilio.Update(ctx, tfe.AdminTwilioSettingsUpdateOptions{
		Enabled: tfe.Bool(false),
	})
	if err != nil {
		return fmt.Errorf("failed to delete Twilio settings: %s", err)
	}

	return nil
}
