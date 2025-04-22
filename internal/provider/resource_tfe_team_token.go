// // Copyright (c) HashiCorp, Inc.
// // SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework-validators/boolvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.ResourceWithConfigure   = &resourceTFETeamToken{}
	_ resource.ResourceWithImportState = &resourceTFETeamToken{}
)

func NewTeamTokenResource() resource.Resource {
	return &resourceTFETeamToken{}
}

type resourceTFETeamToken struct {
	config ConfiguredClient
}

type modelTFETeamToken struct {
	ID              types.String `tfsdk:"id"`
	TeamID          types.String `tfsdk:"team_id"`
	ForceRegenerate types.Bool   `tfsdk:"force_regenerate"`
	Token           types.String `tfsdk:"token"`
	ExpiredAt       types.String `tfsdk:"expired_at"`
	Description     types.String `tfsdk:"description"`
}

func (r *resourceTFETeamToken) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Early exit if provider is unconfigured (i.e. we're only validating config or something)
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

func (r *resourceTFETeamToken) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team_token"
}

func (r *resourceTFETeamToken) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the token",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"team_id": schema.StringAttribute{
				Description: "ID of the team.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"force_regenerate": schema.BoolAttribute{
				Description: "When set to true will force the audit trail token to be recreated.",
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
				Validators: []validator.Bool{
					boolvalidator.ConflictsWith(path.MatchRoot("description")),
				},
			},
			"token": schema.StringAttribute{
				Description: "The generated token.",
				Computed:    true,
				Sensitive:   true,
			},
			"expired_at": schema.StringAttribute{
				Description: "The token's expiration date.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Description: "The description of the token, which must be unique per team.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("force_regenerate")),
				},
			},
		},
		Description: "Generates a new team token. If no description is provided, it follows the legacy behavior to override the existing, descriptionless token if one exists.",
	}
}

func (r *resourceTFETeamToken) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan modelTFETeamToken
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	teamID := plan.TeamID.ValueString()
	if plan.Description.IsNull() {
		// No description indicates legacy behavior where token will be regenerated if it does not exist
		tflog.Debug(ctx, fmt.Sprintf("Check if a token already exists for team: %s", teamID))
		_, err := r.config.Client.TeamTokens.Read(ctx, teamID)
		if err != nil && !errors.Is(err, tfe.ErrResourceNotFound) {
			resp.Diagnostics.AddError(
				fmt.Sprintf("Error checking if a token exists for team %s", teamID),
				err.Error(),
			)
			return
		}
		if err == nil {
			if !plan.ForceRegenerate.ValueBool() {
				resp.Diagnostics.AddError(
					fmt.Sprintf("A token already exists for team: %s", teamID),
					"Set force_regenerate to true to regenerate the token.",
				)
				return
			}
			tflog.Debug(ctx, fmt.Sprintf("Regenerating existing token for team: %s", teamID))
		}
	}

	expiredAt := plan.ExpiredAt.ValueString()
	options := tfe.TeamTokenCreateOptions{
		Description: plan.Description.ValueStringPointer(),
	}
	if !plan.ExpiredAt.IsNull() && expiredAt != "" {
		expiry, err := time.Parse(time.RFC3339, expiredAt)
		if err != nil {
			resp.Diagnostics.AddError(
				fmt.Sprintf("%s must be a valid date or time, provided in iso8601 format", expiredAt),
				err.Error(),
			)
			return
		}
		options.ExpiredAt = &expiry
	}

	token, err := r.config.Client.TeamTokens.CreateWithOptions(ctx, teamID, options)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error creating new token for team %s", teamID),
			err.Error(),
		)
		return
	}

	result := modelFromTFEToken(plan.TeamID, types.StringValue(token.ID), types.StringValue(token.Token), plan.ForceRegenerate, plan.ExpiredAt, plan.Description)
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

func modelFromTFEToken(teamID types.String, tokenID types.String, stateValue types.String, forceRegenerate types.Bool, expiredAt types.String, description types.String) modelTFETeamToken {
	m := modelTFETeamToken{
		TeamID:          teamID,
		ForceRegenerate: forceRegenerate,
		ExpiredAt:       types.StringNull(),
		Token:           stateValue,
		Description:     types.StringNull(),
	}
	if !expiredAt.IsNull() {
		m.ExpiredAt = expiredAt
	}

	if !description.IsNull() {
		m.Description = description
		m.ID = tokenID
	} else {
		m.ID = teamID
	}

	return m
}

func (r *resourceTFETeamToken) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state modelTFETeamToken
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	teamID := state.TeamID.ValueString()
	tflog.Debug(ctx, fmt.Sprintf("Read the token from team: %s", teamID))
	var err error
	if isTokenID(state.ID.ValueString()) {
		_, err = r.config.Client.TeamTokens.ReadByID(ctx, state.ID.ValueString())
	} else {
		_, err = r.config.Client.TeamTokens.Read(ctx, teamID)
	}
	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
			tflog.Debug(ctx, fmt.Sprintf("Token for team %s no longer exists", teamID))
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error reading token from team %s", teamID),
			err.Error(),
		)
		return
	}
	result := modelFromTFEToken(state.TeamID, state.ID, state.Token, state.ForceRegenerate, state.ExpiredAt, state.Description)
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

func (r *resourceTFETeamToken) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// This should never be called, based on the schema
	resp.Diagnostics.AddError("Update not supported.", "Please recreate the resource")
}

func (r *resourceTFETeamToken) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state modelTFETeamToken
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	teamID := state.TeamID.ValueString()
	tflog.Debug(ctx, fmt.Sprintf("Delete the token from team: %s", teamID))
	var err error
	if isTokenID(state.ID.ValueString()) {
		err = r.config.Client.TeamTokens.DeleteByID(ctx, state.ID.ValueString())
	} else {
		err = r.config.Client.TeamTokens.Delete(ctx, teamID)
	}
	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
			tflog.Debug(ctx, fmt.Sprintf("Token for team %s no longer exists", teamID))
			return
		}
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error deleting token from team %s", teamID),
			err.Error(),
		)
	}
}

func (r *resourceTFETeamToken) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("team_id"), req, resp)
}

// Determines whether the ID of the resource is the ID of the authentication token
// or the ID of the team the token belongs to.
func isTokenID(id string) bool {
	return strings.HasPrefix(id, "at-")
}
