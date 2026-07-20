// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	tfev1 "github.com/hashicorp/go-tfe"
	tfe "github.com/hashicorp/go-tfe/v2"
	"github.com/hashicorp/go-tfe/v2/api/models"
	organizationsapi "github.com/hashicorp/go-tfe/v2/api/organizations"
	workspacesapi "github.com/hashicorp/go-tfe/v2/api/workspaces"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	kiota "github.com/microsoft/kiota-abstractions-go"
)

const minTFEVersionWorkspaceRunTaskStages = "v202404-1"

func workspaceRunTaskEnforcementLevels() []string {
	return []string{
		"advisory",
		"mandatory",
	}
}

func workspaceRunTaskStages() []string {
	return []string{
		"pre_plan",
		"post_plan",
		"pre_apply",
		"post_apply",
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
	config         ConfiguredClient
	supportsStages *bool
}

var _ resource.Resource = &resourceWorkspaceRunTask{}
var _ resource.ResourceWithConfigure = &resourceWorkspaceRunTask{}
var _ resource.ResourceWithImportState = &resourceWorkspaceRunTask{}

func NewWorkspaceRunTaskResource() resource.Resource {
	return &resourceWorkspaceRunTask{}
}

func modelFromTFEWorkspaceRunTask(v *tfev1.WorkspaceRunTask) modelTFEWorkspaceRunTaskV1 {
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

func modelFromTFEWorkspaceRunTaskV2(v models.WorkspaceTasksable) modelTFEWorkspaceRunTaskV1 {
	result := modelTFEWorkspaceRunTaskV1{
		ID:               types.StringValue(""),
		WorkspaceID:      types.StringValue(""),
		TaskID:           types.StringValue(""),
		EnforcementLevel: types.StringValue(""),
		Stage:            types.StringValue(""),
		Stages:           types.ListNull(types.StringType),
	}

	if v == nil {
		return result
	}

	if id := v.GetId(); id != nil {
		result.ID = types.StringValue(*id)
	}

	if attrs := v.GetAttributes(); attrs != nil {
		if enforcementLevel := attrs.GetEnforcementLevel(); enforcementLevel != nil {
			result.EnforcementLevel = types.StringValue(enforcementLevel.String())
		}

		if stage := attrs.GetStage(); stage != nil {
			result.Stage = types.StringValue(*stage)
		}

		stages := attrs.GetStages()
		if stages == nil && attrs.GetStage() != nil {
			stages = []string{*attrs.GetStage()}
		}
		if stages, err := types.ListValueFrom(ctx, types.StringType, stages); err == nil {
			result.Stages = stages
		}
	}

	if relationships := v.GetRelationships(); relationships != nil {
		if workspace := relationships.GetWorkspace(); workspace != nil {
			if workspaceData := workspace.GetData(); workspaceData != nil && workspaceData.GetId() != nil {
				result.WorkspaceID = types.StringValue(*workspaceData.GetId())
			}
		}

		if task := relationships.GetTask(); task != nil {
			if taskData := task.GetData(); taskData != nil && taskData.GetId() != nil {
				result.TaskID = types.StringValue(*taskData.GetId())
			}
		}
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
	wstask, err := r.config.ClientV2.API.Workspaces().ByWorkspace_id(workspaceID).Tasks().ById(wstaskID).Get(ctx, nil)
	if err != nil {
		if errors.Is(err, tfe.ErrNotFound) {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError("Error reading Workspace Run Task", "Could not read Workspace Run Task, unexpected error: "+err.Error())
		}
		return
	}

	if wstask == nil || wstask.GetData() == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	result := modelFromTFEWorkspaceRunTaskV2(wstask.GetData())

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
	task, err := r.config.ClientV2.API.Tasks().ById(taskID).Get(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving task", "Could not read Organization Run Task "+taskID+", unexpected error: "+err.Error())
		return
	}
	if task == nil || task.GetData() == nil {
		resp.Diagnostics.AddError("Error retrieving task", "Could not read Organization Run Task "+taskID+", unexpected error: no data returned")
		return
	}

	workspaceID := plan.WorkspaceID.ValueString()
	if _, err := r.config.ClientV2.API.Workspaces().ByWorkspace_id(workspaceID).Get(ctx, nil); err != nil {
		resp.Diagnostics.AddError("Error retrieving workspace", "Could not read Workspace "+workspaceID+", unexpected error: "+err.Error())
		return
	}

	stage, stages := r.extractStageAndStages(plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	options, err := newWorkspaceRunTaskCreateEnvelope(taskID, plan.EnforcementLevel.ValueString(), stage, stages)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create workspace task", err.Error())
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Create task %s in workspace: %s", taskID, workspaceID))
	wstask, err := r.config.ClientV2.API.Workspaces().ByWorkspace_id(workspaceID).Tasks().Post(ctx, options, nil)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create workspace task", err.Error())
		return
	}
	if wstask == nil || wstask.GetData() == nil {
		resp.Diagnostics.AddError("Unable to create workspace task", "No Workspace Run Task data was returned by the API")
		return
	}

	result := modelFromTFEWorkspaceRunTaskV2(wstask.GetData())

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *resourceWorkspaceRunTask) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan modelTFEWorkspaceRunTaskV1

	// Read Terraform planned changes into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	stage, stages := r.extractStageAndStages(plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	wstaskID := plan.ID.ValueString()

	options, err := newWorkspaceRunTaskUpdateEnvelope(wstaskID, plan.EnforcementLevel.ValueString(), stage, stages)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update workspace task", err.Error())
		return
	}

	workspaceID := plan.WorkspaceID.ValueString()

	tflog.Debug(ctx, fmt.Sprintf("Update task %s in workspace %s", wstaskID, workspaceID))
	wstask, err := r.config.ClientV2.API.Workspaces().ByWorkspace_id(workspaceID).Tasks().ById(wstaskID).Patch(ctx, options, nil)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update workspace task", err.Error())
		return
	}
	if wstask == nil || wstask.GetData() == nil {
		resp.Diagnostics.AddError("Unable to update workspace task", "No Workspace Run Task data was returned by the API")
		return
	}

	result := modelFromTFEWorkspaceRunTaskV2(wstask.GetData())

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
	err := r.config.ClientV2.API.Workspaces().ByWorkspace_id(workspaceID).Tasks().ById(wstaskID).Delete(ctx, nil)
	// Ignore 404s for delete
	if err != nil && !errors.Is(err, tfe.ErrNotFound) {
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

	if wstask, err := fetchWorkspaceRunTaskV2(taskName, workspaceName, orgName, r.config.ClientV2); err != nil {
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
		result := modelFromTFEWorkspaceRunTaskV2(wstask)
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

				wstask, err := r.config.ClientV2.API.Workspaces().ByWorkspace_id(oldWorkspaceID).Tasks().ById(oldID).Get(ctx, nil)
				if err != nil {
					resp.Diagnostics.AddError(
						"Error reading workspace run task",
						fmt.Sprintf("Couldn't read workspace run task %s while trying to upgrade state of tfe_workspace_run_task: %s", oldID, err.Error()),
					)
					return
				}
				if wstask == nil || wstask.GetData() == nil {
					resp.Diagnostics.AddError(
						"Error reading workspace run task",
						fmt.Sprintf("Couldn't read workspace run task %s while trying to upgrade state of tfe_workspace_run_task: no data returned", oldID),
					)
					return
				}

				newData := modelFromTFEWorkspaceRunTaskV2(wstask.GetData())
				diags = resp.State.Set(ctx, newData)
				resp.Diagnostics.Append(diags...)
			},
		},
	}
}

