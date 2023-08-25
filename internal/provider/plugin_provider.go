// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"os"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type pluginProviderServer struct {
	providerSchema     *tfprotov5.Schema
	providerMetaSchema *tfprotov5.Schema
	resourceSchemas    map[string]*tfprotov5.Schema
	dataSourceSchemas  map[string]*tfprotov5.Schema
	tfeClient          *tfe.Client
	organization       string

	resourceRouter
	dataSourceRouter map[string]func(ConfiguredClient) tfprotov5.DataSourceServer
}

type errUnsupportedDataSource string

func (e errUnsupportedDataSource) Error() string {
	return "unsupported data source: " + string(e)
}

type errUnsupportedResource string

func (e errUnsupportedResource) Error() string {
	return "unsupported resource: " + string(e)
}

type providerMeta struct {
	token         string
	hostname      string
	sslSkipVerify bool
	organization  string
}

func (p *pluginProviderServer) GetProviderSchema(ctx context.Context, req *tfprotov5.GetProviderSchemaRequest) (*tfprotov5.GetProviderSchemaResponse, error) {
	return &tfprotov5.GetProviderSchemaResponse{
		Provider:          p.providerSchema,
		ProviderMeta:      p.providerMetaSchema,
		ResourceSchemas:   p.resourceSchemas,
		DataSourceSchemas: p.dataSourceSchemas,
	}, nil
}

func (p *pluginProviderServer) PrepareProviderConfig(ctx context.Context, req *tfprotov5.PrepareProviderConfigRequest) (*tfprotov5.PrepareProviderConfigResponse, error) {
	return nil, nil
}

func (p *pluginProviderServer) ConfigureProvider(ctx context.Context, req *tfprotov5.ConfigureProviderRequest) (*tfprotov5.ConfigureProviderResponse, error) {
	resp := &tfprotov5.ConfigureProviderResponse{
		Diagnostics: []*tfprotov5.Diagnostic{},
	}
	meta, err := retrieveProviderMeta(req)
	if err != nil {
		resp.Diagnostics = append(resp.Diagnostics, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  "Error retrieving provider meta values for internal provider.",
			Detail:   fmt.Sprintf("This should never happen; please report it to https://github.com/hashicorp/terraform-provider-tfe/issues\n\nThe error received was: %q", err.Error()),
		})
		return resp, nil
	}

	client, err := getClient(meta.hostname, meta.token, meta.sslSkipVerify)
	if err != nil {
		resp.Diagnostics = append(resp.Diagnostics, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  "Error getting client",
			Detail:   fmt.Sprintf("Error getting client: %v", err),
		})
		return resp, nil
	}

	if meta.organization == "" {
		meta.organization = os.Getenv("TFE_ORGANIZATION")
	}

	p.tfeClient = client
	p.organization = meta.organization
	return resp, nil
}

func (p *pluginProviderServer) StopProvider(ctx context.Context, req *tfprotov5.StopProviderRequest) (*tfprotov5.StopProviderResponse, error) {
	return &tfprotov5.StopProviderResponse{}, nil
}

func (p *pluginProviderServer) ValidateDataSourceConfig(ctx context.Context, req *tfprotov5.ValidateDataSourceConfigRequest) (*tfprotov5.ValidateDataSourceConfigResponse, error) {
	ds, ok := p.dataSourceRouter[req.TypeName]
	if !ok {
		return nil, errUnsupportedDataSource(req.TypeName)
	}
	return ds(ConfiguredClient{p.tfeClient, p.organization}).ValidateDataSourceConfig(ctx, req)
}

func (p *pluginProviderServer) ReadDataSource(ctx context.Context, req *tfprotov5.ReadDataSourceRequest) (*tfprotov5.ReadDataSourceResponse, error) {
	ds, ok := p.dataSourceRouter[req.TypeName]
	if !ok {
		return nil, errUnsupportedDataSource(req.TypeName)
	}
	return ds(ConfiguredClient{p.tfeClient, p.organization}).ReadDataSource(ctx, req)
}

type resourceRouter map[string]tfprotov5.ResourceServer

func (r resourceRouter) ValidateResourceTypeConfig(ctx context.Context, req *tfprotov5.ValidateResourceTypeConfigRequest) (*tfprotov5.ValidateResourceTypeConfigResponse, error) {
	res, ok := r[req.TypeName]
	if !ok {
		return nil, errUnsupportedResource(req.TypeName)
	}
	return res.ValidateResourceTypeConfig(ctx, req)
}

func (r resourceRouter) UpgradeResourceState(ctx context.Context, req *tfprotov5.UpgradeResourceStateRequest) (*tfprotov5.UpgradeResourceStateResponse, error) {
	res, ok := r[req.TypeName]
	if !ok {
		return nil, errUnsupportedResource(req.TypeName)
	}
	return res.UpgradeResourceState(ctx, req)
}

func (r resourceRouter) ReadResource(ctx context.Context, req *tfprotov5.ReadResourceRequest) (*tfprotov5.ReadResourceResponse, error) {
	res, ok := r[req.TypeName]
	if !ok {
		return nil, errUnsupportedResource(req.TypeName)
	}
	return res.ReadResource(ctx, req)
}

