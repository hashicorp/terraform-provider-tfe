// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	tfe "github.com/hashicorp/go-tfe/v2"
	"github.com/hashicorp/go-tfe/v2/api/models"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &resourceTaskConfig{}
var _ resource.ResourceWithConfigure = &resourceTaskConfig{}
var _ resource.ResourceWithImportState = &resourceTaskConfig{}
var _ resource.ResourceWithModifyPlan = &resourceTaskConfig{}

type modelTaskConfig struct {
	Enabled          types.Bool   `tfsdk:"enabled"`
	EnforcementLevel types.String `tfsdk:"enforcement_level"`
	ID               types.String `tfsdk:"id"`
	Owner            types.String `tfsdk:"owner"`
	Stages           types.List   `tfsdk:"stages"`
	TaskID           types.String `tfsdk:"task_id"`
}

type resourceTaskConfig struct {
	config ConfiguredClient
}

func NewTaskConfigResource() resource.Resource {
	return &resourceTaskConfig{}
}

func (r *resourceTaskConfig) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_task_config"
}

func (r *resourceTaskConfig) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages the global configuration for an organization run task. This resource manages the same task configuration as `tfe_organization_run_task_global_settings`; do not manage the same owner and task with both resources. Destroying this resource disables the configuration.",
		Version:             0,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the task configuration.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"owner": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The name of the organization that owns the task configuration. If omitted, the provider-level organization must be configured.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"task_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the run task to configure.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether the run task is applied globally.",
			},
			"enforcement_level": schema.StringAttribute{
				Required:    true,
				Description: fmt.Sprintf("The enforcement level of the global task. Valid values are %s.", sentenceList(workspaceRunTaskEnforcementLevels(), "`", "`", "and")),
				Validators: []validator.String{
					stringvalidator.OneOf(workspaceRunTaskEnforcementLevels()...),
				},
			},
			"stages": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: fmt.Sprintf("Which stages the task will run in. Valid values are one or more of %s.", sentenceList(workspaceRunTaskStages(), "`", "`", "and")),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.UniqueValues(),
					listvalidator.ValueStringsAre(stringvalidator.OneOf(workspaceRunTaskStages()...)),
				},
			},
		},
	}
}

func (r *resourceTaskConfig) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *resourceTaskConfig) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.State.Raw.IsNull() || r.config.Organization == "" {
		return
	}

	ownerPath := path.Root("owner")
	var configuredOwner, plannedOwner types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, ownerPath, &configuredOwner)...)
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, ownerPath, &plannedOwner)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if configuredOwner.IsNull() && !plannedOwner.IsNull() && r.config.Organization != plannedOwner.ValueString() {
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, ownerPath, types.StringValue(r.config.Organization))...)
		resp.RequiresReplace.Append(ownerPath)
	}
}

func (r *resourceTaskConfig) resolveOwner(ctx context.Context, data AttrGettable, diagnostics *diag.Diagnostics) string {
	ownerPath := path.Root("owner")
	var owner types.String
	diagnostics.Append(data.GetAttribute(ctx, ownerPath, &owner)...)
	if diagnostics.HasError() {
		return ""
	}

	if !owner.IsNull() && !owner.IsUnknown() {
		return owner.ValueString()
	}
	if r.config.Organization == "" {
		diagnostics.AddAttributeError(ownerPath, "No owner was specified on the resource or provider", "")
		return ""
	}
	return r.config.Organization
}

func modelFromTaskConfig(ctx context.Context, config models.TaskConfigsable, owner, taskID string) modelTaskConfig {
	result := modelTaskConfig{
		ID:               types.StringValue(valueOrZero(config.GetId())),
		Owner:            types.StringValue(owner),
		TaskID:           types.StringValue(taskID),
		Enabled:          types.BoolNull(),
		EnforcementLevel: types.StringNull(),
		Stages:           types.ListNull(types.StringType),
	}

	if config.GetRelationships() != nil {
		if configOwner := config.GetRelationships().GetOwner(); configOwner != nil && configOwner.GetData() != nil {
			result.Owner = types.StringValue(valueOrZero(configOwner.GetData().GetId()))
		}
		if task := config.GetRelationships().GetTask(); task != nil && task.GetData() != nil {
			result.TaskID = types.StringValue(valueOrZero(task.GetData().GetId()))
		}
	}

	if attributes := config.GetAttributes(); attributes != nil {
		result.Enabled = types.BoolValue(valueOrZero(attributes.GetGlobal()))
		if level := attributes.GetEnforcementLevel(); level != nil {
			result.EnforcementLevel = types.StringValue(level.String())
		}
		stages := make([]string, len(attributes.GetAllowedStages()))
		for i, stage := range attributes.GetAllowedStages() {
			stages[i] = stage.String()
		}
		if value, diagnostics := types.ListValueFrom(ctx, types.StringType, stages); !diagnostics.HasError() {
			result.Stages = value
		}
	}

	return result
}

