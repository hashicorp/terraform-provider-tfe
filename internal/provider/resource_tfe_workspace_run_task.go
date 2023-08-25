// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"log"
	"strings"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func workspaceRunTaskEnforcementLevels() []string {
	return []string{
		string(tfe.Advisory),
		string(tfe.Mandatory),
	}
}

func workspaceRunTaskStages() []string {
	return []string{
		string(tfe.PrePlan),
		string(tfe.PostPlan),
		string(tfe.PreApply),
	}
}

// nolint: unparam
// Helper function to turn a slice of strings into an english sentence for documentation
func sentenceList(items []string, prefix string, suffix string, conjunction string) string {
	var b strings.Builder
	for i, v := range items {
		fmt.Fprint(&b, prefix, v, suffix)
		if i < len(items)-1 {
			if i < len(items)-2 {
				fmt.Fprint(&b, ", ")
			} else {
				fmt.Fprintf(&b, " %s ", conjunction)
			}
		}
	}
	return b.String()
}

func resourceTFEWorkspaceRunTask() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFEWorkspaceRunTaskCreate,
		Read:   resourceTFEWorkspaceRunTaskRead,
		Delete: resourceTFEWorkspaceRunTaskDelete,
		Update: resourceTFEWorkspaceRunTaskUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: resourceTFEWorkspaceRunTaskImporter,
		},

		Schema: map[string]*schema.Schema{
			"workspace_id": {
				Description: "The id of the workspace to associate the Run task to.",
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
			},

			"task_id": {
				Description: "The id of the Run task to associate to the Workspace.",

				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},

			"enforcement_level": {
				Description: fmt.Sprintf("The enforcement level of the task. Valid values are %s.", sentenceList(
					workspaceRunTaskEnforcementLevels(),
					"`",
					"`",
					"and",
				)),
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice(
					workspaceRunTaskEnforcementLevels(),
					false,
				),
			},

			"stage": {
				Description: fmt.Sprintf("The stage to run the task in. Valid values are %s.", sentenceList(
					workspaceRunTaskStages(),
					"`",
					"`",
					"and",
				)),
				Type:     schema.TypeString,
				Optional: true,
				Default:  tfe.PostPlan,
				ValidateFunc: validation.StringInSlice(
					workspaceRunTaskStages(),
					false,
				),
			},
		},
	}
}

func resourceTFEWorkspaceRunTaskCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	workspaceID := d.Get("workspace_id").(string)
	taskID := d.Get("task_id").(string)

	task, err := config.Client.RunTasks.Read(ctx, taskID)
	if err != nil {
		return fmt.Errorf(
			"Error retrieving task %s: %w", taskID, err)
	}

	ws, err := config.Client.Workspaces.ReadByID(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf(
			"Error retrieving workspace %s: %w", workspaceID, err)
	}
	stage := tfe.Stage(d.Get("stage").(string))

	options := tfe.WorkspaceRunTaskCreateOptions{
		RunTask:          task,
		EnforcementLevel: tfe.TaskEnforcementLevel(d.Get("enforcement_level").(string)),
		Stage:            &stage,
	}

	log.Printf("[DEBUG] Create task %s in workspace %s", task.ID, ws.ID)
	wstask, err := config.Client.WorkspaceRunTasks.Create(ctx, ws.ID, options)
	if err != nil {
		return fmt.Errorf("Error creating task %s in workspace %s: %w", task.ID, ws.ID, err)
	}

	d.SetId(wstask.ID)

	return resourceTFEWorkspaceRunTaskRead(d, meta)
}

func resourceTFEWorkspaceRunTaskDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	// Get the workspace
	workspaceID := d.Get("workspace_id").(string)

	log.Printf("[DEBUG] Delete task %s in workspace %s", d.Id(), workspaceID)
	err := config.Client.WorkspaceRunTasks.Delete(ctx, workspaceID, d.Id())
	if err != nil && !isErrResourceNotFound(err) {
		return fmt.Errorf("Error deleting task %s in workspace %s: %w", d.Id(), workspaceID, err)
	}

	return nil
}

func resourceTFEWorkspaceRunTaskUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	// Get the workspace
	workspaceID := d.Get("workspace_id").(string)

	// Setup the options struct
	options := tfe.WorkspaceRunTaskUpdateOptions{}
	if d.HasChange("enforcement_level") {
		options.EnforcementLevel = tfe.TaskEnforcementLevel(d.Get("enforcement_level").(string))
	}
	if d.HasChange("stage") {
		stage := tfe.Stage(d.Get("stage").(string))
		options.Stage = &stage
	}

	log.Printf("[DEBUG] Update configuration of task %s in workspace %s", d.Id(), workspaceID)
	_, err := config.Client.WorkspaceRunTasks.Update(ctx, workspaceID, d.Id(), options)
	if err != nil {
		return fmt.Errorf("Error updating task %s in workspace %s: %w", d.Id(), workspaceID, err)
	}

	return nil
}

func resourceTFEWorkspaceRunTaskRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	// Get the workspace
	workspaceID := d.Get("workspace_id").(string)

	wstask, err := config.Client.WorkspaceRunTasks.Read(ctx, workspaceID, d.Id())
	if err != nil {
		if isErrResourceNotFound(err) {
			log.Printf("[DEBUG] Workspace Task %s does not exist in workspace %s", d.Id(), workspaceID)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading configuration of task %s in workspace %s: %w", d.Id(), workspaceID, err)
	}

	// Update the config.
	d.Set("workspace_id", wstask.Workspace.ID)
	d.Set("task_id", wstask.RunTask.ID)
	d.Set("enforcement_level", string(wstask.EnforcementLevel))
	d.Set("stage", string(wstask.Stage))

	return nil
}

func resourceTFEWorkspaceRunTaskImporter(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	config := meta.(ConfiguredClient)

	s := strings.Split(d.Id(), "/")
	if len(s) != 3 {
		return nil, fmt.Errorf(
			"invalid task input format: %s (expected <ORGANIZATION>/<WORKSPACE NAME>/<TASK NAME>)",
			d.Id(),
		)
	}

	wstask, err := fetchWorkspaceRunTask(s[2], s[1], s[0], config.Client)
	if err != nil {
		return nil, err
	}

	d.Set("workspace_id", wstask.Workspace.ID)
	d.Set("task_id", wstask.RunTask.ID)
	d.SetId(wstask.ID)

	return []*schema.ResourceData{d}, nil
}
