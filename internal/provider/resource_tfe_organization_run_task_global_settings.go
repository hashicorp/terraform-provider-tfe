// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"

	tfe "github.com/hashicorp/go-tfe"
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

func dataModelFromTFEOrganizationRunTaskGlobalSettings(v tfe.RunTask) modelDataTFEOrganizationRunTaskGlobalSettings {
	result := modelDataTFEOrganizationRunTaskGlobalSettings{
		Enabled:          types.BoolNull(),
		ID:               types.StringValue(v.ID),
		TaskID:           types.StringValue(v.ID),
		EnforcementLevel: types.StringNull(),
		Stages:           types.ListNull(types.StringType),
	}

	if v.Global == nil {
		return result
	}

	result.Enabled = types.BoolValue(v.Global.Enabled)
	result.EnforcementLevel = types.StringValue(string(v.Global.EnforcementLevel))
	if stages, err := types.ListValueFrom(ctx, types.StringType, v.Global.Stages); err == nil {
		result.Stages = stages
	}

	return result
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
		Version: 0,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Service-generated identifier for the task",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the run task will be applied globally",
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
				Description: fmt.Sprintf("Which stages the task will run in. Valid values are %s.", sentenceList(
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
				Description: "The id of the run task.",
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

func (r *resourceOrganizationRunTaskGlobalSettings) getRunTask(ctx context.Context, taskID string, diags *diag.Diagnostics) *tfe.RunTask {
	tflog.Error(ctx, fmt.Sprintf("Reading organization run task %s", taskID))
	task, err := r.config.Client.RunTasks.Read(ctx, taskID)

	if err != nil || task == nil {
		diags.AddError("Error reading Organization Run Task", "Could not read Organization Run Task, unexpected error: "+err.Error())
		return nil
	}

	if task.Global == nil {
		diags.AddError("Organization does not support global run tasks",
			fmt.Sprintf("The task %s exists however it does not support global run tasks.", taskID),
		)
		return nil
	}

	return task
}

func (r *resourceOrganizationRunTaskGlobalSettings) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state modelDataTFEOrganizationRunTaskGlobalSettings

	// Read Terraform current state into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	taskID := state.TaskID.ValueString()

	task := r.getRunTask(ctx, taskID, &resp.Diagnostics)
	if task == nil {
		return
	}

	result := dataModelFromTFEOrganizationRunTaskGlobalSettings(*task)

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

	task := r.getRunTask(ctx, taskID, diagnostics)
	if task == nil {
		return
	}

	var stageStrings []types.String
	if err := plan.Stages.ElementsAs(ctx, &stageStrings, false); err != nil && err.HasError() {
		diagnostics.Append(err...)
		return
	}

	stages := make([]tfe.Stage, len(stageStrings))
	for idx, s := range stageStrings {
		stages[idx] = tfe.Stage(s.ValueString())
	}

	options := tfe.RunTaskUpdateOptions{
		Global: &tfe.GlobalRunTaskOptions{
			Enabled:          plan.Enabled.ValueBoolPointer(),
			Stages:           &stages,
			EnforcementLevel: (*tfe.TaskEnforcementLevel)(plan.EnforcementLevel.ValueStringPointer()),
		},
	}

	tflog.Debug(ctx, fmt.Sprintf("Update task %s global settings", taskID))
	task, err := r.config.Client.RunTasks.Update(ctx, taskID, options)
	if err != nil || task == nil {
		diagnostics.AddError("Unable to update organization task", err.Error())
		return
	}
	result := dataModelFromTFEOrganizationRunTaskGlobalSettings(*task)

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

	e := false
	options := tfe.RunTaskUpdateOptions{
		Global: &tfe.GlobalRunTaskOptions{
			Enabled: &e,
		},
	}

	tflog.Debug(ctx, fmt.Sprintf("Disabling task %s global settings", taskID))
	task, err := r.config.Client.RunTasks.Update(ctx, taskID, options)
	if err != nil || task == nil {
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

	if task, err := fetchOrganizationRunTask(taskName, orgName, r.config.Client); err != nil {
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
		result := dataModelFromTFEOrganizationRunTaskGlobalSettings(*task)
		resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
	}
}
