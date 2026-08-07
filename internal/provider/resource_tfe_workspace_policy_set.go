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

func resourceTFEWorkspacePolicySet() *schema.Resource {
	return &schema.Resource{
		Description: "Adds and removes policy sets from a workspace." +
			"\n\n~> **Note:** `tfe_policy_set` has an argument `workspace_ids` that should not be used alongside this resource as they manage the same attachments." +
			"\n\n~> **Note:** Tag-based scoping and explicit workspace/project associations are mutually exclusive on a policy set. To switch between them, first remove the existing association (`terraform apply`), then add the new one (`terraform apply`).",

		Create: resourceTFEWorkspacePolicySetCreate,
		Read:   resourceTFEWorkspacePolicySetRead,
		Delete: resourceTFEWorkspacePolicySetDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceTFEWorkspacePolicySetImporter,
		},

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The ID of the policy set attachment. ID format: `<workspace-id>_<policy-set-id>`.",
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
				Description: "Workspace ID to add the policy set to.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
		},
	}
}

func resourceTFEWorkspacePolicySetCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	policySetID := d.Get("policy_set_id").(string)
	workspaceID := d.Get("workspace_id").(string)

	body := makeWorkspaceIdentifierArrayDocument([]interface{}{workspaceID})

	err := config.ClientV2.API.PolicySets().ByPolicy_set_id(policySetID).Relationships().Workspaces().Post(ctx, body, nil)
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
	policySetEnv, err := config.ClientV2.API.PolicySets().ByPolicy_set_id(policySetID).Get(ctx, nil)
	if err != nil {
		if errors.Is(err, tfeV2.ErrNotFound) {
			log.Printf("[DEBUG] Policy set %s no longer exists", policySetID)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading configuration of policy set %s: %w", policySetID, err)
	}

	policySet := policySetEnv.GetData()
	rels := policySet.GetRelationships()

	isWorkspaceAttached := false
	if rels != nil && rels.GetWorkspaces() != nil {
		for _, ws := range rels.GetWorkspaces().GetData() {
			if valueOrZero(ws.GetId()) == workspaceID {
				isWorkspaceAttached = true
				d.Set("workspace_id", workspaceID)
				break
			}
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
	body := makeWorkspaceIdentifierArrayDocument([]interface{}{workspaceID})

	err := config.ClientV2.API.PolicySets().ByPolicy_set_id(policySetID).Relationships().Workspaces().Delete(ctx, body, nil)
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

	pageSize := int32(100)
	queryParams := &organizations.ItemPolicySetsRequestBuilderGetQueryParameters{
		Pagesize:   &pageSize,
		Searchname: &pSName,
	}
	for {
		list, err := config.ClientV2.API.Organizations().ByOrganization_name(organization).PolicySets().Get(ctx, withQueryParams(queryParams))
		if err != nil {
			return nil, fmt.Errorf("Error retrieving policy sets: %w", err)
		}
		for _, policySet := range list.GetData() {
			psAttrs := policySet.GetAttributes()
			if psAttrs == nil || valueOrZero(psAttrs.GetName()) != pSName {
				continue
			}

			rels := policySet.GetRelationships()
			if rels == nil || rels.GetWorkspaces() == nil {
				continue
			}
			for _, ws := range rels.GetWorkspaces().GetData() {
				wsID := valueOrZero(ws.GetId())
				if wsID == "" {
					continue
				}
				// We need the workspace name; fetch it via v1 to compare
				wsObj, err := config.Client.Workspaces.ReadByID(ctx, wsID)
				if err != nil || wsObj == nil || wsObj.Name != wsName {
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

	return nil, fmt.Errorf("workspace %s has not been assigned to policy set %s", wsName, pSName)
}
