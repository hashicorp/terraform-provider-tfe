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
	"github.com/hashicorp/go-tfe/v2/api/policysets"
	policysetitem "github.com/hashicorp/go-tfe/v2/api/policysets/item"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	abstractions "github.com/microsoft/kiota-abstractions-go"
)

var (
	_ resource.Resource                = &resourceStackPolicySet{}
	_ resource.ResourceWithConfigure   = &resourceStackPolicySet{}
	_ resource.ResourceWithImportState = &resourceStackPolicySet{}
)

type resourceStackPolicySet struct {
	config ConfiguredClient
}

type modelStackPolicySet struct {
	ID          types.String `tfsdk:"id"`
	StackID     types.String `tfsdk:"stack_id"`
	PolicySetID types.String `tfsdk:"policy_set_id"`
}

func NewStackPolicySetResource() resource.Resource {
	return &resourceStackPolicySet{}
}

func (r *resourceStackPolicySet) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stack_policy_set"
}

func (r *resourceStackPolicySet) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an explicit stack association with a tfpolicy policy set. " +
			"The policy set must not be global. A stack cannot be both included and excluded from the same policy set. " +
			"Tag-based scoping and explicit stack associations are mutually exclusive; remove existing tag scope in a separate apply before adding this association. " +
			"Import using `<stack_id>/<policy_set_id>`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Computed synthetic ID in the format `<stack_id>/<policy_set_id>`.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"stack_id": schema.StringAttribute{
				Description: "ID of the stack.",
				Required:    true,
				Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"policy_set_id": schema.StringAttribute{
				Description: "ID of the tfpolicy policy set.",
				Required:    true,
				Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *resourceStackPolicySet) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
		resp.Diagnostics.AddError("Missing API Client", "The tfe_stack_policy_set resource requires the v2 API client.")
		return
	}
	r.config = client
}

func (r *resourceStackPolicySet) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan modelStackPolicySet
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	stackID := plan.StackID.ValueString()
	policySetID := plan.PolicySetID.ValueString()

	if err := r.checkStack(ctx, stackID); err != nil {
		if errors.Is(err, tfev2.ErrNotFound) {
			resp.Diagnostics.AddError("Stack Not Found",
				fmt.Sprintf("Stack %q was not found. Verify the stack ID and that the stack exists.", stackID))
			return
		}
		resp.Diagnostics.AddError("Error Reading Stack", fmt.Sprintf("Stack %q: %s", stackID, apiErrorDetail(err)))
		return
	}

	ps, err := r.readPolicySet(ctx, policySetID)
	if err == nil {
		var excluded bool
		excluded, err = policySetHasStack(ps, stackID, true)
		if err == nil && excluded {
			err = fmt.Errorf("stack %q is excluded from this policy set; remove the exclusion first", stackID)
		}
	}
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Stack Policy Set Association",
			fmt.Sprintf("Policy set %q, stack %q: %s", policySetID, stackID, apiErrorDetail(err)))
		return
	}

	if err := r.writeStackRelationship(ctx, policySetID, stackID, false); err != nil {
		resp.Diagnostics.AddError("Error Creating Stack Policy Set Association",
			fmt.Sprintf("Policy set %q, stack %q: %s", policySetID, stackID, apiErrorDetail(err)))
		return
	}
	plan.ID = types.StringValue(stackID + "/" + policySetID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceStackPolicySet) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state modelStackPolicySet
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	stackID := state.StackID.ValueString()
	policySetID := state.PolicySetID.ValueString()

	ps, err := r.readPolicySet(ctx, policySetID)
	if errors.Is(err, tfev2.ErrNotFound) {
		tflog.Debug(ctx, fmt.Sprintf("Policy set %s no longer exists", policySetID))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Policy Set",
			fmt.Sprintf("Policy set %q: %s", policySetID, apiErrorDetail(err)))
		return
	}
	present, err := policySetHasStack(ps, stackID, false)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Policy Set Relationship Response",
			fmt.Sprintf("Policy set %q, stack %q: %s", policySetID, stackID, err))
		return
	}
	if !present {
		tflog.Debug(ctx, fmt.Sprintf("Stack %s is no longer attached to policy set %s", stackID, policySetID))
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update is required by the framework but never changes anything, because
// every configurable attribute requires replacement.
func (r *resourceStackPolicySet) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan modelStackPolicySet
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceStackPolicySet) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state modelStackPolicySet
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	stackID := state.StackID.ValueString()
	policySetID := state.PolicySetID.ValueString()

	err := r.writeStackRelationship(ctx, policySetID, stackID, true)
	if errors.Is(err, tfev2.ErrNotFound) {
		tflog.Debug(ctx, fmt.Sprintf("Policy set %s no longer exists", policySetID))
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Stack Policy Set Association",
			fmt.Sprintf("Policy set %q, stack %q: %s", policySetID, stackID, apiErrorDetail(err)))
	}
}

