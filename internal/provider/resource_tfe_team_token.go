// // Copyright (c) HashiCorp, Inc.
// // SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// import (
// 	"context"
// 	"errors"
// 	"fmt"
// 	"log"
// 	"time"

//	tfe "github.com/hashicorp/go-tfe"
//	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
//
// )
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

// Metadata implements resource.Resource.
func (r *resourceTFETeamToken) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team_token"
}

// Schema implements resource.Resource.
func (r *resourceTFETeamToken) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "", //TODO ADD DESCRIPTIONS
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"team_id": schema.StringAttribute{
				Description: "", //TODO ADD DESCRIPTIONS
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"force_regenerate": schema.BoolAttribute{
				Description: "",
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"token": schema.StringAttribute{
				Description: "",
				Computed:    true,
				Sensitive:   true,
			},
			"expired_at": schema.StringAttribute{
				Description: "",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Description: "",
	}
}

// Create implements resource.Resource.
func (r *resourceTFETeamToken) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan modelTFETeamToken
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	teamID := plan.TeamID.ValueString()
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
	//TODO: use timetypes
	expiredAt := plan.ExpiredAt.ValueString()
	options := tfe.TeamTokenCreateOptions{}
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

	result := modelFromTFEToken(token, plan.TeamID, plan.ForceRegenerate, plan.ExpiredAt)
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

func modelFromTFEToken(token *tfe.TeamToken, teamID types.String, forceRegenerate types.Bool, expiredAt types.String) modelTFETeamToken {
	m := modelTFETeamToken{
		ID:              teamID,
		TeamID:          teamID,
		ForceRegenerate: forceRegenerate,
		ExpiredAt:       types.StringNull(),
		Token:           types.StringValue(token.Token),
	}
	if !expiredAt.IsNull() {
		m.ExpiredAt = expiredAt
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
	token, err := r.config.Client.TeamTokens.Read(ctx, teamID)
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
	result := modelFromTFEToken(token, state.TeamID, state.ForceRegenerate, state.ExpiredAt)
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
	if err := r.config.Client.TeamTokens.Delete(ctx, teamID); err != nil {
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
