// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"

	tfev2 "github.com/hashicorp/go-tfe/v2"
	"github.com/hashicorp/go-tfe/v2/api/models"
	"github.com/hashicorp/terraform-plugin-framework/path"
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
		MarkdownDescription: "Manages a GCP OIDC configuration." +
			"\n\n~> **Note:** This resource requires using the provider with HCP Terraform on the HCP Terraform Premium edition. Refer to [HCP Terraform pricing](https://www.hashicorp.com/en/pricing?product_intent=terraform&tab=terraform) for details.",
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
				Description: "The fully qualified workload provider path. This should be in the format `projects/{project_number}/locations/global/workloadIdentityPools/{workload_identity_pool_id}/providers/{workload_identity_pool_provider_id}`.",
				Required:    true,
			},
			"organization": schema.StringAttribute{
				Description: "Name of the organization. If omitted, organization must be defined in the provider config.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
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

	serviceAccountEmail := plan.ServiceAccountEmail.ValueString()
	projectNumber := plan.ProjectNumber.ValueString()
	workloadProviderName := plan.WorkloadProviderName.ValueString()

	attrs := models.NewGcpOidcConfigurations_attributes()
	attrs.SetServiceAccountEmail(&serviceAccountEmail)
	attrs.SetProjectNumber(&projectNumber)
	attrs.SetWorkloadProviderName(&workloadProviderName)

	gcpData := models.NewGcpOidcConfigurations()
	gcpData.SetAttributes(attrs)
	gcpType := models.GCPOIDCCONFIGURATIONS_GCPOIDCCONFIGURATIONS_TYPE
	gcpData.SetTypeEscaped(&gcpType)

	inner := models.NewOidcConfigurationEnvelope_OidcConfigurationEnvelope_data()
	inner.SetGcpOidcConfigurations(gcpData)

	envelope := models.NewOidcConfigurationEnvelope()
	envelope.SetData(inner)

	tflog.Debug(ctx, fmt.Sprintf("Create TFE GCP OIDC Configuration for organization %s", orgName))
	oidcEnvelope, err := r.config.ClientV2.API.Organizations().ByOrganization_name(orgName).OidcConfigurations().Post(ctx, envelope, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error creating TFE GCP OIDC Configuration", err.Error())
		return
	}

	oidc, err := extractGCPOIDCData(oidcEnvelope)
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
	oidcEnvelope, err := r.config.ClientV2.API.OidcConfigurations().ByOidc_configuration_id(oidcID).Get(ctx, nil)
	if err != nil {
		if errors.Is(err, tfev2.ErrNotFound) {
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

	oidc, err := extractGCPOIDCData(oidcEnvelope)
	if err != nil {
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

	oidcID := state.ID.ValueString()

	serviceAccountEmail := plan.ServiceAccountEmail.ValueString()
	projectNumber := plan.ProjectNumber.ValueString()
	workloadProviderName := plan.WorkloadProviderName.ValueString()

	attrs := models.NewGcpOidcConfigurations_attributes()
	attrs.SetServiceAccountEmail(&serviceAccountEmail)
	attrs.SetProjectNumber(&projectNumber)
	attrs.SetWorkloadProviderName(&workloadProviderName)

	gcpData := models.NewGcpOidcConfigurations()
	gcpData.SetAttributes(attrs)
	gcpType := models.GCPOIDCCONFIGURATIONS_GCPOIDCCONFIGURATIONS_TYPE
	gcpData.SetTypeEscaped(&gcpType)

	inner := models.NewOidcConfigurationEnvelope_OidcConfigurationEnvelope_data()
	inner.SetGcpOidcConfigurations(gcpData)

	envelope := models.NewOidcConfigurationEnvelope()
	envelope.SetData(inner)

	tflog.Debug(ctx, fmt.Sprintf("Update TFE GCP OIDC Configuration %s", oidcID))
	oidcEnvelope, err := r.config.ClientV2.API.OidcConfigurations().ByOidc_configuration_id(oidcID).Patch(ctx, envelope, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error updating TFE GCP OIDC Configuration", err.Error())
		return
	}

	oidc, err := extractGCPOIDCData(oidcEnvelope)
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
	err := r.config.ClientV2.API.OidcConfigurations().ByOidc_configuration_id(oidcID).Delete(ctx, nil)
	if err != nil {
		if errors.Is(err, tfev2.ErrNotFound) {
			tflog.Debug(ctx, fmt.Sprintf("TFE GCP OIDC configuration %s no longer exists", oidcID))
			return
		}

		resp.Diagnostics.AddError("Error deleting TFE GCP OIDC Configuration", err.Error())
		return
	}
}

// extractGCPOIDCData pulls the GcpOidcConfigurations typed value out of a composed-type envelope.
func extractGCPOIDCData(envelope models.OidcConfigurationEnvelopeable) (models.GcpOidcConfigurationsable, error) {
	if envelope == nil || envelope.GetData() == nil {
		return nil, fmt.Errorf("no data returned by API")
	}
	data := envelope.GetData().GetGcpOidcConfigurations()
	if data == nil {
		return nil, fmt.Errorf("unexpected OIDC configuration type in API response")
	}
	return data, nil
}

func modelFromTFEGCPOIDCConfiguration(p models.GcpOidcConfigurationsable) modelTFEGCPOIDCConfiguration {
	m := modelTFEGCPOIDCConfiguration{
		ID: types.StringValue(valueOrZero(p.GetId())),
	}
	if attrs := p.GetAttributes(); attrs != nil {
		m.ServiceAccountEmail = types.StringValue(valueOrZero(attrs.GetServiceAccountEmail()))
		m.ProjectNumber = types.StringValue(valueOrZero(attrs.GetProjectNumber()))
		m.WorkloadProviderName = types.StringValue(valueOrZero(attrs.GetWorkloadProviderName()))
	}
	if rel := p.GetRelationships(); rel != nil {
		if org := rel.GetOrganization(); org != nil && org.GetData() != nil {
			m.Organization = types.StringValue(valueOrZero(org.GetData().GetId()))
		}
	}
	return m
}
