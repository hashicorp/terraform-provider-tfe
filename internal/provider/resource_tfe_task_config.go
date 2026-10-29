// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"

	tfe "github.com/hashicorp/go-tfe/v2"
	"github.com/hashicorp/go-tfe/v2/api/models"
	taskconfigsapi "github.com/hashicorp/go-tfe/v2/api/taskconfigs"
	taskconfigsitem "github.com/hashicorp/go-tfe/v2/api/taskconfigs/item"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type resourceTaskConfig struct {
	config ConfiguredClient
}

var _ resource.Resource = &resourceTaskConfig{}
var _ resource.ResourceWithConfigure = &resourceTaskConfig{}
var _ resource.ResourceWithImportState = &resourceTaskConfig{}
var _ resource.ResourceWithValidateConfig = &resourceTaskConfig{}

func NewTaskConfigResource() resource.Resource {
	return &resourceTaskConfig{}
}

type modelTFETaskConfigV0 struct {
	EnforcementLevel types.String `tfsdk:"enforcement_level"`
	Global           types.Bool   `tfsdk:"global"`
	ID               types.String `tfsdk:"id"`
	Organization     types.String `tfsdk:"organization"`
	OwnerID          types.String `tfsdk:"owner_id"`
	ProjectIDs       types.Set    `tfsdk:"project_ids"`
	Stages           types.Set    `tfsdk:"stages"`
	TaskID           types.String `tfsdk:"task_id"`
}

// taskConfigInput holds the user-configurable values sent to the API on create and update.
// A nil projectIDs omits the projects relationship, leaving existing projects untouched.
type taskConfigInput struct {
	global           bool
	stages           []string
	enforcementLevel string
	projectIDs       []string
}

func modelFromTFETaskConfig(v models.TaskConfigsable, fallback taskConfigFallback) (modelTFETaskConfigV0, diag.Diagnostics) {
	result := modelTFETaskConfigV0{
		ID:           types.StringValue(valueOrZero(v.GetId())),
		Organization: types.StringValue(fallback.organization),
		// Task configs are only owned by organizations for now, so the fallback is the owner too.
		OwnerID: types.StringValue(fallback.organization),
		TaskID:  types.StringValue(fallback.taskID),
	}

	stages := []attr.Value{}
	if attrs := v.GetAttributes(); attrs != nil {
		result.Global = types.BoolValue(valueOrZero(attrs.GetGlobal()))
		result.EnforcementLevel = types.StringValue(enumStringOrEmpty(attrs.GetEnforcementLevel()))
		for _, stage := range attrs.GetAllowedStages() {
			stages = append(stages, types.StringValue(stage.String()))
		}
	}

	projectIDs := []attr.Value{}
	for _, id := range fallback.projectIDs {
		projectIDs = append(projectIDs, types.StringValue(id))
	}
	if rels := v.GetRelationships(); rels != nil {
		orgId := getIDForOrg(rels.GetOrganization())
		if orgId != "" {
			result.Organization = types.StringValue(orgId)
		}
		taskId := getIDForTask(rels.GetTask())
		if taskId != "" {
			result.TaskID = types.StringValue(taskId)
		}
		ownerId := getIDForOwner(rels.GetOwner())
		if ownerId != "" {
			result.OwnerID = types.StringValue(ownerId)
		}
		projectIDs = getIDsForProjects(projectIDs, rels.GetProjects())
	}

	var diags diag.Diagnostics
	var d diag.Diagnostics
	result.Stages, d = types.SetValue(types.StringType, stages)
	diags.Append(d...)
	result.ProjectIDs, d = types.SetValue(types.StringType, projectIDs)
	diags.Append(d...)

	return result, diags
}

func getIDForOrg(organizationsHasOneable models.OrganizationsHasOneable) string {
	if organizationsHasOneable != nil && organizationsHasOneable.GetData() != nil {
		return valueOrZero(organizationsHasOneable.GetData().GetId())
	}
	return ""
}

func getIDForTask(taskHasOneable models.TasksHasOneable) string {
	if taskHasOneable != nil && taskHasOneable.GetData() != nil {
		return valueOrZero(taskHasOneable.GetData().GetId())
	}
	return ""
}

func getIDsForProjects(projectIDs []attr.Value, projectsHasManyable models.ProjectsHasManyable) []attr.Value {
	// Present data, even when empty, is authoritative over the fallback.
	if projectsHasManyable != nil && projectsHasManyable.GetData() != nil {
		projectIDs = []attr.Value{}
		for _, project := range projectsHasManyable.GetData() {
			projectIDs = append(projectIDs, types.StringValue(valueOrZero(project.GetId())))
		}
	}
	return projectIDs
}

