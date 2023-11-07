// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"log"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTFEWorkspaceAgentPoolExecution() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFEWorkspaceAgentPoolExecutionCreate,
		Read:   resourceTFEWorkspaceAgentPoolExecutionRead,
		Update: resourceTFEWorkspaceAgentPoolExecutionUpdate,
		Delete: resourceTFEWorkspaceAgentPoolExecutionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"agent_pool_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"workspace_id": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceTFEWorkspaceAgentPoolExecutionCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	poolID := d.Get("agent_pool_id").(string)
	workspaceID := d.Get("workspace_id").(string)

	// Create a new options struct to attach the agent pool to workspace
	options := tfe.WorkspaceUpdateOptions{
		AgentPoolID: tfe.String(poolID),
	}

	log.Printf("[DEBUG] Create attachment on workspace with agent pool ID: %s", poolID)
	workspace, err := config.Client.Workspaces.UpdateByID(ctx, workspaceID, options)
	if err != nil {
		return fmt.Errorf("error attaching agent pool ID %s: to workspace ID %s: %w", poolID, workspaceID, err)
	}

	d.SetId(workspace.ID)

	return resourceTFEWorkspaceAgentPoolExecutionRead(d, meta)
}

func resourceTFEWorkspaceAgentPoolExecutionRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Read configuration: %s", d.Id())
	workspace, err := config.Client.Workspaces.ReadByID(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			log.Printf("[DEBUG] Workspace %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading configuration of workspace %s: %w", d.Id(), err)
	}

	d.Set("workspace_id", workspace.ID)

	var poolID string
	if workspace.AgentPool != nil {
		poolID = workspace.AgentPool.ID
	}
	d.Set("agent_pool_id", poolID)

	return nil
}

func resourceTFEWorkspaceAgentPoolExecutionUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	workspaceID := d.Get("workspace_id").(string)

	if d.HasChange("agent_pool_id") {
		poolID := d.Get("agent_pool_id").(string)
		if poolID != "" {
			_, err := config.Client.Workspaces.UpdateByID(ctx, workspaceID, tfe.WorkspaceUpdateOptions{
				AgentPoolID: tfe.String(poolID),
			})
			if err != nil {
				return fmt.Errorf("error updating workspace %s: %w", id, err)
			}
		}
	}

	return resourceTFEWorkspaceAgentPoolExecutionRead(d, meta)
}

// func resourceTFEWorkspaceAgentPoolExecutionDelete(d *schema.ResourceData, meta interface{}) error {
// 	config := meta.(ConfiguredClient)

// 	poolID := d.Get("agent_pool_id").(string)
// 	workpaceID := d.Get("workspace_id").(string)

// 	return nil
// }
