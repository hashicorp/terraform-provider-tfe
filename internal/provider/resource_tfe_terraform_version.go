// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"log"
	"strings"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTFETerraformVersion() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFETerraformVersionCreate,
		Read:   resourceTFETerraformVersionRead,
		Update: resourceTFETerraformVersionUpdate,
		Delete: resourceTFETerraformVersionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceTFETerraformVersionImporter,
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

func resourceTFETerraformVersionCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	opts := tfe.AdminTerraformVersionCreateOptions{
		Version:          tfe.String(d.Get("version").(string)),
		URL:              tfe.String(d.Get("url").(string)),
		Sha:              tfe.String(d.Get("sha").(string)),
		Official:         tfe.Bool(d.Get("official").(bool)),
		Enabled:          tfe.Bool(d.Get("enabled").(bool)),
		Beta:             tfe.Bool(d.Get("beta").(bool)),
		Deprecated:       tfe.Bool(d.Get("deprecated").(bool)),
		DeprecatedReason: tfe.String(d.Get("deprecated_reason").(string)),
	}

	log.Printf("[DEBUG] Create new Terraform version: %s", *opts.Version)
	v, err := config.Client.Admin.TerraformVersions.Create(ctx, opts)
	if err != nil {
		return fmt.Errorf("Error creating the new Terraform version %s: %w", *opts.Version, err)
	}

	d.SetId(v.ID)

	return resourceTFETerraformVersionUpdate(d, meta)
}

func resourceTFETerraformVersionRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Read configuration of Terraform version: %s", d.Id())
	v, err := config.Client.Admin.TerraformVersions.Read(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			log.Printf("[DEBUG] Terraform version %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("version", v.Version)
	d.Set("url", v.URL)
	d.Set("sha", v.Sha)
	d.Set("official", v.Official)
	d.Set("enabled", v.Enabled)
	d.Set("beta", v.Beta)
	d.Set("deprecated", v.Deprecated)
	d.Set("deprecated_reason", v.DeprecatedReason)

	return nil
}

func resourceTFETerraformVersionUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	opts := tfe.AdminTerraformVersionUpdateOptions{
		Version:          tfe.String(d.Get("version").(string)),
		URL:              tfe.String(d.Get("url").(string)),
		Sha:              tfe.String(d.Get("sha").(string)),
		Official:         tfe.Bool(d.Get("official").(bool)),
		Enabled:          tfe.Bool(d.Get("enabled").(bool)),
		Beta:             tfe.Bool(d.Get("beta").(bool)),
		Deprecated:       tfe.Bool(d.Get("deprecated").(bool)),
		DeprecatedReason: tfe.String(d.Get("deprecated_reason").(string)),
	}

	log.Printf("[DEBUG] Update configuration of Terraform version: %s", d.Id())
	v, err := config.Client.Admin.TerraformVersions.Update(ctx, d.Id(), opts)
	if err != nil {
		return fmt.Errorf("Error updating Terraform version %s: %w", d.Id(), err)
	}

	d.SetId(v.ID)

	return resourceTFETerraformVersionRead(d, meta)
}

func resourceTFETerraformVersionDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Delete Terraform version: %s", d.Id())
	err := config.Client.Admin.TerraformVersions.Delete(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return nil
		}
		return fmt.Errorf("Error deleting Terraform version %s: %w", d.Id(), err)
	}

	return nil
}

func resourceTFETerraformVersionImporter(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	config := meta.(ConfiguredClient)

	// Splitting by '-' and checking if the first elem is equal to tool
	// determines if the string is a tool version ID
	s := strings.Split(d.Id(), "-")
	if s[0] != "tool" {
		versionID, err := fetchTerraformVersionID(d.Id(), config.Client)
		if err != nil {
			return nil, fmt.Errorf("error retrieving terraform version %s: %w", d.Id(), err)
		}

		d.SetId(versionID)
	}

	return []*schema.ResourceData{d}, nil
}
