// Copyright (c) HashiCorp, Inc.
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
	_ datasource.DataSource              = &dataSourceOrganizationAuditConfiguration{}
	_ datasource.DataSourceWithConfigure = &dataSourceOrganizationAuditConfiguration{}
)

type modelDataTFEOrganizationAuditConfigurationV0 struct {
	ID           types.String `tfsdk:"id"`
	Organization types.String `tfsdk:"organization"`

	AuditTrailsEnabled          types.Bool   `tfsdk:"audit_trails_enabled"`
	HCPAuditLogStreamingEnabled types.Bool   `tfsdk:"hcp_log_streaming_enabled"`
	HCPOrganization             types.String `tfsdk:"hcp_organization"`
}

func dataTFEOrganizationAuditConfiguration(v *tfe.OrganizationAuditConfiguration) modelDataTFEOrganizationAuditConfigurationV0 {
	result := modelDataTFEOrganizationAuditConfigurationV0{
		ID:           types.StringValue(v.ID),
		Organization: types.StringValue(v.Organization.Name),

		AuditTrailsEnabled:          types.BoolValue(v.AuditTrails.Enabled),
		HCPAuditLogStreamingEnabled: types.BoolValue(v.HCPAuditLogStreaming.Enabled),
		HCPOrganization:             types.StringValue(v.HCPAuditLogStreaming.OrganizationID),
	}

	return result
}

// NewOrganizationAuditConfigurationDataSource is a helper function to simplify the provider implementation.
func NewOrganizationAuditConfigurationDataSource() datasource.DataSource {
	return &dataSourceOrganizationAuditConfiguration{}
}

// dataSourceOrganizationAuditConfiguration is the data source implementation.
type dataSourceOrganizationAuditConfiguration struct {
	config ConfiguredClient
}

// Metadata returns the data source type name.
func (d *dataSourceOrganizationAuditConfiguration) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_audit_configuration"
}

func (d *dataSourceOrganizationAuditConfiguration) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Service-generated identifier for the configuration.",
			},
			"organization": schema.StringAttribute{
				Required:    true,
				Description: "Name of the organization.",
			},

			"audit_trails_enabled": schema.BoolAttribute{
				Optional:    true,
				Description: "Whether Audit Trails is enabled for the organization.",
			},
			"hcp_log_streaming_enabled": schema.BoolAttribute{
				Optional:    true,
				Description: "Whether HCP Audit Log Streaming is enabled for the organization.",
			},
			"hcp_organization": schema.StringAttribute{
				Optional:    true,
				Description: "The destination HCP Organization for HCP Audit Log Streaming.",
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *dataSourceOrganizationAuditConfiguration) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read refreshes the Terraform state with the latest data.
func (d *dataSourceOrganizationAuditConfiguration) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data modelDataTFEOrganizationAuditConfigurationV0

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var organization string
	resp.Diagnostics.Append(d.config.dataOrDefaultOrganization(ctx, req.Config, &organization)...)
	if resp.Diagnostics.HasError() {
		return
	}

	org, err := d.config.Client.Organizations.Read(ctx, organization)
	if err != nil {
		resp.Diagnostics.AddError("Error reading Organization Audit Configuration",
			fmt.Sprintf("Could not read Organization %q, unexpected error: %s", organization, err.Error()),
		)
		return
	}

	if !org.Permissions.CanManageAuditing {
		resp.Diagnostics.AddWarning("Cannot read audit configuration",
			fmt.Sprintf("Cannot not read the audit configuration in organization %q due to insufficient permissions or the organization not supporting auditing", organization),
		)
		return
	}

	ac, err := d.config.Client.OrganizationAuditConfigurations.Read(ctx, organization)
	if err != nil {
		resp.Diagnostics.AddError("Error reading Organization Audit Configuration",
			fmt.Sprintf("Could not read Audit Configuration in organization %q, unexpected error: %s", organization, err.Error()),
		)
		return
	}

	result := dataTFEOrganizationAuditConfiguration(ac)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}
