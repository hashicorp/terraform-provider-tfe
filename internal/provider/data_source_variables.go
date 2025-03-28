// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strconv"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource              = &dataSourceTFEVariables{}
	_ datasource.DataSourceWithConfigure = &dataSourceTFEVariables{}
)

// NewVariablesDataSource is a helper function to simplify the provider
// implementation.
func NewVariablesDataSource() datasource.DataSource {
	return &dataSourceTFEVariables{}
}

// dataSourceTFEVariables is the data source implementation.
type dataSourceTFEVariables struct {
	config ConfiguredClient
}

// modelFromVariables builds a modelVariables struct from a tfe.Variable value.
func modelFromVariables(
	workspaceID types.String,
	variableSetID types.String,
	env []any,
	terraform []any,
	variables []any,
) (modelVariables, diag.Diagnostics) {
	var diags diag.Diagnostics
	var model modelVariables

	// Set workspace or variable set ID
	if !workspaceID.IsNull() {
		model.ID = types.StringValue(fmt.Sprintf("variables/%s", workspaceID.ValueString()))
		model.WorkspaceID = workspaceID
	} else if !variableSetID.IsNull() {
		model.ID = types.StringValue(fmt.Sprintf("variables/%s", variableSetID.ValueString()))
		model.VariableSetID = variableSetID
	}

	// Set the environment variables
	envList, diags := varListFromVariables(ctx, env)
	diags.Append(diags...)
	model.Env = envList

	// Set the terraform variables
	terraformList, diags := varListFromVariables(ctx, terraform)
	diags.Append(diags...)
	model.Terraform = terraformList

	// Set the variables
	variablesList, diags := varListFromVariables(ctx, variables)
	diags.Append(diags...)
	model.Variables = variablesList

	return model, diags
}

func varListFromVariables(ctx context.Context, variables []any) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	var varSlice []any

	variableFieldCount := 6
	mapType := types.MapType{ElemType: types.StringType}
	varList := types.ListNull(mapType)

	for _, variable := range variables {
		rawMap := make(map[string]string, variableFieldCount)
		switch v := variable.(type) {
		case *tfe.Variable:
			rawMap["id"] = v.ID
			rawMap["name"] = v.Key
			rawMap["value"] = v.Value
			rawMap["category"] = string(v.Category)
			rawMap["sensitive"] = strconv.FormatBool(v.Sensitive)
			rawMap["hcl"] = strconv.FormatBool(v.HCL)

		case *tfe.VariableSetVariable:
			rawMap["id"] = v.ID
			rawMap["name"] = v.Key
			rawMap["value"] = v.Value
			rawMap["category"] = string(v.Category)
			rawMap["sensitive"] = strconv.FormatBool(v.Sensitive)
			rawMap["hcl"] = strconv.FormatBool(v.HCL)

		default: // should not happen
			diags.AddError("Error reading variable", fmt.Sprintf("unexpected type %T", variable))
			return varList, diags
		}

		varMap, d := types.MapValueFrom(ctx, types.StringType, rawMap)
		if d.HasError() {
			diags.Append(d...)
			return varList, diags
		}

		varSlice = append(varSlice, varMap)
	}

	varList, d := types.ListValueFrom(ctx, mapType, varSlice)
	if d.HasError() {
		diags.Append(d...)
		return varList, diags
	}

	return varList, diags
}

// modelVariables maps the overall data source schema data.
type modelVariables struct {
	ID            types.String `tfsdk:"id"`
	WorkspaceID   types.String `tfsdk:"workspace_id"`
	VariableSetID types.String `tfsdk:"variable_set_id"`
	Env           types.List   `tfsdk:"env"`
	Terraform     types.List   `tfsdk:"terraform"`
	Variables     types.List   `tfsdk:"variables"`
}

// Configure implements datasource.DataSourceWithConfigure
func (d *dataSourceTFEVariables) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(ConfiguredClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected tfe.ConfiguredClient, got %T. This is a bug in the tfe provider, so please report it on GitHub.", req.ProviderData),
		)

		return
	}
	d.config = client
}