func (r *resourceStackPolicySet) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
		resp.Diagnostics.AddError("Invalid Import ID Format",
			fmt.Sprintf("The import ID must be in the format <stack_id>/<policy_set_id>. Got: %q", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("stack_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("policy_set_id"), parts[1])...)
}

func (r *resourceStackPolicySet) checkStack(ctx context.Context, id string) error {
	envelope, err := r.config.ClientV2.API.Stacks().ByStack_id(id).Get(ctx, nil)
	if err != nil {
		return fmt.Errorf("reading stack %q: %w", id, err)
	}
	if envelope == nil || envelope.GetData() == nil {
		return fmt.Errorf("API returned no data for stack %q", id)
	}
	return nil
}

// readPolicySet includes both stack relationships so the inclusion can be
// checked against existing exclusions.
func (r *resourceStackPolicySet) readPolicySet(ctx context.Context, id string) (models.PolicySetsable, error) {
	envelope, err := r.config.ClientV2.API.PolicySets().ByPolicy_set_id(id).Get(ctx,
		&abstractions.RequestConfiguration[policysets.WithPolicy_set_ItemRequestBuilderGetQueryParameters]{
			QueryParameters: &policysets.WithPolicy_set_ItemRequestBuilderGetQueryParameters{
				Include: []policysetitem.GetIncludeQueryParameterType{
					policysetitem.STACKS_GETINCLUDEQUERYPARAMETERTYPE,
					policysetitem.STACK_EXCLUSIONS_GETINCLUDEQUERYPARAMETERTYPE,
				},
			},
		})
	if err != nil {
		return nil, fmt.Errorf("reading policy set %q: %w", id, err)
	}
	if envelope == nil || envelope.GetData() == nil {
		return nil, fmt.Errorf("API returned no data for policy set %q", id)
	}
	return envelope.GetData(), nil
}

func (r *resourceStackPolicySet) writeStackRelationship(ctx context.Context, policySetID, stackID string, remove bool) error {
	item := models.NewStacksIdentifierArrayDocument_data()
	item.SetId(&stackID)
	kind := models.STACKS_STACKSIDENTIFIERARRAYDOCUMENT_DATA_TYPE
	item.SetTypeEscaped(&kind)
	body := models.NewStacksIdentifierArrayDocument()
	body.SetData([]models.StacksIdentifierArrayDocument_dataable{item})

	stacks := r.config.ClientV2.API.PolicySets().ByPolicy_set_id(policySetID).Relationships().Stacks()
	if remove {
		return stacks.Delete(ctx, body, nil)
	}
	return stacks.Post(ctx, body, nil)
}

// policySetHasStack reports whether stackID is in the policy set's "stacks"
// relationship, or its "stack-exclusions" relationship when exclusion is true.
func policySetHasStack(ps models.PolicySetsable, stackID string, exclusion bool) (bool, error) {
	// Global policy sets omit explicit inclusions, as they do for workspaces and projects.
	if attrs := ps.GetAttributes(); !exclusion && attrs != nil && attrs.GetGlobal() != nil && *attrs.GetGlobal() {
		return false, nil
	}
	if ps.GetRelationships() == nil {
		return false, fmt.Errorf("policy set response is missing relationships")
	}
	name := "stacks"
	relationship := ps.GetRelationships().GetStacks()
	if exclusion {
		name = "stack-exclusions"
		relationship = ps.GetRelationships().GetStackExclusions()
	}
	if relationship == nil || relationship.GetData() == nil {
		return false, fmt.Errorf("policy set response is missing %s relationship data", name)
	}
	for _, stack := range relationship.GetData() {
		if stack == nil || stack.GetId() == nil || *stack.GetId() == "" {
			return false, fmt.Errorf("policy set %s relationship contains a stack without an ID", name)
		}
		if *stack.GetId() == stackID {
			return true, nil
		}
	}
	return false, nil
}