func getIDForOwner(ownerHasOneable models.TaskConfigOwnerHasOneable) string {
	if ownerHasOneable != nil && ownerHasOneable.GetData() != nil {
		return valueOrZero(ownerHasOneable.GetData().GetId())
	}
	return ""
}

// taskConfigFallback holds known values used when the API response omits a relationship's data.
type taskConfigFallback struct {
	organization string
	taskID       string
	projectIDs   []string
}

func taskConfigFallbackFromModel(ctx context.Context, m modelTFETaskConfigV0, organization string) (taskConfigFallback, diag.Diagnostics) {
	var diags diag.Diagnostics
	fallback := taskConfigFallback{organization: organization, taskID: m.TaskID.ValueString()}
	if !m.ProjectIDs.IsNull() && !m.ProjectIDs.IsUnknown() {
		diags.Append(m.ProjectIDs.ElementsAs(ctx, &fallback.projectIDs, false)...)
	}
	return fallback, diags
}

// getTaskConfig reads a task config with its relationships side-loaded, since the API may otherwise omit their data.
func getTaskConfig(ctx context.Context, client *tfe.Client, id string) (models.TaskConfigsEnvelopeable, error) {
	return client.API.TaskConfigs().ByExternal_id(id).Get(ctx, withQueryParams(&taskconfigsapi.WithExternal_ItemRequestBuilderGetQueryParameters{
		Include: []taskconfigsitem.GetIncludeQueryParameterType{
			taskconfigsitem.TASK_GETINCLUDEQUERYPARAMETERTYPE,
			taskconfigsitem.OWNER_GETINCLUDEQUERYPARAMETERTYPE,
			taskconfigsitem.ORGANIZATION_GETINCLUDEQUERYPARAMETERTYPE,
			taskconfigsitem.PROJECTS_GETINCLUDEQUERYPARAMETERTYPE,
		},
	}))
}

// newTaskConfigAttributes builds the attributes shared by the create and update envelopes.
func newTaskConfigAttributes(in taskConfigInput) (*models.TaskConfigs_attributes, error) {
	allowedStages := make([]models.TaskConfigs_attributes_allowedStages, len(in.stages))
	for i, stage := range in.stages {
		parsed := parseEnumPtr[models.TaskConfigs_attributes_allowedStages](models.ParseTaskConfigs_attributes_allowedStages, stage)
		if parsed == nil {
			return nil, fmt.Errorf("invalid task config stage %q", stage)
		}
		allowedStages[i] = *parsed
	}

	enforcementLevel := parseEnumPtr[models.TaskConfigs_attributes_enforcementLevel](models.ParseTaskConfigs_attributes_enforcementLevel, in.enforcementLevel)
	if enforcementLevel == nil {
		return nil, fmt.Errorf("invalid task config enforcement level %q", in.enforcementLevel)
	}

	attributes := models.NewTaskConfigs_attributes()
	attributes.SetGlobal(&in.global)
	attributes.SetAllowedStages(allowedStages)
	attributes.SetEnforcementLevel(enforcementLevel)
	return attributes, nil
}

func newTaskConfigProjectsRelationship(projectIDs []string) *models.ProjectsHasMany {
	identifiers := make([]models.ProjectsIdentifierable, len(projectIDs))
	for i, id := range projectIDs {
		identifier := models.NewProjectsIdentifier()
		identifier.SetId(ptr(id))
		identifier.SetTypeEscaped(ptr(models.PROJECTS_PROJECTSIDENTIFIER_TYPE))
		identifiers[i] = identifier
	}

	projects := models.NewProjectsHasMany()
	projects.SetData(identifiers)
	return projects
}

// newTaskConfigCreateEnvelope builds the request body for creating an organization-owned task config.
func newTaskConfigCreateEnvelope(taskID, organization string, in taskConfigInput) (*models.TaskConfigsEnvelope, error) {
	attributes, err := newTaskConfigAttributes(in)
	if err != nil {
		return nil, err
	}

	taskIdentifier := models.NewTasksIdentifier()
	taskIdentifier.SetId(&taskID)
	taskIdentifier.SetTypeEscaped(ptr(models.TASKS_TASKSIDENTIFIER_TYPE))
	task := models.NewTasksHasOne()
	task.SetData(taskIdentifier)

	ownerIdentifier := models.NewOrganizationsIdentifier()
	ownerIdentifier.SetId(&organization)
	ownerIdentifier.SetTypeEscaped(ptr(models.ORGANIZATIONS_ORGANIZATIONSIDENTIFIER_TYPE))
	owner := models.NewTaskConfigOwnerHasOne()
	owner.SetData(ownerIdentifier)

	relationships := models.NewTaskConfigs_relationships()
	relationships.SetTask(task)
	relationships.SetOwner(owner)
	if in.projectIDs != nil {
		relationships.SetProjects(newTaskConfigProjectsRelationship(in.projectIDs))
	}

	data := models.NewTaskConfigs()
	data.SetAttributes(attributes)
	data.SetRelationships(relationships)
	data.SetTypeEscaped(ptr(models.TASKCONFIGS_TASKCONFIGS_TYPE))

	envelope := models.NewTaskConfigsEnvelope()
	envelope.SetData(data)
	return envelope, nil
}

