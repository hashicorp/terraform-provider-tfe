package tfe

import (
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var (
	backoffMin = 1000.0
	backoffMax = 3000.0
)

var planPendingStatuses = map[tfe.RunStatus]bool{
	tfe.RunPending:        true,
	tfe.RunPlanQueued:     true,
	tfe.RunPlanning:       true,
	tfe.RunCostEstimating: true,
	tfe.RunPolicyChecking: true,
	tfe.RunQueuing:        true,
	tfe.RunFetching:       true,
}

var planTerminalStatuses = map[tfe.RunStatus]bool{
	tfe.RunPlanned:            true,
	tfe.RunPlannedAndFinished: true,
	tfe.RunErrored:            true,
	tfe.RunCostEstimated:      true,
	tfe.RunPolicyChecked:      true,
	tfe.RunPolicySoftFailed:   true,
	tfe.RunPolicyOverride:     true,
}

var applyPendingStatuses = map[tfe.RunStatus]bool{
	tfe.RunConfirmed:   true,
	tfe.RunApplyQueued: true,
	tfe.RunApplying:    true,
	tfe.RunQueuing:     true,
	tfe.RunFetching:    true,
}

var applyDoneStatuses = map[tfe.RunStatus]bool{
	tfe.RunApplied: true,
	tfe.RunErrored: true,
}

var confirmationDoneStatuses = map[tfe.RunStatus]bool{
	tfe.RunConfirmed:   true,
	tfe.RunApplyQueued: true,
	tfe.RunApplying:    true,
}

func resourceTFEWorkspaceRun() *schema.Resource {
	return &schema.Resource{
		Create:        resourceTFEWorkspaceRunCreate,
		Delete:        resourceTFEWorkspaceRunDelete,
		Read:          resourceTFEWorkspaceRunRead,
		Update:        resourceTFEWorkspaceRunUpdate,
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"workspace_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"apply": {
				Type:         schema.TypeList,
				Elem:         resourceTFEWorkspaceRunSchema(),
				Optional:     true,
				AtLeastOneOf: []string{"apply", "destroy"},
				MaxItems:     1,
			},
			"destroy": {
				Type:     schema.TypeList,
				Elem:     resourceTFEWorkspaceRunSchema(),
				Optional: true,
				MaxItems: 1,
			},
		},
	}
}

func resourceTFEWorkspaceRunCreate(d *schema.ResourceData, meta interface{}) error {
	// var isDestroyRun & currentRetryAttempts is declared for the sole purpose of code readability
	isDestroyRun := false
	currentRetryAttempts := 0
	return createWorkspaceRun(d, meta, isDestroyRun, currentRetryAttempts)
}

func resourceTFEWorkspaceRunDelete(d *schema.ResourceData, meta interface{}) error {
	// var isDestroyRun & currentRetryAttempts is declared for the sole purpose of code readability
	isDestroyRun := true
	currentRetryAttempts := 0
	return createWorkspaceRun(d, meta, isDestroyRun, currentRetryAttempts)
}

func resourceTFEWorkspaceRunUpdate(d *schema.ResourceData, meta interface{}) error {
	// update is a noop since this resource only creates a run during a destroy or an initial apply phase
	return nil
}

func resourceTFEWorkspaceRunRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Read run for: %s", d.Id())
	runID := d.Id()
	_, err := config.Client.Runs.Read(ctx, runID)
	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
			log.Printf("[DEBUG] Run %s does not exist", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading run %s: %w", d.Id(), err)
	}

	return nil
}

func resourceTFEWorkspaceRunSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"manual_confirm": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"retry": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"retry_attempts": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  3,
			},
			"retry_backoff_min": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1,
			},
			"retry_backoff_max": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  30,
			},
			"wait_for_run": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
		},
	}
}
