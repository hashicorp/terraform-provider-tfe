// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	tfev2 "github.com/hashicorp/go-tfe/v2"
	"github.com/hashicorp/go-tfe/v2/api/models"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
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
	ID             types.String `tfsdk:"id"`
	Key            types.String `tfsdk:"key"`
	Value          types.String `tfsdk:"value"`
	ValueWO        types.String `tfsdk:"value_wo"`
	ValueWOVersion types.Int64  `tfsdk:"value_wo_version"`
	ReadableValue  types.String `tfsdk:"readable_value"`
	Category       types.String `tfsdk:"category"`
	Description    types.String `tfsdk:"description"`
	HCL            types.Bool   `tfsdk:"hcl"`
	Sensitive      types.Bool   `tfsdk:"sensitive"`
	WorkspaceID    types.String `tfsdk:"workspace_id"`
	VariableSetID  types.String `tfsdk:"variable_set_id"`
}

type modelTFEVariableIdentity struct {
	ID types.String `tfsdk:"id"`
	// Can be either a variable set id or workspace id
	ConfigurableID types.String `tfsdk:"configurable_id"`
	Hostname       types.String `tfsdk:"hostname"`
}

// modelFromV2Var builds a modelTFEVariable struct from a go-tfe v2 VarsEnvelopeable
// (plus the last known value and write-only version).
func modelFromV2Var(resp models.VarsEnvelopeable, lastValue types.String, valueWOVersion types.Int64) modelTFEVariable {
	data := resp.GetData()
	if data == nil {
		return modelTFEVariable{}
	}

	m := modelTFEVariable{
		ID:             types.StringValue(valueOrZero(data.GetId())),
		ValueWOVersion: valueWOVersion,
		WorkspaceID:    types.StringNull(),
		VariableSetID:  types.StringNull(),
	}

	if attrs := data.GetAttributes(); attrs != nil {
		m.Key = types.StringValue(valueOrZero(attrs.GetKey()))
		m.Description = types.StringValue(valueOrZero(attrs.GetDescription()))
		m.HCL = types.BoolValue(valueOrZero(attrs.GetHcl()))
		m.Sensitive = types.BoolValue(valueOrZero(attrs.GetSensitive()))

		if cat := attrs.GetCategory(); cat != nil {
			m.Category = types.StringValue(cat.String())
		} else {
			m.Category = types.StringValue("")
		}

		// Value: sensitive vars return null from the API; carry forward last known value.
		rawValue := valueOrZero(attrs.GetValue())
		m.Value = types.StringValue(rawValue)
		if valueOrZero(attrs.GetSensitive()) {
			m.Value = lastValue
			m.ReadableValue = types.StringNull()
		} else {
			m.ReadableValue = m.Value
		}
	}

	// Write-only mode: clear value and readable_value.
	if !valueWOVersion.IsNull() {
		m.Value = types.StringValue("")
		m.ReadableValue = types.StringValue("")
	}

	// Extract workspace or varset ID from relationships.
	rels := data.GetRelationships()
	if wsID := v2VarWorkspaceID(rels); wsID != "" {
		m.WorkspaceID = types.StringValue(wsID)
	}
	if vsID := v2VarVarsetID(rels); vsID != "" {
		m.VariableSetID = types.StringValue(vsID)
	}

	return m
}

// v2VarEnvelope constructs a VarsEnvelope request body from variable attributes.
// v2VarWorkspaceID returns the workspace ID from a variable relationships
// object, or empty string when absent.
func v2VarWorkspaceID(rels models.Vars_relationshipsable) string {
	if rels == nil {
		return ""
	}
	ws := rels.GetWorkspace()
	if ws == nil || ws.GetData() == nil {
		return ""
	}
	return valueOrZero(ws.GetData().GetId())
}

// v2VarVarsetID returns the variable-set ID from a variable relationships
// object, or empty string when absent.
func v2VarVarsetID(rels models.Vars_relationshipsable) string {
	if rels == nil {
		return ""
	}
	vs := rels.GetVarset()
	if vs == nil || vs.GetData() == nil {
		return ""
	}
	return valueOrZero(vs.GetData().GetId())
}

