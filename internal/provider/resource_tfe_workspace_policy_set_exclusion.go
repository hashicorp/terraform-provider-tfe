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

func resourceTFEWorkspacePolicySetExclusion() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFEWorkspacePolicySetExclusionCreate,
		Read:   resourceTFEWorkspacePolicySetExclusionRead,
		Delete: resourceTFEWorkspacePolicySetExclusionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceTFEWorkspacePolicySetExclusionImporter,
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

func resourceTFEWorkspacePolicySetExclusionCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	policySetID := d.Get("policy_set_id").(string)
	workspaceExclusionID := d.Get("workspace_id").(string)

	policySetAddWorkspaceExclusionsOptions := tfe.PolicySetAddWorkspaceExclusionsOptions{}
	policySetAddWorkspaceExclusionsOptions.WorkspaceExclusions = append(policySetAddWorkspaceExclusionsOptions.WorkspaceExclusions, &tfe.Workspace{ID: workspaceExclusionID})

	err := config.Client.PolicySets.AddWorkspaceExclusions(ctx, policySetID, policySetAddWorkspaceExclusionsOptions)
	if err != nil {
		return fmt.Errorf(
			"error adding workspace exclusion %s to policy set id %s: %w", workspaceExclusionID, policySetID, err)
	}

	d.SetId(fmt.Sprintf("%s_%s", workspaceExclusionID, policySetID))

	return resourceTFEWorkspacePolicySetExclusionRead(d, meta)
}

func resourceTFEWorkspacePolicySetExclusionRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	policySetID := d.Get("policy_set_id").(string)
	workspaceExclusionsID := d.Get("workspace_id").(string)

	log.Printf("[DEBUG] Read configuration of excluded workspace policy set: %s", policySetID)
	policySet, err := config.Client.PolicySets.ReadWithOptions(ctx, policySetID, &tfe.PolicySetReadOptions{
		Include: []tfe.PolicySetIncludeOpt{tfe.PolicySetWorkspaceExclusions},
	})
	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
			log.Printf("[DEBUG] Policy set %s no longer exists", policySetID)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading configuration of policy set %s: %w", policySetID, err)
	}

	isWorkspaceExclusionsAttached := false
	for _, excludedWorkspace := range policySet.WorkspaceExclusions {
		if excludedWorkspace.ID == workspaceExclusionsID {
			isWorkspaceExclusionsAttached = true
			d.Set("workspace_id", workspaceExclusionsID)
			break
		}
	}

	if !isWorkspaceExclusionsAttached {
		log.Printf("[DEBUG] Excluded workspace %s not attached to policy set %s. Removing from state.", workspaceExclusionsID, policySetID)
		d.SetId("")
		return nil
	}

	d.Set("policy_set_id", policySetID)
	return nil
}

func resourceTFEWorkspacePolicySetExclusionDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	policySetID := d.Get("policy_set_id").(string)
	workspaceExclusionsID := d.Get("workspace_id").(string)

	log.Printf("[DEBUG] Removing excluded workspace (%s) from policy set (%s)", workspaceExclusionsID, policySetID)
	policySetRemoveWorkspaceExclusionsOptions := tfe.PolicySetRemoveWorkspaceExclusionsOptions{}
	policySetRemoveWorkspaceExclusionsOptions.WorkspaceExclusions = append(policySetRemoveWorkspaceExclusionsOptions.WorkspaceExclusions, &tfe.Workspace{ID: workspaceExclusionsID})

	err := config.Client.PolicySets.RemoveWorkspaceExclusions(ctx, policySetID, policySetRemoveWorkspaceExclusionsOptions)
	if err != nil {
		return fmt.Errorf(
			"error removing excluded workspace %s from policy set %s: %w", workspaceExclusionsID, policySetID, err)
	}

	return nil
}

func resourceTFEWorkspacePolicySetExclusionImporter(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// The format of the import ID is <ORGANIZATION/WORKSPACE NAME/POLICYSET NAME>
	splitID := strings.SplitN(d.Id(), "/", 3)
	if len(splitID) != 3 {
		return nil, fmt.Errorf(
			"invalid excluded workspace policy set input format: %s (expected <ORGANIZATION>/<WORKSPACE NAME>/<POLICYSET NAME>)",
			splitID,
		)
	}

	organization, wsName, pSName := splitID[0], splitID[1], splitID[2]

	config := meta.(ConfiguredClient)

	// Ensure the named workspace exists before fetching all the policy sets in the org
	_, err := config.Client.Workspaces.Read(ctx, organization, wsName)
	if err != nil {
		return nil, fmt.Errorf("error reading configuration of the workspace to exclude %s in organization %s: %w", wsName, organization, err)
	}

	options := &tfe.PolicySetListOptions{Include: []tfe.PolicySetIncludeOpt{tfe.PolicySetWorkspaceExclusions}}
	for {
		list, err := config.Client.PolicySets.List(ctx, organization, options)
		if err != nil {
			return nil, fmt.Errorf("error retrieving policy sets: %w", err)
		}
		for _, policySet := range list.Items {
			if policySet.Name != pSName {
				continue
			}

			for _, ws := range policySet.WorkspaceExclusions {
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

	return nil, fmt.Errorf("excluded workspace %s has not been added to policy set %s", wsName, pSName)
}
