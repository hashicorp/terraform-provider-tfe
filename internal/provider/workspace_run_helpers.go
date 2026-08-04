// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"strings"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	tfev2api "github.com/hashicorp/go-tfe/v2/api"
	"github.com/hashicorp/go-tfe/v2/api/models"
	v2runs "github.com/hashicorp/go-tfe/v2/api/runs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// runState captures the subset of run fields needed by the workspace-run helpers.
// It is populated from a go-tfe v2 RunsEnvelopeable response.
type runState struct {
	ID               string
	Status           string
	HasChanges       bool
	AllowEmptyApply  bool
	IsConfirmable    bool
	PolicyCheckCount int
	HasCostEstimate  bool
	WorkspaceID      string
}

// wsState captures the subset of workspace fields needed by the helpers.
type wsState struct {
	ID           string
	Name         string
	OrgName      string
	Locked       bool
	CurrentRunID string
}

// runStateFromEnvelope extracts a runState from a v2 RunsEnvelopeable.
func runStateFromEnvelope(resp models.RunsEnvelopeable) *runState {
	if resp == nil || resp.GetData() == nil {
		return nil
	}
	data := resp.GetData()
	rs := &runState{
		ID: valueOrZero(data.GetId()),
	}
	if attrs := data.GetAttributes(); attrs != nil {
		if s := attrs.GetStatus(); s != nil {
			rs.Status = s.String()
		}
		rs.HasChanges = valueOrZero(attrs.GetHasChanges())
		rs.AllowEmptyApply = valueOrZero(attrs.GetAllowEmptyApply())
		if actions := attrs.GetActions(); actions != nil {
			rs.IsConfirmable = valueOrZero(actions.GetIsConfirmable())
		}
	}
	if rels := data.GetRelationships(); rels != nil {
		if pc := rels.GetPolicyChecks(); pc != nil {
			rs.PolicyCheckCount = len(pc.GetData())
		}
		if ce := rels.GetCostEstimate(); ce != nil && ce.GetData() != nil {
			rs.HasCostEstimate = true
		}
		if ws := rels.GetWorkspace(); ws != nil && ws.GetData() != nil {
			rs.WorkspaceID = valueOrZero(ws.GetData().GetId())
		}
	}
	return rs
}

