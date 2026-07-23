// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	legacytfe "github.com/hashicorp/go-tfe"
	tfe "github.com/hashicorp/go-tfe/v2"
	"github.com/hashicorp/go-tfe/v2/api/models"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
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
)

var providerSetPathSegmentRegexp = regexp.MustCompile(`^[^/\s]+$`)

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
		Description: "Manages provider sets.",
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
				Description: "Whether the provider set applies globally.",
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
				Description: "The workspace IDs attached to the provider set.",
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
				Description: "The project IDs attached to the provider set.",
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
				Description: "Source address of the provider, e.g. registry.terraform.io/hashicorp/tfe.",
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
				Description: "The provider configuration managed by the provider set, expressed as a single HCL provider block",
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
				Description: "The provider configuration managed by the provider set, expressed as a single HCL provider block in write-only mode.",
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
				Description: "Version of the write-only provider configuration to trigger updates.",
				Validators: []validator.Int64{
					int64validator.ConflictsWith(path.MatchRoot("provider_config_hcl")),
					int64validator.AlsoRequires(path.MatchRoot("provider_config_hcl_wo")),
				},
			},
		},
	}
}

// providerSetDataFromEnvelope extracts Provider Set data from a v2 response
// envelope and rejects incomplete responses before callers dereference them.
func providerSetDataFromEnvelope(env models.ProviderSetsEnvelopeable) (models.ProviderSetsable, error) {
	if env == nil {
		return nil, errors.New("the API returned an empty provider set response")
	}

	providerSet := env.GetData()
	if providerSet == nil {
		return nil, errors.New("the API returned a provider set response without data")
	}
	if providerSet.GetId() == nil {
		return nil, errors.New("the API returned provider set data without an ID")
	}
	if providerSet.GetAttributes() == nil {
		return nil, errors.New("the API returned provider set data without attributes")
	}

	return providerSet, nil
}

func providerSetStringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func providerSetBoolValue(value *bool) bool {
	if value == nil {
		return false
	}
	return *value
}

func providerSetLegacyError(err error) error {
	if errors.Is(err, tfe.ErrNotFound) {
		return legacytfe.ErrResourceNotFound
	}
	return err
}

func providerSetOrganizationValidationError(organization string) error {
	if !providerSetPathSegmentRegexp.MatchString(organization) {
		return legacytfe.ErrInvalidOrg
	}
	return nil
}

func providerSetRequiredNameValidationError(name string) error {
	if name == "" {
		return legacytfe.ErrRequiredName
	}
	return providerSetNameValidationError(name)
}

func providerSetNameValidationError(name string) error {
	if !providerSetPathSegmentRegexp.MatchString(name) {
		return legacytfe.ErrInvalidName
	}
	return nil
}

func providerSetCreateValidationError(organization, name string, configurationHCL *string) error {
	if err := providerSetOrganizationValidationError(organization); err != nil {
		return err
	}
	if err := providerSetRequiredNameValidationError(name); err != nil {
		return err
	}
	if configurationHCL == nil || *configurationHCL == "" {
		return legacytfe.ErrRequiredConfigurationHcl
	}
	return nil
}

func providerSetProjectIDs(v models.ProviderSetsable) []string {
	if v.GetRelationships() == nil || v.GetRelationships().GetProjects() == nil {
		return nil
	}

	projects := v.GetRelationships().GetProjects().GetData()
	projectIDs := make([]string, 0, len(projects))
	for _, project := range projects {
		if project == nil {
			continue
		}
		id := project.GetId()
		if id == nil {
			continue
		}
		projectIDs = append(projectIDs, *id)
	}
	return projectIDs
}

func providerSetWorkspaceIDs(v models.ProviderSetsable) []string {
	if v.GetRelationships() == nil || v.GetRelationships().GetWorkspaces() == nil {
		return nil
	}

	workspaces := v.GetRelationships().GetWorkspaces().GetData()
	workspaceIDs := make([]string, 0, len(workspaces))
	for _, workspace := range workspaces {
		if workspace == nil {
			continue
		}
		id := workspace.GetId()
		if id == nil {
			continue
		}
		workspaceIDs = append(workspaceIDs, *id)
	}
	return workspaceIDs
}