func (r *resourceWorkspaceRunTask) addStageSupportDiag(d *diag.Diagnostics, isError bool) {
	summary := "Terraform Enterprise version"
	detail := fmt.Sprintf("The version of Terraform Enterprise does not support the stages attribute on Workspace Run Tasks. Got %s but requires %s+", r.config.RemoteTFEVersion(), minTFEVersionWorkspaceRunTaskStages)
	if isError {
		d.AddError(detail, summary)
	} else {
		d.AddWarning(detail, summary)
	}
}

func (r *resourceWorkspaceRunTask) extractStageAndStages(plan modelTFEWorkspaceRunTaskV1, d *diag.Diagnostics) (*string, []string) {
	// There are some complex interactions here between deprecated values in the TF model, and whether the backend server even supports the newer
	// API call style. This function attempts to extract the Stage and Stages properties and emit useful diagnostics

	// If neither stage or stages is set, then it's all fine, we use the server defaults
	if plan.Stage.IsUnknown() && plan.Stages.IsUnknown() {
		return nil, nil
	}

	meets, err := r.config.MeetsMinRemoteTFEVersion(minTFEVersionWorkspaceRunTaskStages)
	if err != nil {
		d.AddError(
			"Error checking minimum TFE version",
			fmt.Sprintf("Could not determine if Terraform Enterprise version %s meets minimum required version %s: %v",
				r.config.RemoteTFEVersion(), minTFEVersionWorkspaceRunTaskStages, err),
		)
		return nil, nil
	}
	r.supportsStages = &meets

	if meets {
		if plan.Stages.IsUnknown() {
			// The user has supplied Stage but not Stages. They would already have received the deprecation warning so just munge
			// the stage into a slice and we're fine
			stages := []string{plan.Stage.ValueString()}
			return nil, stages
		}

		// Convert the plan values into the slice we need
		var stageStrings []types.String
		if err := plan.Stages.ElementsAs(ctx, &stageStrings, false); err != nil && err.HasError() {
			d.Append(err...)
			return nil, nil
		}
		stages := make([]string, len(stageStrings))
		for idx, s := range stageStrings {
			stages[idx] = s.ValueString()
		}
		return nil, stages
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
		stage := stageStrings[0].ValueString()
		return &stage, nil
	}

	// The user supplied a Stage value to a server that doesn't support stages
	return plan.Stage.ValueStringPointer(), nil
}

