// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

// NOTE: This is a legacy resource and should be migrated to the Plugin
// Framework if substantial modifications are planned. See
// docs/new-resources.md if planning to use this code as boilerplate for
// a new resource.

package provider

import (
	"errors"
	"fmt"
	"log"

	tfev2 "github.com/hashicorp/go-tfe/v2"
	"github.com/hashicorp/go-tfe/v2/api/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTFEAgentPoolAllowedProjects() *schema.Resource {
	return &schema.Resource{
		Description: "Adds and removes allowed projects on an agent pool." +
			"\n\n~> **Note:** This resource requires using the provider with HCP Terraform and a HCP Terraform for Business tier plan. [Learn more about HCP Terraform pricing here](https://www.hashicorp.com/products/terraform/pricing).",

		Create: resourceTFEAgentPoolAllowedProjectsCreate,
		Read:   resourceTFEAgentPoolAllowedProjectsRead,
		Update: resourceTFEAgentPoolAllowedProjectsUpdate,
		Delete: resourceTFEAgentPoolAllowedProjectsDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The ID of this resource. Do not rely on this value — use `agent_pool_id` instead.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"agent_pool_id": {
				Description: "The ID of the agent pool.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},

			"allowed_project_ids": {
				Description: "IDs of projects to be added as allowed projects on the agent pool.",
				Type:        schema.TypeSet,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

// buildAgentPoolAllowedProjectsBody constructs a v2 PATCH body that sets the
// allowed-projects relationship to exactly the given project IDs.
func buildAgentPoolAllowedProjectsBody(projectIDs []string) models.AgentPoolsEnvelopeable {
	body := models.NewAgentPoolsEnvelope()
	pool := models.NewAgentPools()
	poolType := models.AGENTPOOLS_AGENTPOOLS_TYPE
	pool.SetTypeEscaped(&poolType)

	rels := models.NewAgentPools_relationships()
	projRel := models.NewProjectsHasMany()
	data := make([]models.ProjectsIdentifierable, 0, len(projectIDs))
	projType := models.PROJECTS_PROJECTSIDENTIFIER_TYPE
	for _, id := range projectIDs {
		ident := models.NewProjectsIdentifier()
		ident.SetId(&id)
		ident.SetTypeEscaped(&projType)
		data = append(data, ident)
	}
	projRel.SetData(data)
	rels.SetAllowedProjects(projRel)
	pool.SetRelationships(rels)
	body.SetData(pool)
	return body
}

func resourceTFEAgentPoolAllowedProjectsCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	apID := d.Get("agent_pool_id").(string)

	var projectIDs []string
	if allowedProjectIDs, ok := d.GetOk("allowed_project_ids"); ok {
		for _, projectID := range allowedProjectIDs.(*schema.Set).List() {
			if val, ok := projectID.(string); ok {
				projectIDs = append(projectIDs, val)
			}
		}
	}

	log.Printf("[DEBUG] Update agent pool: %s", apID)
	_, err := config.ClientV2.API.AgentPools().ByAgent_pool_id(apID).Patch(ctx, buildAgentPoolAllowedProjectsBody(projectIDs), nil)
	if err != nil {
		return fmt.Errorf("Error updating agent pool %s: %w", apID, err)
	}

	d.SetId(apID)

	return nil
}

func resourceTFEAgentPoolAllowedProjectsRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	env, err := config.ClientV2.API.AgentPools().ByAgent_pool_id(d.Id()).Get(ctx, nil)
	if err != nil {
		if errors.Is(err, tfev2.ErrNotFound) {
			log.Printf("[DEBUG] agent pool %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading configuration of agent pool %s: %w", d.Id(), err)
	}

	agentPool := env.GetData()
	if agentPool == nil {
		log.Printf("[DEBUG] agent pool %s no longer exists", d.Id())
		d.SetId("")
		return nil
	}

	var allowedProjectIDs []string
	if rels := agentPool.GetRelationships(); rels != nil {
		if ap := rels.GetAllowedProjects(); ap != nil {
			for _, proj := range ap.GetData() {
				if proj != nil {
					allowedProjectIDs = append(allowedProjectIDs, valueOrZero(proj.GetId()))
				}
			}
		}
	}
	d.Set("allowed_project_ids", allowedProjectIDs)
	d.Set("agent_pool_id", valueOrZero(agentPool.GetId()))

	return nil
}

func resourceTFEAgentPoolAllowedProjectsUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	apID := d.Get("agent_pool_id").(string)

	var projectIDs []string
	if allowedProjectIDs, ok := d.GetOk("allowed_project_ids"); ok {
		for _, projectID := range allowedProjectIDs.(*schema.Set).List() {
			if val, ok := projectID.(string); ok {
				projectIDs = append(projectIDs, val)
			}
		}
	}

	log.Printf("[DEBUG] Update agent pool: %s", apID)
	_, err := config.ClientV2.API.AgentPools().ByAgent_pool_id(apID).Patch(ctx, buildAgentPoolAllowedProjectsBody(projectIDs), nil)
	if err != nil {
		return fmt.Errorf("Error updating agent pool %s: %w", apID, err)
	}

	d.SetId(apID)

	return nil
}

func resourceTFEAgentPoolAllowedProjectsDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	apID := d.Get("agent_pool_id").(string)

	log.Printf("[DEBUG] Update agent pool: %s", apID)
	_, err := config.ClientV2.API.AgentPools().ByAgent_pool_id(apID).Patch(ctx, buildAgentPoolAllowedProjectsBody(nil), nil)
	if err != nil {
		return fmt.Errorf("Error updating agent pool %s: %w", apID, err)
	}

	return nil
}
