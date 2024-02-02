// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource                   = &dataSourceTFENoCodeModule{}
	_ datasource.DataSourceWithConfigure      = &dataSourceTFENoCodeModule{}
	_ datasource.DataSourceWithValidateConfig = &dataSourceTFENoCodeModule{}
)

// NewNoCodeModuleDataSource is a helper function to simplify the implementation.
func NewNoCodeModuleDataSource() datasource.DataSource {
	return &dataSourceTFENoCodeModule{}
}

// dataSourceTFENoCodeModule is the data source implementation.
type dataSourceTFENoCodeModule struct {
	config ConfiguredClient
}

// modelNoCodeModule maps the data source schema data.
type modelNoCodeModule struct {
	ID               types.String `tfsdk:"id"`
	Organization     types.String `tfsdk:"organization"`
	Namespace        types.String `tfsdk:"namespace"`
	VersionPin       types.String `tfsdk:"version_pin"`
	RegistryModuleID types.String `tfsdk:"registry_module_id"`
	Enabled          types.Bool   `tfsdk:"enabled"`
}

// Metadata returns the data source type name.
func (d *dataSourceTFENoCodeModule) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_no_code_module"
}

// Schema defines the schema for the data source.
func (d *dataSourceTFENoCodeModule) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This data source can be used to retrieve a public or private no-code module.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID of the no-code module.",
				Required:    true,
			},
			"organization": schema.StringAttribute{
				Description: "Name of the organization.",
				Optional:    true,
				Computed:    true,
			},
			"namespace": schema.StringAttribute{
				Description: "The namespace of the no-code module.",
				Computed:    true,
			},
			"registry_module_id": schema.StringAttribute{
				Description: "ID of the registry module.",
				Computed:    true,
			},
			"version_pin": schema.StringAttribute{
				Description: "Version pin of the no-code module.",
				Computed:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Indiate if this no-code module is currently enabled.",
				Computed:    true,
			},
		},
	}
}

func (d *dataSourceTFENoCodeModule) ValidateConfig(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	var config modelNoCodeModule

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured client to the data source.
func (d *dataSourceTFENoCodeModule) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *dataSourceTFENoCodeModule) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data modelNoCodeModule

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	options := &tfe.RegistryNoCodeModuleReadOptions{
		Include: []tfe.RegistryNoCodeModuleIncludeOpt{tfe.RegistryNoCodeIncludeVariableOptions},
	}

	tflog.Debug(ctx, "Reading no code module")
	module, err := d.config.Client.RegistryNoCodeModules.Read(ctx, data.ID.ValueString(), options)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read no code module", err.Error())
		return
	}

	data = modelFromTFENoCodeModule(module)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
func modelFromTFENoCodeModule(v *tfe.RegistryNoCodeModule) modelNoCodeModule {
	return modelNoCodeModule{
		ID:               types.StringValue(v.ID),
		Organization:     types.StringValue(v.Organization.Name),
		RegistryModuleID: types.StringValue(v.RegistryModule.ID),
		Namespace:        types.StringValue(v.RegistryModule.Namespace),
		VersionPin:       types.StringValue(v.VersionPin),
		Enabled:          types.BoolValue(v.Enabled),
	}
}
