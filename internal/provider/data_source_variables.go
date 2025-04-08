// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource              = &dataSourceTFEVariables{}
	_ datasource.DataSourceWithConfigure = &dataSourceTFEVariables{}

	variableAttrTypes = map[string]attr.Type{
		"category":  types.StringType,
		"hcl":       types.BoolType,
		"id":        types.StringType,
		"name":      types.StringType,
		"sensitive": types.BoolType,
		"value":     types.StringType,
	}

	variableType = types.ObjectType{AttrTypes: variableAttrTypes}
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
) modelVariables {
	var model modelVariables

	// Set workspace or variable set ID
	if !workspaceID.IsNull() {
		model.ID = types.StringValue(fmt.Sprintf("variables/%s", workspaceID.ValueString()))
		model.WorkspaceID = workspaceID
	} else if !variableSetID.IsNull() {
		model.ID = types.StringValue(fmt.Sprintf("variables/%s", variableSetID.ValueString()))
		model.VariableSetID = variableSetID
	}

	model.Env = varListFromVariables(env)
	model.Terraform = varListFromVariables(terraform)
	model.Variables = varListFromVariables(variables)

	return model
}

func objectValueFromVariable(v tfe.Variable) types.Object {
	return types.ObjectValueMust(
		variableAttrTypes,
		map[string]attr.Value{
			"category":  types.StringValue(string(v.Category)),
			"hcl":       types.BoolValue(v.HCL),
			"id":        types.StringValue(v.ID),
			"name":      types.StringValue(v.Key),
			"sensitive": types.BoolValue(v.Sensitive),
			"value":     types.StringValue(v.Value),
		},
	)
}

func objectValueFromVariableSetVariable(v tfe.VariableSetVariable) types.Object {
	return types.ObjectValueMust(
		variableAttrTypes,
		map[string]attr.Value{
			"category":  types.StringValue(string(v.Category)),
			"hcl":       types.BoolValue(v.HCL),
			"id":        types.StringValue(v.ID),
			"name":      types.StringValue(v.Key),
			"sensitive": types.BoolValue(v.Sensitive),
			"value":     types.StringValue(v.Value),
		},
	)
}

func varListFromVariables(variables []any) types.List {
	varSlice := make([]attr.Value, 0, len(variables))

	var objVar types.Object
	for _, variable := range variables {
		switch v := variable.(type) {
		case *tfe.Variable:
			objVar = objectValueFromVariable(*v)

		case *tfe.VariableSetVariable:
			objVar = objectValueFromVariableSetVariable(*v)

		default: // should not happen
			panic(fmt.Sprintf("unexpected type %T reading variable", variable))
		}

		varSlice = append(varSlice, objVar)
	}

	varList := types.ListValueMust(variableType, varSlice)

	return varList
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

			"env": schema.ListAttribute{
				Computed:    true,
				ElementType: variableType,
			},

			"terraform": schema.ListAttribute{
				Computed:    true,
				ElementType: variableType,
			},

			"variables": schema.ListAttribute{
				Computed:    true,
				ElementType: variableType,
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

	model := modelFromVariables(config.WorkspaceID, config.VariableSetID, env, terraform, variables)

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

	model := modelFromVariables(config.WorkspaceID, config.VariableSetID, env, terraform, variables)

	// Update state
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
