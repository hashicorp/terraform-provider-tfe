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

// resourceTFEVariable implements the tfe_variable resource type. Note: Much of
// the complexity of this type's Resource implementation is because the
// tfe_variable resource is an abstraction over two parallel APIs, so each
// primary CRUD method needs to call different client methods (with different
// argument types and return types) depending on whether the workspace_id or
// variable_set_id attribute is defined.
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
func (r *resourceTFEVariable) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Early exit if provider is unconfigured (i.e. we're only validating config or something)
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(ConfiguredClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected resource Configure type",
			fmt.Sprintf("Expected tfe.ConfiguredClient, got %T. This is a bug in the tfe provider, so please report it on GitHub.", req.ProviderData),
		)
	}
	r.config = client
}

// Metadata implements resource.Resource
func (r *resourceTFEVariable) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "tfe_variable"
}

// Schema implements resource.Resource
func (r *resourceTFEVariable) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
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
						func(ctx context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
							var stateSensitive types.Bool
							diags := req.State.GetAttribute(ctx, path.Root("sensitive"), &stateSensitive)
							if diags.HasError() {
								resp.Diagnostics.Append(diags...)
								return
							}
							if stateSensitive.ValueBool() && req.PlanValue.ValueString() != req.StateValue.ValueString() {
								resp.RequiresReplace = true
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
						func(ctx context.Context, req planmodifier.BoolRequest, resp *boolplanmodifier.RequiresReplaceIfFuncResponse) {
							if req.StateValue.ValueBool() && !req.ConfigValue.ValueBool() {
								resp.RequiresReplace = true
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

// isWorkspaceVariable is a helper function for switching between tfe_variable's
// two separate CRUD implementations.
func isWorkspaceVariable(ctx context.Context, data AttrGettable) bool {
	var variableSetID types.Bool
	// We're ignoring the diagnostics returned by GetAttribute, because we'll
	// be destructuring the entire schema value shortly in the real
	// implementations; any notable problems will be reported at that point.
	data.GetAttribute(ctx, path.Root("variable_set_id"), &variableSetID)
	return variableSetID.IsNull()
}

// Create implements resource.Resource
func (r *resourceTFEVariable) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if isWorkspaceVariable(ctx, &req.Plan) {
		r.createWithWorkspace(ctx, req, resp)
	} else {
		r.createWithVariableSet(ctx, req, resp)
	}
}

// createWithWorkspace is the workspace version of Create.
func (r *resourceTFEVariable) createWithWorkspace(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data modelTFEVariable
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	key := data.Key.ValueString()
	category := data.Category.ValueString()
	workspaceID := data.WorkspaceID.ValueString()

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
		resp.Diagnostics.AddError(
			"Couldn't create variable",
			fmt.Sprintf("Error creating %s variable %s: %s", category, key, err.Error()),
		)
		return
	}

	// Got a variable back, so set state to new values
	data.refreshFromTFEVariable(*variable)
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

// createWithVariableSet is the variable set version of Create.
func (r *resourceTFEVariable) createWithVariableSet(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data modelTFEVariable
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
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
		resp.Diagnostics.AddError(
			"Couldn't create variable",
			fmt.Sprintf("Error creating %s variable %s: %s", category, key, err.Error()),
		)
		return
	}

	// We got a variable, so set state to new values
	data.refreshFromTFEVariableSetVariable(*variable)
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)

}

// Read implements resource.Resource
func (r *resourceTFEVariable) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if isWorkspaceVariable(ctx, &req.State) {
		r.readWithWorkspace(ctx, req, resp)
	} else {
		r.readWithVariableSet(ctx, req, resp)
	}
}

// readWithWorkspace is the workspace version of Read.
func (r *resourceTFEVariable) readWithWorkspace(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data modelTFEVariable
	// Get prior state
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	variableID := data.ID.ValueString()
	workspaceID := data.WorkspaceID.ValueString()
	variable, err := r.config.Client.Variables.Read(ctx, workspaceID, variableID)
	if err != nil {
		// If it's gone: that's not an error, but we are done.
		if err == tfe.ErrResourceNotFound {
			log.Printf("[DEBUG] Variable %s no longer exists", variableID)
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError(
				"Couldn't read variable",
				fmt.Sprintf("Error reading variable %s: %s", variableID, err.Error()),
			)
		}
		return
	}

	// We got a variable, so update state:
	data.refreshFromTFEVariable(*variable)
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

// readWithVariableSet is the variable set version of Read.
func (r *resourceTFEVariable) readWithVariableSet(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data modelTFEVariable
	// Get prior state
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	variableID := data.ID.ValueString()
	variableSetID := data.VariableSetID.ValueString()
	variable, err := r.config.Client.VariableSetVariables.Read(ctx, variableSetID, variableID)
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			// If it's gone: that's not an error, but we are done.
			log.Printf("[DEBUG] Variable %s no longer exists", variableID)
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError(
				"Couldn't read variable",
				fmt.Sprintf("Error reading variable %s: %s", variableID, err.Error()),
			)
		}
		return
	}

	// We got a variable, so update state:
	data.refreshFromTFEVariableSetVariable(*variable)
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

// Update implements resource.Resource
func (r *resourceTFEVariable) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if isWorkspaceVariable(ctx, &req.Plan) {
		r.updateWithWorkspace(ctx, req, resp)
	} else {
		r.updateWithVariableSet(ctx, req, resp)
	}
}

