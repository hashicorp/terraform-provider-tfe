package tfe

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	//"math/big"
	"os"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tftypes"
	"github.com/zclconf/go-cty/cty"
	//gocty "github.com/zclconf/go-cty/cty/gocty"
	ctyjson "github.com/zclconf/go-cty/cty/json"
)

type dataSourceRemoteState struct {
	provider *providerServer
}

var stderr *os.File

func init() {
	stderr = os.Stderr
}

func newDataSourceRemoteState(p *providerServer) tfprotov5.DataSourceServer {
	return dataSourceRemoteState{
		provider: p,
	}
}

func (d dataSourceRemoteState) ReadDataSource(ctx context.Context, req *tfprotov5.ReadDataSourceRequest) (*tfprotov5.ReadDataSourceResponse, error) {
	resp := &tfprotov5.ReadDataSourceResponse{
		Diagnostics: []*tfprotov5.Diagnostic{},
	}
	log.Printf("[DEBUG] @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@OMAR: (dataSourceRemoteState) ReadDataSource")
	log.Printf("[DEBUG] @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@OMAR: (dataSourceRemoteState) ReadDataSource")
	log.Printf("[DEBUG] @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@OMAR: (dataSourceRemoteState) ReadDataSource")
	client, err := getClient(d.provider.meta.hostname, d.provider.meta.token, false)
	if err != nil {
		return nil, err
	}

	orgName, wsName, err := retrieveValues(req)
	if err != nil {
		resp.Diagnostics = append(resp.Diagnostics, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  "Error retrieving values from the config",
			Detail:   fmt.Sprintf("Error retrieving values from the config: %v", err),
		})
		return resp, nil
	}

	remoteStateOutput, err := d.readRemoteStateOutput(ctx, client, orgName, wsName)
	log.Printf("[DEBUG] @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@OMAR: remoteStateOutput: ", remoteStateOutput)
	if err != nil {
		resp.Diagnostics = append(resp.Diagnostics, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  "Error reading remote state output",
			Detail:   fmt.Sprintf("Error reading remote state output: %v", err),
		})
		return resp, nil
	}
	//parsedStateOutput := parseRemoteStateOutput(remoteStateOutput)
	//log.Printf("[DEBUG] @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@OMAR: parsedStateOutput: ", parsedStateOutput)

	tftypesState := map[string]tftypes.Value{}
	stateTypes := map[string]tftypes.Type{}

	for name, outputValues := range remoteStateOutput.OutputValues {
		marshData, err := outputValues.Value.Type().MarshalJSON()
		if err != nil {
			log.Printf("[DEBUG] @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@OMAR: EEROOOOOOOOORRRRRRRR %v", err)
			//return err
		}
		tfType, err := tftypes.ParseJSONType(marshData)
		if err != nil {
			log.Printf("[DEBUG] @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@OMAR: EEROOOOOOOOORRRRRRRR %v", err)
			//return err
		}
		mByte, err := ctyjson.Marshal(outputValues.Value, outputValues.Value.Type())
		if err != nil {
			log.Printf("[DEBUG] @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@OMAR: EEROOOOOOOOORRRRRRRR %v", err)
			//return err
		}
		rawState := tfprotov5.RawState{
			JSON: mByte,
		}
		newVal, err := rawState.Unmarshal(tfType)
		if err != nil {
			log.Printf("[DEBUG] @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@OMAR: EEROOOOOOOOORRRRRRRR")
			//	return fmt.Errorf("ERROOR %v", err)
		}
		//newVal := tftypes.NewValue(tfType, outV)
		tftypesState[name] = newVal
		stateTypes[name] = tfType
	}

	state, err := tfprotov5.NewDynamicValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"workspace":    tftypes.String,
			"organization": tftypes.String,
			"values":       tftypes.DynamicPseudoType,
		},
	}, tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"workspace":    tftypes.String,
			"organization": tftypes.String,
			"values":       tftypes.Object{AttributeTypes: stateTypes},
		},
	}, map[string]tftypes.Value{
		"workspace":    tftypes.NewValue(tftypes.String, wsName),
		"organization": tftypes.NewValue(tftypes.String, orgName),
		"values":       tftypes.NewValue(tftypes.Object{AttributeTypes: stateTypes}, tftypesState),
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
			"values":       tftypes.String,
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

type remoteState struct {
	Outputs map[string]outputValue `json:"outputs"`
}

type outputValue struct {
	Type  interface{} `json:"type"`
	Value interface{} `json:"value"`
}

type stateV4 struct {
	RootOutputs map[string]outputStateV4 `json:"outputs"`
}

type outputStateV4 struct {
	ValueRaw     json.RawMessage `json:"value"`
	ValueTypeRaw json.RawMessage `json:"type"`
	Sensitive    bool            `json:"sensitive,omitempty"`
}

type OutputValue struct {
	Addr      AbsOutputValue
	Value     cty.Value
	Sensitive bool
}

type AbsOutputValue struct {
	OutputValue OutputValueTwo
}

type OutputValueTwo struct {
	Name string
}

type FinalOutputValue struct {
	OutputValues map[string]*OutputValue
}

func (d dataSourceRemoteState) readRemoteStateOutput(ctx context.Context, tfeClient *tfe.Client, orgName, wsName string) (*FinalOutputValue, error) {
	log.Printf("[DEBUG] Reading the Workspace %s in Organization %s", wsName, orgName)
	ws, err := tfeClient.Workspaces.Read(ctx, orgName, wsName)
	if err != nil {
		return nil, fmt.Errorf("Error reading workspace: %v", err)
	}

	/*

		TODO: CHeck this query
		   st, err := cli.StateVersions.CurrentWithOptions(ctx, ws.ID, &tfe.StateVersionCurrentOptions{
		              Include: "outputs",
		      })
	*/
	log.Printf("[DEBUG] Reading the current StateVersion for Workspace ID %s.", ws.ID)
	sv, err := tfeClient.StateVersions.Current(ctx, ws.ID)
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return nil, fmt.Errorf("Could not read  remote state for workspace '%s'", wsName)
		}
		return nil, fmt.Errorf("Error remote state: %v", err)
	}

	log.Printf("[DEBUG] Downloading State Version")
	stateData, err := tfeClient.StateVersions.Download(ctx, sv.DownloadURL)
	if err != nil {
		return nil, fmt.Errorf("Error downloading remote state: %v", err)
	}

	log.Printf("[DEBUG] Unmarshalling remote state output")
	read := bytes.NewReader(stateData)
	src, err := ioutil.ReadAll(read)
	if err != nil {
		return nil, fmt.Errorf("ERROOR %v", err)
	}
	sV4 := &stateV4{}
	err = json.Unmarshal(src, sV4)
	if err != nil {
		return nil, fmt.Errorf("ERROOR %v", err)
	}

	fov := &FinalOutputValue{
		OutputValues: map[string]*OutputValue{},
	}
	for name, fos := range sV4.RootOutputs {
		os := &OutputValue{
			Addr: AbsOutputValue{
				OutputValue: OutputValueTwo{
					Name: name,
				},
			},
		}
		os.Sensitive = fos.Sensitive

		ty, err := ctyjson.UnmarshalType([]byte(fos.ValueTypeRaw))
		if err != nil {
			return nil, fmt.Errorf("ERROOR %v", err)
		}

		val, err := ctyjson.Unmarshal([]byte(fos.ValueRaw), ty)
		if err != nil {
			return nil, fmt.Errorf("ERROOR %v", err)
		}

		os.Value = val
		fov.OutputValues[name] = os
	}

	//stateOuptput := &remoteState{}
	//if err := json.Unmarshal(stateData, stateOuptput); err != nil {
	//	return nil, fmt.Errorf("Could not unmarshal remote state output %v", err)
	//}

	return fov, nil
}

