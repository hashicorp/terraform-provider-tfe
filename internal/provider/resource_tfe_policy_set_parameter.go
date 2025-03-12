// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
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

func modelFromTFEPolicySetParameter(v *tfe.PolicySetParameter, lastValue types.String, isWriteOnlyValue bool) modelTFEPolicySetParameter {
	p := modelTFEPolicySetParameter{
		ID:          types.StringValue(v.ID),
		Key:         types.StringValue(v.Key),
		Value:       types.StringValue(v.Value),
		ValueWO:     types.StringValue(v.Value),
		Sensitive:   types.BoolValue(v.Sensitive),
		PolicySetID: types.StringValue(v.PolicySet.ID),
	}

	// If the parameter is sensitive, carry forward the last known value
	// instead, because the API never lets us read it again.
	if v.Sensitive {
		p.Value = lastValue
	}

	// Don't retrieve values if write-only is being used. Unset the value and readable_value fields before updating the state.
	if isWriteOnlyValue {
		p.Value = types.StringValue("")
	}

	return p
}

func isWriteOnlyValueInPrivateState(req resource.ReadRequest, resp *resource.ReadResponse) bool {
	storedValueWO, diags := req.Private.GetKey(ctx, "value_wo")
	resp.Diagnostics.Append(diags...)
	return len(storedValueWO) != 0
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
						r.requiresReplaceIfValueWOModifiedFunc,
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
					stringplanmodifier.RequiresReplaceIf(func(ctx context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
						storedValueWO, diags := req.Private.GetKey(ctx, "value_wo")
						resp.Diagnostics.Append(diags...)
						if resp.Diagnostics.HasError() {
							return
						}

						if !req.ConfigValue.IsNull() {
							hashedValue := getSHA256Hash(req.ConfigValue.ValueString())

							if string(storedValueWO) != hashedValue {
								log.Printf("[DEBUG] Replacing resource because the value of `value_wo` attribute has changed")
								resp.RequiresReplace = true
							}
						} else if len(storedValueWO) != 0 {
							// `value_wo` was previously set in the config, but is no longer
							resp.RequiresReplace = true
						}
					},
						"Force replacement if value_wo changed",
						"Force replacement if value_wo changed",
					),
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

func (r *resourceTFEPolicySetParameter) requiresReplaceIfValueWOModifiedFunc(ctx context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
	storedValueWO, diags := req.Private.GetKey(ctx, "value_wo")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(storedValueWO) != 0 {
		if req.ConfigValue.IsNull() {
			log.Printf("[DEBUG] Replacing resource because the value of `value_wo` attribute has been cleared")
			resp.RequiresReplace = true
			return
		}

		hashedValue := getSHA256Hash(req.ConfigValue.ValueString())
		if string(storedValueWO) != hashedValue {
			log.Printf("[DEBUG] Replacing resource because the value of `value_wo` attribute has changed")
			resp.RequiresReplace = true
		}
	} else if !req.ConfigValue.IsNull() {
		log.Printf("[DEBUG] Replacing resource because `value_wo` attribute has been added to a pre-existing policy set parameter resource")
		resp.RequiresReplace = true
	}
}

func (r *resourceTFEPolicySetParameter) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read the plan into the model
	var plan modelTFEPolicySetParameter
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read the config into the model
	var config modelTFEPolicySetParameter
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

	if !config.ValueWO.IsNull() {
		options.Value = config.ValueWO.ValueStringPointer()
	} else {
		options.Value = plan.Value.ValueStringPointer()
	}

	// Create the policy set parameter
	log.Printf("[DEBUG] Create %s parameter: %s", tfe.CategoryPolicySet, plan.Key.ValueString())
	p, err := r.config.Client.PolicySetParameters.Create(ctx, plan.PolicySetID.ValueString(), options)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Error creating %s parameter %s", tfe.CategoryPolicySet, plan.Key), err.Error())
		return
	}

	result := modelFromTFEPolicySetParameter(p, plan.Value, !config.ValueWO.IsNull())

	if !config.ValueWO.IsNull() {
		// Use the resource's private state to store secure hashes of write-only argument values, the provider during planmodify will use the hash to determine if a write-only argument value has changed in later Terraform runs.
		diags := resp.Private.SetKey(ctx, "value_wo", fmt.Appendf(nil, `"%s"`, config.ValueWO.ValueString()))
		resp.Diagnostics.Append(diags...)
	} else {
		// if the value is not configured as write-only, then remove valueWO key from private state. Setting a key with an empty byte slice is interpreted by the framework as a request to remove the key from the ProviderData map.
		diags := resp.Private.SetKey(ctx, "value_wo", []byte(""))
		resp.Diagnostics.Append(diags...)
	}

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
	log.Printf("[DEBUG] Read parameter: %s", state.ID)
	p, err := r.config.Client.PolicySetParameters.Read(ctx, state.PolicySetID.ValueString(), state.ID.ValueString())
	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
			log.Printf("[DEBUG] Parameter %s no longer exists", state.ID)
			resp.State.RemoveResource(ctx)
		}

		resp.Diagnostics.AddError(fmt.Sprintf("Error reading %s parameter %s", tfe.CategoryPolicySet, state.ID), err.Error())
		return
	}

	isWriteOnlyValue := isWriteOnlyValueInPrivateState(req, resp) // to avoid reading from written-only values
	if resp.Diagnostics.HasError() {
		return
	}

	// update state
	result := modelFromTFEPolicySetParameter(p, state.Value, isWriteOnlyValue)
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

