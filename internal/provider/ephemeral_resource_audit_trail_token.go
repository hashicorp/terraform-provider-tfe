// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ ephemeral.EphemeralResource              = &AuditTrailTokenEphemeralResource{}
	_ ephemeral.EphemeralResourceWithConfigure = &AuditTrailTokenEphemeralResource{}
)

func NewAuditTrailTokenEphemeralResource() ephemeral.EphemeralResource {
	return &AuditTrailTokenEphemeralResource{}
}

type AuditTrailTokenEphemeralResource struct {
	config ConfiguredClient
}

type AuditTrailTokenEphemeralResourceModel struct {
	Organization types.String      `tfsdk:"organization"`
	Token        types.String      `tfsdk:"token"`
	ExpiredAt    timetypes.RFC3339 `tfsdk:"expired_at"`
}

func (e *AuditTrailTokenEphemeralResource) Schema(ctx context.Context, req ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This ephemeral resource can be used to retrieve an audit trail token without saving its value in state. Using this ephemeral resource will generate a new token each time it is used, invalidating any existing audit trail token.",
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
				Description: `The token's expiration date. The expiration date must be a date/time string in RFC3339 format (e.g., "2024-12-31T23:59:59Z"). If no expiration date is supplied, the expiration date will default to null and never expire.`,
				Optional:    true,
				CustomType:  timetypes.RFC3339Type{},
			},
		},
		DeprecationMessage: "Use of this ephemeral resource is deprecated. Please use the `tfe_audit_trail_token` managed resource instead.",
	}
}

// Configure adds the provider configured client to the data source.
func (e *AuditTrailTokenEphemeralResource) Configure(_ context.Context, req ephemeral.ConfigureRequest, resp *ephemeral.ConfigureResponse) {
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

func (e *AuditTrailTokenEphemeralResource) Metadata(ctx context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_audit_trail_token"
}

func (e *AuditTrailTokenEphemeralResource) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	// No-op
}
