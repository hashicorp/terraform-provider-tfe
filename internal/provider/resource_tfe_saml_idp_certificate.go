// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	tfe "github.com/hashicorp/go-tfe/v2"
	"github.com/hashicorp/go-tfe/v2/api/models"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-tfe/internal/provider/customtypes"
)

const (
	// minTFEVersionSAMLIDPCertificates is the first Terraform Enterprise
	// release that ships the /admin/saml-settings/idp-certificates API.
	minTFEVersionSAMLIDPCertificates string = "2.1.0"

	samlIDPCertRoleManaged string = "managed"
)

var (
	_ resource.Resource                = &resourceTFESAMLIDPCertificate{}
	_ resource.ResourceWithConfigure   = &resourceTFESAMLIDPCertificate{}
	_ resource.ResourceWithImportState = &resourceTFESAMLIDPCertificate{}
	_ resource.ResourceWithModifyPlan  = &resourceTFESAMLIDPCertificate{}
)

// NewSAMLIDPCertificateResource returns the SAML IdP certificate resource.
func NewSAMLIDPCertificateResource() resource.Resource {
	return &resourceTFESAMLIDPCertificate{}
}

type resourceTFESAMLIDPCertificate struct {
	config ConfiguredClient
}

type modelTFESAMLIDPCertificate struct {
	ID          types.String                    `tfsdk:"id"`
	DisplayName types.String                    `tfsdk:"display_name"`
	Cert        customtypes.PEMCertificateValue `tfsdk:"cert"`
	Fingerprint types.String                    `tfsdk:"fingerprint"`
	ExpiresAt   types.String                    `tfsdk:"expires_at"`
	CreatedAt   types.String                    `tfsdk:"created_at"`
	Issuer      types.String                    `tfsdk:"issuer"`
	CertRole    types.String                    `tfsdk:"cert_role"`
}

// samlIDPCertificateEnvelope wraps the given attributes in the JSON:API
// document shape the idp-certificates endpoints expect. Nil values are omitted
// from the request body.
func samlIDPCertificateEnvelope(displayName, cert *string) models.SamlIdpCertificatesEnvelopeable {
	attrs := models.NewSamlIdpCertificates_attributes()
	attrs.SetDisplayName(displayName)
	attrs.SetCert(cert)

	data := models.NewSamlIdpCertificates()
	data.SetTypeEscaped(ptr(models.SAMLIDPCERTIFICATES_SAMLIDPCERTIFICATES_TYPE))
	data.SetAttributes(attrs)

	envelope := models.NewSamlIdpCertificatesEnvelope()
	envelope.SetData(data)
	return envelope
}

// modelFromSAMLIDPCertificate builds the resource model from an API object.
func modelFromSAMLIDPCertificate(data models.SamlIdpCertificatesable) (modelTFESAMLIDPCertificate, error) {
	if data == nil || data.GetAttributes() == nil {
		return modelTFESAMLIDPCertificate{}, errors.New("SAML IdP certificate response did not contain any data")
	}
	attrs := data.GetAttributes()

	return modelTFESAMLIDPCertificate{
		ID:          types.StringValue(valueOrZero(data.GetId())),
		DisplayName: types.StringValue(valueOrZero(attrs.GetDisplayName())),
		Cert:        customtypes.NewPEMCertificateValue(valueOrZero(attrs.GetCert())),
		Fingerprint: types.StringValue(valueOrZero(attrs.GetFingerprint())),
		ExpiresAt:   timeStringOrNull(valueOrZero(attrs.GetExpiresAt()).UTC()),
		CreatedAt:   timeStringOrNull(valueOrZero(attrs.GetCreatedAt()).UTC()),
		Issuer:      types.StringValue(valueOrZero(attrs.GetIssuer())),
		CertRole:    types.StringValue(enumStringOrEmpty(attrs.GetCertRole())),
	}, nil
}

func modelFromSAMLIDPCertificateEnvelope(env models.SamlIdpCertificatesEnvelopeable) (modelTFESAMLIDPCertificate, error) {
	if env == nil {
		return modelTFESAMLIDPCertificate{}, errors.New("SAML IdP certificate response did not contain any data")
	}
	return modelFromSAMLIDPCertificate(env.GetData())
}

