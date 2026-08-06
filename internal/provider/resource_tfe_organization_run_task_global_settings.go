// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	tfe "github.com/hashicorp/go-tfe/v2"
	"github.com/hashicorp/go-tfe/v2/api/models"
	organizationsapi "github.com/hashicorp/go-tfe/v2/api/organizations"
	forownerapi "github.com/hashicorp/go-tfe/v2/api/organizations/item/taskconfigs/forowner"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	kiota "github.com/microsoft/kiota-abstractions-go"
)

var _ resource.Resource = &resourceOrganizationRunTaskGlobalSettings{}
var _ resource.ResourceWithConfigure = &resourceOrganizationRunTaskGlobalSettings{}
var _ resource.ResourceWithImportState = &resourceOrganizationRunTaskGlobalSettings{}

type modelDataTFEOrganizationRunTaskGlobalSettings struct {
	Enabled          types.Bool   `tfsdk:"enabled"`
	EnforcementLevel types.String `tfsdk:"enforcement_level"`
	ID               types.String `tfsdk:"id"`
	Stages           types.List   `tfsdk:"stages"`
	TaskID           types.String `tfsdk:"task_id"`
}

// newOrganizationRunTaskGlobalSettingsEnvelope builds the request body for updating an organization
// run task's global settings. stages and enforcementLevel are left unset (nil) when the caller only
// intends to change enabled, e.g. on Delete.
func newOrganizationRunTaskGlobalSettingsEnvelope(taskID string, enabled *bool, stages []string, enforcementLevel *string) *models.TasksEnvelope {
	global := models.NewTasks_attributes_globalConfiguration()
	global.SetEnabled(enabled)
	global.SetStages(stages)
	global.SetEnforcementLevel(enforcementLevel)

	attributes := models.NewTasks_attributes()
	attributes.SetGlobalConfiguration(global)

	data := models.NewTasks()
	data.SetId(&taskID)
	data.SetAttributes(attributes)
	taskType := models.TASKS_TASKS_TYPE
	data.SetTypeEscaped(&taskType)

	envelope := models.NewTasksEnvelope()
	envelope.SetData(data)
	return envelope
}

func newOrganizationRunTaskGlobalTaskConfigEnvelope(taskID, organization string, enabled *bool, stages []string, enforcementLevel *string, create bool) (*models.TaskConfigsEnvelope, error) {
	attributes := models.NewTaskConfigs_attributes()
	attributes.SetGlobal(enabled)

	if stages != nil {
		allowedStages := make([]models.TaskConfigs_attributes_allowedStages, len(stages))
		for i, stage := range stages {
			parsed, err := models.ParseTaskConfigs_attributes_allowedStages(stage)
			if err != nil || parsed == nil {
				return nil, fmt.Errorf("invalid run task stage %q", stage)
			}
			allowedStages[i] = *(parsed.(*models.TaskConfigs_attributes_allowedStages))
		}
		attributes.SetAllowedStages(allowedStages)
	}

	if enforcementLevel != nil {
		parsed, err := models.ParseTaskConfigs_attributes_enforcementLevel(*enforcementLevel)
		if err != nil || parsed == nil {
			return nil, fmt.Errorf("invalid run task enforcement level %q", *enforcementLevel)
		}
		attributes.SetEnforcementLevel(parsed.(*models.TaskConfigs_attributes_enforcementLevel))
	}

	data := models.NewTaskConfigs()
	data.SetAttributes(attributes)
	taskConfigType := models.TASKCONFIGS_TASKCONFIGS_TYPE
	data.SetTypeEscaped(&taskConfigType)

	if create {
		taskIdentifier := models.NewTasksIdentifier()
		taskIdentifier.SetId(&taskID)
		taskType := models.TASKS_TASKSIDENTIFIER_TYPE
		taskIdentifier.SetTypeEscaped(&taskType)
		task := models.NewTasksHasOne()
		task.SetData(taskIdentifier)

		organizationIdentifier := models.NewOrganizationsIdentifier()
		organizationIdentifier.SetId(&organization)
		organizationType := models.ORGANIZATIONS_ORGANIZATIONSIDENTIFIER_TYPE
		organizationIdentifier.SetTypeEscaped(&organizationType)
		ownerData := models.NewTaskConfigOwnerHasOne_TaskConfigOwnerHasOne_data()
		ownerData.SetOrganizationsIdentifier(organizationIdentifier)
		owner := models.NewTaskConfigOwnerHasOne()
		owner.SetData(ownerData)

		relationships := models.NewTaskConfigs_relationships()
		relationships.SetTask(task)
		relationships.SetOwner(owner)
		data.SetRelationships(relationships)
	}

	envelope := models.NewTaskConfigsEnvelope()
	envelope.SetData(data)
	return envelope, nil
}

