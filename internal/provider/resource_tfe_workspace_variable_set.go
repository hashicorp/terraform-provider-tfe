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

func resourceTFEWorkspaceVariableSet() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFEWorkspaceVariableSetCreate,
		Read:   resourceTFEWorkspaceVariableSetRead,
		Delete: resourceTFEWorkspaceVariableSetDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceTFEWorkspaceVariableSetImporter,
		},

		Schema: map[string]*schema.Schema{
			"variable_set_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"workspace_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceTFEWorkspaceVariableSetCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	vSID := d.Get("variable_set_id").(string)
	wID := d.Get("workspace_id").(string)

	applyOptions := tfe.VariableSetApplyToWorkspacesOptions{}
	applyOptions.Workspaces = append(applyOptions.Workspaces, &tfe.Workspace{ID: wID})

	err := config.Client.VariableSets.ApplyToWorkspaces(ctx, vSID, &applyOptions)
	if err != nil {
		return fmt.Errorf(
			"Error applying variable set id %s to workspace %s: %w", vSID, wID, err)
	}

	id := encodeVariableSetWorkspaceAttachment(wID, vSID)
	d.SetId(id)

	return resourceTFEWorkspaceVariableSetRead(d, meta)
}

func resourceTFEWorkspaceVariableSetRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	wID := d.Get("workspace_id").(string)
	vSID := d.Get("variable_set_id").(string)

	log.Printf("[DEBUG] Read configuration of workspace variable set: %s", d.Id())
	vS, err := config.Client.VariableSets.Read(ctx, vSID, &tfe.VariableSetReadOptions{
		Include: &[]tfe.VariableSetIncludeOpt{tfe.VariableSetWorkspaces},
	})
	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
			log.Printf("[DEBUG] Variable set %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading configuration of variable set %s: %w", d.Id(), err)
	}

	// Verify workspace listed in variable set
	check := false
	for _, workspace := range vS.Workspaces {
		if workspace.ID == wID {
			check = true
			d.Set("workspace_id", wID)
		}
	}
	if !check {
		log.Printf("[DEBUG] Workspace %s not attached to variable set %s. Removing from state.", wID, vSID)
		d.SetId("")
		return nil
	}

	d.Set("variable_set_id", vSID)
	return nil
}

func resourceTFEWorkspaceVariableSetDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	wID := d.Get("workspace_id").(string)
	vSID := d.Get("variable_set_id").(string)

	log.Printf("[DEBUG] Delete workspace (%s) from variable set (%s)", wID, vSID)
	removeOptions := tfe.VariableSetRemoveFromWorkspacesOptions{}
	removeOptions.Workspaces = append(removeOptions.Workspaces, &tfe.Workspace{ID: wID})

	err := config.Client.VariableSets.RemoveFromWorkspaces(ctx, vSID, &removeOptions)
	if err != nil {
		return fmt.Errorf(
			"Error removing workspace %s from variable set %s: %w", wID, vSID, err)
	}

	return nil
}

func encodeVariableSetWorkspaceAttachment(wID, vSID string) string {
	return fmt.Sprintf("%s_%s", wID, vSID)
}

func resourceTFEWorkspaceVariableSetImporter(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// The format of the import ID is <ORGANIZATION/WORKSPACE NAME/VARSET NAME> but be aware
	// that variable set names can contain forward slash characters but organization/workspace
	// names cannot. Therefore, we split the import ID into at most 3 substrings.
	organization, wsName, vSName, err := destructureImportID(strings.SplitN(d.Id(), "/", 3))
	if err != nil {
		return nil, err
	}

	config := meta.(ConfiguredClient)

	// Ensure a workspace of this name exists before fetching all the variable sets in the org
	_, err = config.Client.Workspaces.Read(ctx, organization, wsName)
	if err != nil {
		return nil, fmt.Errorf("error reading configuration of workspace %s in organization %s: %w", wsName, organization, err)
	}

	options := &tfe.VariableSetListOptions{
		Include: string(tfe.VariableSetWorkspaces),
	}
	for {
		list, err := config.Client.VariableSets.List(ctx, organization, options)
		if err != nil {
			return nil, fmt.Errorf("Error retrieving variable sets: %w", err)
		}
		for _, vs := range list.Items {
			if vs.Name != vSName {
				continue
			}

			for _, ws := range vs.Workspaces {
				if ws.Name != wsName {
					continue
				}

				d.Set("workspace_id", ws.ID)
				d.Set("variable_set_id", vs.ID)
				d.SetId(encodeVariableSetWorkspaceAttachment(ws.ID, vs.ID))

				return []*schema.ResourceData{d}, nil
			}
		}

		// Exit the loop when we've seen all pages.
		if list.CurrentPage >= list.TotalPages {
			break
		}

		// Update the page number to get the next page.
		options.PageNumber = list.NextPage
	}

	return nil, fmt.Errorf("workspace %s has not been assigned to variable set %s", wsName, vSName)
}

func destructureImportID(splitID []string) (string, string, string, error) {
	if len(splitID) != 3 {
		return "", "", "", fmt.Errorf(
			"invalid workspace variable set input format: %s (expected <ORGANIZATION><WORKSPACE NAME>/<VARIABLE SET NAME>)",
			splitID,
		)
	}

	return splitID[0], splitID[1], splitID[2], nil
}
