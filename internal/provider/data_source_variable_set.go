// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

// NOTE: This is a legacy resource and should be migrated to the Plugin
// Framework if substantial modifications are planned. See
// docs/new-resources.md if planning to use this code as boilerplate for
// a new resource.

package provider

import (
	"errors"
	"fmt"

	tfev2 "github.com/hashicorp/go-tfe/v2"
	v2orgs "github.com/hashicorp/go-tfe/v2/api/organizations"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTFEVariableSet() *schema.Resource {
	return &schema.Resource{
		Description: "Gets information on a named variable set.",

		Read: dataSourceTFEVariableSetRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The ID of the variable set.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"name": {
				Description: "Name of the variable set.",
				Type:        schema.TypeString,
				Required:    true,
			},

			"organization": {
				Description: "Name of the organization.",
				Type:        schema.TypeString,
				Optional:    true,
			},

			"description": {
				Description: "Description of the variable set.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"global": {
				Description: "Whether the variable set applies to all workspaces in the organization.",
				Type:        schema.TypeBool,
				Computed:    true,
			},

			"priority": {
				Description: "When true, the variables in this set take priority over workspace-level variables and cannot be overridden.",
				Type:        schema.TypeBool,
				Computed:    true,
			},

			"workspace_ids": {
				Description: "IDs of the workspaces that use the variable set.",
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},

			"variable_ids": {
				Description: "IDs of the variables attached to the variable set.",
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},

			"project_ids": {
				Description: "IDs of the projects that use the variable set.",
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},

			"stack_ids": {
				Description: "IDs of the stacks that use the variable set.",
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},

			"parent_project_id": {
				Description: "ID of the project that owns the variable set.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func dataSourceTFEVariableSetRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)
	api := config.ClientV2.API

	// Get the name and organization.
	name := d.Get("name").(string)
	organization, err := config.schemaOrDefaultOrganization(d)
	if err != nil {
		return err
	}

	// List varsets in the org, searching by name, and find the matching one.
	pageSize := int32(100)
	queryParams := &v2orgs.ItemVarsetsRequestBuilderGetQueryParameters{
		Q:        &name,
		Pagesize: &pageSize,
	}

	for {
		l, err := api.Organizations().ByOrganization_name(organization).Varsets().Get(ctx, withQueryParams(queryParams))
		if err != nil {
			if errors.Is(err, tfev2.ErrNotFound) {
				return fmt.Errorf("could not find variable set%s/%s", organization, name)
			}
			return fmt.Errorf("Error retrieving variable set: %w", err)
		}

		for _, vs := range l.GetData() {
			vsAttrs := vs.GetAttributes()
			if vsAttrs == nil || valueOrZero(vsAttrs.GetName()) != name {
				continue
			}

			vsID := valueOrZero(vs.GetId())

			d.Set("name", valueOrZero(vsAttrs.GetName()))
			d.Set("description", valueOrZero(vsAttrs.GetDescription()))
			d.Set("global", valueOrZero(vsAttrs.GetGlobal()))
			d.Set("priority", valueOrZero(vsAttrs.GetPriority()))

			// Fetch the full varset to get relationship data (workspaces, vars, projects, stacks, parent).
			vsResp, err := api.Varsets().ByVarset_id(vsID).Get(ctx, nil)
			if err != nil {
				return fmt.Errorf("Error retrieving variable set relations: %w", err)
			}

			if vsResp.GetData() == nil || vsResp.GetData().GetRelationships() == nil {
				d.SetId(vsID)
				return nil
			}

			rels := vsResp.GetData().GetRelationships()

			var workspaces []interface{}
			if wsRel := rels.GetWorkspaces(); wsRel != nil {
				for _, ws := range wsRel.GetData() {
					workspaces = append(workspaces, valueOrZero(ws.GetId()))
				}
			}
			d.Set("workspace_ids", workspaces)

			var variables []interface{}
			if varRel := rels.GetVars(); varRel != nil {
				for _, v := range varRel.GetData() {
					variables = append(variables, valueOrZero(v.GetId()))
				}
			}
			d.Set("variable_ids", variables)

			var projects []interface{}
			if prjRel := rels.GetProjects(); prjRel != nil {
				for _, p := range prjRel.GetData() {
					projects = append(projects, valueOrZero(p.GetId()))
				}
			}
			d.Set("project_ids", projects)

			var stacks []interface{}
			if stkRel := rels.GetStacks(); stkRel != nil {
				for _, s := range stkRel.GetData() {
					stacks = append(stacks, valueOrZero(s.GetId()))
				}
			}
			d.Set("stack_ids", stacks)

			// Parent project ID: present when parent relationship type is "projects".
			if parent := rels.GetParent(); parent != nil && parent.GetData() != nil {
				parentData := parent.GetData()
				parentType := parentData.GetTypeEscaped()
				if parentType != nil && parentType.String() == "projects" {
					d.Set("parent_project_id", valueOrZero(parentData.GetId()))
				}
			}

			d.SetId(vsID)
			return nil
		}

		// Exit the loop when we've seen all pages.
		nextPage := nextPageNumber(l.GetMeta())
		if nextPage == nil {
			break
		}
		queryParams.Pagenumber = nextPage
	}

	return fmt.Errorf("could not find variable set %s/%s", organization, name)
}