// wsStateFromEnvelope extracts a wsState from a v2 workspace response.
func wsStateFromEnvelope(wsID string, wsEnv models.WorkspacesEnvelopeable) *wsState {
	if wsEnv == nil || wsEnv.GetData() == nil {
		return nil
	}
	data := wsEnv.GetData()
	ws := &wsState{
		ID: valueOrZero(data.GetId()),
	}
	if attrs := data.GetAttributes(); attrs != nil {
		ws.Name = valueOrZero(attrs.GetName())
		ws.Locked = valueOrZero(attrs.GetLocked())
	}
	if rels := data.GetRelationships(); rels != nil {
		if org := rels.GetOrganization(); org != nil && org.GetData() != nil {
			ws.OrgName = valueOrZero(org.GetData().GetId())
		}
		if cr := rels.GetCurrentRun(); cr != nil && cr.GetData() != nil {
			ws.CurrentRunID = valueOrZero(cr.GetData().GetId())
		}
	}
	if ws.ID == "" {
		ws.ID = wsID
	}
	return ws
}

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
	api := config.ClientV2.API

	workspaceID := d.Get("workspace_id").(string)
	log.Printf("[DEBUG] Read workspace by ID %s", workspaceID)

	wsResp, err := api.Workspaces().ByWorkspace_id(workspaceID).Get(ctx, nil)
	if err != nil {
		return fmt.Errorf("error reading workspace %s: %w", workspaceID, err)
	}
	ws := wsStateFromEnvelope(workspaceID, wsResp)
	if ws == nil {
		return fmt.Errorf("workspace %s returned empty response", workspaceID)
	}

	waitForRun := runArgs["wait_for_run"].(bool)
	manualConfirm := runArgs["manual_confirm"].(bool)
	msg, _ := runArgs["message"].(string)

	run, err := createRunV2(api, ws.ID, waitForRun, manualConfirm, isDestroyRun, msg)
	if err != nil {
		if isDestroyRun && isConfigVersionMissingErr(err) {
			log.Printf("[WARN] Configuration version is missing for workspace %s; treating destroy as a no-op because there is nothing to destroy", ws.ID)
			d.SetId(fmt.Sprintf("%d", rand.New(rand.NewSource(time.Now().UnixNano())).Int()))
			return nil
		}
		return err
	}

	if !waitForRun {
		d.SetId(run.ID)
		return nil
	}

	isPlanOp := true
	hasPostPlanTaskStage, err := readPostPlanTaskStageInRunV2(api, run.ID)
	if err != nil {
		return err
	}

	planPendingStatuses, planTerminalStatuses := planStatusesV2(run, hasPostPlanTaskStage)
	run, err = awaitRunV2(api, config.Client, run.ID, ws.OrgName, isPlanOp, planPendingStatuses, isPlanCompleteV2(planTerminalStatuses))
	if err != nil {
		return err
	}

	if run.Status == runStatusErrored || run.Status == runStatusPolicySoftFailed {
		if retry && currentRetryAttempts < retryMaxAttempts {
			currentRetryAttempts++
			log.Printf("[INFO] Run errored during plan, retrying run, retry count: %d", currentRetryAttempts)
			return createWorkspaceRun(d, meta, isDestroyRun, currentRetryAttempts)
		}

		return fmt.Errorf("run errored during plan, use the run ID %s to debug error", run.ID)
	}

	if run.Status == runStatusPolicyOverride {
		log.Printf("[INFO] Policy check soft-failed, awaiting manual override for run %q", run.ID)
		run, err = awaitRunV2(api, config.Client, run.ID, ws.OrgName, isPlanOp, policyOverridePendingStatuses, isManuallyOverridenV2)
		if err != nil {
			return err
		}
	}

	if !run.HasChanges && !run.AllowEmptyApply {
		run, err = awaitRunV2(api, config.Client, run.ID, ws.OrgName, isPlanOp, confirmationPendingStatuses, isPlannedAndFinishedV2)
		if err != nil {
			return err
		}
	}

	if run.Status == runStatusPlannedAndFinished {
		log.Printf("[INFO] Plan finished, no changes to apply")
		d.SetId(run.ID)
		return nil
	}

	run, err = awaitRunV2(api, config.Client, run.ID, ws.OrgName, isPlanOp, confirmationPendingStatuses, isConfirmableV2)
	if err != nil {
		return err
	}

	err = confirmRunV2(api, manualConfirm, isPlanOp, run, ws)
	if err != nil {
		return err
	}

	isPlanOp = false
	run, err = awaitRunV2(api, config.Client, run.ID, ws.OrgName, isPlanOp, applyPendingStatuses, isCompletedV2)
	if err != nil {
		return err
	}

	return completeOrRetryRunV2(meta, run, d, retry, currentRetryAttempts, retryMaxAttempts, isDestroyRun)
}

func getRunArgs(d *schema.ResourceData, isDestroyRun bool) map[string]interface{} {
	var runArgs map[string]interface{}

	if isDestroyRun {
		destroyArgs, ok := d.GetOk("destroy")
		if !ok {
			return nil
		}
		runArgs = destroyArgs.([]interface{})[0].(map[string]interface{})
	} else {
		createArgs, ok := d.GetOk("apply")
		if !ok {
			d.SetId(fmt.Sprintf("%d", rand.New(rand.NewSource(time.Now().UnixNano())).Int()))
			return nil
		}
		runArgs = createArgs.([]interface{})[0].(map[string]interface{})
	}

	return runArgs
}

// isConfigVersionMissingErr reports whether err was caused by the workspace
// lacking a configuration version when creating a run.
func isConfigVersionMissingErr(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(strings.ToLower(err.Error()), "configuration version is missing")
}

