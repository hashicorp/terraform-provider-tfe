// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"log"
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
	ID         types.String `tfsdk:"id"`
	Status     types.String `tfsdk:"status"`
	Error      types.String `tfsdk:"error"`
	KeyVersion types.String `tfsdk:"key_version"`
	CreatedAt  types.String `tfsdk:"created_at"`
	UpdatedAt  types.String `tfsdk:"updated_at"`
	RevokedAt  types.String `tfsdk:"revoked_at"`
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
		Description: "This data source can be used to retrieve a HYOK customer key version.",
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
				Description: "The version number of the customer key.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "The timestamp when the key version was created.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "The timestamp when the key version was last updated.",
				Computed:    true,
			},
			"revoked_at": schema.StringAttribute{
				Description: "The timestamp when the key version was revoked, if applicable.",
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

	log.Printf("[DEBUG] Reading HYOK customer key version: %s", data.ID.ValueString())

	// Make API call to fetch the HYOK customer key version
	keyVersion, err := d.config.Client.HYOKCustomerKeyVersions.Read(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read HYOK customer key version", err.Error())
		return
	}

	// Set the computed attributes from the API response
	data.Status = types.StringValue(string(keyVersion.Status))
	data.KeyVersion = types.StringValue(keyVersion.KeyVersion)
	data.CreatedAt = types.StringValue(keyVersion.CreatedAt.Format(time.RFC3339)) // TODO DOM: Check this format
	data.UpdatedAt = types.StringValue(keyVersion.UpdatedAt.Format(time.RFC3339)) // TODO DOM: Check this format
	data.RevokedAt = types.StringValue(keyVersion.RevokedAt.Format(time.RFC3339)) // TODO DOM: Check this format
	data.Error = types.StringValue(keyVersion.Error)

	//if keyVersion.Error != "" {
	//} else {
	//	data.Error = types.StringNull()
	//}
	//
	//if keyVersion.RevokedAt != nil {
	//} else {
	//	data.RevokedAt = types.StringNull()
	//}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