func getOrganizationRunTaskConfig(ctx context.Context, client *tfe.Client, taskID, organization string) (models.TaskConfigsable, error) {
	ownerType := forownerapi.ORGANIZATIONS_GETQOWNERTYPEQUERYPARAMETERTYPE
	requestConfig := &kiota.RequestConfiguration[organizationsapi.ItemTaskConfigsForOwnerRequestBuilderGetQueryParameters]{
		QueryParameters: &organizationsapi.ItemTaskConfigsForOwnerRequestBuilderGetQueryParameters{
			QownerId:   &organization,
			QownerType: &ownerType,
			QtaskId:    &taskID,
		},
	}
	envelope, err := client.API.Organizations().ByOrganization_name(organization).TaskConfigs().ForOwner().Get(ctx, requestConfig)
	if err != nil {
		return nil, err
	}
	if envelope == nil {
		return nil, nil
	}
	return envelope.GetData(), nil
}

func NewOrganizationRunTaskGlobalSettingsResource() resource.Resource {
	return &resourceOrganizationRunTaskGlobalSettings{}
}

type resourceOrganizationRunTaskGlobalSettings struct {
	config ConfiguredClient
}

func (r *resourceOrganizationRunTaskGlobalSettings) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_run_task_global_settings"
}

func (r *resourceOrganizationRunTaskGlobalSettings) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The tfe_organization_run_task_global_settings resource creates, updates and destroys the [global settings](https://developer.hashicorp.com/terraform/cloud-docs/workspaces/settings/run-tasks#global-run-tasks) for an [Organization Run task](https://developer.hashicorp.com/terraform/cloud-docs/workspaces/settings/run-tasks#creating-a-run-task). Your organization must have the `global-run-task` [entitlement](https://developer.hashicorp.com/terraform/cloud-docs/api-docs#feature-entitlements) to use global run tasks.",
		Version:             0,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the Run task global settings.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the run task will be applied globally.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"enforcement_level": schema.StringAttribute{
				Description: fmt.Sprintf("The enforcement level of the global task. Valid values are %s.", sentenceList(
					workspaceRunTaskEnforcementLevels(),
					"`",
					"`",
					"and",
				)),
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOf(workspaceRunTaskEnforcementLevels()...),
				},
			},
			"stages": schema.ListAttribute{
				ElementType: types.StringType,
				Description: fmt.Sprintf("Which stages the task will run in. Valid values are one or more of %s.", sentenceList(
					workspaceRunTaskStages(),
					"`",
					"`",
					"and",
				)),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.UniqueValues(),
					listvalidator.ValueStringsAre(
						stringvalidator.OneOf(workspaceRunTaskStages()...),
					),
				},
				Required: true,
			},
			"task_id": schema.StringAttribute{
				Description: "The ID of the run task which will have the global settings applied.",
				Required:    true,
				// When the task changes force a replace
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *resourceOrganizationRunTaskGlobalSettings) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(ConfiguredClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected tfe.ConfiguredClient, got %T. This is a bug in the tfe provider, so please report it on GitHub.", req.ProviderData),
		)

		return
	}
	r.config = client
}

func (r *resourceOrganizationRunTaskGlobalSettings) getRunTask(ctx context.Context, taskID string, diags *diag.Diagnostics) (models.Tasksable, bool) {
	tflog.Error(ctx, fmt.Sprintf("Reading organization run task %s", taskID))
	taskEnvelope, err := r.config.ClientV2.API.Tasks().ById(taskID).Get(ctx, nil)
	if err != nil {
		if errors.Is(err, tfe.ErrNotFound) {
			return nil, true
		}
		diags.AddError("Error reading Organization Run Task", "Could not read Organization Run Task, unexpected error: "+err.Error())
		return nil, false
	}
	if taskEnvelope == nil || taskEnvelope.GetData() == nil {
		return nil, true
	}
	task := taskEnvelope.GetData()

	if taskGlobalConfiguration(task) == nil {
		diags.AddError("Organization does not support global run tasks",
			fmt.Sprintf("The task %s exists however it does not support global run tasks.", taskID),
		)
		return nil, false
	}

	return task, false
}

