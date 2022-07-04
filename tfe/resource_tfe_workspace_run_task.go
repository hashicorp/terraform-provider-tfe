package tfe

import (
	"fmt"
	"log"
	"strings"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceTFEWorkspaceRunTask() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFEWorkspaceRunTaskCreate,
		Read:   resourceTFEWorkspaceRunTaskRead,
		Delete: resourceTFEWorkspaceRunTaskDelete,
		Update: resourceTFEWorkspaceRunTaskUpdate,
		Importer: &schema.ResourceImporter{
			State: resourceTFEWorkspaceRunTaskImporter,
		},

		Schema: map[string]*schema.Schema{
			"workspace_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},

			"task_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},

			"enforcement_level": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice(
					[]string{
						string(tfe.Advisory),
						string(tfe.Mandatory),
					},
					false,
				),
			},
		},
	}
}

func resourceTFEWorkspaceRunTaskCreate(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	workspaceID := d.Get("workspace_id").(string)
	taskID := d.Get("task_id").(string)

	task, err := tfeClient.RunTasks.Read(ctx, taskID)
	if err != nil {
		return fmt.Errorf(
			"Error retrieving task %s: %w", taskID, err)
	}

	ws, err := tfeClient.Workspaces.ReadByID(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf(
			"Error retrieving workspace %s: %w", workspaceID, err)
	}

	options := tfe.WorkspaceRunTaskCreateOptions{
		RunTask:          task,
		EnforcementLevel: tfe.TaskEnforcementLevel(d.Get("enforcement_level").(string)),
	}

	log.Printf("[DEBUG] Create task %s in workspace %s", task.ID, ws.ID)
	wstask, err := tfeClient.WorkspaceRunTasks.Create(ctx, ws.ID, options)
	if err != nil {
		return fmt.Errorf("Error creating task %s in workspace %s: %w", task.ID, ws.ID, err)
	}

	d.SetId(wstask.ID)

	return resourceTFEWorkspaceRunTaskRead(d, meta)
}

func resourceTFEWorkspaceRunTaskDelete(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	// Get the workspace
	workspaceID := d.Get("workspace_id").(string)

	log.Printf("[DEBUG] Delete task %s in workspace %s", d.Id(), workspaceID)
	err := tfeClient.WorkspaceRunTasks.Delete(ctx, workspaceID, d.Id())
	if err != nil && !isErrResourceNotFound(err) {
		return fmt.Errorf("Error deleting task %s in workspace %s: %w", d.Id(), workspaceID, err)
	}

	return nil
}

func resourceTFEWorkspaceRunTaskUpdate(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	// Get the workspace
	workspaceID := d.Get("workspace_id").(string)

	// Setup the options struct
	options := tfe.WorkspaceRunTaskUpdateOptions{}
	if d.HasChange("enforcement_level") {
		options.EnforcementLevel = tfe.TaskEnforcementLevel(d.Get("enforcement_level").(string))
	}

	log.Printf("[DEBUG] Update configuration of task %s in workspace %s", d.Id(), workspaceID)
	_, err := tfeClient.WorkspaceRunTasks.Update(ctx, workspaceID, d.Id(), options)
	if err != nil {
		return fmt.Errorf("Error updating task %s in workspace %s: %w", d.Id(), workspaceID, err)
	}

	return nil
}

func resourceTFEWorkspaceRunTaskRead(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	// Get the workspace
	workspaceID := d.Get("workspace_id").(string)

	wstask, err := tfeClient.WorkspaceRunTasks.Read(ctx, workspaceID, d.Id())
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

	return nil
}

func resourceTFEWorkspaceRunTaskImporter(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	tfeClient := meta.(*tfe.Client)

	s := strings.Split(d.Id(), "/")
	if len(s) != 3 {
		return nil, fmt.Errorf(
			"invalid task input format: %s (expected <ORGANIZATION>/<WORKSPACE NAME>/<TASK NAME>)",
			d.Id(),
		)
	}

	wstask, err := fetchWorkspaceRunTask(s[2], s[1], s[0], tfeClient)
	if err != nil {
		return nil, err
	}

	d.Set("workspace_id", wstask.Workspace.ID)
	d.Set("task_id", wstask.RunTask.ID)
	d.SetId(wstask.ID)

	return []*schema.ResourceData{d}, nil
}