func parseRemoteStateOutput(stateOutput *FinalOutputValue) error {
	log.Printf("[DEBUG] @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@OMAR: parseRemoteStateOutput")
	log.Printf("[DEBUG] @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@OMAR: parseRemoteStateOutput")
	log.Printf("[DEBUG] @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@OMAR: parseRemoteStateOutput")

	tftypesState := map[string]tftypes.Value{}
	stateTypes := map[string]tftypes.Type{}

	for name, outputValues := range stateOutput.OutputValues {
		log.Printf("[DEBUG] @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@OMAR: NAME: ", name)
		log.Printf("[DEBUG] @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@OMAR: VALUE: ", outputValues.Value)
		log.Printf("[DEBUG] @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@OMAR: TYPE: ", outputValues.Value.Type())
		log.Printf("[DEBUG] @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@OMAR: KEY: ", outputValues)

		marshData, err := outputValues.Value.Type().MarshalJSON()
		if err != nil {
			log.Printf("[DEBUG] @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@OMAR: EEROOOOOOOOORRRRRRRR %v", err)
			//return err
		}
		tfType, err := tftypes.ParseJSONType(marshData)
		if err != nil {
			//return err
			log.Printf("[DEBUG] @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@OMAR: EEROOOOOOOOORRRRRRRR %v", err)
		}
		log.Printf("[DEBUG] @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@OMAR: TF TYPE: ", tfType)

		mByte, err := ctyjson.Marshal(outputValues.Value, outputValues.Value.Type())
		if err != nil {
			//return err
			log.Printf("[DEBUG] @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@OMAR: EEROOOOOOOOORRRRRRRR %v", err)
		}
		log.Printf("[DEBUG] @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@OMAR: BYTE: ", string(mByte))

		outV := []string{}
		//err = gocty.FromCtyValue(outputValues.Value, &outV)
		//err = json.Unmarshal(mByte, &outV)
		if err != nil {
			log.Printf("[DEBUG] @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@OMAR: EEROOOOOOOOORRRRRRRR %v", err)
			//return fmt.Errorf("ERROOR %v", err)
		}
		log.Printf("[DEBUG] @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@OMAR: OUTPUT VALUE: ", outV)

		log.Printf("[DEBUG] @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@OMAR: VALUE INTERFCE: ", tfType)

		rawState := tfprotov5.RawState{
			JSON: mByte,
		}
		newVal, err := rawState.Unmarshal(tfType)
		if err != nil {
			log.Printf("[DEBUG] @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@OMAR: EEROOOOOOOOORRRRRRRR")
			//return fmt.Errorf("ERROOR %v", err)
		}
		//newVal := tftypes.NewValue(tfType, outV)
		log.Printf("[DEBUG] @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@OMAR: NEW VAL: ", newVal)
		tftypesState[name] = newVal
		stateTypes[name] = tfType
	}
	return nil
}