func createRunV2(api *tfev2api.ApiClient, wsID string, waitForRun bool, manualConfirm bool, isDestroyRun bool, message string) (*runState, error) {
	autoApply := false
	if !waitForRun {
		autoApply = !manualConfirm
	}

	runAttrs := models.NewRuns_attributes()
	runAttrs.SetIsDestroy(ptr(isDestroyRun))
	runAttrs.SetAutoApply(ptr(autoApply))
	if message != "" {
		runAttrs.SetMessage(ptr(message))
	}

	wsIdentType := models.WORKSPACES_WORKSPACESIDENTIFIER_TYPE
	wsData := models.NewWorkspacesHasOne_data()
	wsData.SetId(ptr(wsID))
	wsData.SetTypeEscaped(&wsIdentType)

	wsRel := models.NewWorkspacesHasOne()
	wsRel.SetData(wsData)

	runRels := models.NewRuns_relationships()
	runRels.SetWorkspace(wsRel)

	runType := models.RUNS_RUNS_TYPE
	runData := models.NewRuns()
	runData.SetTypeEscaped(&runType)
	runData.SetAttributes(runAttrs)
	runData.SetRelationships(runRels)

	envelope := models.NewRunsEnvelope()
	envelope.SetData(runData)

	log.Printf("[DEBUG] Create run for workspace: %s", wsID)
	resp, err := api.Runs().Post(ctx, envelope, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating run for workspace %s: %w", wsID, err)
	}
	if resp == nil || resp.GetData() == nil {
		return nil, fmt.Errorf("the client returned both a nil run and nil error for workspace %s", wsID)
	}

	run := runStateFromEnvelope(resp)
	if run == nil {
		return nil, fmt.Errorf("failed to parse run response for workspace %s", wsID)
	}

	log.Printf("[DEBUG] Run %s created for workspace %s", run.ID, wsID)
	return run, nil
}

func confirmRunV2(api *tfev2api.ApiClient, manualConfirm bool, isPlanOp bool, run *runState, ws *wsState) error {
	if manualConfirm {
		confirmationPendingStatus := map[string]bool{run.Status: true}
		log.Printf("[INFO] Plan complete, waiting for manual confirm before proceeding run %q", run.ID)
		_, err := awaitRunV2(api, nil, run.ID, ws.OrgName, isPlanOp, confirmationPendingStatus, isConfirmedV2)
		return err
	}
	return applyRunV2(api, run)
}

func applyRunV2(api *tfev2api.ApiClient, run *runState) error {
	log.Printf("[INFO] Plan complete, confirming an apply for run %q", run.ID)

	comment := fmt.Sprintf("Run confirmed by tfe_workspace_run resource via terraform-provider-tfe on %s",
		time.Now().Format(time.UnixDate))

	actionComment := models.NewActionComment()
	actionComment.SetComment(ptr(comment))

	_, err := api.Runs().ById(run.ID).Actions().Apply().Post(ctx, actionComment, nil)
	if err != nil {
		// Try to read the run to get current status for a better error message.
		refreshed, fetchErr := api.Runs().ById(run.ID).Get(ctx, nil)
		currentStatus := "unknown"
		if fetchErr == nil && refreshed.GetData() != nil && refreshed.GetData().GetAttributes() != nil {
			if s := refreshed.GetData().GetAttributes().GetStatus(); s != nil {
				currentStatus = s.String()
			}
		}
		return fmt.Errorf("run errored while applying run %s (waited til status %s, currently status %s): %w", run.ID, run.Status, currentStatus, err)
	}

	return nil
}

func completeOrRetryRunV2(meta interface{}, run *runState, d *schema.ResourceData, retry bool, currentRetryAttempts int, retryMaxAttempts int, isDestroyRun bool) error {
	switch run.Status {
	case runStatusApplied:
		log.Printf("[INFO] Apply complete for run %q", run.ID)
		d.SetId(run.ID)
		return nil
	case runStatusErrored:
		if retry && currentRetryAttempts < retryMaxAttempts {
			currentRetryAttempts++
			log.Printf("[INFO] Run errored during apply, retrying run, retry count: %d", currentRetryAttempts)
			return createWorkspaceRun(d, meta, isDestroyRun, currentRetryAttempts)
		}
		return fmt.Errorf("run errored during apply, use the run ID %s to debug error", run.ID)
	default:
		return fmt.Errorf("run %s entered unexpected state %s, expected %s state", run.ID, run.Status, runStatusApplied)
	}
}

