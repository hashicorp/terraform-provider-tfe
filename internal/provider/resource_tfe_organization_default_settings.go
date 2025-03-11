// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"log"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	RemoteExecutionMode = "remote"
	LocalExecutionMode  = "local"
	AgentExecutionMode  = "agent"

	DefaultExecutionMode = RemoteExecutionMode
)

var (
	_ resource.Resource                = (*resourceTFEOrganizationDefaultSettings)(nil)
	_ resource.ResourceWithConfigure   = (*resourceTFEOrganizationDefaultSettings)(nil)
	_ resource.ResourceWithImportState = (*resourceTFEOrganizationDefaultSettings)(nil)
	_ resource.ResourceWithModifyPlan  = (*resourceTFEOrganizationDefaultSettings)(nil)

	ValidExecutionModes = []string{
		AgentExecutionMode,
		LocalExecutionMode,
		RemoteExecutionMode,
	}
)

func NewOrganizationDefaultSettings() resource.Resource {
	return &resourceTFEOrganizationDefaultSettings{}
}

type resourceTFEOrganizationDefaultSettings struct {
	config ConfiguredClient
}

type modelTFEOrganizationDefaultSettings struct {
	Organization         types.String `tfsdk:"organization"`
	DefaultExecutionMode types.String `tfsdk:"default_execution_mode"`
	DefaultAgentPoolID   types.String `tfsdk:"default_agent_pool_id"`
}

func modelFromTFEOrganization(v *tfe.Organization) modelTFEOrganizationDefaultSettings {
	model := modelTFEOrganizationDefaultSettings{
		Organization:         types.StringValue(v.Name),
		DefaultExecutionMode: types.StringValue(v.DefaultExecutionMode),
	}

	if v.DefaultAgentPool != nil {
		model.DefaultAgentPoolID = types.StringValue(v.DefaultAgentPool.ID)
	}

	return model
}

// Configure implements resource.ResourceWithConfigure
func (r *resourceTFEOrganizationDefaultSettings) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *resourceTFEOrganizationDefaultSettings) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_default_settings"
}

// Schema implements resource.Resource
func (r *resourceTFEOrganizationDefaultSettings) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"organization": schema.StringAttribute{
				Description: "The name of the organization.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			"default_execution_mode": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOf(ValidExecutionModes...),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			"default_agent_pool_id": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

// Create implements resource.Resource
func (r *resourceTFEOrganizationDefaultSettings) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read Terraform plan data
	var data modelTFEOrganizationDefaultSettings
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get org name or default
	var orgName string
	resp.Diagnostics.Append(r.config.dataOrDefaultOrganization(ctx, req.Config, &orgName)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create options struct
	options := tfe.OrganizationUpdateOptions{}

	if !data.DefaultExecutionMode.IsNull() {
		options.DefaultExecutionMode = data.DefaultExecutionMode.ValueStringPointer()
	}

	if !data.DefaultAgentPoolID.IsNull() {
		options.DefaultAgentPool = &tfe.AgentPool{
			ID: data.DefaultAgentPoolID.ValueString(),
		}
	}

	o, err := r.config.Client.Organizations.Update(ctx, orgName, options)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update organization default settings", err.Error())
		return
	}

	result := modelFromTFEOrganization(o)

	// Write the data back to the resource
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

// Update implements resource.Resource
func (r *resourceTFEOrganizationDefaultSettings) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform plan data
	var data modelTFEOrganizationDefaultSettings
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get org name or default
	var orgName string
	resp.Diagnostics.Append(r.config.dataOrDefaultOrganization(ctx, req.Config, &orgName)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create options struct
	options := tfe.OrganizationUpdateOptions{}

	if !data.DefaultExecutionMode.IsNull() {
		options.DefaultExecutionMode = data.DefaultExecutionMode.ValueStringPointer()
	}

	if !data.DefaultAgentPoolID.IsNull() {
		options.DefaultAgentPool = &tfe.AgentPool{
			ID: data.DefaultAgentPoolID.ValueString(),
		}
	}

	o, err := r.config.Client.Organizations.Update(ctx, orgName, options)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update organization default settings", err.Error())
		return
	}

	result := modelFromTFEOrganization(o)

	// Write the data back to the resource
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

// Read implements resource.Resource
func (r *resourceTFEOrganizationDefaultSettings) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read Terraform state data
	var data modelTFEOrganizationDefaultSettings
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get org name or default
	var orgName string
	resp.Diagnostics.Append(r.config.dataOrDefaultOrganization(ctx, req.State, &orgName)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get organization
	o, err := r.config.Client.Organizations.Read(ctx, orgName)
	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
			resp.Diagnostics.AddError("Organization not found", err.Error())
			return
		}

		resp.Diagnostics.AddError("Unable to read organization", err.Error())
		return
	}

	result := modelFromTFEOrganization(o)
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

// Delete implements resource.Resource
func (r *resourceTFEOrganizationDefaultSettings) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Read Terraform state data
	var data modelTFEOrganizationDefaultSettings
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get org name or default
	var orgName string
	resp.Diagnostics.Append(r.config.dataOrDefaultOrganization(ctx, req.State, &orgName)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create options struct with system defaults
	options := tfe.OrganizationUpdateOptions{
		DefaultExecutionMode: tfe.String(DefaultExecutionMode),
		DefaultAgentPool:     nil,
	}

	// Reset organization settings
	log.Printf("[DEBUG] Reseting default execution mode of organization: %s", orgName)
	o, err := r.config.Client.Organizations.Update(ctx, orgName, options)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update organization default settings", err.Error())
		return
	}

	result := modelFromTFEOrganization(o)
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

// ImportState implements resource.ResourceWithImportState
func (r *resourceTFEOrganizationDefaultSettings) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization"), req.ID)...)
}

func (r *resourceTFEOrganizationDefaultSettings) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}

	// Read Terraform plan data
	var data modelTFEOrganizationDefaultSettings
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !data.DefaultAgentPoolID.IsNull() && data.DefaultExecutionMode.ValueString() != AgentExecutionMode {
		resp.Diagnostics.AddError(
			"Invalid default_execution_mode",
			"Default execution mode must be set to 'agent' when default_agent_pool_id is set",
		)
	}
}
