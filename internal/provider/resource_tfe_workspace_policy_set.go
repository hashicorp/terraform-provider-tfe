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

func resourceTFEWorkspacePolicySet() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFEWorkspacePolicySetCreate,
		Read:   resourceTFEWorkspacePolicySetRead,
		Delete: resourceTFEWorkspacePolicySetDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceTFEWorkspacePolicySetImporter,
		},

		Schema: map[string]*schema.Schema{
			"policy_set_id": {
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

func resourceTFEWorkspacePolicySetCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	policySetID := d.Get("policy_set_id").(string)
	workspaceID := d.Get("workspace_id").(string)

	policySetAddWorkspacesOptions := tfe.PolicySetAddWorkspacesOptions{}
	policySetAddWorkspacesOptions.Workspaces = append(policySetAddWorkspacesOptions.Workspaces, &tfe.Workspace{ID: workspaceID})

	err := config.Client.PolicySets.AddWorkspaces(ctx, policySetID, policySetAddWorkspacesOptions)
	if err != nil {
		return fmt.Errorf(
			"Error attaching policy set id %s to workspace %s: %w", policySetID, workspaceID, err)
	}

	d.SetId(fmt.Sprintf("%s_%s", workspaceID, policySetID))

	return resourceTFEWorkspacePolicySetRead(d, meta)
}

func resourceTFEWorkspacePolicySetRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	policySetID := d.Get("policy_set_id").(string)
	workspaceID := d.Get("workspace_id").(string)

	log.Printf("[DEBUG] Read configuration of workspace policy set: %s", policySetID)
	policySet, err := config.Client.PolicySets.ReadWithOptions(ctx, policySetID, &tfe.PolicySetReadOptions{
		Include: []tfe.PolicySetIncludeOpt{tfe.PolicySetWorkspaces},
	})
	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
			log.Printf("[DEBUG] Policy set %s no longer exists", policySetID)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading configuration of policy set %s: %w", policySetID, err)
	}

	isWorkspaceAttached := false
	for _, workspace := range policySet.Workspaces {
		if workspace.ID == workspaceID {
			isWorkspaceAttached = true
			d.Set("workspace_id", workspaceID)
			break
		}
	}

	if !isWorkspaceAttached {
		log.Printf("[DEBUG] Workspace %s not attached to policy set %s. Removing from state.", workspaceID, policySetID)
		d.SetId("")
		return nil
	}

	d.Set("policy_set_id", policySetID)
	return nil
}

func resourceTFEWorkspacePolicySetDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	policySetID := d.Get("policy_set_id").(string)
	workspaceID := d.Get("workspace_id").(string)

	log.Printf("[DEBUG] Detaching workspace (%s) from policy set (%s)", workspaceID, policySetID)
	policySetRemoveWorkspacesOptions := tfe.PolicySetRemoveWorkspacesOptions{}
	policySetRemoveWorkspacesOptions.Workspaces = append(policySetRemoveWorkspacesOptions.Workspaces, &tfe.Workspace{ID: workspaceID})

	err := config.Client.PolicySets.RemoveWorkspaces(ctx, policySetID, policySetRemoveWorkspacesOptions)
	if err != nil {
		return fmt.Errorf(
			"Error detaching workspace %s from policy set %s: %w", workspaceID, policySetID, err)
	}

	return nil
}

func resourceTFEWorkspacePolicySetImporter(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// The format of the import ID is <ORGANIZATION/WORKSPACE NAME/POLICYSET NAME>
	splitID := strings.SplitN(d.Id(), "/", 3)
	if len(splitID) != 3 {
		return nil, fmt.Errorf(
			"invalid workspace policy set input format: %s (expected <ORGANIZATION>/<WORKSPACE NAME>/<POLICYSET NAME>)",
			splitID,
		)
	}

	organization, wsName, pSName := splitID[0], splitID[1], splitID[2]

	config := meta.(ConfiguredClient)

	// Ensure the named workspace exists before fetching all the policy sets in the org
	_, err := config.Client.Workspaces.Read(ctx, organization, wsName)
	if err != nil {
		return nil, fmt.Errorf("error reading configuration of workspace %s in organization %s: %w", wsName, organization, err)
	}

	options := &tfe.PolicySetListOptions{Include: []tfe.PolicySetIncludeOpt{tfe.PolicySetWorkspaces}}
	for {
		list, err := config.Client.PolicySets.List(ctx, organization, options)
		if err != nil {
			return nil, fmt.Errorf("Error retrieving policy sets: %w", err)
		}
		for _, policySet := range list.Items {
			if policySet.Name != pSName {
				continue
			}

			for _, ws := range policySet.Workspaces {
				if ws.Name != wsName {
					continue
				}

				d.Set("workspace_id", ws.ID)
				d.Set("policy_set_id", policySet.ID)
				d.SetId(fmt.Sprintf("%s_%s", ws.ID, policySet.ID))

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

	return nil, fmt.Errorf("workspace %s has not been assigned to policy set %s", wsName, pSName)
}