// awaitRunV2 polls a run until isDone returns true or a terminal/unexpected state is reached.
// v1Client is optional; when non-nil, it is used for logging capacity/queue info (documented v1 fallback).
func awaitRunV2(api *tfev2api.ApiClient, v1Client *tfe.Client, runID string, organization string, isPlanOp bool, runPendingStatus map[string]bool, isDone func(*runState) bool) (*runState, error) {
	for i := 0; ; i++ {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("context canceled: %w", ctx.Err())
		case <-time.After(backoff(backoffMin, backoffMax, i)):
			log.Printf("[DEBUG] Polling run %s", runID)
			runResp, err := api.Runs().ById(runID).Get(ctx, nil)
			if err != nil {
				log.Printf("[ERROR] Could not read run %s: %v", runID, err)
				continue
			}

			run := runStateFromEnvelope(runResp)
			if run == nil {
				log.Printf("[ERROR] Could not parse run response for %s", runID)
				continue
			}

			result, err := hasFinalStatusV2(api, v1Client, run, organization, isPlanOp, runPendingStatus, isDone)
			if result == nil && err == nil {
				continue
			}
			return result, err
		}
	}
}

func hasFinalStatusV2(api *tfev2api.ApiClient, v1Client *tfe.Client, run *runState, organization string, isPlanOp bool, runPendingStatus map[string]bool, isDone func(*runState) bool) (*runState, error) {
	_, runIsInProgress := runPendingStatus[run.Status]

	switch {
	case isDone(run):
		log.Printf("[INFO] Run %s has reached a terminal state: %s", run.ID, run.Status)
		return run, nil
	case runIsInProgress:
		logRunProgressV2(api, v1Client, organization, isPlanOp, run)
		return nil, nil
	case run.Status == runStatusCanceled:
		log.Printf("[INFO] Run %s has been canceled, status is %s", run.ID, run.Status)
		return nil, fmt.Errorf("run %s has been canceled, status is %s", run.ID, run.Status)
	default:
		log.Printf("[INFO] Run %s has entered unexpected state: %s", run.ID, run.Status)
		return nil, fmt.Errorf("run %s has entered unexpected state: %s", run.ID, run.Status)
	}
}

