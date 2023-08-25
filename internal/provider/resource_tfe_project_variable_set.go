// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTFEProjectVariableSet() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFEProjectVariableSetCreate,
		Read:   resourceTFEProjectVariableSetRead,
		Delete: resourceTFEProjectVariableSetDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceTFEProjectVariableSetImporter,
		},

		Schema: map[string]*schema.Schema{
			"variable_set_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"project_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceTFEProjectVariableSetCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	vSID := d.Get("variable_set_id").(string)
	prjID := d.Get("project_id").(string)

	applyOptions := tfe.VariableSetApplyToProjectsOptions{}
	applyOptions.Projects = append(applyOptions.Projects, &tfe.Project{ID: prjID})

	err := config.Client.VariableSets.ApplyToProjects(ctx, vSID, applyOptions)
	if err != nil {
		return fmt.Errorf(
			"Error applying variable set id %s to project %s: %w", vSID, prjID, err)
	}

	id := encodeVariableSetProjectAttachment(prjID, vSID)
	d.SetId(id)

	return resourceTFEProjectVariableSetRead(d, meta)
}

func resourceTFEProjectVariableSetRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	prjID := d.Get("project_id").(string)
	vSID := d.Get("variable_set_id").(string)

	log.Printf("[DEBUG] Read configuration of project variable set: %s", d.Id())
	vS, err := config.Client.VariableSets.Read(ctx, vSID, &tfe.VariableSetReadOptions{
		Include: &[]tfe.VariableSetIncludeOpt{tfe.VariableSetProjects},
	})
	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
			log.Printf("[DEBUG] Variable set %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading configuration of variable set %s: %w", d.Id(), err)
	}

	// Verify project listed in variable set
	check := false
	for _, project := range vS.Projects {
		if project.ID == prjID {
			check = true
			d.Set("project_id", prjID)
			break
		}
	}
	if !check {
		log.Printf("[DEBUG] Project %s not attached to variable set %s. Removing from state.", prjID, vSID)
		d.SetId("")
		return nil
	}

	d.Set("variable_set_id", vSID)
	return nil
}

func resourceTFEProjectVariableSetDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	prjID := d.Get("project_id").(string)
	vSID := d.Get("variable_set_id").(string)

	log.Printf("[DEBUG] Delete project (%s) from variable set (%s)", prjID, vSID)
	removeOptions := tfe.VariableSetRemoveFromProjectsOptions{}
	removeOptions.Projects = append(removeOptions.Projects, &tfe.Project{ID: prjID})

	err := config.Client.VariableSets.RemoveFromProjects(ctx, vSID, removeOptions)
	if err != nil {
		return fmt.Errorf(
			"Error removing project %s from variable set %s: %w", prjID, vSID, err)
	}

	return nil
}

func encodeVariableSetProjectAttachment(prjID, vSID string) string {
	return fmt.Sprintf("%s_%s", prjID, vSID)
}

func resourceTFEProjectVariableSetImporter(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// The format of the import ID is <ORGANIZATION/PROJECT ID/VARSET NAME> but be aware
	// that variable set names can contain forward slash characters but organization/project
	// names cannot. Therefore, we split the import ID into at most 3 substrings.
	organization, prjID, vSName, err := destructureProjectImportID(strings.SplitN(d.Id(), "/", 3))
	if err != nil {
		return nil, err
	}

	config := meta.(ConfiguredClient)

	// Ensure a project with this ID exists before fetching all the variable sets in the org
	_, err = config.Client.Projects.Read(ctx, prjID)
	if err != nil {
		return nil, fmt.Errorf("error reading project %s in organization %s: %w", prjID, organization, err)
	}

	options := &tfe.VariableSetListOptions{}
	for {
		list, err := config.Client.VariableSets.ListForProject(ctx, prjID, options)
		if err != nil {
			return nil, fmt.Errorf("Error retrieving variable sets for project %s: %w", prjID, err)
		}
		for _, vs := range list.Items {
			if vs.Name != vSName {
				continue
			}

			d.Set("project_id", prjID)
			d.Set("variable_set_id", vs.ID)
			d.SetId(encodeVariableSetProjectAttachment(prjID, vs.ID))

			return []*schema.ResourceData{d}, nil
		}

		// Exit the loop when we've seen all pages.
		if list.CurrentPage >= list.TotalPages {
			break
		}

		// Update the page number to get the next page.
		options.PageNumber = list.NextPage
	}

	return nil, fmt.Errorf("project %s has not been assigned to variable set %s", prjID, vSName)
}

func destructureProjectImportID(splitID []string) (string, string, string, error) {
	if len(splitID) != 3 {
		return "", "", "", fmt.Errorf(
			"invalid project variable set input format: %s (expected <ORGANIZATION>/<PROJECT ID>/<VARIABLE SET NAME>)",
			splitID,
		)
	}

	return splitID[0], splitID[1], splitID[2], nil
}
