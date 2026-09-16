// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &resourceTFEWorkspaceHYOKEnabled{}
	_ resource.ResourceWithConfigure   = &resourceTFEWorkspaceHYOKEnabled{}
	_ resource.ResourceWithImportState = &resourceTFEWorkspaceHYOKEnabled{}
)

type resourceTFEWorkspaceHYOKEnabled struct {
	config ConfiguredClient
}

func (r *resourceTFEWorkspaceHYOKEnabled) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *resourceTFEWorkspaceHYOKEnabled) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	workspaceID := req.ID

	ws, err := r.config.Client.Workspaces.ReadByID(ctx, workspaceID)
	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
			resp.Diagnostics.AddError(
				"Workspace not found",
				fmt.Sprintf("No workspace with ID %s exists.", workspaceID),
			)
			return
		}
		resp.Diagnostics.AddError(fmt.Sprintf("Error reading workspace %s", workspaceID), err.Error())
		return
	}

	if ws.HYOKEnabled == nil || !*ws.HYOKEnabled {
		resp.Diagnostics.AddError(
			"HYOK is not enabled on this workspace",
			fmt.Sprintf("Workspace %s does not have HYOK enabled. This resource can only be imported for workspaces that already have HYOK enabled.", workspaceID),
		)
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Importing tfe_workspace_hyok_enabled for workspace %s", workspaceID))
	resource.ImportStatePassthroughID(ctx, path.Root("workspace_id"), req, resp)
}

func (r *resourceTFEWorkspaceHYOKEnabled) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workspace_hyok_enabled"
}
func (r *resourceTFEWorkspaceHYOKEnabled) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Enables HYOK (Hold Your Own Key) encryption on a workspace.\n\n~> **Note:** HYOK enablement is **irreversible**. Once enabled on a workspace, it cannot be disabled. Destroying this resource removes it from Terraform state but does **not** disable HYOK on the workspace. This resource requires HCP Terraform Premium. Refer to [HCP Terraform pricing](https://www.hashicorp.com/en/pricing?product_intent=terraform&tab=terraform) for details.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID of the workspace on which HYOK has been enabled",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"workspace_id": schema.StringAttribute{
				Description: "ID of the workspace in which to enable HYOK",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}
func NewWorkspaceHYOKEnabledResource() resource.Resource {
	return &resourceTFEWorkspaceHYOKEnabled{}
}

type modelTFEWorkspaceHYOKEnabled struct {
	ID          types.String `tfsdk:"id"`
	WorkspaceID types.String `tfsdk:"workspace_id"`
}

func (r *resourceTFEWorkspaceHYOKEnabled) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan modelTFEWorkspaceHYOKEnabled
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	workspaceID := plan.WorkspaceID.ValueString()
	plan.ID = plan.WorkspaceID

	ws, err := r.config.Client.Workspaces.ReadByID(ctx, workspaceID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading workspace", err.Error())
		return
	}

	if ws.HYOKEnabled != nil && *ws.HYOKEnabled {
		tflog.Debug(ctx, fmt.Sprintf("HYOK already enabled on workspace %s", workspaceID))
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		return
	}

	options := tfe.WorkspaceUpdateOptions{
		HYOKEnabled: tfe.Bool(true),
	}

	tflog.Debug(ctx, "Enabling HYOK")
	ws, err = r.config.Client.Workspaces.UpdateByID(ctx, workspaceID, options)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Error Enabling HYOK on workspace %s, ", workspaceID), err.Error())
		return
	}

	if ws.HYOKEnabled == nil || !*ws.HYOKEnabled {
		resp.Diagnostics.AddError(fmt.Sprintf("Error Enabling HYOK on workspace %s.", workspaceID), "API accepted request, but HYOK is still false for workspace")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceTFEWorkspaceHYOKEnabled) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state modelTFEWorkspaceHYOKEnabled
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	workspaceID := state.WorkspaceID.ValueString()
	ws, err := r.config.Client.Workspaces.ReadByID(ctx, workspaceID)
	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
			tflog.Debug(ctx, fmt.Sprintf("Workspace %s no longer exists", workspaceID))
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(fmt.Sprintf("Error reading workspace %s", workspaceID), err.Error())
		return
	}

	if ws.HYOKEnabled == nil || !*ws.HYOKEnabled {
		tflog.Debug(ctx, fmt.Sprintf("HYOK is not enabled on workspace %s; removing from state so it will be re-enabled on next apply", workspaceID))
		resp.State.RemoveResource(ctx)
		return
	}

	state.ID = state.WorkspaceID
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceTFEWorkspaceHYOKEnabled) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state modelTFEWorkspaceHYOKEnabled
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	tflog.Debug(ctx, fmt.Sprintf("Unexpected update to tfe_workspace_hyok_enabled for workspace %s was attempted and skipped", state.WorkspaceID.ValueString()))
}

func (r *resourceTFEWorkspaceHYOKEnabled) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state modelTFEWorkspaceHYOKEnabled
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	tflog.Debug(ctx, "HYOK will continue to be enabled despite resource being deleted")
	if resp.Diagnostics.HasError() {
		return
	}
}
