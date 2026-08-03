// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	tfe "github.com/hashicorp/go-tfe/v2"
	"github.com/hashicorp/go-tfe/v2/api/models"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ resource.Resource               = &resourceTFEProviderSet{}
	_ resource.ResourceWithConfigure  = &resourceTFEProviderSet{}
	_ resource.ResourceWithModifyPlan = &resourceTFEProviderSet{}
	_ resource.ResourceWithIdentity   = &resourceTFEProviderSet{}
)

func NewProviderSetResource() resource.Resource {
	return &resourceTFEProviderSet{}
}

type resourceTFEProviderSet struct {
	config ConfiguredClient
}

type modelTFEProviderSet struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	Global         types.Bool   `tfsdk:"global"`
	Priority       types.Bool   `tfsdk:"priority"`
	Organization   types.String `tfsdk:"organization"`
	WorkspaceIDs   types.Set    `tfsdk:"workspace_ids"`
	ProjectIDs     types.Set    `tfsdk:"project_ids"`
	ProviderSource types.String `tfsdk:"provider_source"`
	// ProviderConfigHCL is the normal Terraform-managed mode and is persisted in state.
	ProviderConfigHCL types.String `tfsdk:"provider_config_hcl"`
	// ProviderConfigHCLWO is the write-only mode for HCL that should not be retained in state.
	ProviderConfigHCLWO types.String `tfsdk:"provider_config_hcl_wo"`
	// ProviderConfigHCLWOVersion is the explicit update trigger for write-only HCL.
	ProviderConfigHCLWOVersion types.Int64 `tfsdk:"provider_config_hcl_wo_version"`
}

type TFEProviderSetIdentityModel struct {
	ID types.String `tfsdk:"id"`
}

// collectionLengthOptions returns the options for determining the length of
// collections in the model, treating null and unknown as zero to simplify
// validation logic.
func (r modelTFEProviderSet) collectionLengthOptions() basetypes.CollectionLengthOptions {
	return basetypes.CollectionLengthOptions{
		UnhandledNullAsZero:    true,
		UnhandledUnknownAsZero: true,
	}
}

// Metadata returns the terraform resource type name.
func (r *resourceTFEProviderSet) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_provider_set"
}

