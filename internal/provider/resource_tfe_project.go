// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	tfe "github.com/hashicorp/go-tfe"
	tfev2 "github.com/hashicorp/go-tfe/v2"
	"github.com/hashicorp/go-tfe/v2/api/models"
	"github.com/hashicorp/jsonapi"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &resourceTFEProject{}
	_ resource.ResourceWithConfigure   = &resourceTFEProject{}
	_ resource.ResourceWithImportState = &resourceTFEProject{}
	_ resource.ResourceWithModifyPlan  = &resourceTFEProject{}
)

func NewProjectResource() resource.Resource {
	return &resourceTFEProject{}
}

type resourceTFEProject struct {
	config ConfiguredClient
}

var projectIDRegexp = regexp.MustCompile("^prj-[a-zA-Z0-9]{16}$")

// modelTFEProject maps the resource schema data to a struct.
type modelTFEProject struct {
	ID                          types.String `tfsdk:"id"`
	Name                        types.String `tfsdk:"name"`
	Description                 types.String `tfsdk:"description"`
	Organization                types.String `tfsdk:"organization"`
	AutoDestroyActivityDuration types.String `tfsdk:"auto_destroy_activity_duration"`
	Tags                        types.Map    `tfsdk:"tags"`
	IgnoreAdditionalTags        types.Bool   `tfsdk:"ignore_additional_tags"`
}

type modelProjectIdentity struct {
	ID       types.String `tfsdk:"id"`
	Hostname types.String `tfsdk:"hostname"`
}

// modelFromTFEProject builds a modelTFEProject struct from a v2 project resource. tags is a plain
// key/value map since its two callers source it differently: Create/Update echo back the tags
// just sent (trusting local input), while Read sources it from the server's effective tag
// bindings via GET /projects/{id}/effective-tag-bindings.
func modelFromTFEProject(p models.Projectsable, tags map[string]string, ignoreAdditionalTags types.Bool) modelTFEProject {
	model := modelTFEProject{
		ID:                   types.StringValue(valueOrZero(p.GetId())),
		Organization:         types.StringValue(projectOrganizationID(p.GetRelationships())),
		IgnoreAdditionalTags: ignoreAdditionalTags,
	}

	if attrs := p.GetAttributes(); attrs != nil {
		model.Name = types.StringValue(valueOrZero(attrs.GetName()))
		model.Description = types.StringValue(valueOrZero(attrs.GetDescription()))
		if duration := attrs.GetAutoDestroyActivityDuration(); duration != nil {
			model.AutoDestroyActivityDuration = types.StringValue(*duration)
		}
	}

	tagElems := make(map[string]attr.Value, len(tags))
	for key, value := range tags {
		tagElems[key] = types.StringValue(value)
	}
	model.Tags = types.MapValueMust(types.StringType, tagElems)

	return model
}

// projectPlanTagBindings extracts the configured tag key/value pairs from a plan's tags map.
func projectPlanTagBindings(tags types.Map) map[string]string {
	bindings := make(map[string]string)
	for key, val := range tags.Elements() {
		if strVal, ok := val.(types.String); ok && !strVal.IsNull() {
			bindings[key] = strVal.ValueString()
		}
	}
	return bindings
}

// newTagBindingsCollection builds the request body for replacing all of a project's direct tag
// bindings via the dedicated /projects/{id}/relationships/tag-bindings endpoint. go-tfe/v2 has no
// way to embed tag bindings in the same request as the project attributes update, unlike v1.
func newTagBindingsCollection(bindings map[string]string) *models.TagBindingsCollection {
	data := make([]models.TagBindingsable, 0, len(bindings))
	for key, value := range bindings {
		attributes := models.NewTagBindings_attributes()
		attributes.SetKey(&key)
		attributes.SetValue(&value)

		tb := models.NewTagBindings()
		tb.SetAttributes(attributes)
		tbType := models.TAGBINDINGS_TAGBINDINGS_TYPE
		tb.SetTypeEscaped(&tbType)

		data = append(data, tb)
	}

	collection := models.NewTagBindingsCollection()
	collection.SetData(data)
	return collection
}

// newProjectAttributes builds the attributes shared by the create and update envelopes.
func newProjectAttributes(name, description string, autoDestroyActivityDuration *string) *models.Projects_attributes {
	attributes := models.NewProjects_attributes()
	attributes.SetName(&name)
	attributes.SetDescription(&description)
	attributes.SetAutoDestroyActivityDuration(autoDestroyActivityDuration)
	return attributes
}

