// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func createWorkspaceRun(d *schema.ResourceData, meta interface{}, isDestroyRun bool, currentRetryAttempts int) error {
	runArgs := getRunArgs(d, isDestroyRun)
	if runArgs == nil {
		return nil
	}

	retryBOMin := runArgs["retry_backoff_min"].(int)
	retryBOMax := runArgs["retry_backoff_max"].(int)
	retry := runArgs["retry"].(bool)
	retryMaxAttempts := runArgs["retry_attempts"].(int)

	isInitialRunAttempt := currentRetryAttempts == 0

	// only perform exponential backoff during retries, not during initial attempt
	if !isInitialRunAttempt {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context canceled: %w", ctx.Err())
		case <-time.After(backoff(float64(retryBOMin), float64(retryBOMax), currentRetryAttempts)):
		}
	}

	config := meta.(ConfiguredClient)

	workspaceID := d.Get("workspace_id").(string)
	log.Printf("[DEBUG] Read workspace by ID %s", workspaceID)
	ws, err := config.Client.Workspaces.ReadByID(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf(
			"error reading workspace %s: %w", workspaceID, err)
	}

	waitForRun := runArgs["wait_for_run"].(bool)
	manualConfirm := runArgs["manual_confirm"].(bool)

	run, err := createRun(config.Client, waitForRun, manualConfirm, isDestroyRun, ws)
	if err != nil {
		return err
	}

	// in fire-and-forget mode, that's all we need to do
	if !waitForRun {
		d.SetId(run.ID)
		return nil
	}

	isPlanOp := true
	hasPostPlanTaskStage, err := readPostPlanTaskStageInRun(config.Client, run.ID)
	if err != nil {
		return err
	}

	planPendingStatuses, planTerminalStatuses := planStatuses(run, hasPostPlanTaskStage)
	run, err = awaitRun(config.Client, run.ID, ws.Organization.Name, isPlanOp, planPendingStatuses, isPlanComplete(planTerminalStatuses))
	if err != nil {
		return err
	}

	if (run.Status == tfe.RunErrored) || (run.Status == tfe.RunStatus(tfe.PolicySoftFailed)) {
		if retry && currentRetryAttempts < retryMaxAttempts {
			currentRetryAttempts++
			log.Printf("[INFO] Run errored during plan, retrying run, retry count: %d", currentRetryAttempts)
			return createWorkspaceRun(d, meta, isDestroyRun, currentRetryAttempts)
		}

		return fmt.Errorf("run errored during plan, use the run ID %s to debug error", run.ID)
	}

	if run.Status == tfe.RunPolicyOverride {
		log.Printf("[INFO] Policy check soft-failed, awaiting manual override for run %q", run.ID)
		run, err = awaitRun(config.Client, run.ID, ws.Organization.Name, isPlanOp, policyOverridePendingStatuses, isManuallyOverriden)
		if err != nil {
			return err
		}
	}

	if !run.HasChanges && !run.AllowEmptyApply {
		run, err = awaitRun(config.Client, run.ID, ws.Organization.Name, isPlanOp, confirmationPendingStatuses, isPlannedAndFinished)
		if err != nil {
			return err
		}
	}
	// A run is complete when it is successfully planned with no changes to apply
	if run.Status == tfe.RunPlannedAndFinished {
		log.Printf("[INFO] Plan finished, no changes to apply")
		d.SetId(run.ID)
		return nil
	}

	// wait for run to be comfirmable before attempting to confirm
	run, err = awaitRun(config.Client, run.ID, ws.Organization.Name, isPlanOp, confirmationPendingStatuses, isConfirmable)
	if err != nil {
		return err
	}

	err = confirmRun(config.Client, manualConfirm, isPlanOp, run, ws)
	if err != nil {
		return err
	}

	isPlanOp = false
	run, err = awaitRun(config.Client, run.ID, ws.Organization.Name, isPlanOp, applyPendingStatuses, isCompleted)
	if err != nil {
		return err
	}

	return completeOrRetryRun(meta, run, d, retry, currentRetryAttempts, retryMaxAttempts, isDestroyRun)
}

