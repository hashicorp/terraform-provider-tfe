// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"time"

	tfe "github.com/hashicorp/go-tfe"
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
	Organization types.String `tfsdk:"organization"`
	ExpiredAt    types.String `tfsdk:"expired_at"`
	Token        types.String `tfsdk:"token"`
}

func (e *OrganizationTokenEphemeralResource) Schema(ctx context.Context, req ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This ephemeral resource can be used to retrieve an organization token without saving its value in state. Using this ephemeral resource will generate a new token each time it is used, invalidating any existing organization token.",
		Attributes: map[string]schema.Attribute{
			"organization": schema.StringAttribute{
				Description: `Name of the organization. If omitted, organization must be defined in the provider config.`,
				Optional:    true,
				Computed:    true,
			},
			"expired_at": schema.StringAttribute{
				Description: `The token's expiration date. The expiration date must be a date/time string in RFC3339 format (e.g., "2024-12-31T23:59:59Z"). If no expiration date is supplied, the expiration date will default to null and never expire.`,
				Optional:    true,
			},
			"token": schema.StringAttribute{
				Description: `The generated token.`,
				Computed:    true,
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
	// Read Terraform config config
	var config OrganizationTokenEphemeralResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get org name or default
	var orgName string
	resp.Diagnostics.Append(e.config.dataOrDefaultOrganization(ctx, req.Config, &orgName)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create options
	var expiredAt *time.Time
	if !config.ExpiredAt.IsNull() {
		parsed, err := time.Parse(time.RFC3339, config.ExpiredAt.String())
		if err != nil {
			resp.Diagnostics.AddError("Invalid expired_at value", err.Error())
			return
		}

		expiredAt = &parsed
	}

	opts := tfe.OrganizationTokenCreateOptions{
		ExpiredAt: expiredAt,
	}

	// Create a new token
	result, err := e.config.Client.OrganizationTokens.CreateWithOptions(ctx, orgName, opts)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create organization token", err.Error())
		return
	}

	// Set the token in the model
	config.Token = types.StringValue(result.Token)

	// Write the data back to the ephemeral resource
	resp.Diagnostics.Append(resp.Result.Set(ctx, &config)...)
}
