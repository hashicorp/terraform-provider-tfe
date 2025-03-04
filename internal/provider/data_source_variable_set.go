// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// NOTE: This is a legacy resource and should be migrated to the Plugin
// Framework if substantial modifications are planned. See
// docs/new-resources.md if planning to use this code as boilerplate for
// a new resource.

package provider

import (
	"fmt"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTFEVariableSet() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTFEVariableSetRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"organization": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"global": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"priority": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"workspace_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"variable_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"project_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"parent_project_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func dataSourceTFEVariableSetRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	// Get the name and organization.
	name := d.Get("name").(string)
	organization, err := config.schemaOrDefaultOrganization(d)
	if err != nil {
		return err
	}

	// Create an options struct.
	options := tfe.VariableSetListOptions{}

	for {
		// Variable Set relations, vars and workspaces, are omitted from the querying until
		// we find the desired variable set.
		l, err := config.Client.VariableSets.List(ctx, organization, &options)
		if err != nil {
			if err == tfe.ErrResourceNotFound {
				return fmt.Errorf("could not find variable set%s/%s", organization, name)
			}
			return fmt.Errorf("Error retrieving variable set: %w", err)
		}

		for _, vs := range l.Items {
			if vs.Name == name {
				d.Set("name", vs.Name)
				d.Set("description", vs.Description)
				d.Set("global", vs.Global)
				d.Set("priority", vs.Priority)

				if vs.Parent != nil && vs.Parent.Project != nil {
					d.Set("parent_project_id", vs.Parent.Project.ID)
				}

				// Only now include vars and workspaces to cut down on request load.
				readOptions := tfe.VariableSetReadOptions{
					Include: &[]tfe.VariableSetIncludeOpt{tfe.VariableSetWorkspaces, tfe.VariableSetVars},
				}

				vs, err = config.Client.VariableSets.Read(ctx, vs.ID, &readOptions)
				if err != nil {
					return fmt.Errorf("Error retrieving variable set relations: %w", err)
				}

				var workspaces []interface{}
				for _, workspace := range vs.Workspaces {
					workspaces = append(workspaces, workspace.ID)
				}
				d.Set("workspace_ids", workspaces)

				var variables []interface{}
				for _, variable := range vs.Variables {
					variables = append(variables, variable.ID)
				}
				d.Set("variable_ids", variables)

				var projects []interface{}
				for _, project := range vs.Projects {
					projects = append(projects, project.ID)
				}
				d.Set("project_ids", projects)

				d.SetId(vs.ID)
				return nil
			}
		}

		// Exit the loop when we've seen all pages.
		if l.CurrentPage >= l.TotalPages {
			break
		}

		// Update the page number to get the next page.
		options.PageNumber = l.NextPage
	}

	return fmt.Errorf("could not find variable set %s/%s", organization, name)
}
