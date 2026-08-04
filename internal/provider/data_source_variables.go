// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-tfe/v2/api/models"
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

// objectValueFromV2Var builds a types.Object from a go-tfe v2 Varsable.
func objectValueFromV2Var(v models.Varsable) types.Object {
	id := valueOrZero(v.GetId())
	key, value, category, hcl, sensitive := v2VarAttrFields(v.GetAttributes())

	return types.ObjectValueMust(
		variableAttrTypes,
		map[string]attr.Value{
			"category":  types.StringValue(category),
			"hcl":       types.BoolValue(hcl),
			"id":        types.StringValue(id),
			"name":      types.StringValue(key),
			"sensitive": types.BoolValue(sensitive),
			"value":     types.StringValue(value),
		},
	)
}

// v2VarAttrFields extracts the scalar fields from a v2 Vars_attributesable,
// returning zero values when attrs or any pointer is nil.
func v2VarAttrFields(attrs models.Vars_attributesable) (key, value, category string, hcl, sensitive bool) {
	if attrs == nil {
		return
	}
	key = valueOrZero(attrs.GetKey())
	value = valueOrZero(attrs.GetValue())
	if attrs.GetCategory() != nil {
		category = attrs.GetCategory().String()
	}
	hcl = valueOrZero(attrs.GetHcl())
	sensitive = valueOrZero(attrs.GetSensitive())
	return
}

// modelFromVariables builds a modelVariables struct.
func modelFromVariables(
	workspaceID types.String,
	variableSetID types.String,
	env []types.Object,
	terraform []types.Object,
	variables []types.Object,
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

	model.Env = varListFromObjects(env)
	model.Terraform = varListFromObjects(terraform)
	model.Variables = varListFromObjects(variables)

	return model
}

func varListFromObjects(variables []types.Object) types.List {
	varSlice := make([]attr.Value, len(variables))
	for i, v := range variables {
		varSlice[i] = v
	}
	return types.ListValueMust(variableType, varSlice)
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
		Description: "Gets all variables defined in a workspace or variable set." +
			"\n\n-> **Note:** One of `workspace_id` or `variable_set_id` is required.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Static identifier for the variables group in the workspace.",
				Computed:    true,
			},

			"workspace_id": schema.StringAttribute{
				Description: "ID of the workspace.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("variable_set_id")),
				},
			},

			"variable_set_id": schema.StringAttribute{
				MarkdownDescription: "ID of the variable set.",
				Optional:            true,
			},

			"env": schema.ListNestedAttribute{
				Description: "List containing environment variables configured on the workspace.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The variable ID.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The variable Key name.",
							Computed:    true,
						},
						"value": schema.StringAttribute{
							Description: "The variable value. If the variable is sensitive this value will be empty.",
							Computed:    true,
						},
						"category": schema.StringAttribute{
							MarkdownDescription: "The category of the variable. Valid values are `terraform` or `env`.",
							Computed:            true,
						},
						"hcl": schema.BoolAttribute{
							Description: "Whether the variable is marked as HCL or not.",
							Computed:    true,
						},
						"sensitive": schema.BoolAttribute{
							Description: "Whether the variable's value is sensitive and hidden.",
							Computed:    true,
						},
					},
				},
			},

			"terraform": schema.ListNestedAttribute{
				Description: "List containing terraform variables configured on the workspace.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The variable ID.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The variable Key name.",
							Computed:    true,
						},
						"value": schema.StringAttribute{
							Description: "The variable value. If the variable is sensitive this value will be empty.",
							Computed:    true,
						},
						"category": schema.StringAttribute{
							MarkdownDescription: "The category of the variable. Valid values are `terraform` or `env`.",
							Computed:            true,
						},
						"hcl": schema.BoolAttribute{
							Description: "Whether the variable is marked as HCL or not.",
							Computed:    true,
						},
						"sensitive": schema.BoolAttribute{
							Description: "Whether the variable's value is sensitive and hidden.",
							Computed:    true,
						},
					},
				},
			},

			"variables": schema.ListNestedAttribute{
				Description: "List containing terraform and environment variables configured on the workspace.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The variable ID.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The variable Key name.",
							Computed:    true,
						},
						"value": schema.StringAttribute{
							Description: "The variable value. If the variable is sensitive this value will be empty.",
							Computed:    true,
						},
						"category": schema.StringAttribute{
							MarkdownDescription: "The category of the variable. Valid values are `terraform` or `env`.",
							Computed:            true,
						},
						"hcl": schema.BoolAttribute{
							Description: "Whether the variable is marked as HCL or not.",
							Computed:    true,
						},
						"sensitive": schema.BoolAttribute{
							Description: "Whether the variable's value is sensitive and hidden.",
							Computed:    true,
						},
					},
				},
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
	workspaceID := config.WorkspaceID.ValueString()
	api := d.config.ClientV2.API

	tflog.Debug(ctx, fmt.Sprintf("Reading workspace: %s", workspaceID))

	// Workspace vars are returned all at once (no pagination in the spec).
	variableList, err := api.Workspaces().ByWorkspace_id(workspaceID).Vars().Get(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Error retrieving variables for workspace %s:", workspaceID), err.Error())
		return
	}

	var env []types.Object
	var terraform []types.Object
	var variables []types.Object

	for _, variable := range variableList.GetData() {
		obj := objectValueFromV2Var(variable)
		variables = append(variables, obj)

		category := ""
		if attrs := variable.GetAttributes(); attrs != nil && attrs.GetCategory() != nil {
			category = attrs.GetCategory().String()
		}
		switch category {
		case "env":
			env = append(env, obj)
		case "terraform":
			terraform = append(terraform, obj)
		}
	}

	model := modelFromVariables(config.WorkspaceID, config.VariableSetID, env, terraform, variables)

	// Update state
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (d *dataSourceTFEVariables) readFromVariableSet(ctx context.Context, config modelVariables, resp *datasource.ReadResponse) {
	variableSetID := config.VariableSetID.ValueString()
	api := d.config.ClientV2.API

	tflog.Debug(ctx, fmt.Sprintf("Reading variable set: %s", variableSetID))

	var env []types.Object
	var terraform []types.Object
	var variables []types.Object

	// Follow link-based pagination for varset variables.
	varsBuilder := api.Varsets().ByVarset_id(variableSetID).Relationships().Vars()
	for {
		variableList, err := varsBuilder.Get(ctx, nil)
		if err != nil {
			if variableList == nil {
				resp.Diagnostics.AddError(fmt.Sprintf("Error retrieving variables for variable set %s:", variableSetID), err.Error())
				return
			}
			resp.Diagnostics.AddError(fmt.Sprintf("Error retrieving variables for variable set %s:", variableSetID), err.Error())
			return
		}
		if variableList == nil {
			break
		}

		for _, variable := range variableList.GetData() {
			obj := objectValueFromV2Var(variable)
			variables = append(variables, obj)

			category := ""
			if attrs := variable.GetAttributes(); attrs != nil && attrs.GetCategory() != nil {
				category = attrs.GetCategory().String()
			}
			switch category {
			case "env":
				env = append(env, obj)
			case "terraform":
				terraform = append(terraform, obj)
			}
		}

		// Follow the next page link if present.
		links := variableList.GetLinks()
		if links == nil || links.GetNext() == nil || *links.GetNext() == "" {
			break
		}
		varsBuilder = varsBuilder.WithUrl(*links.GetNext())
	}

	model := modelFromVariables(config.WorkspaceID, config.VariableSetID, env, terraform, variables)

	// Update state
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
