// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-tfe/internal/client"
)

// frameworkProvider is a type that implements the terraform-plugin-framework
// provider.Provider interface. Someday, this will probably encompass the entire
// behavior of the tfe provider. Today, it is a small but growing subset.
type frameworkProvider struct{}

// Compile-time interface check
var _ provider.Provider = &frameworkProvider{}

// FrameworkProviderConfig is a helper type for extracting the provider
// configuration from the provider block.
type FrameworkProviderConfig struct {
	Hostname      types.String `tfsdk:"hostname"`
	Token         types.String `tfsdk:"token"`
	Organization  types.String `tfsdk:"organization"`
	SSLSkipVerify types.Bool   `tfsdk:"ssl_skip_verify"`
}

// NewFrameworkProvider is a helper function for initializing the portion of
// the tfe provider implemented via the terraform-plugin-framework.
func NewFrameworkProvider() provider.Provider {
	return &frameworkProvider{}
}

// Metadata (a Provider interface function) lets the provider identify itself.
// Resources and data sources can access this information from their request
// objects.
func (p *frameworkProvider) Metadata(_ context.Context, _ provider.MetadataRequest, res *provider.MetadataResponse) {
	res.TypeName = "tfe"
}

// Schema (a Provider interface function) returns the schema for the Terraform
// block that configures the provider itself.
func (p *frameworkProvider) Schema(_ context.Context, _ provider.SchemaRequest, res *provider.SchemaResponse) {
	res.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"hostname": schema.StringAttribute{
				Description: descriptions["hostname"],
				Optional:    true,
			},
			"token": schema.StringAttribute{
				Optional:    true,
				Description: descriptions["token"],
				// TODO: should be sensitive, but that's a breaking change.
			},
			"organization": schema.StringAttribute{
				Description: descriptions["organization"],
				Optional:    true,
			},
			"ssl_skip_verify": schema.BoolAttribute{
				Description: descriptions["ssl_skip_verify"],
				Optional:    true,
			},
		},
	}
}

// Configure (a Provider interface function) sets up the TFC client per the
// specified provider configuration block and env vars.
func (p *frameworkProvider) Configure(ctx context.Context, req provider.ConfigureRequest, res *provider.ConfigureResponse) {
	var data FrameworkProviderConfig
	diags := req.Config.Get(ctx, &data)

	res.Diagnostics.Append(diags...)
	if res.Diagnostics.HasError() {
		return
	}

	// TODO: add .IsUnknown() error handling in the case where user passed
	// unresolvable references for provider config values, c.f. what hashicups
	// does around
	// https://github.com/hashicorp/terraform-provider-hashicups-pf/blob/main/hashicups/provider.go#L79

	// Read default organization from environment if it wasn't set in the
	// config. All other env defaults are handled by getClient().
	if data.Organization.IsNull() {
		// Falling back to Getenv will collapse the new type system's handling
		// of null/unknown into a plain zero-value, but that's OK at this point.
		data.Organization = types.StringValue(os.Getenv("TFE_ORGANIZATION"))
	}

	client, err := client.GetClient(data.Hostname.ValueString(), data.Token.ValueString(), data.SSLSkipVerify.ValueBool())

	if err != nil {
		res.Diagnostics.AddError("Failed to initialize HTTP client", err.Error())
		return
	}

	configuredClient := ConfiguredClient{
		Client:       client,
		Organization: data.Organization.ValueString(),
	}

	res.DataSourceData = configuredClient
	res.ResourceData = configuredClient
}

func (p *frameworkProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewSAMLSettingsDataSource,
	}
}

func (p *frameworkProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewResourceVariable,
		NewSAMLSettingsResource,
	}
}
