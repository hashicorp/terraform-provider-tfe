// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
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
		string(tfe.PostApply),
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

type resourceWorkspaceRunTask struct {
	config       ConfiguredClient
	capabilities capabilitiesResolver
}

var _ resource.Resource = &resourceWorkspaceRunTask{}
var _ resource.ResourceWithConfigure = &resourceWorkspaceRunTask{}
var _ resource.ResourceWithImportState = &resourceWorkspaceRunTask{}

func NewWorkspaceRunTaskResource() resource.Resource {
	return &resourceWorkspaceRunTask{}
}

func modelFromTFEWorkspaceRunTask(v *tfe.WorkspaceRunTask) modelTFEWorkspaceRunTaskV1 {
	result := modelTFEWorkspaceRunTaskV1{
		ID:               types.StringValue(v.ID),
		WorkspaceID:      types.StringValue(v.Workspace.ID),
		TaskID:           types.StringValue(v.RunTask.ID),
		EnforcementLevel: types.StringValue(string(v.EnforcementLevel)),
		Stage:            types.StringValue(string(v.Stage)),
		Stages:           types.ListNull(types.StringType),
	}

	if stages, err := types.ListValueFrom(ctx, types.StringType, v.Stages); err == nil {
		result.Stages = stages
	}

	return result
}

func (r *resourceWorkspaceRunTask) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workspace_run_task"
}

// Configure implements resource.ResourceWithConfigure
func (r *resourceWorkspaceRunTask) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(ConfiguredClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected resource Configure type",
			fmt.Sprintf("Expected tfe.ConfiguredClient, got %T. This is a bug in the tfe provider, so please report it on GitHub.", req.ProviderData),
		)
	}
	r.config = client
	r.capabilities = newDefaultCapabilityResolver(client.Client)
}

func (r *resourceWorkspaceRunTask) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resourceWorkspaceRunTaskSchemaV1
}

func (r *resourceWorkspaceRunTask) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state modelTFEWorkspaceRunTaskV1

	// Read Terraform current state into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wstaskID := state.ID.ValueString()
	workspaceID := state.WorkspaceID.ValueString()

	tflog.Debug(ctx, "Reading workspace run task")
	wstask, err := r.config.Client.WorkspaceRunTasks.Read(ctx, workspaceID, wstaskID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading Workspace Run Task", "Could not read Workspace Run Task, unexpected error: "+err.Error())
		return
	}

	result := modelFromTFEWorkspaceRunTask(wstask)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *resourceWorkspaceRunTask) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan modelTFEWorkspaceRunTaskV1

	// Read Terraform planned changes into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	taskID := plan.TaskID.ValueString()
	task, err := r.config.Client.RunTasks.Read(ctx, taskID)
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving task", "Could not read Organization Run Task "+taskID+", unexpected error: "+err.Error())
		return
	}

	workspaceID := plan.WorkspaceID.ValueString()
	if _, err := r.config.Client.Workspaces.ReadByID(ctx, workspaceID); err != nil {
		resp.Diagnostics.AddError("Error retrieving workspace", "Could not read Workspace "+workspaceID+", unexpected error: "+err.Error())
		return
	}

	level := tfe.TaskEnforcementLevel(plan.EnforcementLevel.ValueString())

	options := tfe.WorkspaceRunTaskCreateOptions{
		RunTask:          task,
		EnforcementLevel: level,
	}

	stage, stages := r.extractStageAndStages(plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if stage != nil {
		// Needed for older TFE instances
		options.Stage = stage //nolint:staticcheck
	}
	if stages != nil {
		options.Stages = stages
	}

	tflog.Debug(ctx, fmt.Sprintf("Create task %s in workspace: %s", taskID, workspaceID))
	wstask, err := r.config.Client.WorkspaceRunTasks.Create(ctx, workspaceID, options)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create workspace task", err.Error())
		return
	}

	result := modelFromTFEWorkspaceRunTask(wstask)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *resourceWorkspaceRunTask) stringPointerToStagePointer(val *string) *tfe.Stage {
	if val == nil {
		return nil
	}
	newVal := tfe.Stage(*val)
	return &newVal
}

func (r *resourceWorkspaceRunTask) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan modelTFEWorkspaceRunTaskV1

	// Read Terraform planned changes into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	level := tfe.TaskEnforcementLevel(plan.EnforcementLevel.ValueString())

	options := tfe.WorkspaceRunTaskUpdateOptions{
		EnforcementLevel: level,
	}

	stage, stages := r.extractStageAndStages(plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if stage != nil {
		// Needed for older TFE instances
		options.Stage = stage //nolint:staticcheck
	}
	if stages != nil {
		options.Stages = stages
	}

	wstaskID := plan.ID.ValueString()
	workspaceID := plan.WorkspaceID.ValueString()

	tflog.Debug(ctx, fmt.Sprintf("Update task %s in workspace %s", wstaskID, workspaceID))
	wstask, err := r.config.Client.WorkspaceRunTasks.Update(ctx, workspaceID, wstaskID, options)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update workspace task", err.Error())
		return
	}

	result := modelFromTFEWorkspaceRunTask(wstask)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *resourceWorkspaceRunTask) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state modelTFEWorkspaceRunTaskV1
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	wstaskID := state.ID.ValueString()
	workspaceID := state.WorkspaceID.ValueString()

	tflog.Debug(ctx, fmt.Sprintf("Delete task %s in workspace %s", wstaskID, workspaceID))
	err := r.config.Client.WorkspaceRunTasks.Delete(ctx, workspaceID, wstaskID)
	// Ignore 404s for delete
	if err != nil && !errors.Is(err, tfe.ErrResourceNotFound) {
		resp.Diagnostics.AddError(
			"Error deleting workspace run task",
			fmt.Sprintf("Couldn't delete task %s in workspace %s: %s", wstaskID, workspaceID, err.Error()),
		)
	}
	// Resource is implicitly deleted from resp.State if diagnostics have no errors.
}

