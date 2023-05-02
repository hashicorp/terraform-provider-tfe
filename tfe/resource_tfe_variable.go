package tfe

import (
	"context"
	"fmt"
	"log"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type resourceTFEVariable struct {
	config ConfiguredClient
}

// modelTFEVariable maps the resource schema data to a struct.
type modelTFEVariable struct {
	ID            types.String `tfsdk:"id"`
	Key           types.String `tfsdk:"key"`
	Value         types.String `tfsdk:"value"`
	Category      types.String `tfsdk:"category"`
	Description   types.String `tfsdk:"description"`
	HCL           types.Bool   `tfsdk:"hcl"`
	Sensitive     types.Bool   `tfsdk:"sensitive"`
	WorkspaceID   types.String `tfsdk:"workspace_id"`
	VariableSetID types.String `tfsdk:"variable_set_id"`
}

func (m *modelTFEVariable) refreshFromTFEVariable(v tfe.Variable) {
	// For most fields, the server is authoritative:
	m.ID = types.StringValue(v.ID)
	m.Key = types.StringValue(v.Key)
	m.Category = types.StringValue(string(v.Category))
	m.Description = types.StringValue(v.Description) // can be null in API, but becomes zero value in tfe.Variable.
	m.HCL = types.BoolValue(v.HCL)
	m.Sensitive = types.BoolValue(v.Sensitive)
	m.WorkspaceID = types.StringValue(v.Workspace.ID)
	m.VariableSetID = types.StringNull() // never present on workspace vars.

	// But: if the variable is sensitive, our client always gets an empty Value,
	// so our last-known info is the best we're gonna get.
	if !v.Sensitive {
		m.Value = types.StringValue(v.Value)
	}
}

func (m *modelTFEVariable) refreshFromTFEVariableSetVariable(v tfe.VariableSetVariable) {
	// For most fields, the server is authoritative:
	m.ID = types.StringValue(v.ID)
	m.Key = types.StringValue(v.Key)
	m.Category = types.StringValue(string(v.Category))
	m.Description = types.StringValue(v.Description) // can be null in API, but becomes zero value in tfe.VariableSetVariable.
	m.HCL = types.BoolValue(v.HCL)
	m.Sensitive = types.BoolValue(v.Sensitive)
	m.WorkspaceID = types.StringNull() // never present on variable set vars.
	m.VariableSetID = types.StringValue(v.VariableSet.ID)

	// But: if the variable is sensitive, our client always gets an empty Value,
	// so our last-known info is the best we're gonna get.
	if !v.Sensitive {
		m.Value = types.StringValue(v.Value)
	}
}

// Configure implements resource.ResourceWithConfigure. TODO: dry this out for other rscs
func (r *resourceTFEVariable) Configure(ctx context.Context, req resource.ConfigureRequest, res *resource.ConfigureResponse) {
	// Early exit if provider is unconfigured (i.e. we're only validating config or something)
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(ConfiguredClient)
	if !ok {
		res.Diagnostics.AddError(
			"Unexpected resource Configure type",
			fmt.Sprintf("Expected tfe.ConfiguredClient, got %T. This is a bug in the tfe provider, so please report it on GitHub.", req.ProviderData),
		)
	}
	r.config = client
}

// Metadata implements resource.Resource
func (r *resourceTFEVariable) Metadata(_ context.Context, _ resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = "tfe_variable"
}

// Schema implements resource.Resource
func (r *resourceTFEVariable) Schema(ctx context.Context, req resource.SchemaRequest, res *resource.SchemaResponse) {
	res.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Service-generated identifier for the variable",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"key": schema.StringAttribute{
				Required:    true,
				Description: "Name of the variable.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIf(
						func(ctx context.Context, req planmodifier.StringRequest, res *stringplanmodifier.RequiresReplaceIfFuncResponse) {
							var stateSensitive types.Bool
							diags := req.State.GetAttribute(ctx, path.Root("sensitive"), &stateSensitive)
							if diags.HasError() {
								res.Diagnostics.Append(diags...)
								return
							}
							if stateSensitive.ValueBool() && req.PlanValue.ValueString() != req.StateValue.ValueString() {
								res.RequiresReplace = true
							}
						},
						"Force replacement if key changed and sensitive is true",
						"Force replacement if key changed and sensitive is true",
					),
				},
			},
			"value": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Sensitive:   true,
				Description: "Value of the variable",
				// TODO: do descriptions cause a schema upgrade? how bout the rest of the stuff I'm doing here?
			},
			"category": schema.StringAttribute{
				Required:    true,
				Description: `Whether this is a Terraform or environment variable. Valid values are "terraform" or "env".`,
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(tfe.CategoryEnv),
						string(tfe.CategoryTerraform),
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
			},
			"hcl": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"sensitive": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplaceIf(
						func(ctx context.Context, req planmodifier.BoolRequest, res *boolplanmodifier.RequiresReplaceIfFuncResponse) {
							if req.StateValue.ValueBool() && !req.ConfigValue.ValueBool() {
								res.RequiresReplace = true
							}
						},
						"Force replacement if sensitive argument changed from true to false.",
						"Force replacement if sensitive argument changed from true to false.",
					),
				},
			},
			"workspace_id": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(
						path.MatchRelative().AtParent().AtName("variable_set_id"),
						// TODO: double-check behavior and ensure it includes current attr in that list
					),
					stringvalidator.RegexMatches(
						workspaceIDRegexp,
						"must be a valid workspace ID (ws-<RANDOM STRING>)",
					),
				},
			},
			"variable_set_id": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(
						path.MatchRelative().AtParent().AtName("workspace_id"),
					),
					stringvalidator.RegexMatches(
						variableSetIDRegexp,
						"must be a valid variable set ID (varset-<RANDOM STRING>)",
					),
				},
			},
		},
		Description:         "",
		MarkdownDescription: "",
		DeprecationMessage:  "",
		Version:             1,
	}
}

