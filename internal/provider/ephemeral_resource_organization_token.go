// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
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
	ForceRegenerate types.Bool   `tfsdk:"force_regenerate"`
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
			"token": schema.StringAttribute{
				Description: `The generated token.`,
				Computed:    true,
			},
			"force_regenerate": schema.BoolAttribute{
				Description: "If set to `false`, an existing organization token will cause the run to fail. If set to `true`, the check for an existing organization token will be suppressed. Defaults to `false`.",
				Optional:    true,
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
	// Read Terraform config data
	var data OrganizationTokenEphemeralResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get org name or default
	var orgName string
	resp.Diagnostics.Append(e.config.dataOrDefaultOrganization(ctx, req.Config, &orgName)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check for existing token
	existingToken, err := e.config.Client.OrganizationTokens.Read(ctx, orgName)
	if err != nil && !errors.Is(err, tfe.ErrResourceNotFound) {
		resp.Diagnostics.AddError("Unable to read organization token", err.Error())
		return
	}

	// Fail if a token exists unless `force_regenerate` is set
	if existingToken != nil && !data.ForceRegenerate.ValueBool() {
		resp.Diagnostics.AddError("Organization token already exists", "An organization token already exists. Set `force_regenerate` to `true` to suppress this check.")
		return
	}

	result, err := e.config.Client.OrganizationTokens.Create(ctx, orgName)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create organization token", err.Error())
		return
	}

	diags := resp.Private.SetKey(ctx, "organization", []byte(orgName))
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	data = ephemeralResourceModelFromTFEOrganizationToken(orgName, result)

	// Save to ephemeral result data
	resp.Diagnostics.Append(resp.Result.Set(ctx, &data)...)
}

func (e *OrganizationTokenEphemeralResource) Close(ctx context.Context, req ephemeral.CloseRequest, resp *ephemeral.CloseResponse) {
	privateBytes, diags := req.Private.GetKey(ctx, "organization")
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	if err := e.config.Client.OrganizationTokens.Delete(ctx, string(privateBytes)); err != nil && !errors.Is(err, tfe.ErrResourceNotFound) {
		fmt.Printf("%+v\n", err)
		resp.Diagnostics.AddError("Unable to delete organization token", err.Error())
		return
	}
}

// ephemeralResourceModelFromTFEOrganizationToken builds a OrganizationTokenEphemeralResourceModel struct from a
// tfe.OrganizationToken value.
func ephemeralResourceModelFromTFEOrganizationToken(organization string, v *tfe.OrganizationToken) OrganizationTokenEphemeralResourceModel {
	return OrganizationTokenEphemeralResourceModel{
		Organization: types.StringValue(organization),
		Token:        types.StringValue(v.Token),
	}
}
