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
		},
	}
}

func resourceTFEWorkspaceAgentPoolExecutionCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	poolID := d.Get("agent_pool_id").(string)
	workspaceID := d.Get("workspace_id").(string)

	// Create a new options struct to attach the agent pool to workspace
	options := tfe.WorkspaceUpdateOptions{
		AgentPoolID:   tfe.String(poolID),
		ExecutionMode: tfe.String("agent"),
	}

	log.Printf("[DEBUG] Create attachment on workspace with agent pool ID: %s", poolID)
	workspace, err := config.Client.Workspaces.UpdateByID(ctx, workspaceID, options)
	if err != nil {
		return fmt.Errorf("error attaching agent pool ID %s to workspace ID %s: %w", poolID, workspaceID, err)
	}
	// log a WARN if exeuction mode is "agent" because this mode requires
	// also providing an agent_pool_id to the workspace resource and that attr is now deprecated
	if workspace.ExecutionMode == "agent" {
		warnWorkspaceAgentPoolIDDeprecation()
	}

	d.SetId(workspace.ID)

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

func warnWorkspaceAgentPoolIDDeprecation() {
	log.Printf("[WARN] The agent_pool_id field of tfe_workspace is deprecated as of release 0.50.0 and may be removed in a future version. The preferred method of associating an agent pool to a workspace is by using the tfe_workspace_agent_pool_execution resource.")
}
