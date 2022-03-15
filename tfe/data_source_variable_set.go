package tfe

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
				Required: true,
			},

			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"global": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"workspaces": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceTFEVariableSetRead(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	// Get the name and organization.
	name := d.Get("name").(string)
	organization := d.Get("organization").(string)

	// Create an options struct.
	options := tfe.VariableSetListOptions{}

	for {
		// Variable Set relations, vars and workspaces, are omitted from the querying until
		// we find the desired variable set.
		l, err := tfeClient.VariableSets.List(ctx, organization, &options)
		if err != nil {
			if err == tfe.ErrResourceNotFound {
				return fmt.Errorf("could not find variable set%s/%s", organization, name)
			}
			return fmt.Errorf("Error retrieving variable set: %v", err)
		}

		for _, vs := range l.Items {
			if vs.Name == name {
				d.Set("name", vs.Name)
				d.Set("description", vs.Description)
				d.Set("global", vs.Global)

				//Only now include vars and workspaces to cut down on request load.
				readOptions := tfe.VariableSetReadOptions{
					Include: &[]tfe.VariableSetIncludeOps{tfe.VariableSetWorkspaces, tfe.VariableSetVars},
				}

				vs, err = tfeClient.VariableSets.Read(ctx, vs.ID, &readOptions)
				if err != nil {
					return fmt.Errorf("Error retrieving variable set relations: %v", err)
				}

				var workspaces []interface{}
				for _, workspace := range vs.Workspaces {
					workspaces = append(workspaces, workspace.ID)
				}
				d.Set("workspaces", workspaces)

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

	return fmt.Errorf("Could not find variable set %s/%s", organization, name)
}
