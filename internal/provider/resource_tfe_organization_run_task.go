// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
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

var _ resource.Resource = &resourceOrgRunTask{}
var _ resource.ResourceWithConfigure = &resourceOrgRunTask{}
var _ resource.ResourceWithImportState = &resourceOrgRunTask{}
var _ resource.ResourceWithModifyPlan = &resourceOrgRunTask{}

func NewOrganizationRunTaskResource() resource.Resource {
	return &resourceOrgRunTask{}
}

type modelTFEOrganizationRunTaskV0 struct {
	Category         types.String `tfsdk:"category"`
	Description      types.String `tfsdk:"description"`
	Enabled          types.Bool   `tfsdk:"enabled"`
	HMACKey          types.String `tfsdk:"hmac_key"`
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Organization     types.String `tfsdk:"organization"`
	URL              types.String `tfsdk:"url"`
	HMACKeyWO        types.String `tfsdk:"hmac_key_wo"`
	HMACKeyWOVersion types.Int64  `tfsdk:"hmac_key_wo_version"`
}

func modelFromTFEOrganizationRunTask(v *tfe.RunTask, hmacKey types.String, hmacKeyWOVersion types.Int64) modelTFEOrganizationRunTaskV0 {
	result := modelTFEOrganizationRunTaskV0{
		Category:         types.StringValue(v.Category),
		Description:      types.StringValue(v.Description),
		Enabled:          types.BoolValue(v.Enabled),
		HMACKey:          types.StringValue(""), // This value is never emitted by the API so we inject it later
		ID:               types.StringValue(v.ID),
		Name:             types.StringValue(v.Name),
		Organization:     types.StringValue(v.Organization.Name),
		URL:              types.StringValue(v.URL),
		HMACKeyWOVersion: hmacKeyWOVersion,
	}

	if len(hmacKey.String()) > 0 {
		result.HMACKey = hmacKey
	}

	// Don't retrieve values if write-only is being used. Unset the hmac key field before updating the state.
	isWriteOnlyValue := !hmacKeyWOVersion.IsNull()
	if isWriteOnlyValue {
		result.HMACKey = types.StringValue("")
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
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("hmac_key_wo")),
				},
			},
			// since the hmac_key_wo write-only values are not saved to state, they will not trigger updates on their own.
			// Instead the hmac_key_wo_version responsibility is to trigger updates to the hmac_key_wo attribute when version number changes.
			"hmac_key_wo": schema.StringAttribute{
				Optional:    true,
				WriteOnly:   true,
				Sensitive:   true,
				Description: "HMAC key in write-only mode",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("hmac_key")),
					stringvalidator.AlsoRequires(path.MatchRoot("hmac_key_wo_version")),
				},
			},

			"hmac_key_wo_version": schema.Int64Attribute{
				Optional:    true,
				Description: "Version of the write-only HMAC key to trigger updates",
				Validators: []validator.Int64{
					int64validator.ConflictsWith(path.MatchRoot("hmac_key")),
					int64validator.AlsoRequires(path.MatchRoot("hmac_key_wo")),
				},
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
		if errors.Is(err, tfe.ErrResourceNotFound) {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError("Error reading Organization Run Task", "Could not read Organization Run Task, unexpected error: "+err.Error())
		}
		return
	}

	// update state
	result := modelFromTFEOrganizationRunTask(task, state.HMACKey, state.HMACKeyWOVersion)
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

	var config modelTFEOrganizationRunTaskV0
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
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
		Enabled:     plan.Enabled.ValueBoolPointer(),
		Description: plan.Description.ValueStringPointer(),
	}
	// Set Value from "hmac_key_wo" if set, otherwise use the normal value
	if !config.HMACKeyWO.IsNull() {
		options.HMACKey = config.HMACKeyWO.ValueStringPointer()
	} else {
		options.HMACKey = plan.HMACKey.ValueStringPointer()
	}

	tflog.Debug(ctx, fmt.Sprintf("Create task %s for organization: %s", options.Name, organization))
	task, err := r.config.Client.RunTasks.Create(ctx, organization, options)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create organization task", err.Error())
		return
	}

	result := modelFromTFEOrganizationRunTask(task, plan.HMACKey, config.HMACKeyWOVersion)

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
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var config modelTFEOrganizationRunTaskV0
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
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
	keyToUpdate := r.determineHMACKeyForUpdate(plan, state, config)
	if keyToUpdate != nil {
		options.HMACKey = keyToUpdate
	}

	taskID := plan.ID.ValueString()

	tflog.Debug(ctx, fmt.Sprintf("Update task %s", taskID))

	task, err := r.config.Client.RunTasks.Update(ctx, taskID, options)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update organization task", err.Error())
		return
	}

	result := modelFromTFEOrganizationRunTask(task, plan.HMACKey, config.HMACKeyWOVersion)
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
		result := modelFromTFEOrganizationRunTask(task, types.StringValue(""), types.Int64Null())
		resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
	}
}

// determineHMACKeyForUpdate is invoked only after terraform determines that an attribute update is needed.
// note that the update can be triggered by other attributes outside of the key/key_wo attributes.
// this function compares the KeyWOVersion vs Key to ensure that during api update call, key is not mistakenly unset.
// Returns nil if no value update is needed.
func (r *resourceOrgRunTask) determineHMACKeyForUpdate(plan, state, config modelTFEOrganizationRunTaskV0) *string {
	// Determine if we're using write-only HMAC key in plan vs state
	usingWriteOnlyInPlan := !plan.HMACKeyWOVersion.IsNull()
	usingWriteOnlyInState := !state.HMACKeyWOVersion.IsNull()

	// Case 1: Switching FROM hmac_key TO hmac_key_wo
	if !usingWriteOnlyInState && usingWriteOnlyInPlan && !config.HMACKeyWO.IsNull() {
		return config.HMACKeyWO.ValueStringPointer()
	}
	// Case 2: Switching FROM hmac_key_wo TO hmac_key
	if usingWriteOnlyInState && !usingWriteOnlyInPlan && !plan.HMACKey.IsNull() {
		return plan.HMACKey.ValueStringPointer()
	}
	// Case 3: hmac_key_wo version changed in plan
	if usingWriteOnlyInPlan && plan.HMACKeyWOVersion.ValueInt64() != state.HMACKeyWOVersion.ValueInt64() && !config.HMACKeyWO.IsNull() {
		return config.HMACKeyWO.ValueStringPointer()
	}
	// Case 4: Regular hmac_key changed. Only set HMACKey if our planned value would be a CHANGE from
	// the prior state. This prevents accidentally resetting the HMAC key on unrelated changes.
	if state.HMACKey.ValueString() != plan.HMACKey.ValueString() {
		return plan.HMACKey.ValueStringPointer()
	}
	return nil
}
