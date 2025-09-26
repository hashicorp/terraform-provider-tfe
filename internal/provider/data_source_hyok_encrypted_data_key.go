// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"time"
)

var (
	_ datasource.DataSource              = &dataSourceHYOKEncryptedDataKey{}
	_ datasource.DataSourceWithConfigure = &dataSourceHYOKEncryptedDataKey{}
)

func NewHYOKEncryptedDataKeyDataSource() datasource.DataSource {
	return &dataSourceHYOKEncryptedDataKey{}
}

type dataSourceHYOKEncryptedDataKey struct {
	config ConfiguredClient
}

type HYOKEncryptedDataKeyDataSourceModel struct {
	ID              types.String `tfsdk:"id"`
	EncryptedDEK    types.String `tfsdk:"encrypted_dek"`
	CustomerKeyName types.String `tfsdk:"customer_key_name"`
	CreatedAt       types.String `tfsdk:"created_at"`
}

func (d *dataSourceHYOKEncryptedDataKey) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *dataSourceHYOKEncryptedDataKey) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_hyok_encrypted_data_key"
}

func (d *dataSourceHYOKEncryptedDataKey) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This data source can be used to retrieve a HYOK customer key version.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the HYOK encrypted data key.",
				Required:    true,
			},
			"encrypted_dek": schema.StringAttribute{
				Description: "The encrypted data encryption key of the HYOK encrypted data key.",
				Computed:    true,
			},
			"customer_key_name": schema.StringAttribute{
				Description: "The customer provided name of the HYOK encrypted data key.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "The timestamp when the key version was created.",
				Computed:    true,
			},
		},
	}
}

func (d *dataSourceHYOKEncryptedDataKey) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data HYOKEncryptedDataKeyDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Make API call to fetch the HYOK customer key version
	keyVersion, err := d.config.Client.HYOKEncryptedDataKeys.Read(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read HYOK customer key version", err.Error())
		return
	}

	// Set the computed attributes from the API response
	data.EncryptedDEK = types.StringValue(keyVersion.EncryptedDEK)
	data.CustomerKeyName = types.StringValue(keyVersion.CustomerKeyName)
	data.CreatedAt = types.StringValue(keyVersion.CreatedAt.Format(time.RFC3339))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