func getRunArgs(d *schema.ResourceData, isDestroyRun bool) map[string]interface{} {
	var runArgs map[string]interface{}

	if isDestroyRun {
		// destroy block is optional, if it is not set then destroy action is noop for a destroy type run
		destroyArgs, ok := d.GetOk("destroy")
		if !ok {
			return nil
		}
		runArgs = destroyArgs.([]interface{})[0].(map[string]interface{})
	} else {
		createArgs, ok := d.GetOk("apply")
		if !ok {
			// apply block is optional, if it is not set then set a random ID to allow for consistent result after apply ops
			d.SetId(fmt.Sprintf("%d", rand.New(rand.NewSource(time.Now().UnixNano())).Int()))
			return nil
		}
		runArgs = createArgs.([]interface{})[0].(map[string]interface{})
	}

	return runArgs
}

func createRun(client *tfe.Client, waitForRun bool, manualConfirm bool, isDestroyRun bool, ws *tfe.Workspace) (*tfe.Run, error) {
	// In fire-and-forget mode (waitForRun=false), autoapply is set to !manualConfirm
	// This should be intuitive, as "manual confirm" is the opposite of "auto apply"
	//
	// In apply-and-wait mode (waitForRun=true), autoapply is set to false to give the tfe_workspace_run resource full control of run confirmation.
	autoApply := false
	if !waitForRun {
		autoApply = !manualConfirm
	}

	runConfig := tfe.RunCreateOptions{
		Workspace: ws,
		IsDestroy: tfe.Bool(isDestroyRun),
		Message: tfe.String(fmt.Sprintf(
			"Triggered by tfe_workspace_run resource via terraform-provider-tfe on %s",
			time.Now().Format(time.UnixDate),
		)),
		AutoApply: tfe.Bool(autoApply),
	}
	log.Printf("[DEBUG] Create run for workspace: %s", ws.ID)
	run, err := client.Runs.Create(ctx, runConfig)
	if err != nil {
		return nil, fmt.Errorf(
			"error creating run for workspace %s: %w", ws.ID, err)
	}

	if run == nil {
		log.Printf("[ERROR] The client returned both a nil run and nil error, this should not happen")
		return nil, fmt.Errorf(
			"the client returned both a nil run and nil error for workspace %s, this should not happen", ws.ID)
	}

	log.Printf("[DEBUG] Run %s created for workspace %s", run.ID, ws.ID)
	return run, nil
}

func confirmRun(client *tfe.Client, manualConfirm bool, isPlanOp bool, run *tfe.Run, ws *tfe.Workspace) error {
	// if human approval is required, an apply will auto kick off when run is manually approved
	if manualConfirm {
		confirmationPendingStatus := map[tfe.RunStatus]bool{}
		confirmationPendingStatus[run.Status] = true

		log.Printf("[INFO] Plan complete, waiting for manual confirm before proceeding run %q", run.ID)
		_, err := awaitRun(client, run.ID, ws.Organization.Name, isPlanOp, confirmationPendingStatus, isConfirmed)
		if err != nil {
			return err
		}
	} else {
		// if human approval is NOT required, go ahead and kick off an apply
		log.Printf("[INFO] Plan complete, confirming an apply for run %q", run.ID)
		err := client.Runs.Apply(ctx, run.ID, tfe.RunApplyOptions{
			Comment: tfe.String(fmt.Sprintf("Run confirmed by tfe_workspace_run resource via terraform-provider-tfe on %s",
				time.Now().Format(time.UnixDate))),
		})
		if err != nil {
			refreshed, fetchErr := client.Runs.Read(ctx, run.ID)
			if fetchErr != nil {
				err = fmt.Errorf("%w\n additionally, got an error while reading the run: %s", err, fetchErr.Error())
			}
			return fmt.Errorf("run errored while applying run %s (waited til status %s, currently status %s): %w", run.ID, run.Status, refreshed.Status, err)
		}
	}
	return nil
}