// Update implements resource.Resource
func (r *resourceTFEPolicySetParameter) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read the plan into the model
	var plan modelTFEPolicySetParameter
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read the state into the model
	var state modelTFEPolicySetParameter
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read the config into the model
	var config modelTFEPolicySetParameter
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
	// This is so we don't accidentally reset the value of a sensitive parameter on
	// unrelated changes when `ignore_changes = [value]` is set.
	if state.Value.ValueString() != plan.Value.ValueString() {
		options.Value = plan.Value.ValueStringPointer()
	}

	// Update the policy set parameter
	log.Printf("[DEBUG] Update parameter: %s", plan.ID.ValueString())
	p, err := r.config.Client.PolicySetParameters.Update(ctx, plan.PolicySetID.ValueString(), plan.ID.ValueString(), options)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Error updating parameter %s", plan.ID), err.Error())
	}

	result := modelFromTFEPolicySetParameter(p, plan.Value, !config.ValueWO.IsNull())
	r.updatePrivateState(ctx, resp, config.ValueWO)
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *resourceTFEPolicySetParameter) updatePrivateState(ctx context.Context, resp *resource.UpdateResponse, configValueWO types.String) {
	if !configValueWO.IsNull() {
		// Use the resource's private state to store secure hashes of write-only argument values, planModify will use the hash to determine if a write-only argument value has changed in later Terraform runs.
		hashedValue := getSHA256Hash(configValueWO.ValueString())
		diags := resp.Private.SetKey(ctx, "value_wo", fmt.Appendf(nil, `"%s"`, hashedValue))
		resp.Diagnostics.Append(diags...)
	} else {
		// if value is not configured as write-only, remove valueWO key from private state
		diags := resp.Private.SetKey(ctx, "value_wo", []byte(""))
		resp.Diagnostics.Append(diags...)
	}
}

func getSHA256Hash(data string) string {
	hasher := sha256.New()
	hasher.Write([]byte(data))
	return hex.EncodeToString(hasher.Sum(nil))
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
	log.Printf("[DEBUG] Delete parameter: %s", state.ID)
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
			"Error importing policy set parameter",
			fmt.Sprintf("Invalid policy set parameter import format: %s (expected <POLICY SET ID>/<PARAMETER ID>)", req.ID),
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
