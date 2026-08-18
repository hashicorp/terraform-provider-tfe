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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.ResourceWithConfigure   = &resourceTFEVaultOIDCConfiguration{}
	_ resource.ResourceWithImportState = &resourceTFEVaultOIDCConfiguration{}
)

func NewVaultOIDCConfigurationResource() resource.Resource {
	return &resourceTFEVaultOIDCConfiguration{}
}

type resourceTFEVaultOIDCConfiguration struct {
	config ConfiguredClient
}

type modelTFEVaultOIDCConfiguration struct {
	ID               types.String `tfsdk:"id"`
	Address          types.String `tfsdk:"address"`
	RoleName         types.String `tfsdk:"role_name"`
	Namespace        types.String `tfsdk:"namespace"`
	JWTAuthPath      types.String `tfsdk:"auth_path"`
	TLSCACertificate types.String `tfsdk:"encoded_cacert"`
	Organization     types.String `tfsdk:"organization"`
}

func (r *resourceTFEVaultOIDCConfiguration) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *resourceTFEVaultOIDCConfiguration) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vault_oidc_configuration"
}

func (r *resourceTFEVaultOIDCConfiguration) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Vault OIDC configuration.\n\n~> **Note:** This resource requires using the provider with HCP Terraform on the HCP Terraform Premium edition. Refer to [HCP Terraform pricing](https://www.hashicorp.com/en/pricing?product_intent=terraform&tab=terraform) for details.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the Vault OIDC configuration.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"address": schema.StringAttribute{
				Description: "The full address of your Vault instance.",
				Required:    true,
			},
			"role_name": schema.StringAttribute{
				Description: "The name of a role in your Vault JWT auth path, with permission to encrypt and decrypt with a Transit secrets engine key.",
				Required:    true,
			},
			"namespace": schema.StringAttribute{
				Description: "The namespace your JWT auth path is mounted in.",
				Required:    true,
			},
			"auth_path": schema.StringAttribute{
				MarkdownDescription: "The mount path of the JWT authentication method. Defaults to `\"jwt\"`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("jwt"),
			},
			"encoded_cacert": schema.StringAttribute{
				Description: "A base64 encoded certificate which can be used to authenticate your Vault certificate. Only needed for self-hosted Vault Enterprise instances with a self-signed certificate.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
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

func (r *resourceTFEVaultOIDCConfiguration) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *resourceTFEVaultOIDCConfiguration) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read Terraform plan into the model
	var plan modelTFEVaultOIDCConfiguration
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

	address := plan.Address.ValueString()
	roleName := plan.RoleName.ValueString()
	namespace := plan.Namespace.ValueString()
	authPath := plan.JWTAuthPath.ValueString()
	encodedCACert := plan.TLSCACertificate.ValueString()

	attrs := models.NewVaultOidcConfigurations_attributes()
	attrs.SetAddress(&address)
	attrs.SetRole(&roleName)
	attrs.SetNamespace(&namespace)
	attrs.SetAuthPath(&authPath)
	attrs.SetEncodedCaCert(&encodedCACert)

	vaultData := models.NewVaultOidcConfigurations()
	vaultData.SetAttributes(attrs)
	vaultType := models.VAULTOIDCCONFIGURATIONS_VAULTOIDCCONFIGURATIONS_TYPE
	vaultData.SetTypeEscaped(&vaultType)

	inner := models.NewOidcConfigurationEnvelope_OidcConfigurationEnvelope_data()
	inner.SetVaultOidcConfigurations(vaultData)

	envelope := models.NewOidcConfigurationEnvelope()
	envelope.SetData(inner)

	tflog.Debug(ctx, fmt.Sprintf("Create TFE Vault OIDC Configuration for organization %s", orgName))
	oidcEnvelope, err := r.config.ClientV2.API.Organizations().ByOrganization_name(orgName).OidcConfigurations().Post(ctx, envelope, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error creating TFE Vault OIDC Configuration", err.Error())
		return
	}

	oidc, err := extractVaultOIDCData(oidcEnvelope)
	if err != nil {
		resp.Diagnostics.AddError("Error creating TFE Vault OIDC Configuration", err.Error())
		return
	}

	result := modelFromTFEVaultOIDCConfiguration(oidc)
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

func (r *resourceTFEVaultOIDCConfiguration) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read Terraform state into the model
	var state modelTFEVaultOIDCConfiguration
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	oidcID := state.ID.ValueString()
	tflog.Debug(ctx, fmt.Sprintf("Read Vault OIDC configuration: %s", oidcID))
	oidcEnvelope, err := r.config.ClientV2.API.OidcConfigurations().ByOidc_configuration_id(oidcID).Get(ctx, nil)
	if err != nil {
		if errors.Is(err, tfev2.ErrNotFound) {
			tflog.Debug(ctx, fmt.Sprintf("Vault OIDC configuration %s no longer exists", oidcID))
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error reading Vault OIDC configuration %s", oidcID),
			err.Error(),
		)
		return
	}

	oidc, err := extractVaultOIDCData(oidcEnvelope)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error reading Vault OIDC configuration %s", oidcID),
			err.Error(),
		)
		return
	}

	result := modelFromTFEVaultOIDCConfiguration(oidc)
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