func v2VarEnvelope(key, value, category, description string, hcl, sensitive bool) models.VarsEnvelopeable {
	attrs := models.NewVars_attributes()
	attrs.SetKey(ptr(key))
	attrs.SetValue(ptr(value))
	attrs.SetDescription(ptr(description))
	attrs.SetHcl(ptr(hcl))
	attrs.SetSensitive(ptr(sensitive))

	// Parse the category enum.
	cat, _ := models.ParseVars_attributes_category(category)
	if catVal, ok := cat.(*models.Vars_attributes_category); ok {
		attrs.SetCategory(catVal)
	}

	varType := models.VARS_VARS_TYPE
	varData := models.NewVars()
	varData.SetTypeEscaped(&varType)
	varData.SetAttributes(attrs)

	envelope := models.NewVarsEnvelope()
	envelope.SetData(varData)
	return envelope
}

// v2VarUpdateEnvelope constructs a VarsEnvelope for PATCH requests (value optional).
func v2VarUpdateEnvelope(key, description string, hcl, sensitive bool, value *string) models.VarsEnvelopeable {
	attrs := models.NewVars_attributes()
	attrs.SetKey(ptr(key))
	attrs.SetDescription(ptr(description))
	attrs.SetHcl(ptr(hcl))
	attrs.SetSensitive(ptr(sensitive))
	if value != nil {
		attrs.SetValue(value)
	}

	varType := models.VARS_VARS_TYPE
	varData := models.NewVars()
	varData.SetTypeEscaped(&varType)
	varData.SetAttributes(attrs)

	envelope := models.NewVarsEnvelope()
	envelope.SetData(varData)
	return envelope
}

// Configure implements resource.ResourceWithConfigure
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
	resp.ResourceBehavior = resource.ResourceBehavior{
		MutableIdentity: true,
	}
}

// Schema implements resource.Resource
func (r *resourceTFEVariable) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Creates, updates and destroys variables." +
			"\n\n-> **Note:** While the `value` field may be referenced in other resources, for safety it is always treated as sensitive. This means that it will always be redacted from plan outputs, and any other resource attributes which depend on it will also be redacted. The `readable_value` attribute is not sensitive, and will not be redacted; instead, it will be null if the variable is sensitive. This allows other resources to reference it, while keeping their plan outputs readable." +
			"\n\n~> **Note:** When `sensitive` is set to `true`, Terraform cannot detect and repair drift if `value` is later changed out-of-band via the HCP Terraform UI. Terraform will only change the value for a sensitive variable if you change `value` in the configuration, so that it no longer matches the last known value in the state.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the variable.",
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
				Description: "Value of the variable. Either `value` or `value_wo` can be provided, but not both.",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("value_wo")),
				},
			},
			"value_wo": schema.StringAttribute{
				Optional:    true,
				WriteOnly:   true,
				Sensitive:   true,
				Description: "Value of the variable in write-only mode. `Write-only` attributes function similarly to their non-write-only counterparts, but are never stored to state and do not display in the Terraform plan output. Can be used in place of `value`. Either `value` or `value_wo` can be provided, but not both.",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("value")),
					stringvalidator.AlsoRequires(path.MatchRoot("value_wo_version")),
				},
			},
			"value_wo_version": schema.Int64Attribute{
				Optional:    true,
				Description: "Version identifier for the write-only value. Required when `value_wo` is specified to trigger updates. Cannot be used with `value`.",
				Validators: []validator.Int64{
					int64validator.ConflictsWith((path.MatchRoot("value"))),
					int64validator.AlsoRequires(path.MatchRoot("value_wo")),
				},
			},
			"category": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Whether this is a Terraform or environment variable. Valid values are `terraform` or `env`.",
				Validators: []validator.String{
					stringvalidator.OneOf("env", "terraform"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Description: "Description of the variable.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"hcl": schema.BoolAttribute{
				Description: "Whether to evaluate the value of the variable as a string of HCL code. Has no effect for environment variables. Defaults to `false`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"sensitive": schema.BoolAttribute{
				Description: "Whether the value is sensitive. If true then the variable is written once and not visible thereafter. Defaults to false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
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
				Description: "ID of the workspace that owns the variable. Exactly one of `workspace_id` or `variable_set_id` must be provided.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(
						path.MatchRelative().AtParent().AtName("variable_set_id"),
					),
					stringvalidator.RegexMatches(
						workspaceIDRegexp,
						"must be a valid workspace ID (ws-<RANDOM STRING>)",
					),
				},
			},
			"variable_set_id": schema.StringAttribute{
				Description: "ID of the variable set that owns the variable. Exactly one of `workspace_id` or `variable_set_id` must be provided.",
				Optional:    true,
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
			"readable_value": schema.StringAttribute{
				Computed:    true,
				Description: "Only present if the variable is non-sensitive. A copy of the value which will not be marked as sensitive in plan outputs. Will be `null` if the variable is sensitive. Cannot be explicitly set in the resource configuration.",
				PlanModifiers: []planmodifier.String{
					&updateReadableValuePlanModifier{},
				},
			},
		},
		Version: 1,
	}
}