// Metadata implements datasource.DataSourceWithMetadata.
func (d *dataSourceTFEVariables) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_variables"
}

func (d *dataSourceTFEVariables) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This data source can be used to retrieve all variables in a workspace or variable set.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},

			"workspace_id": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("variable_set_id")),
				},
			},

			"variable_set_id": schema.StringAttribute{
				Optional: true,
			},

			// TODO(v2): Update the attribute type for the three following attributes
			// to be ListNestedAttribute.
			//
			// This will change the usage pattern from:
			//   data.tfe_variables.workspace_foobar.env[0]["key"]
			//
			// to:
			//   data.tfe_variables.workspace_foobar.env[0].key

			"env": schema.ListAttribute{
				Computed:    true,
				ElementType: types.MapType{ElemType: types.StringType},
			},

			"terraform": schema.ListAttribute{
				Computed:    true,
				ElementType: types.MapType{ElemType: types.StringType},
			},

			"variables": schema.ListAttribute{
				Computed:    true,
				ElementType: types.MapType{ElemType: types.StringType},
			},
		},
	}
}

// Read implements datasource.DataSource.
func (d *dataSourceTFEVariables) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Load the config into the model.
	var config modelVariables
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !config.WorkspaceID.IsNull() {
		d.readFromWorkspace(ctx, config, resp)
	} else if !config.VariableSetID.IsNull() {
		d.readFromVariableSet(ctx, config, resp)
	}
}

func (d *dataSourceTFEVariables) readFromWorkspace(ctx context.Context, config modelVariables, resp *datasource.ReadResponse) {
	var (
		options *tfe.VariableListOptions

		env       []any
		terraform []any
		variables []any
	)

	workspaceID := config.WorkspaceID.ValueString()

	tflog.Debug(ctx, fmt.Sprintf("Reading workspace: %s", workspaceID))

	for {
		variableList, err := d.config.Client.Variables.List(ctx, workspaceID, options)
		if err != nil {
			resp.Diagnostics.AddError(fmt.Sprintf("Error retrieving variables for workspace %s:", workspaceID), err.Error())
			return
		}

		for _, variable := range variableList.Items {
			variables = append(variables, variable)

			switch variable.Category {
			case "env":
				env = append(env, variable)
			case "terraform":
				terraform = append(terraform, variable)
			}
		}

		// Exit the loop when we've seen all pages.
		if variableList.CurrentPage >= variableList.TotalPages {
			break
		}

		// Update the page number to get the next page.
		options.PageNumber = variableList.NextPage
	}

	model, diags := modelFromVariables(config.WorkspaceID, config.VariableSetID, env, terraform, variables)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Update state
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (d *dataSourceTFEVariables) readFromVariableSet(ctx context.Context, config modelVariables, resp *datasource.ReadResponse) {
	var (
		options *tfe.VariableSetVariableListOptions

		env       []any
		terraform []any
		variables []any
	)

	variableSetID := config.VariableSetID.ValueString()

	tflog.Debug(ctx, fmt.Sprintf("Reading variable set: %s", variableSetID))

	for {
		variableList, err := d.config.Client.VariableSetVariables.List(ctx, variableSetID, options)
		if err != nil || variableList == nil {
			resp.Diagnostics.AddError(fmt.Sprintf("Error retrieving variables for variable set %s:", variableSetID), err.Error())
			return
		}

		for _, variable := range variableList.Items {
			variables = append(variables, variable)

			switch variable.Category {
			case "env":
				env = append(env, variable)
			case "terraform":
				terraform = append(terraform, variable)
			}
		}

		// Exit the loop when we've seen all pages.
		if variableList.CurrentPage >= variableList.TotalPages {
			break
		}

		// Update the page number to get the next page.
		options.PageNumber = variableList.NextPage
	}

	model, diags := modelFromVariables(config.WorkspaceID, config.VariableSetID, env, terraform, variables)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Update state
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