func workspaceRunTaskEnforcementLevel(level string) (*models.WorkspaceTasks_attributes_enforcementLevel, error) {
	switch level {
	case "advisory":
		enforcementLevel := models.ADVISORY_WORKSPACETASKS_ATTRIBUTES_ENFORCEMENTLEVEL
		return &enforcementLevel, nil
	case "mandatory":
		enforcementLevel := models.MANDATORY_WORKSPACETASKS_ATTRIBUTES_ENFORCEMENTLEVEL
		return &enforcementLevel, nil
	default:
		return nil, fmt.Errorf("unsupported enforcement level %q", level)
	}
}

func newWorkspaceRunTaskAttributes(enforcementLevel string, stage *string, stages []string) (*models.WorkspaceTasks_attributes, error) {
	parsedEnforcementLevel, err := workspaceRunTaskEnforcementLevel(enforcementLevel)
	if err != nil {
		return nil, err
	}

	attributes := models.NewWorkspaceTasks_attributes()
	attributes.SetEnforcementLevel(parsedEnforcementLevel)
	if stage != nil {
		attributes.SetStage(stage)
	}
	if stages != nil {
		attributes.SetStages(stages)
	}

	return attributes, nil
}

func newWorkspaceRunTaskCreateEnvelope(taskID, enforcementLevel string, stage *string, stages []string) (*models.WorkspaceTasksEnvelope, error) {
	attributes, err := newWorkspaceRunTaskAttributes(enforcementLevel, stage, stages)
	if err != nil {
		return nil, err
	}

	taskRelationshipData := models.NewTasksId_data()
	taskRelationshipData.SetId(&taskID)
	taskType := models.TASKS_TASKSID_DATA_TYPE
	taskRelationshipData.SetTypeEscaped(&taskType)

	taskRelationship := models.NewTasksId()
	taskRelationship.SetData(taskRelationshipData)

	relationships := models.NewWorkspaceTasks_relationships()
	relationships.SetTask(taskRelationship)

	workspaceTask := models.NewWorkspaceTasks()
	workspaceTask.SetAttributes(attributes)
	workspaceTask.SetRelationships(relationships)
	workspaceTaskType := models.WORKSPACETASKS_WORKSPACETASKS_TYPE
	workspaceTask.SetTypeEscaped(&workspaceTaskType)

	envelope := models.NewWorkspaceTasksEnvelope()
	envelope.SetData(workspaceTask)

	return envelope, nil
}

func newWorkspaceRunTaskUpdateEnvelope(id, enforcementLevel string, stage *string, stages []string) (*models.WorkspaceTasksEnvelope, error) {
	attributes, err := newWorkspaceRunTaskAttributes(enforcementLevel, stage, stages)
	if err != nil {
		return nil, err
	}

	workspaceTask := models.NewWorkspaceTasks()
	workspaceTask.SetId(&id)
	workspaceTask.SetAttributes(attributes)
	workspaceTaskType := models.WORKSPACETASKS_WORKSPACETASKS_TYPE
	workspaceTask.SetTypeEscaped(&workspaceTaskType)

	envelope := models.NewWorkspaceTasksEnvelope()
	envelope.SetData(workspaceTask)

	return envelope, nil
}