func (r *resourceTFEVariable) IdentitySchema(ctx context.Context, req resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"id": identityschema.StringAttribute{
				RequiredForImport: true,
			},
			"configurable_id": identityschema.StringAttribute{
				RequiredForImport: true,
			},
			"hostname": identityschema.StringAttribute{
				OptionalForImport: true,
			},
		},
	}
}

// isWorkspaceVariable is a helper function for switching between tfe_variable's
// two separate CRUD implementations.
func isWorkspaceVariable(ctx context.Context, data AttrGettable) bool {
	var workspaceID types.String
	// We're ignoring the diagnostics returned by GetAttribute, because we'll
	// be destructuring the entire schema value shortly in the real
	// implementations; any notable problems will be reported at that point.
	data.GetAttribute(ctx, path.Root("workspace_id"), &workspaceID)
	return !workspaceID.IsNull()
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

	var config modelTFEVariable
	diags = req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	key := data.Key.ValueString()
	category := data.Category.ValueString()
	workspaceID := data.WorkspaceID.ValueString()

	// Determine value: use value_wo if set, otherwise the normal value.
	value := data.Value.ValueString()
	if !config.ValueWO.IsNull() {
		value = config.ValueWO.ValueString()
	}

	envelope := v2VarEnvelope(key, value, category, data.Description.ValueString(), data.HCL.ValueBool(), data.Sensitive.ValueBool())

	log.Printf("[DEBUG] Create %s variable: %s", category, key)
	varResp, err := r.config.ClientV2.API.Workspaces().ByWorkspace_id(workspaceID).Vars().Post(ctx, envelope, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating variable",
			fmt.Sprintf("Couldn't create %s variable %s: %s", category, key, err.Error()),
		)
		return
	}

	result := modelFromV2Var(varResp, data.Value, config.ValueWOVersion)
	diags = resp.State.Set(ctx, &result)
	resp.Diagnostics.Append(diags...)

	identity := modelTFEVariableIdentity{
		ID:             result.ID,
		Hostname:       types.StringValue(r.config.Client.BaseURL().Host),
		ConfigurableID: result.WorkspaceID,
	}
	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identity)...)
}

// createWithVariableSet is the variable set version of Create.
func (r *resourceTFEVariable) createWithVariableSet(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data modelTFEVariable
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var config modelTFEVariable
	diags = req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	key := data.Key.ValueString()
	category := data.Category.ValueString()
	variableSetID := data.VariableSetID.ValueString()

	// Determine value: use value_wo if set, otherwise the normal value.
	value := data.Value.ValueString()
	if !config.ValueWO.IsNull() {
		value = config.ValueWO.ValueString()
	}

	envelope := v2VarEnvelope(key, value, category, data.Description.ValueString(), data.HCL.ValueBool(), data.Sensitive.ValueBool())

	log.Printf("[DEBUG] Create %s variable: %s", category, key)
	varResp, err := r.config.ClientV2.API.Varsets().ByVarset_id(variableSetID).Relationships().Vars().Post(ctx, envelope, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating variable",
			fmt.Sprintf("Couldn't create %s variable %s: %s", category, key, err.Error()),
		)
		return
	}

	result := modelFromV2Var(varResp, data.Value, config.ValueWOVersion)
	// For varset vars, the workspace relationship won't be present; ensure null.
	result.WorkspaceID = types.StringNull()
	if result.VariableSetID.IsNull() {
		result.VariableSetID = types.StringValue(variableSetID)
	}
	diags = resp.State.Set(ctx, &result)
	resp.Diagnostics.Append(diags...)

	identity := modelTFEVariableIdentity{
		ID:             result.ID,
		Hostname:       types.StringValue(r.config.Client.BaseURL().Host),
		ConfigurableID: result.VariableSetID,
	}
	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identity)...)
}

