// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type dataSourceTFEIPAllowlist struct {
	config ConfiguredClient
}

var (
	_ datasource.DataSource              = &dataSourceTFEIPAllowlist{}
	_ datasource.DataSourceWithConfigure = &dataSourceTFEIPAllowlist{}
)

// NewIPAllowlistDataSource returns a new IP allowlist (CIDR range list) data source.
func NewIPAllowlistDataSource() datasource.DataSource {
	return &dataSourceTFEIPAllowlist{}
}

type modelDataSourceTFEIPAllowlist struct {
	ID               types.String `tfsdk:"id"`
	Organization     types.String `tfsdk:"organization"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	EnforcementScope types.String `tfsdk:"enforcement_scope"`
	AgentPoolIDs     types.Set    `tfsdk:"agent_pool_ids"`
	CIDRRanges       types.Set    `tfsdk:"cidr_range"`
}

func (d *dataSourceTFEIPAllowlist) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ip_allowlist"
}

func (d *dataSourceTFEIPAllowlist) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This data source can be used to retrieve an IP allowlist (CIDR range list) in an organization by its ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the IP allowlist.",
				Computed:    true,
			},
			"organization": schema.StringAttribute{
				Description: "Name of the organization. If omitted, organization must be defined in the provider config.",
				Optional:    true,
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the IP allowlist.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "A description for the IP allowlist.",
				Computed:    true,
			},
			"enforcement_scope": schema.StringAttribute{
				Description: "Where the IP allowlist is enforced. One of `organization`, `all_agent_pools`, or `selected_agent_pools`.",
				Computed:    true,
			},
			"agent_pool_ids": schema.SetAttribute{
				Description: "The IDs of the agent pools the IP allowlist applies to.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"cidr_ranges": schema.SetNestedAttribute{
				Description: "The CIDR ranges that belong to the IP allowlist.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"range": schema.StringAttribute{
							Description: "The IPv4 CIDR range.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "A description for the CIDR range.",
							Computed:    true,
						},
						"enabled": schema.BoolAttribute{
							Description: "Whether the CIDR range is enforced.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *dataSourceTFEIPAllowlist) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(ConfiguredClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected tfe.ConfiguredClient, got %T. This is a bug in the tfe provider, so please report it on GitHub.", req.ProviderData),
		)
		return
	}
	d.config = client
}

func (d *dataSourceTFEIPAllowlist) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config modelDataSourceTFEIPAllowlist
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var organization string
	resp.Diagnostics.Append(d.config.dataOrDefaultOrganization(ctx, req.Config, &organization)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := config.Name.ValueString()
	tflog.Debug(ctx, fmt.Sprintf("Read IP allowlist %q in organization %q", name, organization))

	listID, found, err := findIPAllowlistByName(ctx, d.config.ClientV2, organization, name)
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving IP allowlists", err.Error())
		return
	}
	if !found {
		resp.Diagnostics.AddError(
			"Could not find IP allowlist",
			fmt.Sprintf("IP allowlist %s/%s not found", organization, name),
		)
		return
	}

	// Reuse the resource read logic to fully populate the model.
	r := &resourceTFEIPAllowlist{config: d.config}
	result, diags, err := r.fetchIPAllowlist(ctx, listID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading IP allowlist", err.Error())
		return
	}
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// When the API does not return the organization on the relationship, fall
	// back to the resolved organization name.
	if result.Organization.IsNull() || result.Organization.ValueString() == "" {
		result.Organization = types.StringValue(organization)
	}

	model := modelDataSourceTFEIPAllowlist{
		ID:               result.ID,
		Organization:     result.Organization,
		Name:             result.Name,
		Description:      result.Description,
		EnforcementScope: result.EnforcementScope,
		AgentPoolIDs:     result.AgentPoolIDs,
		CIDRRanges:       result.CIDRRanges,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
