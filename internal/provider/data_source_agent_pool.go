// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

// NOTE: This is a legacy resource and should be migrated to the Plugin
// Framework if substantial modifications are planned. See
// docs/new-resources.md if planning to use this code as boilerplate for
// a new resource.

package provider

import (
	"github.com/hashicorp/go-tfe/v2/api/models"
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
		d.Set("allowed_project_ids", projectIDsFromRelationship(rels.GetAllowedProjects()))
		d.Set("allowed_workspace_ids", workspaceIDsFromRelationship(rels.GetAllowedWorkspaces()))
		d.Set("excluded_workspace_ids", workspaceIDsFromRelationship(rels.GetExcludedWorkspaces()))
	}

	return nil
}

// projectIDsFromRelationship extracts the list of project IDs from a
// projects has-many relationship, if present.
func projectIDsFromRelationship(rel models.ProjectsHasManyable) []string {
	if rel == nil {
		return nil
	}
	var ids []string
	for _, proj := range rel.GetData() {
		if proj != nil {
			ids = append(ids, valueOrZero(proj.GetId()))
		}
	}
	return ids
}

// workspaceIDsFromRelationship extracts the list of workspace IDs from a
// workspaces has-many relationship, if present.
func workspaceIDsFromRelationship(rel models.WorkspacesHasManyable) []string {
	if rel == nil {
		return nil
	}
	var ids []string
	for _, ws := range rel.GetData() {
		if ws != nil {
			ids = append(ids, valueOrZero(ws.GetId()))
		}
	}
	return ids
}