// Read implements resource.Resource
func (r *resourceTFEVariable) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if isWorkspaceVariable(ctx, &req.State) {
		r.readWithWorkspace(ctx, req, resp)
	} else {
		r.readWithVariableSet(ctx, req, resp)
	}
}

func (r *resourceTFEVariable) setReadIdentity(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse, variableID string, configurableID string) {
	if resp.Identity == nil {
		return
	}

	if req.Identity != nil {
		currentIdentity := &modelTFEVariableIdentity{}
		resp.Diagnostics.Append(req.Identity.Get(ctx, &currentIdentity)...)
		if resp.Diagnostics.HasError() {
			return
		}
		if currentIdentity != nil && !currentIdentity.ID.IsNull() {
			return
		}
	}

	identity := modelTFEVariableIdentity{
		ID:             types.StringValue(variableID),
		Hostname:       types.StringValue(r.config.Client.BaseURL().Host),
		ConfigurableID: types.StringValue(configurableID),
	}
	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identity)...)
}

// readWithWorkspace is the workspace version of Read.
func (r *resourceTFEVariable) readWithWorkspace(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data modelTFEVariable
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	variableID := data.ID.ValueString()
	workspaceID := data.WorkspaceID.ValueString()

	varResp, err := r.config.ClientV2.API.Workspaces().ByWorkspace_id(workspaceID).Vars().ById(variableID).Get(ctx, nil)
	if err != nil {
		if errors.Is(err, tfev2.ErrNotFound) {
			log.Printf("[DEBUG] Variable %s no longer exists", variableID)
			r.setReadIdentity(ctx, req, resp, variableID, workspaceID)
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError(
				"Error reading variable",
				fmt.Sprintf("Couldn't read variable %s: %s", variableID, err.Error()),
			)
		}
		return
	}

	result := modelFromV2Var(varResp, data.Value, data.ValueWOVersion)
	diags = resp.State.Set(ctx, &result)
	resp.Diagnostics.Append(diags...)

	identity := modelTFEVariableIdentity{
		ID:             result.ID,
		Hostname:       types.StringValue(r.config.Client.BaseURL().Host),
		ConfigurableID: result.WorkspaceID,
	}
	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identity)...)
}