// updateWithWorkspace is the workspace version of Update.
func (r *resourceTFEVariable) updateWithWorkspace(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get both plan and state; must compare them to handle sensitive values safely.
	var plan modelTFEVariable
	var state modelTFEVariable
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	variableID := plan.ID.ValueString()
	workspaceID := plan.WorkspaceID.ValueString()

	// When a tfe update options struct uses pointers, any nil fields are
	// omitted in the API request, preserving the prior value. Here, we always
	// want to omit Category (can't update it, cf. the schema), and only
	// *sometimes* want to include Value.
	options := tfe.VariableUpdateOptions{
		Key:         plan.Key.ValueStringPointer(),
		Description: plan.Description.ValueStringPointer(),
		HCL:         plan.HCL.ValueBoolPointer(),
		Sensitive:   plan.Sensitive.ValueBoolPointer(),
	}
	// Specifically, we ONLY want to set Value if our planned value would be a
	// CHANGE from the prior state. This is so we don't accidentally reset the
	// value of a sensitive variable on unrelated changes when `ignore_changes =
	// [value]` is set. (Basically: since we can't observe the real-world
	// condition of a sensitive variable, we don't KNOW whether setting it to
	// our last-known value is a safely idempotent operation or not. This is why
	// Terraform doesn't promise that it can manage drift at all for write-only
	// attributes.)
	if state.Value.ValueString() != plan.Value.ValueString() {
		options.Value = plan.Value.ValueStringPointer()
	}

	log.Printf("[DEBUG] Update variable: %s", variableID)
	variable, err := r.config.Client.Variables.Update(ctx, workspaceID, variableID, options)
	if err != nil {
		resp.Diagnostics.AddError(
			"Couldn't update variable",
			fmt.Sprintf("Error updating variable %s: %s", variableID, err.Error()),
		)
	}
	// Update state
	plan.refreshFromTFEVariable(*variable)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

// updateWithVariableSet is the variable set version of Update.
func (r *resourceTFEVariable) updateWithVariableSet(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get both plan and state; must compare them to handle sensitive values safely.
	var plan modelTFEVariable
	var state modelTFEVariable
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	variableID := plan.ID.ValueString()
	variableSetID := plan.VariableSetID.ValueString()

	options := &tfe.VariableSetVariableUpdateOptions{
		Key:         plan.Key.ValueStringPointer(),
		Description: plan.Description.ValueStringPointer(),
		HCL:         plan.HCL.ValueBoolPointer(),
		Sensitive:   plan.Sensitive.ValueBoolPointer(),
	}
	// We ONLY want to set Value if our planned value would be a CHANGE from the
	// prior state. See comments in updateWithWorkspace for more color.
	if state.Value.ValueString() != plan.Value.ValueString() {
		options.Value = plan.Value.ValueStringPointer()
	}

	log.Printf("[DEBUG] Update variable: %s", variableID)
	variable, err := r.config.Client.VariableSetVariables.Update(ctx, variableSetID, variableID, options)
	if err != nil {
		resp.Diagnostics.AddError(
			"Couldn't update variable",
			fmt.Sprintf("Error updating variable %s: %s", variableID, err.Error()),
		)
	}
	// Update state
	plan.refreshFromTFEVariableSetVariable(*variable)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

// Delete implements resource.Resource
func (r *resourceTFEVariable) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if isWorkspaceVariable(ctx, &req.State) {
		r.deleteWithWorkspace(ctx, req, resp)
	} else {
		r.deleteWithVariableSet(ctx, req, resp)
	}
}

// deleteWithWorkspace is the workspace version of Delete.
func (r *resourceTFEVariable) deleteWithWorkspace(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data modelTFEVariable
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	variableID := data.ID.ValueString()
	workspaceID := data.WorkspaceID.ValueString()

	log.Printf("[DEBUG] Delete variable: %s", variableID)
	err := r.config.Client.Variables.Delete(ctx, workspaceID, variableID)
	// Ignore 404s for delete
	if err != nil && err != tfe.ErrResourceNotFound {
		resp.Diagnostics.AddError(
			"Couldn't delete variable",
			fmt.Sprintf("Error deleting variable %s: %s", variableID, err.Error()),
		)
	}
	// Resource is implicitly deleted from resp.State if diagnostics have no errors.
}

// deleteWithVariableSet is the variable set version of Delete.
func (r *resourceTFEVariable) deleteWithVariableSet(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data modelTFEVariable
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	variableID := data.ID.ValueString()
	variableSetID := data.VariableSetID.ValueString()

	log.Printf("[DEBUG] Delete variable: %s", variableID)
	err := r.config.Client.VariableSetVariables.Delete(ctx, variableSetID, variableID)
	// Ignore 404s for delete
	if err != nil && err != tfe.ErrResourceNotFound {
		resp.Diagnostics.AddError(
			"Couldn't delete variable",
			fmt.Sprintf("Error deleting variable %s: %s", variableID, err.Error()),
		)
	}
	// Resource is implicitly deleted from resp.State if diagnostics have no errors.
}

// Compile-time interface check
var _ resource.ResourceWithConfigure = &resourceTFEVariable{}

// NewResourceVariable is a resource function for the framework provider.
func NewResourceVariable() resource.Resource {
	return &resourceTFEVariable{}
}
