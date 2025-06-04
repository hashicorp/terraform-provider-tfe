// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ ephemeral.EphemeralResource              = &TeamTokenEphemeralResource{}
	_ ephemeral.EphemeralResourceWithConfigure = &TeamTokenEphemeralResource{}
)

func NewTeamTokenEphemeralResource() ephemeral.EphemeralResource {
	return &TeamTokenEphemeralResource{}
}

type TeamTokenEphemeralResource struct {
	config ConfiguredClient
}

type TeamTokenEphemeralResourceModel struct {
	ID        types.String      `tfsdk:"id"`
	TeamID    types.String      `tfsdk:"team_id"`
	Token     types.String      `tfsdk:"token"`
	ExpiredAt timetypes.RFC3339 `tfsdk:"expired_at"`
}

type TeamTokenPrivateData struct {
	ID     string `json:"id"`
	TeamID string `json:"team_id"`
}

func (e *TeamTokenEphemeralResource) Schema(ctx context.Context, req ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This ephemeral resource can be used to retrieve a team token without saving its value in state.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: `System generated ID of the team token.`,
				Computed:    true,
			},
			"team_id": schema.StringAttribute{
				Description: `ID of the team.`,
				Required:    true,
			},
			"token": schema.StringAttribute{
				Description: `The generated token.`,
				Computed:    true,
				Sensitive:   true,
			},
			"expired_at": schema.StringAttribute{
				Description: `The token's expiration date.`,
				Optional:    true,
				Computed:    true,
				CustomType:  timetypes.RFC3339Type{},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (e *TeamTokenEphemeralResource) Configure(_ context.Context, req ephemeral.ConfigureRequest, resp *ephemeral.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(ConfiguredClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Ephemeral Resource Configure Type",
			fmt.Sprintf("Expected tfe.ConfiguredClient, got %T. This is a bug in the tfe provider, so please report it on GitHub.", req.ProviderData),
		)

		return
	}

	e.config = client
}

func (e *TeamTokenEphemeralResource) Metadata(ctx context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team_token"
}

func (e *TeamTokenEphemeralResource) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	var config TeamTokenEphemeralResourceModel

	// Read Terraform config data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	desc := fmt.Sprintf("ephemeral-team-token-%s", uuid.New())

	// Create a new options struct
	// Set a description to make use of the new multiple token API
	options := tfe.TeamTokenCreateOptions{
		Description: &desc,
	}

	if !config.ExpiredAt.IsNull() {
		expiredAt, diags := config.ExpiredAt.ValueRFC3339Time()
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		options.ExpiredAt = &expiredAt
	}

	var teamID = config.TeamID.ValueString()
	result, err := e.config.Client.TeamTokens.CreateWithOptions(ctx, config.TeamID.ValueString(), options)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read resource", err.Error())
		return
	}

	privateData, _ := json.Marshal(result)
	resp.Private.SetKey(ctx, "team_token_data", privateData)

	config = ephemeralResourceModelFromTFETeamToken(teamID, result)

	// Save to ephemeral result data
	resp.Diagnostics.Append(resp.Result.Set(ctx, &config)...)
}

func (e *TeamTokenEphemeralResource) Close(ctx context.Context, req ephemeral.CloseRequest, resp *ephemeral.CloseResponse) {
	privateBytes, diags := req.Private.GetKey(ctx, "team_token_data")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var privateData TeamTokenPrivateData
	json.Unmarshal(privateBytes, &privateData)

	log.Printf("[DEBUG] Removing team token with ID: %s", privateData.ID)

	err := e.config.Client.TeamTokens.Delete(ctx, privateData.ID)
	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
			log.Printf("[DEBUG] Team token with ID %s not found, skipping deletion", privateData.ID)
			return
		}

		resp.Diagnostics.AddError("Unable to delete team token", err.Error())
		return
	}
}

// ephemeralResourceModelFromTFETeamToken builds a TeamTokenEphemeralResourceModel struct from a
// tfe.TeamToken value.
func ephemeralResourceModelFromTFETeamToken(teamID string, v *tfe.TeamToken) TeamTokenEphemeralResourceModel {
	return TeamTokenEphemeralResourceModel{
		TeamID:    types.StringValue(teamID),
		Token:     types.StringValue(v.Token),
		ExpiredAt: timetypes.NewRFC3339TimeValue(v.ExpiredAt),
	}
}
