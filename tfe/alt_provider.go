package tfe

import (
	"context"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tftypes"
)

type altserver struct {
	providerSchema     *tfprotov5.Schema
	providerMetaSchema *tfprotov5.Schema
	resourceSchemas    map[string]*tfprotov5.Schema
	dataSourceSchemas  map[string]*tfprotov5.Schema

	resourceRouter
	dataSourceRouter
}

func (s altserver) GetProviderSchema(ctx context.Context, req *tfprotov5.GetProviderSchemaRequest) (*tfprotov5.GetProviderSchemaResponse, error) {
	return &tfprotov5.GetProviderSchemaResponse{
		Provider:          s.providerSchema,
		ProviderMeta:      s.providerMetaSchema,
		ResourceSchemas:   s.resourceSchemas,
		DataSourceSchemas: s.dataSourceSchemas,
	}, nil
}

func (s altserver) PrepareProviderConfig(ctx context.Context, req *tfprotov5.PrepareProviderConfigRequest) (*tfprotov5.PrepareProviderConfigResponse, error) {
	return &tfprotov5.PrepareProviderConfigResponse{
		PreparedConfig: req.Config,
	}, nil
}

func (s altserver) ConfigureProvider(ctx context.Context, req *tfprotov5.ConfigureProviderRequest) (*tfprotov5.ConfigureProviderResponse, error) {
	var diags []*tfprotov5.Diagnostic
	return &tfprotov5.ConfigureProviderResponse{
		Diagnostics: diags,
	}, nil
}

func (s altserver) StopProvider(ctx context.Context, req *tfprotov5.StopProviderRequest) (*tfprotov5.StopProviderResponse, error) {
	return &tfprotov5.StopProviderResponse{}, nil
}

func AltServer() tfprotov5.ProviderServer {
	return altserver{
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
						Name:        "ssl_skip_verify",
						Type:        tftypes.Bool,
						Description: descriptions["ssl_skip_verify"],
						Optional:    true,
					},
					&tfprotov5.SchemaAttribute{
						Name:        "token",
						Type:        tftypes.String,
						Description: descriptions["token"],
						Optional:    true,
					},
				},
			},
		},
		dataSourceSchemas: map[string]*tfprotov5.Schema{
			"corner_time": {
				Version: 1,
				Block: &tfprotov5.SchemaBlock{
					Version: 1,
					Attributes: []*tfprotov5.SchemaAttribute{
						{
							Name:            "current",
							Type:            tftypes.String,
							Description:     "The current time in RFC3339 format.",
							DescriptionKind: tfprotov5.StringKindPlain,
							Computed:        true,
						},
						{
							Name:     "id",
							Type:     tftypes.String,
							Optional: true,
							Computed: true,
						},
					},
				},
			},
		},
		dataSourceRouter: dataSourceRouter{
			"corner_time": dataSourceTime{},
		},
	}
}
