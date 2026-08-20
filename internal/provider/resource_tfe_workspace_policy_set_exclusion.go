// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

// NOTE: This is a legacy resource and should be migrated to the Plugin
// Framework if substantial modifications are planned. See
// docs/new-resources.md if planning to use this code as boilerplate for
// a new resource.

package provider

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	tfeV2 "github.com/hashicorp/go-tfe/v2"
	"github.com/hashicorp/go-tfe/v2/api/organizations"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTFEWorkspacePolicySetExclusion() *schema.Resource {
	return &schema.Resource{
		Description: "Adds and removes policy sets from an excluded workspace." +
			"\n\n~> **Note:** `tfe_policy_set` has an argument `workspace_ids` that should not be used alongside this resource because they manage the same attachments." +
			"\n\n~> **Note:** Tag-based scoping and explicit workspace/project associations are mutually exclusive on a policy set. To switch between them, first remove the existing association (`terraform apply`), then add the new one (`terraform apply`). Attempting both in a single apply may fail.",

		Create: resourceTFEWorkspacePolicySetExclusionCreate,
		Read:   resourceTFEWorkspacePolicySetExclusionRead,
		Delete: resourceTFEWorkspacePolicySetExclusionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceTFEWorkspacePolicySetExclusionImporter,
		},

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The ID of the policy set exclusion. ID format: `<workspace-id>_<policy-set-id>`.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"policy_set_id": {
				Description: "ID of the policy set.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},

			"workspace_id": {
				Description: "Excluded workspace ID to add the policy set to.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
		},
	}
}

func resourceTFEWorkspacePolicySetExclusionCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	policySetID := d.Get("policy_set_id").(string)
	workspaceExclusionID := d.Get("workspace_id").(string)

	body := makeWorkspaceIdentifierArrayDocument([]interface{}{workspaceExclusionID})
	err := config.ClientV2.API.PolicySets().ByPolicy_set_id(policySetID).Relationships().WorkspaceExclusions().Post(ctx, body, nil)
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
	policySetEnv, err := config.ClientV2.API.PolicySets().ByPolicy_set_id(policySetID).Get(ctx, nil)
	if err != nil {
		if errors.Is(err, tfeV2.ErrNotFound) {
			log.Printf("[DEBUG] Policy set %s no longer exists", policySetID)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading configuration of policy set %s: %w", policySetID, err)
	}

	policySet := policySetEnv.GetData()
	rels := policySet.GetRelationships()

	isWorkspaceExclusionsAttached := false
	if rels != nil && rels.GetWorkspaceExclusions() != nil {
		for _, ws := range rels.GetWorkspaceExclusions().GetData() {
			if valueOrZero(ws.GetId()) == workspaceExclusionsID {
				isWorkspaceExclusionsAttached = true
				d.Set("workspace_id", workspaceExclusionsID)
				break
			}
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
	body := makeWorkspaceIdentifierArrayDocument([]interface{}{workspaceExclusionsID})
	err := config.ClientV2.API.PolicySets().ByPolicy_set_id(policySetID).Relationships().WorkspaceExclusions().Delete(ctx, body, nil)
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
	workspaceEnv, err := config.ClientV2.API.Organizations().ByOrganization_name(organization).Workspaces().ByWorkspace_name(wsName).Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error reading configuration of the workspace to exclude %s in organization %s: %w", wsName, organization, err)
	}
	workspaceID := valueOrZero(workspaceEnv.GetData().GetId())

	pageSize := int32(100)
	queryParams := &organizations.ItemPolicySetsRequestBuilderGetQueryParameters{
		Pagesize:   &pageSize,
		Searchname: &pSName,
	}
	for {
		list, err := config.ClientV2.API.Organizations().ByOrganization_name(organization).PolicySets().Get(ctx, withQueryParams(queryParams))
		if err != nil {
			return nil, fmt.Errorf("error retrieving policy sets: %w", err)
		}
		for _, policySet := range list.GetData() {
			psAttrs := policySet.GetAttributes()
			if psAttrs == nil || valueOrZero(psAttrs.GetName()) != pSName {
				continue
			}

			rels := policySet.GetRelationships()
			if rels == nil || rels.GetWorkspaceExclusions() == nil {
				continue
			}
			for _, ws := range rels.GetWorkspaceExclusions().GetData() {
				wsID := valueOrZero(ws.GetId())
				if wsID != workspaceID {
					continue
				}

				d.Set("workspace_id", wsID)
				d.Set("policy_set_id", valueOrZero(policySet.GetId()))
				d.SetId(fmt.Sprintf("%s_%s", wsID, valueOrZero(policySet.GetId())))
				return []*schema.ResourceData{d}, nil
			}
		}

		nextPage := nextPageFromMeta(list.GetMeta())
		if nextPage == nil {
			break
		}
		queryParams.Pagenumber = nextPage
	}

	return nil, fmt.Errorf("excluded workspace %s has not been added to policy set %s", wsName, pSName)
}
