package tfe

import (
	"context"
	"fmt"
	"log"
	"math/big"
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
			"state_output": tftypes.String,
		}})
	if err != nil {
		return &tfprotov5.ReadDataSourceResponse{Diagnostics: []*tfprotov5.Diagnostic{}}, err
	}
	log.Printf("[DEBUG] @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@OMAR VAL %s", val)
	/*
		stateFile := map[string]interface{}{
			"foo":   []interface{}{"a", "b", "c"},
			"hello": "world",
		}
	*/
	stateFile := map[string]interface{}{
		"foo":   []interface{}{"a", "b", "c"},
		"hello": int64(123),
		"quuz":  false,
	}
	tftypesState := map[string]tftypes.Value{}
	stateTypes := map[string]tftypes.Type{}
	for k, v := range stateFile {
		switch val := v.(type) {
		case string:
			tftypesState[k] = tftypes.NewValue(tftypes.String, val)
			stateTypes[k] = tftypes.String
		case int64:
			tftypesState[k] = tftypes.NewValue(tftypes.Number, big.NewFloat(float64(val)))
			stateTypes[k] = tftypes.Number
		case bool:
			tftypesState[k] = tftypes.NewValue(tftypes.Bool, val)
			stateTypes[k] = tftypes.Bool
		case []interface{}:
			elements := []tftypes.Value{}
			types := []tftypes.Type{}
			for _, element := range val {
				switch el := element.(type) {
				case string:
					elements = append(elements, tftypes.NewValue(tftypes.String, el))
					types = append(types, tftypes.String)
				default:
					panic(fmt.Sprintf("unknown type %T", element))
				}
			}
			tftypesState[k] = tftypes.NewValue(tftypes.Tuple{ElementTypes: types}, elements)
			stateTypes[k] = tftypes.Tuple{ElementTypes: types}
		}
	}

	state, err := tfprotov5.NewDynamicValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"workspace":    tftypes.String,
			"state_output": tftypes.DynamicPseudoType,
		},
	}, tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"workspace":    tftypes.String,
			"state_output": tftypes.Object{AttributeTypes: stateTypes},
		},
	}, map[string]tftypes.Value{
		"workspace":    tftypes.NewValue(tftypes.String, "foobar"),
		"state_output": tftypes.NewValue(tftypes.Object{AttributeTypes: stateTypes}, tftypesState),
	}))

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
