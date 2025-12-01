// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &resourceStackVariableSet{}
var _ resource.ResourceWithConfigure = &resourceStackVariableSet{}
var _ resource.ResourceWithImportState = &resourceStackVariableSet{}

func NewStackVariableSetResource() resource.Resource {
	return &resourceStackVariableSet{}
}

// resourceStackVariableSet implements the tfe_stack_variable_set resource type
type resourceStackVariableSet struct {
	config ConfiguredClient
}

// Model describes the resource data model.
type stackVariableSetResourceModel struct {
	ID            types.String `tfsdk:"id"`
	StackID       types.String `tfsdk:"stack_id"`
	VariableSetID types.String `tfsdk:"variable_set_id"`
}

func (r *resourceStackVariableSet) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stack_variable_set"
}

func (r *resourceStackVariableSet) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages associations between variable sets and stacks.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the stack variable set association.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"variable_set_id": schema.StringAttribute{
				Description: "The ID of the variable set.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(), // Follows tfe_workspace_variable_set and tfe_project_variable_set pattern
				},
			},
			"stack_id": schema.StringAttribute{
				Description: "The ID of the stack.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(), // Follows tfe_workspace_variable_set and tfe_project_variable_set pattern
				},
			},
		},
	}
}

func (r *resourceStackVariableSet) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(ConfiguredClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected ConfiguredClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.config = client
}

func (r *resourceStackVariableSet) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan stackVariableSetResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vSID := plan.VariableSetID.ValueString()
	stkID := plan.StackID.ValueString()

	applyOptions := tfe.VariableSetApplyToStacksOptions{}
	applyOptions.Stacks = append(applyOptions.Stacks, &tfe.Stack{ID: stkID})

	err := r.config.Client.VariableSets.ApplyToStacks(ctx, vSID, &applyOptions)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error applying variable set to stack",
			fmt.Sprintf("Error applying variable set id %s to stack %s: %s", vSID, stkID, err),
		)
		return
	}

	// Set the ID using stack/varset format
	plan.ID = types.StringValue(fmt.Sprintf("%s/%s", stkID, vSID))

	tflog.Trace(ctx, "Created stack variable set attachment", map[string]interface{}{
		"stack_id":        stkID,
		"variable_set_id": vSID,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceStackVariableSet) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state stackVariableSetResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	stkID := state.StackID.ValueString()
	vSID := state.VariableSetID.ValueString()

	tflog.Debug(ctx, "Reading stack variable set attachment", map[string]interface{}{
		"stack_id":        stkID,
		"variable_set_id": vSID,
	})

	vS, err := r.config.Client.VariableSets.Read(ctx, vSID, &tfe.VariableSetReadOptions{
		Include: &[]tfe.VariableSetIncludeOpt{tfe.VariableSetStacks},
	})
	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
			tflog.Debug(ctx, "Variable set not found, removing from state", map[string]interface{}{
				"variable_set_id": vSID,
			})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading variable set",
			fmt.Sprintf("Error reading variable set %s: %s", vSID, err),
		)
		return
	}

	// Verify stack is still attached
	found := false
	for _, stack := range vS.Stacks {
		if stack.ID == stkID {
			found = true
			break
		}
	}

	if !found {
		tflog.Debug(ctx, "Stack not attached to variable set, removing from state", map[string]interface{}{
			"stack_id":        stkID,
			"variable_set_id": vSID,
		})
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *resourceStackVariableSet) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state stackVariableSetResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	stkID := state.StackID.ValueString()
	vSID := state.VariableSetID.ValueString()

	tflog.Debug(ctx, "Deleting stack variable set attachment", map[string]interface{}{
		"stack_id":        stkID,
		"variable_set_id": vSID,
	})

	removeOptions := tfe.VariableSetRemoveFromStacksOptions{}
	removeOptions.Stacks = append(removeOptions.Stacks, &tfe.Stack{ID: stkID})

	err := r.config.Client.VariableSets.RemoveFromStacks(ctx, vSID, &removeOptions)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error removing stack from variable set",
			fmt.Sprintf("Error removing stack %s from variable set %s: %s", stkID, vSID, err),
		)
		return
	}

	tflog.Trace(ctx, "Deleted stack variable set attachment")
}

func (r *resourceStackVariableSet) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Stack and variable set IDs cannot be updated (they are marked as RequiresReplace)
	// This method is a no-op but required by the framework
	var plan stackVariableSetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceStackVariableSet) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Parse the import ID in format: stack_id_variable_set_id
	// Example: stk-xxx_varset-yyy
	id := req.ID
	parts := strings.Split(id, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid import ID format",
			fmt.Sprintf("Expected format: stack/varset (e.g., stk-xxx/varset-yyy), got: %s", id),
		)
		return
	}

	stackID := parts[0]
	variableSetID := parts[1]

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("stack_id"), stackID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("variable_set_id"), variableSetID)...)
}