// AttrGettable is a small enabler for helper functions that need to read one
// attribute of a Plan or State.
type AttrGettable interface {
	GetAttribute(ctx context.Context, path path.Path, target interface{}) diag.Diagnostics
}

// isWorkspaceVariable takes a pointer to a Plan or State from a CRUD request,
// and determines whether the operation should use the workspace variable
// implementation or the variable set variable one.
func isWorkspaceVariable(ctx context.Context, data AttrGettable) bool {
	var variableSetID types.Bool
	// We're ignoring the diagnostics returned by GetAttribute, because we'll
	// be destructuring the entire schema value shortly in the real
	// implementations; any notable problems will be reported at that point.
	data.GetAttribute(ctx, path.Root("variable_set_id"), &variableSetID)
	return variableSetID.IsNull()
}

// Create implements resource.Resource
func (r *resourceTFEVariable) Create(ctx context.Context, req resource.CreateRequest, res *resource.CreateResponse) {
	if isWorkspaceVariable(ctx, &req.Plan) {
		r.createWorkspaceVariable(ctx, req, res)
	} else {
		r.createVariableSetVariable(ctx, req, res)
	}
}

// createWorkspaceVariable is the workspace version of Create.
func (r *resourceTFEVariable) createWorkspaceVariable(ctx context.Context, req resource.CreateRequest, res *resource.CreateResponse) {
	var data modelTFEVariable
	diags := req.Plan.Get(ctx, &data)
	res.Diagnostics.Append(diags...)
	if res.Diagnostics.HasError() {
		return
	}

	key := data.Key.ValueString()
	category := data.Category.ValueString()
	workspaceID := data.WorkspaceID.ValueString()

	// Make a workspace variable
	options := tfe.VariableCreateOptions{
		Key:         data.Key.ValueStringPointer(),
		Value:       data.Value.ValueStringPointer(),
		Category:    tfe.Category(tfe.CategoryType(category)),
		HCL:         data.HCL.ValueBoolPointer(),
		Sensitive:   data.Sensitive.ValueBoolPointer(),
		Description: data.Description.ValueStringPointer(),
	}

	log.Printf("[DEBUG] Create %s variable: %s", category, key)
	variable, err := r.config.Client.Variables.Create(ctx, workspaceID, options)
	if err != nil {
		res.Diagnostics.AddError(
			"Couldn't create variable",
			fmt.Sprintf("Error creating %s variable %s: %s", category, key, err.Error()),
		)
		return
	}

	// Got a variable back, so set state to new values
	data.refreshFromTFEVariable(*variable)
	diags = res.State.Set(ctx, &data)
	res.Diagnostics.Append(diags...)
}

// createVariableSetVariable is the variable set version of Create.
func (r *resourceTFEVariable) createVariableSetVariable(ctx context.Context, req resource.CreateRequest, res *resource.CreateResponse) {
	var data modelTFEVariable
	diags := req.Plan.Get(ctx, &data)
	res.Diagnostics.Append(diags...)
	if res.Diagnostics.HasError() {
		return
	}

	key := data.Key.ValueString()
	category := data.Category.ValueString()
	variableSetID := data.VariableSetID.ValueString()

	options := tfe.VariableSetVariableCreateOptions{
		Key:         data.Key.ValueStringPointer(),
		Value:       data.Value.ValueStringPointer(),
		Category:    tfe.Category(tfe.CategoryType(category)),
		HCL:         data.HCL.ValueBoolPointer(),
		Sensitive:   data.Sensitive.ValueBoolPointer(),
		Description: data.Description.ValueStringPointer(),
	}

	log.Printf("[DEBUG] Create %s variable: %s", category, key)
	variable, err := r.config.Client.VariableSetVariables.Create(ctx, variableSetID, &options)
	if err != nil {
		res.Diagnostics.AddError(
			"Couldn't create variable",
			fmt.Sprintf("Error creating %s variable %s: %s", category, key, err.Error()),
		)
		return
	}

	// We got a variable, so set state to new values
	data.refreshFromTFEVariableSetVariable(*variable)
	diags = res.State.Set(ctx, &data)
	res.Diagnostics.Append(diags...)

}