// newProjectCreateEnvelope builds the request body for creating a project.
func newProjectCreateEnvelope(name, description string, autoDestroyActivityDuration *string) *models.ProjectsEnvelope {
	attributes := newProjectAttributes(name, description, autoDestroyActivityDuration)

	data := models.NewProjects()
	data.SetAttributes(attributes)
	projType := models.PROJECTS_PROJECTS_TYPE
	data.SetTypeEscaped(&projType)

	envelope := models.NewProjectsEnvelope()
	envelope.SetData(data)
	return envelope
}

// newProjectUpdateEnvelope builds the request body for updating an existing project.
func newProjectUpdateEnvelope(id, name, description string, autoDestroyActivityDuration *string) *models.ProjectsEnvelope {
	attributes := newProjectAttributes(name, description, autoDestroyActivityDuration)

	data := models.NewProjects()
	data.SetId(&id)
	data.SetAttributes(attributes)
	projType := models.PROJECTS_PROJECTS_TYPE
	data.SetTypeEscaped(&projType)

	envelope := models.NewProjectsEnvelope()
	envelope.SetData(data)
	return envelope
}

// Configure implements resource.ResourceWithConfigure
func (r *resourceTFEProject) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *resourceTFEProject) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

// Schema implements resource.Resource
func (r *resourceTFEProject) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a project.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "ID for the project.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"name": schema.StringAttribute{
				Description: "Name of the project. TFE versions v202404-2 and earlier support between 3-36 characters. TFE versions v202405-1 and later support between 3-40 characters.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 40),
					stringvalidator.RegexMatches(regexp.MustCompile(`\A[\w\-][\w\- ]+[\w\-]\z`),
						"can only include letters, numbers, spaces, -, and _.",
					),
				},
			},

			"description": schema.StringAttribute{
				Description: "Description of the project.",
				Optional:    true,
				Computed:    true,
			},

			"organization": schema.StringAttribute{
				Description: "Name of the organization. If omitted, organization must be defined in the provider config.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			"auto_destroy_activity_duration": schema.StringAttribute{
				Description: "A duration string for all workspaces in the project, representing time after each workspace's activity when an auto-destroy run will be triggered.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^\d{1,4}[dh]$`),
						"must be 1-4 digits followed by 'd' or 'h'.",
					),
				},
			},

			"tags": schema.MapAttribute{
				Description: "A map of key-value tags to add to the project.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
			},

			"ignore_additional_tags": schema.BoolAttribute{
				Description: "Explicitly ignores `tags` not defined by config so they will not be overwritten by the configured tags. This creates exceptional behaviour in Terraform with respect to `tags` and is not recommended. This value must be applied before it will be used.",
				Optional:    true,
			},
		},
	}
}

func (r *resourceTFEProject) IdentitySchema(ctx context.Context, req resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"id": identityschema.StringAttribute{
				RequiredForImport: true,
			},
			"hostname": identityschema.StringAttribute{
				OptionalForImport: true,
			},
		},
	}
}

// Create implements resource.Resource
func (r *resourceTFEProject) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan modelTFEProject
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the organization name from resource or provider config
	var orgName string
	resp.Diagnostics.Append(r.config.dataOrDefaultOrganization(ctx, req.Config, &orgName)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()

	var autoDestroy *string
	if !plan.AutoDestroyActivityDuration.IsNull() {
		v := plan.AutoDestroyActivityDuration.ValueString()
		autoDestroy = &v
	}

	tagBindings := projectPlanTagBindings(plan.Tags)

	envelope := newProjectCreateEnvelope(name, plan.Description.ValueString(), autoDestroy)

	tflog.Debug(ctx, fmt.Sprintf("Create project %s", name))
	projEnvelope, err := r.config.ClientV2.API.Organizations().ByOrganization_name(orgName).Projects().Post(ctx, envelope, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error creating project", err.Error())
		return
	}
	if projEnvelope == nil || projEnvelope.GetData() == nil {
		resp.Diagnostics.AddError("Error creating project", "no data was returned by the API")
		return
	}
	projectData := projEnvelope.GetData()

	if len(tagBindings) > 0 {
		projectID := valueOrZero(projectData.GetId())
		collection := newTagBindingsCollection(tagBindings)
		if err := r.config.ClientV2.API.Projects().ByProject_id(projectID).Relationships().TagBindings().Patch(ctx, collection, nil); err != nil {
			resp.Diagnostics.AddError("Error setting tag bindings on project", err.Error())
			return
		}
	}

	result := modelFromTFEProject(projectData, tagBindings, plan.IgnoreAdditionalTags)

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)

	identity := modelProjectIdentity{
		ID:       result.ID,
		Hostname: types.StringValue(r.config.ClientV2.BaseURL().Host),
	}
	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identity)...)
}

