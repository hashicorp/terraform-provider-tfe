// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceTFEWorkspaceAgentPoolExecution() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFEWorkspaceAgentPoolExecutionCreate,
		Read:   resourceTFEWorkspaceAgentPoolExecutionRead,
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
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"execution_mode": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"operations"},
				ValidateFunc: validation.StringInSlice(
					[]string{
						"agent",
						"local",
						"remote",
					},
					false,
				),
			},
		},
	}
}

func resourceTFEWorkspaceAgentPoolExecutionCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	execution_mode := d.Get("execution_mode").(string)
	agent_pool_id := d.Get("agent_pool_id").(string)

	if execution_mode == "agent" && agent_pool_id == "" {
		return fmt.Errorf(`error with either execution mode set as "agent" with no agent_pool_ID,
		or agent_pool_id is set but execution_mode is not set to "agent"`)
	}

	workspaceID := d.Get("workspace_id").(string)

	// Create a new options struct to attach the agent pool to workspace
	options := tfe.WorkspaceUpdateOptions{
		AgentPoolID:   tfe.String(agent_pool_id),
		ExecutionMode: tfe.String(execution_mode),
	}

	log.Printf("[DEBUG] Create attachment on workspace with agent pool ID: %s", agent_pool_id)
	workspace, err := config.Client.Workspaces.UpdateByID(ctx, workspaceID, options)
	if err != nil {
		return fmt.Errorf("error attaching agent pool ID %s to workspace ID %s: %w", agent_pool_id, workspaceID, err)
	}

	d.SetId(workspace.ID)
	// d.Set() will update state file on tfe_workspace for execution_mode and agent pool ID
	d.Set("execution_mode", workspace.ExecutionMode)
	d.Set("agent_pool_id", workspace.AgentPoolID)

	return resourceTFEWorkspaceAgentPoolExecutionRead(d, meta)
}

func resourceTFEWorkspaceAgentPoolExecutionRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Read configuration: %s", d.Id())
	workspace, err := config.Client.Workspaces.ReadByID(ctx, d.Id())
	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
			log.Printf("[DEBUG] Workspace %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading configuration of workspace %s: %w", d.Id(), err)
	}

	d.Set("workspace_id", workspace.ID)

	var poolID string
	if workspace.AgentPool != nil {
		poolID = workspace.AgentPool.ID
	}
	d.Set("agent_pool_id", poolID)

	return nil
}

func resourceTFEWorkspaceAgentPoolExecutionDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	poolID := d.Get("agent_pool_id").(string)
	workspaceID := d.Get("workspace_id").(string)

	log.Printf("[DEBUG] Detach agent pool %s from workspace %s", poolID, workspaceID)

	_, err := config.Client.Workspaces.ReadByID(ctx, workspaceID)
	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
			log.Printf("[DEBUG] Workspace %s no longer exists", workspaceID)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading configuration of workspace %s: %w", workspaceID, err)
	}

	_, errs := config.Client.Workspaces.UpdateByID(ctx, workspaceID, tfe.WorkspaceUpdateOptions{
		AgentPoolID: tfe.String(""),
	})
	if errs != nil {
		return fmt.Errorf("error detaching agent pool %s from workspace %s: %w", poolID, workspaceID, errs)
	}

	return nil
}