// Schema defines the schema for the resource.
func (r *resourceTFEProviderSet) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Creates, updates and destroys provider sets." +
			"\n\n~> **Warning:** This resource is currently in beta and isn't generally available to all users. It is subject to change or be removed.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "The name of the provider set.",
				Required:    true,
			},
			"id": schema.StringAttribute{
				Description: "The ID of the provider set.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				Description: "The description of the provider set.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"global": schema.BoolAttribute{
				Description: "Whether the provider set applies globally. Defaults to `false`. When `global` is `false` or omitted, at least one of `workspace_ids` or `project_ids` must be set.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"priority": schema.BoolAttribute{
				Description: "Whether the provider set takes priority over provider sets with more specific scopes. Defaults to false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"organization": schema.StringAttribute{
				Description: "Name of the organization. If omitted, organization must be defined in the provider config.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"workspace_ids": schema.SetAttribute{
				Description: "IDs of the workspaces attached to the provider set. Required if `global` is `false` or omitted and `project_ids` is not set. Cannot be set if `global` is `true`.",
				ElementType: types.StringType,
				Optional:    true,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
						stringvalidator.RegexMatches(
							workspaceIDRegexp,
							"must be a valid workspace ID (ws-<RANDOM STRING>)",
						),
					),
				},
			},
			"project_ids": schema.SetAttribute{
				Description: "IDs of the projects attached to the provider set. Required if `global` is `false` or omitted and `workspace_ids` is not set. Cannot be set if `global` is `true`.",
				ElementType: types.StringType,
				Optional:    true,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
						stringvalidator.RegexMatches(
							projectIDRegexp,
							"must be a valid project ID (prj-<RANDOM STRING>)",
						),
					),
				},
			},
			"provider_source": schema.StringAttribute{
				Description: "Source address of the provider, e.g. `registry.terraform.io/hashicorp/tfe`.",
				Required:    true,
				CustomType:  types.StringType,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-zA-Z0-9._-]+/[a-zA-Z0-9_-]+/[a-zA-Z0-9_-]+$`),
						"must be in the format 'hostname/namespace/type', e.g. 'registry.terraform.io/hashicorp/tfe'",
					),
				},
			},
			"provider_config_hcl": schema.StringAttribute{
				Description: "The provider configuration HCL stored in Terraform state.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					// The normal and write-only modes are mutually exclusive because they map to the same backend blob.
					stringvalidator.ConflictsWith(path.MatchRoot("provider_config_hcl_wo")),
					stringvalidator.AtLeastOneOf(
						path.MatchRoot("provider_config_hcl"),
						path.MatchRoot("provider_config_hcl_wo"),
					),
				},
			},
			"provider_config_hcl_wo": schema.StringAttribute{
				Description: "Provider configuration HCL that is never stored in state. Cannot be used with `provider_config_hcl`. Requires `provider_config_hcl_wo_version`.",
				Optional:    true,
				WriteOnly:   true,
				Sensitive:   true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("provider_config_hcl")),
					// Write-only values are not stored in state, so a separate version field is required to trigger updates.
					stringvalidator.AlsoRequires(path.MatchRoot("provider_config_hcl_wo_version")),
				},
			},
			"provider_config_hcl_wo_version": schema.Int64Attribute{
				Optional:    true,
				Description: "Version identifier required when `provider_config_hcl_wo` is set to trigger updates.",
				Validators: []validator.Int64{
					int64validator.ConflictsWith(path.MatchRoot("provider_config_hcl")),
					int64validator.AlsoRequires(path.MatchRoot("provider_config_hcl_wo")),
				},
			},
		},
	}
}

func (r *resourceTFEProviderSet) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"id": identityschema.StringAttribute{
				RequiredForImport: true,
			},
		},
	}
}

// modelFromTFEProviderSet builds a modelFromTFEProviderSet struct from a tfe.ProviderSet
func modelFromTFEProviderSet(
	ctx context.Context,
	v models.ProviderSetsable,
	providerConfigHCLWOVersion types.Int64,
) (m modelTFEProviderSet, diags diag.Diagnostics) {
	m = modelTFEProviderSet{
		ProjectIDs:   types.SetNull(types.StringType),
		WorkspaceIDs: types.SetNull(types.StringType),
	}

	if id := v.GetId(); id != nil {
		m.ID = types.StringValue(*id)
	}

	configurationHcl := setModelFromProviderSetAttributes(&m, v.GetAttributes())

	relationships := v.GetRelationships()

	m.Organization = types.StringValue(providerSetOrganizationID(relationships))

	if !providerConfigHCLWOVersion.IsNull() {
		m.ProviderConfigHCL = types.StringNull()
		m.ProviderConfigHCLWOVersion = providerConfigHCLWOVersion
	} else {
		m.ProviderConfigHCL = types.StringValue(configurationHcl)
		m.ProviderConfigHCLWOVersion = types.Int64Null()
	}

	if relationships == nil {
		return m, diags
	}

	var d diag.Diagnostics

	if projectsRel := relationships.GetProjects(); projectsRel != nil {
		projectData := projectsRel.GetData()
		projectIDs := make([]string, 0, len(projectData))
		for _, p := range projectData {
			if p == nil {
				continue
			}
			if id := p.GetId(); id != nil {
				projectIDs = append(projectIDs, *id)
			}
		}
		if len(projectIDs) > 0 {
			m.ProjectIDs, d = types.SetValueFrom(ctx, types.StringType, projectIDs)
			diags.Append(d...)
		}
	}

	if workspacesRel := relationships.GetWorkspaces(); workspacesRel != nil {
		workspaceData := workspacesRel.GetData()
		workspaceIDs := make([]string, 0, len(workspaceData))
		for _, w := range workspaceData {
			if w == nil {
				continue
			}
			if id := w.GetId(); id != nil {
				workspaceIDs = append(workspaceIDs, *id)
			}
		}
		if len(workspaceIDs) > 0 {
			m.WorkspaceIDs, d = types.SetValueFrom(ctx, types.StringType, workspaceIDs)
			diags.Append(d...)
		}
	}

	return m, diags
}

// setModelFromProviderSetAttributes populates m's attribute-derived fields
// from attrs and returns the configuration HCL, if any.
func setModelFromProviderSetAttributes(m *modelTFEProviderSet, attrs models.ProviderSets_attributesable) string {
	if attrs == nil {
		return ""
	}

	if name := attrs.GetName(); name != nil {
		m.Name = types.StringValue(*name)
	}

	description := ""
	if d := attrs.GetDescription(); d != nil {
		description = *d
	}
	m.Description = types.StringValue(description)

	var global bool
	if g := attrs.GetGlobal(); g != nil {
		global = *g
	}
	m.Global = types.BoolValue(global)

	m.Priority = providerSetPriorityValue(attrs.GetPriority())

	if source := attrs.GetProviderSource(); source != nil {
		m.ProviderSource = types.StringValue(*source)
	}

	configurationHcl := ""
	if hcl := attrs.GetConfigurationHcl(); hcl != nil {
		configurationHcl = *hcl
	}
	return configurationHcl
}

func providerSetPriorityValue(priority *bool) types.Bool {
	if priority == nil {
		return types.BoolValue(false)
	}
	return types.BoolValue(*priority)
}

// providerSetOrganizationID extracts the organization ID from a provider
// set's relationships, if present.
func providerSetOrganizationID(relationships models.ProviderSets_relationshipsable) string {
	if relationships == nil {
		return ""
	}
	orgRel := relationships.GetOrganization()
	if orgRel == nil {
		return ""
	}
	orgData := orgRel.GetData()
	if orgData == nil {
		return ""
	}
	id := orgData.GetId()
	if id == nil {
		return ""
	}
	return *id
}

// ModifyPlan enforces that provider sets are either global or scoped to at
// least one project or workspace. It also prevents implicitly detaching scopes
// when a scoped provider set is updated to global.
func (r *resourceTFEProviderSet) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}

	modifyPlanForDefaultOrganizationChange(
		ctx,
		r.config.Organization,
		req.State,
		req.Config,
		req.Plan,
		resp,
	)

	var plan modelTFEProviderSet
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.Global.IsUnknown() {
		return
	}

	workspaceCount := plan.WorkspaceIDs.Length(plan.collectionLengthOptions())
	projectCount := plan.ProjectIDs.Length(plan.collectionLengthOptions())

	if !plan.Global.ValueBool() {
		scopesUnknown := plan.WorkspaceIDs.IsUnknown() || plan.ProjectIDs.IsUnknown()
		if !scopesUnknown && workspaceCount == 0 && projectCount == 0 {
			resp.Diagnostics.AddAttributeError(
				path.Root("global"),
				"Invalid Attribute Combination",
				"global must be true unless workspace_ids or project_ids are set",
			)
		}
		return
	}

	// Validate workspace_ids is not set
	if workspaceCount > 0 {
		resp.Diagnostics.AddAttributeError(
			path.Root("workspace_ids"),
			"Invalid Attribute Combination",
			"workspace_ids cannot be set when global is true",
		)
	}
	// Validate project_ids is not set
	if projectCount > 0 {
		resp.Diagnostics.AddAttributeError(
			path.Root("project_ids"),
			"Invalid Attribute Combination",
			"project_ids cannot be set when global is true",
		)
	}
}

// Configure is used to provide the resource with the configured API client
// from the provider.
func (r *resourceTFEProviderSet) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(ConfiguredClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected resource Configure type",
			fmt.Sprintf("Expected tfe.ConfiguredClient, got %T. This is a bug in the tfe provider, so please report it on GitHub.", req.ProviderData),
		)
		return
	}

	r.config = client
}

// providerSetRelationshipIDs converts a Set of string IDs from the model into
// a plain string slice, always returning a non-nil (possibly empty) slice so
// that the resulting relationship is explicitly set to the plan's exact
// desired state, including clearing it.
func providerSetRelationshipIDs(s types.Set) []string {
	if s.IsNull() || s.IsUnknown() {
		return []string{}
	}
	elements := s.Elements()
	ids := make([]string, len(elements))
	for i, v := range elements {
		ids[i] = v.(types.String).ValueString()
	}
	return ids
}

// providerSetWorkspacesRelationship builds the workspaces relationship for a
// provider set create/update request from the given workspace IDs.
func providerSetWorkspacesRelationship(ids []string) models.WorkspacesHasManyable {
	wsType := models.WORKSPACES_WORKSPACESIDENTIFIER_TYPE
	data := make([]models.WorkspacesIdentifierable, len(ids))
	for i, id := range ids {
		d := models.NewWorkspacesIdentifier()
		d.SetId(&id)
		d.SetTypeEscaped(&wsType)
		data[i] = d
	}

	rel := models.NewWorkspacesHasMany()
	rel.SetData(data)
	return rel
}

// providerSetProjectsRelationship builds the projects relationship for a
// provider set create/update request from the given project IDs.
func providerSetProjectsRelationship(ids []string) models.ProjectsHasManyable {
	projType := models.PROJECTS_PROJECTSIDENTIFIER_TYPE
	data := make([]models.ProjectsIdentifierable, len(ids))
	for i, id := range ids {
		d := models.NewProjectsIdentifier()
		d.SetId(&id)
		d.SetTypeEscaped(&projType)
		data[i] = d
	}

	rel := models.NewProjectsHasMany()
	rel.SetData(data)
	return rel
}

// newProviderSetAttributes builds the attributes shared by provider set
// create and update requests.
func newProviderSetAttributes(name, description, providerSource, configurationHcl string, global, priority bool) *models.ProviderSets_attributes {
	attributes := models.NewProviderSets_attributes()
	attributes.SetName(&name)
	attributes.SetDescription(&description)
	attributes.SetProviderSource(&providerSource)
	attributes.SetConfigurationHcl(&configurationHcl)
	attributes.SetGlobal(&global)
	attributes.SetPriority(&priority)
	return attributes
}

// newProviderSetCreateEnvelope builds the request body for creating a
// provider set belonging to the given organization.
func newProviderSetCreateEnvelope(organization string, plan modelTFEProviderSet, configurationHcl string) *models.ProviderSetsEnvelope {
	attributes := newProviderSetAttributes(
		plan.Name.ValueString(),
		plan.Description.ValueString(),
		plan.ProviderSource.ValueString(),
		configurationHcl,
		plan.Global.ValueBool(),
		plan.Priority.ValueBool(),
	)

	orgType := models.ORGANIZATIONS_ORGANIZATIONSIDENTIFIER_TYPE
	orgData := models.NewOrganizationsIdentifier()
	orgData.SetId(&organization)
	orgData.SetTypeEscaped(&orgType)
	orgRel := models.NewOrganizationsHasOne()
	orgRel.SetData(orgData)

	relationships := models.NewProviderSets_relationships()
	relationships.SetOrganization(orgRel)
	relationships.SetWorkspaces(providerSetWorkspacesRelationship(providerSetRelationshipIDs(plan.WorkspaceIDs)))
	relationships.SetProjects(providerSetProjectsRelationship(providerSetRelationshipIDs(plan.ProjectIDs)))

	ps := models.NewProviderSets()
	ps.SetAttributes(attributes)
	ps.SetRelationships(relationships)
	psType := models.PROVIDERSETS_PROVIDERSETS_TYPE
	ps.SetTypeEscaped(&psType)

	envelope := models.NewProviderSetsEnvelope()
	envelope.SetData(ps)
	return envelope
}

// newProviderSetUpdateEnvelope builds the request body for updating an
// existing provider set. The organization relationship is immutable
// (organization changes force resource replacement) and is intentionally
// left unset.
func newProviderSetUpdateEnvelope(id string, plan modelTFEProviderSet, configurationHcl string) *models.ProviderSetsEnvelope {
	attributes := newProviderSetAttributes(
		plan.Name.ValueString(),
		plan.Description.ValueString(),
		plan.ProviderSource.ValueString(),
		configurationHcl,
		plan.Global.ValueBool(),
		plan.Priority.ValueBool(),
	)

	relationships := models.NewProviderSets_relationships()
	relationships.SetWorkspaces(providerSetWorkspacesRelationship(providerSetRelationshipIDs(plan.WorkspaceIDs)))
	relationships.SetProjects(providerSetProjectsRelationship(providerSetRelationshipIDs(plan.ProjectIDs)))

	ps := models.NewProviderSets()
	ps.SetId(&id)
	ps.SetAttributes(attributes)
	ps.SetRelationships(relationships)
	psType := models.PROVIDERSETS_PROVIDERSETS_TYPE
	ps.SetTypeEscaped(&psType)

	envelope := models.NewProviderSetsEnvelope()
	envelope.SetData(ps)
	return envelope
}

// Create handles the creation of the resource by making an API call to create a
// provider set with the specified configuration, and then storing the
// resulting state. It also handles the logic for the write-only HCL field,
// ensuring that if it is set, its value is used for the API call instead of
// the normal HCL field.
func (r *resourceTFEProviderSet) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan modelTFEProviderSet
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var config modelTFEProviderSet
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the organization name from resource or provider config
	var orgName string
	resp.Diagnostics.Append(r.config.dataOrDefaultOrganization(ctx, req.Config, &orgName)...)
	if resp.Diagnostics.HasError() {
		return
	}

	configurationHcl := plan.ProviderConfigHCL.ValueString()
	if !config.ProviderConfigHCLWO.IsNull() {
		tflog.Debug(
			ctx,
			fmt.Sprintf(
				"Using write-only HCL value for provider set %s",
				plan.Name.ValueString(),
			),
		)
		configurationHcl = config.ProviderConfigHCLWO.ValueString()
	}

	envelope := newProviderSetCreateEnvelope(orgName, plan, configurationHcl)

	tflog.Debug(ctx,
		fmt.Sprintf(
			"Creating provider set with name: %s, organization: %s",
			plan.Name.ValueString(),
			orgName,
		))
	psEnvelope, err := r.config.ClientV2.API.Organizations().ByOrganization_name(orgName).ProviderSets().Post(ctx, envelope, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating provider set",
			fmt.Sprintf("Couldn't create provider set %s: %s", plan.Name.ValueString(), err.Error()),
		)
		return
	}
	if psEnvelope == nil || psEnvelope.GetData() == nil {
		resp.Diagnostics.AddError(
			"Error creating provider set",
			fmt.Sprintf("Couldn't create provider set %s: no data was returned by the API", plan.Name.ValueString()),
		)
		return
	}

	result, diags := modelFromTFEProviderSet(ctx, psEnvelope.GetData(), config.ProviderConfigHCLWOVersion)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)

	identity := TFEProviderSetIdentityModel{
		ID: result.ID,
	}
	resp.Diagnostics.Append(resp.Identity.Set(ctx, identity)...)

}

// Read handles reading the resource state by making an API call to retrieve the
// current configuration of the provider set, and then updating the state with
// the retrieved information. If the provider set no longer exists, it removes
// the resource from state without error.
func (r *resourceTFEProviderSet) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state modelTFEProviderSet
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	providerSetID := state.ID.ValueString()
	psEnvelope, err := r.config.ClientV2.API.ProviderSets().ByProvider_set_id(providerSetID).Get(ctx, nil)
	if err != nil {
		// If it's gone: that's not an error, but we are done.
		if errors.Is(err, tfe.ErrNotFound) {
			tflog.Debug(
				ctx, fmt.Sprintf(
					"Provider Set %s no longer exists", providerSetID,
				),
			)
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError(
				"Error reading Provider set",
				fmt.Sprintf("Couldn't read provider set %s: %s", providerSetID, err.Error()),
			)
		}

		return
	}
	if psEnvelope == nil || psEnvelope.GetData() == nil {
		tflog.Debug(
			ctx, fmt.Sprintf(
				"Provider Set %s no longer exists", providerSetID,
			),
		)
		resp.State.RemoveResource(ctx)
		return
	}

	// update state
	result, diags := modelFromTFEProviderSet(ctx, psEnvelope.GetData(), state.ProviderConfigHCLWOVersion)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)

	identity := TFEProviderSetIdentityModel{
		ID: result.ID,
	}
	resp.Diagnostics.Append(resp.Identity.Set(ctx, identity)...)
}

// Update handles updating the resource by making an API call to update the
// provider set with the new configuration specified in the plan, and then
// updating the state with the resulting configuration. It also handles the
// logic for the write-only HCL field, ensuring that if it is set in the plan,
// its value is used for the API call instead of the normal HCL field, and that
// the version is updated to trigger the update.
func (r *resourceTFEProviderSet) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan modelTFEProviderSet
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var config modelTFEProviderSet
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	configurationHcl := plan.ProviderConfigHCL.ValueString()
	if !config.ProviderConfigHCLWO.IsNull() {
		tflog.Debug(
			ctx,
			fmt.Sprintf(
				"Using write-only HCL value for provider set %s",
				plan.Name.ValueString(),
			),
		)
		configurationHcl = config.ProviderConfigHCLWO.ValueString()
	}

	envelope := newProviderSetUpdateEnvelope(plan.ID.ValueString(), plan, configurationHcl)

	tflog.Debug(ctx,
		fmt.Sprintf(
			"Updating provider set %s",
			plan.ID.String(),
		))
	psEnvelope, err := r.config.ClientV2.API.ProviderSets().ByProvider_set_id(plan.ID.ValueString()).Patch(ctx, envelope, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating provider set",
			fmt.Sprintf("Couldn't update provider set %s: %s", plan.ID.String(), err.Error()),
		)
		return
	}
	if psEnvelope == nil || psEnvelope.GetData() == nil {
		resp.Diagnostics.AddError(
			"Error updating provider set",
			fmt.Sprintf("Couldn't update provider set %s: no data was returned by the API", plan.ID.String()),
		)
		return
	}

	result, diags := modelFromTFEProviderSet(ctx, psEnvelope.GetData(), config.ProviderConfigHCLWOVersion)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)

	identity := TFEProviderSetIdentityModel{
		ID: result.ID,
	}
	resp.Diagnostics.Append(resp.Identity.Set(ctx, identity)...)
}

// Delete handles deleting the resource by making an API call to delete the
// provider set. If the provider set is already gone, it treats that as a
// success and removes the resource from state without error.
func (r *resourceTFEProviderSet) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state modelTFEProviderSet
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	providerSetID := state.ID.ValueString()

	tflog.Debug(ctx, fmt.Sprintf("Delete provider set: %s", providerSetID))
	err := r.config.ClientV2.API.ProviderSets().ByProvider_set_id(providerSetID).Delete(ctx, nil)
	// Ignore 404s for delete
	if err != nil && !errors.Is(err, tfe.ErrNotFound) {
		resp.Diagnostics.AddError(
			"Error deleting provider set",
			fmt.Sprintf("Couldn't delete provider set %s: %s", providerSetID, err.Error()),
		)
	}
	// Resource is implicitly deleted from resp.State if diagnostics have no errors.
}
