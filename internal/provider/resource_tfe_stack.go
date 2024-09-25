// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-tfe"
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

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &resourceTFEStack{}
var _ resource.ResourceWithConfigure = &resourceTFEStack{}
var _ resource.ResourceWithImportState = &resourceTFEStack{}

func NewStackResource() resource.Resource {
	return &resourceTFEStack{}
}

// resourceTFEStack implements the tfe_stack resource type
type resourceTFEStack struct {
	config ConfiguredClient
}

func (r *resourceTFEStack) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stack"
}

func (r *resourceTFEStack) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	pathVCSRepoOAuthTokenID := path.Expressions{
		path.MatchRelative().AtParent().AtName("oauth_token_id"),
	}
	pathGHAInstallationID := path.Expressions{
		path.MatchRelative().AtParent().AtName("github_app_installation_id"),
	}

	resp.Schema = schema.Schema{
		Description: "Defines a Stack resource. Note that a Stack cannot be destroyed if it contains deployments that have underlying managed resources.",
		Version:     1,

		Blocks: map[string]schema.Block{
			"vcs_repo": schema.SingleNestedBlock{
				Description: "VCS repository configuration for the Stack.",
				Attributes: map[string]schema.Attribute{
					"identifier": schema.StringAttribute{
						Description: "Identifier of the VCS repository.",
						Required:    true,
					},
					"branch": schema.StringAttribute{
						Description: "The repository branch that Terraform should use. This defaults to the respository's default branch (e.g. main).",
						Optional:    true,
					},
					"github_app_installation_id": schema.StringAttribute{
						Description: "The installation ID of the GitHub App. This conflicts with `oauth_token_id` and can only be used if `oauth_token_id` is not used.",
						Optional:    true,
						Validators: []validator.String{
							stringvalidator.AtLeastOneOf(pathVCSRepoOAuthTokenID...),
							stringvalidator.ConflictsWith(pathVCSRepoOAuthTokenID...),
						},
					},
					"oauth_token_id": schema.StringAttribute{
						Description: "The VCS Connection to use. This ID can be obtained from a `tfe_oauth_client` resource. This conflicts with `github_app_installation_id` and can only be used if `github_app_installation_id` is not used.",
						Optional:    true,
						Validators: []validator.String{
							stringvalidator.AtLeastOneOf(pathGHAInstallationID...),
							stringvalidator.ConflictsWith(pathGHAInstallationID...),
						},
					},
				},
			},
		},

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID of the Stack.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Description: "ID of the project that the Stack belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the Stack",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the Stack",
				Optional:    true,
			},
			"deployment_names": schema.SetAttribute{
				Description: "The time when the Stack was created.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"created_at": schema.StringAttribute{
				Description: "The time when the stack was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Description: "The time when the stack was last updated.",
				Computed:    true,
			},
		},
	}
}

// Configure implements resource.ResourceWithConfigure
func (r *resourceTFEStack) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *resourceTFEStack) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan modelTFEStack

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if resp.Diagnostics.HasError() {
		return
	}

	options := tfe.StackCreateOptions{
		Name: plan.Name.ValueString(),
		VCSRepo: &tfe.StackVCSRepo{
			Identifier:        plan.VCSRepo.Identifier.ValueString(),
			Branch:            plan.VCSRepo.Branch.ValueString(),
			GHAInstallationID: plan.VCSRepo.GHAInstallationID.ValueString(),
			OAuthTokenID:      plan.VCSRepo.OAuthTokenID.ValueString(),
		},
		Project: &tfe.Project{
			ID: plan.ProjectID.ValueString(),
		},
	}

	if !plan.Description.IsNull() {
		options.Description = tfe.String(plan.Description.ValueString())
	}

	tflog.Debug(ctx, "Creating stack")
	stack, err := r.config.Client.Stacks.Create(ctx, options)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create stack", err.Error())
		return
	}

	result := modelFromTFEStack(stack)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *resourceTFEStack) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state modelTFEStack

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Reading stack %q", state.ID.ValueString()))
	stack, err := r.config.Client.Stacks.Read(ctx, state.ID.ValueString(), nil)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read stack", err.Error())
		return
	}

	result := modelFromTFEStack(stack)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *resourceTFEStack) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan modelTFEStack
	var state modelTFEStack

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	options := tfe.StackUpdateOptions{
		Name:        tfe.String(plan.Name.ValueString()),
		Description: tfe.String(plan.Description.ValueString()),
		VCSRepo: &tfe.StackVCSRepo{
			Identifier:        plan.VCSRepo.Identifier.ValueString(),
			Branch:            plan.VCSRepo.Branch.ValueString(),
			GHAInstallationID: plan.VCSRepo.GHAInstallationID.ValueString(),
			OAuthTokenID:      plan.VCSRepo.OAuthTokenID.ValueString(),
		},
	}

	tflog.Debug(ctx, "Updating stack")
	stack, err := r.config.Client.Stacks.Update(ctx, state.ID.ValueString(), options)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update stack", err.Error())
		return
	}

	result := modelFromTFEStack(stack)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *resourceTFEStack) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state modelTFEStack

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting stack")
	err := r.config.Client.Stacks.Delete(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to delete stack", err.Error())
		return
	}
}

func (r *resourceTFEStack) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