// readWithVariableSet is the variable set version of Read.
func (r *resourceTFEVariable) readWithVariableSet(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data modelTFEVariable
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	variableID := data.ID.ValueString()
	variableSetID := data.VariableSetID.ValueString()

	varResp, err := r.config.ClientV2.API.Varsets().ByVarset_id(variableSetID).Relationships().Vars().ById(variableID).Get(ctx, nil)
	if err != nil {
		if errors.Is(err, tfev2.ErrNotFound) {
			log.Printf("[DEBUG] Variable %s no longer exists", variableID)
			r.setReadIdentity(ctx, req, resp, variableID, variableSetID)
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError(
				"Error reading variable",
				fmt.Sprintf("Couldn't read variable %s: %s", variableID, err.Error()),
			)
		}
		return
	}

	result := modelFromV2Var(varResp, data.Value, data.ValueWOVersion)
	// For varset vars, ensure the varset ID is correct (may not be in response).
	result.WorkspaceID = types.StringNull()
	if result.VariableSetID.IsNull() {
		result.VariableSetID = types.StringValue(variableSetID)
	}
	diags = resp.State.Set(ctx, &result)
	resp.Diagnostics.Append(diags...)

	identity := modelTFEVariableIdentity{
		ID:             result.ID,
		Hostname:       types.StringValue(r.config.Client.BaseURL().Host),
		ConfigurableID: result.VariableSetID,
	}
	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identity)...)
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
	var plan modelTFEVariable
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state modelTFEVariable
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	var config modelTFEVariable
	diags = req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	variableID := plan.ID.ValueString()
	workspaceID := plan.WorkspaceID.ValueString()

	// Determine value to update (nil means no value change).
	valueToUpdate := r.determineValueForUpdate(plan, state, config)

	envelope := v2VarUpdateEnvelope(
		plan.Key.ValueString(),
		plan.Description.ValueString(),
		plan.HCL.ValueBool(),
		plan.Sensitive.ValueBool(),
		valueToUpdate,
	)

	log.Printf("[DEBUG] Update variable: %s", variableID)
	varResp, err := r.config.ClientV2.API.Workspaces().ByWorkspace_id(workspaceID).Vars().ById(variableID).Patch(ctx, envelope, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating variable",
			fmt.Sprintf("Couldn't update variable %s: %s", variableID, err.Error()),
		)
		return
	}

	result := modelFromV2Var(varResp, plan.Value, config.ValueWOVersion)
	diags = resp.State.Set(ctx, &result)
	resp.Diagnostics.Append(diags...)

	currentIdentity := &modelTFEVariableIdentity{}
	resp.Diagnostics.Append(req.Identity.Get(ctx, &currentIdentity)...)
	if !resp.Diagnostics.HasError() && (currentIdentity == nil || currentIdentity.ID.IsNull()) {
		identity := modelTFEVariableIdentity{
			ID:             result.ID,
			Hostname:       types.StringValue(r.config.Client.BaseURL().Host),
			ConfigurableID: result.WorkspaceID,
		}
		resp.Diagnostics.Append(resp.Identity.Set(ctx, &identity)...)
	}
}

// updateWithVariableSet is the variable set version of Update.
func (r *resourceTFEVariable) updateWithVariableSet(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
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
	var config modelTFEVariable
	diags = req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	variableID := plan.ID.ValueString()
	variableSetID := plan.VariableSetID.ValueString()

	valueToUpdate := r.determineValueForUpdate(plan, state, config)

	envelope := v2VarUpdateEnvelope(
		plan.Key.ValueString(),
		plan.Description.ValueString(),
		plan.HCL.ValueBool(),
		plan.Sensitive.ValueBool(),
		valueToUpdate,
	)

	log.Printf("[DEBUG] Update variable: %s", variableID)
	varResp, err := r.config.ClientV2.API.Varsets().ByVarset_id(variableSetID).Relationships().Vars().ById(variableID).Patch(ctx, envelope, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating variable",
			fmt.Sprintf("Couldn't update variable %s: %s", variableID, err.Error()),
		)
		return
	}

	result := modelFromV2Var(varResp, plan.Value, config.ValueWOVersion)
	result.WorkspaceID = types.StringNull()
	if result.VariableSetID.IsNull() {
		result.VariableSetID = types.StringValue(variableSetID)
	}
	diags = resp.State.Set(ctx, &result)
	resp.Diagnostics.Append(diags...)

	currentIdentity := &modelTFEVariableIdentity{}
	resp.Diagnostics.Append(req.Identity.Get(ctx, &currentIdentity)...)
	if !resp.Diagnostics.HasError() && (currentIdentity == nil || currentIdentity.ID.IsNull()) {
		identity := modelTFEVariableIdentity{
			ID:             result.ID,
			Hostname:       types.StringValue(r.config.Client.BaseURL().Host),
			ConfigurableID: result.VariableSetID,
		}
		resp.Diagnostics.Append(resp.Identity.Set(ctx, &identity)...)
	}
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
	err := r.config.ClientV2.API.Workspaces().ByWorkspace_id(workspaceID).Vars().ById(variableID).Delete(ctx, nil)
	// Ignore 404s for delete
	if err != nil && !errors.Is(err, tfev2.ErrNotFound) {
		resp.Diagnostics.AddError(
			"Error deleting variable",
			fmt.Sprintf("Couldn't delete variable %s: %s", variableID, err.Error()),
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
	err := r.config.ClientV2.API.Varsets().ByVarset_id(variableSetID).Relationships().Vars().ById(variableID).Delete(ctx, nil)
	// Ignore 404s for delete
	if err != nil && !errors.Is(err, tfev2.ErrNotFound) {
		resp.Diagnostics.AddError(
			"Error deleting variable",
			fmt.Sprintf("Couldn't delete variable %s: %s", variableID, err.Error()),
		)
	}
	// Resource is implicitly deleted from resp.State if diagnostics have no errors.
}

var resourceTFEVariableSchemaV0 = schema.Schema{
	Version: 0,
	Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed: true,
		},
		"key": schema.StringAttribute{
			Required: true,
		},
		"value": schema.StringAttribute{
			Optional:  true,
			Computed:  true,
			Default:   stringdefault.StaticString(""),
			Sensitive: true,
		},
		"category": schema.StringAttribute{
			Required: true,
			Validators: []validator.String{
				stringvalidator.OneOf("env", "terraform"),
			},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
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
		},
		// Unlike the modern tfe_variable schema, this workspace_id was of the
		// form org_name/ws_name.
		"workspace_id": schema.StringAttribute{
			Required: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
	},
}

