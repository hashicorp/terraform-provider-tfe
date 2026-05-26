// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &dataSourceTFESCIMToken{}
	_ datasource.DataSourceWithConfigure = &dataSourceTFESCIMToken{}
)

// NewSCIMTokenDataSource is a helper function to simplify the provider implementation.
func NewSCIMTokenDataSource() datasource.DataSource {
	return &dataSourceTFESCIMToken{}
}

// dataSourceTFESCIMToken is the data source implementation.
type dataSourceTFESCIMToken struct {
	client *tfe.Client
}

// modelDataTFESCIMToken maps the data source schema data.
type modelDataTFESCIMToken struct {
	ID          types.String `tfsdk:"id"`
	Description types.String `tfsdk:"description"`
	ExpiredAt   types.String `tfsdk:"expired_at"`
	CreatedAt   types.String `tfsdk:"created_at"`
	LastUsedAt  types.String `tfsdk:"last_used_at"`
}

// Metadata returns the data source type name.
func (d *dataSourceTFESCIMToken) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_scim_token"
}

// Schema defines the schema for the data source.
func (d *dataSourceTFESCIMToken) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads a SCIM authentication token by its ID. Requires SCIM to be enabled.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the SCIM token",
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^at-`),
						"must be a valid SCIM token ID starting with 'at-'",
					),
				},
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "The description of the SCIM token.",
			},
			"expired_at": schema.StringAttribute{
				Computed:    true,
				Description: "The time when the SCIM token expires.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "The time when the SCIM token was created.",
			},
			"last_used_at": schema.StringAttribute{
				Computed:    true,
				Description: "The time when the SCIM token was last used.",
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *dataSourceTFESCIMToken) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *dataSourceTFESCIMToken) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data modelDataTFESCIMToken
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	scimTokenID := data.ID.ValueString()
	tflog.Debug(ctx, fmt.Sprintf("Reading SCIM Token %s", scimTokenID))

	scimToken, err := d.client.Admin.Settings.SCIM.Tokens.Read(ctx, scimTokenID)
	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
			resp.Diagnostics.AddError(
				fmt.Sprintf("SCIM token %s not found", scimTokenID),
				fmt.Sprintf("No SCIM token exists with ID %s. Verify the ID is correct and that SCIM is enabled on this Terraform Enterprise instance.", scimTokenID),
			)
			return
		}
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error reading SCIM token %s", scimTokenID),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &modelDataTFESCIMToken{
		ID:          types.StringValue(scimToken.ID),
		Description: types.StringValue(scimToken.Description),
		ExpiredAt:   timeStringOrNull(scimToken.ExpiredAt),
		CreatedAt:   timeStringOrNull(scimToken.CreatedAt),
		LastUsedAt:  timeStringOrNull(scimToken.LastUsedAt),
	})...)
}
