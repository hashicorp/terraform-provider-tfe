// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTFESentinelVersion() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFESentinelVersionCreate,
		Read:   resourceTFESentinelVersionRead,
		Update: resourceTFESentinelVersionUpdate,
		Delete: resourceTFESentinelVersionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceTFESentinelVersionImporter,
		},

		Schema: map[string]*schema.Schema{
			"version": {
				Type:     schema.TypeString,
				Required: true,
			},
			"url": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"sha": {
				Type:     schema.TypeString,
				Optional: true,
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
			"archs": {
				Type:     schema.TypeList,
				Optional: true,
				Default:  nil,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"url": {
							Type:     schema.TypeString,
							Required: true,
						},
						"sha": {
							Type:     schema.TypeString,
							Required: true,
						},
						"os": {
							Type:     schema.TypeString,
							Required: true,
						},
						"arch": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func resourceTFESentinelVersionCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	opts := tfe.AdminSentinelVersionCreateOptions{
		Version:          d.Get("version").(string),
		URL:              d.Get("url").(string),
		SHA:              d.Get("sha").(string),
		Official:         tfe.Bool(d.Get("official").(bool)),
		Enabled:          tfe.Bool(d.Get("enabled").(bool)),
		Beta:             tfe.Bool(d.Get("beta").(bool)),
		Deprecated:       tfe.Bool(d.Get("deprecated").(bool)),
		DeprecatedReason: tfe.String(d.Get("deprecated_reason").(string)),
		Archs:            convertToToolVersionArchitectures(d.Get("archs").([]interface{})),
	}

	log.Printf("[DEBUG] Create new Sentinel version: %s", opts.Version)
	v, err := config.Client.Admin.SentinelVersions.Create(ctx, opts)
	if err != nil {
		return fmt.Errorf("error creating the new Sentinel version %s: %w", opts.Version, err)
	}

	d.SetId(v.ID)

	return resourceTFESentinelVersionRead(d, meta)
}

func resourceTFESentinelVersionRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Read configuration of Sentinel version: %s", d.Id())
	v, err := config.Client.Admin.SentinelVersions.Read(ctx, d.Id())
	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
			log.Printf("[DEBUG] Sentinel version %s no longer exists", d.Id())
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
	d.Set("archs", convertToToolVersionArchitecturesMap(v.Archs))

	return nil
}

func resourceTFESentinelVersionUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	opts := tfe.AdminSentinelVersionUpdateOptions{
		Version:          tfe.String(d.Get("version").(string)),
		URL:              tfe.String(d.Get("url").(string)),
		SHA:              tfe.String(d.Get("sha").(string)),
		Official:         tfe.Bool(d.Get("official").(bool)),
		Enabled:          tfe.Bool(d.Get("enabled").(bool)),
		Beta:             tfe.Bool(d.Get("beta").(bool)),
		Deprecated:       tfe.Bool(d.Get("deprecated").(bool)),
		DeprecatedReason: tfe.String(d.Get("deprecated_reason").(string)),
		Archs:            convertToToolVersionArchitectures(d.Get("archs").([]interface{})),
	}

	log.Printf("[DEBUG] Update configuration of Sentinel version: %s", d.Id())
	v, err := config.Client.Admin.SentinelVersions.Update(ctx, d.Id(), opts)
	if err != nil {
		return fmt.Errorf("error updating Sentinel version %s: %w", d.Id(), err)
	}

	d.SetId(v.ID)

	return resourceTFESentinelVersionRead(d, meta)
}

func resourceTFESentinelVersionDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Delete Sentinel version: %s", d.Id())
	err := config.Client.Admin.SentinelVersions.Delete(ctx, d.Id())
	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
			log.Printf("[DEBUG] Sentinel version: %s not found", d.Id())
			return nil
		}
		return fmt.Errorf("error deleting Sentinel version %s: %w", d.Id(), err)
	}

	return nil
}

func resourceTFESentinelVersionImporter(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	config := meta.(ConfiguredClient)

	// Splitting by '-' and checking if the first elem is equal to tool
	// determines if the string is a tool version ID
	s := strings.Split(d.Id(), "-")
	if s[0] != "tool" {
		versionID, err := fetchSentinelVersionID(d.Id(), config.Client)
		if err != nil {
			return nil, fmt.Errorf("error retrieving sentinel version %s: %w", d.Id(), err)
		}

		d.SetId(versionID)
	}

	return []*schema.ResourceData{d}, nil
}
