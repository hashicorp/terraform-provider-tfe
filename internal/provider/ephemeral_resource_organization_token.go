// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"time"

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
	Organization types.String `tfsdk:"organization"`
	Token        types.String `tfsdk:"token"`
	ExpiredAt    types.String `tfsdk:"expired_at"`
}

func (e *OrganizationTokenEphemeralResource) Schema(ctx context.Context, req ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This ephemeral resource can be used to retrieve an organization token without saving its value in state.",
		Attributes: map[string]schema.Attribute{
			"organization": schema.StringAttribute{
				Description: `Name of the organization. If omitted, organization must be defined in the provider config.`,
				Optional:    true,
				Computed:    true,
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

	var orgName string
	resp.Diagnostics.Append(e.config.dataOrDefaultOrganization(ctx, req.Config, &orgName)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create a new options struct
	options := tfe.OrganizationTokenCreateOptions{}
	if data.ExpiredAt.IsUnknown() {
		expiredAt := data.ExpiredAt.ValueString()
		expiredAtTime, err := time.Parse(time.RFC3339, expiredAt)
		if err != nil {
			resp.Diagnostics.AddError("Invalid ExpiredAt value", "The ExpiredAt value must be set to a valid date.")
			return
		}
		options.ExpiredAt = &expiredAtTime
	}

	result, err := e.config.Client.OrganizationTokens.CreateWithOptions(ctx, orgName, options)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read resource", err.Error())
		return
	}

	data = ephemeralResourceModelFromTFEOrganizationToken(orgName, result)

	// Save to ephemeral result data
	resp.Diagnostics.Append(resp.Result.Set(ctx, &data)...)
}

// ephemeralResourceModelFromTFEOrganizationToken builds a OrganizationTokenEphemeralResourceModel struct from a
// tfe.OrganizationToken value.
func ephemeralResourceModelFromTFEOrganizationToken(organization string, v *tfe.OrganizationToken) OrganizationTokenEphemeralResourceModel {
	return OrganizationTokenEphemeralResourceModel{
		Organization: types.StringValue(organization),
		Token:        types.StringValue(v.Token),
		ExpiredAt:    types.StringValue(v.ExpiredAt.String()),
	}
}
