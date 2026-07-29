// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

// NOTE: This is a legacy resource and should be migrated to the Plugin
// Framework if substantial modifications are planned. See
// docs/new-resources.md if planning to use this code as boilerplate for
// a new resource.

package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTFEAgentPool() *schema.Resource {
	return &schema.Resource{
		Description: "Gets information about an agent pool.",

		Read: dataSourceTFEAgentPoolRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The agent pool ID.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"name": {
				Description: "Name of the agent pool.",
				Type:        schema.TypeString,
				Required:    true,
			},

			"organization": {
				Description: "Name of the organization.",
				Type:        schema.TypeString,
				Optional:    true,
			},

			"organization_scoped": {
				Description: "Whether or not the agent pool can be used by all workspaces in the organization.",
				Type:        schema.TypeBool,
				Computed:    true,
			},

			"allowed_workspace_ids": {
				Description: "The set of workspace IDs that have permission to use the agent pool.",
				Type:        schema.TypeSet,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},

			"allowed_project_ids": {
				Description: "The set of project IDs that have permission to use the agent pool.",
				Type:        schema.TypeSet,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},

			"excluded_workspace_ids": {
				Description: "The set of workspace IDs that are excluded from the scope of the agent pool.",
				Type:        schema.TypeSet,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
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

	pool, err := fetchAgentPool(organization, name, config.ClientV2)
	if err != nil {
		return err
	}

	d.SetId(valueOrZero(pool.GetId()))

	if attrs := pool.GetAttributes(); attrs != nil {
		d.Set("organization_scoped", valueOrZero(attrs.GetOrganizationScoped()))
	}

	if rels := pool.GetRelationships(); rels != nil {
		if ap := rels.GetAllowedProjects(); ap != nil {
			var allowedProjectIDs []string
			for _, proj := range ap.GetData() {
				if proj != nil {
					allowedProjectIDs = append(allowedProjectIDs, valueOrZero(proj.GetId()))
				}
			}
			d.Set("allowed_project_ids", allowedProjectIDs)
		}

		if aw := rels.GetAllowedWorkspaces(); aw != nil {
			var allowedWorkspaceIDs []string
			for _, ws := range aw.GetData() {
				if ws != nil {
					allowedWorkspaceIDs = append(allowedWorkspaceIDs, valueOrZero(ws.GetId()))
				}
			}
			d.Set("allowed_workspace_ids", allowedWorkspaceIDs)
		}

		if ew := rels.GetExcludedWorkspaces(); ew != nil {
			var excludedWorkspaceIDs []string
			for _, ws := range ew.GetData() {
				if ws != nil {
					excludedWorkspaceIDs = append(excludedWorkspaceIDs, valueOrZero(ws.GetId()))
				}
			}
			d.Set("excluded_workspace_ids", excludedWorkspaceIDs)
		}
	}

	return nil
}
