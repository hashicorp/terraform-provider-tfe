// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	tfev2 "github.com/hashicorp/go-tfe/v2"
	"github.com/hashicorp/go-tfe/v2/api/models"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &resourceStackPolicySetExclusion{}
	_ resource.ResourceWithConfigure   = &resourceStackPolicySetExclusion{}
	_ resource.ResourceWithImportState = &resourceStackPolicySetExclusion{}
)

type resourceStackPolicySetExclusion struct {
	config ConfiguredClient
}

type modelStackPolicySetExclusion struct {
	ID          types.String `tfsdk:"id"`
	StackID     types.String `tfsdk:"stack_id"`
	PolicySetID types.String `tfsdk:"policy_set_id"`
}

func NewStackPolicySetExclusionResource() resource.Resource {
	return &resourceStackPolicySetExclusion{}
}

func (r *resourceStackPolicySetExclusion) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stack_policy_set_exclusion"
}

func (r *resourceStackPolicySetExclusion) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Adds and removes Stack exclusions on a Terraform policy (`tfpolicy`) policy set. " +
			"Stack exclusions exempt the specified Stack from policy set enforcement, typically to carve it out of a global policy set. " +
			"On Terraform Enterprise, requires v" + minTFEVersionStackPolicySets + " or later." +
			"\n\n~> **Note:** A Stack cannot be both included in and excluded from the same policy set." +
			"\n\n~> **Note:** Tag-based scoping and explicit workspace/project/Stack associations are mutually exclusive on a policy set. " +
			"To switch between them, first remove the existing association (`terraform apply`), then add the new one (`terraform apply`). " +
			"Attempting both in a single apply may fail.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The composite ID of the Stack exclusion, in the format `<STACK_ID>/<POLICY_SET_ID>`.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"stack_id": schema.StringAttribute{
				Description: "The ID of the Stack to exclude from the policy set.",
				Required:    true,
				Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"policy_set_id": schema.StringAttribute{
				Description: "The ID of the policy set to which to add the Stack exclusion. Must have `kind = \"tfpolicy\"`.",
				Required:    true,
				Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *resourceStackPolicySetExclusion) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(ConfiguredClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type",
			fmt.Sprintf("Expected ConfiguredClient, got %T. This is a bug in the tfe provider, so please report it on GitHub.", req.ProviderData))
		return
	}
	if client.ClientV2 == nil {
		resp.Diagnostics.AddError("Missing API Client", "The tfe_stack_policy_set_exclusion resource requires the v2 API client.")
		return
	}
	r.config = client
}

func (r *resourceStackPolicySetExclusion) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan modelStackPolicySetExclusion
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	checkStackPolicySetsSupported(r.config, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	stackID := plan.StackID.ValueString()
	policySetID := plan.PolicySetID.ValueString()

	if err := checkStackExists(ctx, r.config, stackID); err != nil {
		if errors.Is(err, tfev2.ErrNotFound) {
			resp.Diagnostics.AddError("Stack Not Found",
				fmt.Sprintf("Stack %q was not found. Verify the stack ID and that the stack exists.", stackID))
			return
		}
		resp.Diagnostics.AddError("Error Reading Stack", fmt.Sprintf("Stack %q: %s", stackID, err.Error()))
		return
	}

	// The API returns success without creating the exclusion when the stack
	// is already included, so reject that case before writing.
	ps, err := readPolicySetWithStacks(ctx, r.config, policySetID)
	if err == nil {
		var included bool
		included, err = policySetHasStack(ps, stackID, false)
		if err == nil && included {
			err = fmt.Errorf("stack %q is included in this policy set; remove the inclusion first", stackID)
		}
	}
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Stack Policy Set Exclusion",
			fmt.Sprintf("Policy set %q, stack %q: %s", policySetID, stackID, err.Error()))
		return
	}

	if err := r.writeStackExclusionRelationship(ctx, policySetID, stackID, false); err != nil {
		resp.Diagnostics.AddError("Error Creating Stack Policy Set Exclusion",
			fmt.Sprintf("Policy set %q, stack %q: %s", policySetID, stackID, err.Error()))
		return
	}
	plan.ID = types.StringValue(stackID + "/" + policySetID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceStackPolicySetExclusion) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state modelStackPolicySetExclusion
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	stackID := state.StackID.ValueString()
	policySetID := state.PolicySetID.ValueString()

	ps, err := readPolicySetWithStacks(ctx, r.config, policySetID)
	if errors.Is(err, tfev2.ErrNotFound) {
		tflog.Debug(ctx, fmt.Sprintf("Policy set %s no longer exists", policySetID))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Policy Set",
			fmt.Sprintf("Policy set %q: %s", policySetID, err.Error()))
		return
	}
	excluded, err := policySetHasStack(ps, stackID, true)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Policy Set Relationship Response",
			fmt.Sprintf("Policy set %q, stack %q: %s", policySetID, stackID, err))
		return
	}
	if !excluded {
		tflog.Debug(ctx, fmt.Sprintf("Stack %s is no longer excluded from policy set %s", stackID, policySetID))
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update is required by the framework but never changes anything, because
// every configurable attribute requires replacement.
func (r *resourceStackPolicySetExclusion) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan modelStackPolicySetExclusion
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceStackPolicySetExclusion) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state modelStackPolicySetExclusion
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	stackID := state.StackID.ValueString()
	policySetID := state.PolicySetID.ValueString()

	err := r.writeStackExclusionRelationship(ctx, policySetID, stackID, true)
	if errors.Is(err, tfev2.ErrNotFound) {
		tflog.Debug(ctx, fmt.Sprintf("Policy set %s no longer exists", policySetID))
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Stack Policy Set Exclusion",
			fmt.Sprintf("Policy set %q, stack %q: %s", policySetID, stackID, err.Error()))
	}
}

// ImportState only parses the ID; the Read that Terraform runs after import
// verifies the exclusion exists.
func (r *resourceStackPolicySetExclusion) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
		resp.Diagnostics.AddError("Invalid Import ID Format",
			fmt.Sprintf("The import ID must be in the format <stack_id>/<policy_set_id>. Got: %q", req.ID))
		return
	}
	checkStackPolicySetsSupported(r.config, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("stack_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("policy_set_id"), parts[1])...)
}

func (r *resourceStackPolicySetExclusion) writeStackExclusionRelationship(ctx context.Context, policySetID, stackID string, remove bool) error {
	item := models.NewStacksIdentifierArrayDocument_data()
	item.SetId(&stackID)
	kind := models.STACKS_STACKSIDENTIFIERARRAYDOCUMENT_DATA_TYPE
	item.SetTypeEscaped(&kind)
	body := models.NewStacksIdentifierArrayDocument()
	body.SetData([]models.StacksIdentifierArrayDocument_dataable{item})

	exclusions := r.config.ClientV2.API.PolicySets().ByPolicy_set_id(policySetID).Relationships().StackExclusions()
	if remove {
		return exclusions.Delete(ctx, body, nil)
	}
	return exclusions.Post(ctx, body, nil)
}