func (r *resourceWorkspaceRunTask) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	s := strings.SplitN(req.ID, "/", 3)
	if len(s) != 3 {
		resp.Diagnostics.AddError(
			"Error importing workspace run task",
			fmt.Sprintf("Invalid task input format: %s (expected <ORGANIZATION>/<WORKSPACE NAME>/<TASK NAME>)", req.ID),
		)
		return
	}

	taskName := s[2]
	workspaceName := s[1]
	orgName := s[0]

	if wstask, err := fetchWorkspaceRunTask(taskName, workspaceName, orgName, r.config.Client); err != nil {
		resp.Diagnostics.AddError(
			"Error importing workspace run task",
			err.Error(),
		)
	} else if wstask == nil {
		resp.Diagnostics.AddError(
			"Error importing workspace run task",
			"Workspace task does not exist or has no details",
		)
	} else {
		result := modelFromTFEWorkspaceRunTask(wstask)
		resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
	}
}

func (r *resourceWorkspaceRunTask) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema: &resourceWorkspaceRunTaskSchemaV0,
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				var oldData modelTFEWorkspaceRunTaskV0
				diags := req.State.Get(ctx, &oldData)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}

				oldWorkspaceID := oldData.WorkspaceID.ValueString()
				oldID := oldData.ID.ValueString()

				wstask, err := r.config.Client.WorkspaceRunTasks.Read(ctx, oldWorkspaceID, oldID)
				if err != nil || wstask == nil {
					resp.Diagnostics.AddError(
						"Error reading workspace run task",
						fmt.Sprintf("Couldn't read workspace run task %s while trying to upgrade state of tfe_workspace_run_task: %s", oldID, err.Error()),
					)
					return
				}

				newData := modelFromTFEWorkspaceRunTask(wstask)
				diags = resp.State.Set(ctx, newData)
				resp.Diagnostics.Append(diags...)
			},
		},
	}
}

func (r *resourceWorkspaceRunTask) supportsStagesProperty() bool {
	// The Stages property is available in HCP Terraform and Terraform Enterprise v202404-1 onwards.
	//
	// The version comparison here can use plain string comparisons due to the nature of the naming scheme. If
	// TFE every changes its scheme, the comparison will be problematic.
	return r.capabilities.IsCloud() || r.capabilities.RemoteTFEVersion() > "v202404"
}

func (r *resourceWorkspaceRunTask) addStageSupportDiag(d *diag.Diagnostics, isError bool) {
	summary := "Terraform Enterprise version"
	detail := fmt.Sprintf("The version of Terraform Enterprise does not support the stages attribute on Workspace Run Tasks. Got %s but requires v202404-1+", r.config.Client.RemoteTFEVersion())
	if isError {
		d.AddError(detail, summary)
	} else {
		d.AddWarning(detail, summary)
	}
}

func (r *resourceWorkspaceRunTask) extractStageAndStages(plan modelTFEWorkspaceRunTaskV1, d *diag.Diagnostics) (*tfe.Stage, *[]tfe.Stage) {
	// There are some complex interactions here between deprecated values in the TF model, and whether the backend server even supports the newer
	// API call style. This function attempts to extract the Stage and Stages properties and emit useful diagnostics

	// If neither stage or stages is set, then it's all fine, we use the server defaults
	if plan.Stage.IsUnknown() && plan.Stages.IsUnknown() {
		return nil, nil
	}

	if r.supportsStagesProperty() {
		if plan.Stages.IsUnknown() {
			// The user has supplied Stage but not Stages. They would already have received the deprecation warning so just munge
			// the stage into a slice and we're fine
			stages := []tfe.Stage{tfe.Stage(plan.Stage.ValueString())}
			return nil, &stages
		}

		// Convert the plan values into the slice we need
		var stageStrings []types.String
		if err := plan.Stages.ElementsAs(ctx, &stageStrings, false); err != nil && err.HasError() {
			d.Append(err...)
			return nil, nil
		}
		stages := make([]tfe.Stage, len(stageStrings))
		for idx, s := range stageStrings {
			stages[idx] = tfe.Stage(s.ValueString())
		}
		return nil, &stages
	}

	// The backend server doesn't support Stages
	if !plan.Stages.IsUnknown() {
		// The user has supplied a stages array. We need to figure out if we can munge this into a stage attribute
		stagesCount := len(plan.Stages.Elements())

		if stagesCount > 1 {
			// The user has supplied more than one stage so we can't munge this
			r.addStageSupportDiag(d, true)
			return nil, nil
		}

		// Send the warning
		r.addStageSupportDiag(d, false)

		if stagesCount == 0 {
			// Somehow we've got no stages listed. Use default server values
			return nil, nil
		}

		// ... Otherwise there's a single Stages value which we can munge into Stage.
		var stageStrings []types.String
		if err := plan.Stages.ElementsAs(ctx, &stageStrings, false); err != nil && err.HasError() {
			d.Append(err...)
			return nil, nil
		}
		stage := tfe.Stage(stageStrings[0].ValueString())
		return &stage, nil
	}

	// The user supplied a Stage value to a server that doesn't support stages
	return r.stringPointerToStagePointer(plan.Stage.ValueStringPointer()), nil
}
