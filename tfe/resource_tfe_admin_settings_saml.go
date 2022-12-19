package tfe

import (
	"fmt"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTFEAdminSettingsSAML() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFEAdminSettingsSAMLUpdate,
		Read:   resourceTFEAdminSettingsSAMLRead,
		Update: resourceTFEAdminSettingsSAMLUpdate,
		Delete: resourceTFEAdminSettingsSAMLDelete,
		Schema: map[string]*schema.Schema{
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"debug": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"idp_cert": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"slo_endpoint_url": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"sso_endpoint_url": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"attr_username": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Username",
			},
			"attr_groups": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "MemberOf",
			},
			"attr_site_admin": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "SiteAdmin",
			},
			"site_admin_role": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "site-admins",
			},
			"sso_api_token_session_timeout": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1209600,
			},
			"acs_consumer_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"metadata_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceTFEAdminSettingsSAMLUpdate(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	opts := tfe.AdminSAMLSettingsUpdateOptions{
		Enabled:                   tfe.Bool(d.Get("enabled").(bool)),
		Debug:                     tfe.Bool(d.Get("debug").(bool)),
		AttrUsername:              tfe.String(d.Get("attr_username").(string)),
		AttrGroups:                tfe.String(d.Get("attr_groups").(string)),
		AttrSiteAdmin:             tfe.String(d.Get("attr_site_admin").(string)),
		SiteAdminRole:             tfe.String(d.Get("site_admin_role").(string)),
		SSOAPITokenSessionTimeout: tfe.Int(d.Get("sso_api_token_session_timeout").(int)),
	}

	if v, ok := d.GetOk("idp_cert"); ok {
		opts.IDPCert = tfe.String(v.(string))
	}
	if v, ok := d.GetOk("slo_endpoint_url"); ok {
		opts.SLOEndpointURL = tfe.String(v.(string))
	}
	if v, ok := d.GetOk("sso_endpoint_url"); ok {
		opts.SSOEndpointURL = tfe.String(v.(string))
	}

	_, err := tfeClient.Admin.Settings.SAML.Update(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to update SAML settings: %s", err)
	}

	d.SetId("settings-saml")
	return resourceTFEAdminSettingsSAMLRead(d, meta)
}

func resourceTFEAdminSettingsSAMLRead(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	settings, err := tfeClient.Admin.Settings.SAML.Read(ctx)
	if err != nil {
		return fmt.Errorf("failed to read SAML settings: %s", err)
	}

	d.Set("enabled", settings.Enabled)
	d.Set("debug", settings.Debug)
	d.Set("idp_cert", settings.IDPCert)
	d.Set("slo_endpoint_url", settings.SLOEndpointURL)
	d.Set("sso_endpoint_url", settings.SSOEndpointURL)
	d.Set("attr_username", settings.AttrUsername)
	d.Set("attr_groups", settings.AttrGroups)
	d.Set("attr_site_admin", settings.AttrSiteAdmin)
	d.Set("site_admin_role", settings.SiteAdminRole)
	d.Set("sso_api_token_session_timeout", settings.SSOAPITokenSessionTimeout)
	d.Set("acs_consumer_url", settings.ACSConsumerURL)
	d.Set("metadata_url", settings.MetadataURL)

	return nil
}

func resourceTFEAdminSettingsSAMLDelete(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	_, err := tfeClient.Admin.Settings.SAML.Update(ctx, tfe.AdminSAMLSettingsUpdateOptions{
		Enabled: tfe.Bool(false),
	})
	if err != nil {
		return fmt.Errorf("failed to delete SAML settings: %s", err)
	}

	return nil
}