// UpgradeState implements resource.ResourceWithUpgradeState
func (r *resourceTFEVariable) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		// Upgrader from version 0 to 1 (schema 1 introduced in v0.15.1, commit
		// 88a646c; changed workspace_id to use external ID instead of
		// org/ws_name)
		0: {
			PriorSchema: &resourceTFEVariableSchemaV0,
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				// Using modern model struct for oldData, since it's a superset of the old attrs.
				var oldData modelTFEVariable
				diags := req.State.Get(ctx, &oldData)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}
				// Get the workspace external ID using the v2 client.
				oldWorkspaceID := oldData.WorkspaceID.ValueString()
				newWorkspaceID, err := fetchWorkspaceExternalIDV2(oldWorkspaceID, r.config.ClientV2)
				if err != nil {
					resp.Diagnostics.AddError(
						"Error reading workspace",
						fmt.Sprintf("Couldn't read workspace %s while trying to upgrade state of tfe_variable %s: %s", oldWorkspaceID, oldData.ID.ValueString(), err.Error()),
					)
					return
				}
				newData := modelTFEVariable{
					// Updated ID
					WorkspaceID: types.StringValue(newWorkspaceID),
					// Other existing attrs unchanged
					ID:        oldData.ID,
					Key:       oldData.Key,
					Value:     oldData.Value,
					Category:  oldData.Category,
					HCL:       oldData.HCL,
					Sensitive: oldData.Sensitive,
					// New attrs didn't exist
					Description:   types.StringNull(),
					VariableSetID: types.StringNull(),
				}
				diags = resp.State.Set(ctx, newData)
				resp.Diagnostics.Append(diags...)
			},
		},
	}
}

