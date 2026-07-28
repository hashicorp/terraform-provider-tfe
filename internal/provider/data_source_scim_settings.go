// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &dataSourceTFESCIMSettings{}
	_ datasource.DataSourceWithConfigure = &dataSourceTFESCIMSettings{}
)

// dataSourceTFESCIMSettings is the data source implementation.
type dataSourceTFESCIMSettings struct {
	client *tfe.Client
}

// NewSCIMSettingsDataSource is a helper function to simplify the provider implementation.
func NewSCIMSettingsDataSource() datasource.DataSource {
	return &dataSourceTFESCIMSettings{}
}

// modelDataTFESCIMSettings maps the data source schema data.
type modelDataTFESCIMSettings struct {
	ID                        types.String `tfsdk:"id"`
	Enabled                   types.Bool   `tfsdk:"enabled"`
	Paused                    types.Bool   `tfsdk:"paused"`
	SiteAdminGroupSCIMID      types.String `tfsdk:"site_admin_group_scim_id"`
	SiteAdminGroupDisplayName types.String `tfsdk:"site_admin_group_display_name"`
}

// Metadata returns the data source type name.
func (d *dataSourceTFESCIMSettings) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_scim_settings"
}

// Schema defines the schema for the data source.
func (d *dataSourceTFESCIMSettings) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads the current SCIM provisioning settings for the Terraform Enterprise instance.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the SCIM settings. Always `scim`.",
			},
			"enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether SCIM provisioning is enabled for the Terraform Enterprise instance.",
			},
			"paused": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether SCIM provisioning is paused for the Terraform Enterprise instance.",
			},
			"site_admin_group_scim_id": schema.StringAttribute{
				Computed:    true,
				Description: "The SCIM ID of the group whose members are granted site admin privileges. Empty when no group is linked.",
			},
			"site_admin_group_display_name": schema.StringAttribute{
				Computed:    true,
				Description: "The display name of the group whose members are granted site admin privileges.",
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *dataSourceTFESCIMSettings) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
	d.client = client.Client
}

// Read refreshes the Terraform state with the latest data.
func (d *dataSourceTFESCIMSettings) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	s, err := d.client.Admin.Settings.SCIM.Read(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read SCIM settings", err.Error())
		return
	}

	// Set state
	diags := resp.State.Set(ctx, &modelDataTFESCIMSettings{
		ID:                        types.StringValue(s.ID),
		Enabled:                   types.BoolValue(s.Enabled),
		Paused:                    types.BoolValue(s.Paused),
		SiteAdminGroupSCIMID:      types.StringValue(s.SiteAdminGroupSCIMID),
		SiteAdminGroupDisplayName: types.StringValue(s.SiteAdminGroupDisplayName),
	})
	resp.Diagnostics.Append(diags...)
}
