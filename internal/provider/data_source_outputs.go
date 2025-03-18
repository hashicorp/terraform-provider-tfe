// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"reflect"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &outputsDataSource{}
	_ datasource.DataSourceWithConfigure = &outputsDataSource{}
)

func NewOutputsDataSource() datasource.DataSource {
	return &outputsDataSource{}
}

type outputsDataSource struct {
	config ConfiguredClient
}

type outputsModel struct {
	ID                 types.String  `tfsdk:"id"`
	Organization       types.String  `tfsdk:"organization"`
	Workspace          types.String  `tfsdk:"workspace"`
	Values             types.Dynamic `tfsdk:"values"`
	NonSensitiveValues types.Dynamic `tfsdk:"nonsensitive_values"`
}

func modelFromOutputs(v *tfe.Workspace, sensitiveOutputs types.Dynamic, nonSensitiveOutputs types.Dynamic) outputsModel {
	orgName := v.Organization.Name
	wsName := v.Name

	return outputsModel{
		ID:                 types.StringValue(fmt.Sprintf("%s-%s", orgName, wsName)),
		Organization:       types.StringValue(orgName),
		Workspace:          types.StringValue(wsName),
		Values:             sensitiveOutputs,
		NonSensitiveValues: nonSensitiveOutputs,
	}
}

func (d *outputsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This data source can be used to retrieve a workspace's state outputs.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: `System-generated unique identifier for the resource.`,
				Computed:    true,
			},
			"organization": schema.StringAttribute{
				Description: `Name of the organization.`,
				Optional:    true,
			},
			"workspace": schema.StringAttribute{
				Description: `Name of the workspace.`,
				Required:    true,
			},
			"values": schema.DynamicAttribute{
				Description: `Values of the workspace outputs.`,
				Computed:    true,
				Sensitive:   true,
			},
			"nonsensitive_values": schema.DynamicAttribute{
				Description: `Non-sensitive values of the workspace outputs.`,
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *outputsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(ConfiguredClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Ephemeral Resource Configure Type",
			fmt.Sprintf("Expected tfe.ConfiguredClient, got %T. This is a bug in the tfe provider, so please report it on GitHub.", req.ProviderData),
		)

		return
	}

	d.config = client
}

func (d *outputsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_outputs"
}

func (d *outputsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	config := outputsModel{}
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get org name or default
	var orgName string
	resp.Diagnostics.Append(d.config.dataOrDefaultOrganization(ctx, req.Config, &orgName)...)
	if resp.Diagnostics.HasError() {
		return
	}

	log.Printf("[DEBUG] Reading the workspace %s in organization %s", config.Workspace.ValueString(), orgName)
	opts := &tfe.WorkspaceReadOptions{
		Include: []tfe.WSIncludeOpt{tfe.WSOutputs},
	}

	ws, err := d.config.Client.Workspaces.ReadWithOptions(ctx, orgName, config.Workspace.ValueString(), opts)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read workspace", err.Error())
		return
	}

	sensitiveTypes := map[string]attr.Type{}
	sensitiveValues := map[string]attr.Value{}
	nonSensitiveTypes := map[string]attr.Type{}
	nonSensitiveValues := map[string]attr.Value{}

	for _, op := range ws.Outputs {
		if op.Sensitive {
			// An additional API call is required to read sensitive output values.
			result, err := d.config.Client.StateVersionOutputs.Read(ctx, op.ID)
			if err != nil {
				resp.Diagnostics.AddError("Unable to read resource", err.Error())
				return
			}

			op.Value = result.Value
		}

		attrType, err := inferAttrType(op.Value)
		if err != nil {
			resp.Diagnostics.AddError("Error inferring attribute type", err.Error())
			return
		}

		attrValue, diags := convertToAttrValue(op.Value, attrType)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		sensitiveTypes[op.Name] = attrType
		sensitiveValues[op.Name] = attrValue

		if !op.Sensitive {
			nonSensitiveTypes[op.Name] = attrType
			nonSensitiveValues[op.Name] = attrValue
		}
	}

	// Create dynamic attribute value for `sensitive_values`
	obj, diags := types.ObjectValue(sensitiveTypes, sensitiveValues)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	sensitiveOutputs := types.DynamicValue(obj)

	// Create dynamic attribute value for `nonsensitive_values`
	obj, diags = types.ObjectValue(nonSensitiveTypes, nonSensitiveValues)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	nonSensitiveOutputs := types.DynamicValue(obj)

	diags.Append(resp.State.Set(ctx, modelFromOutputs(ws, sensitiveOutputs, nonSensitiveOutputs))...)
}