func (r *resourceOrganizationRunTaskGlobalSettings) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state modelDataTFEOrganizationRunTaskGlobalSettings

	// Read Terraform current state into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	taskID := state.TaskID.ValueString()

	task, notFound := r.getRunTask(ctx, taskID, &resp.Diagnostics)
	if notFound {
		resp.State.RemoveResource(ctx)
		return
	}
	if task == nil {
		return
	}

	result := dataModelFromTFEOrganizationRunTaskGlobalSettingsV2(task)
	organization := taskOrganizationID(task.GetRelationships())
	if organization != "" {
		taskConfig, err := getOrganizationRunTaskConfig(ctx, r.config.ClientV2, taskID, organization)
		if err != nil && !errors.Is(err, tfe.ErrNotFound) {
			resp.Diagnostics.AddError("Error reading Organization Run Task global settings", err.Error())
			return
		}
		if taskConfig != nil {
			result = dataModelFromTFEOrganizationRunTaskGlobalTaskConfig(taskID, taskConfig)
		}
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *resourceOrganizationRunTaskGlobalSettings) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	r.updateRunTask(ctx, &req.Plan, &resp.State, &resp.Diagnostics)
}

func (r *resourceOrganizationRunTaskGlobalSettings) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	r.updateRunTask(ctx, &req.Plan, &resp.State, &resp.Diagnostics)
}