// checkSAMLIDPCertificatesSupported records an error when the connected
// Terraform Enterprise predates the idp-certificates API. Shared with the
// SAML IdP certificate data sources.
func checkSAMLIDPCertificatesSupported(c ConfiguredClient, d *diag.Diagnostics) {
	meets, err := c.MeetsMinRemoteTFEVersion(minTFEVersionSAMLIDPCertificates)
	if err != nil {
		d.AddError(
			"Error checking minimum Terraform Enterprise version",
			fmt.Sprintf("Could not determine whether Terraform Enterprise version %s meets the minimum required version %s: %v",
				c.RemoteTFEVersion(), minTFEVersionSAMLIDPCertificates, err),
		)
		return
	}
	if !meets {
		d.AddError(
			"Terraform Enterprise version does not support SAML IdP certificates",
			fmt.Sprintf("Managing SAML IdP certificates requires Terraform Enterprise %s or later. This instance reports %s. Upgrade Terraform Enterprise or use the idp_cert attribute of tfe_saml_settings.",
				minTFEVersionSAMLIDPCertificates, c.RemoteTFEVersion()),
		)
	}
}

// preserveCertFormatting keeps the practitioner's PEM formatting in state when
// the server returned the same certificate body, re-wrapped or re-armored.
func preserveCertFormatting(ctx context.Context, server, prior customtypes.PEMCertificateValue) customtypes.PEMCertificateValue {
	if prior.IsNull() || prior.IsUnknown() {
		return server
	}
	equal, diags := server.StringSemanticEquals(ctx, prior)
	if diags.HasError() || !equal {
		return server
	}
	return prior
}

// Configure implements resource.ResourceWithConfigure
func (r *resourceTFESAMLIDPCertificate) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *resourceTFESAMLIDPCertificate) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_saml_idp_certificate"
}

