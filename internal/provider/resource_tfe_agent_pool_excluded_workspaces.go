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

func resourceTFEAgentPoolExcludedWorkspaces() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFEAgentPoolExcludedWorkspacesCreate,
		Read:   resourceTFEAgentPoolExcludedWorkspacesRead,
		Update: resourceTFEAgentPoolExcludedWorkspacesUpdate,
		Delete: resourceTFEAgentPoolExcludedWorkspacesDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"agent_pool_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"excluded_workspace_ids": {
				Type:     schema.TypeSet,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceTFEAgentPoolExcludedWorkspacesCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	apID := d.Get("agent_pool_id").(string)

	// Create a new options struct.
	options := tfe.AgentPoolExcludedWorkspacesUpdateOptions{}

	if excludedWorkspaceIDs, excludedWorkspaceSet := d.GetOk("excluded_workspace_ids"); excludedWorkspaceSet {
		options.ExcludedWorkspaces = []*tfe.Workspace{}
		for _, workspaceID := range excludedWorkspaceIDs.(*schema.Set).List() {
			if val, ok := workspaceID.(string); ok {
				options.ExcludedWorkspaces = append(options.ExcludedWorkspaces, &tfe.Workspace{ID: val})
			}
		}
	}

	log.Printf("[DEBUG] Update agent pool: %s", apID)
	_, err := config.Client.AgentPools.UpdateExcludedWorkspaces(ctx, apID, options)
	if err != nil {
		return fmt.Errorf("Error updating agent pool %s: %w", apID, err)
	}

	d.SetId(apID)

	return nil
}

func resourceTFEAgentPoolExcludedWorkspacesRead(d *schema.ResourceData, meta interface{}) error {
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

	var excludedWorkspaceIDs []string
	for _, workspace := range agentPool.ExcludedWorkspaces {
		excludedWorkspaceIDs = append(excludedWorkspaceIDs, workspace.ID)
	}
	d.Set("excluded_workspace_ids", excludedWorkspaceIDs)
	d.Set("agent_pool_id", agentPool.ID)

	return nil
}

func resourceTFEAgentPoolExcludedWorkspacesUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	apID := d.Get("agent_pool_id").(string)

	// Create a new options struct.
	options := tfe.AgentPoolExcludedWorkspacesUpdateOptions{
		ExcludedWorkspaces: []*tfe.Workspace{},
	}

	if excludedWorkspaceIDs, excludedWorkspaceSet := d.GetOk("excluded_workspace_ids"); excludedWorkspaceSet {
		options.ExcludedWorkspaces = []*tfe.Workspace{}
		for _, workspaceID := range excludedWorkspaceIDs.(*schema.Set).List() {
			if val, ok := workspaceID.(string); ok {
				options.ExcludedWorkspaces = append(options.ExcludedWorkspaces, &tfe.Workspace{ID: val})
			}
		}
	}

	log.Printf("[DEBUG] Update agent pool: %s", apID)
	_, err := config.Client.AgentPools.UpdateExcludedWorkspaces(ctx, apID, options)
	if err != nil {
		return fmt.Errorf("Error updating agent pool %s: %w", apID, err)
	}

	d.SetId(apID)

	return nil
}

func resourceTFEAgentPoolExcludedWorkspacesDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	apID := d.Get("agent_pool_id").(string)

	// Create a new options struct.
	options := tfe.AgentPoolExcludedWorkspacesUpdateOptions{
		ExcludedWorkspaces: []*tfe.Workspace{},
	}

	log.Printf("[DEBUG] Update agent pool: %s", apID)
	_, err := config.Client.AgentPools.UpdateExcludedWorkspaces(ctx, apID, options)
	if err != nil {
		return fmt.Errorf("Error updating agent pool %s: %w", apID, err)
	}

	return nil
}