func completeOrRetryRun(meta interface{}, run *tfe.Run, d *schema.ResourceData, retry bool, currentRetryAttempts int, retryMaxAttempts int, isDestroyRun bool) error {
	switch run.Status {
	case tfe.RunApplied:
		log.Printf("[INFO] Apply complete for run %q", run.ID)
		d.SetId(run.ID)
		return nil
	case tfe.RunErrored:
		if retry && currentRetryAttempts < retryMaxAttempts {
			currentRetryAttempts++
			log.Printf("[INFO] Run errored during apply, retrying run, retry count: %d", currentRetryAttempts)
			return createWorkspaceRun(d, meta, isDestroyRun, currentRetryAttempts)
		}
		return fmt.Errorf("run errored during apply, use the run ID %s to debug error", run.ID)
	default:
		// unexpected run states including canceled and discarded is handled by this block
		return fmt.Errorf("run %s entered unexpected state %s, expected %s state", run.ID, run.Status, tfe.RunApplied)
	}
}

func awaitRun(client *tfe.Client, runID string, organization string, isPlanOp bool, runPendingStatus map[tfe.RunStatus]bool, isDone func(*tfe.Run) bool) (*tfe.Run, error) {
	for i := 0; ; i++ {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("context canceled: %w", ctx.Err())
		case <-time.After(backoff(backoffMin, backoffMax, i)):
			log.Printf("[DEBUG] Polling run %s", runID)
			run, err := client.Runs.Read(ctx, runID)
			if err != nil {
				log.Printf("[ERROR] Could not read run %s: %v", runID, err)
				continue
			}

			run, err = hasFinalStatus(client, run, organization, isPlanOp, runPendingStatus, isDone)
			if run == nil && err == nil {
				// if both error and run is nil, then run is still in progress
				continue
			}

			return run, err
		}
	}
}

func hasFinalStatus(client *tfe.Client, run *tfe.Run, organization string, isPlanOp bool, runPendingStatus map[tfe.RunStatus]bool, isDone func(*tfe.Run) bool) (*tfe.Run, error) {
	_, runIsInProgress := runPendingStatus[run.Status]

	switch {
	case isDone(run):
		log.Printf("[INFO] Run %s has reached a terminal state: %s", run.ID, run.Status)
		return run, nil
	case runIsInProgress:
		logRunProgress(client, organization, isPlanOp, run)
		return nil, nil
	case run.Status == tfe.RunCanceled:
		log.Printf("[INFO] Run %s has been canceled, status is %s", run.ID, run.Status)
		return nil, fmt.Errorf("run %s has been canceled, status is %s", run.ID, run.Status)
	default:
		log.Printf("[INFO] Run %s has entered unexpected state: %s", run.ID, run.Status)
		return nil, fmt.Errorf("run %s has entered unexpected state: %s", run.ID, run.Status)
	}
}

