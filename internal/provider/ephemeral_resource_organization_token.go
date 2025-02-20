// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ ephemeral.EphemeralResource              = &OrganizationTokenEphemeralResource{}
	_ ephemeral.EphemeralResourceWithConfigure = &OrganizationTokenEphemeralResource{}
)

func NewOrganizationTokenEphemeralResource() ephemeral.EphemeralResource {
	return &OrganizationTokenEphemeralResource{}
}

type OrganizationTokenEphemeralResource struct {
	config ConfiguredClient
}

type OrganizationTokenEphemeralResourceModel struct {
	Organization    types.String `tfsdk:"organization"`
	Token           types.String `tfsdk:"token"`
	ForceRegenerate types.Bool   `tfsdk:"force_generate"`
	ExpiredAt       types.String `tfsdk:"expired_at"`
}

func (e *OrganizationTokenEphemeralResource) Schema(ctx context.Context, req ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"organization": schema.StringAttribute{
				Description: `Name of the organization. If omitted, organization must be defined in the provider config.`,
			},
			"token": schema.StringAttribute{
				Description: `The generated token.`,
				Computed:    true,
			},
			"force_generate": schema.BoolAttribute{
				Description: `If set to true, a new token will be generated even if a token already exists. This will invalidate the existing token!`,
			},
			"expired_at": schema.StringAttribute{
				Description: `The token's expiration date. The expiration date must be a date/time string in RFC3339 format (e.g., "2024-12-31T23:59:59Z"). If no expiration date is supplied, the expiration date will default to null and never expire.`,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (e *OrganizationTokenEphemeralResource) Configure(_ context.Context, req ephemeral.ConfigureRequest, resp *ephemeral.ConfigureResponse) {
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

func (e *OrganizationTokenEphemeralResource) Metadata(ctx context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_token"
}

func (e *OrganizationTokenEphemeralResource) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	var data OrganizationTokenEphemeralResourceModel

	// Read Terraform config data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := e.config.Client.OrganizationTokens.Read(ctx, data.Organization.String())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read resource", err.Error())
		return
	}

	data = ephemeralResourceModelFromTFEOrganizationToken(result)

	// Save to ephemeral result data
	resp.Diagnostics.Append(resp.Result.Set(ctx, &data)...)
}

// ephemeralResourceModelFromTFEOrganizationToken builds a OrganizationTokenEphemeralResourceModel struct from a
// tfe.OrganizationToken value.
func ephemeralResourceModelFromTFEOrganizationToken(v *tfe.OrganizationToken) OrganizationTokenEphemeralResourceModel {
	return OrganizationTokenEphemeralResourceModel{
		Token:     types.StringValue(v.Token),
		ExpiredAt: types.StringValue(v.ExpiredAt.String()),
	}
}