func (r *resourceTFEVaultOIDCConfiguration) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan modelTFEVaultOIDCConfiguration
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	var state modelTFEVaultOIDCConfiguration
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	oidcID := state.ID.ValueString()

	address := plan.Address.ValueString()
	roleName := plan.RoleName.ValueString()
	namespace := plan.Namespace.ValueString()
	authPath := plan.JWTAuthPath.ValueString()
	encodedCACert := plan.TLSCACertificate.ValueString()

	attrs := models.NewVaultOidcConfigurations_attributes()
	attrs.SetAddress(&address)
	attrs.SetRole(&roleName)
	attrs.SetNamespace(&namespace)
	attrs.SetAuthPath(&authPath)
	attrs.SetEncodedCaCert(&encodedCACert)

	vaultData := models.NewVaultOidcConfigurations()
	vaultData.SetAttributes(attrs)
	vaultType := models.VAULTOIDCCONFIGURATIONS_VAULTOIDCCONFIGURATIONS_TYPE
	vaultData.SetTypeEscaped(&vaultType)

	inner := models.NewOidcConfigurationEnvelope_OidcConfigurationEnvelope_data()
	inner.SetVaultOidcConfigurations(vaultData)

	envelope := models.NewOidcConfigurationEnvelope()
	envelope.SetData(inner)

	tflog.Debug(ctx, fmt.Sprintf("Update TFE Vault OIDC Configuration %s", oidcID))
	oidcEnvelope, err := r.config.ClientV2.API.OidcConfigurations().ByOidc_configuration_id(oidcID).Patch(ctx, envelope, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error updating TFE Vault OIDC Configuration", err.Error())
		return
	}

	oidc, err := extractVaultOIDCData(oidcEnvelope)
	if err != nil {
		resp.Diagnostics.AddError("Error updating TFE Vault OIDC Configuration", err.Error())
		return
	}

	result := modelFromTFEVaultOIDCConfiguration(oidc)
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

func (r *resourceTFEVaultOIDCConfiguration) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state modelTFEVaultOIDCConfiguration
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	oidcID := state.ID.ValueString()
	tflog.Debug(ctx, fmt.Sprintf("Delete TFE Vault OIDC configuration: %s", oidcID))
	err := r.config.ClientV2.API.OidcConfigurations().ByOidc_configuration_id(oidcID).Delete(ctx, nil)
	if err != nil {
		if errors.Is(err, tfev2.ErrNotFound) {
			tflog.Debug(ctx, fmt.Sprintf("TFE Vault OIDC configuration %s no longer exists", oidcID))
			return
		}

		resp.Diagnostics.AddError("Error deleting TFE Vault OIDC Configuration", err.Error())
		return
	}
}

// extractVaultOIDCData pulls the VaultOidcConfigurations typed value out of a composed-type envelope.
func extractVaultOIDCData(envelope models.OidcConfigurationEnvelopeable) (models.VaultOidcConfigurationsable, error) {
	if envelope == nil || envelope.GetData() == nil {
		return nil, fmt.Errorf("no data returned by API")
	}
	data := envelope.GetData().GetVaultOidcConfigurations()
	if data == nil {
		return nil, fmt.Errorf("unexpected OIDC configuration type in API response")
	}
	return data, nil
}

func modelFromTFEVaultOIDCConfiguration(p models.VaultOidcConfigurationsable) modelTFEVaultOIDCConfiguration {
	m := modelTFEVaultOIDCConfiguration{
		ID: types.StringValue(valueOrZero(p.GetId())),
	}
	if attrs := p.GetAttributes(); attrs != nil {
		m.Address = types.StringValue(valueOrZero(attrs.GetAddress()))
		m.RoleName = types.StringValue(valueOrZero(attrs.GetRole()))
		m.Namespace = types.StringValue(valueOrZero(attrs.GetNamespace()))
		m.JWTAuthPath = types.StringValue(valueOrZero(attrs.GetAuthPath()))
		m.TLSCACertificate = types.StringValue(valueOrZero(attrs.GetEncodedCaCert()))
	}
	if rel := p.GetRelationships(); rel != nil {
		if org := rel.GetOrganization(); org != nil && org.GetData() != nil {
			m.Organization = types.StringValue(valueOrZero(org.GetData().GetId()))
		}
	}
	return m
}
