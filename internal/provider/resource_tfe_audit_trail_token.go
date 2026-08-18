// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0
package provider

import (
	"context"
	"errors"
	"fmt"
	"time"

	tfev2 "github.com/hashicorp/go-tfe/v2"
	"github.com/hashicorp/go-tfe/v2/api/models"
	orgbuilder "github.com/hashicorp/go-tfe/v2/api/organizations"
	authtokenparams "github.com/hashicorp/go-tfe/v2/api/organizations/item/authenticationtoken"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-tfe/internal/provider/planmodifiers"
)

type resourceAuditTrailToken struct {
	config ConfiguredClient
}

var _ resource.Resource = &resourceAuditTrailToken{}
var _ resource.ResourceWithConfigure = &resourceAuditTrailToken{}
var _ resource.ResourceWithImportState = &resourceAuditTrailToken{}
var _ resource.ResourceWithModifyPlan = &resourceAuditTrailToken{}

func NewAuditTrailTokenResource() resource.Resource {
	return &resourceAuditTrailToken{}
}

type modelTFEAuditTrailTokenV0 struct {
	ID              types.String      `tfsdk:"id"`
	Organization    types.String      `tfsdk:"organization"`
	Token           types.String      `tfsdk:"token"`
	ExpiredAt       timetypes.RFC3339 `tfsdk:"expired_at"`
	ForceRegenerate types.Bool        `tfsdk:"force_regenerate"`
}

func modelFromTFEOrganizationToken(v models.AuthenticationTokensable, organization string, token types.String, forceRegen types.Bool) modelTFEAuditTrailTokenV0 {
	result := modelTFEAuditTrailTokenV0{
		Organization:    types.StringValue(organization),
		ID:              types.StringValue(organization),
		ForceRegenerate: forceRegen,
		Token:           token,
	}

	if attrs := v.GetAttributes(); attrs != nil {
		if expiredAt := attrs.GetExpiredAt(); expiredAt != nil && !expiredAt.IsZero() {
			result.ExpiredAt = timetypes.NewRFC3339TimeValue(*expiredAt)
		}
	}

	return result
}

func (r *resourceAuditTrailToken) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_audit_trail_token"
}

func (r *resourceAuditTrailToken) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// If an audit trail token uses the default organization, then if the default org. changes, it should trigger a modification
	modifyPlanForDefaultOrganizationChange(ctx, r.config.Organization, req.State, req.Config, req.Plan, resp)
}

