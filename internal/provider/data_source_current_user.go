// Copyright IBM Corp. 2018, 2025
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
	ID        types.String `tfsdk:"id"`
	Username  types.String `tfsdk:"username"`
	Email     types.String `tfsdk:"email"`
	AvatarURL types.String `tfsdk:"avatar_url"`
}

func (d *dataSourceCurrentUser) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_current_user"
}

func (d *dataSourceCurrentUser) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get the current user associated with the API token.",

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
			"avatar_url": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Avatar URL of the current user.",
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

	user, err := d.config.Client.Users.ReadCurrent(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read current user", err.Error())
		return
	}

	model := modelCurrentUser{
		ID:        types.StringValue(user.ID),
		Username:  types.StringValue(user.Username),
		Email:     types.StringValue(user.Email),
		AvatarURL: types.StringValue(user.AvatarURL),
	}

	tflog.Trace(ctx, "Read current user successfully", map[string]any{
		"user_id":  user.ID,
		"username": user.Username,
	})

	diags := resp.State.Set(ctx, &model)
	resp.Diagnostics.Append(diags...)
}
