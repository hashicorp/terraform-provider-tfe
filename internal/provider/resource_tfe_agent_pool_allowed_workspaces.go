// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTFEAgentPoolAllowedWorkspaces() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFEAgentPoolAllowedWorkspacesCreate,
		Read:   resourceTFEAgentPoolAllowedWorkspacesRead,
		Update: resourceTFEAgentPoolAllowedWorkspacesUpdate,
		Delete: resourceTFEAgentPoolAllowedWorkspacesDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"agent_pool_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"allowed_workspace_ids": {
				Type:     schema.TypeSet,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceTFEAgentPoolAllowedWorkspacesCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	apID := d.Get("agent_pool_id").(string)

	// Create a new options struct.
	options := tfe.AgentPoolAllowedWorkspacesUpdateOptions{}

	if allowedWorkspaceIDs, allowedWorkspaceSet := d.GetOk("allowed_workspace_ids"); allowedWorkspaceSet {
		options.AllowedWorkspaces = []*tfe.Workspace{}
		for _, workspaceID := range allowedWorkspaceIDs.(*schema.Set).List() {
			if val, ok := workspaceID.(string); ok {
				options.AllowedWorkspaces = append(options.AllowedWorkspaces, &tfe.Workspace{ID: val})
			}
		}
	}

	log.Printf("[DEBUG] Update agent pool: %s", apID)
	_, err := config.Client.AgentPools.UpdateAllowedWorkspaces(ctx, apID, options)
	if err != nil {
		return fmt.Errorf("Error updating agent pool %s: %w", apID, err)
	}

	d.SetId(apID)

	return nil
}

func resourceTFEAgentPoolAllowedWorkspacesRead(d *schema.ResourceData, meta interface{}) error {
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

	var allowedWorkspaceIDs []string
	for _, workspace := range agentPool.AllowedWorkspaces {
		allowedWorkspaceIDs = append(allowedWorkspaceIDs, workspace.ID)
	}
	d.Set("allowed_workspace_ids", allowedWorkspaceIDs)
	d.Set("agent_pool_id", agentPool.ID)

	return nil
}

func resourceTFEAgentPoolAllowedWorkspacesUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	apID := d.Get("agent_pool_id").(string)

	// Create a new options struct.
	options := tfe.AgentPoolAllowedWorkspacesUpdateOptions{
		AllowedWorkspaces: []*tfe.Workspace{},
	}

	if allowedWorkspaceIDs, allowedWorkspaceSet := d.GetOk("allowed_workspace_ids"); allowedWorkspaceSet {
		options.AllowedWorkspaces = []*tfe.Workspace{}
		for _, workspaceID := range allowedWorkspaceIDs.(*schema.Set).List() {
			if val, ok := workspaceID.(string); ok {
				options.AllowedWorkspaces = append(options.AllowedWorkspaces, &tfe.Workspace{ID: val})
			}
		}
	}

	log.Printf("[DEBUG] Update agent pool: %s", apID)
	_, err := config.Client.AgentPools.UpdateAllowedWorkspaces(ctx, apID, options)
	if err != nil {
		return fmt.Errorf("Error updating agent pool %s: %w", apID, err)
	}

	d.SetId(apID)

	return nil
}

func resourceTFEAgentPoolAllowedWorkspacesDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	apID := d.Get("agent_pool_id").(string)

	// Create a new options struct.
	options := tfe.AgentPoolAllowedWorkspacesUpdateOptions{
		AllowedWorkspaces: []*tfe.Workspace{},
	}

	log.Printf("[DEBUG] Update agent pool: %s", apID)
	_, err := config.Client.AgentPools.UpdateAllowedWorkspaces(ctx, apID, options)
	if err != nil {
		return fmt.Errorf("Error updating agent pool %s: %w", apID, err)
	}

	return nil
}