// Read implements resource.Resource
func (r *resourceTFEVariable) Read(ctx context.Context, req resource.ReadRequest, res *resource.ReadResponse) {
	var data modelTFEVariable
	// Get prior state
	diags := req.State.Get(ctx, &data)
	res.Diagnostics.Append(diags...)
	if res.Diagnostics.HasError() {
		return
	}
	variableID := data.ID.ValueString()

	if data.VariableSetID.IsNull() {
		// Read a workspace variable
		workspaceID := data.WorkspaceID.ValueString()
		variable, err := r.config.Client.Variables.Read(ctx, workspaceID, variableID)
		if err != nil {
			// If it's gone, just say so:
			if err == tfe.ErrResourceNotFound {
				log.Printf("[DEBUG] Variable %s no longer exists", variableID)
				res.State.RemoveResource(ctx)
				return
			}

			// If something worse happened, complain:
			res.Diagnostics.AddError(
				"Couldn't read variable",
				fmt.Sprintf("Error reading variable %s: %s", variableID, err.Error()),
			)
			return
		}

		// We got a variable, so update state:
		data.refreshFromTFEVariable(*variable)
		// Important: If sensitive, transfer over the value from prior state, since that's the last time we were able to know anything about it.
		diags = res.State.Set(ctx, &data)
		res.Diagnostics.Append(diags...)
	} else {
		// TODO Read a variable set variable
	}
}

// Update implements resource.Resource
func (r *resourceTFEVariable) Update(ctx context.Context, req resource.UpdateRequest, res *resource.UpdateResponse) {
	var plan modelTFEVariable
	var state modelTFEVariable
	// Get plan
	diags := req.Plan.Get(ctx, &plan)
	res.Diagnostics.Append(diags...)
	if res.Diagnostics.HasError() {
		return
	}
	// Get state too
	diags = req.State.Get(ctx, &state)
	res.Diagnostics.Append(diags...)
	if res.Diagnostics.HasError() {
		return
	}
	variableID := plan.ID.ValueString()

	if plan.VariableSetID.IsNull() {
		// Update a workspace variable
		workspaceID := plan.WorkspaceID.ValueString()

		// Make update options, BUT:
		//
		// - Omit Value IF no change was planned and the variable is sensitive!
		// (If we don't do that, we can accidentally reset it to the last known
		// value when we shouldn't, e.g. when ignore_changes is used.) That
		// means it's possible for our knowledge about the value to be out of
		// date, but this is about the best we can do when something's
		// impossible to inspect and not always safe to hard-overwrite.
		//
		// - Always omit Category, which we can never update anyway (see schema).
		var value *string
		if state.Sensitive.ValueBool() && state.Value.ValueString() == plan.Value.ValueString() {
			value = nil
		} else {
			value = plan.Value.ValueStringPointer()
		}
		options := tfe.VariableUpdateOptions{
			Key:         plan.Key.ValueStringPointer(),
			Value:       value,
			Description: plan.Description.ValueStringPointer(),
			HCL:         plan.HCL.ValueBoolPointer(),
			Sensitive:   plan.Sensitive.ValueBoolPointer(),
		}

		// Do it
		log.Printf("[DEBUG] Update variable: %s", variableID)
		variable, err := r.config.Client.Variables.Update(ctx, workspaceID, variableID, options)
		if err != nil {
			res.Diagnostics.AddError(
				"Couldn't update variable",
				fmt.Sprintf("Error updating variable %s: %s", variableID, err.Error()),
			)
		}
		// Update state
		plan.refreshFromTFEVariable(*variable)
		diags = res.State.Set(ctx, &plan)
		res.Diagnostics.Append(diags...)
	} else {
		// TODO update a variable set variable
	}
}

// Delete implements resource.Resource
func (r *resourceTFEVariable) Delete(ctx context.Context, req resource.DeleteRequest, res *resource.DeleteResponse) {
	var data modelTFEVariable
	diags := req.State.Get(ctx, &data)
	res.Diagnostics.Append(diags...)
	if res.Diagnostics.HasError() {
		return
	}

	variableID := data.ID.ValueString()

	if data.VariableSetID.IsNull() {
		// Delete a workspace variable
		workspaceID := data.WorkspaceID.ValueString()
		log.Printf("[DEBUG] Delete variable: %s", variableID)
		err := r.config.Client.Variables.Delete(ctx, workspaceID, variableID)
		// Ignore 404s for delete
		if err != nil && err != tfe.ErrResourceNotFound {
			res.Diagnostics.AddError(
				"Couldn't delete variable",
				fmt.Sprintf("Error deleting variable %s: %s", variableID, err.Error()),
			)
			return
		}
		// Resource gets implicitly deleted from response state if no error.
	} else {
		// TODO delete a variable set variable
	}
}

// Compile-time interface check
var _ resource.ResourceWithConfigure = &resourceTFEVariable{}

// NewResourceVariable is a resource function for the framework provider.
func NewResourceVariable() resource.Resource {
	return &resourceTFEVariable{}
}
