// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

// NOTE: This is a legacy resource and should be migrated to the Plugin
// Framework if substantial modifications are planned. See
// docs/new-resources.md if planning to use this code as boilerplate for
// a new resource.

package provider

import (
	"context"
	"log"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTFEGHAInstallation() *schema.Resource {
	return &schema.Resource{
		Description: "Gets information about the Github App installation.",
		Read:        dataSourceGHAInstallationRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The internal ID of the Github Installation. This is different from the `installation_id`",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"installation_id": {
				Description:  "ID of the Github Installation. The installation ID can be found in the URL slug when visiting the installation's configuration page, e.g `https://github.com/settings/installations/12345678`.",
				Type:         schema.TypeInt,
				Optional:     true,
				AtLeastOneOf: []string{"name", "installation_id"},
			},
			"name": {
				Description: "Name of the Github user or organization account that installed the app.",
				Type:        schema.TypeString,
				Optional:    true,
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