// ImportState implements resource.ResourceWithImportState
func (r *resourceTFEVariable) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID != "" {
		r.legacyImportByID(ctx, req, resp)
		return
	}

	var identityData modelTFEVariableIdentity
	// We are reading an identity here
	resp.Diagnostics.Append(req.Identity.Get(ctx, &identityData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data := modelTFEVariable{
		ID: identityData.ID,
	}
	varsetIDUsed := variableSetIDRegexp.MatchString(identityData.ConfigurableID.ValueString())

	if varsetIDUsed {
		// The Configurable ID is a varset
		data.VariableSetID = identityData.ConfigurableID
		data.WorkspaceID = types.StringNull()
	} else {
		// The Configurable ID is a workspace
		data.VariableSetID = types.StringNull()
		data.WorkspaceID = identityData.ConfigurableID
	}

	diags := resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

// determineValueForUpdate is invoked only after terraform determines that an attribute update is needed.
// note that the update can be triggered by other attributes outside of the value/value_wo attributes.
// this function compares the ValueWOVersion vs Value to ensure that during api update call, value is not mistakenly unset.
// Returns nil if no value update is needed.
func (r *resourceTFEVariable) determineValueForUpdate(plan, state, config modelTFEVariable) *string {
	// Determine if we're using write-only value in plan vs state
	usingWriteOnlyInPlan := !plan.ValueWOVersion.IsNull()
	usingWriteOnlyInState := !state.ValueWOVersion.IsNull()

	// Case 1: Switching FROM value TO value_wo
	if !usingWriteOnlyInState && usingWriteOnlyInPlan && !config.ValueWO.IsNull() {
		return config.ValueWO.ValueStringPointer()
	}
	// Case 2: Switching FROM value_wo TO value
	if usingWriteOnlyInState && !usingWriteOnlyInPlan && !plan.Value.IsNull() {
		return plan.Value.ValueStringPointer()
	}
	// Case 3: value_wo version changed in plan
	if usingWriteOnlyInPlan && plan.ValueWOVersion.ValueInt64() != state.ValueWOVersion.ValueInt64() && !config.ValueWO.IsNull() {
		return config.ValueWO.ValueStringPointer()
	}
	// Case 4: Regular value changed. Only set Value if our planned value would be a CHANGE from
	// the prior state. This prevents accidentally resetting the value of sensitive variables on
	// unrelated changes when ignore_changes=[value] is set.
	if state.Value.ValueString() != plan.Value.ValueString() {
		return plan.Value.ValueStringPointer()
	}
	return nil
}

func (r *resourceTFEVariable) legacyImportByID(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	s := strings.SplitN(req.ID, "/", 3)
	if len(s) != 3 {
		resp.Diagnostics.AddError(
			"Error importing variable",
			fmt.Sprintf("Invalid variable import format: %s (expected <ORGANIZATION>/<WORKSPACE NAME|VARIABLE SET ID>/<VARIABLE ID>)", req.ID),
		)
		return
	}
	organization := s[0]
	container := s[1]
	id := s[2]
	data := modelTFEVariable{
		ID:            types.StringValue(id),
		WorkspaceID:   types.StringNull(),
		VariableSetID: types.StringNull(),
	}

	varsetIDUsed := variableSetIDRegexp.MatchString(container)
	if varsetIDUsed {
		data.VariableSetID = types.StringValue(container)
	} else {
		workspaceID, err := fetchWorkspaceExternalIDV2(organization+"/"+container, r.config.ClientV2)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error importing variable",
				fmt.Sprintf("Couldn't retrieve workspace %s from organization %s: %s", container, organization, err.Error()),
			)
			return
		}
		data.WorkspaceID = types.StringValue(workspaceID)
	}

	diags := resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

type updateReadableValuePlanModifier struct{}

func (u *updateReadableValuePlanModifier) Description(ctx context.Context) string {
	return "The readable_value will match the value if sensitive is false, or be empty otherwise"
}

func (u *updateReadableValuePlanModifier) MarkdownDescription(ctx context.Context) string {
	return u.Description(ctx)
}

func (u *updateReadableValuePlanModifier) PlanModifyString(ctx context.Context, request planmodifier.StringRequest, response *planmodifier.StringResponse) {
	var valueWO types.String
	diags := request.Config.GetAttribute(ctx, path.Root("value_wo"), &valueWO)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
	if valueWO.IsNull() {
		var sensitive types.Bool
		diags := request.Plan.GetAttribute(ctx, path.Root("sensitive"), &sensitive)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}

		// If the variable is sensitive, unset the readable_value
		if sensitive.ValueBool() {
			response.PlanValue = types.StringNull()
			return
		}

		var actualValue types.String
		diags = request.Plan.GetAttribute(ctx, path.Root("value"), &actualValue)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}
		response.PlanValue = actualValue
	} else {
		// it is a write-only value, so unset any previously set readable_value
		response.PlanValue = types.StringValue("")
	}
}

// Compile-time interface check
var _ resource.Resource = &resourceTFEVariable{}
var _ resource.ResourceWithConfigure = &resourceTFEVariable{}
var _ resource.ResourceWithUpgradeState = &resourceTFEVariable{}
var _ resource.ResourceWithImportState = &resourceTFEVariable{}
var _ planmodifier.String = &updateReadableValuePlanModifier{}

// NewResourceVariable is a resource function for the framework provider.
func NewResourceVariable() resource.Resource {
	return &resourceTFEVariable{}
}
