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

func resourceTFEProjectPolicySet() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFEProjectPolicySetCreate,
		Read:   resourceTFEProjectPolicySetRead,
		Delete: resourceTFEProjectPolicySetDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceTFEProjectPolicySetImporter,
		},

		Schema: map[string]*schema.Schema{
			"policy_set_id": {
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

func resourceTFEProjectPolicySetCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	policySetID := d.Get("policy_set_id").(string)
	projectID := d.Get("project_id").(string)

	policySetAddProjectsOptions := tfe.PolicySetAddProjectsOptions{}
	policySetAddProjectsOptions.Projects = append(policySetAddProjectsOptions.Projects, &tfe.Project{ID: projectID})

	err := config.Client.PolicySets.AddProjects(ctx, policySetID, policySetAddProjectsOptions)
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
	policySet, err := config.Client.PolicySets.ReadWithOptions(ctx, policySetID, &tfe.PolicySetReadOptions{
		Include: []tfe.PolicySetIncludeOpt{tfe.PolicySetProjects},
	})
	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
			log.Printf("[DEBUG] Policy set %s no longer exists", policySetID)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading configuration of policy set %s: %w", policySetID, err)
	}

	isProjectAttached := false
	for _, project := range policySet.Projects {
		if project.ID == projectID {
			isProjectAttached = true
			d.Set("project_id", projectID)
			break
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
	policySetRemoveProjectsOptions := tfe.PolicySetRemoveProjectsOptions{}
	policySetRemoveProjectsOptions.Projects = append(policySetRemoveProjectsOptions.Projects, &tfe.Project{ID: projectID})

	err := config.Client.PolicySets.RemoveProjects(ctx, policySetID, policySetRemoveProjectsOptions)
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
	_, err := config.Client.Projects.Read(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("error reading configuration of project %s in organization %s: %w", projectID, organization, err)
	}

	options := &tfe.PolicySetListOptions{Include: []tfe.PolicySetIncludeOpt{tfe.PolicySetProjects}}
	for {
		list, err := config.Client.PolicySets.List(ctx, organization, options)
		if err != nil {
			return nil, fmt.Errorf("error retrieving organization's list of policy sets: %w", err)
		}
		for _, policySet := range list.Items {
			if policySet.Name != policySetName {
				continue
			}

			for _, project := range policySet.Projects {
				if project.ID != projectID {
					continue
				}

				d.Set("project_id", project.ID)
				d.Set("policy_set_id", policySet.ID)
				d.SetId(fmt.Sprintf("%s_%s", project.ID, policySet.ID))

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

	return nil, fmt.Errorf("project %s has not been assigned to policy set %s", projectID, policySetName)
}
