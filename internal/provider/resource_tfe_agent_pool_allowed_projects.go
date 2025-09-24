// Copyright (c) HashiCorp, Inc.
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

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTFEAgentPoolAllowedProjects() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFEAgentPoolAllowedProjectsCreate,
		Read:   resourceTFEAgentPoolAllowedProjectsRead,
		Update: resourceTFEAgentPoolAllowedProjectsUpdate,
		Delete: resourceTFEAgentPoolAllowedProjectsDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"agent_pool_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"allowed_project_ids": {
				Type:     schema.TypeSet,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceTFEAgentPoolAllowedProjectsCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	apID := d.Get("agent_pool_id").(string)

	// Create a new options struct.
	options := tfe.AgentPoolAllowedProjectsUpdateOptions{}

	if allowedProjectIDs, allowedProjectSet := d.GetOk("allowed_project_ids"); allowedProjectSet {
		options.AllowedProjects = []*tfe.Project{}
		for _, projectID := range allowedProjectIDs.(*schema.Set).List() {
			if val, ok := projectID.(string); ok {
				options.AllowedProjects = append(options.AllowedProjects, &tfe.Project{ID: val})
			}
		}
	}

	log.Printf("[DEBUG] Update agent pool: %s", apID)
	_, err := config.Client.AgentPools.UpdateAllowedProjects(ctx, apID, options)
	if err != nil {
		return fmt.Errorf("Error updating agent pool %s: %w", apID, err)
	}

	d.SetId(apID)

	return nil
}

func resourceTFEAgentPoolAllowedProjectsRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	agentPool, err := config.Client.AgentPools.Read(ctx, d.Id())
	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
			log.Printf("[DEBUG] agent pool %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading configuration of agent pool %s: %w", d.Id(), err)
	}

	var allowedProjectIDs []string
	for _, project := range agentPool.AllowedProjects {
		allowedProjectIDs = append(allowedProjectIDs, project.ID)
	}
	d.Set("allowed_project_ids", allowedProjectIDs)
	d.Set("agent_pool_id", agentPool.ID)

	return nil
}

func resourceTFEAgentPoolAllowedProjectsUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	apID := d.Get("agent_pool_id").(string)

	// Create a new options struct.
	options := tfe.AgentPoolAllowedProjectsUpdateOptions{
		AllowedProjects: []*tfe.Project{},
	}

	if allowedProjectIDs, allowedProjectSet := d.GetOk("allowed_project_ids"); allowedProjectSet {
		options.AllowedProjects = []*tfe.Project{}
		for _, projectID := range allowedProjectIDs.(*schema.Set).List() {
			if val, ok := projectID.(string); ok {
				options.AllowedProjects = append(options.AllowedProjects, &tfe.Project{ID: val})
			}
		}
	}

	log.Printf("[DEBUG] Update agent pool: %s", apID)
	_, err := config.Client.AgentPools.UpdateAllowedProjects(ctx, apID, options)
	if err != nil {
		return fmt.Errorf("Error updating agent pool %s: %w", apID, err)
	}

	d.SetId(apID)

	return nil
}

func resourceTFEAgentPoolAllowedProjectsDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	apID := d.Get("agent_pool_id").(string)

	// Create a new options struct.
	options := tfe.AgentPoolAllowedProjectsUpdateOptions{
		AllowedProjects: []*tfe.Project{},
	}

	log.Printf("[DEBUG] Update agent pool: %s", apID)
	_, err := config.Client.AgentPools.UpdateAllowedProjects(ctx, apID, options)
	if err != nil {
		return fmt.Errorf("Error updating agent pool %s: %w", apID, err)
	}

	return nil
}
