// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"log"
	"regexp"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var variableSetIDRegexp = regexp.MustCompile("varset-[a-zA-Z0-9]{16}$")

func resourceTFEVariableSet() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFEVariableSetCreate,
		Read:   resourceTFEVariableSetRead,
		Update: resourceTFEVariableSetUpdate,
		Delete: resourceTFEVariableSetDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"global": {
				Type:          schema.TypeBool,
				Optional:      true,
				Default:       false,
				ConflictsWith: []string{"workspace_ids"},
			},

			"organization": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
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

func resourceTFEVariableSetCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	// Get the name and organization.
	name := d.Get("name").(string)
	organization, err := config.schemaOrDefaultOrganization(d)
	if err != nil {
		return err
	}

	// Create a new options struct.
	options := tfe.VariableSetCreateOptions{
		Name:   tfe.String(name),
		Global: tfe.Bool(d.Get("global").(bool)),
	}

	if description, descriptionSet := d.GetOk("description"); descriptionSet {
		options.Description = tfe.String(description.(string))
	}

	variableSet, err := config.Client.VariableSets.Create(ctx, organization, &options)
	if err != nil {
		return fmt.Errorf(
			"Error creating variable set %s, for organization: %s: %w", name, organization, err)
	}

	d.SetId(variableSet.ID)

	if workspaceIDs, workspacesSet := d.GetOk("workspace_ids"); !*options.Global && workspacesSet {
		log.Printf("[DEBUG] Apply variable set %s to workspaces %v", name, workspaceIDs)
		warnWorkspaceIdsDeprecation()

		applyOptions := tfe.VariableSetUpdateWorkspacesOptions{}
		for _, workspaceID := range workspaceIDs.(*schema.Set).List() {
			if val, ok := workspaceID.(string); ok {
				applyOptions.Workspaces = append(applyOptions.Workspaces, &tfe.Workspace{ID: val})
			}
		}

		_, err := config.Client.VariableSets.UpdateWorkspaces(ctx, variableSet.ID, &applyOptions)
		if err != nil {
			return fmt.Errorf(
				"Error applying variable set %s (%s) to given workspaces: %w", name, variableSet.ID, err)
		}
	}

	return resourceTFEVariableSetRead(d, meta)
}

func resourceTFEVariableSetRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Read configuration of variable set: %s", d.Id())
	variableSet, err := config.Client.VariableSets.Read(ctx, d.Id(), &tfe.VariableSetReadOptions{
		Include: &[]tfe.VariableSetIncludeOpt{tfe.VariableSetWorkspaces},
	})
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			log.Printf("[DEBUG] Variable set %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading configuration of variable set %s: %w", d.Id(), err)
	}

	// Update the config.
	d.Set("name", variableSet.Name)
	d.Set("description", variableSet.Description)
	d.Set("global", variableSet.Global)
	d.Set("organization", variableSet.Organization.Name)

	var wids []interface{}
	for _, workspace := range variableSet.Workspaces {
		wids = append(wids, workspace.ID)
	}
	d.Set("workspace_ids", wids)

	return nil
}

func resourceTFEVariableSetUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	if d.HasChange("name") || d.HasChange("description") || d.HasChange("global") {
		options := tfe.VariableSetUpdateOptions{
			Name:        tfe.String(d.Get("name").(string)),
			Description: tfe.String(d.Get("description").(string)),
			Global:      tfe.Bool(d.Get("global").(bool)),
		}

		log.Printf("[DEBUG] Update variable set: %s", d.Id())
		_, err := config.Client.VariableSets.Update(ctx, d.Id(), &options)
		if err != nil {
			return fmt.Errorf("Error updateing variable %s: %w", d.Id(), err)
		}
	}

	if d.HasChanges("workspace_ids") {
		workspaceIDs := d.Get("workspace_ids")
		applyOptions := tfe.VariableSetUpdateWorkspacesOptions{}
		applyOptions.Workspaces = []*tfe.Workspace{}
		for _, workspaceID := range workspaceIDs.(*schema.Set).List() {
			if val, ok := workspaceID.(string); ok {
				applyOptions.Workspaces = append(applyOptions.Workspaces, &tfe.Workspace{ID: val})
			}
		}

		log.Printf("[DEBUG] Apply variable set %s to workspaces %v", d.Id(), workspaceIDs)
		warnWorkspaceIdsDeprecation()
		_, err := config.Client.VariableSets.UpdateWorkspaces(ctx, d.Id(), &applyOptions)
		if err != nil {
			return fmt.Errorf(
				"Error applying variable set %s to given workspaces: %w", d.Id(), err)
		}
	}

	return resourceTFEVariableSetRead(d, meta)
}

func resourceTFEVariableSetDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Delete variable set: %s", d.Id())
	err := config.Client.VariableSets.Delete(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return nil
		}
		return fmt.Errorf("Error deleting variable set %s: %w", d.Id(), err)
	}

	return nil
}

func warnWorkspaceIdsDeprecation() {
	log.Printf("[WARN] The workspace_ids field of tfe_variable_set is deprecated as of release 0.33.0 and may be removed in a future version. The preferred method of associating a variable set to a workspace is by using the tfe_workspace_variable_set resource.")
}