func logRunProgressV2(api *tfev2api.ApiClient, v1Client *tfe.Client, organization string, isPlanOp bool, run *runState) {
	log.Printf("[DEBUG] Reading workspace %s", run.WorkspaceID)
	wsResp, err := api.Workspaces().ByWorkspace_id(run.WorkspaceID).Get(ctx, nil)
	if err != nil {
		log.Printf("[ERROR] Unable to read workspace %s: %v", run.WorkspaceID, err)
		return
	}
	ws := wsStateFromEnvelope(run.WorkspaceID, wsResp)
	if ws == nil {
		return
	}

	if ws.Locked && ws.CurrentRunID != "" {
		// Check if the current run is pending (meaning workspace is manually locked).
		crResp, err := api.Runs().ById(ws.CurrentRunID).Get(ctx, nil)
		if err != nil {
			log.Printf("[ERROR] Unable to read current run %s: %v", ws.CurrentRunID, err)
			return
		}
		cr := runStateFromEnvelope(crResp)
		if cr != nil && cr.Status == runStatusPending {
			log.Printf("[INFO] Waiting for manually locked workspace to be unlocked")
			return
		}
	}

	if ws.CurrentRunID == run.ID {
		// Organizations.ReadCapacity and ReadRunQueue remain on go-tfe v1:
		// These endpoints are not available in go-tfe/v2 (no generated builders).
		// They are used only for logging and do not affect run outcomes.
		// Removal condition: when these endpoints are added to the Atlas OpenAPI
		// spec and regenerated in go-tfe/v2, this fallback can be migrated.
		if v1Client != nil {
			runPositionInOrg, err := readRunPositionInOrgQueue(v1Client, run.ID, organization)
			if err != nil {
				log.Printf("[ERROR] Unable to read run position in organization queue %v", err)
				return
			}

			orgCapacity, err := v1Client.Organizations.ReadCapacity(ctx, organization)
			if err != nil {
				log.Printf("[ERROR] Unable to read capacity for organization %s: %v", organization, err)
				return
			}
			if runPositionInOrg > 0 {
				log.Printf("[INFO] Waiting for %d queued run(s) before starting run", runPositionInOrg-orgCapacity.Running)
				return
			}
		}
	}

	runPositionInWorkspace, err := readRunPositionInWorkspaceQueueV2(api, run.ID, ws.ID, isPlanOp, ws.CurrentRunID)
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

func readRunPositionInOrgQueue(tfeClient *tfe.Client, runID string, organization string) (int, error) {
	position := 0
	options := tfe.ReadRunQueueOptions{}

	for {
		runQueue, err := tfeClient.Organizations.ReadRunQueue(ctx, organization, options)
		if err != nil {
			return position, fmt.Errorf("unable to read run queue for organization %s: %w", organization, err)
		}
		for _, item := range runQueue.Items {
			if runID == item.ID {
				position = item.PositionInQueue
				return position, nil
			}
		}

		if runQueue.CurrentPage >= runQueue.TotalPages {
			break
		}
		options.PageNumber = runQueue.NextPage
	}

	return position, nil
}

func readRunPositionInWorkspaceQueueV2(api *tfev2api.ApiClient, runID string, wsID string, isPlanOp bool, currentRunID string) (int, error) {
	position := 0
	found := false

	pageSize := int32(100)
	queryParams := &v2runs.RunsRequestBuilderGetQueryParameters{
		Workspace_id: ptr(wsID),
		Pagesize:     &pageSize,
	}

	for {
		runList, err := api.Runs().Get(ctx, withQueryParams(queryParams))
		if err != nil {
			return position, fmt.Errorf("unable to read run list for workspace %s: %w", wsID, err)
		}

		for _, item := range runList.GetData() {
			itemID := valueOrZero(item.GetId())
			if !found {
				if runID == itemID {
					found = true
				}
				continue
			}

			// ignore runs with final states while computing queue count
			itemStatus := ""
			if attrs := item.GetAttributes(); attrs != nil && attrs.GetStatus() != nil {
				itemStatus = attrs.GetStatus().String()
			}
			switch itemStatus {
			case runStatusApplied, runStatusCanceled, runStatusDiscarded, runStatusErrored, runStatusPlannedAndFinished:
				continue
			case runStatusPlanned:
				if isPlanOp {
					continue
				}
			}

			position++

			if currentRunID != "" && currentRunID == itemID {
				return position, nil
			}
		}

		// Exit the loop when we've seen all pages.
		nextPage := nextPageFromMeta(runList.GetMeta())
		if nextPage == nil {
			break
		}
		queryParams.Pagenumber = nextPage
	}

	return position, nil
}

func readPostPlanTaskStageInRunV2(api *tfev2api.ApiClient, runID string) (bool, error) {
	hasPostPlanTaskStage := false

	taskStagesBuilder := api.Runs().ById(runID).TaskStages()
	for {
		taskStages, err := taskStagesBuilder.Get(ctx, nil)
		if err != nil {
			return hasPostPlanTaskStage, fmt.Errorf("[ERROR] Could not read task stages for run %s: %w", runID, err)
		}
		for _, item := range taskStages.GetData() {
			if attrs := item.GetAttributes(); attrs != nil {
				stage := valueOrZero(attrs.GetStage())
				if stage == "post_plan" {
					hasPostPlanTaskStage = true
					return hasPostPlanTaskStage, nil
				}
			}
		}

		// Follow link-based pagination for task stages.
		links := taskStages.GetLinks()
		if links == nil || links.GetNext() == nil || *links.GetNext() == "" {
			break
		}
		taskStagesBuilder = taskStagesBuilder.WithUrl(*links.GetNext())
	}

	return hasPostPlanTaskStage, nil
}

func planStatusesV2(run *runState, hasPostPlanTaskStage bool) (map[string]bool, map[string]bool) {
	hasPolicyCheck := run.PolicyCheckCount > 0
	hasCostEstimate := run.HasCostEstimate

	var planTerminalStatuses = map[string]bool{
		runStatusErrored:            true,
		runStatusPlannedAndFinished: true,
		runStatusPolicySoftFailed:   true,
		runStatusPolicyOverride:     true,
	}

	var planPendingStatuses = map[string]bool{
		runStatusPending:           true,
		runStatusPlanQueued:        true,
		runStatusPlanning:          true,
		runStatusCostEstimating:    true,
		runStatusPolicyChecking:    true,
		runStatusQueuing:           true,
		runStatusFetching:          true,
		runStatusPostPlanRunning:   true,
		runStatusPostPlanCompleted: true,
		runStatusPrePlanRunning:    true,
		runStatusPrePlanCompleted:  true,
	}

	if hasPolicyCheck {
		planTerminalStatuses[runStatusPolicyChecked] = true
		planPendingStatuses[runStatusCostEstimated] = true
		planPendingStatuses[runStatusPlanned] = true
	} else if hasCostEstimate {
		planTerminalStatuses[runStatusCostEstimated] = true
		planPendingStatuses[runStatusPlanned] = true
	} else if hasPostPlanTaskStage {
		planTerminalStatuses[runStatusPostPlanCompleted] = true
		planPendingStatuses[runStatusPlanned] = true
	} else {
		planTerminalStatuses[runStatusPlanned] = true
	}

	return planPendingStatuses, planTerminalStatuses
}

func isPlanCompleteV2(planTerminalStatuses map[string]bool) func(run *runState) bool {
	return func(run *runState) bool {
		_, found := planTerminalStatuses[run.Status]
		return found
	}
}

func isManuallyOverridenV2(run *runState) bool {
	_, found := policyOverriddenStatuses[run.Status]
	return found
}

func isPlannedAndFinishedV2(run *runState) bool {
	return runStatusPlannedAndFinished == run.Status
}

func isConfirmableV2(run *runState) bool {
	return run.IsConfirmable
}

func isConfirmedV2(run *runState) bool {
	_, found := confirmationDoneStatuses[run.Status]
	return found
}

func isCompletedV2(run *runState) bool {
	_, found := applyDoneStatuses[run.Status]
	return found
}

// perform exponential backoff based on the iteration and
// limited by the provided min and max durations in milliseconds.
func backoff(minVal float64, maxVal float64, iter int) time.Duration {
	backoffVal := math.Pow(2, float64(iter)/5) * minVal
	if backoffVal > maxVal {
		backoffVal = maxVal
	}
	return time.Duration(backoffVal) * time.Millisecond
}

// readRunPositionInWorkspaceQueue is the v1-based queue-position helper retained
// for unit tests (workspace_run_helpers_test.go). Production code uses
// readRunPositionInWorkspaceQueueV2 instead.
func readRunPositionInWorkspaceQueue(tfeClient *tfe.Client, runID string, wsID string, isPlanOp bool, currentRun *tfe.Run) (int, error) {
	position := 0
	options := tfe.RunListOptions{}
	found := false

	currentRunID := ""
	if currentRun != nil {
		currentRunID = currentRun.ID
	}

	for {
		runList, err := tfeClient.Runs.List(ctx, wsID, &options)
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

			switch item.Status {
			case tfe.RunApplied, tfe.RunCanceled, tfe.RunDiscarded, tfe.RunErrored, tfe.RunPlannedAndFinished:
				continue
			case tfe.RunPlanned:
				if isPlanOp {
					continue
				}
			}

			position++

			if currentRunID != "" && currentRunID == item.ID {
				return position, nil
			}
		}

		if runList.CurrentPage >= runList.TotalPages {
			break
		}
		options.PageNumber = runList.NextPage
	}

	return position, nil
}