func fetchWorkspaceRunTaskV2(name, workspace, organization string, client *tfe.Client) (models.WorkspaceTasksable, error) {
	task, err := fetchOrganizationRunTaskV2(name, organization, client)
	if err != nil {
		return nil, fmt.Errorf("Error reading configuration of task %s in organization %s: %w", name, organization, err)
	}

	ws, err := fetchWorkspaceV2(workspace, organization, client)
	if err != nil {
		return nil, fmt.Errorf("Error reading configuration of workspace %s in organization %s: %w", workspace, organization, err)
	}

	if ws.GetId() == nil {
		return nil, fmt.Errorf("Error reading configuration of workspace %s in organization %s: workspace has no ID", workspace, organization)
	}

	workspaceID := *ws.GetId()
	if task.GetId() == nil {
		return nil, fmt.Errorf("Error reading configuration of task %s in organization %s: task has no ID", name, organization)
	}
	taskID := *task.GetId()

	pageNumber := int32(1)
	for {
		query := &workspacesapi.ItemTasksRequestBuilderGetQueryParameters{
			Pagenumber: &pageNumber,
		}
		requestConfig := &kiota.RequestConfiguration[workspacesapi.ItemTasksRequestBuilderGetQueryParameters]{
			QueryParameters: query,
		}

		list, err := client.API.Workspaces().ByWorkspace_id(workspaceID).Tasks().Get(ctx, requestConfig)
		if err != nil {
			return nil, fmt.Errorf("Error retrieving workspace run tasks: %w", err)
		}
		if list == nil {
			break
		}
		if list.GetMeta() == nil {
			break
		}

		for _, wstask := range list.GetData() {
			if wstask == nil || wstask.GetRelationships() == nil || wstask.GetRelationships().GetTask() == nil {
				continue
			}
			taskData := wstask.GetRelationships().GetTask().GetData()
			if taskData != nil && taskData.GetId() != nil && *taskData.GetId() == taskID {
				return wstask, nil
			}
		}

		nextPage, hasNextPage := nextPageFromPagination(list.GetMeta().GetPagination())
		if !hasNextPage {
			break
		}
		pageNumber = nextPage
	}

	return nil, fmt.Errorf("could not find organization run task %s for workspace %s in organization %s", name, workspace, organization)
}

func fetchOrganizationRunTaskV2(name, organization string, client *tfe.Client) (models.Tasksable, error) {
	pageNumber := int32(1)
	for {
		query := &organizationsapi.ItemTasksRequestBuilderGetQueryParameters{
			Pagenumber: &pageNumber,
		}
		requestConfig := &kiota.RequestConfiguration[organizationsapi.ItemTasksRequestBuilderGetQueryParameters]{
			QueryParameters: query,
		}

		list, err := client.API.Organizations().ByOrganization_name(organization).Tasks().Get(ctx, requestConfig)
		if err != nil {
			return nil, fmt.Errorf("Error retrieving organization tasks: %w", err)
		}
		if list == nil {
			break
		}
		if list.GetMeta() == nil {
			break
		}

		for _, task := range list.GetData() {
			if task == nil || task.GetAttributes() == nil || task.GetAttributes().GetName() == nil {
				continue
			}
			if *task.GetAttributes().GetName() == name {
				return task, nil
			}
		}

		nextPage, hasNextPage := nextPageFromPagination(list.GetMeta().GetPagination())
		if !hasNextPage {
			break
		}
		pageNumber = nextPage
	}

	return nil, fmt.Errorf("could not find organization run task for organization %s and name %s", organization, name)
}

func fetchWorkspaceV2(workspace, organization string, client *tfe.Client) (models.Workspacesable, error) {
	ws, err := client.API.Organizations().ByOrganization_name(organization).Workspaces().ByWorkspace_name(workspace).Get(ctx, nil)
	if err != nil {
		return nil, err
	}
	if ws == nil || ws.GetData() == nil {
		return nil, fmt.Errorf("workspace %s has no details", workspace)
	}
	return ws.GetData(), nil
}

func nextPageFromPagination(pagination models.Paginationable) (int32, bool) {
	if pagination == nil {
		return 0, false
	}

	currentPage := pagination.GetCurrentPage()
	totalPages := pagination.GetTotalPages()
	if currentPage == nil || totalPages == nil || *currentPage >= *totalPages {
		return 0, false
	}

	if nextPage := pagination.GetNextPage(); nextPage != nil {
		return *nextPage, true
	}

	return *currentPage + 1, true
}
