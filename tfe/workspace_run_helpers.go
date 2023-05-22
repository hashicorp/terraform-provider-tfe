// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tfe

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

	runConfig := tfe.RunCreateOptions{
		Workspace: ws,
		IsDestroy: tfe.Bool(isDestroyRun),
		Message: tfe.String(fmt.Sprintf(
			"Triggered by tfe_workspace_run resource via terraform-provider-tfe on %s",
			time.Now().Format(time.UnixDate),
		)),
		// autoapply is set to false to give the tfe_workspace_run resource full control of run confirmation.
		AutoApply: tfe.Bool(false),
	}
	log.Printf("[DEBUG] Create run for workspace: %s", ws.ID)
	run, err := config.Client.Runs.Create(ctx, runConfig)
	if err != nil {
		return fmt.Errorf(
			"error creating run for workspace %s: %w", ws.ID, err)
	}

	if run == nil {
		log.Printf("[ERROR] The client returned both a nil run and nil error, this should not happen")
		return fmt.Errorf(
			"the client returned both a nil run and nil error for workspace %s, this should not happen", ws.ID)
	}

	log.Printf("[DEBUG] Run %s created for workspace %s", run.ID, ws.ID)

	isPlanOp := true
	run, err = awaitRun(meta, run.ID, ws.ID, ws.Organization.Name, isPlanOp, planPendingStatuses, planTerminalStatuses)
	if err != nil {
		return err
	}

	if run.Status == tfe.RunErrored || run.Status == tfe.RunStatus(tfe.PolicySoftFailed) {
		if retry && currentRetryAttempts < retryMaxAttempts {
			currentRetryAttempts++
			log.Printf("[INFO] Run errored during plan, retrying run, retry count: %d", currentRetryAttempts)
			return createWorkspaceRun(d, meta, isDestroyRun, currentRetryAttempts)
		}

		return fmt.Errorf("run errored during plan, use the run ID %s to debug error", run.ID)
	}

	// A run is complete when it is successfully planned with no changes to apply
	if run.Status == tfe.RunPlannedAndFinished {
		log.Printf("[INFO] Plan finished, no changes to apply")
		d.SetId(run.ID)
		return nil
	}

	if run.Status == tfe.RunPolicyOverride {
		log.Printf("[INFO] Policy check soft-failed, awaiting manual override for run %q", run.ID)
		run, err = awaitRun(meta, run.ID, ws.ID, ws.Organization.Name, isPlanOp, map[tfe.RunStatus]bool{tfe.RunPolicyOverride: true}, confirmationDoneStatuses)
		if err != nil {
			return err
		}
	}

	manualConfirm := runArgs["manual_confirm"].(bool)
	// if human approval is required, an apply will auto kick off when run is manually approved
	if manualConfirm {
		confirmationPendingStatus := map[tfe.RunStatus]bool{}
		confirmationPendingStatus[run.Status] = true

		log.Printf("[INFO] Plan complete, waiting for manual confirm before proceeding run %q", run.ID)
		run, err = awaitRun(meta, run.ID, ws.ID, ws.Organization.Name, isPlanOp, confirmationPendingStatus, confirmationDoneStatuses)
		if err != nil {
			return err
		}
	} else {
		// if human approval is NOT required, go ahead and kick off an apply
		log.Printf("[INFO] Plan complete, confirming an apply for run %q", run.ID)
		err = config.Client.Runs.Apply(ctx, run.ID, tfe.RunApplyOptions{
			Comment: tfe.String(fmt.Sprintf("Run confirmed by tfe_workspace_run resource via terraform-provider-tfe on %s",
				time.Now().Format(time.UnixDate))),
		})
		if err != nil {
			return fmt.Errorf("run errored while applying run %s: %w", run.ID, err)
		}
	}

	/**
		Checking for waitForRun arg towards the tail of this function ensures that all
		human actions necessary above has already been done.
	**/
	waitForRun := runArgs["wait_for_run"].(bool)
	if !waitForRun {
		d.SetId(run.ID)
		return nil
	}

	isPlanOp = false
	run, err = awaitRun(meta, run.ID, ws.ID, ws.Organization.Name, isPlanOp, applyPendingStatuses, applyDoneStatuses)
	if err != nil {
		return err
	}

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

