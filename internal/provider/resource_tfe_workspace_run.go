// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

// NOTE: This is a legacy resource and should be migrated to the Plugin
// Framework if substantial modifications are planned. See
// docs/new-resources.md if planning to use this code as boilerplate for
// a new resource.

package provider

import (
	"context"
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

var applyPendingStatuses = map[tfe.RunStatus]bool{
	tfe.RunConfirmed:          true,
	tfe.RunApplyQueued:        true,
	tfe.RunApplying:           true,
	tfe.RunQueuing:            true,
	tfe.RunFetching:           true,
	tfe.RunQueuingApply:       true,
	tfe.RunPreApplyRunning:    true,
	tfe.RunPreApplyCompleted:  true,
	tfe.RunPostApplyRunning:   true,
	tfe.RunPostApplyCompleted: true,
}

var applyDoneStatuses = map[tfe.RunStatus]bool{
	tfe.RunApplied: true,
	tfe.RunErrored: true,
}

var confirmationPendingStatuses = map[tfe.RunStatus]bool{
	tfe.RunPostPlanCompleted: true,
	tfe.RunPlanned:           true,
	tfe.RunCostEstimated:     true,
	tfe.RunPolicyChecked:     true,
}

var confirmationDoneStatuses = map[tfe.RunStatus]bool{
	tfe.RunConfirmed:         true,
	tfe.RunApplyQueued:       true,
	tfe.RunApplying:          true,
	tfe.RunPrePlanCompleted:  true,
	tfe.RunPrePlanRunning:    true,
	tfe.RunQueuingApply:      true,
	tfe.RunPreApplyCompleted: true,
}

var policyOverriddenStatuses = map[tfe.RunStatus]bool{
	tfe.RunPolicyChecked:    true,
	tfe.RunConfirmed:        true,
	tfe.RunApplyQueued:      true,
	tfe.RunApplying:         true,
	tfe.RunPrePlanCompleted: true,
	tfe.RunPrePlanRunning:   true,
}

var policyOverridePendingStatuses = map[tfe.RunStatus]bool{
	tfe.RunPolicyOverride: true,
}

func resourceTFEWorkspaceRun() *schema.Resource {
	return &schema.Resource{
		Description: "Provides a resource to manage the initial and/or final Terraform run in a given workspace. These initial and final runs often have a special relationship to other things that depend on the workspace's existence, so it can be useful to manage the completion of these runs in the same Terraform configuration that manages the workspace.\n\n" +
			"~> **Note:** Use caution when removing `tfe_workspace_run` from configuration. Destroying with a `destroy` block present creates a destroy run for underlying managed resources.\n\n" +
			"There are a few main use cases this resource was designed for: \n - **Workspaces that depend on other workspaces.** If a workspace will create infrastructure that other workspaces rely on (for example, a Kubernetes cluster to deploy resources into), those downstream workspaces can depend on an initial `apply` with `wait_for_run = true`, so they aren't created before their infrastructure dependencies.\n- **A more reliable `queue_all_runs = true`.** The `queue_all_runs` argument on `tfe_workspace` requests an initial run, which can complete asynchronously outside of the Terraform run that creates the workspace. Unfortunately, it can't be used with workspaces that require variables to be set, because the `tfe_variable` resources themselves depend on the `tfe_workspace`. By managing an initial `apply` with `wait_for_run = false` that depends on your `tfe_variables`, you can accomplish the same goal without a circular dependency.\n- **Safe workspace destruction.** To ensure a workspace's managed resources are destroyed before deleting it, add a `destroy` block with `wait_for_run = true`. When you destroy the `tfe_workspace_run` resource, Terraform will wait for the destroy run to complete before deleting the workspace. This pattern is compatible with the `tfe_workspace` resource's default safe deletion behavior.\nThe `tfe_workspace_run` expects to own exactly one apply during a creation and/or one destroy during a destruction. This implies that even if previous successful applies exist in the workspace, a `tfe_workspace_run` resource that includes an `apply` block will queue a new apply when added to a config.",

		Create:        resourceTFEWorkspaceRunCreate,
		Delete:        resourceTFEWorkspaceRunDelete,
		Read:          resourceTFEWorkspaceRunRead,
		Update:        resourceTFEWorkspaceRunUpdate,
		CustomizeDiff: resourceTFEWorkspaceRunCustomizeDiff,
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The ID of the run created by this resource.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"workspace_id": {
				Description: "ID of the workspace to execute the run.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"apply": {
				Description:  "Adding an apply block ensures an apply run is queued when the resource is created. The block controls settings for the workspace's apply run during creation.",
				Type:         schema.TypeList,
				Elem:         resourceTFEWorkspaceRunSchema(),
				Optional:     true,
				AtLeastOneOf: []string{"apply", "destroy"},
				MaxItems:     1,
			},
			"destroy": {
				Description: "Adding a destroy block ensures a destroy run is queued when the resource is destroyed. The block controls settings for the workspace's destroy run during destruction.",
				Type:        schema.TypeList,
				Elem:        resourceTFEWorkspaceRunSchema(),
				Optional:    true,
				MaxItems:    1,
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

// resourceTFEWorkspaceRunCustomizeDiff rejects allow_config_version_missing on
// the apply block. The flag is only coherent for destroy runs: on the apply
// path, a no-op would set a synthetic ID that resourceTFEWorkspaceRunRead
// cannot resolve, dropping the resource from state and producing a perpetual
// diff. The schema is shared between the apply and destroy blocks, so this diff
// check is where we constrain the flag to destroy only.
func resourceTFEWorkspaceRunCustomizeDiff(_ context.Context, d *schema.ResourceDiff, _ interface{}) error {
	if v, ok := d.GetOk("apply.0.allow_config_version_missing"); ok && v.(bool) {
		return errors.New("allow_config_version_missing is only supported in the destroy block, not the apply block")
	}
	return nil
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
	// First check whether this is a destroy-only run
	_, ok := d.GetOk("apply")
	if !ok {
		// If there's no apply, then there won't be anything to "read" until we
		// do a destroy run. Return now and leave the ID alone, so that we keep
		// the resource in the state and get a destroy run when the time comes.
		log.Printf("[DEBUG] Run %s (random ID) has no apply; nothing to read for refresh", d.Id())
		return nil
	}

	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Read run for: %s", d.Id())
	runID := d.Id()
	_, err := config.Client.Runs.Read(ctx, runID)
	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
			// It would be very strange for this to happen, since runs can't
			// normally be deleted independently. But this *probably* means we
			// never performed the initial apply, so we'll remove the missing
			// run from the state to force an apply to happen.
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
				Description: "If set to true a human will have to manually confirm a plan in HCP Terraform's UI to start an apply. If set to false, this resource will be automatically applied.",
				Type:        schema.TypeBool,
				Required:    true,
			},
			"message": {
				Description: "A custom message to associate with the run. If omitted, the default run message is used.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"retry": {
				Description: "Whether or not to retry on plan or apply errors. When set to true, retry_attempts must also be greater than zero in order for retries to happen. Defaults to true.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
			"retry_attempts": {
				Description: "The number of retry attempts made after an initial error. Defaults to 3.",
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     3,
			},
			"retry_backoff_min": {
				Description: "The minimum time in seconds to backoff before attempting a retry. Defaults to 1.",
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     1,
			},
			"retry_backoff_max": {
				Description: "The maximum time in seconds to backoff before attempting a retry. Defaults to 30.",
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     30,
			},
			"wait_for_run": {
				Description: "Whether or not to wait for a run to reach completion before considering this a success. When set to false, the provider considers the tfe_workspace_run resource to have been created immediately after the run has been queued. When set to true, the provider waits for a successful apply on the target workspace (or a no-change plan). Defaults to true.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
			"allow_config_version_missing": {
				Description: "Whether or not to treat a missing configuration version as a success rather than an error when creating the run. This is only supported in the destroy block and is useful for destroy runs against workspaces that never had a configuration version uploaded (for example, an empty workspace). When set to true and the run cannot be created because the configuration version is missing, the destroy is treated as a no-op success. Setting this in the apply block is not allowed. Defaults to false.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
		},
	}
}