// modelFromTFEProviderSet builds a Terraform model from v2 Provider Set data.
func modelFromTFEProviderSet(
	ctx context.Context,
	v models.ProviderSetsable,
	organization string,
	providerConfigHCLWOVersion types.Int64,
) (m modelTFEProviderSet, diags diag.Diagnostics) {
	attributes := v.GetAttributes()

	// Initialize all fields from the provided API struct
	m = modelTFEProviderSet{
		ID:             types.StringValue(*v.GetId()),
		Name:           types.StringValue(providerSetStringValue(attributes.GetName())),
		Description:    types.StringValue(providerSetStringValue(attributes.GetDescription())),
		Global:         types.BoolValue(providerSetBoolValue(attributes.GetGlobal())),
		Priority:       types.BoolValue(providerSetBoolValue(attributes.GetPriority())),
		Organization:   types.StringValue(organization),
		ProviderSource: types.StringValue(providerSetStringValue(attributes.GetProviderSource())),
		ProjectIDs:     types.SetNull(types.StringType),
		WorkspaceIDs:   types.SetNull(types.StringType),
	}

	if !providerConfigHCLWOVersion.IsNull() {
		m.ProviderConfigHCL = types.StringNull()
		m.ProviderConfigHCLWOVersion = providerConfigHCLWOVersion
	} else {
		m.ProviderConfigHCL = types.StringValue(providerSetStringValue(attributes.GetConfigurationHcl()))
		m.ProviderConfigHCLWOVersion = types.Int64Null()
	}

	projectIDs := providerSetProjectIDs(v)

	var d diag.Diagnostics

	if len(projectIDs) > 0 {
		m.ProjectIDs, d = types.SetValueFrom(ctx, types.StringType, projectIDs)
		diags.Append(d...)
	}

	workspaceIDs := providerSetWorkspaceIDs(v)
	if len(workspaceIDs) > 0 {
		m.WorkspaceIDs, d = types.SetValueFrom(ctx, types.StringType, workspaceIDs)
		diags.Append(d...)
	}
	return m, diags
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

func providerSetRelationshipsFromModel(m modelTFEProviderSet) *models.ProviderSets_relationships {
	relationships := models.NewProviderSets_relationships()

	projects := models.NewProviderSets_relationships_projects()
	projectData := make([]models.ProviderSets_relationships_projects_dataable, 0)
	if !m.ProjectIDs.IsNull() && !m.ProjectIDs.IsUnknown() {
		projectData = make([]models.ProviderSets_relationships_projects_dataable, 0, len(m.ProjectIDs.Elements()))
		projectType := models.PROJECTS_PROVIDERSETS_RELATIONSHIPS_PROJECTS_DATA_TYPE
		for _, value := range m.ProjectIDs.Elements() {
			project := models.NewProviderSets_relationships_projects_data()
			project.SetId(value.(types.String).ValueStringPointer())
			project.SetTypeEscaped(&projectType)
			projectData = append(projectData, project)
		}
	}
	projects.SetData(projectData)
	relationships.SetProjects(projects)

	workspaces := models.NewProviderSets_relationships_workspaces()
	workspaceData := make([]models.ProviderSets_relationships_workspaces_dataable, 0)
	if !m.WorkspaceIDs.IsNull() && !m.WorkspaceIDs.IsUnknown() {
		workspaceData = make([]models.ProviderSets_relationships_workspaces_dataable, 0, len(m.WorkspaceIDs.Elements()))
		workspaceType := models.WORKSPACES_PROVIDERSETS_RELATIONSHIPS_WORKSPACES_DATA_TYPE
		for _, value := range m.WorkspaceIDs.Elements() {
			workspace := models.NewProviderSets_relationships_workspaces_data()
			workspace.SetId(value.(types.String).ValueStringPointer())
			workspace.SetTypeEscaped(&workspaceType)
			workspaceData = append(workspaceData, workspace)
		}
	}
	workspaces.SetData(workspaceData)
	relationships.SetWorkspaces(workspaces)

	return relationships
}

func newProviderSetEnvelope(m modelTFEProviderSet, configurationHCL *string) *models.ProviderSetsEnvelope {
	attributes := models.NewProviderSets_attributes()
	attributes.SetName(m.Name.ValueStringPointer())
	attributes.SetDescription(m.Description.ValueStringPointer())
	attributes.SetProviderSource(m.ProviderSource.ValueStringPointer())
	attributes.SetConfigurationHcl(configurationHCL)
	attributes.SetGlobal(m.Global.ValueBoolPointer())
	// Always set priority, including false, so PATCH requests do not omit it.
	attributes.SetPriority(m.Priority.ValueBoolPointer())

	providerSet := models.NewProviderSets()
	providerSetType := models.PROVIDERSETS_PROVIDERSETS_TYPE
	providerSet.SetTypeEscaped(&providerSetType)
	providerSet.SetAttributes(attributes)
	providerSet.SetRelationships(providerSetRelationshipsFromModel(m))

	envelope := models.NewProviderSetsEnvelope()
	envelope.SetData(providerSet)
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

	configurationHCL := plan.ProviderConfigHCL.ValueStringPointer()

	if !config.ProviderConfigHCLWO.IsNull() {
		tflog.Debug(
			ctx,
			fmt.Sprintf(
				"Using write-only HCL value for provider set %s",
				plan.Name.ValueString(),
			),
		)
		configurationHCL = config.ProviderConfigHCLWO.ValueStringPointer()
	}
	if err := providerSetCreateValidationError(orgName, plan.Name.ValueString(), configurationHCL); err != nil {
		resp.Diagnostics.AddError(
			"Error creating provider set",
			fmt.Sprintf("Couldn't create provider set %s: %s", plan.Name.ValueString(), err.Error()),
		)
		return
	}
	options := newProviderSetEnvelope(plan, configurationHCL)

	tflog.Debug(ctx,
		fmt.Sprintf(
			"Creating provider set with name: %s, organization: %s",
			plan.Name.ValueString(),
			orgName,
		))
	psEnvelope, err := r.config.ClientV2.API.Organizations().ByOrganization_name(orgName).ProviderSets().Post(ctx, options, nil)
	if err != nil {
		err = providerSetLegacyError(err)
		resp.Diagnostics.AddError(
			"Error creating provider set",
			fmt.Sprintf("Couldn't create provider set %s: %s", plan.Name.ValueString(), err.Error()),
		)
		return
	}
	ps, err := providerSetDataFromEnvelope(psEnvelope)
	if err != nil {
		resp.Diagnostics.AddError("Error creating provider set", err.Error())
		return
	}

	result, diags := modelFromTFEProviderSet(ctx, ps, orgName, config.ProviderConfigHCLWOVersion)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
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
	ps, err := providerSetDataFromEnvelope(psEnvelope)
	if err != nil {
		resp.Diagnostics.AddError("Error reading Provider set", err.Error())
		return
	}

	// update state
	result, diags := modelFromTFEProviderSet(ctx, ps, state.Organization.ValueString(), state.ProviderConfigHCLWOVersion)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
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
	if err := providerSetNameValidationError(plan.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError(
			"Error updating provider set",
			fmt.Sprintf("Couldn't update provider set %s: %s", plan.ID.String(), err.Error()),
		)
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

	configurationHCL := plan.ProviderConfigHCL.ValueStringPointer()

	if !config.ProviderConfigHCLWO.IsNull() {
		tflog.Debug(
			ctx,
			fmt.Sprintf(
				"Using write-only HCL value for provider set %s",
				plan.Name.ValueString(),
			),
		)
		configurationHCL = config.ProviderConfigHCLWO.ValueStringPointer()
	}
	options := newProviderSetEnvelope(plan, configurationHCL)

	tflog.Debug(ctx,
		fmt.Sprintf(
			"Updating provider set %s",
			plan.ID.String(),
		))
	psEnvelope, err := r.config.ClientV2.API.ProviderSets().ByProvider_set_id(plan.ID.ValueString()).Patch(ctx, options, nil)
	if err != nil {
		err = providerSetLegacyError(err)
		resp.Diagnostics.AddError(
			"Error updating provider set",
			fmt.Sprintf("Couldn't update provider set %s: %s", plan.ID.String(), err.Error()),
		)
		return
	}
	ps, err := providerSetDataFromEnvelope(psEnvelope)
	if err != nil {
		resp.Diagnostics.AddError("Error updating provider set", err.Error())
		return
	}

	result, diags := modelFromTFEProviderSet(ctx, ps, orgName, config.ProviderConfigHCLWOVersion)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
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
