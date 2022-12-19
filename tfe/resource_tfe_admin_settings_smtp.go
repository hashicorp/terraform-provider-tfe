package tfe

import (
	"fmt"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTFEAdminSettingsSMTP() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFEAdminSettingsSMTPUpdate,
		Read:   resourceTFEAdminSettingsSMTPRead,
		Update: resourceTFEAdminSettingsSMTPUpdate,
		Delete: resourceTFEAdminSettingsSMTPDelete,
		Schema: map[string]*schema.Schema{
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"host": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"port": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"sender": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"auth": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "none",
			},
			"username": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"password": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"test_email_address": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func smtpAuthType(v string) *tfe.SMTPAuthType {
	t := tfe.SMTPAuthType(v)
	return &t
}

func resourceTFEAdminSettingsSMTPUpdate(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	opts := tfe.AdminSMTPSettingsUpdateOptions{
		Enabled: tfe.Bool(d.Get("enabled").(bool)),
		Auth:    smtpAuthType(d.Get("auth").(string)),
	}

	if v, ok := d.GetOk("host"); ok {
		opts.Host = tfe.String(v.(string))
	}
	if v, ok := d.GetOk("port"); ok {
		opts.Port = tfe.Int(v.(int))
	}
	if v, ok := d.GetOk("sender"); ok {
		opts.Sender = tfe.String(v.(string))
	}
	if v, ok := d.GetOk("username"); ok {
		opts.Username = tfe.String(v.(string))
	}
	if v, ok := d.GetOk("password"); ok {
		opts.Password = tfe.String(v.(string))
	}
	if v, ok := d.GetOk("test_email_address"); ok {
		opts.TestEmailAddress = tfe.String(v.(string))
	}

	_, err := tfeClient.Admin.Settings.SMTP.Update(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to update SMTP settings: %s", err)
	}

	d.SetId("settings-smtp")
	return resourceTFEAdminSettingsSMTPRead(d, meta)
}

func resourceTFEAdminSettingsSMTPRead(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	settings, err := tfeClient.Admin.Settings.SMTP.Read(ctx)
	if err != nil {
		return fmt.Errorf("failed to read SMTP settings: %s", err)
	}

	d.Set("enabled", settings.Enabled)
	d.Set("host", settings.Host)
	d.Set("port", settings.Port)
	d.Set("sender", settings.Sender)
	d.Set("auth", settings.Auth)
	d.Set("username", settings.Username)

	return nil
}

func resourceTFEAdminSettingsSMTPDelete(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	_, err := tfeClient.Admin.Settings.SMTP.Update(ctx, tfe.AdminSMTPSettingsUpdateOptions{
		Enabled: tfe.Bool(false),
	})
	if err != nil {
		return fmt.Errorf("failed to delete SMTP settings: %s", err)
	}

	return nil
}
