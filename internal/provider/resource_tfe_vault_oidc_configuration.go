// // Copyright (c) HashiCorp, Inc.
// // SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
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
				Description: `The mounting path of JWT auth path of JWT auth. Defaults to "jwt".`,
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("jwt"),
			},
			"encoded_cacert": schema.StringAttribute{
				Description: "A base64 encoded certificate which can be used to authenticate your Vault certificate. Only needed for self-hosted Vault Enterprise instances with a self-signed certificate.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"organization": schema.StringAttribute{
				Description: "Name of the organization to which the TFE Vault OIDC configuration belongs.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Description: "Generates a new TFE Vault OIDC Configuration.",
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

	options := tfe.VaultOIDCConfigurationCreateOptions{
		Address:          plan.Address.ValueString(),
		RoleName:         plan.RoleName.ValueString(),
		Namespace:        plan.Namespace.ValueString(),
		JWTAuthPath:      plan.JWTAuthPath.ValueString(),
		TLSCACertificate: plan.TLSCACertificate.ValueString(),
	}

	tflog.Debug(ctx, fmt.Sprintf("Create TFE Vault OIDC Configuration for organization %s", orgName))
	oidc, err := r.config.Client.VaultOIDCConfigurations.Create(ctx, orgName, options)
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
	oidc, err := r.config.Client.VaultOIDCConfigurations.Read(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
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

	options := tfe.VaultOIDCConfigurationUpdateOptions{
		Address:          plan.Address.ValueStringPointer(),
		RoleName:         plan.RoleName.ValueStringPointer(),
		Namespace:        plan.Namespace.ValueStringPointer(),
		JWTAuthPath:      plan.JWTAuthPath.ValueStringPointer(),
		TLSCACertificate: plan.TLSCACertificate.ValueStringPointer(),
	}

	oidcID := state.ID.ValueString()
	tflog.Debug(ctx, fmt.Sprintf("Update TFE Vault OIDC Configuration %s", oidcID))
	oidc, err := r.config.Client.VaultOIDCConfigurations.Update(ctx, oidcID, options)
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
	err := r.config.Client.VaultOIDCConfigurations.Delete(ctx, oidcID)
	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
			tflog.Debug(ctx, fmt.Sprintf("TFE Vault OIDC configuration %s no longer exists", oidcID))
		}

		resp.Diagnostics.AddError("Error deleting TFE Vault OIDC Configuration", err.Error())
		return
	}
}

func modelFromTFEVaultOIDCConfiguration(p *tfe.VaultOIDCConfiguration) modelTFEVaultOIDCConfiguration {
	return modelTFEVaultOIDCConfiguration{
		ID:               types.StringValue(p.ID),
		Address:          types.StringValue(p.Address),
		RoleName:         types.StringValue(p.RoleName),
		Namespace:        types.StringValue(p.Namespace),
		JWTAuthPath:      types.StringValue(p.JWTAuthPath),
		TLSCACertificate: types.StringValue(p.TLSCACertificate),
		Organization:     types.StringValue(p.Organization.Name),
	}
}
