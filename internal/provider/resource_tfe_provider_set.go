// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	tfe "github.com/hashicorp/go-tfe"
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

// modelFromTFEProviderSet builds a modelFromTFEProviderSet struct from a tfe.ProviderSet
func modelFromTFEProviderSet(
	ctx context.Context,
	v tfe.ProviderSet,
	providerConfigHCLWOVersion types.Int64,
) (m modelTFEProviderSet, diags diag.Diagnostics) {
	// Initialize all fields from the provided API struct
	m = modelTFEProviderSet{
		ID:             types.StringValue(v.ID),
		Name:           types.StringValue(v.Name),
		Description:    types.StringValue(v.Description),
		Global:         types.BoolValue(v.Global),
		Organization:   types.StringValue(v.Organization.Name),
		ProviderSource: types.StringValue(v.ProviderSource),
		ProjectIDs:     types.SetNull(types.StringType),
		WorkspaceIDs:   types.SetNull(types.StringType),
	}

	if !providerConfigHCLWOVersion.IsNull() {
		m.ProviderConfigHCL = types.StringNull()
		m.ProviderConfigHCLWOVersion = providerConfigHCLWOVersion
	} else {
		m.ProviderConfigHCL = types.StringValue(v.ConfigurationHcl)
		m.ProviderConfigHCLWOVersion = types.Int64Null()
	}

	projectIDs := make([]string, len(v.Projects))
	for i, project := range v.Projects {
		projectIDs[i] = project.ID
	}

	var d diag.Diagnostics

	if len(projectIDs) > 0 {
		m.ProjectIDs, d = types.SetValueFrom(ctx, types.StringType, projectIDs)
		diags.Append(d...)
	}

	workspaceIDs := make([]string, len(v.Workspaces))
	for i, workspace := range v.Workspaces {
		workspaceIDs[i] = workspace.ID
	}
	if len(workspaceIDs) > 0 {
		m.WorkspaceIDs, d = types.SetValueFrom(ctx, types.StringType, workspaceIDs)
		diags.Append(d...)
	}
	return m, diags
}

// ModifyPlan is used to enforce that when global is updated from false to
// true, workspace_ids and project_ids must be empty, since global provider
// sets cannot have workspaces or projects attached. This is necessary
// because the API will automatically detach all workspaces and projects when
// a provider set is updated to global, which would be unexpected and
// potentially destructive behavior for users if it were allowed to happen
// implicitly.
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

	if !plan.Global.ValueBool() {
		return
	}

	// Validate workspace_ids is not set
	if plan.WorkspaceIDs.Length(plan.collectionLengthOptions()) > 0 {
		resp.Diagnostics.AddAttributeError(
			path.Root("workspace_ids"),
			"Invalid Attribute Combination",
			"workspace_ids cannot be set when global is true",
		)
	}
	// Validate project_ids is not set
	if plan.ProjectIDs.Length(plan.collectionLengthOptions()) > 0 {
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

// tfeWorkspacesFromModel converts the workspace IDs in the model to a slice of
// tfe.Workspace pointers for API calls.
func tfeWorkspacesFromModel(m modelTFEProviderSet) []*tfe.Workspace {
	if m.WorkspaceIDs.IsNull() || m.WorkspaceIDs.IsUnknown() {
		return []*tfe.Workspace{}
	}
	workspaces := make([]*tfe.Workspace, m.WorkspaceIDs.Length(m.collectionLengthOptions()))

	for i, v := range m.WorkspaceIDs.Elements() {
		workspaces[i] = &tfe.Workspace{ID: v.(types.String).ValueString()}
	}
	return workspaces
}

// tfeProjectsFromModel converts the project IDs in the model to a slice of
// tfe.Project pointers for API calls.
func tfeProjectsFromModel(m modelTFEProviderSet) []*tfe.Project {
	if m.ProjectIDs.IsNull() || m.ProjectIDs.IsUnknown() {
		return []*tfe.Project{}
	}
	projects := make([]*tfe.Project, m.ProjectIDs.Length(m.collectionLengthOptions()))

	for i, v := range m.ProjectIDs.Elements() {
		projects[i] = &tfe.Project{ID: v.(types.String).ValueString()}
	}
	return projects
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

	options := tfe.ProviderSetCreateOptions{
		Name:             plan.Name.ValueString(),
		Description:      plan.Description.ValueStringPointer(),
		ProviderSource:   plan.ProviderSource.ValueString(),
		Global:           plan.Global.ValueBoolPointer(),
		Workspaces:       tfeWorkspacesFromModel(plan),
		Projects:         tfeProjectsFromModel(plan),
		ConfigurationHcl: plan.ProviderConfigHCL.ValueString(),
	}

	if !config.ProviderConfigHCLWO.IsNull() {
		tflog.Debug(
			ctx,
			fmt.Sprintf(
				"Using write-only HCL value for provider set %s",
				plan.Name.ValueString(),
			),
		)
		options.ConfigurationHcl = config.ProviderConfigHCLWO.ValueString()
	}

	tflog.Debug(ctx,
		fmt.Sprintf(
			"Creating provider set with name: %s, organization: %s",
			options.Name,
			orgName,
		))
	ps, err := r.config.Client.ProviderSets.Create(ctx, orgName, options)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating provider set",
			fmt.Sprintf("Couldn't create provider set %s: %s", options.Name, err.Error()),
		)
		return
	}

	result, diags := modelFromTFEProviderSet(ctx, *ps, config.ProviderConfigHCLWOVersion)
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
	ps, err := r.config.Client.ProviderSets.Read(ctx, providerSetID)
	if err != nil {
		// If it's gone: that's not an error, but we are done.
		if errors.Is(err, tfe.ErrResourceNotFound) {
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

	// update state
	result, diags := modelFromTFEProviderSet(ctx, *ps, state.ProviderConfigHCLWOVersion)
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

	options := tfe.ProviderSetUpdateOptions{
		Name:             plan.Name.ValueStringPointer(),
		Description:      plan.Description.ValueStringPointer(),
		ProviderSource:   plan.ProviderSource.ValueStringPointer(),
		Global:           plan.Global.ValueBoolPointer(),
		Workspaces:       tfeWorkspacesFromModel(plan),
		Projects:         tfeProjectsFromModel(plan),
		ConfigurationHcl: plan.ProviderConfigHCL.ValueStringPointer(),
	}

	if !config.ProviderConfigHCLWO.IsNull() {
		tflog.Debug(
			ctx,
			fmt.Sprintf(
				"Using write-only HCL value for provider set %s",
				plan.Name.ValueString(),
			),
		)
		options.ConfigurationHcl = config.ProviderConfigHCLWO.ValueStringPointer()
	}

	tflog.Debug(ctx,
		fmt.Sprintf(
			"Updating provider set %s",
			plan.ID.String(),
		))
	ps, err := r.config.Client.ProviderSets.Update(ctx, plan.ID.ValueString(), options)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating provider set",
			fmt.Sprintf("Couldn't update provider set %s: %s", plan.ID.String(), err.Error()),
		)
		return
	}

	result, diags := modelFromTFEProviderSet(ctx, *ps, config.ProviderConfigHCLWOVersion)
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
	err := r.config.Client.ProviderSets.Delete(ctx, providerSetID)
	// Ignore 404s for delete
	if err != nil && !errors.Is(err, tfe.ErrResourceNotFound) {
		resp.Diagnostics.AddError(
			"Error deleting provider set",
			fmt.Sprintf("Couldn't delete provider set %s: %s", providerSetID, err.Error()),
		)
	}
	// Resource is implicitly deleted from resp.State if diagnostics have no errors.
}
