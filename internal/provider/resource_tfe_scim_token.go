// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-tfe/internal/provider/planmodifiers"
)

type modelTFESCIMToken struct {
	ID          types.String `tfsdk:"id"`
	Description types.String `tfsdk:"description"`
	Token       types.String `tfsdk:"token"`
	ExpiredAt   types.String `tfsdk:"expired_at"`
	CreatedAt   types.String `tfsdk:"created_at"`
	LastUsedAt  types.String `tfsdk:"last_used_at"`
}

// scimTokenExpiredAtUserSetKey records, in private state, whether the user
// explicitly set expired_at on the last Create. Without this marker we can't
// tell "user removed expired_at" from "user never set it" at plan time —
// both look like ConfigValue=null with a non-null Computed state.
const scimTokenExpiredAtUserSetKey = "expired_at_user_set"

// replaceWhenExpiredAtRemoved forces replacement when the user drops
// expired_at from config after previously setting it.
type replaceWhenExpiredAtRemoved struct{}

var _ planmodifier.String = replaceWhenExpiredAtRemoved{}

func (m replaceWhenExpiredAtRemoved) Description(_ context.Context) string {
	return "Replaces the resource when expired_at is removed from config after being set."
}

func (m replaceWhenExpiredAtRemoved) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m replaceWhenExpiredAtRemoved) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// Skip on create and destroy.
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}
	// Value changes are handled by the built-in RequiresReplace.
	if !req.ConfigValue.IsNull() {
		return
	}

	marker, diags := req.Private.GetKey(ctx, scimTokenExpiredAtUserSetKey)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// No marker -> user never set expired_at; leave the plan alone.
	if len(marker) == 0 {
		return
	}

	// UseStateForUnknown ran earlier in the modifier chain and copied state
	// into the plan; reset to Unknown so the framework sees a real diff and
	// honors RequiresReplace.
	resp.PlanValue = types.StringUnknown()
	resp.RequiresReplace = true
}

// resourceTFESCIMToken implements the tfe_scim_token resource type
type resourceTFESCIMToken struct {
	client *tfe.Client
}

// modelFromTFEAdminSCIMToken builds a modelTFESCIMToken struct from a tfe.AdminSCIMToken value
func modelFromTFEAdminSCIMToken(v tfe.AdminSCIMToken) modelTFESCIMToken {
	m := modelTFESCIMToken{
		ID:          types.StringValue(v.ID),
		Description: types.StringValue(v.Description),
		Token:       types.StringValue(v.Token),
		ExpiredAt:   timeStringOrNull(v.ExpiredAt),
		CreatedAt:   timeStringOrNull(v.CreatedAt),
		LastUsedAt:  timeStringOrNull(v.LastUsedAt),
	}
	return m
}

var (
	_ resource.Resource                = &resourceTFESCIMToken{}
	_ resource.ResourceWithConfigure   = &resourceTFESCIMToken{}
	_ resource.ResourceWithImportState = &resourceTFESCIMToken{}
)

// NewSCIMTokenResource is a resource function for the framework provider.
func NewSCIMTokenResource() resource.Resource {
	return &resourceTFESCIMToken{}
}

// Metadata implements resource.Resource
func (r *resourceTFESCIMToken) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_scim_token"
}

// Configure implements resource.ResourceWithConfigure
func (r *resourceTFESCIMToken) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Early exit if provider is not properly configured (i.e. we're only validating config or something)
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(ConfiguredClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected resource Configure type",
			fmt.Sprintf("Expected tfe.ConfiguredClient, got %T. This is a bug in the tfe provider, so please report it on GitHub.", req.ProviderData),
		)
		return
	}
	r.client = client.Client
}

func (r *resourceTFESCIMToken) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a resource which manages TFE SCIM tokens. These tokens are used for authentication when using the TFE SCIM API.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the SCIM token.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "The description of the SCIM token.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"token": schema.StringAttribute{
				Description: "The value of the SCIM token. This is only set on creation and cannot be read afterwards for security purposes.",
				Computed:    true,
				Sensitive:   true,
			},
			"expired_at": schema.StringAttribute{
				Description: "The time when the SCIM token expires.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					// Order matters: UseStateForUnknown first (keeps re-applies
					// no-ops), then the removal check, then RequiresReplace.
					stringplanmodifier.UseStateForUnknown(),
					replaceWhenExpiredAtRemoved{},
					stringplanmodifier.RequiresReplace(),
					planmodifiers.WarnIfNullOnCreate(
						"SCIM token expiration defaults to 365 days when unset.",
					),
				},
			},
			"created_at": schema.StringAttribute{
				Description: "The time when the SCIM token was created.",
				Computed:    true,
			},
			"last_used_at": schema.StringAttribute{
				Description: "The time when the SCIM token was last used.",
				Computed:    true,
			},
		},
	}
}

