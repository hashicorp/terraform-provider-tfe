// // Copyright IBM Corp. 2018, 2025
// // SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	tfev2 "github.com/hashicorp/go-tfe/v2"
	tfev2api "github.com/hashicorp/go-tfe/v2/api"
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
	api := r.config.ClientV2.API
	legacy := plan.Description.IsNull()
	if legacy {
		// No description indicates legacy behavior where token will be regenerated if it does not exist
		tflog.Debug(ctx, fmt.Sprintf("Check if a token already exists for team: %s", teamID))
		_, err := api.Teams().ById(teamID).AuthenticationToken().Get(ctx, nil)
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

	expiredAt := plan.ExpiredAt.ValueString()
	var expiry *time.Time
	if !plan.ExpiredAt.IsNull() && expiredAt != "" {
		parsed, err := time.Parse(time.RFC3339, expiredAt)
		if err != nil {
			resp.Diagnostics.AddError(
				fmt.Sprintf("%s must be a valid date or time, provided in iso8601 format", expiredAt),
				err.Error(),
			)
			return
		}
		expiry = &parsed
	}

	var tokenID, tokenValue string
	var expiredAtValue types.String
	var err error
	if legacy {
		tokenID, tokenValue, expiredAtValue, err = createLegacyTeamToken(ctx, api, teamID, expiry)
	} else {
		tokenID, tokenValue, expiredAtValue, err = createDescribedTeamToken(ctx, r.config.Client, teamID, plan.Description.ValueStringPointer(), expiry)
	}
	if err != nil {
		errDetails := err.Error()
		if errors.Is(err, tfev2.ErrNotFound) || errors.Is(err, tfe.ErrResourceNotFound) {
			errDetails = fmt.Sprintf("%s, team does not exist or version of Terraform Enterprise "+
				"does not support multiple team tokens with descriptions", errDetails)
		}
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error creating new token for team %s", teamID),
			errDetails,
		)
		return
	}

	result := modelFromTFEToken(plan.TeamID, types.StringValue(tokenID), types.StringValue(tokenValue), plan.ForceRegenerate, expiredAtValue, plan.Description)
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

// createLegacyTeamToken creates (or, if one already exists, regenerates) the team's
// single, descriptionless authentication token on the singular "authentication-token"
// endpoint.
func createLegacyTeamToken(ctx context.Context, api *tfev2api.ApiClient, teamID string, expiry *time.Time) (id string, value string, expiredAt types.String, err error) {
	expiredAt = types.StringNull()

	attributes := models.NewAuthenticationTokens_attributes()
	attributes.SetExpiredAt(expiry)

	tokenType := models.AUTHENTICATIONTOKENS_AUTHENTICATIONTOKENS_TYPE
	tokenData := models.NewAuthenticationTokens()
	tokenData.SetTypeEscaped(&tokenType)
	tokenData.SetAttributes(attributes)
	envelope := models.NewAuthenticationTokensEnvelope()
	envelope.SetData(tokenData)

	tokenEnvelope, err := api.Teams().ById(teamID).AuthenticationToken().Post(ctx, envelope, nil)
	if err != nil {
		return "", "", expiredAt, err
	}
	if tokenEnvelope == nil || tokenEnvelope.GetData() == nil || tokenEnvelope.GetData().GetId() == nil {
		return "", "", expiredAt, errors.New("API response did not include the created token")
	}

	token := tokenEnvelope.GetData()
	id = *token.GetId()
	if attrs := token.GetAttributes(); attrs != nil {
		if v := attrs.GetToken(); v != nil {
			value = *v
		}
		if v := attrs.GetExpiredAt(); v != nil {
			expiredAt = types.StringValue(v.Format(time.RFC3339))
		}
	}
	return id, value, expiredAt, nil
}

