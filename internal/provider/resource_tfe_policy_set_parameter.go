// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-tfe/internal/provider/helpers"
	"github.com/hashicorp/terraform-provider-tfe/internal/provider/planmodifiers"
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
	ID          types.String `tfsdk:"id"`
	Key         types.String `tfsdk:"key"`
	Value       types.String `tfsdk:"value"`
	ValueWO     types.String `tfsdk:"value_wo"`
	Sensitive   types.Bool   `tfsdk:"sensitive"`
	PolicySetID types.String `tfsdk:"policy_set_id"`
}

func modelFromTFEPolicySetParameter(v *tfe.PolicySetParameter, lastValue types.String, isWriteOnly bool) modelTFEPolicySetParameter {
	p := modelTFEPolicySetParameter{
		ID:          types.StringValue(v.ID),
		Key:         types.StringValue(v.Key),
		Value:       types.StringValue(v.Value),
		Sensitive:   types.BoolValue(v.Sensitive),
		PolicySetID: types.StringValue(v.PolicySet.ID),
	}

	// If the variable is sensitive, carry forward the last known value
	// instead, because the API never lets us read it again.
	if v.Sensitive {
		p.Value = lastValue
	}

	// If the variable is write-only, clear the value.
	if isWriteOnly {
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
		Version:     1,
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
				},
				PlanModifiers: []planmodifier.String{
					planmodifiers.NewReplaceForWriteOnlyStringValue("value_wo"),
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
	isWriteOnly := !config.ValueWO.IsNull()
	if isWriteOnly {
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

	// Store the hashed write-only value in the private state
	store := r.writeOnlyValueStore(resp.Private)
	resp.Diagnostics.Append(store.SetPriorValue(ctx, config.ValueWO)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update state
	result := modelFromTFEPolicySetParameter(p, plan.Value, isWriteOnly)
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

	// Check if the parameter is write-only
	isWriteOnly, diags := r.writeOnlyValueStore(resp.Private).PriorValueExists(ctx)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	result := modelFromTFEPolicySetParameter(p, state.Value, isWriteOnly)
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

	// Only set Value if our planned value would be a change from the prior state.
	// This is so we don't accidentally reset the value of a sensitive variable on
	// unrelated changes when `ignore_changes = [value]` is set.
	if state.Value.ValueString() != plan.Value.ValueString() {
		options.Value = plan.Value.ValueStringPointer()
	}

	// Update the policy set parameter
	tflog.Debug(ctx, fmt.Sprintf("Update parameter: %s", plan.ID.ValueString()))
	p, err := r.config.Client.PolicySetParameters.Update(ctx, plan.PolicySetID.ValueString(), plan.ID.ValueString(), options)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Error updating parameter %s", plan.ID), err.Error())
	}

	// Store the hashed write-only value in the private state
	store := r.writeOnlyValueStore(resp.Private)
	resp.Diagnostics.Append(store.SetPriorValue(ctx, config.ValueWO)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update state
	result := modelFromTFEPolicySetParameter(p, plan.Value, !config.ValueWO.IsNull())
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

func (r *resourceTFEPolicySetParameter) writeOnlyValueStore(private helpers.PrivateState) *helpers.WriteOnlyValueStore {
	return helpers.NewWriteOnlyValueStore(private, "value_wo")
}
