// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource              = &dataSourceTeam{}
	_ datasource.DataSourceWithConfigure = &dataSourceTeam{}
)

func NewTeamDataSource() datasource.DataSource {
	return &dataSourceTeam{}
}

type dataSourceTeam struct {
	config ConfiguredClient
}

type dataSourceTeamModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Organization types.String `tfsdk:"organization"`
	SSOTeamID    types.String `tfsdk:"sso_team_id"`
}

func (d *dataSourceTeam) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team"
}

func (d *dataSourceTeam) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This data source can be used to retrieve a team by name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the team.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the team.",
			},
			"organization": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Name of the organization. If omitted, organization must be defined in the provider config.",
			},
			"sso_team_id": schema.StringAttribute{
				Computed:    true,
				Description: "The SSO team identifier configured for the team.",
			},
		},
	}
}

func (d *dataSourceTeam) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *dataSourceTeam) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data dataSourceTeamModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var organization string
	resp.Diagnostics.Append(d.config.dataOrDefaultOrganization(ctx, req.Config, &organization)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading team", map[string]any{
		"name":         data.Name.ValueString(),
		"organization": organization,
	})

	team, err := d.findTeamByName(ctx, organization, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read team", err.Error())
		return
	}

	data.ID = types.StringValue(team.ID)
	data.Name = types.StringValue(team.Name)
	data.Organization = types.StringValue(organization)
	data.SSOTeamID = types.StringValue(team.SSOTeamID)

	tflog.Trace(ctx, "Read team successfully", map[string]any{
		"team_id":      team.ID,
		"name":         team.Name,
		"organization": organization,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *dataSourceTeam) findTeamByName(ctx context.Context, organization, name string) (*tfe.Team, error) {
	tl, err := d.config.Client.Teams.List(ctx, organization, &tfe.TeamListOptions{
		Names: []string{name},
	})
	if err != nil {
		return nil, fmt.Errorf("error retrieving teams: %w", err)
	}

	switch len(tl.Items) {
	case 0:
		return nil, fmt.Errorf("could not find team %s/%s", organization, name)
	case 1:
		// We check this just in case a user's TFE instance only has one team
		// and doesn't support the filter query param
		if tl.Items[0].Name != name {
			return nil, fmt.Errorf("could not find team %s/%s", organization, name)
		}

		return tl.Items[0], nil
	default:
		options := &tfe.TeamListOptions{}

		for {
			for i := range tl.Items {
				if tl.Items[i].Name == name {
					return tl.Items[i], nil
				}
			}

			if tl.CurrentPage >= tl.TotalPages {
				break
			}

			options.PageNumber = tl.NextPage

			tl, err = d.config.Client.Teams.List(ctx, organization, options)
			if err != nil {
				return nil, fmt.Errorf("error retrieving teams: %w", err)
			}
		}
	}

	return nil, fmt.Errorf("could not find team %s/%s", organization, name)
}
