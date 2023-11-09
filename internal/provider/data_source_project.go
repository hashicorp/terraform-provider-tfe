// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// Read context to implement cancellation
//

package provider

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTFEProject() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceTFEProjectRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"organization": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"workspace_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceTFEProjectRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ConfiguredClient)

	// Get the project name and organization
	projName := d.Get("name").(string)
	orgName, err := config.schemaOrDefaultOrganizationKey(d, "organization")
	if err != nil {
		return diag.Errorf("Error retrieving organization name: %v", err)
	}

	// Create an options struct.
	options := &tfe.ProjectListOptions{
		Name: projName,
	}

	l, err := config.Client.Projects.List(ctx, orgName, options)
	if err != nil {
		return diag.Errorf("Error retrieving projects: %v", err)
	}

	for _, proj := range l.Items {
		// Case-insensitive uniqueness is enforced in TFC
		if strings.EqualFold(proj.Name, projName) {
			// Only now include workspaces to cut down on request load.
			readOptions := &tfe.WorkspaceListOptions{
				ProjectID: proj.ID,
			}
			var workspaces []interface{}

			for {
				wl, err := config.Client.Workspaces.List(ctx, orgName, readOptions)
				if err != nil {
					return diag.Errorf("Error retrieving workspaces: %v", err)
				}

				for _, workspace := range wl.Items {
					workspaces = append(workspaces, workspace.ID)
				}

				// Exit the loop when we've seen all pages.
				if wl.CurrentPage >= wl.TotalPages {
					break
				}

				// Update the page number to get the next page.
				readOptions.PageNumber = wl.NextPage
			}

			d.Set("workspace_ids", workspaces)
			d.SetId(proj.ID)
			return nil
		}
	}
	return diag.Errorf("could not find project %s/%s", orgName, projName)
}
