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
		MarkdownDescription: "Manages an Azure OIDC configuration." +
			"\n\n~> **Note:** This resource requires using the provider with HCP Terraform on the HCP Terraform Premium edition. Refer to [HCP Terraform pricing](https://www.hashicorp.com/en/pricing?product_intent=terraform&tab=terraform) for details.",
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

	clientID := plan.ClientID.ValueString()
	subscriptionID := plan.SubscriptionID.ValueString()
	tenantID := plan.TenantID.ValueString()

	attrs := models.NewAzureOidcConfigurations_attributes()
	attrs.SetClientId(&clientID)
	attrs.SetSubscriptionId(&subscriptionID)
	attrs.SetTenantId(&tenantID)

	azureData := models.NewAzureOidcConfigurations()
	azureData.SetAttributes(attrs)
	azureType := models.AZUREOIDCCONFIGURATIONS_AZUREOIDCCONFIGURATIONS_TYPE
	azureData.SetTypeEscaped(&azureType)

	inner := models.NewOidcConfigurationEnvelope_OidcConfigurationEnvelope_data()
	inner.SetAzureOidcConfigurations(azureData)

	envelope := models.NewOidcConfigurationEnvelope()
	envelope.SetData(inner)

	tflog.Debug(ctx, fmt.Sprintf("Create TFE Azure OIDC Configuration for organization %s", orgName))
	oidcEnvelope, err := r.config.ClientV2.API.Organizations().ByOrganization_name(orgName).OidcConfigurations().Post(ctx, envelope, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error creating TFE Azure OIDC Configuration", err.Error())
		return
	}

	oidc, err := extractAzureOIDCData(oidcEnvelope)
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
	oidcEnvelope, err := r.config.ClientV2.API.OidcConfigurations().ByOidc_configuration_id(oidcID).Get(ctx, nil)
	if err != nil {
		if errors.Is(err, tfev2.ErrNotFound) {
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

	oidc, err := extractAzureOIDCData(oidcEnvelope)
	if err != nil {
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

	oidcID := state.ID.ValueString()

	clientID := plan.ClientID.ValueString()
	subscriptionID := plan.SubscriptionID.ValueString()
	tenantID := plan.TenantID.ValueString()

	attrs := models.NewAzureOidcConfigurations_attributes()
	attrs.SetClientId(&clientID)
	attrs.SetSubscriptionId(&subscriptionID)
	attrs.SetTenantId(&tenantID)

	azureData := models.NewAzureOidcConfigurations()
	azureData.SetAttributes(attrs)
	azureType := models.AZUREOIDCCONFIGURATIONS_AZUREOIDCCONFIGURATIONS_TYPE
	azureData.SetTypeEscaped(&azureType)

	inner := models.NewOidcConfigurationEnvelope_OidcConfigurationEnvelope_data()
	inner.SetAzureOidcConfigurations(azureData)

	envelope := models.NewOidcConfigurationEnvelope()
	envelope.SetData(inner)

	tflog.Debug(ctx, fmt.Sprintf("Update TFE Azure OIDC Configuration %s", oidcID))
	oidcEnvelope, err := r.config.ClientV2.API.OidcConfigurations().ByOidc_configuration_id(oidcID).Patch(ctx, envelope, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error updating TFE Azure OIDC Configuration", err.Error())
		return
	}

	oidc, err := extractAzureOIDCData(oidcEnvelope)
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
	err := r.config.ClientV2.API.OidcConfigurations().ByOidc_configuration_id(oidcID).Delete(ctx, nil)
	if err != nil {
		if errors.Is(err, tfev2.ErrNotFound) {
			tflog.Debug(ctx, fmt.Sprintf("TFE Azure OIDC configuration %s no longer exists", oidcID))
			return
		}

		resp.Diagnostics.AddError("Error deleting TFE Azure OIDC Configuration", err.Error())
		return
	}
}

// extractAzureOIDCData pulls the AzureOidcConfigurations typed value out of a composed-type envelope.
func extractAzureOIDCData(envelope models.OidcConfigurationEnvelopeable) (models.AzureOidcConfigurationsable, error) {
	if envelope == nil || envelope.GetData() == nil {
		return nil, fmt.Errorf("no data returned by API")
	}
	data := envelope.GetData().GetAzureOidcConfigurations()
	if data == nil {
		return nil, fmt.Errorf("unexpected OIDC configuration type in API response")
	}
	return data, nil
}

func modelFromTFEAzureOIDCConfiguration(p models.AzureOidcConfigurationsable) modelTFEAzureOIDCConfiguration {
	m := modelTFEAzureOIDCConfiguration{
		ID: types.StringValue(valueOrZero(p.GetId())),
	}
	if attrs := p.GetAttributes(); attrs != nil {
		m.ClientID = types.StringValue(valueOrZero(attrs.GetClientId()))
		m.SubscriptionID = types.StringValue(valueOrZero(attrs.GetSubscriptionId()))
		m.TenantID = types.StringValue(valueOrZero(attrs.GetTenantId()))
	}
	if rel := p.GetRelationships(); rel != nil {
		if org := rel.GetOrganization(); org != nil && org.GetData() != nil {
			m.Organization = types.StringValue(valueOrZero(org.GetData().GetId()))
		}
	}
	return m
}
