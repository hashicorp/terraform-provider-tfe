// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	// Compile-time proof of interface implementation.
	_ datasource.DataSource              = &dataSourceTFEOAuthClient{}
	_ datasource.DataSourceWithConfigure = &dataSourceTFEOAuthClient{}
)

func NewOAuthClientDataSource() datasource.DataSource {
	return &dataSourceTFEOAuthClient{}
}

type dataSourceTFEOAuthClient struct {
	config ConfiguredClient
}

type modelDataSourceTFEOAuthClient struct {
	ID                         types.String `tfsdk:"id"`
	Name                       types.String `tfsdk:"name"`
	Organization               types.String `tfsdk:"organization"`
	OAuthClientID              types.String `tfsdk:"oauth_client_id"`
	ServiceProvider            types.String `tfsdk:"service_provider"`
	APIURL                     types.String `tfsdk:"api_url"`
	CallbackURL                types.String `tfsdk:"callback_url"`
	CreatedAt                  types.String `tfsdk:"created_at"`
	HTTPURL                    types.String `tfsdk:"http_url"`
	OAuthTokenID               types.String `tfsdk:"oauth_token_id"`
	ServiceProviderDisplayName types.String `tfsdk:"service_provider_display_name"`
	OrganizationScoped         types.Bool   `tfsdk:"organization_scoped"`
	ProjectIDs                 types.Set    `tfsdk:"project_ids"`
}

func modelDataSourceFromTFEOAuthClient(ctx context.Context, c *tfe.OAuthClient) (*modelDataSourceTFEOAuthClient, diag.Diagnostics) {
	var diags diag.Diagnostics
	m := modelDataSourceTFEOAuthClient{
		ID:                         types.StringValue(c.ID),
		Name:                       types.StringPointerValue(c.Name),
		Organization:               types.StringValue(c.Organization.Name),
		OAuthClientID:              types.StringValue(c.ID),
		ServiceProvider:            types.StringValue(string(c.ServiceProvider)),
		APIURL:                     types.StringValue(c.APIURL),
		CallbackURL:                types.StringValue(c.CallbackURL),
		CreatedAt:                  types.StringValue(c.CreatedAt.Format(time.RFC3339)),
		OrganizationScoped:         types.BoolPointerValue(c.OrganizationScoped),
		HTTPURL:                    types.StringValue(c.HTTPURL),
		ServiceProviderDisplayName: types.StringValue(c.ServiceProviderName),
	}

	// Set project IDs
	projectIDs := make([]string, len(c.Projects))
	for i, project := range c.Projects {
		projectIDs[i] = project.ID
	}

	projectIDSet, diags := types.SetValueFrom(ctx, types.StringType, projectIDs)
	if diags.HasError() {
		return nil, diags
	}
	m.ProjectIDs = projectIDSet

	// Set OAuth token ID
	switch len(c.OAuthTokens) {
	case 0:
		m.OAuthTokenID = types.StringValue("")
	case 1:
		m.OAuthTokenID = types.StringValue(c.OAuthTokens[0].ID)
	default:
		diags.AddError("Error parsing API result", fmt.Sprintf("unexpected number of OAuth tokens: %d", len(c.OAuthTokens)))
	}

	return &m, diags
}

// Configure implements datasource.ResourceWithConfigure
func (d *dataSourceTFEOAuthClient) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Early exit if provider is unconfigured (i.e. we're only validating config or something)
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(ConfiguredClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected resource Configure type",
			fmt.Sprintf("Expected tfe.ConfiguredClient, got %T. This is a bug in the tfe provider, so please report it on GitHub.", req.ProviderData),
		)
	}
	d.config = client
}

// Metadata implements datasource.Resource
func (d *dataSourceTFEOAuthClient) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_oauth_client"
}