// newTaskConfigUpdateEnvelope builds the request body for updating an existing task config. The task and owner are immutable.
func newTaskConfigUpdateEnvelope(id string, in taskConfigInput) (*models.TaskConfigsEnvelope, error) {
	attributes, err := newTaskConfigAttributes(in)
	if err != nil {
		return nil, err
	}

	relationships := models.NewTaskConfigs_relationships()
	if in.projectIDs != nil {
		relationships.SetProjects(newTaskConfigProjectsRelationship(in.projectIDs))
	}

	data := models.NewTaskConfigs()
	data.SetId(&id)
	data.SetAttributes(attributes)
	data.SetRelationships(relationships)
	data.SetTypeEscaped(ptr(models.TASKCONFIGS_TASKCONFIGS_TYPE))

	envelope := models.NewTaskConfigsEnvelope()
	envelope.SetData(data)
	return envelope, nil
}

func taskConfigInputFromModel(ctx context.Context, m modelTFETaskConfigV0) (taskConfigInput, diag.Diagnostics) {
	var diags diag.Diagnostics
	in := taskConfigInput{
		global:           m.Global.ValueBool(),
		enforcementLevel: m.EnforcementLevel.ValueString(),
		projectIDs:       []string{},
	}
	diags.Append(m.Stages.ElementsAs(ctx, &in.stages, false)...)
	diags.Append(m.ProjectIDs.ElementsAs(ctx, &in.projectIDs, false)...)
	return in, diags
}

func (r *resourceTaskConfig) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_task_config"
}

func (r *resourceTaskConfig) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var global types.Bool
	var projectIDs types.Set
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("global"), &global)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("project_ids"), &projectIDs)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// An unknown set may still resolve to empty, so only reject values known to be non-empty.
	if global.ValueBool() && !projectIDs.IsNull() && (projectIDs.IsUnknown() || len(projectIDs.Elements()) > 0) {
		resp.Diagnostics.AddAttributeError(
			path.Root("project_ids"),
			"Conflicting configuration",
			"`project_ids` cannot be set when `global` is true. A global task config applies to all workspaces in the organization.",
		)
	}
}

// organizationForTask returns the configured organization, or looks it up from the task when none is configured.
func (r *resourceTaskConfig) organizationForTask(ctx context.Context, configured types.String, taskID string) (string, diag.Diagnostics) {
	var diags diag.Diagnostics
	if !configured.IsNull() && !configured.IsUnknown() && configured.ValueString() != "" {
		return configured.ValueString(), diags
	}

	taskEnvelope, err := r.config.ClientV2.API.Tasks().ById(taskID).Get(ctx, nil)
	if err != nil {
		diags.AddError("Unable to determine the task's organization", fmt.Sprintf("Could not read task %s: %s", taskID, apiErrorDetail(err)))
		return "", diags
	}
	if taskEnvelope == nil || taskEnvelope.GetData() == nil {
		diags.AddError("Unable to determine the task's organization", fmt.Sprintf("No data was returned for task %s", taskID))
		return "", diags
	}

	organization := taskOrganizationID(taskEnvelope.GetData().GetRelationships())
	if organization == "" {
		diags.AddError("Unable to determine the task's organization", fmt.Sprintf("The response for task %s did not include its organization", taskID))
	}
	return organization, diags
}

// Configure implements resource.ResourceWithConfigure
func (r *resourceTaskConfig) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
		return
	}
	r.config = client
}

