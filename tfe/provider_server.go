package tfe

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tftypes"
)

type providerServer struct {
	providerSchema     *tfprotov5.Schema
	providerMetaSchema *tfprotov5.Schema
	resourceSchemas    map[string]*tfprotov5.Schema
	dataSourceSchemas  map[string]*tfprotov5.Schema
	meta               providerMeta

	resourceRouter
	dataSourceRouter map[string]func(p *providerServer) tfprotov5.DataSourceServer
}

type providerMeta struct {
	token    string
	hostname string
}

func (p *providerServer) GetProviderSchema(ctx context.Context, req *tfprotov5.GetProviderSchemaRequest) (*tfprotov5.GetProviderSchemaResponse, error) {
	return &tfprotov5.GetProviderSchemaResponse{
		Provider:          p.providerSchema,
		ProviderMeta:      p.providerMetaSchema,
		ResourceSchemas:   p.resourceSchemas,
		DataSourceSchemas: p.dataSourceSchemas,
	}, nil
}

func (p *providerServer) PrepareProviderConfig(ctx context.Context, req *tfprotov5.PrepareProviderConfigRequest) (*tfprotov5.PrepareProviderConfigResponse, error) {
	return nil, nil
}

func (p *providerServer) ConfigureProvider(ctx context.Context, req *tfprotov5.ConfigureProviderRequest) (*tfprotov5.ConfigureProviderResponse, error) {
	resp := &tfprotov5.ConfigureProviderResponse{
		Diagnostics: []*tfprotov5.Diagnostic{},
	}
	meta, err := retrieveProviderMeta(req)
	if err != nil {
		resp.Diagnostics = append(resp.Diagnostics, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  "Error retrieving provider meta values from provider request",
			Detail:   fmt.Sprintf("Error retrieving provider meta values from provider request", err.Error()),
		})
		return resp, nil
	}
	p.meta = meta

	return resp, nil
}

func (p *providerServer) StopProvider(ctx context.Context, req *tfprotov5.StopProviderRequest) (*tfprotov5.StopProviderResponse, error) {
	return &tfprotov5.StopProviderResponse{}, nil
}

func (p *providerServer) ValidateDataSourceConfig(ctx context.Context, req *tfprotov5.ValidateDataSourceConfigRequest) (*tfprotov5.ValidateDataSourceConfigResponse, error) {
	log.Printf("[DEBUG] @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@OMAR DATA SOURCE ROUTER VALIDATE")
	log.Printf("[DEBUG] @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@OMAR DATA SOURCE ROUTER VALIDATE")
	log.Printf("[DEBUG] @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@OMAR DATA SOURCE ROUTER VALIDATE")
	log.Printf("[DEBUG] @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@OMAR DATA SOURCE ROUTER VALIDATE")
	log.Printf("[DEBUG] @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@OMAR DATA SOURCE ROUTER VALIDATE")
	ds, ok := p.dataSourceRouter[req.TypeName]
	if !ok {
		return nil, errUnsupportedDataSource(req.TypeName)
	}
	return ds(p).ValidateDataSourceConfig(ctx, req)
}

func (p *providerServer) ReadDataSource(ctx context.Context, req *tfprotov5.ReadDataSourceRequest) (*tfprotov5.ReadDataSourceResponse, error) {
	log.Printf("[DEBUG] @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@OMAR DATA SOURCE ROUTER READ")
	log.Printf("[DEBUG] @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@OMAR DATA SOURCE ROUTER READ")
	log.Printf("[DEBUG] @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@OMAR DATA SOURCE ROUTER READ")
	log.Printf("[DEBUG] @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@OMAR DATA SOURCE ROUTER READ")
	log.Printf("[DEBUG] @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@OMAR DATA SOURCE ROUTER READ")
	log.Printf("[DEBUG] @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@OMAR DATA SOURCE ROUTER READ")
	ds, ok := p.dataSourceRouter[req.TypeName]
	if !ok {
		return nil, errUnsupportedDataSource(req.TypeName)
	}
	return ds(p).ReadDataSource(ctx, req)
}

func ProviderServer() tfprotov5.ProviderServer {
	return &providerServer{
		providerSchema: &tfprotov5.Schema{
			Block: &tfprotov5.SchemaBlock{
				Attributes: []*tfprotov5.SchemaAttribute{
					&tfprotov5.SchemaAttribute{
						Name:        "hostname",
						Type:        tftypes.String,
						Description: descriptions["hostname"],
						Optional:    true,
					},
					&tfprotov5.SchemaAttribute{
						Name:        "token",
						Type:        tftypes.String,
						Description: descriptions["token"],
						Optional:    true,
					},
					&tfprotov5.SchemaAttribute{
						Name:        "ssl_skip_verify",
						Type:        tftypes.Bool,
						Description: descriptions["ssl_skip_verify"],
						Optional:    true,
					},
				},
			},
		},
		dataSourceSchemas: map[string]*tfprotov5.Schema{
			"tfe_state_outputs": {
				Version: 1,
				Block: &tfprotov5.SchemaBlock{
					Version: 1,
					Attributes: []*tfprotov5.SchemaAttribute{
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
							Name:     "values",
							Type:     tftypes.DynamicPseudoType,
							Optional: true,
							Computed: true,
						},
					},
				},
			},
		},
		dataSourceRouter: map[string]func(p *providerServer) tfprotov5.DataSourceServer{
			"tfe_state_outputs": newDataSourceRemoteState,

			//	"tfe_remote_state": dataSourceRemoteState{},
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
		return meta, fmt.Errorf("Could not set the value to map %v", err)
	}
	err = valMap["hostname"].As(&hostname)
	if err != nil {
		return meta, fmt.Errorf("Could not set the hostname value to string %v", err)
	}
	err = valMap["token"].As(&token)
	if err != nil {
		return meta, fmt.Errorf("Could not set the token value to string %v", err)
	}
	meta.hostname = hostname
	meta.token = token

	return meta, nil
}