func (r *resourceAuditTrailToken) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *resourceAuditTrailToken) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Generates a new audit trail token in organization, replacing any existing token." +
			"\n\n-> **Note:** Only organizations that have the [audit-logging entitlement](https://developer.hashicorp.com/terraform/cloud-docs/api-docs#audit-logging) may create audit trail tokens.",
		Version: 0,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID for the audit trail token.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"expired_at": schema.StringAttribute{
				Description: "The token's expiration date. The expiration date must be a date/time string in RFC3339 format (e.g., \"2024-12-31T23:59:59Z\"). If no expiration date is supplied, the token will expire 24 months from creation and a warning during plan and apply phases will be displayed.",
				CustomType:  timetypes.RFC3339Type{},
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
					planmodifiers.WarnIfNullOnCreate(
						"Audit Trail Token expiration null values defaults to 24 months",
					),
				},
			},
			"organization": schema.StringAttribute{
				Description: "Name of the organization. If omitted, organization must be defined in the provider config.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"token": schema.StringAttribute{
				Description: "The authentication token for accessing Audit Trails.",
				Sensitive:   true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"force_regenerate": schema.BoolAttribute{
				Description: "If set to `true`, a new token will be generated even if a token already exists. This will invalidate the existing token!",
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *resourceAuditTrailToken) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state modelTFEAuditTrailTokenV0

	// Read Terraform current state into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var organization string
	resp.Diagnostics.Append(r.config.dataOrDefaultOrganization(ctx, req.State, &organization)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tokenType := authtokenparams.AUDITTRAILS_GETTOKENQUERYPARAMETERTYPE

	tflog.Debug(ctx, "Reading audit trail token")
	tokenEnvelope, err := r.config.ClientV2.API.Organizations().ByOrganization_name(organization).AuthenticationToken().Get(ctx,
		withQueryParams(&orgbuilder.ItemAuthenticationTokenRequestBuilderGetQueryParameters{Token: &tokenType}))
	if err != nil {
		if errors.Is(err, tfev2.ErrNotFound) {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError("Error reading Organization Audit Trail Token", "Could not read Organization Audit Trail Token, unexpected error: "+err.Error())
		}
		return
	}
	if tokenEnvelope == nil || tokenEnvelope.GetData() == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	result := modelFromTFEOrganizationToken(tokenEnvelope.GetData(), organization, state.Token, state.ForceRegenerate)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *resourceAuditTrailToken) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan modelTFEAuditTrailTokenV0

	// Read Terraform planned changes into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var organization string
	resp.Diagnostics.Append(r.config.dataOrDefaultOrganization(ctx, req.Plan, &organization)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if an audit trail token already exists for the organization and only
	// continue if the force_regenerate flag is set.
	tflog.Debug(ctx, fmt.Sprintf("Check if an audit trail token already exists for organization: %s", organization))
	checkTokenType := authtokenparams.AUDITTRAILS_GETTOKENQUERYPARAMETERTYPE
	existingEnvelope, err := r.config.ClientV2.API.Organizations().ByOrganization_name(organization).AuthenticationToken().Get(ctx,
		withQueryParams(&orgbuilder.ItemAuthenticationTokenRequestBuilderGetQueryParameters{Token: &checkTokenType}))
	if err != nil && !errors.Is(err, tfev2.ErrNotFound) {
		resp.Diagnostics.AddError("Error while checking if an audit token exists for organization", fmt.Sprintf("error checking if an audit token exists for organization %s: %s", organization, err))
		return
	}
	if err == nil && existingEnvelope != nil && existingEnvelope.GetData() != nil {
		if !plan.ForceRegenerate.ValueBool() {
			resp.Diagnostics.AddError("An audit trail token already exists", fmt.Sprintf("an audit trail token already exists for organization: %s", organization))
			return
		}
		tflog.Debug(ctx, fmt.Sprintf("Regenerating existing audit trail token for organization: %s", organization))
	}

	// Build request body with optional expiry
	attrs := models.NewAuthenticationTokens_attributes()
	expireString := plan.ExpiredAt.ValueString()
	if expireString != "" {
		expiry, err := time.Parse(time.RFC3339, expireString)
		if err != nil {
			resp.Diagnostics.AddError("Invalid date", fmt.Sprintf("%s must be a valid date or time, provided in iso8601 format", expireString))
			return
		}
		attrs.SetExpiredAt(&expiry)
	}

	tokenDataType := models.AUTHENTICATIONTOKENS_AUTHENTICATIONTOKENS_TYPE
	tokenData := models.NewAuthenticationTokens()
	tokenData.SetTypeEscaped(&tokenDataType)
	tokenData.SetAttributes(attrs)

	envelope := models.NewAuthenticationTokensEnvelope()
	envelope.SetData(tokenData)

	postTokenType := authtokenparams.AUDITTRAILS_POSTTOKENQUERYPARAMETERTYPE
	tflog.Debug(ctx, fmt.Sprintf("Create audit trail token for organization %s", organization))
	tokenEnvelope, err := r.config.ClientV2.API.Organizations().ByOrganization_name(organization).AuthenticationToken().Post(ctx, envelope,
		withQueryParams(&orgbuilder.ItemAuthenticationTokenRequestBuilderPostQueryParameters{Token: &postTokenType}))
	if err != nil {
		resp.Diagnostics.AddError("Unable to create organization audit trail token", err.Error())
		return
	}
	if tokenEnvelope == nil || tokenEnvelope.GetData() == nil {
		resp.Diagnostics.AddError("Unable to create organization audit trail token", "no data was returned by the API")
		return
	}

	tokenStr := types.StringNull()
	if attrs := tokenEnvelope.GetData().GetAttributes(); attrs != nil {
		if t := attrs.GetToken(); t != nil {
			tokenStr = types.StringValue(*t)
		}
	}

	result := modelFromTFEOrganizationToken(tokenEnvelope.GetData(), organization, tokenStr, plan.ForceRegenerate)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *resourceAuditTrailToken) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Audit trail tokens cannot be updated", "Audit trail tokens cannot be updated. Please regenerate token.")
}

func (r *resourceAuditTrailToken) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state modelTFEAuditTrailTokenV0
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var organization string
	resp.Diagnostics.Append(r.config.dataOrDefaultOrganization(ctx, req.State, &organization)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteTokenType := authtokenparams.AUDITTRAILS_DELETETOKENQUERYPARAMETERTYPE

	tflog.Debug(ctx, fmt.Sprintf("Delete organization audit trail token %s", organization))
	err := r.config.ClientV2.API.Organizations().ByOrganization_name(organization).AuthenticationToken().Delete(ctx,
		withQueryParams(&orgbuilder.ItemAuthenticationTokenRequestBuilderDeleteQueryParameters{Token: &deleteTokenType}))
	// Ignore 404s for delete
	if err != nil && !errors.Is(err, tfev2.ErrNotFound) {
		resp.Diagnostics.AddError(
			"Error deleting organization audit trail token",
			fmt.Sprintf("Couldn't delete organization audit trail token %s: %s", organization, err.Error()),
		)
	}
	// Resource is implicitly deleted from resp.State if diagnostics have no errors.
}

func (r *resourceAuditTrailToken) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	organization := req.ID

	tokenType := authtokenparams.AUDITTRAILS_GETTOKENQUERYPARAMETERTYPE

	tflog.Debug(ctx, "Reading audit trail token")
	tokenEnvelope, err := r.config.ClientV2.API.Organizations().ByOrganization_name(organization).AuthenticationToken().Get(ctx,
		withQueryParams(&orgbuilder.ItemAuthenticationTokenRequestBuilderGetQueryParameters{Token: &tokenType}))
	if err != nil {
		resp.Diagnostics.AddError("Error importing organization audit trail token", err.Error())
		return
	}
	if tokenEnvelope == nil || tokenEnvelope.GetData() == nil {
		resp.Diagnostics.AddError(
			"Error importing organization audit trail token",
			"Audit trail token does not exist or has no details",
		)
		return
	}

	result := modelFromTFEOrganizationToken(tokenEnvelope.GetData(), organization, basetypes.NewStringNull(), basetypes.NewBoolNull())
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}
