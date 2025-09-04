// // Copyright (c) HashiCorp, Inc.
// // SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/path"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.ResourceWithConfigure   = &resourceTFEAzureOIDCConfiguration{}
	_ resource.ResourceWithImportState = &resourceTFEAzureOIDCConfiguration{}
)

func NewAzureOIDCConfigurationResource() resource.Resource {
	return &resourceTFEAzureOIDCConfiguration{}
}

type resourceTFEAzureOIDCConfiguration struct {
	config ConfiguredClient
}

type modelTFEAzureOIDCConfiguration struct {
	ID             types.String `tfsdk:"id"`
	ClientID       types.String `tfsdk:"client_id"`
	SubscriptionID types.String `tfsdk:"subscription_id"`
	TenantID       types.String `tfsdk:"tenant_id"`
	Organization   types.String `tfsdk:"organization"`
}

func (r *resourceTFEAzureOIDCConfiguration) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *resourceTFEAzureOIDCConfiguration) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_azure_oidc_configuration"
}

func (r *resourceTFEAzureOIDCConfiguration) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the Azure OIDC configuration.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"client_id": schema.StringAttribute{
				Description: "The Client (or Application) ID of your Entra ID application.",
				Required:    true,
			},
			"subscription_id": schema.StringAttribute{
				Description: "The ID of your Azure subscription.",
				Required:    true,
			},
			"tenant_id": schema.StringAttribute{
				Description: "The Tenant (or Directory) ID of your Entra ID application.",
				Required:    true,
			},
			"organization": schema.StringAttribute{
				Description: "Name of the organization to which the TFE Azure OIDC configuration belongs.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Description: "Generates a new TFE Azure OIDC Configuration.",
	}
}

func (r *resourceTFEAzureOIDCConfiguration) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *resourceTFEAzureOIDCConfiguration) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read Terraform plan into the model
	var plan modelTFEAzureOIDCConfiguration
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the organization name from resource or provider config
	var orgName string
	resp.Diagnostics.Append(r.config.dataOrDefaultOrganization(ctx, req.Config, &orgName)...)
	if resp.Diagnostics.HasError() {
		return
	}

	options := tfe.AzureOIDCConfigurationCreateOptions{
		ClientID:       plan.ClientID.ValueString(),
		SubscriptionID: plan.SubscriptionID.ValueString(),
		TenantID:       plan.TenantID.ValueString(),
	}

	tflog.Debug(ctx, fmt.Sprintf("Create TFE Azure OIDC Configuration for organization %s", orgName))
	oidc, err := r.config.Client.AzureOIDCConfigurations.Create(ctx, orgName, options)
	if err != nil {
		resp.Diagnostics.AddError("Error creating TFE Azure OIDC Configuration", err.Error())
		return
	}
	result := modelFromTFEAzureOIDCConfiguration(oidc)
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

func (r *resourceTFEAzureOIDCConfiguration) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read Terraform state into the model
	var state modelTFEAzureOIDCConfiguration
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	oidcID := state.ID.ValueString()
	tflog.Debug(ctx, fmt.Sprintf("Read Azure OIDC configuration: %s", oidcID))
	oidc, err := r.config.Client.AzureOIDCConfigurations.Read(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
			tflog.Debug(ctx, fmt.Sprintf("Azure OIDC configuration %s no longer exists", oidcID))
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error reading Azure OIDC configuration %s", oidcID),
			err.Error(),
		)
		return
	}
	result := modelFromTFEAzureOIDCConfiguration(oidc)
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

func (r *resourceTFEAzureOIDCConfiguration) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan modelTFEAzureOIDCConfiguration
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	var state modelTFEAzureOIDCConfiguration
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	options := tfe.AzureOIDCConfigurationUpdateOptions{
		ClientID:       plan.ClientID.ValueStringPointer(),
		SubscriptionID: plan.SubscriptionID.ValueStringPointer(),
		TenantID:       plan.TenantID.ValueStringPointer(),
	}

	oidcID := state.ID.ValueString()
	tflog.Debug(ctx, fmt.Sprintf("Update TFE Azure OIDC Configuration %s", oidcID))
	oidc, err := r.config.Client.AzureOIDCConfigurations.Update(ctx, oidcID, options)
	if err != nil {
		resp.Diagnostics.AddError("Error updating TFE Azure OIDC Configuration", err.Error())
		return
	}

	result := modelFromTFEAzureOIDCConfiguration(oidc)
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

func (r *resourceTFEAzureOIDCConfiguration) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state modelTFEAzureOIDCConfiguration
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	oidcID := state.ID.ValueString()
	tflog.Debug(ctx, fmt.Sprintf("Delete TFE Azure OIDC configuration: %s", oidcID))
	err := r.config.Client.AzureOIDCConfigurations.Delete(ctx, oidcID)
	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
			tflog.Debug(ctx, fmt.Sprintf("TFE Azure OIDC configuration %s no longer exists", oidcID))
		}

		resp.Diagnostics.AddError("Error deleting TFE Azure OIDC Configuration", err.Error())
		return
	}
}

func modelFromTFEAzureOIDCConfiguration(p *tfe.AzureOIDCConfiguration) modelTFEAzureOIDCConfiguration {
	return modelTFEAzureOIDCConfiguration{
		ID:             types.StringValue(p.ID),
		ClientID:       types.StringValue(p.ClientID),
		SubscriptionID: types.StringValue(p.SubscriptionID),
		TenantID:       types.StringValue(p.TenantID),
		Organization:   types.StringValue(p.Organization.Name),
	}
}