func awaitRun(meta interface{}, runID string, wsID string, organization string, isPlanOp bool, runPendingStatus map[tfe.RunStatus]bool,
	runTerminalStatus map[tfe.RunStatus]bool) (*tfe.Run, error) {
	config := meta.(ConfiguredClient)

	for i := 0; ; i++ {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("context canceled: %w", ctx.Err())
		case <-time.After(backoff(backoffMin, backoffMax, i)):
			log.Printf("[DEBUG] Polling run %s", runID)
			run, err := config.Client.Runs.Read(ctx, runID)
			if err != nil {
				log.Printf("[ERROR] Could not read run %s: %v", runID, err)
				continue
			}

			_, runHasEnded := runTerminalStatus[run.Status]
			_, runIsInProgress := runPendingStatus[run.Status]

			switch {
			case runHasEnded:
				log.Printf("[INFO] Run %s has reached a terminal state: %s", runID, run.Status)
				return run, nil
			case runIsInProgress:
				log.Printf("[DEBUG] Reading workspace %s", wsID)
				ws, err := config.Client.Workspaces.ReadByID(ctx, wsID)
				if err != nil {
					log.Printf("[ERROR] Unable to read workspace %s: %v", wsID, err)
					continue
				}

				// if the workspace is locked and the current run has not started, assume that workspace was locked for other purposes.
				// display a message to indicate that the workspace is waiting to be manually unlocked before the run can proceed
				if ws.Locked && ws.CurrentRun != nil {
					currentRun, err := config.Client.Runs.Read(ctx, ws.CurrentRun.ID)
					if err != nil {
						log.Printf("[ERROR] Unable to read current run %s: %v", ws.CurrentRun.ID, err)
						continue
					}

					if currentRun.Status == tfe.RunPending {
						log.Printf("[INFO] Waiting for manually locked workspace to be unlocked")
						continue
					}
				}

				// if this run is the current run in it's workspace, display it's position in the organization queue
				if ws.CurrentRun != nil && ws.CurrentRun.ID == runID {
					runPositionInOrg, err := readRunPositionInOrgQueue(meta, runID, organization)
					if err != nil {
						log.Printf("[ERROR] Unable to read run position in organization queue %v", err)
						continue
					}

					orgCapacity, err := config.Client.Organizations.ReadCapacity(ctx, organization)
					if err != nil {
						log.Printf("[ERROR] Unable to read capacity for organization %s: %v", organization, err)
						continue
					}
					if runPositionInOrg > 0 {
						log.Printf("[INFO] Waiting for %d queued run(s) before starting run", runPositionInOrg-orgCapacity.Running)
						continue
					}
				}

				// if this run is not the current run in it's workspace, display it's position in the workspace queue
				runPositionInWorkspace, err := readRunPositionInWorkspaceQueue(meta, runID, wsID, isPlanOp, ws.CurrentRun)
				if err != nil {
					log.Printf("[ERROR] Unable to read run position in workspace queue %v", err)
					continue
				}

				if runPositionInWorkspace > 0 {
					log.Printf(
						"[INFO] Waiting for %d run(s) to finish in workspace %s before being queued...",
						runPositionInWorkspace,
						ws.Name,
					)
					continue
				}

				log.Printf("[INFO] Waiting for run %s, status is %s", runID, run.Status)
			default:
				log.Printf("[INFO] Run %s has entered state: %s", runID, run.Status)
				return run, nil
			}
		}
	}
}

func readRunPositionInOrgQueue(meta interface{}, runID string, organization string) (int, error) {
	config := meta.(ConfiguredClient)
	position := 0
	options := tfe.ReadRunQueueOptions{}

	for {
		runQueue, err := config.Client.Organizations.ReadRunQueue(ctx, organization, options)
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

func readRunPositionInWorkspaceQueue(meta interface{}, runID string, wsID string, isPlanOp bool, currentRun *tfe.Run) (int, error) {
	config := meta.(ConfiguredClient)
	position := 0
	options := tfe.RunListOptions{}
	found := false

	for {
		runList, err := config.Client.Runs.List(ctx, wsID, &options)
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
