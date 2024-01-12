// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// NOTE: This is a legacy resource and should be migrated to the Plugin
// Framework if substantial modifications are planned. See
// docs/new-resources.md if planning to use this code as boilerplate for
// a new resource.

package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTFEOrganizationRunTask() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTFEOrganizationRunTaskRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"organization": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"url": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"category": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func dataSourceTFEOrganizationRunTaskRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)
	name := d.Get("name").(string)
	organization, err := config.schemaOrDefaultOrganization(d)
	if err != nil {
		return err
	}

	task, err := fetchOrganizationRunTask(name, organization, config.Client)
	if err != nil {
		return err
	}

	d.Set("url", task.URL)
	d.Set("category", task.Category)
	d.Set("enabled", task.Enabled)
	d.Set("description", task.Description)
	d.SetId(task.ID)

	return nil
}