func (r *resourceTaskConfig) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Creates, updates and destroys the configuration of an [Organization Run task](https://developer.hashicorp.com/terraform/cloud-docs/workspaces/settings/run-tasks), controlling the stages it runs in, its enforcement level, and whether it applies globally or to specific projects.",
		Version:             0,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the task config.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"organization": schema.StringAttribute{
				Description: "Name of the organization that owns the task. If omitted, it is inferred from `task_id`.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"owner_id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the resource that owns the task config. Currently always the organization.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"task_id": schema.StringAttribute{
				Description: "The ID of the run task this configuration applies to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"global": schema.BoolAttribute{
				Description: "Whether the run task applies to all workspaces in the organization. Cannot be true when `project_ids` is set.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"enforcement_level": schema.StringAttribute{
				Description: fmt.Sprintf("The enforcement level of the task. Valid values are %s.", sentenceList(
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
			"stages": schema.SetAttribute{
				ElementType: types.StringType,
				Description: fmt.Sprintf("The stages the task will run in. Valid values are one or more of %s.", sentenceList(
					workspaceRunTaskStages(),
					"`",
					"`",
					"and",
				)),
				Required: true,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
					setvalidator.ValueStringsAre(
						stringvalidator.OneOf(workspaceRunTaskStages()...),
					),
				},
			},
			"project_ids": schema.SetAttribute{
				ElementType: types.StringType,
				Description: "The IDs of the projects the task applies to. Cannot be set when `global` is true.",
				Optional:    true,
				Computed:    true,
				Default:     setdefault.StaticValue(types.SetValueMust(types.StringType, []attr.Value{})),
			},
		},
	}
}

func (r *resourceTaskConfig) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state modelTFETaskConfigV0

	// Read Terraform current state into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	taskConfigID := state.ID.ValueString()

	tflog.Debug(ctx, fmt.Sprintf("Reading task config %s", taskConfigID))
	envelope, err := getTaskConfig(ctx, r.config.ClientV2, taskConfigID)
	if err != nil {
		if errors.Is(err, tfe.ErrNotFound) {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError("Error reading task config", "Could not read task config, unexpected error: "+apiErrorDetail(err))
		}
		return
	}
	if envelope == nil || envelope.GetData() == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	fallback, diags := taskConfigFallbackFromModel(ctx, state, state.Organization.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, diags := modelFromTFETaskConfig(envelope.GetData(), fallback)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *resourceTaskConfig) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan modelTFETaskConfigV0

	// Read Terraform planned changes into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	organization, diags := r.organizationForTask(ctx, plan.Organization, plan.TaskID.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, diags := taskConfigInputFromModel(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	envelope, err := newTaskConfigCreateEnvelope(plan.TaskID.ValueString(), organization, input)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create task config", err.Error())
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Create task config for task %s in organization: %s", plan.TaskID.ValueString(), organization))
	created, err := r.config.ClientV2.API.Organizations().ByOrganization_name(organization).TaskConfigs().Post(ctx, envelope, nil)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create task config", apiErrorDetail(err))
		return
	}
	if created == nil || created.GetData() == nil {
		resp.Diagnostics.AddError("Unable to create task config", "No task config data was returned by the API")
		return
	}

	fallback, diags := taskConfigFallbackFromModel(ctx, plan, organization)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, diags := modelFromTFETaskConfig(created.GetData(), fallback)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *resourceTaskConfig) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan modelTFETaskConfigV0

	// Read Terraform planned changes into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, diags := taskConfigInputFromModel(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	taskConfigID := plan.ID.ValueString()

	envelope, err := newTaskConfigUpdateEnvelope(taskConfigID, input)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update task config", err.Error())
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Update task config %s", taskConfigID))
	updated, err := r.config.ClientV2.API.TaskConfigs().ByExternal_id(taskConfigID).Patch(ctx, envelope, nil)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update task config", apiErrorDetail(err))
		return
	}
	if updated == nil || updated.GetData() == nil {
		resp.Diagnostics.AddError("Unable to update task config", "No task config data was returned by the API")
		return
	}

	fallback, diags := taskConfigFallbackFromModel(ctx, plan, plan.Organization.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, diags := modelFromTFETaskConfig(updated.GetData(), fallback)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *resourceTaskConfig) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state modelTFETaskConfigV0
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	taskConfigID := state.ID.ValueString()

	tflog.Debug(ctx, fmt.Sprintf("Delete task config %s", taskConfigID))
	err := r.config.ClientV2.API.TaskConfigs().ByExternal_id(taskConfigID).Delete(ctx, nil)
	// Ignore 404s for delete
	if err != nil && !errors.Is(err, tfe.ErrNotFound) {
		resp.Diagnostics.AddError(
			"Error deleting task config",
			fmt.Sprintf("Couldn't delete task config %s: %s", taskConfigID, apiErrorDetail(err)),
		)
	}
	// Resource is implicitly deleted from resp.State if diagnostics have no errors.
}

func (r *resourceTaskConfig) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
