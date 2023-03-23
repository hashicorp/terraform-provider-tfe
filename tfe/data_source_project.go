// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// Read context to implement cancellation
//

package tfe

import (
	"context"
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
				Required: true,
			},

			"project_id": {
				Type:     schema.TypeString,
				Computed: true,
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

	for {
		l, err := config.Client.Projects.List(ctx, orgName, options)
		if err != nil {
			return diag.Errorf("Error retrieving projects: %v", err)
		}

		for _, proj := range l.Items {
			if proj.Name == projName {
				d.SetId(proj.ID)
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
	return diag.Errorf("could not find project %s/%s", orgName, projName)
}