// Schema implements resource.Resource
func (r *resourceTFESAMLIDPCertificate) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "(Only for Terraform Enterprise) Manages a trusted SAML Identity Provider certificate. Requires Terraform Enterprise v2.1.0 or later and an admin token.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the IdP certificate (idpc-*).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"display_name": schema.StringAttribute{
				Description: "A human-readable label for the certificate.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"cert": schema.StringAttribute{
				Description: "The PEM encoded X.509 certificate as provided by the IdP. Legacy certificates cannot have their body changed.",
				Required:    true,
				CustomType:  customtypes.PEMCertificateType{},
			},
			"fingerprint": schema.StringAttribute{
				Description: "Colon-separated uppercase SHA-256 fingerprint of the certificate.",
				Computed:    true,
			},
			"expires_at": schema.StringAttribute{
				Description: "The time the certificate expires, in RFC 3339 format.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "The time the certificate was added, in RFC 3339 format.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"issuer": schema.StringAttribute{
				Description: "The issuer of the certificate, derived from its O and CN fields.",
				Computed:    true,
			},
			"cert_role": schema.StringAttribute{
				Description: "The role of the certificate: `managed`, `legacy_primary` or `legacy_old`.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// ModifyPlan implements resource.ResourceWithModifyPlan. It keeps the
// certificate-derived attributes stable when the body is unchanged, and rejects
// body changes to legacy certificates at plan time.
func (r *resourceTFESAMLIDPCertificate) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Create or destroy: nothing to carry over.
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	var plan, state modelTFESAMLIDPCertificate
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.Cert.IsUnknown() {
		return
	}

	certChanged := true
	if equal, diags := state.Cert.StringSemanticEquals(ctx, plan.Cert); !diags.HasError() && equal {
		certChanged = false
	}

	if !certChanged {
		plan.Fingerprint = state.Fingerprint
		plan.ExpiresAt = state.ExpiresAt
		plan.Issuer = state.Issuer
		resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
		return
	}

	role := state.CertRole.ValueString()
	if role != "" && role != samlIDPCertRoleManaged {
		resp.Diagnostics.AddAttributeError(
			path.Root("cert"),
			"Cannot change the body of a legacy SAML IdP certificate",
			fmt.Sprintf("Certificate %s has role %q and is owned by the deprecated idp_cert/old_idp_cert attributes of SAML settings. Only display_name can be changed. Rotate it through the idp_cert attribute of tfe_saml_settings, or add a new tfe_saml_idp_certificate and remove this one.",
				state.ID.ValueString(), role),
		)
	}
}

// Create implements resource.Resource
func (r *resourceTFESAMLIDPCertificate) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan modelTFESAMLIDPCertificate
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	checkSAMLIDPCertificatesSupported(r.config, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	body := samlIDPCertificateEnvelope(plan.DisplayName.ValueStringPointer(), plan.Cert.ValueStringPointer())

	tflog.Debug(ctx, "Create SAML IdP certificate")
	env, err := r.config.ClientV2.API.Admin().SamlSettings().IdpCertificates().Post(ctx, body, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error creating SAML IdP certificate", apiErrorDetail(err))
		return
	}

	result, err := modelFromSAMLIDPCertificateEnvelope(env)
	if err != nil {
		resp.Diagnostics.AddError("Error creating SAML IdP certificate", err.Error())
		return
	}
	result.Cert = preserveCertFormatting(ctx, result.Cert, plan.Cert)

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

// Read implements resource.Resource
func (r *resourceTFESAMLIDPCertificate) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state modelTFESAMLIDPCertificate
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	checkSAMLIDPCertificatesSupported(r.config, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	tflog.Debug(ctx, fmt.Sprintf("Read SAML IdP certificate %s", id))
	env, err := r.config.ClientV2.API.Admin().SamlSettings().IdpCertificates().ByExternal_id(id).Get(ctx, nil)
	if err != nil {
		if errors.Is(err, tfe.ErrNotFound) {
			tflog.Debug(ctx, fmt.Sprintf("SAML IdP certificate %s no longer exists", id))
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(fmt.Sprintf("Error reading SAML IdP certificate %s", id), apiErrorDetail(err))
		return
	}

	result, err := modelFromSAMLIDPCertificateEnvelope(env)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Error reading SAML IdP certificate %s", id), err.Error())
		return
	}
	result.Cert = preserveCertFormatting(ctx, result.Cert, state.Cert)

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

// Update implements resource.Resource
func (r *resourceTFESAMLIDPCertificate) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state modelTFESAMLIDPCertificate
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	checkSAMLIDPCertificatesSupported(r.config, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Only send the body when it actually changed, so renaming a legacy
	// certificate does not trip the API's legacy body guard.
	var cert *string
	if equal, diags := state.Cert.StringSemanticEquals(ctx, plan.Cert); diags.HasError() || !equal {
		cert = plan.Cert.ValueStringPointer()
	}
	body := samlIDPCertificateEnvelope(plan.DisplayName.ValueStringPointer(), cert)

	id := state.ID.ValueString()
	tflog.Debug(ctx, fmt.Sprintf("Update SAML IdP certificate %s", id))
	env, err := r.config.ClientV2.API.Admin().SamlSettings().IdpCertificates().ByExternal_id(id).Patch(ctx, body, nil)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Error updating SAML IdP certificate %s", id), apiErrorDetail(err))
		return
	}

	result, err := modelFromSAMLIDPCertificateEnvelope(env)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Error updating SAML IdP certificate %s", id), err.Error())
		return
	}
	result.Cert = preserveCertFormatting(ctx, result.Cert, plan.Cert)

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

// Delete implements resource.Resource
func (r *resourceTFESAMLIDPCertificate) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state modelTFESAMLIDPCertificate
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	checkSAMLIDPCertificatesSupported(r.config, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	tflog.Debug(ctx, fmt.Sprintf("Delete SAML IdP certificate %s", id))
	err := r.config.ClientV2.API.Admin().SamlSettings().IdpCertificates().ByExternal_id(id).Delete(ctx, nil)
	if err != nil {
		if errors.Is(err, tfe.ErrNotFound) {
			tflog.Debug(ctx, fmt.Sprintf("SAML IdP certificate %s no longer exists", id))
			return
		}
		resp.Diagnostics.AddError(fmt.Sprintf("Error deleting SAML IdP certificate %s", id), apiErrorDetail(err))
	}
}

// ImportState implements resource.ResourceWithImportState
func (r *resourceTFESAMLIDPCertificate) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if !strings.HasPrefix(req.ID, "idpc-") {
		resp.Diagnostics.AddError(
			"Invalid SAML IdP certificate ID",
			fmt.Sprintf("Expected an ID of the form idpc-<id>, got %q.", req.ID),
		)
		return
	}
	checkSAMLIDPCertificatesSupported(r.config, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
