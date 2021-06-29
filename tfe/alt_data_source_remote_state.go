package tfe

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tftypes"
)

type dataSourceRemoteState struct{}

var stderr *os.File

func init() {
	stderr = os.Stderr
}

func (d dataSourceRemoteState) ReadDataSource(ctx context.Context, req *tfprotov5.ReadDataSourceRequest) (*tfprotov5.ReadDataSourceResponse, error) {
	log.Printf("[DEBUG] @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@OMAR")
	log.Printf("[DEBUG] !!!!!!!!!!!!!!!!!!!!!!!!!!OMAR")
	config := req.Config
	val, err := config.Unmarshal(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"workspace":    tftypes.String,
			"download_url": tftypes.String,
		}})
	if err != nil {
		return &tfprotov5.ReadDataSourceResponse{Diagnostics: []*tfprotov5.Diagnostic{}}, err
	}
	log.Printf("[DEBUG] @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@OMAR VAL %s", val)
	state, err := tfprotov5.NewDynamicValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"workspace":    tftypes.String,
			"download_url": tftypes.String,
		},
	}, tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"workspace":    tftypes.String,
			"download_url": tftypes.String,
		},
	}, map[string]tftypes.Value{
		"workspace":    tftypes.NewValue(tftypes.String, "foobar"),
		"download_url": tftypes.NewValue(tftypes.String, "static_id"),
	}))

	//state, err := tfprotov5.NewDynamicValue(tftypes.Object{
	//	AttributeTypes: map[string]tftypes.Type{
	//		"download_url": tftypes.String,
	//	},
	//}, tftypes.NewValue(tftypes.Object{
	//	AttributeTypes: map[string]tftypes.Type{
	//		"download_url": tftypes.String,
	//	},
	//}, map[string]tftypes.Value{
	//	"download_url": tftypes.NewValue(tftypes.String, "foobar"),
	//}))
	if err != nil {
		return &tfprotov5.ReadDataSourceResponse{
			Diagnostics: []*tfprotov5.Diagnostic{
				{
					Severity: tfprotov5.DiagnosticSeverityError,
					Summary:  "Error encoding state",
					Detail:   fmt.Sprintf("Error encoding state: %s", err.Error()),
				},
			},
		}, nil
	}
	return &tfprotov5.ReadDataSourceResponse{
		State: &state,
	}, nil
}

func (d dataSourceRemoteState) ValidateDataSourceConfig(ctx context.Context, req *tfprotov5.ValidateDataSourceConfigRequest) (*tfprotov5.ValidateDataSourceConfigResponse, error) {
	return &tfprotov5.ValidateDataSourceConfigResponse{}, nil
}
