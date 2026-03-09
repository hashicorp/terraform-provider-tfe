// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &resourceTFEPolicySetParameter{}
	_ resource.ResourceWithConfigure   = &resourceTFEPolicySetParameter{}
	_ resource.ResourceWithImportState = &resourceTFEPolicySetParameter{}
)

type resourceTFEPolicySetParameter struct {
	config ConfiguredClient
}

func NewPolicySetParameterResource() resource.Resource {
	return &resourceTFEPolicySetParameter{}
}

type modelTFEPolicySetParameter struct {
	ID             types.String `tfsdk:"id"`
	Key            types.String `tfsdk:"key"`
	Value          types.String `tfsdk:"value"`
	ValueWO        types.String `tfsdk:"value_wo"`
	ValueWOVersion types.Int64  `tfsdk:"value_wo_version"`
	Sensitive      types.Bool   `tfsdk:"sensitive"`
	PolicySetID    types.String `tfsdk:"policy_set_id"`
}

func modelFromTFEPolicySetParameter(v *tfe.PolicySetParameter, lastValue types.String, valueWOVersion types.Int64) modelTFEPolicySetParameter {
	p := modelTFEPolicySetParameter{
		ID:             types.StringValue(v.ID),
		Key:            types.StringValue(v.Key),
		Value:          types.StringValue(v.Value),
		ValueWOVersion: valueWOVersion,
		Sensitive:      types.BoolValue(v.Sensitive),
		PolicySetID:    types.StringValue(v.PolicySet.ID),
	}

	// If the variable is sensitive, carry forward the last known value
	// instead, because the API never lets us read it again.
	if v.Sensitive {
		p.Value = lastValue
	}

	// Don't retrieve values if write-only is being used. Unset the value field before updating the state.
	isWriteOnlyValue := !valueWOVersion.IsNull()
	if isWriteOnlyValue {
		p.Value = types.StringValue("")
	}

	return p
}

func (r *resourceTFEPolicySetParameter) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Metadata implements resource.Resource
func (r *resourceTFEPolicySetParameter) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policy_set_parameter"
}

