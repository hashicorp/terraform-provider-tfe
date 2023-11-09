// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTFEAgentPool() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTFEAgentPoolRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"organization": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"organization_scoped": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"allowed_workspace_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceTFEAgentPoolRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	// Get the name and organization.
	name := d.Get("name").(string)
	organization, err := config.schemaOrDefaultOrganization(d)
	if err != nil {
		return err
	}

	pool, err := fetchAgentPool(organization, name, config.Client)
	if err != nil {
		return err
	}

	d.SetId(pool.ID)
	d.Set("organization_scoped", pool.OrganizationScoped)

	var allowedWorkspaceIDs []string
	for _, allowedWorkspaceID := range pool.AllowedWorkspaces {
		allowedWorkspaceIDs = append(allowedWorkspaceIDs, allowedWorkspaceID.ID)
	}
	d.Set("allowed_workspace_ids", allowedWorkspaceIDs)

	return nil
}
