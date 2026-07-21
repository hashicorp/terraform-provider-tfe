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

var (
	_ datasource.DataSource              = &dataSourceCurrentUser{}
	_ datasource.DataSourceWithConfigure = &dataSourceCurrentUser{}
)

func NewCurrentUserDataSource() datasource.DataSource {
	return &dataSourceCurrentUser{}
}

type dataSourceCurrentUser struct {
	config ConfiguredClient
}

type modelCurrentUser struct {
	ID               types.String `tfsdk:"id"`
	Username         types.String `tfsdk:"username"`
	Email            types.String `tfsdk:"email"`
	IsServiceAccount types.Bool   `tfsdk:"is_service_account"`
}

func (d *dataSourceCurrentUser) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_current_user"
}

func (d *dataSourceCurrentUser) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get the current user associated with the API token. When authenticated with a team or organization token, HCP Terraform returns a synthetic service user rather than a real user account, so attributes like `email` and `username` will not reflect a real person.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Service-generated identifier for the user.",
			},
			"username": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The username of the current user.",
			},
			"email": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The email address of the current user.",
			},
			"is_service_account": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the current user is a service account.",
			},
		},
	}
}

func (d *dataSourceCurrentUser) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(ConfiguredClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected ConfiguredClient, got %T. This is a bug in the tfe provider, so please report it on GitHub.", req.ProviderData),
		)

		return
	}

	d.config = client
}

func (d *dataSourceCurrentUser) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Debug(ctx, "Reading current user")

	userEnvelope, err := d.config.ClientV2.API.Account().Details().Get(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read current user", err.Error())
		return
	}

	user := userEnvelope.GetData()
	if user == nil {
		resp.Diagnostics.AddError("Unable to read current user", "The API response contained no user data.")
		return
	}

	var username, email string
	var isServiceAccount bool
	if attributes := user.GetAttributes(); attributes != nil {
		username = valueOrZero(attributes.GetUsername())
		email = valueOrZero(attributes.GetEmail())
		isServiceAccount = valueOrZero(attributes.GetIsServiceAccount())
	}

	model := modelCurrentUser{
		ID:               types.StringValue(valueOrZero(user.GetId())),
		Username:         types.StringValue(username),
		Email:            types.StringValue(email),
		IsServiceAccount: types.BoolValue(isServiceAccount),
	}

	tflog.Trace(ctx, "Read current user successfully", map[string]any{
		"user_id":            valueOrZero(user.GetId()),
		"username":           username,
		"is_service_account": isServiceAccount,
	})

	diags := resp.State.Set(ctx, &model)
	resp.Diagnostics.Append(diags...)
}