// Read implements resource.Resource
func (r *resourceTFESCIMToken) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state modelTFESCIMToken
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	scimTokenID := state.ID.ValueString()

	var scimToken *tfe.AdminSCIMToken
	var err error

	tflog.Debug(ctx, "Reading SCIM Token")
	scimToken, err = r.client.Admin.Settings.SCIM.Tokens.Read(ctx, scimTokenID)

	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
			tflog.Debug(ctx, fmt.Sprintf("Token ID %s, no longer exists", scimTokenID))
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error reading SCIM Token of ID %s", scimTokenID), "Could not read SCIM Token, unexpected error: "+err.Error(),
		)
		return
	}

	result := modelTFESCIMToken{
		ID:          types.StringValue(scimToken.ID),
		Description: types.StringValue(scimToken.Description),
		Token:       state.Token, // Token is only returned at creation; preserve existing state value
		ExpiredAt:   timeStringOrNull(scimToken.ExpiredAt),
		CreatedAt:   timeStringOrNull(scimToken.CreatedAt),
		LastUsedAt:  timeStringOrNull(scimToken.LastUsedAt),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

func (r *resourceTFESCIMToken) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan modelTFESCIMToken
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating SCIM Token")

	expiredAt := plan.ExpiredAt.ValueString()
	options := tfe.AdminSCIMTokenCreateOptions{
		Description: plan.Description.ValueStringPointer(),
	}

	if !plan.ExpiredAt.IsNull() && !plan.ExpiredAt.IsUnknown() {
		if expiredAt == "" {
			resp.Diagnostics.AddError(
				"Invalid expired_at value",
				`expired_at must be omitted or set to a non-empty RFC3339/iso8601 format timestamp; use null or remove the attribute instead of "".`,
			)
			return
		}

		expiry, err := time.Parse(time.RFC3339, expiredAt)
		if err != nil {
			resp.Diagnostics.AddError(
				fmt.Sprintf("%s must be a valid date or time, provided in RFC3339/iso8601 format", expiredAt),
				err.Error(),
			)
			return
		}
		options.ExpiredAt = &expiry
	}

	scimToken, err := r.client.Admin.Settings.SCIM.Tokens.CreateWithOptions(ctx, options)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating new SCIM token",
			err.Error(),
		)
		return
	}

	result := modelFromTFEAdminSCIMToken(*scimToken)
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)

	// Track whether the user set expired_at so the plan modifier can detect
	// a later removal. Empty bytes clear the key; "1" is an arbitrary
	// non-empty sentinel (the modifier only checks for length > 0).
	marker := []byte("")
	if !plan.ExpiredAt.IsNull() && plan.ExpiredAt.ValueString() != "" {
		marker = []byte("1")
	}
	resp.Diagnostics.Append(resp.Private.SetKey(ctx, scimTokenExpiredAtUserSetKey, marker)...)
}

func (r *resourceTFESCIMToken) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	// This should never be called, based on the schema
	resp.Diagnostics.AddError("Update not supported.", "Please recreate the resource")
}

func (r *resourceTFESCIMToken) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state modelTFESCIMToken
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	scimTokenID := state.ID.ValueString()

	tflog.Debug(ctx, fmt.Sprintf("Deleting SCIM Token with ID %s", scimTokenID))
	err := r.client.Admin.Settings.SCIM.Tokens.Delete(ctx, scimTokenID)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error deleting SCIM Token with ID %s", scimTokenID),
			"Could not delete SCIM Token, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *resourceTFESCIMToken) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if !isTokenID(req.ID) {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier to be a SCIM token ID, got: %s. Please use the format `terraform import tfe_scim_token.<name> <token_id>`.", req.ID),
		)
		return
	}

	scimTokenID := req.ID
	scimToken, err := r.client.Admin.Settings.SCIM.Tokens.Read(ctx, scimTokenID)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error reading SCIM Token with ID %s", scimTokenID),
			"Could not read SCIM Token, unexpected error: "+err.Error(),
		)
		return
	}

	// Token is only returned at creation time and is not available via the read endpoint.
	result := modelTFESCIMToken{
		ID:          types.StringValue(scimToken.ID),
		Description: types.StringValue(scimToken.Description),
		Token:       types.StringNull(),
		ExpiredAt:   timeStringOrNull(scimToken.ExpiredAt),
		CreatedAt:   timeStringOrNull(scimToken.CreatedAt),
		LastUsedAt:  timeStringOrNull(scimToken.LastUsedAt),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}
