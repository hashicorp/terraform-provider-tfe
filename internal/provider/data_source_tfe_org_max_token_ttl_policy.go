// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = &dataSourceTFEOrgMaxTokenTTLPolicy{}
var _ datasource.DataSourceWithConfigure = &dataSourceTFEOrgMaxTokenTTLPolicy{}

func NewOrgMaxTokenTTLPolicyDataSource() datasource.DataSource {
	return &dataSourceTFEOrgMaxTokenTTLPolicy{}
}

type dataSourceTFEOrgMaxTokenTTLPolicy struct {
	config ConfiguredClient
}

type modelTFEOrgMaxTokenTTLPolicyData struct {
	Organization            types.String `tfsdk:"organization"`
	OrgTokenMaxTTLMs        types.Int64  `tfsdk:"org_token_max_ttl_ms"`
	TeamTokenMaxTTLMs       types.Int64  `tfsdk:"team_token_max_ttl_ms"`
	AuditTrailTokenMaxTTLMs types.Int64  `tfsdk:"audit_trail_token_max_ttl_ms"`
	UserTokenMaxTTLMs       types.Int64  `tfsdk:"user_token_max_ttl_ms"`
}

func (d *dataSourceTFEOrgMaxTokenTTLPolicy) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_org_max_token_ttl_policy"
}

func (d *dataSourceTFEOrgMaxTokenTTLPolicy) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves the maximum time-to-live (TTL) policy for API tokens in an organization. " +
			"This data source fetches the current TTL limits for organization, team, audit trail, " +
			"and user tokens from the database.",

		Attributes: map[string]schema.Attribute{
			"organization": schema.StringAttribute{
				Description: "Name of the organization. If omitted, organization must be defined in the provider config.",
				Optional:    true,
				Computed:    true,
			},
			"org_token_max_ttl_ms": schema.Int64Attribute{
				Description: "Maximum lifespan allowed for organization tokens in milliseconds.",
				Computed:    true,
			},
			"team_token_max_ttl_ms": schema.Int64Attribute{
				Description: "Maximum lifespan allowed for team tokens in milliseconds.",
				Computed:    true,
			},
			"audit_trail_token_max_ttl_ms": schema.Int64Attribute{
				Description: "Maximum lifespan allowed for audit trail tokens in milliseconds.",
				Computed:    true,
			},
			"user_token_max_ttl_ms": schema.Int64Attribute{
				Description: "Maximum lifespan allowed for user tokens in milliseconds.",
				Computed:    true,
			},
		},
	}
}

func (d *dataSourceTFEOrgMaxTokenTTLPolicy) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(ConfiguredClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected data source Configure type",
			fmt.Sprintf("Expected tfe.ConfiguredClient, got %T", req.ProviderData),
		)
		return
	}
	d.config = client
}

func (d *dataSourceTFEOrgMaxTokenTTLPolicy) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config modelTFEOrgMaxTokenTTLPolicyData

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var organization string
	resp.Diagnostics.Append(d.config.dataOrDefaultOrganization(ctx, req.Config, &organization)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Reading token TTL policies for organization: %s", organization))

	policyList, err := d.config.Client.OrganizationTokenTTLPolicies.List(ctx, organization, nil)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read organization token TTL policies", err.Error())
		return
	}

	// Convert the API response to the data source model
	result := modelFromTokenTTLPoliciesData(organization, policyList.Items)

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func modelFromTokenTTLPoliciesData(organization string, policies []*tfe.OrganizationTokenTTLPolicy) modelTFEOrgMaxTokenTTLPolicyData {
	// Default TTL: 2 years in milliseconds
	defaultTTLMs := int64(63072000000)

	result := modelTFEOrgMaxTokenTTLPolicyData{
		Organization:            types.StringValue(organization),
		OrgTokenMaxTTLMs:        types.Int64Value(defaultTTLMs),
		TeamTokenMaxTTLMs:       types.Int64Value(defaultTTLMs),
		AuditTrailTokenMaxTTLMs: types.Int64Value(defaultTTLMs),
		UserTokenMaxTTLMs:       types.Int64Value(defaultTTLMs),
	}

	// Set actual values from policies
	for _, policy := range policies {
		switch policy.TokenType {
		case tfe.TokenTypeOrganization:
			result.OrgTokenMaxTTLMs = types.Int64Value(policy.MaxTTLMs)
		case tfe.TokenTypeTeam:
			result.TeamTokenMaxTTLMs = types.Int64Value(policy.MaxTTLMs)
		case tfe.TokenTypeAuditTrails:
			result.AuditTrailTokenMaxTTLMs = types.Int64Value(policy.MaxTTLMs)
		case tfe.TokenTypeUser:
			result.UserTokenMaxTTLMs = types.Int64Value(policy.MaxTTLMs)
		}
	}

	return result
}