func inferAttrType(raw interface{}) (attr.Type, error) {
	switch v := raw.(type) {
	case bool:
		return types.BoolType, nil
	case int, int8, int16, int32, int64, float32, float64:
		return types.NumberType, nil
	case string:
		return types.StringType, nil
	case []interface{}:
		// For slices, if the slice is empty return a List with a dynamic element type.
		if len(v) == 0 {
			return types.ListType{ElemType: types.DynamicType}, nil
		}

		// Infer the type for the first element.
		firstType, err := inferAttrType(v[0])
		if err != nil {
			return nil, err
		}

		// Check if all elements are of the same type.
		homogeneous := true
		for i := 1; i < len(v); i++ {
			currType, err := inferAttrType(v[i])
			if err != nil {
				return nil, err
			}
			if !reflect.DeepEqual(firstType, currType) {
				homogeneous = false
				break
			}
		}
		if homogeneous {
			return types.ListType{ElemType: firstType}, nil
		}

		// If not homogeneous, build a Tuple with each element’s inferred type.
		tupleTypes := make([]attr.Type, len(v))
		for i, elem := range v {
			t, err := inferAttrType(elem)
			if err != nil {
				return nil, err
			}
			tupleTypes[i] = t
		}
		return types.TupleType{ElemTypes: tupleTypes}, nil
	case map[string]interface{}:
		// Build an Object type by inferring each attribute's type.
		attrTypes := make(map[string]attr.Type)
		for key, value := range v {
			inferred, err := inferAttrType(value)
			if err != nil {
				return nil, fmt.Errorf("error inferring type for key %q: %w", key, err)
			}
			attrTypes[key] = inferred
		}
		return types.ObjectType{AttrTypes: attrTypes}, nil
	default:
		return nil, fmt.Errorf("unsupported type %T", raw)
	}
}

func convertToAttrValue(raw interface{}, t attr.Type) (attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics

	if t == types.BoolType {
		b, ok := raw.(bool)
		if !ok {
			diags.AddError("Conversion Error", "expected bool")
			return types.BoolNull(), diags
		}
		return types.BoolValue(b), diags
	}

	if t == types.NumberType {
		// Use a float64 conversion to handle all numeric types.
		n, ok := raw.(float64)
		if !ok {
			diags.AddError("Conversion Error", "expected number")
			return types.NumberNull(), diags
		}
		return types.NumberValue(big.NewFloat(n)), diags
	}

	if t == types.StringType {
		s, ok := raw.(string)
		if !ok {
			diags.AddError("Conversion Error", "expected string")
			return types.StringNull(), diags
		}
		return types.StringValue(s), diags
	}

	// For composite types, use a type switch on the expected type.
	switch tt := t.(type) {
	case types.ListType:
		// Expect raw to be a slice.
		slice, ok := raw.([]interface{})
		if !ok {
			diags.AddError("Conversion Error", "expected slice for ListType")
			return types.ListNull(tt.ElemType), diags
		}

		var elems []attr.Value
		for _, elem := range slice {
			v, ds := convertToAttrValue(elem, tt.ElemType)
			diags.Append(ds...)
			elems = append(elems, v)
		}
		return types.ListValue(tt.ElemType, elems)

	case types.TupleType:
		// Expect raw to be a slice.
		slice, ok := raw.([]interface{})
		if !ok {
			diags.AddError("Conversion Error", "expected slice for TupleType")
			return types.TupleNull(tt.ElemTypes), diags
		}
		if len(slice) != len(tt.ElemTypes) {
			diags.AddError("Conversion Error", "tuple length mismatch")
			return types.TupleNull(tt.ElemTypes), diags
		}

		var elems []attr.Value
		for i, elem := range slice {
			v, ds := convertToAttrValue(elem, tt.ElemTypes[i])
			diags.Append(ds...)
			elems = append(elems, v)
		}
		return types.TupleValue(tt.ElemTypes, elems)

	case types.ObjectType:
		// Expect raw to be a map[string]interface{}.
		m, ok := raw.(map[string]interface{})
		if !ok {
			diags.AddError("Conversion Error", "expected map for ObjectType")
			return types.ObjectNull(tt.AttrTypes), diags
		}

		objValues := make(map[string]attr.Value)
		// Iterate over the expected attributes defined in the ObjectType.
		for key, expectedType := range tt.AttrTypes {
			value := m[key]
			v, ds := convertToAttrValue(value, expectedType)
			diags.Append(ds...)
			if ds.HasError() {
				return types.ObjectNull(tt.AttrTypes), diags
			}

			objValues[key] = v
		}
		return types.ObjectValue(tt.AttrTypes, objValues)

	case types.MapType:
		// Expect raw to be a map[string]interface{}.
		m, ok := raw.(map[string]interface{})
		if !ok {
			diags.AddError("Conversion Error", "expected map for MapType")
			return types.MapValue(tt.ElemType, nil)
		}

		mapValues := make(map[string]attr.Value)
		for key, value := range m {
			v, ds := convertToAttrValue(value, tt.ElemType)
			diags.Append(ds...)
			if ds.HasError() {
				return types.MapValue(tt.ElemType, nil)
			}

			mapValues[key] = v
		}
		return types.MapValue(tt.ElemType, mapValues)

	default:
		diags.AddError("Conversion Error", fmt.Sprintf("unsupported type %T", t))
	}

	return nil, diags
}