func logRunProgress(client *tfe.Client, organization string, isPlanOp bool, run *tfe.Run) {
	log.Printf("[DEBUG] Reading workspace %s", run.Workspace.ID)
	ws, err := client.Workspaces.ReadByID(ctx, run.Workspace.ID)
	if err != nil {
		log.Printf("[ERROR] Unable to read workspace %s: %v", run.Workspace.ID, err)
		return
	}

	// if the workspace is locked and the current run has not started, assume that workspace was locked for other purposes.
	// display a message to indicate that the workspace is waiting to be manually unlocked before the run can proceed
	if ws.Locked && ws.CurrentRun != nil {
		currentRun, err := client.Runs.Read(ctx, ws.CurrentRun.ID)
		if err != nil {
			log.Printf("[ERROR] Unable to read current run %s: %v", ws.CurrentRun.ID, err)
			return
		}

		if currentRun.Status == tfe.RunPending {
			log.Printf("[INFO] Waiting for manually locked workspace to be unlocked")
			return
		}
	}

	// if this run is the current run in it's workspace, display it's position in the organization queue
	if ws.CurrentRun != nil && ws.CurrentRun.ID == run.ID {
		runPositionInOrg, err := readRunPositionInOrgQueue(client, run.ID, organization)
		if err != nil {
			log.Printf("[ERROR] Unable to read run position in organization queue %v", err)
			return
		}

		orgCapacity, err := client.Organizations.ReadCapacity(ctx, organization)
		if err != nil {
			log.Printf("[ERROR] Unable to read capacity for organization %s: %v", organization, err)
			return
		}
		if runPositionInOrg > 0 {
			log.Printf("[INFO] Waiting for %d queued run(s) before starting run", runPositionInOrg-orgCapacity.Running)
			return
		}
	}

	// if this run is not the current run in it's workspace, display it's position in the workspace queue
	runPositionInWorkspace, err := readRunPositionInWorkspaceQueue(client, run.ID, ws.ID, isPlanOp, ws.CurrentRun)
	if err != nil {
		log.Printf("[ERROR] Unable to read run position in workspace queue %v", err)
		return
	}

	if runPositionInWorkspace > 0 {
		log.Printf(
			"[INFO] Waiting for %d run(s) to finish in workspace %s before being queued...",
			runPositionInWorkspace,
			ws.Name,
		)
		return
	}

	log.Printf("[INFO] Waiting for run %s, status is %s", run.ID, run.Status)
}

func readRunPositionInOrgQueue(client *tfe.Client, runID string, organization string) (int, error) {
	position := 0
	options := tfe.ReadRunQueueOptions{}

	for {
		runQueue, err := client.Organizations.ReadRunQueue(ctx, organization, options)
		if err != nil {
			return position, fmt.Errorf("unable to read run queue for organization %s: %w", organization, err)
		}
		for _, item := range runQueue.Items {
			if runID == item.ID {
				position = item.PositionInQueue
				return position, nil
			}
		}

		// Exit the loop when we've seen all pages.
		if runQueue.CurrentPage >= runQueue.TotalPages {
			break
		}

		options.PageNumber = runQueue.NextPage
	}

	return position, nil
}

func readRunPositionInWorkspaceQueue(client *tfe.Client, runID string, wsID string, isPlanOp bool, currentRun *tfe.Run) (int, error) {
	position := 0
	options := tfe.RunListOptions{}
	found := false

	for {
		runList, err := client.Runs.List(ctx, wsID, &options)
		if err != nil {
			return position, fmt.Errorf("unable to read run list for workspace %s: %w", wsID, err)
		}

		for _, item := range runList.Items {
			if !found {
				if runID == item.ID {
					found = true
				}

				continue
			}

			// ignore runs with final states while computing queue count
			switch item.Status {
			case tfe.RunApplied, tfe.RunCanceled, tfe.RunDiscarded, tfe.RunErrored, tfe.RunPlannedAndFinished:
				continue
			case tfe.RunPlanned:
				if isPlanOp {
					continue
				}
			}

			position++

			if currentRun != nil && currentRun.ID == item.ID {
				return position, nil
			}
		}

		// Exit the loop when we've seen all pages.
		if runList.CurrentPage >= runList.TotalPages {
			break
		}

		options.PageNumber = runList.NextPage
	}

	return position, nil
}

// perform exponential backoff based on the iteration and
// limited by the provided min and max durations in milliseconds.
func backoff(min, max float64, iter int) time.Duration {
	backoff := math.Pow(2, float64(iter)/5) * min
	if backoff > max {
		backoff = max
	}
	return time.Duration(backoff) * time.Millisecond
}

