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

func resourceTFEAgentPoolAllowedWorkspaces() *schema.Resource {
	return &schema.Resource{
		Description: "Adds and removes allowed workspaces on an agent pool." +
			"\n\n~> **Note:** This resource requires using the provider with HCP Terraform and a HCP Terraform for Business account. [Learn more about HCP Terraform pricing here](https://www.hashicorp.com/products/terraform/pricing).",

		Create: resourceTFEAgentPoolAllowedWorkspacesCreate,
		Read:   resourceTFEAgentPoolAllowedWorkspacesRead,
		Update: resourceTFEAgentPoolAllowedWorkspacesUpdate,
		Delete: resourceTFEAgentPoolAllowedWorkspacesDelete,
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

			"allowed_workspace_ids": {
				Description: "IDs of workspaces to be added as allowed workspaces on the agent pool.",
				Type:        schema.TypeSet,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

// buildAgentPoolAllowedWorkspacesBody constructs a v2 PATCH body that sets the
// allowed-workspaces relationship to exactly the given workspace IDs.
func buildAgentPoolAllowedWorkspacesBody(workspaceIDs []string) models.AgentPoolsEnvelopeable {
	body := models.NewAgentPoolsEnvelope()
	pool := models.NewAgentPools()
	poolType := models.AGENTPOOLS_AGENTPOOLS_TYPE
	pool.SetTypeEscaped(&poolType)

	rels := models.NewAgentPools_relationships()
	wsRel := models.NewWorkspacesHasMany()
	data := make([]models.WorkspacesIdentifierable, 0, len(workspaceIDs))
	wsType := models.WORKSPACES_WORKSPACESIDENTIFIER_TYPE
	for _, id := range workspaceIDs {
		id := id
		ident := models.NewWorkspacesIdentifier()
		ident.SetId(&id)
		ident.SetTypeEscaped(&wsType)
		data = append(data, ident)
	}
	wsRel.SetData(data)
	rels.SetAllowedWorkspaces(wsRel)
	pool.SetRelationships(rels)
	body.SetData(pool)
	return body
}

func resourceTFEAgentPoolAllowedWorkspacesCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	apID := d.Get("agent_pool_id").(string)

	var workspaceIDs []string
	if allowedWorkspaceIDs, ok := d.GetOk("allowed_workspace_ids"); ok {
		for _, workspaceID := range allowedWorkspaceIDs.(*schema.Set).List() {
			if val, ok := workspaceID.(string); ok {
				workspaceIDs = append(workspaceIDs, val)
			}
		}
	}

	log.Printf("[DEBUG] Update agent pool: %s", apID)
	_, err := config.ClientV2.API.AgentPools().ByAgent_pool_id(apID).Patch(ctx, buildAgentPoolAllowedWorkspacesBody(workspaceIDs), nil)
	if err != nil {
		return fmt.Errorf("Error updating agent pool %s: %w", apID, err)
	}

	d.SetId(apID)

	return nil
}

func resourceTFEAgentPoolAllowedWorkspacesRead(d *schema.ResourceData, meta interface{}) error {
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

	var allowedWorkspaceIDs []string
	if rels := agentPool.GetRelationships(); rels != nil {
		if aw := rels.GetAllowedWorkspaces(); aw != nil {
			for _, ws := range aw.GetData() {
				if ws != nil {
					allowedWorkspaceIDs = append(allowedWorkspaceIDs, valueOrZero(ws.GetId()))
				}
			}
		}
	}
	d.Set("allowed_workspace_ids", allowedWorkspaceIDs)
	d.Set("agent_pool_id", valueOrZero(agentPool.GetId()))

	return nil
}

func resourceTFEAgentPoolAllowedWorkspacesUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	apID := d.Get("agent_pool_id").(string)

	var workspaceIDs []string
	if allowedWorkspaceIDs, ok := d.GetOk("allowed_workspace_ids"); ok {
		for _, workspaceID := range allowedWorkspaceIDs.(*schema.Set).List() {
			if val, ok := workspaceID.(string); ok {
				workspaceIDs = append(workspaceIDs, val)
			}
		}
	}

	log.Printf("[DEBUG] Update agent pool: %s", apID)
	_, err := config.ClientV2.API.AgentPools().ByAgent_pool_id(apID).Patch(ctx, buildAgentPoolAllowedWorkspacesBody(workspaceIDs), nil)
	if err != nil {
		return fmt.Errorf("Error updating agent pool %s: %w", apID, err)
	}

	d.SetId(apID)

	return nil
}

func resourceTFEAgentPoolAllowedWorkspacesDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	apID := d.Get("agent_pool_id").(string)

	log.Printf("[DEBUG] Update agent pool: %s", apID)
	_, err := config.ClientV2.API.AgentPools().ByAgent_pool_id(apID).Patch(ctx, buildAgentPoolAllowedWorkspacesBody(nil), nil)
	if err != nil {
		return fmt.Errorf("Error updating agent pool %s: %w", apID, err)
	}

	return nil
}