// createDescribedTeamToken creates a team authentication token with a description,
// allowing multiple tokens per team.
//
// go-tfe/v2 does not yet expose a working endpoint for this: there is no generated
// builder for POST /teams/{team_id}/authentication-tokens, and the previously generated
// (but never actually wired up in Atlas) top-level POST /authentication-tokens/{id} has
// since been removed from the client entirely. This falls back to the v1 client, which
// still supports it, until go-tfe/v2 catches up.
func createDescribedTeamToken(ctx context.Context, client *tfe.Client, teamID string, description *string, expiry *time.Time) (id string, value string, expiredAt types.String, err error) {
	expiredAt = types.StringNull()

	token, err := client.TeamTokens.CreateWithOptions(ctx, teamID, tfe.TeamTokenCreateOptions{
		Description: description,
		ExpiredAt:   expiry,
	})
	if err != nil {
		return "", "", expiredAt, err
	}

	if !token.ExpiredAt.IsZero() {
		expiredAt = types.StringValue(token.ExpiredAt.Format(time.RFC3339))
	}
	return token.ID, token.Token, expiredAt, nil
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
	api := r.config.ClientV2.API
	var tokenEnvelope models.AuthenticationTokensEnvelopeable
	var err error
	if isTokenID(state.ID.ValueString()) {
		tokenEnvelope, err = api.AuthenticationTokens().ById(state.ID.ValueString()).Get(ctx, nil)
	} else {
		tokenEnvelope, err = api.Teams().ById(teamID).AuthenticationToken().Get(ctx, nil)
	}
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

	// if expired_at was set to null at creation, the API returns a default value of 24 months from the creation date.
	expiredAt := types.StringNull()
	if tokenEnvelope != nil && tokenEnvelope.GetData() != nil {
		if attrs := tokenEnvelope.GetData().GetAttributes(); attrs != nil {
			if v := attrs.GetExpiredAt(); v != nil {
				expiredAt = types.StringValue(v.Format(time.RFC3339))
			}
		}
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
	api := r.config.ClientV2.API
	var err error
	if isTokenID(state.ID.ValueString()) {
		err = api.AuthenticationTokens().ById(state.ID.ValueString()).Delete(ctx, nil)
	} else {
		err = api.Teams().ById(teamID).AuthenticationToken().Delete(ctx, nil)
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

	// Fetch token by ID to set attributes
	tokenEnvelope, err := r.config.ClientV2.API.AuthenticationTokens().ById(req.ID).Get(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error importing team token", err.Error())
		return
	}
	if tokenEnvelope == nil || tokenEnvelope.GetData() == nil {
		resp.Diagnostics.AddError("Error importing team token", "API response did not include the token")
		return
	}
	token := tokenEnvelope.GetData()

	teamID := tokenTeamID(token)
	if teamID == "" {
		resp.Diagnostics.AddError("Error importing team token", "token did not return associated team")
		return
	}

	tokenValue := types.StringNull()
	expiredAt := types.StringNull()
	description := types.StringNull()
	if attrs := token.GetAttributes(); attrs != nil {
		if v := attrs.GetToken(); v != nil {
			tokenValue = types.StringValue(*v)
		}
		if v := attrs.GetExpiredAt(); v != nil {
			expiredAt = types.StringValue(v.Format(time.RFC3339))
		}
		if v := attrs.GetDescription(); v != nil {
			description = types.StringValue(*v)
		}
	}

	result := modelFromTFEToken(types.StringValue(teamID), types.StringValue(req.ID), tokenValue, basetypes.NewBoolNull(), expiredAt, description)
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

// tokenTeamID extracts the ID of the token's team from its team relationship. The API
// returns this as a resource identifier (relationships.team.data.id); older Terraform
// Enterprise releases may instead (or additionally) expose it only as a related link
// (e.g. /api/v2/teams/team-abc123), so that's used as a defensive fallback.
func tokenTeamID(token models.AuthenticationTokensable) string {
	relationships := token.GetRelationships()
	if relationships == nil || relationships.GetTeam() == nil {
		return ""
	}
	team := relationships.GetTeam()

	if data := team.GetData(); data != nil && data.GetId() != nil {
		return *data.GetId()
	}

	if team.GetLinks() == nil {
		return ""
	}
	related := team.GetLinks().GetRelated()
	if related == nil {
		return ""
	}
	parts := strings.Split(strings.TrimSuffix(*related, "/"), "/")
	return parts[len(parts)-1]
}

// Determines whether the ID of the resource is the ID of the authentication token
// or the ID of the team the token belongs to.
func isTokenID(id string) bool {
	return strings.HasPrefix(id, "at-")
}
