package tfe

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"os"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tftypes"
)

type dataSourceRemoteState struct {
	meta providerMeta
}

var stderr *os.File

func init() {
	stderr = os.Stderr
}

func (d dataSourceRemoteState) ReadDataSource(ctx context.Context, req *tfprotov5.ReadDataSourceRequest) (*tfprotov5.ReadDataSourceResponse, error) {
	log.Printf("[DEBUG] @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@OMAR")
	log.Printf("[DEBUG] @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@OMAR ---------META")
	log.Printf("[DEBUG] @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@OMAR ---------META")
	log.Printf("[DEBUG] @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@OMAR ---------META: %v", d.meta)
	hostname, token, err := retrieveMeta(req)
	if err != nil {
		return nil, err
	}

	client, err := getClient(hostname, token, false)
	if err != nil {
		return nil, err
	}

	orgName, wsName, err := retrieveValues(req)
	if err != nil {
		return &tfprotov5.ReadDataSourceResponse{
			Diagnostics: []*tfprotov5.Diagnostic{
				{
					Severity: tfprotov5.DiagnosticSeverityError,
					Summary:  "Error retrieving values from the config",
					Detail:   fmt.Sprintf("Error retrieving values from the config: %s", err.Error()),
				},
			},
		}, nil
	}

	_ = d.readRemoteStateOutput(ctx, client, orgName, wsName)

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
			"organization": tftypes.String,
			"state_output": tftypes.DynamicPseudoType,
		},
	}, tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"workspace":    tftypes.String,
			"organization": tftypes.String,
			"state_output": tftypes.Object{AttributeTypes: stateTypes},
		},
	}, map[string]tftypes.Value{
		"workspace":    tftypes.NewValue(tftypes.String, wsName),
		"organization": tftypes.NewValue(tftypes.String, orgName),
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

func retrieveValues(req *tfprotov5.ReadDataSourceRequest) (string, string, error) {
	var orgName string
	var wsName string
	var err error

	config := req.Config
	val, err := config.Unmarshal(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"workspace":    tftypes.String,
			"organization": tftypes.String,
			"state_output": tftypes.String,
		}})
	if err != nil {
		return orgName, wsName, fmt.Errorf("Error unmarshalling config: %v", err)
	}

	var valMap map[string]tftypes.Value
	err = val.As(&valMap)
	if err != nil {
		return orgName, wsName, fmt.Errorf("Error assigning Value to Golang map: %v", err)
	}

	err = valMap["organization"].As(&orgName)
	if err != nil {
		return orgName, wsName, fmt.Errorf("Error assigning 'organization' Value to Golang string: %v", err)
	}
	err = valMap["workspace"].As(&wsName)
	if err != nil {
		return orgName, wsName, fmt.Errorf("Error assigning 'workspace' Value to Golang string: %v", err)
	}

	return orgName, wsName, nil
}

func retrieveMeta(req *tfprotov5.ReadDataSourceRequest) (string, string, error) {
	var hostname string
	var token string
	config := req.ProviderMeta
	log.Printf("[DEBUG] @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@OMAR META CONFIG %v", config)
	val, err := config.Unmarshal(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"hostname":        tftypes.String,
			"token":           tftypes.String,
			"ssl_skip_verify": tftypes.Bool,
		}})
	if err != nil {
		return hostname, token, fmt.Errorf("Error unmarshalling config: %v", err)
	}
	log.Printf("[DEBUG] @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@OMAR META VAL %v", val)

	var valMap map[string]tftypes.Value
	err = val.As(&valMap)
	if err != nil {
		return hostname, token, fmt.Errorf("Error assigning Value to Golang map: %v", err)
	}

	log.Printf("[DEBUG] @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@OMAR META VAL %v", valMap)

	err = valMap["hostname"].As(&hostname)
	if err != nil {
		return hostname, token, fmt.Errorf("Error assigning 'organization' Value to Golang string: %v", err)
	}
	err = valMap["token"].As(&token)
	if err != nil {
		return hostname, token, fmt.Errorf("Error assigning 'workspace' Value to Golang string: %v", err)
	}

	log.Printf("[DEBUG] @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@OMAR META VAL hostname %s", hostname)
	log.Printf("[DEBUG] @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@OMAR META VAL token %s", token)

	hostname = "app.terraform.io"
	token = "foo"
	return hostname, token, nil

}

type remoteStateFile struct {
	Outputs map[string]interface{} `json:"outputs"`
}

type outputValue struct {
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

func (d dataSourceRemoteState) readRemoteStateOutput(ctx context.Context, tfeClient *tfe.Client, orgName, wsName string) error {

	ws, err := tfeClient.Workspaces.Read(ctx, orgName, wsName)
	if err != nil {
		return fmt.Errorf("Error reading workspace: %v", err)
	}

	sv, err := tfeClient.StateVersions.Current(ctx, ws.ID)
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return fmt.Errorf("Could not read  remote state for workspace '%s'", wsName)
		}
		return fmt.Errorf("Error remote state: %v", err)
	}

	log.Printf("[DEBUG] Setting Remote State Output")

	stateData, err := tfeClient.StateVersions.Download(ctx, sv.DownloadURL)
	if err != nil {
		return fmt.Errorf("Error downloading remote state: %v", err)
	}
	stateOuptput := &remoteStateFile{}
	if err := json.Unmarshal(stateData, stateOuptput); err != nil {
		return err
	}
	log.Printf("[DEBUG] @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@OMAR STATE OUTPUT: %v", stateOuptput)

	for k, v := range stateOuptput.Outputs {
		log.Printf("[DEBUG] ============OMAR=========== STATE KEY: %s", k)
		log.Printf("[DEBUG] ============OMAR=========== STATE VALUE: %s", v)
	}

	return nil
}
