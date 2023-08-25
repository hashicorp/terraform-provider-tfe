// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/zclconf/go-cty/cty"
	ctyjson "github.com/zclconf/go-cty/cty/json"
)

type dataSourceOutputs struct {
	tfeClient    *tfe.Client
	organization string
}

func newDataSourceOutputs(config ConfiguredClient) tfprotov5.DataSourceServer {
	return dataSourceOutputs{
		tfeClient:    config.Client,
		organization: config.Organization,
	}
}

func (d dataSourceOutputs) ReadDataSource(ctx context.Context, req *tfprotov5.ReadDataSourceRequest) (*tfprotov5.ReadDataSourceResponse, error) {
	resp := &tfprotov5.ReadDataSourceResponse{
		Diagnostics: []*tfprotov5.Diagnostic{},
	}

	orgName, wsName, err := d.readConfigValues(req)
	if err != nil {
		resp.Diagnostics = append(resp.Diagnostics, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  "Error retrieving values from the config",
			Detail:   fmt.Sprintf("Error retrieving values from the config: %v", err),
		})
		return resp, nil
	}

	remoteStateOutput, err := d.readStateOutput(ctx, orgName, wsName)
	if err != nil {
		resp.Diagnostics = append(resp.Diagnostics, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  "Error reading remote state output",
			Detail:   fmt.Sprintf("Error reading remote state output: %v", err),
		})
		return resp, nil
	}

	tftypesValues, stateTypes, tftypesNonsensitiveValues, nonsensitiveStateTypes, err := parseStateOutput(remoteStateOutput)
	if err != nil {
		resp.Diagnostics = append(resp.Diagnostics, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  "Error parsing remote state output",
			Detail:   fmt.Sprintf("Error parsing remote state output: %v", err),
		})
		return resp, nil
	}

	id := fmt.Sprintf("%s-%s", orgName, wsName)
	state, err := tfprotov5.NewDynamicValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"workspace":           tftypes.String,
			"organization":        tftypes.String,
			"values":              tftypes.DynamicPseudoType,
			"nonsensitive_values": tftypes.DynamicPseudoType,
			"id":                  tftypes.String,
		},
	}, tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"workspace":           tftypes.String,
			"organization":        tftypes.String,
			"values":              tftypes.Object{AttributeTypes: stateTypes},
			"nonsensitive_values": tftypes.Object{AttributeTypes: nonsensitiveStateTypes},
			"id":                  tftypes.String,
		},
	}, map[string]tftypes.Value{
		"workspace":           tftypes.NewValue(tftypes.String, wsName),
		"organization":        tftypes.NewValue(tftypes.String, orgName),
		"values":              tftypes.NewValue(tftypes.Object{AttributeTypes: stateTypes}, tftypesValues),
		"nonsensitive_values": tftypes.NewValue(tftypes.Object{AttributeTypes: nonsensitiveStateTypes}, tftypesNonsensitiveValues),
		"id":                  tftypes.NewValue(tftypes.String, id),
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

func (d dataSourceOutputs) ValidateDataSourceConfig(ctx context.Context, req *tfprotov5.ValidateDataSourceConfigRequest) (*tfprotov5.ValidateDataSourceConfigResponse, error) {
	return &tfprotov5.ValidateDataSourceConfigResponse{}, nil
}

func (d dataSourceOutputs) readConfigValues(req *tfprotov5.ReadDataSourceRequest) (string, string, error) {
	var orgName string
	var wsName string
	var err error

	config := req.Config
	val, err := config.Unmarshal(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"workspace":           tftypes.String,
			"organization":        tftypes.String,
			"values":              tftypes.DynamicPseudoType,
			"nonsensitive_values": tftypes.DynamicPseudoType,
			"id":                  tftypes.String,
		}})
	if err != nil {
		return "", "", fmt.Errorf("Error unmarshalling config: %w", err)
	}

	var valMap map[string]tftypes.Value
	err = val.As(&valMap)
	if err != nil {
		return "", "", fmt.Errorf("error assigning configuration attributes to map: %w", err)
	}

	err = valMap["organization"].As(&orgName)
	if err != nil || orgName == "" {
		if d.organization == "" {
			return "", "", errMissingOrganization
		}
		orgName = d.organization
	}

	if valMap["workspace"].IsNull() {
		return orgName, "", fmt.Errorf("workspace cannot be nil: %w", err)
	}

	err = valMap["workspace"].As(&wsName)
	if err != nil {
		return orgName, wsName, fmt.Errorf("error assigning 'workspace' value to string: %w", err)
	}

	return orgName, wsName, nil
}

type stateData struct {
	outputs map[string]*outputData
}

type outputData struct {
	Value     cty.Value
	Sensitive cty.Value
}

func (d dataSourceOutputs) readStateOutput(ctx context.Context, orgName, wsName string) (*stateData, error) {
	log.Printf("[DEBUG] Reading the Workspace %s in Organization %s", wsName, orgName)
	opts := &tfe.WorkspaceReadOptions{
		Include: []tfe.WSIncludeOpt{tfe.WSOutputs},
	}
	ws, err := d.tfeClient.Workspaces.ReadWithOptions(ctx, orgName, wsName, opts)
	if err != nil {
		return nil, fmt.Errorf("error reading workspace: %w", err)
	}

	sd := &stateData{
		outputs: map[string]*outputData{},
	}

	for _, op := range ws.Outputs {
		if op.Sensitive {
			sensitiveOutput, err := d.tfeClient.StateVersionOutputs.Read(ctx, op.ID)
			if err != nil {
				return nil, fmt.Errorf("could not read sensitive output: %w", err)
			}
			op.Value = sensitiveOutput.Value
		}

		buf, err := json.Marshal(op.Value)
		if err != nil {
			return nil, fmt.Errorf("could not marshal output value: %w", err)
		}

		v := ctyjson.SimpleJSONValue{}
		err = v.UnmarshalJSON(buf)
		if err != nil {
			return nil, fmt.Errorf("could not unmarshal output value: %w", err)
		}
		sd.outputs[op.Name] = &outputData{
			Value:     v.Value,
			Sensitive: cty.BoolVal(op.Sensitive),
		}
	}

	return sd, nil
}

func parseStateOutput(stateOutput *stateData) (map[string]tftypes.Value, map[string]tftypes.Type, map[string]tftypes.Value, map[string]tftypes.Type, error) {
	tftypesValues := map[string]tftypes.Value{}
	stateTypes := map[string]tftypes.Type{}

	tftypesNonsensitiveValues := map[string]tftypes.Value{}
	nonsensitiveStateTypes := map[string]tftypes.Type{}

	for name, output := range stateOutput.outputs {
		marshData, err := output.Value.Type().MarshalJSON()
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("could not marshal output type: %w", err)
		}
		tfType, err := tftypes.ParseJSONType(marshData)
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("could not parse json type data: %w", err)
		}
		mByte, err := ctyjson.Marshal(output.Value, output.Value.Type())
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("could not marshal output value and output type: %w", err)
		}
		tfRawState := tfprotov5.RawState{
			JSON: mByte,
		}
		newVal, err := tfRawState.Unmarshal(tfType)
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("could not unmarshal tftype into value: %w", err)
		}
		if output.Sensitive.False() {
			tftypesNonsensitiveValues[name] = newVal
			nonsensitiveStateTypes[name] = tfType
		}

		tftypesValues[name] = newVal
		stateTypes[name] = tfType
	}

	return tftypesValues, stateTypes, tftypesNonsensitiveValues, nonsensitiveStateTypes, nil
}
