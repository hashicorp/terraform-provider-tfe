// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/jsonapi"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
	Tags                        types.Map    `tfsdk:"tags"`
	IgnoreAdditionalTags        types.Bool   `tfsdk:"ignore_additional_tags"`
}

// modelFromTFEProject builds a modelTFEProject struct from a tfe.Project value.
func modelFromTFEProject(p *tfe.Project, tags []*tfe.TagBinding, ignoreAdditionalTags types.Bool) (modelTFEProject, diag.Diagnostics) {
	var diags diag.Diagnostics

	model := modelTFEProject{
		ID:                   types.StringValue(p.ID),
		Name:                 types.StringValue(p.Name),
		Description:          types.StringValue(p.Description),
		Organization:         types.StringValue(p.Organization.Name),
		IgnoreAdditionalTags: ignoreAdditionalTags,
	}

	if p.AutoDestroyActivityDuration.IsSpecified() {
		duration, err := p.AutoDestroyActivityDuration.Get()
		if err != nil {
			diags.AddAttributeError(path.Root("auto_destroy_activity_duration"), "Invalid duration", fmt.Sprintf("Error reading auto destroy activity duration: %v", err))
		}

		model.AutoDestroyActivityDuration = types.StringValue(duration)
	}

	tagElems := make(map[string]attr.Value)
	for _, binding := range tags {
		tagElems[binding.Key] = types.StringValue(binding.Value)
	}
	model.Tags = types.MapValueMust(types.StringType, tagElems)

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

			"tags": schema.MapAttribute{
				Description: "A map of key-value tags to add to the project.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
			},

			"ignore_additional_tags": schema.BoolAttribute{
				Description: "Explicitly ignores tags created outside of Terraform so they will not be overwritten by tags defined in configuration.",
				Optional:    true,
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

	tags := plan.Tags.Elements()
	for key, val := range tags {
		if strVal, ok := val.(types.String); ok && !strVal.IsNull() {
			options.TagBindings = append(options.TagBindings, &tfe.TagBinding{
				Key:   key,
				Value: strVal.ValueString(),
			})
		}
	}

	tflog.Debug(ctx, fmt.Sprintf("Create project %s", name))
	project, err := r.config.Client.Projects.Create(ctx, orgName, options)

	if err != nil {
		resp.Diagnostics.AddError("Error creating project", err.Error())
		return
	}

	result, diags := modelFromTFEProject(project, options.TagBindings, plan.IgnoreAdditionalTags)
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

	id := state.ID.ValueString()

	tflog.Debug(ctx, fmt.Sprintf("Read project %s", id))
	project, err := r.config.Client.Projects.ReadWithOptions(ctx, id, tfe.ProjectReadOptions{
		Include: []tfe.ProjectIncludeOpt{tfe.ProjectEffectiveTagBindings},
	})
	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
			tflog.Debug(ctx, fmt.Sprintf("Project %s no longer exists", id))
			resp.State.RemoveResource(ctx)
			return
		}

		if errors.Is(err, tfe.ErrInvalidIncludeValue) {
			tflog.Debug(ctx, fmt.Sprintf("[DEBUG] Project %s read failed due to unsupported Include; retrying without it", id))
			project, err = r.config.Client.Projects.Read(ctx, id)
			if err != nil && errors.Is(err, tfe.ErrResourceNotFound) {
				tflog.Debug(ctx, fmt.Sprintf("Project %s no longer exists", id))
				resp.State.RemoveResource(ctx)
				return
			} else if err != nil {
				resp.Diagnostics.AddError("Error reading project", err.Error())
				return
			}
		} else {
			resp.Diagnostics.AddError("Error reading project", err.Error())
			return
		}
	}

	tagBindings := []*tfe.TagBinding{}
	for _, binding := range project.EffectiveTagBindings {
		tagBindings = append(tagBindings, &tfe.TagBinding{
			Key:   binding.Key,
			Value: binding.Value,
		})
	}

	var result modelTFEProject
	var diags diag.Diagnostics
	if state.IgnoreAdditionalTags.ValueBool() {
		allowedTags := []*tfe.TagBinding{}

		currentTags := state.Tags.Elements()
		for _, binding := range tagBindings {
			if _, ok := currentTags[binding.Key]; ok {
				allowedTags = append(allowedTags, binding)
			}
		}
		result, diags = modelFromTFEProject(project, allowedTags, state.IgnoreAdditionalTags)
	} else {
		result, diags = modelFromTFEProject(project, tagBindings, state.IgnoreAdditionalTags)
	}

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

	id := state.ID.ValueString()
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

	for key, val := range plan.Tags.Elements() {
		if strVal, ok := val.(types.String); ok && !strVal.IsNull() {
			options.TagBindings = append(options.TagBindings, &tfe.TagBinding{
				Key:   key,
				Value: strVal.ValueString(),
			})
		}
	}

	if len(options.TagBindings) == 0 && !plan.IgnoreAdditionalTags.ValueBool() {
		err := r.config.Client.Projects.DeleteAllTagBindings(ctx, id)
		if err != nil {
			resp.Diagnostics.AddError("Error removing tag bindings from project", err.Error())
			return
		}
	}

	tflog.Debug(ctx, fmt.Sprintf("Update project %s", id))
	project, err := r.config.Client.Projects.Update(ctx, id, options)
	if err != nil {
		resp.Diagnostics.AddError("Error updating project", err.Error())
		return
	}

	result, diags := modelFromTFEProject(project, options.TagBindings, plan.IgnoreAdditionalTags)
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

	id := state.ID.ValueString()

	tflog.Debug(ctx, fmt.Sprintf("Delete project %s", id))
	err := r.config.Client.Projects.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
			tflog.Debug(ctx, fmt.Sprintf("Project %s no longer exists", id))
			// The resource is implicitly deleted from state after returning
			return
		}
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
