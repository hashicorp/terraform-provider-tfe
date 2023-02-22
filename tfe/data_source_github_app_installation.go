// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package tfe

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTFEGHAInstallation() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGHAInstallationRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:         schema.TypeString,
				Optional:     true,
				AtLeastOneOf: []string{"id", "name", "github_installation_id"},
			},
			"github_installation_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func dataSourceGHAInstallationRead(d *schema.ResourceData, meta interface{}) error {
	ctx := context.TODO()
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Reading github app installations")

	var ghai *tfe.GHAInstallation
	var err error

	switch v, ok := d.GetOk("id"); {
	case ok:
		ghai, err = config.Client.GHAInstallations.Read(ctx, v.(string))
		if err != nil {
			return fmt.Errorf("Error retrieving Github App Installation: %w", err)
		}
	default:
		// search by name or id
		if err != nil {
			return err
		}

		var name string
		var GHInstallationID int
		vName, ok := d.GetOk("name")
		if ok {
			name = vName.(string)
		}
		vInstallationId, ok := d.GetOk("github_installation_id")
		if ok {
			GHInstallationID = vInstallationId.(int)
		}
		ghai, err = fetchGithubAppInstallationByNameOrGHID(ctx, config.Client, name, GHInstallationID)
		if err != nil {
			return err
		}
	}

	d.SetId(ghai.ID)
	d.Set("id", ghai.ID)
	d.Set("github_installation_id", ghai.GHInstallationId)
	d.Set("name", ghai.Name)

	return nil
}