func (r *resourceTFEPolicySetParameter) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Creates, updates and destroys policy set parameters.",
		Version:     0,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Service-generated identifier for the parameter.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"key": schema.StringAttribute{
				Description: "Name of the parameter.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIf(
						func(ctx context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
							var stateSensitive types.Bool
							diags := req.State.GetAttribute(ctx, path.Root("sensitive"), &stateSensitive)
							if diags.HasError() {
								resp.Diagnostics.Append(diags...)
								return
							}
							if stateSensitive.ValueBool() && req.PlanValue.ValueString() != req.StateValue.ValueString() {
								resp.RequiresReplace = true
							}
						},
						"Force replacement if key changed and sensitive is true",
						"Force replacement if key changed and sensitive is true",
					),
				},
			},

			"value": schema.StringAttribute{
				Description: "Value of the parameter.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Sensitive:   true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("value_wo")),
				},
			},

			"value_wo": schema.StringAttribute{
				Optional:    true,
				WriteOnly:   true,
				Sensitive:   true,
				Description: "Value of the parameter in write-only mode",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("value")),
					stringvalidator.AlsoRequires(path.MatchRoot("value_wo_version")),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIf(
						func(ctx context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
							var stateVersion types.Int64
							diags := req.State.GetAttribute(ctx, path.Root("value_wo_version"), &stateVersion)
							resp.Diagnostics.Append(diags...)
							if resp.Diagnostics.HasError() {
								return
							}
							var planVersion types.Int64
							diags = req.Plan.GetAttribute(ctx, path.Root("value_wo_version"), &planVersion)
							resp.Diagnostics.Append(diags...)
							if resp.Diagnostics.HasError() {
								return
							}

							if !stateVersion.IsNull() && !planVersion.IsNull() && stateVersion.ValueInt64() != planVersion.ValueInt64() {
								resp.RequiresReplace = true
							}
						},
						"Force replacement if value_wo_version changed.",
						"Force replacement if value_wo_version changed.",
					),
				},
			},

			"value_wo_version": schema.Int64Attribute{
				Optional:    true,
				Description: "Version of the write-only value to trigger updates",
				Validators: []validator.Int64{
					int64validator.ConflictsWith(path.MatchRoot("value")),
					int64validator.AlsoRequires(path.MatchRoot("value_wo")),
				},
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},

			"sensitive": schema.BoolAttribute{
				Description: "Whether the value is sensitive. If true then the parameter is written once and not visible thereafter.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplaceIf(
						func(ctx context.Context, req planmodifier.BoolRequest, resp *boolplanmodifier.RequiresReplaceIfFuncResponse) {
							if req.StateValue.ValueBool() && !req.ConfigValue.ValueBool() {
								resp.RequiresReplace = true
							}
						},
						"Force replacement if sensitive argument changed from true to false.",
						"Force replacement if sensitive argument changed from true to false.",
					),
				},
			},

			"policy_set_id": schema.StringAttribute{
				Description: "The ID of the policy set that owns the parameter.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(
						path.MatchRelative().AtParent().AtName("policy_set_id"),
						// TODO: double-check behavior and ensure it includes current attr in that list
					),
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^polset-[a-zA-Z0-9]{16}$`),
						"must be a valid policy set ID (polset-<RANDOM STRING>)",
					),
				},
			},
		},
	}
}

func (r *resourceTFEPolicySetParameter) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read the Terraform plan and config into the model
	var plan, config modelTFEPolicySetParameter
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create an options struct
	options := tfe.PolicySetParameterCreateOptions{
		Key:       plan.Key.ValueStringPointer(),
		Category:  tfe.Category(tfe.CategoryPolicySet),
		Sensitive: plan.Sensitive.ValueBoolPointer(),
	}

	// Set Value from `value_wo` if set, otherwise use the normal value
	if !config.ValueWO.IsNull() {
		options.Value = config.ValueWO.ValueStringPointer()
	} else {
		options.Value = plan.Value.ValueStringPointer()
	}

	// Create the policy set parameter
	tflog.Debug(ctx, fmt.Sprintf("Create %s parameter: %s", tfe.CategoryPolicySet, plan.Key.ValueString()))
	p, err := r.config.Client.PolicySetParameters.Create(ctx, plan.PolicySetID.ValueString(), options)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Error creating %s parameter %s", tfe.CategoryPolicySet, plan.Key), err.Error())
		return
	}

	// We got a parameter, so set state to new values
	result := modelFromTFEPolicySetParameter(p, plan.Value, config.ValueWOVersion)
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *resourceTFEPolicySetParameter) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read the Terraform state into the model
	var state modelTFEPolicySetParameter
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check that policy set exists before continuing
	_, err := r.config.Client.PolicySets.Read(ctx, state.PolicySetID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Error retrieving policy set %s", state.PolicySetID), err.Error())
		return
	}

	// Read the policy set parameter
	tflog.Debug(ctx, fmt.Sprintf("Read parameter: %s", state.ID))
	p, err := r.config.Client.PolicySetParameters.Read(ctx, state.PolicySetID.ValueString(), state.ID.ValueString())
	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
			tflog.Debug(ctx, fmt.Sprintf("Parameter %s no longer exists", state.ID))
			resp.State.RemoveResource(ctx)
		}

		resp.Diagnostics.AddError(fmt.Sprintf("Error reading %s parameter %s", tfe.CategoryPolicySet, state.ID), err.Error())
		return
	}

	// We got a parameter, so update state:
	result := modelFromTFEPolicySetParameter(p, state.Value, state.ValueWOVersion)
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

// Update implements resource.Resource
func (r *resourceTFEPolicySetParameter) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read the Terraform plan, state, and config into the model
	var plan, state, config modelTFEPolicySetParameter
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create an options struct
	options := tfe.PolicySetParameterUpdateOptions{
		Key:       plan.Key.ValueStringPointer(),
		Sensitive: plan.Sensitive.ValueBoolPointer(),
	}

	options.Value = r.determineValueForUpdate(plan, state, config)

	// Update the policy set parameter
	tflog.Debug(ctx, fmt.Sprintf("Update parameter: %s", plan.ID.ValueString()))
	p, err := r.config.Client.PolicySetParameters.Update(ctx, plan.PolicySetID.ValueString(), plan.ID.ValueString(), options)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Error updating parameter %s", plan.ID), err.Error())
	}

	// Update state
	result := modelFromTFEPolicySetParameter(p, plan.Value, config.ValueWOVersion)
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

// Delete implements resource.Resource
func (r *resourceTFEPolicySetParameter) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Read the Terraform state into the model
	var state modelTFEPolicySetParameter
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check that policy set exists before continuing
	_, err := r.config.Client.PolicySets.Read(ctx, state.PolicySetID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Error retrieving policy set %s", state.PolicySetID), err.Error())
		return
	}

	// Delete the policy set parameter
	tflog.Debug(ctx, fmt.Sprintf("Delete parameter: %s", state.ID))
	err = r.config.Client.PolicySetParameters.Delete(ctx, state.PolicySetID.ValueString(), state.ID.ValueString())
	if err != nil && !errors.Is(err, tfe.ErrResourceNotFound) {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error deleting parameter %s", state.ID), err.Error())
	}
	// Resource is implicitly deleted from resp.State if diagnostics have no errors.
}

func (r *resourceTFEPolicySetParameter) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	s := strings.SplitN(req.ID, "/", 2)
	if len(s) != 2 {
		resp.Diagnostics.AddError(
			"Error importing variable",
			fmt.Sprintf("Invalid variable import format: %s (expected <POLICY SET ID>/<PARAMETER ID>)", req.ID),
		)
		return
	}

	policySetID := s[0]
	parameterID := s[1]

	data := modelTFEPolicySetParameter{
		ID:          types.StringValue(parameterID),
		PolicySetID: types.StringValue(policySetID),
	}

	diags := resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

// determineValueForUpdate returns what value to send to the API during an update,
// selecting from plan, state, or config based on four scenarios: switching between value/value_wo,
// version changes, or regular value changes. Returns nil if no value update is needed.
func (r *resourceTFEPolicySetParameter) determineValueForUpdate(plan, state, config modelTFEPolicySetParameter) *string {
	// Determine if we're using write-only value in plan vs state
	usingWriteOnlyInPlan := !plan.ValueWOVersion.IsNull()
	usingWriteOnlyInState := !state.ValueWOVersion.IsNull()

	// Case 1: Switching FROM value TO value_wo
	if !usingWriteOnlyInState && usingWriteOnlyInPlan && !config.ValueWO.IsNull() {
		return config.ValueWO.ValueStringPointer()
	}
	// Case 2: Switching FROM value_wo TO value
	if usingWriteOnlyInState && !usingWriteOnlyInPlan && !plan.Value.IsNull() {
		return plan.Value.ValueStringPointer()
	}
	// Case 3: value_wo version changed in plan
	if usingWriteOnlyInPlan && plan.ValueWOVersion.ValueInt64() != state.ValueWOVersion.ValueInt64() && !config.ValueWO.IsNull() {
		return config.ValueWO.ValueStringPointer()
	}
	// Case 4: Regular value changed. Only set Value if our planned value would be a CHANGE from
	// the prior state. This prevents accidentally resetting the value of sensitive variables on
	// unrelated changes when ignore_changes=[value] is set.
	if state.Value.ValueString() != plan.Value.ValueString() {
		return plan.Value.ValueStringPointer()
	}
	return nil
}