// Read implements resource.Resource
func (r *resourceTFEProject) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state modelTFEProject
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	tflog.Debug(ctx, fmt.Sprintf("Read project %s", id))
	projEnvelope, err := r.config.ClientV2.API.Projects().ByProject_id(id).Get(ctx, nil)
	if err != nil {
		if errors.Is(err, tfev2.ErrNotFound) {
			tflog.Debug(ctx, fmt.Sprintf("Project %s no longer exists", id))
			r.setReadIdentity(ctx, req, resp, id)
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading project", err.Error())
		return
	}
	if projEnvelope == nil || projEnvelope.GetData() == nil {
		tflog.Debug(ctx, fmt.Sprintf("Project %s no longer exists", id))
		r.setReadIdentity(ctx, req, resp, id)
		resp.State.RemoveResource(ctx)
		return
	}
	projectData := projEnvelope.GetData()

	bindingsColl, err := r.config.ClientV2.API.Projects().ByProject_id(id).EffectiveTagBindings().Get(ctx, nil)
	if err != nil && !errors.Is(err, tfev2.ErrNotFound) {
		resp.Diagnostics.AddError("Error reading project", err.Error())
		return
	}

	tagBindings := make(map[string]string)
	if bindingsColl != nil {
		for _, binding := range bindingsColl.GetData() {
			if binding == nil || binding.GetAttributes() == nil {
				continue
			}
			tagBindings[valueOrZero(binding.GetAttributes().GetKey())] = valueOrZero(binding.GetAttributes().GetValue())
		}
	}

	if state.IgnoreAdditionalTags.ValueBool() {
		currentTags := state.Tags.Elements()
		for key := range tagBindings {
			if _, ok := currentTags[key]; !ok {
				delete(tagBindings, key)
			}
		}
	}

	result := modelFromTFEProject(projectData, tagBindings, state.IgnoreAdditionalTags)

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)

	identity := modelProjectIdentity{
		ID:       result.ID,
		Hostname: types.StringValue(r.config.ClientV2.BaseURL().Host),
	}
	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identity)...)
}

func (r *resourceTFEProject) setReadIdentity(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse, projectID string) {
	if resp.Identity == nil {
		return
	}

	if req.Identity != nil {
		currentIdentity := &modelProjectIdentity{}
		resp.Diagnostics.Append(req.Identity.Get(ctx, &currentIdentity)...)
		if resp.Diagnostics.HasError() {
			return
		}
		if currentIdentity != nil && !currentIdentity.ID.IsNull() {
			return
		}
	}

	identity := modelProjectIdentity{
		ID:       types.StringValue(projectID),
		Hostname: types.StringValue(r.config.ClientV2.BaseURL().Host),
	}
	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identity)...)
}

// clearProjectAutoDestroyV1 updates a project's name and description, and explicitly nulls out
// auto_destroy_activity_duration, via the v1 client. See the comment in Update for why this one
// case can't go through go-tfe/v2.
func (r *resourceTFEProject) clearProjectAutoDestroyV1(ctx context.Context, id, name, description string) error {
	v1Options := tfe.ProjectUpdateOptions{
		Name:                        &name,
		Description:                 &description,
		AutoDestroyActivityDuration: jsonapi.NewNullNullableAttr[string](),
	}
	_, err := r.config.Client.Projects.Update(ctx, id, v1Options)
	return err
}

// updateProjectV2 updates a project's name, description, and auto_destroy_activity_duration via
// the v2 client.
func (r *resourceTFEProject) updateProjectV2(ctx context.Context, id, name, description string, autoDestroyActivityDuration types.String) (models.Projectsable, error) {
	var autoDestroy *string
	if !autoDestroyActivityDuration.IsNull() {
		v := autoDestroyActivityDuration.ValueString()
		autoDestroy = &v
	}

	envelope := newProjectUpdateEnvelope(id, name, description, autoDestroy)

	projEnvelope, err := r.config.ClientV2.API.Projects().ByProject_id(id).Patch(ctx, envelope, nil)
	if err != nil {
		return nil, err
	}
	if projEnvelope == nil || projEnvelope.GetData() == nil {
		return nil, errors.New("no data was returned by the API")
	}
	return projEnvelope.GetData(), nil
}

