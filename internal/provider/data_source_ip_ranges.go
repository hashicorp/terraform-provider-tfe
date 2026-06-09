// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &dataSourceTFEIPRanges{}
	_ datasource.DataSourceWithConfigure = &dataSourceTFEIPRanges{}
)

func NewIPRangesDataSource() datasource.DataSource {
	return &dataSourceTFEIPRanges{}
}

type dataSourceTFEIPRanges struct {
	config ConfiguredClient
}

// model TFEIPRages maps the data source schema data to a struct.
type modelTFEIPRanges struct {
	ID            types.String `tfsdk:"id"`
	API           types.List   `tfsdk:"api"`
	Notifications types.List   `tfsdk:"notifications"`
	Sentinel      types.List   `tfsdk:"sentinel"`
	VCS           types.List   `tfsdk:"vcs"`
}

// modelFromTFEIPRanges builds a modelTFEIPRanges struct from a tfe.IPRanges value.
func modelFromTFEIPRanges(i *tfe.IPRange) (modelTFEIPRanges, diag.Diagnostics) {
	model := modelTFEIPRanges{
		ID:            types.StringValue("ip_ranges"),
		API:           types.ListNull(types.StringType),
		Notifications: types.ListNull(types.StringType),
		Sentinel:      types.ListNull(types.StringType),
		VCS:           types.ListNull(types.StringType),
	}

	var diags diag.Diagnostics

	// Since calls are low-stakes computationally, these are not short circuited so all errors are visible
	api, err := types.ListValueFrom(ctx, types.StringType, i.API)
	diags.Append(err...)
	model.API = api

	n, err := types.ListValueFrom(ctx, types.StringType, i.Notifications)
	diags.Append(err...)
	model.Notifications = n

	s, err := types.ListValueFrom(ctx, types.StringType, i.Sentinel)
	diags.Append(err...)
	model.Sentinel = s

	v, err := types.ListValueFrom(ctx, types.StringType, i.VCS)
	diags.Append(err...)
	model.VCS = v

	return model, diags
}

func (d *dataSourceTFEIPRanges) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ip_ranges"
}

func (d *dataSourceTFEIPRanges) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This data source can be used to retrieve a list of HCP Terraform's IP ranges.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Static identifier for HCP IP ranges.",
				Computed:    true,
			},
			"api": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "The list of IP ranges in CIDR notation used for connections from user site to HCP Terraform APIs.",
				Computed:    true,
			},
			"notifications": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "The list of IP ranges in CIDR notation used for notifications.",
				Computed:    true,
			},
			"sentinel": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "The list of IP ranges in CIDR notation used for outbound requests from Sentinel policies. Applicable for Policy Checks mode only.",
				Computed:    true,
			},
			"vcs": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "The list of IP ranges in CIDR notation used for connecting to VCS providers.",
				Computed:    true,
			},
		},
	}
}

func (d *dataSourceTFEIPRanges) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data modelTFEIPRanges

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ipRanges, err := d.config.Client.Meta.IPRanges.Read(ctx, "")
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving IP Ranges", err.Error())
		return
	}

	result, diags := modelFromTFEIPRanges(ipRanges)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

// When we need configuration data from or about the provider e.g. via login, we need Configure
func (d *dataSourceTFEIPRanges) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Early exit if provider is unconfigured (i.e. we're only validating config or something)
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(ConfiguredClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure type",
			fmt.Sprintf("Expected tfe.ConfiguredClient, got %T. This is a bug in the tfe provider, so please report it on GitHub.", req.ProviderData),
		)
	}
	d.config = client
}
