package tfe

import (
	"context"
	"fmt"
	"os"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var tfeClient *tfe.Client

type pluginProviderServer struct {
	providerSchema     *tfprotov5.Schema
	providerMetaSchema *tfprotov5.Schema
	resourceSchemas    map[string]*tfprotov5.Schema
	dataSourceSchemas  map[string]*tfprotov5.Schema
	tfeClient          *tfe.Client

	resourceRouter
	dataSourceRouter map[string]func(*tfe.Client) tfprotov5.DataSourceServer
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
	token    string
	hostname string
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
			Summary:  "Error retrieving provider meta values from provider request",
			Detail:   fmt.Sprintf("Error retrieving provider meta values from provider request %v", err),
		})
		return resp, nil
	}

	client, err := getClient(meta.hostname, meta.token, false)
	if err != nil {
		resp.Diagnostics = append(resp.Diagnostics, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  "Error getting client",
			Detail:   fmt.Sprintf("Error getting client: %v", err),
		})
		return resp, nil
	}

	p.tfeClient = client
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
	return ds(p.tfeClient).ValidateDataSourceConfig(ctx, req)
}

func (p *pluginProviderServer) ReadDataSource(ctx context.Context, req *tfprotov5.ReadDataSourceRequest) (*tfprotov5.ReadDataSourceResponse, error) {
	ds, ok := p.dataSourceRouter[req.TypeName]
	if !ok {
		return nil, errUnsupportedDataSource(req.TypeName)
	}
	return ds(p.tfeClient).ReadDataSource(ctx, req)
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
							Required:        true,
						},
						{
							Name:      "values",
							Type:      tftypes.DynamicPseudoType,
							Optional:  true,
							Computed:  true,
							Sensitive: true,
						},
					},
				},
			},
		},
		dataSourceRouter: map[string]func(*tfe.Client) tfprotov5.DataSourceServer{
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
		}})

	if err != nil {
		return meta, fmt.Errorf("Could not unmarshal ConfigureProviderRequest %v", err)
	}
	var hostname string
	var token string
	var valMap map[string]tftypes.Value
	err = val.As(&valMap)
	if err != nil {
		return meta, fmt.Errorf("Could not set the schema attributes to map %v", err)
	}
	if !valMap["hostname"].IsNull() {
		err = valMap["hostname"].As(&hostname)
		if err != nil {
			return meta, fmt.Errorf("Could not set the hostname value to string %v", err)
		}
	}
	if !valMap["token"].IsNull() {
		err = valMap["token"].As(&token)
		if err != nil {
			return meta, fmt.Errorf("Could not set the token value to string %v", err)
		}
	}

	if hostname == "" && os.Getenv("TFE_HOSTNAME") != "" {
		hostname = os.Getenv("TFE_HOSTNAME")
	}

	if token == "" && os.Getenv("TFE_TOKEN") != "" {
		token = os.Getenv("TFE_TOKEN")
	}

	if hostname == "" || token == "" {
		return meta, fmt.Errorf("the hostname and token must be present.")
	}

	meta.hostname = hostname
	meta.token = token

	return meta, nil
}