// Update implements resource.Resource
func (r *resourceTFEProject) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan modelTFEProject
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state modelTFEProject
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	name := plan.Name.ValueString()
	description := plan.Description.ValueString()

	// clearingAutoDestroy is true when auto_destroy_activity_duration was previously specified and
	// is now being cleared out. That one case falls back to the v1 client: go-tfe/v2's
	// Projects_attributes models this field as a plain *string, and the generated JSON serializer
	// omits nil-pointer fields entirely instead of emitting an explicit `null` (confirmed by
	// reading kiota-serialization-json-go's WriteStringValue), so there is no way to send the
	// explicit null this API needs in order to clear the field via the generated client.
	clearingAutoDestroy := !state.AutoDestroyActivityDuration.IsNull() && plan.AutoDestroyActivityDuration.IsNull()

	tflog.Debug(ctx, fmt.Sprintf("Update project %s", id))

	var projectData models.Projectsable
	var err error
	if clearingAutoDestroy {
		err = r.clearProjectAutoDestroyV1(ctx, id, name, description)
	} else {
		projectData, err = r.updateProjectV2(ctx, id, name, description, plan.AutoDestroyActivityDuration)
	}
	if err != nil {
		resp.Diagnostics.AddError("Error updating project", err.Error())
		return
	}

	// Tag bindings always go through the dedicated /projects/{id}/relationships/tag-bindings
	// endpoint now; go-tfe/v2 has no way to embed them in the same request as the attributes
	// update above.
	tagBindings := projectPlanTagBindings(plan.Tags)
	if len(tagBindings) > 0 || !plan.IgnoreAdditionalTags.ValueBool() {
		collection := newTagBindingsCollection(tagBindings)
		if err := r.config.ClientV2.API.Projects().ByProject_id(id).Relationships().TagBindings().Patch(ctx, collection, nil); err != nil {
			resp.Diagnostics.AddError("Error updating tag bindings on project", err.Error())
			return
		}
	}

	if projectData == nil {
		// The auto-destroy-clearing branch above used the v1 client, which doesn't hand back a
		// v2 model; re-read canonically via v2 to build the result.
		projEnvelope, err := r.config.ClientV2.API.Projects().ByProject_id(id).Get(ctx, nil)
		if err != nil {
			resp.Diagnostics.AddError("Error updating project", err.Error())
			return
		}
		if projEnvelope == nil || projEnvelope.GetData() == nil {
			resp.Diagnostics.AddError("Error updating project", "no data was returned by the API")
			return
		}
		projectData = projEnvelope.GetData()
	}

	result := modelFromTFEProject(projectData, tagBindings, plan.IgnoreAdditionalTags)

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)

	currentIdentity := &modelProjectIdentity{}
	resp.Diagnostics.Append(req.Identity.Get(ctx, &currentIdentity)...)
	// Only set the identity if it is null/empty in the current state
	if !resp.Diagnostics.HasError() && (currentIdentity == nil || currentIdentity.ID.IsNull()) {
		identity := modelProjectIdentity{
			ID:       result.ID,
			Hostname: types.StringValue(r.config.ClientV2.BaseURL().Host),
		}
		resp.Diagnostics.Append(resp.Identity.Set(ctx, &identity)...)
	}
}

// Delete implements resource.Resource
func (r *resourceTFEProject) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state modelTFEProject
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	tflog.Debug(ctx, fmt.Sprintf("Delete project %s", id))
	err := r.config.ClientV2.API.Projects().ByProject_id(id).Delete(ctx, nil)
	if err != nil {
		if errors.Is(err, tfev2.ErrNotFound) {
			tflog.Debug(ctx, fmt.Sprintf("Project %s no longer exists", id))
			// The resource is implicitly deleted from state after returning
			return
		}
		resp.Diagnostics.AddError("Error deleting project", err.Error())
		return
	}
}

func (r *resourceTFEProject) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}

	modifyPlanForDefaultOrganizationChange(ctx, r.config.Organization, req.State, req.Config, req.Plan, resp)
}

func (r *resourceTFEProject) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughWithIdentity(ctx, path.Root("id"), path.Root("id"), req, resp)
}
