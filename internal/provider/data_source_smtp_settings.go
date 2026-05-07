// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &dataSourceTFESMTPSettings{}
	_ datasource.DataSourceWithConfigure = &dataSourceTFESMTPSettings{}
)

// NewSMTPSettingsDataSource is a helper function to simplify the provider implementation.
func NewSMTPSettingsDataSource() datasource.DataSource {
	return &dataSourceTFESMTPSettings{}
}

// dataSourceTFESMTPSettings is the data source implementation.
type dataSourceTFESMTPSettings struct {
	client *tfe.Client
}

// modelDataTFESMTPSettings maps the data source schema data.
type modelDataTFESMTPSettings struct {
	ID       types.String `tfsdk:"id"`
	Enabled  types.Bool   `tfsdk:"enabled"`
	Host     types.String `tfsdk:"host"`
	Port     types.Int64  `tfsdk:"port"`
	Sender   types.String `tfsdk:"sender"`
	Auth     types.String `tfsdk:"auth"`
	Username types.String `tfsdk:"username"`
}

// Metadata returns the data source type name.
func (d *dataSourceTFESMTPSettings) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_smtp_settings"
}

// Schema defines the schema for the data source.
func (d *dataSourceTFESMTPSettings) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads SMTP settings for Terraform Enterprise.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the SMTP settings. Always 'smtp'.",
				Computed:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether SMTP is enabled.",
				Computed:    true,
			},
			"host": schema.StringAttribute{
				Description: "The hostname of the SMTP server.",
				Computed:    true,
			},
			"port": schema.Int64Attribute{
				Description: "The port of the SMTP server.",
				Computed:    true,
			},
			"sender": schema.StringAttribute{
				Description: "The sender email address.",
				Computed:    true,
			},
			"auth": schema.StringAttribute{
				Description: "The authentication type. Valid values are 'none', 'plain', and 'login'.",
				Computed:    true,
			},
			"username": schema.StringAttribute{
				Description: "The username used to authenticate to the SMTP server.",
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *dataSourceTFESMTPSettings) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *dataSourceTFESMTPSettings) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data modelDataTFESMTPSettings

	smtpSettings, err := d.client.Admin.Settings.SMTP.Read(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading SMTP Settings",
			"Could not read SMTP Settings: "+err.Error(),
		)
		return
	}

	// Map response to model
	data.ID = types.StringValue(smtpSettings.ID)
	data.Enabled = types.BoolValue(smtpSettings.Enabled)
	data.Host = types.StringValue(smtpSettings.Host)
	data.Port = types.Int64Value(int64(smtpSettings.Port))
	data.Sender = types.StringValue(smtpSettings.Sender)
	data.Auth = types.StringValue(string(smtpSettings.Auth))
	data.Username = types.StringValue(smtpSettings.Username)

	// Set state
	diags := resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}
