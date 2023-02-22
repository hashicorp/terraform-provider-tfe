// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package tfe

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
	"time"
)

func dataSourceTFEGHAInstallation() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGHAInstallationRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:         schema.TypeString,
				Optional:     true,
				AtLeastOneOf: []string{"id", "name", "gh_installation_id"},
			},
			"gh_installation_id": {
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

	var oc *tfe.GHAInstallation
	var err error

	switch v, ok := d.GetOk("id"); {
	case ok:
		oc, err = config.Client.GHAInstallations.Read(ctx, v.(string))
		if err != nil {
			return fmt.Errorf("Error retrieving Github App Installation: %w", err)
		}
	default:
		// search by name or id
		if err != nil {
			return err
		}

		var name string
		var installationId int32
		vName, ok := d.GetOk("name")
		if ok {
			name = vName.(string)
		}
		vInstallationId, ok := d.GetOk("gh_installation_id")
		if ok {
			installationId = vInstallationId.(int32)
		}

		oc, err = fetchGithubAppInstallationByNameOrGHID(ctx, config.Client, name, installationId)
		if err != nil {
			return err
		}
	}

	d.SetId(oc.ID)
	d.Set("installation_d", oc.InstallationId)
	if oc.Name != nil {
		d.Set("name", *oc.Name)
	}

	return nil
}
}