func (r *resourceOrganizationRunTaskGlobalSettings) updateRunTask(ctx context.Context, tfPlan *tfsdk.Plan, tfState *tfsdk.State, diagnostics *diag.Diagnostics) {
	var plan modelDataTFEOrganizationRunTaskGlobalSettings

	// Read Terraform planned changes into the model
	diagnostics.Append(tfPlan.Get(ctx, &plan)...)
	if diagnostics.HasError() {
		return
	}

	taskID := plan.TaskID.ValueString()

	task, notFound := r.getRunTask(ctx, taskID, diagnostics)
	if notFound {
		diagnostics.AddError("Error reading Organization Run Task", fmt.Sprintf("Could not find Organization Run Task %s", taskID))
		return
	}
	if task == nil {
		return
	}

	var stageStrings []types.String
	if err := plan.Stages.ElementsAs(ctx, &stageStrings, false); err != nil && err.HasError() {
		diagnostics.Append(err...)
		return
	}

	stages := make([]string, len(stageStrings))
	for idx, s := range stageStrings {
		stages[idx] = s.ValueString()
	}

	organization := taskOrganizationID(task.GetRelationships())
	if organization == "" {
		diagnostics.AddError("Unable to update organization task", "The task response did not include its organization")
		return
	}

	taskConfig, err := getOrganizationRunTaskConfig(ctx, r.config.ClientV2, taskID, organization)
	if err != nil && !errors.Is(err, tfe.ErrNotFound) {
		diagnostics.AddError("Unable to update organization task", err.Error())
		return
	}

	taskConfigEnvelope, err := newOrganizationRunTaskGlobalTaskConfigEnvelope(taskID, organization, plan.Enabled.ValueBoolPointer(), stages, plan.EnforcementLevel.ValueStringPointer(), taskConfig == nil)
	if err != nil {
		diagnostics.AddError("Unable to update organization task", err.Error())
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Update task %s global settings", taskID))
	var updatedTaskConfigEnvelope models.TaskConfigsEnvelopeable
	if taskConfig == nil {
		updatedTaskConfigEnvelope, err = r.config.ClientV2.API.Organizations().ByOrganization_name(organization).TaskConfigs().Post(ctx, taskConfigEnvelope, nil)
	} else {
		updatedTaskConfigEnvelope, err = r.config.ClientV2.API.TaskConfigs().ByTask_config_id(valueOrZero(taskConfig.GetId())).Patch(ctx, taskConfigEnvelope, nil)
	}
	if errors.Is(err, tfe.ErrNotFound) {
		// Task configs are feature-gated on older TFE releases; preserve the existing task API behavior.
		taskEnvelope, fallbackErr := r.config.ClientV2.API.Tasks().ById(taskID).Patch(ctx, newOrganizationRunTaskGlobalSettingsEnvelope(taskID, plan.Enabled.ValueBoolPointer(), stages, plan.EnforcementLevel.ValueStringPointer()), nil)
		if fallbackErr != nil {
			diagnostics.AddError("Unable to update organization task", fallbackErr.Error())
			return
		}
		if taskEnvelope == nil || taskEnvelope.GetData() == nil {
			diagnostics.AddError("Unable to update organization task", "No task data was returned by the API")
			return
		}
		diagnostics.Append(tfState.Set(ctx, dataModelFromTFEOrganizationRunTaskGlobalSettingsV2(taskEnvelope.GetData()))...)
		return
	}
	if err != nil {
		diagnostics.AddError("Unable to update organization task", err.Error())
		return
	}
	if updatedTaskConfigEnvelope == nil || updatedTaskConfigEnvelope.GetData() == nil {
		diagnostics.AddError("Unable to update organization task", "No task configuration data was returned by the API")
		return
	}
	result := dataModelFromTFEOrganizationRunTaskGlobalTaskConfig(taskID, updatedTaskConfigEnvelope.GetData())

	diagnostics.Append(tfState.Set(ctx, &result)...)
}

func (r *resourceOrganizationRunTaskGlobalSettings) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state modelDataTFEOrganizationRunTaskGlobalSettings

	// Read Terraform planned changes into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	taskID := state.TaskID.ValueString()
	task, notFound := r.getRunTask(ctx, taskID, &resp.Diagnostics)
	if notFound {
		return
	}
	if task == nil {
		return
	}
	organization := taskOrganizationID(task.GetRelationships())
	if organization == "" {
		resp.Diagnostics.AddError("Unable to update organization task", "The task response did not include its organization")
		return
	}

	var stageValues []types.String
	resp.Diagnostics.Append(state.Stages.ElementsAs(ctx, &stageValues, false)...)
	if resp.Diagnostics.HasError() {
		return
	}
	stages := make([]string, len(stageValues))
	for i, stage := range stageValues {
		stages[i] = stage.ValueString()
	}

	taskConfig, err := getOrganizationRunTaskConfig(ctx, r.config.ClientV2, taskID, organization)
	if err != nil && !errors.Is(err, tfe.ErrNotFound) {
		resp.Diagnostics.AddError("Unable to update organization task", err.Error())
		return
	}

	e := false

	tflog.Debug(ctx, fmt.Sprintf("Disabling task %s global settings", taskID))
	if taskConfig != nil {
		envelope, envelopeErr := newOrganizationRunTaskGlobalTaskConfigEnvelope(taskID, organization, &e, stages, state.EnforcementLevel.ValueStringPointer(), false)
		if envelopeErr != nil {
			resp.Diagnostics.AddError("Unable to update organization task", envelopeErr.Error())
			return
		}
		_, err = r.config.ClientV2.API.TaskConfigs().ByTask_config_id(valueOrZero(taskConfig.GetId())).Patch(ctx, envelope, nil)
	} else {
		_, err = r.config.ClientV2.API.Tasks().ById(taskID).Patch(ctx, newOrganizationRunTaskGlobalSettingsEnvelope(taskID, &e, stages, state.EnforcementLevel.ValueStringPointer()), nil)
	}
	if err != nil && !errors.Is(err, tfe.ErrNotFound) {
		resp.Diagnostics.AddError("Unable to update organization task", err.Error())
		return
	}
	// Resource is implicitly deleted from resp.State if diagnostics have no errors.
}

func (r *resourceOrganizationRunTaskGlobalSettings) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	s := strings.SplitN(req.ID, "/", 2)
	if len(s) != 2 {
		resp.Diagnostics.AddError(
			"Error importing organization run task global settings",
			fmt.Sprintf("Invalid task input format: %s (expected <ORGANIZATION>/<TASK NAME>)", req.ID),
		)
		return
	}

	taskName := s[1]
	orgName := s[0]

	if task, err := fetchOrganizationRunTaskV2(taskName, orgName, r.config.ClientV2); err != nil {
		resp.Diagnostics.AddError(
			"Error importing organization run task",
			err.Error(),
		)
	} else if task == nil {
		resp.Diagnostics.AddError(
			"Error importing organization run task",
			"Task does not exist or does not support global settings",
		)
	} else {
		// We can never import the HMACkey (Write-only) so assume it's the default (empty)
		result := dataModelFromTFEOrganizationRunTaskGlobalSettingsV2(task)
		resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
	}
}
