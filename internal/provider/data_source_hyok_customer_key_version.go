// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-tfe/v2/api/models"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"time"
)

var (
	_ datasource.DataSource              = &dataSourceHYOKCustomerKeyVersion{}
	_ datasource.DataSourceWithConfigure = &dataSourceHYOKCustomerKeyVersion{}
)

func NewHYOKCustomerKeyVersionDataSource() datasource.DataSource {
	return &dataSourceHYOKCustomerKeyVersion{}
}

type dataSourceHYOKCustomerKeyVersion struct {
	config ConfiguredClient
}

type HYOKCustomerKeyVersionDataSourceModel struct {
	ID                types.String `tfsdk:"id"`
	Status            types.String `tfsdk:"status"`
	Error             types.String `tfsdk:"error"`
	KeyVersion        types.String `tfsdk:"key_version"`
	CreatedAt         types.String `tfsdk:"created_at"`
	WorkspacesSecured types.Int64  `tfsdk:"workspaces_secured"`
}

func (d *dataSourceHYOKCustomerKeyVersion) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *dataSourceHYOKCustomerKeyVersion) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_hyok_customer_key_version"
}

func (d *dataSourceHYOKCustomerKeyVersion) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This data source can be used to retrieve a Hold Your Own Keys (HYOK) customer key version.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the HYOK customer key version.",
				Required:    true,
			},
			"status": schema.StringAttribute{
				Description: "The status of the HYOK customer key version.",
				Computed:    true,
			},
			"error": schema.StringAttribute{
				Description: "Any error message associated with the HYOK customer key version.",
				Computed:    true,
			},
			"key_version": schema.StringAttribute{
				Description: "The version number of the customer key version.",
				Computed:    true,
			},
			"workspaces_secured": schema.Int64Attribute{
				Description: "The number of workspaces secured by this customer key version.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "The timestamp when the key version was created.",
				Computed:    true,
			},
		},
	}
}

func (d *dataSourceHYOKCustomerKeyVersion) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data HYOKCustomerKeyVersionDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Make API call to fetch the HYOK customer key version
	id := data.ID.ValueString()
	envelope, err := d.config.ClientV2.API.HyokCustomerKeyVersions().ByHyok_customer_key_version_id(id).Get(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read HYOK customer key version", err.Error())
		return
	}

	// Set the computed attributes from the API response
	result := modelFromTFEHYOKCustomerKeyVersion(data.ID, envelope)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func modelFromTFEHYOKCustomerKeyVersion(id types.String, p models.HyokCustomerKeyVersionsEnvelopeable) HYOKCustomerKeyVersionDataSourceModel {
	model := HYOKCustomerKeyVersionDataSourceModel{ID: id}

	data := p.GetData()
	if data == nil {
		return model
	}
	if data.GetId() != nil {
		model.ID = types.StringValue(*data.GetId())
	}

	attributes := data.GetAttributes()
	if attributes == nil {
		return model
	}

	model.Status = types.StringValue(attributes.GetStatus().String())
	model.WorkspacesSecured = types.Int64Value(int64(*attributes.GetWorkspacesSecured()))
	model.CreatedAt = types.StringValue(attributes.GetCreatedAt().Format(time.RFC3339))

	if attributes.GetKeyVersion() != nil {
		model.KeyVersion = types.StringValue(*attributes.GetKeyVersion())
	}

	if attributes.GetError() != nil {
		model.Error = types.StringValue(*attributes.GetError())
	}

	return model
}
