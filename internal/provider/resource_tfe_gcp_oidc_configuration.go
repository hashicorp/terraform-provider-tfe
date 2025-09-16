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
	_ resource.ResourceWithConfigure   = &resourceTFEGCPOIDCConfiguration{}
	_ resource.ResourceWithImportState = &resourceTFEGCPOIDCConfiguration{}
)

func NewGCPOIDCConfigurationResource() resource.Resource {
	return &resourceTFEGCPOIDCConfiguration{}
}

type resourceTFEGCPOIDCConfiguration struct {
	config ConfiguredClient
}

type modelTFEGCPOIDCConfiguration struct {
	ID                   types.String `tfsdk:"id"`
	ServiceAccountEmail  types.String `tfsdk:"service_account_email"`
	ProjectNumber        types.String `tfsdk:"project_number"`
	WorkloadProviderName types.String `tfsdk:"workload_provider_name"`
	Organization         types.String `tfsdk:"organization"`
}

func (r *resourceTFEGCPOIDCConfiguration) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *resourceTFEGCPOIDCConfiguration) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gcp_oidc_configuration"
}

func (r *resourceTFEGCPOIDCConfiguration) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the GCP OIDC configuration.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"service_account_email": schema.StringAttribute{
				Description: "The email of your GCP service account, with permissions to encrypt and decrypt using a Cloud KMS key.",
				Required:    true,
			},
			"project_number": schema.StringAttribute{
				Description: "The GCP Project containing the workload provider and service account.",
				Required:    true,
			},
			"workload_provider_name": schema.StringAttribute{
				Description: "The fully qualified workload provider path.",
				Required:    true,
			},
			"organization": schema.StringAttribute{
				Description: "Name of the organization to which the TFE GCP OIDC configuration belongs.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Description: "Generates a new TFE GCP OIDC Configuration.",
	}
}

func (r *resourceTFEGCPOIDCConfiguration) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *resourceTFEGCPOIDCConfiguration) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read Terraform plan into the model
	var plan modelTFEGCPOIDCConfiguration
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

	options := tfe.GCPOIDCConfigurationCreateOptions{
		ServiceAccountEmail:  plan.ServiceAccountEmail.ValueString(),
		ProjectNumber:        plan.ProjectNumber.ValueString(),
		WorkloadProviderName: plan.WorkloadProviderName.ValueString(),
	}

	tflog.Debug(ctx, fmt.Sprintf("Create TFE GCP OIDC Configuration for organization %s", orgName))
	oidc, err := r.config.Client.GCPOIDCConfigurations.Create(ctx, orgName, options)
	if err != nil {
		resp.Diagnostics.AddError("Error creating TFE GCP OIDC Configuration", err.Error())
		return
	}
	result := modelFromTFEGCPOIDCConfiguration(oidc)
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

func (r *resourceTFEGCPOIDCConfiguration) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read Terraform state into the model
	var state modelTFEGCPOIDCConfiguration
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	oidcID := state.ID.ValueString()
	tflog.Debug(ctx, fmt.Sprintf("Read GCP OIDC configuration: %s", oidcID))
	oidc, err := r.config.Client.GCPOIDCConfigurations.Read(ctx, oidcID)
	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
			tflog.Debug(ctx, fmt.Sprintf("GCP OIDC configuration %s no longer exists", oidcID))
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error reading GCP OIDC configuration %s", oidcID),
			err.Error(),
		)
		return
	}
	result := modelFromTFEGCPOIDCConfiguration(oidc)
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

func (r *resourceTFEGCPOIDCConfiguration) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan modelTFEGCPOIDCConfiguration
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	var state modelTFEGCPOIDCConfiguration
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	options := tfe.GCPOIDCConfigurationUpdateOptions{
		ServiceAccountEmail:  plan.ServiceAccountEmail.ValueStringPointer(),
		ProjectNumber:        plan.ProjectNumber.ValueStringPointer(),
		WorkloadProviderName: plan.WorkloadProviderName.ValueStringPointer(),
	}

	oidcID := state.ID.ValueString()
	tflog.Debug(ctx, fmt.Sprintf("Update TFE GCP OIDC Configuration %s", oidcID))
	oidc, err := r.config.Client.GCPOIDCConfigurations.Update(ctx, oidcID, options)
	if err != nil {
		resp.Diagnostics.AddError("Error updating TFE GCP OIDC Configuration", err.Error())
		return
	}

	result := modelFromTFEGCPOIDCConfiguration(oidc)
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

func (r *resourceTFEGCPOIDCConfiguration) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state modelTFEGCPOIDCConfiguration
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	oidcID := state.ID.ValueString()
	tflog.Debug(ctx, fmt.Sprintf("Delete TFE GCP OIDC configuration: %s", oidcID))
	err := r.config.Client.GCPOIDCConfigurations.Delete(ctx, oidcID)
	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
			tflog.Debug(ctx, fmt.Sprintf("TFE GCP OIDC configuration %s no longer exists", oidcID))
		}

		resp.Diagnostics.AddError("Error deleting TFE GCP OIDC Configuration", err.Error())
		return
	}
}

func modelFromTFEGCPOIDCConfiguration(p *tfe.GCPOIDCConfiguration) modelTFEGCPOIDCConfiguration {
	return modelTFEGCPOIDCConfiguration{
		ID:                   types.StringValue(p.ID),
		ServiceAccountEmail:  types.StringValue(p.ServiceAccountEmail),
		WorkloadProviderName: types.StringValue(p.WorkloadProviderName),
		ProjectNumber:        types.StringValue(p.ProjectNumber),
		Organization:         types.StringValue(p.Organization.Name),
	}
}
