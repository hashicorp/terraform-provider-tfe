// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceTFEWorkspaceExecutionMode() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFEWorkspaceExecutionModeCreate,
		Update: resourceTFEWorkspaceExecutionModeUpdate,
		Read:   resourceTFEWorkspaceExecutionModeRead,
		Delete: resourceTFEWorkspaceExecutionModeDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceTFEWorkspaceExectionModeImporter,
		},

		CustomizeDiff: validateExecutionMode,

		Schema: map[string]*schema.Schema{
			"agent_pool_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"workspace_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"execution_mode": {
				Type:     schema.TypeString,
				Required: true,
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

func resourceTFEWorkspaceExecutionModeCreate(d *schema.ResourceData, meta interface{}) error {
	return createUpdateWorkspaceExecutionMode(d, meta)
}

func resourceTFEWorkspaceExecutionModeRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

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

func resourceTFEWorkspaceExecutionModeUpdate(d *schema.ResourceData, meta interface{}) error {
	return createUpdateWorkspaceExecutionMode(d, meta)
}

func resourceTFEWorkspaceExecutionModeDelete(d *schema.ResourceData, meta interface{}) error {
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
		AgentPoolID:   tfe.String(""),
		ExecutionMode: tfe.String("remote"),
	})
	if errs != nil {
		return fmt.Errorf("error detaching agent pool from workspace: %w", errs)
	}

	return nil
}

func validateExecutionMode(_ context.Context, d *schema.ResourceDiff, meta interface{}) error {
	executionMode := d.Get("execution_mode").(string)
	agentPoolID := d.Get("agent_pool_id").(string)

	if executionMode == "agent" {
		if d.NewValueKnown("agent_pool_id") && agentPoolID == "" {
			return fmt.Errorf(`agent_pool_id must be provided when execution_mode is "agent"`)
		}
	}

	if executionMode != "agent" && agentPoolID != "" {
		return fmt.Errorf(`execution_mode must be set to "agent" to assign agent_pool_id`)
	}

	if executionMode != "agent" && !d.GetRawConfig().GetAttr("agent_pool_id").IsNull() {
		return fmt.Errorf(`agent_pool_id must be null, when execution_mode is not set to "agent"`)
	}

	return nil
}

func resourceTFEWorkspaceExectionModeImporter(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	config := meta.(ConfiguredClient)

	s := strings.Split(d.Id(), "/")
	if len(s) >= 3 {
		return nil, fmt.Errorf(
			"invalid workspace input format: %s (expected <ORGANIZATION>/<WORKSPACE NAME> or <WORKSPACE ID>)",
			d.Id(),
		)
	} else if len(s) == 2 {
		workspaceID, err := fetchWorkspaceExternalID(s[0]+"/"+s[1], config.Client)
		if err != nil {
			return nil, fmt.Errorf(
				"error retrieving workspace with name %s from organization %s %w", s[1], s[0], err)
		}

		d.SetId(workspaceID)
	}

	return []*schema.ResourceData{d}, nil
}

func createUpdateWorkspaceExecutionMode(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	workspaceID := d.Get("workspace_id").(string)
	executionMode := d.Get("execution_mode").(string)
	agentPoolID := d.Get("agent_pool_id").(string)

	// Create options struct to attach the agent pool to workspace
	options := tfe.WorkspaceUpdateOptions{
		AgentPoolID:   tfe.String(agentPoolID),
		ExecutionMode: tfe.String(executionMode),
	}

	log.Printf("[DEBUG] Update execution mode on workspace: %s", d.Get("workspace_id"))

	workspace, err := config.Client.Workspaces.UpdateByID(ctx, workspaceID, options)
	if err != nil {
		return fmt.Errorf("error updating execution mode %s on workspace ID %s: %w", d.Get("execution_mode"), workspaceID, err)
	}

	d.SetId(workspace.ID)
	d.Set("execution_mode", workspace.ExecutionMode)

	// to avoid a panic, check for agent pool ID == nil
	if workspace.AgentPool == nil {
		d.Set("agent_pool_id", nil)
	} else {
		d.Set("agent_pool_id", workspace.AgentPool.ID)
	}

	return resourceTFEWorkspaceExecutionModeRead(d, meta)
}