func (r resourceRouter) PlanResourceChange(ctx context.Context, req *tfprotov5.PlanResourceChangeRequest) (*tfprotov5.PlanResourceChangeResponse, error) {
	res, ok := r[req.TypeName]
	if !ok {
		return nil, errUnsupportedResource(req.TypeName)
	}
	return res.PlanResourceChange(ctx, req)
}

func (r resourceRouter) ApplyResourceChange(ctx context.Context, req *tfprotov5.ApplyResourceChangeRequest) (*tfprotov5.ApplyResourceChangeResponse, error) {
	res, ok := r[req.TypeName]
	if !ok {
		return nil, errUnsupportedResource(req.TypeName)
	}
	return res.ApplyResourceChange(ctx, req)
}

func (r resourceRouter) ImportResourceState(ctx context.Context, req *tfprotov5.ImportResourceStateRequest) (*tfprotov5.ImportResourceStateResponse, error) {
	res, ok := r[req.TypeName]
	if !ok {
		return nil, errUnsupportedResource(req.TypeName)
	}
	return res.ImportResourceState(ctx, req)
}

// PluginProviderServer returns the implementation of an interface for a lower
// level usage of the Provider to Terraform protocol.
// This relies on the terraform-plugin-go library, which provides low level
// bindings for the Terraform plugin protocol.
func PluginProviderServer() tfprotov5.ProviderServer {
	return &pluginProviderServer{
		providerSchema: &tfprotov5.Schema{
			Block: &tfprotov5.SchemaBlock{
				Attributes: []*tfprotov5.SchemaAttribute{
					{
						Name:        "hostname",
						Type:        tftypes.String,
						Description: descriptions["hostname"],
						Optional:    true,
					},
					{
						Name:        "token",
						Type:        tftypes.String,
						Description: descriptions["token"],
						Optional:    true,
					},
					{
						Name:        "ssl_skip_verify",
						Type:        tftypes.Bool,
						Description: descriptions["ssl_skip_verify"],
						Optional:    true,
					},
					{
						Name:        "organization",
						Type:        tftypes.String,
						Description: descriptions["organization"],
						Optional:    true,
					},
				},
			},
		},
		dataSourceSchemas: map[string]*tfprotov5.Schema{
			"tfe_outputs": {
				Version: 1,
				Block: &tfprotov5.SchemaBlock{
					Version: 1,
					Attributes: []*tfprotov5.SchemaAttribute{
						{
							Name:     "id",
							Type:     tftypes.String,
							Computed: true,
						},
						{
							Name:            "workspace",
							Type:            tftypes.String,
							Description:     "The workspace to fetch the remote state from.",
							DescriptionKind: tfprotov5.StringKindPlain,
							Required:        true,
						},
						{
							Name:            "organization",
							Type:            tftypes.String,
							Description:     "The organization to fetch the remote state from.",
							DescriptionKind: tfprotov5.StringKindPlain,
							Optional:        true,
						},
						{
							Name:      "values",
							Type:      tftypes.DynamicPseudoType,
							Optional:  true,
							Computed:  true,
							Sensitive: true,
						},
						{
							Name:      "nonsensitive_values",
							Type:      tftypes.DynamicPseudoType,
							Computed:  true,
							Sensitive: false,
						},
					},
				},
			},
		},
		dataSourceRouter: map[string]func(ConfiguredClient) tfprotov5.DataSourceServer{
			"tfe_outputs": newDataSourceOutputs,
		},
	}
}

func retrieveProviderMeta(req *tfprotov5.ConfigureProviderRequest) (providerMeta, error) {
	meta := providerMeta{}
	config := req.Config
	val, err := config.Unmarshal(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"hostname":        tftypes.String,
			"token":           tftypes.String,
			"ssl_skip_verify": tftypes.Bool,
			"organization":    tftypes.String,
		}})

	if err != nil {
		return meta, fmt.Errorf("could not unmarshal ConfigureProviderRequest %w", err)
	}
	var hostname string
	var token string
	var sslSkipVerify bool
	var organization string
	var valMap map[string]tftypes.Value
	err = val.As(&valMap)
	if err != nil {
		return meta, fmt.Errorf("could not set the schema attributes to map %w", err)
	}
	if !valMap["hostname"].IsNull() {
		err = valMap["hostname"].As(&hostname)
		if err != nil {
			return meta, fmt.Errorf("could not set the hostname value to string %w", err)
		}
	}
	if !valMap["token"].IsNull() {
		err = valMap["token"].As(&token)
		if err != nil {
			return meta, fmt.Errorf("could not set the token value to string %w", err)
		}
	}
	if !valMap["ssl_skip_verify"].IsNull() {
		err = valMap["ssl_skip_verify"].As(&sslSkipVerify)
		if err != nil {
			return meta, fmt.Errorf("could not set the ssl_skip_verify value to boolean %w", err)
		}
	} else {
		sslSkipVerify = defaultSSLSkipVerify
	}
	if !valMap["organization"].IsNull() {
		err = valMap["organization"].As(&organization)
		if err != nil {
			return meta, fmt.Errorf("failed to set the organization value to string: %w", err)
		}
	}

	meta.hostname = hostname
	meta.token = token
	meta.sslSkipVerify = sslSkipVerify
	meta.organization = organization

	return meta, nil
}
