// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	tfev2 "github.com/hashicorp/go-tfe/v2"
	"github.com/hashicorp/go-tfe/v2/api/models"
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
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-tfe/internal/provider/planmodifiers"
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
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
					planmodifiers.WarnIfNullOnCreate(
						"Team Token expiration null values defaults to 24 months",
					),
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
	hasDescription := !plan.Description.IsNull()

	if !hasDescription {
		// No description indicates legacy behavior where token will be regenerated if it does not exist
		tflog.Debug(ctx, fmt.Sprintf("Check if a token already exists for team: %s", teamID))
		_, err := r.config.ClientV2.API.Teams().ById(teamID).AuthenticationToken().Get(ctx, nil)
		if err != nil && !errors.Is(err, tfev2.ErrNotFound) {
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

	expiredAtString := plan.ExpiredAt.ValueString()
	var expiry *time.Time
	if !plan.ExpiredAt.IsNull() && expiredAtString != "" {
		parsed, err := time.Parse(time.RFC3339, expiredAtString)
		if err != nil {
			resp.Diagnostics.AddError(
				fmt.Sprintf("%s must be a valid date or time, provided in iso8601 format", expiredAtString),
				err.Error(),
			)
			return
		}
		expiry = &parsed
	}

	var tokenID, tokenValue string
	var tokenExpiredAt time.Time

	if hasDescription {
		// go-tfe v2 does not generate a route for POST
		// /teams/{id}/authentication-tokens (plural), which Atlas requires
		// to create more than one named/described token per team - the
		// generated client only exposes the singular, legacy
		// /teams/{id}/authentication-token route. This call remains on
		// go-tfe v1 until that route is generated. Every other operation on
		// this resource (the description-less create path below, read, and
		// delete) uses the go-tfe v2 client.
		options := tfe.TeamTokenCreateOptions{
			Description: plan.Description.ValueStringPointer(),
		}
		if expiry != nil {
			options.ExpiredAt = expiry
		}
		token, err := r.config.Client.TeamTokens.CreateWithOptions(ctx, teamID, options)
		if err != nil {
			errDetails := err.Error()
			if errors.Is(err, tfe.ErrResourceNotFound) {
				errDetails = fmt.Sprintf("%s, team does not exist or version of Terraform Enterprise "+
					"does not support multiple team tokens with descriptions", errDetails)
			}
			resp.Diagnostics.AddError(
				fmt.Sprintf("Error creating new token for team %s", teamID),
				errDetails,
			)
			return
		}
		tokenID = token.ID
		tokenValue = token.Token
		tokenExpiredAt = token.ExpiredAt
	} else {
		attributes := models.NewAuthenticationTokens_attributes()
		if expiry != nil {
			attributes.SetExpiredAt(expiry)
		}
		authToken := models.NewAuthenticationTokens()
		authToken.SetTypeEscaped(ptr(models.AUTHENTICATIONTOKENS_AUTHENTICATIONTOKENS_TYPE))
		authToken.SetAttributes(attributes)
		envelope := models.NewAuthenticationTokensEnvelope()
		envelope.SetData(authToken)

		created, err := r.config.ClientV2.API.Teams().ById(teamID).AuthenticationToken().Post(ctx, envelope, nil)
		if err != nil {
			resp.Diagnostics.AddError(
				fmt.Sprintf("Error creating new token for team %s", teamID),
				err.Error(),
			)
			return
		}
		data := created.GetData()
		if len(data) == 0 {
			resp.Diagnostics.AddError(
				fmt.Sprintf("Error creating new token for team %s", teamID),
				"no data returned by the API",
			)
			return
		}

		tokenID = valueOrZero(data[0].GetId())
		tokenValue = valueOrZero(data[0].GetAttributes().GetToken())
		if ea := data[0].GetAttributes().GetExpiredAt(); ea != nil {
			tokenExpiredAt = *ea
		}
	}

	var expiredAtValue types.String
	if !tokenExpiredAt.IsZero() {
		expiredAtValue = types.StringValue(tokenExpiredAt.Format(time.RFC3339))
	} else {
		expiredAtValue = types.StringNull()
	}

	result := modelFromTFEToken(plan.TeamID, types.StringValue(tokenID), types.StringValue(tokenValue), plan.ForceRegenerate, expiredAtValue, plan.Description)
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

	var tokenExpiredAt *time.Time
	if isTokenID(state.ID.ValueString()) {
		result, err := r.config.ClientV2.API.AuthenticationTokens().ById(state.ID.ValueString()).Get(ctx, nil)
		if err != nil {
			if errors.Is(err, tfev2.ErrNotFound) {
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
		if result != nil && result.GetData() != nil {
			tokenExpiredAt = result.GetData().GetAttributes().GetExpiredAt()
		}
	} else {
		result, err := r.config.ClientV2.API.Teams().ById(teamID).AuthenticationToken().Get(ctx, nil)
		if err != nil {
			if errors.Is(err, tfev2.ErrNotFound) {
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
		if result != nil && result.GetData() != nil {
			tokenExpiredAt = result.GetData().GetAttributes().GetExpiredAt()
		}
	}

	// if expired_at was set to null at creation, the API returns a default value of 24 months from the creation date.
	expiredAt := types.StringNull()
	if tokenExpiredAt != nil && !tokenExpiredAt.IsZero() {
		expiredAt = types.StringValue(tokenExpiredAt.Format(time.RFC3339))
	}

	result := modelFromTFEToken(state.TeamID, state.ID, state.Token, state.ForceRegenerate, expiredAt, state.Description)
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
		err = r.config.ClientV2.API.AuthenticationTokens().ById(state.ID.ValueString()).Delete(ctx, nil)
	} else {
		err = r.config.ClientV2.API.Teams().ById(teamID).AuthenticationToken().Delete(ctx, nil)
	}
	if err != nil {
		if errors.Is(err, tfev2.ErrNotFound) {
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
	if !isTokenID(req.ID) {
		// Set the team ID field
		resource.ImportStatePassthroughID(ctx, path.Root("team_id"), req, resp)
		return
	}

	// go-tfe v2's generated AuthenticationTokens relationship model only
	// exposes a "related" link for the owning team
	// (AuthenticationTokens_relationships.GetTeam() returns
	// Links_relatedable), not the relationship's JSON:API resource
	// identifier (id). Import-by-token-ID needs the team ID to populate
	// team_id, so this lookup remains on go-tfe v1 until go-tfe generates
	// that relationship's "data" member. This is the only remaining go-tfe
	// v1 call in this resource; create, read, and delete all use the
	// go-tfe v2 client.
	//
	// Fetch token by ID to set attributes
	token, err := r.config.Client.TeamTokens.ReadByID(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing team token", err.Error())
		return
	}
	if token.Team == nil {
		resp.Diagnostics.AddError("Error importing team token", "token did not return associated team")
		return
	}

	var expiredAt types.String
	if !token.ExpiredAt.IsZero() {
		expiredAt = types.StringValue(token.ExpiredAt.Format(time.RFC3339))
	} else {
		expiredAt = types.StringNull()
	}

	var description types.String
	if token.Description != nil {
		description = types.StringValue(*token.Description)
	} else {
		description = types.StringNull()
	}

	result := modelFromTFEToken(types.StringValue(token.Team.ID), types.StringValue(token.ID), types.StringValue(token.Token), basetypes.NewBoolNull(), expiredAt, description)
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

// Determines whether the ID of the resource is the ID of the authentication token
// or the ID of the team the token belongs to.
func isTokenID(id string) bool {
	return strings.HasPrefix(id, "at-")
}