func (r *resourceTaskConfig) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state modelTaskConfig
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	owner := state.Owner.ValueString()
	taskID := state.TaskID.ValueString()
	config, err := getOrganizationRunTaskConfig(ctx, r.config.ClientV2, taskID, owner)
	if errors.Is(err, tfe.ErrNotFound) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Error reading task config", err.Error())
		return
	}
	if config == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	result := modelFromTaskConfig(ctx, config, owner, taskID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *resourceTaskConfig) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	r.upsert(ctx, &req.Plan, &resp.State, &resp.Diagnostics)
}

func (r *resourceTaskConfig) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	r.upsert(ctx, &req.Plan, &resp.State, &resp.Diagnostics)
}

func (r *resourceTaskConfig) upsert(ctx context.Context, planData *tfsdk.Plan, state *tfsdk.State, diagnostics *diag.Diagnostics) {
	var plan modelTaskConfig
	diagnostics.Append(planData.Get(ctx, &plan)...)
	if diagnostics.HasError() {
		return
	}

	owner := r.resolveOwner(ctx, planData, diagnostics)
	if diagnostics.HasError() {
		return
	}
	taskID := plan.TaskID.ValueString()

	var stages []types.String
	diagnostics.Append(plan.Stages.ElementsAs(ctx, &stages, false)...)
	if diagnostics.HasError() {
		return
	}
	stageValues := make([]string, len(stages))
	for i, stage := range stages {
		stageValues[i] = stage.ValueString()
	}

	current, err := getOrganizationRunTaskConfig(ctx, r.config.ClientV2, taskID, owner)
	if err != nil && !errors.Is(err, tfe.ErrNotFound) {
		diagnostics.AddError("Unable to read task config before update", err.Error())
		return
	}

	envelope, err := newOrganizationRunTaskGlobalTaskConfigEnvelope(taskID, owner, plan.Enabled.ValueBoolPointer(), stageValues, plan.EnforcementLevel.ValueStringPointer(), current == nil)
	if err != nil {
		diagnostics.AddError("Unable to build task config request", err.Error())
		return
	}

	var updated models.TaskConfigsEnvelopeable
	if current == nil {
		tflog.Debug(ctx, "Creating task config", map[string]any{"owner": owner, "task_id": taskID})
		updated, err = r.config.ClientV2.API.Organizations().ByOrganization_name(owner).TaskConfigs().Post(ctx, envelope, nil)
	} else {
		tflog.Debug(ctx, "Updating task config", map[string]any{"owner": owner, "task_id": taskID})
		updated, err = r.config.ClientV2.API.TaskConfigs().ByExternal_id(valueOrZero(current.GetId())).Patch(ctx, envelope, nil)
	}
	if err != nil {
		diagnostics.AddError("Unable to save task config", err.Error())
		return
	}
	if updated == nil || updated.GetData() == nil {
		diagnostics.AddError("Unable to save task config", "No task config data was returned by the API")
		return
	}

	result := modelFromTaskConfig(ctx, updated.GetData(), owner, taskID)
	diagnostics.Append(state.Set(ctx, &result)...)
}

func (r *resourceTaskConfig) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state modelTaskConfig
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	owner := state.Owner.ValueString()
	taskID := state.TaskID.ValueString()
	current, err := getOrganizationRunTaskConfig(ctx, r.config.ClientV2, taskID, owner)
	if errors.Is(err, tfe.ErrNotFound) {
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to read task config before disabling", err.Error())
		return
	}
	if current == nil {
		return
	}

	var stages []string
	var enforcementLevel *string
	if attributes := current.GetAttributes(); attributes != nil {
		for _, stage := range attributes.GetAllowedStages() {
			stages = append(stages, stage.String())
		}
		if level := attributes.GetEnforcementLevel(); level != nil {
			value := level.String()
			enforcementLevel = &value
		}
	}
	enabled := false
	envelope, err := newOrganizationRunTaskGlobalTaskConfigEnvelope(taskID, owner, &enabled, stages, enforcementLevel, false)
	if err != nil {
		resp.Diagnostics.AddError("Unable to build task config disable request", err.Error())
		return
	}

	_, err = r.config.ClientV2.API.TaskConfigs().ByExternal_id(valueOrZero(current.GetId())).Patch(ctx, envelope, nil)
	if err != nil && !errors.Is(err, tfe.ErrNotFound) {
		resp.Diagnostics.AddError("Unable to disable task config", err.Error())
	}
}

func (r *resourceTaskConfig) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Error importing task config", fmt.Sprintf("Invalid import ID %q; expected <OWNER>/<TASK NAME>", req.ID))
		return
	}
	owner, taskName := parts[0], parts[1]

	task, err := fetchOrganizationRunTaskV2(taskName, owner, r.config.ClientV2)
	if err != nil {
		resp.Diagnostics.AddError("Error importing task config", err.Error())
		return
	}
	if task == nil || task.GetId() == nil {
		resp.Diagnostics.AddError("Error importing task config", "Run task was not found for the specified owner")
		return
	}

	taskID := valueOrZero(task.GetId())
	config, err := getOrganizationRunTaskConfig(ctx, r.config.ClientV2, taskID, owner)
	if err != nil {
		resp.Diagnostics.AddError("Error importing task config", err.Error())
		return
	}
	if config == nil {
		resp.Diagnostics.AddError("Error importing task config", "Task config was not found for the specified owner and task")
		return
	}

	result := modelFromTaskConfig(ctx, config, owner, taskID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}
