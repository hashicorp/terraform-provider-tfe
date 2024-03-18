// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	customValidators "github.com/hashicorp/terraform-provider-tfe/internal/provider/validators"
)

type resourceOrgRunTask struct {
	config ConfiguredClient
}

func NewOrganizationRunTaskResource() resource.Resource {
	return &resourceOrgRunTask{}
}

type modelTFEOrganizationRunTaskV0 struct {
	Category     types.String `tfsdk:"category"`
	Description  types.String `tfsdk:"description"`
	Enabled      types.Bool   `tfsdk:"enabled"`
	HMACKey      types.String `tfsdk:"hmac_key"`
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Organization types.String `tfsdk:"organization"`
	URL          types.String `tfsdk:"url"`
}

func modelFromTFEOrganizationRunTask(v *tfe.RunTask, hmacKey types.String) modelTFEOrganizationRunTaskV0 {
	result := modelTFEOrganizationRunTaskV0{
		Category:     types.StringValue(v.Category),
		Description:  types.StringValue(v.Description),
		Enabled:      types.BoolValue(v.Enabled),
		HMACKey:      types.StringValue(""), // This value is never emitted by the API so we inject it later
		ID:           types.StringValue(v.ID),
		Name:         types.StringValue(v.Name),
		Organization: types.StringValue(v.Organization.Name),
		URL:          types.StringValue(v.URL),
	}

	if len(hmacKey.String()) > 0 {
		result.HMACKey = hmacKey
	}

	return result
}

func (r *resourceOrgRunTask) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_run_task"
}

func (r *resourceOrgRunTask) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// If a Run Tasks uses the default organization, then if the deafault org. changes, it should trigger a modification
	modifyPlanForDefaultOrganizationChange(ctx, r.config.Organization, req.State, req.Config, req.Plan, resp)
}

// Configure implements resource.ResourceWithConfigure
func (r *resourceOrgRunTask) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *resourceOrgRunTask) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			"name": schema.StringAttribute{
				Required: true,
			},
			"organization": schema.StringAttribute{
				Optional: true,
				Computed: true,
				// From ForceNew: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"url": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					customValidators.IsURLWithHTTPorHTTPS(),
				},
			},
			"category": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("task"),
			},
			"hmac_key": schema.StringAttribute{
				Sensitive: true,
				Optional:  true,
				Computed:  true,
				Default:   stringdefault.StaticString(""),
			},
			"enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
			"description": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
			},
		},
	}
}

func (r *resourceOrgRunTask) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state modelTFEOrganizationRunTaskV0

	// Read Terraform current state into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	taskID := state.ID.ValueString()

	tflog.Debug(ctx, "Reading organization run task")
	task, err := r.config.Client.RunTasks.Read(ctx, taskID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading Organization Run Task", "Could not read Organization Run Task, unexpected error: "+err.Error())
		return
	}

	result := modelFromTFEOrganizationRunTask(task, state.HMACKey)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *resourceOrgRunTask) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan modelTFEOrganizationRunTaskV0

	// Read Terraform planned changes into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var organization string
	resp.Diagnostics.Append(r.config.dataOrDefaultOrganization(ctx, req.Plan, &organization)...)

	if resp.Diagnostics.HasError() {
		return
	}

	options := tfe.RunTaskCreateOptions{
		Name:        plan.Name.ValueString(),
		URL:         plan.URL.ValueString(),
		Category:    plan.Category.ValueString(),
		HMACKey:     plan.HMACKey.ValueStringPointer(),
		Enabled:     plan.Enabled.ValueBoolPointer(),
		Description: plan.Description.ValueStringPointer(),
	}

	tflog.Debug(ctx, fmt.Sprintf("Create task %s for organization: %s", options.Name, organization))
	task, err := r.config.Client.RunTasks.Create(ctx, organization, options)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create organization task", err.Error())
		return
	}

	result := modelFromTFEOrganizationRunTask(task, plan.HMACKey)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *resourceOrgRunTask) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan modelTFEOrganizationRunTaskV0

	// Read Terraform planned changes into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state modelTFEOrganizationRunTaskV0
	// Read Terraform state into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	options := tfe.RunTaskUpdateOptions{
		Name:        plan.Name.ValueStringPointer(),
		URL:         plan.URL.ValueStringPointer(),
		Category:    plan.Category.ValueStringPointer(),
		Enabled:     plan.Enabled.ValueBoolPointer(),
		Description: plan.Description.ValueStringPointer(),
	}

	// HMAC Key is a write-only value so we should only send it if
	// it really has changed.
	if plan.HMACKey.ValueString() != state.HMACKey.ValueString() {
		options.HMACKey = plan.HMACKey.ValueStringPointer()
	}

	taskID := plan.ID.ValueString()

	tflog.Debug(ctx, fmt.Sprintf("Update task %s", taskID))
	task, err := r.config.Client.RunTasks.Update(ctx, taskID, options)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update organization task", err.Error())
		return
	}

	result := modelFromTFEOrganizationRunTask(task, plan.HMACKey)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *resourceOrgRunTask) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state modelTFEOrganizationRunTaskV0
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	taskID := state.ID.ValueString()

	tflog.Debug(ctx, fmt.Sprintf("Delete task %s", taskID))
	err := r.config.Client.RunTasks.Delete(ctx, taskID)
	// Ignore 404s for delete
	if err != nil && !errors.Is(err, tfe.ErrResourceNotFound) {
		resp.Diagnostics.AddError(
			"Error deleting organization run task",
			fmt.Sprintf("Couldn't delete organization run task %s: %s", taskID, err.Error()),
		)
	}
	// Resource is implicitly deleted from resp.State if diagnostics have no errors.
}

func (r *resourceOrgRunTask) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	s := strings.SplitN(req.ID, "/", 2)
	if len(s) != 2 {
		resp.Diagnostics.AddError(
			"Error importing organization run task",
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
			"Task does not exist or has no details",
		)
	} else {
		// We can never import the HMACkey (Write-only) so assume it's the default (empty)
		result := modelFromTFEOrganizationRunTask(task, types.StringValue(""))
		resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
	}
}
