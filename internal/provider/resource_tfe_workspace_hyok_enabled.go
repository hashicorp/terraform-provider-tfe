// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
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

// assuming we'll need to update the docs when this comes out too? if yes which docs and where
// configure
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

// import state
func (r *resourceTFEWorkspaceHYOKEnabled) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("workspace_id"), req, resp)
}

// resource.Resource
func (r *resourceTFEWorkspaceHYOKEnabled) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workspace_hyok_enabled"
}
func (r *resourceTFEWorkspaceHYOKEnabled) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Enables HYOK on selected workspace. \n\n~> **Note:** HYOK is *irreversible* once enabled in a workspace and will persist being destroyed. " +
			"Requires HCP Terraform Premium",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "????",
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

// CRUD
// Create

// Maybe add workspace name here and in the schema for better error messages?
type modelTFEWorkspaceHYOKEnabled struct {
	ID          types.String `tfsdk:"id"`
	WorkspaceID types.String `tfsdk:"workspace_id"`
}

//API Model not needed?

func (r *resourceTFEWorkspaceHYOKEnabled) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	//boilerplate error checking
	var plan modelTFEWorkspaceHYOKEnabled
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	//check if already enabled ----
	//return debug already done
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

	//don't need org name right?
	//actually maybe we do becuase we need to verify that it can enable this but that should be handled higher up?
	// var organization string
	// resp.Diagnostics.Appen(r.config.dataOrDefaultOrganization(ctx, req.Plan, &organization)...)
	// if resp.Diagnostics.hasError() {
	// 	return
	// }

	options := tfe.WorkspaceUpdateOptions{
		HYOKEnabled: tfe.Bool(true),
	}

	tflog.Debug(ctx, "Enabling HYOK")
	_, err = r.config.Client.Workspaces.UpdateByID(ctx, workspaceID, options)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Error Enabling HYOK on workspace %s, ", workspaceID), err.Error())
		return
	}

	//not sure about this but also no other meaningful thing to put for an id
	//could hash it but no real point to it
	// I guess it could later be used to verify whether or not a given workspace has hyok enabled
	//plan.ID = plan.WorkspaceID
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// read
func (r *resourceTFEWorkspaceHYOKEnabled) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state modelTFEWorkspaceHYOKEnabled
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	_, err := r.config.Client.Workspaces.ReadByID(ctx, state.WorkspaceID.ValueString())
	if err != nil {
		//workspace not found or bad connection
		//????????
		//delete if workspace no longer exists?
		resp.Diagnostics.AddError(fmt.Sprintf("Error finding workspace %s, ", state.WorkspaceID.String()), err.Error())
		return
	}
	// should not be a case where hyok is turned off so there really shouldnt be any new info when the state is refreshed?

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// update - can't update
func (r *resourceTFEWorkspaceHYOKEnabled) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	//no-op
}

// delete - hyok should still be running even if resource is deleted
func (r *resourceTFEWorkspaceHYOKEnabled) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state modelTFEWorkspaceHYOKEnabled
	//
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	//should auto delete the resource?
	//do not destroy
}
