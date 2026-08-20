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

func resourceTFEProjectPolicySet() *schema.Resource {
	return &schema.Resource{
		Description: "Adds and removes policy sets from a project." +
			"\n\nPolicies are rules enforced on Terraform runs. Two policy-as-code frameworks are integrated with Terraform Enterprise: Sentinel and Open Policy Agent (OPA)." +
			"\n\nPolicy sets are groups of policies that are applied together to related workspaces. By using policy sets, you can group your policies by attributes such as environment or region. Individual policies that are members of policy sets will only be checked for workspaces that the policy set is attached to." +
			"\n\n~> **Note:** Tag-based scoping and explicit workspace/project associations are mutually exclusive on a policy set. To switch between them, first remove the existing association (`terraform apply`), then add the new one (`terraform apply`).",

		Create: resourceTFEProjectPolicySetCreate,
		Read:   resourceTFEProjectPolicySetRead,
		Delete: resourceTFEProjectPolicySetDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceTFEProjectPolicySetImporter,
		},

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The ID of the policy set attachment. ID format: `<project-id>_<policy-set-id>`.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"policy_set_id": {
				Description: "ID of the policy set.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},

			"project_id": {
				Description: "Project ID to add the policy set to.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
		},
	}
}

func resourceTFEProjectPolicySetCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	policySetID := d.Get("policy_set_id").(string)
	projectID := d.Get("project_id").(string)

	body := makeProjectIdentifierArrayDocument([]interface{}{projectID})
	err := config.ClientV2.API.PolicySets().ByPolicy_set_id(policySetID).Relationships().Projects().Post(ctx, body, nil)
	if err != nil {
		return fmt.Errorf(
			"error attaching policy set id %s to project %s: %w", policySetID, projectID, err)
	}

	d.SetId(fmt.Sprintf("%s_%s", projectID, policySetID))

	return resourceTFEProjectPolicySetRead(d, meta)
}

func resourceTFEProjectPolicySetRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	policySetID := d.Get("policy_set_id").(string)
	projectID := d.Get("project_id").(string)

	log.Printf("[DEBUG] Read configuration of project policy set: %s", policySetID)
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

	isProjectAttached := false
	if rels != nil && rels.GetProjects() != nil {
		for _, proj := range rels.GetProjects().GetData() {
			if valueOrZero(proj.GetId()) == projectID {
				isProjectAttached = true
				d.Set("project_id", projectID)
				break
			}
		}
	}

	if !isProjectAttached {
		log.Printf("[DEBUG] Project %s not attached to policy set %s. Removing from state.", projectID, policySetID)
		d.SetId("")
		return nil
	}

	d.Set("policy_set_id", policySetID)
	return nil
}

func resourceTFEProjectPolicySetDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	policySetID := d.Get("policy_set_id").(string)
	projectID := d.Get("project_id").(string)

	log.Printf("[DEBUG] Detaching project (%s) from policy set (%s)", projectID, policySetID)
	body := makeProjectIdentifierArrayDocument([]interface{}{projectID})
	err := config.ClientV2.API.PolicySets().ByPolicy_set_id(policySetID).Relationships().Projects().Delete(ctx, body, nil)
	if err != nil {
		return fmt.Errorf(
			"error detaching project %s from policy set %s: %w", projectID, policySetID, err)
	}

	return nil
}

func resourceTFEProjectPolicySetImporter(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// The format of the import ID is <ORGANIZATION/PROJECT ID/POLICYSET NAME>
	splitID := strings.SplitN(d.Id(), "/", 3)
	if len(splitID) != 3 {
		return nil, fmt.Errorf(
			"invalid project policy set input format: %s (expected <ORGANIZATION>/<PROJECT ID>/<POLICYSET NAME>)",
			splitID,
		)
	}

	organization, projectID, policySetName := splitID[0], splitID[1], splitID[2]

	config := meta.(ConfiguredClient)

	// Ensure the named project exists before fetching all the policy sets in the org
	_, err := config.ClientV2.API.Projects().ByProject_id(projectID).Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error reading configuration of project %s in organization %s: %w", projectID, organization, err)
	}

	pageSize := int32(100)
	queryParams := &organizations.ItemPolicySetsRequestBuilderGetQueryParameters{
		Pagesize:   &pageSize,
		Searchname: &policySetName,
	}
	for {
		list, err := config.ClientV2.API.Organizations().ByOrganization_name(organization).PolicySets().Get(ctx, withQueryParams(queryParams))
		if err != nil {
			return nil, fmt.Errorf("error retrieving organization's list of policy sets: %w", err)
		}
		for _, policySet := range list.GetData() {
			psAttrs := policySet.GetAttributes()
			if psAttrs == nil || valueOrZero(psAttrs.GetName()) != policySetName {
				continue
			}

			rels := policySet.GetRelationships()
			if rels == nil || rels.GetProjects() == nil {
				continue
			}
			for _, proj := range rels.GetProjects().GetData() {
				if valueOrZero(proj.GetId()) != projectID {
					continue
				}

				d.Set("project_id", projectID)
				d.Set("policy_set_id", valueOrZero(policySet.GetId()))
				d.SetId(fmt.Sprintf("%s_%s", projectID, valueOrZero(policySet.GetId())))
				return []*schema.ResourceData{d}, nil
			}
		}

		nextPage := nextPageFromMeta(list.GetMeta())
		if nextPage == nil {
			break
		}
		queryParams.Pagenumber = nextPage
	}

	return nil, fmt.Errorf("project %s has not been assigned to policy set %s", projectID, policySetName)
}
