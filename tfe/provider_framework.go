package tfe

import (
	"context"

	"github.com/hashicorp/terraform-provider-tfe/internal/admin"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// FrameworkProvider is the implementation of the provider using the Terraform Plugin Framework
type FrameworkProvider struct {
	version string
}

// ProviderData stores the provider block configuration values
type ProviderData struct {
	Hostname      types.String `tfsdk:"hostname"`
	Token         types.String `tfsdk:"token"`
	SSLSkipVerify types.Bool   `tfsdk:"ssl_skip_verify"`
}

// Configure configures the provider with the client SDK using provider block data
func (p *FrameworkProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data ProviderData
	diags := req.Config.Get(ctx, &data)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := getClient(data.Hostname.String(), data.Token.String(), data.SSLSkipVerify.ValueBool())

	if err != nil {
		resp.Diagnostics.AddError("Failed to initialize HTTP client", err.Error())
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

// GetSchema returns the schema for the provider block configuration
func (p *FrameworkProvider) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"hostname": {
				Description: "The Terraform Enterprise hostname to connect to. Defaults to app.terraform.io.",
				Optional:    true,
				Type:        types.StringType,
			},
			"token": {
				Description: "The token used to authenticate with Terraform Enterprise. We recommend omitting\n" +
					"the token which can be set as credentials in the CLI config file.",
				Optional: true,
				Type:     types.StringType,
			},
			"ssl_skip_verify": {
				Description: "Whether or not to skip certificate verifications.",
				Optional:    true,
				Type:        types.BoolType,
			},
		},
	}, nil
}

// Resources is the list of resources available to the provider
func (p *FrameworkProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
	}
}

// Data Sources is the list of data sources available to the provider
func (p *FrameworkProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

// New returns a new provider implementation
func NewFrameworkProvider(version string) provider.Provider {
	return &FrameworkProvider{
		version: version,
	}
}