/*
	stateFile := map[string]interface{}{
		"foo":   []interface{}{"a", "b", "c"},
		"hello": "world",
	}
*/
//	stateFile := map[string]interface{}{
//		"foo":   []interface{}{"a", "b", "c"},
//		"hello": int64(123),
//		"quuz":  false,
//	}
//for k, v := range stateFile {
//	switch val := v.(type) {
//	case string:
//		tftypesState[k] = tftypes.NewValue(tftypes.String, val)
//		stateTypes[k] = tftypes.String
//	case int64:
//		tftypesState[k] = tftypes.NewValue(tftypes.Number, big.NewFloat(float64(val)))
//		stateTypes[k] = tftypes.Number
//	case bool:
//		tftypesState[k] = tftypes.NewValue(tftypes.Bool, val)
//		stateTypes[k] = tftypes.Bool
//	case []interface{}:
//		elements := []tftypes.Value{}
//		types := []tftypes.Type{}
//		for _, element := range val {
//			switch el := element.(type) {
//			case string:
//				elements = append(elements, tftypes.NewValue(tftypes.String, el))
//				types = append(types, tftypes.String)
//			default:
//				panic(fmt.Sprintf("unknown type %T", element))
//			}
//		}
//		tftypesState[k] = tftypes.NewValue(tftypes.Tuple{ElementTypes: types}, elements)
//		stateTypes[k] = tftypes.Tuple{ElementTypes: types}
//	}
//}
