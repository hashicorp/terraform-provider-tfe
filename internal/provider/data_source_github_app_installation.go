// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package provider

import (
	"context"
	"log"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTFEGHAInstallation() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGHAInstallationRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"installation_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				AtLeastOneOf: []string{"name", "installation_id"},
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

	log.Printf("[DEBUG] Reading github app installation")

	var ghai *tfe.GHAInstallation
	var err error

	// search by name or installation_id
	var name string
	var GHInstallationID int
	vName, ok := d.GetOk("name")
	if ok {
		name = vName.(string)
	}

	vInstallationID, ok := d.GetOk("installation_id")
	if ok {
		GHInstallationID = vInstallationID.(int)
	}
	ghai, err = fetchGithubAppInstallationByNameOrGHID(ctx, config.Client, name, GHInstallationID)
	if err != nil {
		return err
	}

	d.SetId(*ghai.ID)
	d.Set("id", *ghai.ID)
	d.Set("installation_id", *ghai.InstallationID)
	d.Set("name", *ghai.Name)
	return nil
}