func readPostPlanTaskStageInRun(client *tfe.Client, runID string) (bool, error) {
	hasPostPlanTaskStage := false
	options := tfe.TaskStageListOptions{}

	for {
		taskStages, err := client.TaskStages.List(ctx, runID, &options)
		if err != nil {
			return hasPostPlanTaskStage, fmt.Errorf("[ERROR] Could not read task stages for run %s: %v", runID, err)
		}
		for _, item := range taskStages.Items {
			if item.Stage == tfe.PostPlan {
				hasPostPlanTaskStage = true
				return hasPostPlanTaskStage, nil
			}
		}

		// Exit the loop when we've seen all pages.
		if taskStages.CurrentPage >= taskStages.TotalPages {
			break
		}

		options.PageNumber = taskStages.NextPage
	}

	return hasPostPlanTaskStage, nil
}

func planStatuses(run *tfe.Run, hasPostPlanTaskStage bool) (map[tfe.RunStatus]bool, map[tfe.RunStatus]bool) {
	hasPolicyCheck := len(run.PolicyChecks) > 0
	hasCostEstimate := run.CostEstimate != nil

	/*
		tfe.RunCostEstimated and tfe.RunPolicyChecked are optional terminal statuses that may or may not be present in a plan.
		Policy checks are currently the last checks performed as a post plan op if present
		These following statuses are added at compute-time. When plan has changes, the following occur in order:
		1. tfe.RunPolicyChecked is the plan terminal status if present, otherwise
		2. tfe.RunCostEstimated is the plan terminal status if present, otherwise
		3. tfe.RunPostPlanCompleted is the plan terminal status if present, otherwise
		4. tfe.RunPlanned is the plan terminal status if all of above is absent
	*/
	var planTerminalStatuses = map[tfe.RunStatus]bool{
		tfe.RunErrored:            true,
		tfe.RunPlannedAndFinished: true,
		tfe.RunPolicySoftFailed:   true,
		tfe.RunPolicyOverride:     true,
	}

	var planPendingStatuses = map[tfe.RunStatus]bool{
		tfe.RunPending:           true,
		tfe.RunPlanQueued:        true,
		tfe.RunPlanning:          true,
		tfe.RunCostEstimating:    true,
		tfe.RunPolicyChecking:    true,
		tfe.RunQueuing:           true,
		tfe.RunFetching:          true,
		tfe.RunPostPlanRunning:   true,
		tfe.RunPostPlanCompleted: true,
		tfe.RunPrePlanRunning:    true,
		tfe.RunPrePlanCompleted:  true,
	}

	if hasPolicyCheck {
		planTerminalStatuses[tfe.RunPolicyChecked] = true

		planPendingStatuses[tfe.RunCostEstimated] = true
		planPendingStatuses[tfe.RunPlanned] = true
	} else if hasCostEstimate {
		planTerminalStatuses[tfe.RunCostEstimated] = true

		planPendingStatuses[tfe.RunPlanned] = true
	} else if hasPostPlanTaskStage {
		planTerminalStatuses[tfe.RunPostPlanCompleted] = true

		planPendingStatuses[tfe.RunPlanned] = true
	} else {
		// there are no post plan ops
		planTerminalStatuses[tfe.RunPlanned] = true
	}

	return planPendingStatuses, planTerminalStatuses
}

func isPlanComplete(planTerminalStatuses map[tfe.RunStatus]bool) func(run *tfe.Run) bool {
	return func(run *tfe.Run) bool {
		_, found := planTerminalStatuses[run.Status]
		return found
	}
}

func isManuallyOverriden(run *tfe.Run) bool {
	_, found := policyOverriddenStatuses[run.Status]
	return found
}

func isPlannedAndFinished(run *tfe.Run) bool {
	return tfe.RunPlannedAndFinished == run.Status
}

func isConfirmable(run *tfe.Run) bool {
	return run.Actions.IsConfirmable
}

func isConfirmed(run *tfe.Run) bool {
	_, found := confirmationDoneStatuses[run.Status]
	return found
}

func isCompleted(run *tfe.Run) bool {
	_, found := applyDoneStatuses[run.Status]
	return found
}
