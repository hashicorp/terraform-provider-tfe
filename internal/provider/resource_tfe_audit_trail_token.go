// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package provider

import (
	"context"
	"errors"
	"fmt"
	"time"

	tfe "github.com/hashicorp/go-tfe"
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

func modelFromTFEOrganizationToken(v *tfe.OrganizationToken, organization string, token types.String, forceRegen types.Bool) modelTFEAuditTrailTokenV0 {
	result := modelTFEAuditTrailTokenV0{
		Organization:    types.StringValue(organization),
		ID:              types.StringValue(organization),
		ForceRegenerate: forceRegen,
		Token:           token,
	}

	if !v.ExpiredAt.IsZero() {
		result.ExpiredAt = timetypes.NewRFC3339TimeValue(v.ExpiredAt)
	}

	return result
}

func (r *resourceAuditTrailToken) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_audit_trail_token"
}

func (r *resourceAuditTrailToken) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// If an audit trail token uses the default organization, then if the deafault org. changes, it should trigger a modification
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
		Version: 0,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Service-generated identifier for the token",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"expired_at": schema.StringAttribute{
				Description: "The time when the audit trail token will expire. This must be a valid ISO8601 timestamp.",
				CustomType:  timetypes.RFC3339Type{},
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
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
				Description: "When set to true will force the audit trail token to be recreated.",
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

	tokenType := tfe.AuditTrailToken

	tflog.Debug(ctx, "Reading audit trail token")
	token, err := r.config.Client.OrganizationTokens.ReadWithOptions(ctx, organization, tfe.OrganizationTokenReadOptions{TokenType: &tokenType})
	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError("Error reading Organization Audit Trail Token", "Could not read Organization Audit Trail Token, unexpected error: "+err.Error())
		}
		return
	}

	result := modelFromTFEOrganizationToken(token, organization, state.Token, state.ForceRegenerate)

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

	tokenType := tfe.AuditTrailToken

	var organization string
	resp.Diagnostics.Append(r.config.dataOrDefaultOrganization(ctx, req.Plan, &organization)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Check if an audit trail token already exists for the organization and only
	// continue if the force_regenerate flag is set.
	tflog.Debug(ctx, fmt.Sprintf("Check if an audit trail token already exists for organization: %s", organization))
	if token, err := r.config.Client.OrganizationTokens.ReadWithOptions(ctx, organization, tfe.OrganizationTokenReadOptions{TokenType: &tokenType}); err != nil {
		if !errors.Is(err, tfe.ErrResourceNotFound) {
			resp.Diagnostics.AddError("Error while checking if an audit token exists for organization", fmt.Sprintf("error checking if an audit token exists for organization %s: %s", organization, err))
			return
		}
	} else if token != nil {
		if !plan.ForceRegenerate.ValueBool() {
			resp.Diagnostics.AddError("An audit trail token already exists", fmt.Sprintf("an audit trail token already exists for organization: %s", organization))
			return
		}
		tflog.Debug(ctx, fmt.Sprintf("Regenerating existing audit trail token for organization: %s", organization))
	}

	options := tfe.OrganizationTokenCreateOptions{
		TokenType: &tokenType,
	}

	// Optional ExpiryAt
	expireString := plan.ExpiredAt.ValueString()
	if expireString != "" {
		expiry, err := time.Parse(time.RFC3339, expireString)
		if err != nil {
			resp.Diagnostics.AddError("Invalid date", fmt.Sprintf("%s must be a valid date or time, provided in iso8601 format", expireString))
			return
		}
		options.ExpiredAt = &expiry
	}

	tflog.Debug(ctx, fmt.Sprintf("Create audit trail token for organization %s", organization))
	token, err := r.config.Client.OrganizationTokens.CreateWithOptions(ctx, organization, options)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create organization audit trail token", err.Error())
		return
	}

	result := modelFromTFEOrganizationToken(token, organization, types.StringValue(token.Token), plan.ForceRegenerate)

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
	tokenType := tfe.AuditTrailToken

	options := tfe.OrganizationTokenDeleteOptions{
		TokenType: &tokenType,
	}

	tflog.Debug(ctx, fmt.Sprintf("Delete organization audit trail token %s", organization))
	err := r.config.Client.OrganizationTokens.DeleteWithOptions(ctx, organization, options)
	// Ignore 404s for delete
	if err != nil && !errors.Is(err, tfe.ErrResourceNotFound) {
		resp.Diagnostics.AddError(
			"Error deleting organization audit trail token",
			fmt.Sprintf("Couldn't delete organization audit trail token %s: %s", organization, err.Error()),
		)
	}
	// Resource is implicitly deleted from resp.State if diagnostics have no errors.
}

func (r *resourceAuditTrailToken) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	organization := req.ID

	tokenType := tfe.AuditTrailToken

	tflog.Debug(ctx, "Reading audit trail token")
	if token, err := r.config.Client.OrganizationTokens.ReadWithOptions(ctx, organization, tfe.OrganizationTokenReadOptions{TokenType: &tokenType}); err != nil {
		resp.Diagnostics.AddError("Error importing organization audit trail token", err.Error())
	} else if token == nil {
		resp.Diagnostics.AddError(
			"Error importing organization audit trail token",
			"Audit trail token does not exist or has no details",
		)
	} else {
		result := modelFromTFEOrganizationToken(token, organization, basetypes.NewStringNull(), basetypes.NewBoolNull())
		resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
	}
}