// Schema implements datasource.Resource
func (d *dataSourceTFEOAuthClient) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Service-generated identifier for the variable",
			},

			"organization": schema.StringAttribute{
				Description: "Name of the organization",
				Computed:    true,
				Optional:    true,
			},

			"name": schema.StringAttribute{
				Description: "Display name for the OAuth Client",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.AtLeastOneOf(
						path.MatchRoot("oauth_client_id"),
						path.MatchRoot("service_provider"),
					),
				},
			},

			"oauth_client_id": schema.StringAttribute{
				Description: "OAuth Token ID for the OAuth Client",
				Optional:    true,
			},

			"service_provider": schema.StringAttribute{
				Description: "The VCS provider being connected with",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(tfe.ServiceProviderGithub),
						string(tfe.ServiceProviderGithubEE),
						string(tfe.ServiceProviderGitlab),
						string(tfe.ServiceProviderGitlabCE),
						string(tfe.ServiceProviderGitlabEE),
						string(tfe.ServiceProviderBitbucket),
						string(tfe.ServiceProviderBitbucketServer),
						string(tfe.ServiceProviderBitbucketServerLegacy),
						string(tfe.ServiceProviderBitbucketDataCenter),
						string(tfe.ServiceProviderAzureDevOpsServer),
						string(tfe.ServiceProviderAzureDevOpsServices),
					),
				},
			},

			"api_url": schema.StringAttribute{
				Description: "The base URL of the VCS provider's API",
				Computed:    true,
			},

			"callback_url": schema.StringAttribute{
				Description: "The base URL of the VCS provider's API",
				Computed:    true,
			},

			"created_at": schema.StringAttribute{
				Description: "The base URL of the VCS provider's API",
				Computed:    true,
			},

			"http_url": schema.StringAttribute{
				Description: "The homepage of the VCS provider",
				Computed:    true,
			},

			"oauth_token_id": schema.StringAttribute{
				Description: "OAuth Token ID for the OAuth Client",
				Computed:    true,
			},

			"service_provider_display_name": schema.StringAttribute{
				Description: "The display name of the VCS provider",
				Computed:    true,
			},

			"organization_scoped": schema.BoolAttribute{
				Description: "Whether or not the oauth client is scoped to all projects and workspaces in the organization",
				Computed:    true,
			},

			"project_ids": schema.SetAttribute{
				Description: "The IDs of the projects that the OAuth client is associated with",
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

// Read implements datasource.Resource
func (d *dataSourceTFEOAuthClient) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Load the config into the model
	var config modelDataSourceTFEOAuthClient
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the organization name from the data source or provider config
	var organization string
	resp.Diagnostics.Append(d.config.dataOrDefaultOrganization(ctx, req.Config, &organization)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := config.OAuthClientID.ValueString()
	name := config.Name.ValueString()
	serviceProvider := tfe.ServiceProviderType(config.ServiceProvider.ValueString())

	var err error
	var oc *tfe.OAuthClient
	tflog.Debug(ctx, fmt.Sprintf("Read OAuth client: %s", id))

	if !config.OAuthClientID.IsNull() {
		// Read the OAuth client using its ID
		oc, err = d.config.Client.OAuthClients.Read(ctx, id)
		if err != nil && errors.Is(err, tfe.ErrResourceNotFound) {
			tflog.Debug(ctx, fmt.Sprintf("OAuth client %s no longer exists", id))
			resp.State.RemoveResource(ctx)
			return
		}

		if err != nil {
			resp.Diagnostics.AddError("Error reading OAuth client", err.Error())
			return
		}
	} else {
		// Read the OAuth client using its name or service provider
		oc, err = fetchOAuthClientByNameOrServiceProvider(ctx, d.config.Client, organization, name, serviceProvider)
		if err != nil {
			resp.Diagnostics.AddError("Error reading OAuth client", err.Error())
			return
		}
	}

	// Load the result into the model
	result, diags := modelDataSourceFromTFEOAuthClient(ctx, oc)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Update state
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}
