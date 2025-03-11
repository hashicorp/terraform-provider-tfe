// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// NOTE: This is a legacy resource and should be migrated to the Plugin
// Framework if substantial modifications are planned. See
// docs/new-resources.md if planning to use this code as boilerplate for
// a new resource.

package provider

import (
	"context"
	"fmt"
	"log"
	"regexp"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/jsonapi"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &resourceTFEProject{}
	_ resource.ResourceWithConfigure   = &resourceTFEProject{}
	_ resource.ResourceWithImportState = &resourceTFEProject{}
	_ resource.ResourceWithModifyPlan  = &resourceTFEProject{}
)

func NewProjectResource() resource.Resource {
	return &resourceTFEProject{}
}

type resourceTFEProject struct {
	config ConfiguredClient
}

// modelTFEProject maps the resource schema data to a struct.
type modelTFEProject struct {
	ID                          types.String `tfsdk:"id"`
	Name                        types.String `tfsdk:"name"`
	Description                 types.String `tfsdk:"description"`
	Organization                types.String `tfsdk:"organization"`
	AutoDestroyActivityDuration types.String `tfsdk:"auto_destroy_activity_duration"`
}

// modelFromTFEProject builds a modelTFEProject struct from a tfe.Project value.
func modelFromTFEProject(p *tfe.Project) (modelTFEProject, diag.Diagnostics) {
	var diags diag.Diagnostics

	model := modelTFEProject{
		ID:           types.StringValue(p.ID),
		Name:         types.StringValue(p.Name),
		Description:  types.StringValue(p.Description),
		Organization: types.StringValue(p.Organization.Name),
	}

	if p.AutoDestroyActivityDuration.IsSpecified() {
		duration, err := p.AutoDestroyActivityDuration.Get()
		if err != nil {
			diags.AddAttributeError(path.Root("auto_destroy_activity_duration"), "Invalid duration", fmt.Sprintf("Error reading auto destroy activity duration: %v", err))
		}

		model.AutoDestroyActivityDuration = types.StringValue(duration)
	}

	return model, diags
}

// Configure implements resource.ResourceWithConfigure
func (r *resourceTFEProject) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Early exit if provider is unconfigured (i.e. we're only validating config or something)
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
func (r *resourceTFEProject) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

// Schema implements resource.Resource
func (r *resourceTFEProject) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Service-generated identifier for the variable",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"name": schema.StringAttribute{
				Description: "Name of the project.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 40),
					stringvalidator.RegexMatches(regexp.MustCompile(`\A[\w\-][\w\- ]+[\w\-]\z`),
						"can only include letters, numbers, spaces, -, and _.",
					),
				},
			},

			"description": schema.StringAttribute{
				Description: "Description of the project.",
				Optional:    true,
				Computed:    true,
			},

			"organization": schema.StringAttribute{
				Description: "Name of the organization to which the project belongs.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			"auto_destroy_activity_duration": schema.StringAttribute{
				Description: "Duration after which the project will be auto-destroyed.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^\d{1,4}[dh]$`),
						"must be 1-4 digits followed by 'd' or 'h'.",
					),
				},
			},
		},
	}
}

// Create implements resource.Resource
func (r *resourceTFEProject) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan modelTFEProject
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the organization name from resource or provider config
	var orgName string
	resp.Diagnostics.Append(r.config.dataOrDefaultOrganization(ctx, req.Config, &orgName)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()

	options := tfe.ProjectCreateOptions{
		Name:        name,
		Description: plan.Description.ValueStringPointer(),
	}

	if !plan.AutoDestroyActivityDuration.IsNull() {
		options.AutoDestroyActivityDuration = jsonapi.NewNullableAttrWithValue(plan.AutoDestroyActivityDuration.ValueString())
	}

	log.Printf("[DEBUG] Create project %s", name)
	project, err := r.config.Client.Projects.Create(ctx, orgName, options)

	if err != nil {
		resp.Diagnostics.AddError("Error creating project", err.Error())
		return
	}

	result, diags := modelFromTFEProject(project)
	if diags.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

// Read implements resource.Resource
func (r *resourceTFEProject) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state modelTFEProject
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	log.Printf("[DEBUG] Read project %s", state.ID.ValueString())
	project, err := r.config.Client.Projects.Read(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading project", err.Error())
		return
	}

	result, diags := modelFromTFEProject(project)
	if diags.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

// Update implements resource.Resource
func (r *resourceTFEProject) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan modelTFEProject
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state modelTFEProject
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()

	options := tfe.ProjectUpdateOptions{
		Name:        &name,
		Description: plan.Description.ValueStringPointer(),
	}

	// If auto_destroy_activity_duration was previously specified and is now being
	// cleared out, set an explicit null in the update options struct.
	if !state.AutoDestroyActivityDuration.IsNull() && plan.AutoDestroyActivityDuration.IsNull() {
		options.AutoDestroyActivityDuration = jsonapi.NewNullNullableAttr[string]()
	} else if !plan.AutoDestroyActivityDuration.IsNull() {
		options.AutoDestroyActivityDuration = jsonapi.NewNullableAttrWithValue(plan.AutoDestroyActivityDuration.ValueString())
	}

	log.Printf("[DEBUG] Update project %s", plan.ID.ValueString())
	project, err := r.config.Client.Projects.Update(ctx, plan.ID.ValueString(), options)
	if err != nil {
		resp.Diagnostics.AddError("Error updating project", err.Error())
		return
	}

	result, diags := modelFromTFEProject(project)
	if diags.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

// Delete implements resource.Resource
func (r *resourceTFEProject) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state modelTFEProject
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	log.Printf("[DEBUG] Delete project %s", state.ID.ValueString())
	err := r.config.Client.Projects.Delete(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting project", err.Error())
		return
	}
}

func (r *resourceTFEProject) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}

	modifyPlanForDefaultOrganizationChange(ctx, r.config.Organization, req.State, req.Config, req.Plan, resp)
}

func (r *resourceTFEProject) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
