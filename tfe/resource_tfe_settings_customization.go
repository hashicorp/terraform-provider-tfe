package tfe

import (
	"fmt"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTFESettingsCustomization() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFESettingsCustomizationUpdate,
		Read:   resourceTFESettingsCustomizationRead,
		Update: resourceTFESettingsCustomizationUpdate,
		Delete: resourceTFESettingsCustomizationDelete,
		Schema: map[string]*schema.Schema{
			"support_email_address": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "support@hashicorp.com",
			},
			"login_help": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"footer": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"error": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"new_user": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
		},
	}
}

func resourceTFESettingsCustomizationUpdate(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	opts := tfe.AdminCustomizationSettingsUpdateOptions{
		SupportEmail: tfe.String(d.Get("support_email_address").(string)),
		LoginHelp:    tfe.String(d.Get("login_help").(string)),
		Footer:       tfe.String(d.Get("footer").(string)),
		Error:        tfe.String(d.Get("error").(string)),
		NewUser:      tfe.String(d.Get("new_user").(string)),
	}

	_, err := tfeClient.Admin.Settings.Customization.Update(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to update Customization settings: %s", err)
	}

	d.SetId("settings-customization")
	return resourceTFESettingsCustomizationRead(d, meta)
}

func resourceTFESettingsCustomizationRead(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	settings, err := tfeClient.Admin.Settings.Customization.Read(ctx)
	if err != nil {
		return fmt.Errorf("failed to read Customization settings: %s", err)
	}

	d.Set("support_email_address", settings.SupportEmail)
	d.Set("login_help", settings.LoginHelp)
	d.Set("footer", settings.Footer)
	d.Set("error", settings.Error)
	d.Set("new_user", settings.NewUser)

	return nil
}

func resourceTFESettingsCustomizationDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}
