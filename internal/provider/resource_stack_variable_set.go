// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const minTFEVersionVariableSetStacks = "1.0.0"

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

func (r *resourceStackVariableSet) checkStackVariableSetSupport(diagnostics *diag.Diagnostics) bool {
	meetsMinVersionRequirement, err := r.config.MeetsMinRemoteTFEVersion(minTFEVersionVariableSetStacks)
	if err != nil {
		diagnostics.AddError(
			"Error checking TFE version",
			fmt.Sprintf("Could not determine if Terraform Enterprise version %s meets minimum required version %s: %v",
				r.config.RemoteTFEVersion(), minTFEVersionVariableSetStacks, err),
		)
		return false
	}
	if !meetsMinVersionRequirement {
		diagnostics.AddError(
			"Feature not supported",
			fmt.Sprintf("Associating variable sets with stacks requires Terraform Enterprise version %s or later. Current version: %s",
				minTFEVersionVariableSetStacks, r.config.RemoteTFEVersion()),
		)
		return false
	}
	return true
}

func (r *resourceStackVariableSet) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan stackVariableSetResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !r.checkStackVariableSetSupport(&resp.Diagnostics) {
		return
	}

	variableSetID := plan.VariableSetID.ValueString()
	stackID := plan.StackID.ValueString()

	applyOptions := tfe.VariableSetApplyToStacksOptions{}
	applyOptions.Stacks = append(applyOptions.Stacks, &tfe.Stack{ID: stackID})

	if err := r.config.Client.VariableSets.ApplyToStacks(ctx, variableSetID, &applyOptions); err != nil {
		resp.Diagnostics.AddError(
			"Error applying variable set to stack",
			fmt.Sprintf("Error applying variable set id %s to stack %s: %s", variableSetID, stackID, err),
		)
		return
	}

	// Set the ID using stack/varset format
	plan.ID = types.StringValue(fmt.Sprintf("%s/%s", stackID, variableSetID))
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceStackVariableSet) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state stackVariableSetResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !r.checkStackVariableSetSupport(&resp.Diagnostics) {
		return
	}

	stackID := state.StackID.ValueString()
	variableSetID := state.VariableSetID.ValueString()

	tflog.Debug(ctx, "Reading stack variable set attachment", map[string]interface{}{
		"stack_id":        stackID,
		"variable_set_id": variableSetID,
	})

	vS, err := r.config.Client.VariableSets.Read(ctx, variableSetID, &tfe.VariableSetReadOptions{
		Include: &[]tfe.VariableSetIncludeOpt{tfe.VariableSetStacks},
	})
	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
			tflog.Debug(ctx, "Variable set not found, removing from state", map[string]interface{}{
				"variable_set_id": variableSetID,
			})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading variable set",
			fmt.Sprintf("Error reading variable set %s: %s", variableSetID, err),
		)
		return
	}

	// Verify stack is still attached
	found := false
	for _, stack := range vS.Stacks {
		if stack.ID == stackID {
			found = true
			break
		}
	}

	if !found {
		tflog.Debug(ctx, "Stack not attached to variable set, removing from state", map[string]interface{}{
			"stack_id":        stackID,
			"variable_set_id": variableSetID,
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

	if !r.checkStackVariableSetSupport(&resp.Diagnostics) {
		return
	}

	stackID := state.StackID.ValueString()
	variableSetID := state.VariableSetID.ValueString()

	tflog.Debug(ctx, "Deleting stack variable set attachment", map[string]interface{}{
		"stack_id":        stackID,
		"variable_set_id": variableSetID,
	})

	removeOptions := tfe.VariableSetRemoveFromStacksOptions{}
	removeOptions.Stacks = append(removeOptions.Stacks, &tfe.Stack{ID: stackID})

	err := r.config.Client.VariableSets.RemoveFromStacks(ctx, variableSetID, &removeOptions)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error removing stack from variable set",
			fmt.Sprintf("Error removing stack %s from variable set %s: %s", stackID, variableSetID, err),
		)
		return
	}
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
	// Example: st-xxx/varset-yyy
	id := req.ID
	parts := strings.Split(id, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid import ID format",
			fmt.Sprintf("Expected format: stack/varset (e.g., st-xxx/varset-yyy), got: %s", id),
		)
		return
	}

	stackID := parts[0]
	variableSetID := parts[1]

	if !isResourceIDFormat("st", stackID) || !isResourceIDFormat("varset", variableSetID) {
		resp.Diagnostics.AddError(
			"Invalid import ID format",
			fmt.Sprintf("Import ID must be in format: stack/varset (e.g., st-xxx/varset-yyy), got: %s", id),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("stack_id"), stackID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("variable_set_id"), variableSetID)...)
}
