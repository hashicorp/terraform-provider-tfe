// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"log"

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
)

const (
	RemoteExecutionMode = "remote"
	LocalExecutionMode  = "local"
	AgentExecutionMode  = "agent"

	DefaultExecutionMode = RemoteExecutionMode
)

var (
	_ resource.Resource                   = (*resourceTFEOrganizationDefaultSettings)(nil)
	_ resource.ResourceWithConfigure      = (*resourceTFEOrganizationDefaultSettings)(nil)
	_ resource.ResourceWithImportState    = (*resourceTFEOrganizationDefaultSettings)(nil)
	_ resource.ResourceWithValidateConfig = (*resourceTFEOrganizationDefaultSettings)(nil)

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

func modelFromOrganizationsV2(org models.Organizationsable, orgName string) modelTFEOrganizationDefaultSettings {
	model := modelTFEOrganizationDefaultSettings{
		Organization: types.StringValue(orgName),
	}

	if attrs := org.GetAttributes(); attrs != nil {
		if mode := attrs.GetDefaultExecutionMode(); mode != nil {
			model.DefaultExecutionMode = types.StringValue(mode.String())
		}
	}

	if rels := org.GetRelationships(); rels != nil {
		if ap := rels.GetDefaultAgentPool(); ap != nil {
			if data := ap.GetData(); data != nil {
				if id := data.GetId(); id != nil && *id != "" {
					model.DefaultAgentPoolID = types.StringValue(*id)
				}
			}
		}
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
		MarkdownDescription: "Manages default settings for an organization." +
			"\n\nPrimarily used to set the default execution mode of an organization. Settings configured here will be used as the default for all workspaces in the organization, unless they specify their own values with a [`tfe_workspace_settings` resource](workspace_settings.html) (or deprecated attributes on the workspace resource).",
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
				MarkdownDescription: "Which [execution mode](https://developer.hashicorp.com/terraform/cloud-docs/workspaces/settings#execution-mode) to use as the default for all workspaces in the organization. Valid values are `remote`, `local` or `agent`.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(ValidExecutionModes...),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			"default_agent_pool_id": schema.StringAttribute{
				Description: "The ID of an agent pool to assign as the default for workspaces in the organization. Requires `default_execution_mode` to be set to `agent`. This value must not be provided if `default_execution_mode` is set to any other value.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

// buildOrganizationUpdateBody builds a v2 PATCH body for organization default settings.
func buildOrganizationUpdateBody(execMode string, agentPoolID string) models.OrganizationsEnvelopeable {
	body := models.NewOrganizationsEnvelope()
	org := models.NewOrganizations()
	orgType := models.ORGANIZATIONS_ORGANIZATIONS_TYPE
	org.SetTypeEscaped(&orgType)

	attrs := models.NewOrganizations_attributes()
	if execMode != "" {
		mode, _ := models.ParseOrganizations_attributes_defaultExecutionMode(execMode)
		if mode != nil {
			modeTyped := mode.(*models.Organizations_attributes_defaultExecutionMode)
			attrs.SetDefaultExecutionMode(modeTyped)
		}
	}
	org.SetAttributes(attrs)

	if agentPoolID != "" {
		rels := models.NewOrganizations_relationships()
		ap := models.NewAgentPoolsHasOne()
		apData := models.NewAgentPoolsHasOne_data()
		apData.SetId(&agentPoolID)
		apType := models.AGENTPOOLS_AGENTPOOLSIDENTIFIER_TYPE
		apData.SetTypeEscaped(&apType)
		ap.SetData(apData)
		rels.SetDefaultAgentPool(ap)
		org.SetRelationships(rels)
	}

	body.SetData(org)
	return body
}

// buildOrganizationResetBody builds a v2 PATCH body to reset organization default settings.
func buildOrganizationResetBody() models.OrganizationsEnvelopeable {
	body := models.NewOrganizationsEnvelope()
	org := models.NewOrganizations()
	orgType := models.ORGANIZATIONS_ORGANIZATIONS_TYPE
	org.SetTypeEscaped(&orgType)

	attrs := models.NewOrganizations_attributes()
	mode, _ := models.ParseOrganizations_attributes_defaultExecutionMode(DefaultExecutionMode)
	if mode != nil {
		modeTyped := mode.(*models.Organizations_attributes_defaultExecutionMode)
		attrs.SetDefaultExecutionMode(modeTyped)
	}
	org.SetAttributes(attrs)

	// Do not send a relationship body; the Atlas controller disassociates the
	// default agent pool automatically when execution mode is reset to "remote".
	body.SetData(org)
	return body
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

	execMode := data.DefaultExecutionMode.ValueString()
	agentPoolID := data.DefaultAgentPoolID.ValueString()

	env, err := r.config.ClientV2.API.Organizations().ByOrganization_name(orgName).Patch(ctx, buildOrganizationUpdateBody(execMode, agentPoolID), nil)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update organization default settings", err.Error())
		return
	}

	org := env.GetData()
	if org == nil {
		resp.Diagnostics.AddError("Unable to update organization default settings", "API returned no data")
		return
	}

	result := modelFromOrganizationsV2(org, orgName)
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

	execMode := data.DefaultExecutionMode.ValueString()
	agentPoolID := data.DefaultAgentPoolID.ValueString()

	env, err := r.config.ClientV2.API.Organizations().ByOrganization_name(orgName).Patch(ctx, buildOrganizationUpdateBody(execMode, agentPoolID), nil)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update organization default settings", err.Error())
		return
	}

	org := env.GetData()
	if org == nil {
		resp.Diagnostics.AddError("Unable to update organization default settings", "API returned no data")
		return
	}

	result := modelFromOrganizationsV2(org, orgName)
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
	env, err := r.config.ClientV2.API.Organizations().ByOrganization_name(orgName).Get(ctx, nil)
	if err != nil {
		if errors.Is(err, tfev2.ErrNotFound) {
			resp.Diagnostics.AddError("Organization not found", err.Error())
			return
		}

		resp.Diagnostics.AddError("Unable to read organization", err.Error())
		return
	}

	org := env.GetData()
	if org == nil {
		resp.Diagnostics.AddError("Unable to read organization", "API returned no data")
		return
	}

	result := modelFromOrganizationsV2(org, orgName)
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

	// Reset organization settings to defaults
	log.Printf("[DEBUG] Resetting default execution mode of organization: %s", orgName)
	_, err := r.config.ClientV2.API.Organizations().ByOrganization_name(orgName).Patch(ctx, buildOrganizationResetBody(), nil)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update organization default settings", err.Error())
		return
	}
}

// ImportState implements resource.ResourceWithImportState
func (r *resourceTFEOrganizationDefaultSettings) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization"), req.ID)...)
}

func (r *resourceTFEOrganizationDefaultSettings) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	// Read Terraform plan data
	var config modelTFEOrganizationDefaultSettings
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !config.DefaultAgentPoolID.IsNull() && config.DefaultExecutionMode.ValueString() != AgentExecutionMode {
		resp.Diagnostics.AddError(
			"Invalid default_execution_mode",
			"Default execution mode must be set to 'agent' when default_agent_pool_id is set",
		)
	}
}
