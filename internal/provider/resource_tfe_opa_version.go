// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tfe

import (
	"context"
	"fmt"
	"log"
	"strings"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTFEOPAVersion() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFEOPAVersionCreate,
		Read:   resourceTFEOPAVersionRead,
		Update: resourceTFEOPAVersionUpdate,
		Delete: resourceTFEOPAVersionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceTFEOPAVersionImporter,
		},

		Schema: map[string]*schema.Schema{
			"version": {
				Type:     schema.TypeString,
				Required: true,
			},
			"url": {
				Type:     schema.TypeString,
				Required: true,
			},
			"sha": {
				Type:     schema.TypeString,
				Required: true,
			},
			"official": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"beta": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"deprecated": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"deprecated_reason": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  nil,
			},
		},
	}
}

func resourceTFEOPAVersionCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	opts := tfe.AdminOPAVersionCreateOptions{
		Version:          *tfe.String(d.Get("version").(string)),
		URL:              *tfe.String(d.Get("url").(string)),
		SHA:              *tfe.String(d.Get("sha").(string)),
		Official:         tfe.Bool(d.Get("official").(bool)),
		Enabled:          tfe.Bool(d.Get("enabled").(bool)),
		Beta:             tfe.Bool(d.Get("beta").(bool)),
		Deprecated:       tfe.Bool(d.Get("deprecated").(bool)),
		DeprecatedReason: tfe.String(d.Get("deprecated_reason").(string)),
	}

	log.Printf("[DEBUG] Create new OPA version: %s", opts.Version)
	v, err := config.Client.Admin.OPAVersions.Create(ctx, opts)
	if err != nil {
		return fmt.Errorf("Error creating the new OPA version %s: %w", opts.Version, err)
	}

	d.SetId(v.ID)

	return resourceTFEOPAVersionUpdate(d, meta)
}

func resourceTFEOPAVersionRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Read configuration of OPA version: %s", d.Id())
	v, err := config.Client.Admin.OPAVersions.Read(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			log.Printf("[DEBUG] OPA version %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("version", v.Version)
	d.Set("url", v.URL)
	d.Set("sha", v.SHA)
	d.Set("official", v.Official)
	d.Set("enabled", v.Enabled)
	d.Set("beta", v.Beta)
	d.Set("deprecated", v.Deprecated)
	d.Set("deprecated_reason", v.DeprecatedReason)

	return nil
}

func resourceTFEOPAVersionUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	opts := tfe.AdminOPAVersionUpdateOptions{
		Version:          tfe.String(d.Get("version").(string)),
		URL:              tfe.String(d.Get("url").(string)),
		SHA:              tfe.String(d.Get("sha").(string)),
		Official:         tfe.Bool(d.Get("official").(bool)),
		Enabled:          tfe.Bool(d.Get("enabled").(bool)),
		Beta:             tfe.Bool(d.Get("beta").(bool)),
		Deprecated:       tfe.Bool(d.Get("deprecated").(bool)),
		DeprecatedReason: tfe.String(d.Get("deprecated_reason").(string)),
	}

	log.Printf("[DEBUG] Update configuration of OPA version: %s", d.Id())
	v, err := config.Client.Admin.OPAVersions.Update(ctx, d.Id(), opts)
	if err != nil {
		return fmt.Errorf("Error updating OPA version %s: %w", d.Id(), err)
	}

	d.SetId(v.ID)

	return resourceTFEOPAVersionRead(d, meta)
}

func resourceTFEOPAVersionDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Delete OPA version: %s", d.Id())
	err := config.Client.Admin.OPAVersions.Delete(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return nil
		}
		return fmt.Errorf("Error deleting OPA version %s: %w", d.Id(), err)
	}

	return nil
}

func resourceTFEOPAVersionImporter(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	config := meta.(ConfiguredClient)

	// Splitting by '-' and checking if the first elem is equal to tool
	// determines if the string is a tool version ID
	s := strings.Split(d.Id(), "-")
	if s[0] != "tool" {
		versionID, err := fetchOPAVersionID(d.Id(), config.Client)
		if err != nil {
			return nil, fmt.Errorf("error retrieving OPA version %s: %w", d.Id(), err)
		}

		d.SetId(versionID)
	}

	return []*schema.ResourceData{d}, nil
}
