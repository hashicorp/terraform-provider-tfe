// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &resourceTFERegistryArtifactTag{}
	_ resource.ResourceWithConfigure   = &resourceTFERegistryArtifactTag{}
	_ resource.ResourceWithImportState = &resourceTFERegistryArtifactTag{}
)

type resourceTFERegistryArtifactTag struct {
	config ConfiguredClient
}

const (
	ArtifactTypeRegistryModule    string = "registry-module"
	ArtifactTypeRegistryProvider  string = "registry-provider"
	ArtifactTypeRegistryComponent string = "registry-component"
)

func NewTFERegistryArtifactTagResource() resource.Resource {
	return &resourceTFERegistryArtifactTag{}
}

type modelRegistryArtifactTag struct {
	ID       types.String  `tfsdk:"id"`
	Tags     []modelTag    `tfsdk:"tags"`
	Artifact modelArtifact `tfsdk:"artifact"`
}

type modelTag struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

type modelArtifact struct {
	ID   types.String `tfsdk:"id"`
	Type types.String `tfsdk:"type"`
}

func (r *resourceTFERegistryArtifactTag) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(ConfiguredClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected tfe.ConfiguredClient, got %T. This is a bug in the tfe provider, so please report it on GitHub.", req.ProviderData),
		)
		return
	}

	r.config = client
}

// Metadata implements [resource.Resource].
func (r *resourceTFERegistryArtifactTag) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_registry_artifact_tag"
}

// Schema implements [resource.Resource].
func (r *resourceTFERegistryArtifactTag) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a resource which manages the tags on registry artifacts.",
		Version:     0,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the tag binding.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"tags": schema.ListNestedAttribute{
				Description: "The list of the project tags to assign to the registry artifact. This will replace all existing tags on the artifact.",
				Required:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							Description: "The tag key.",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.LengthAtLeast(1),
							},
						},
						"value": schema.StringAttribute{
							Description: "The tag value.",
							Required:    true,
						},
					},
				},
			},
			"artifact": schema.SingleNestedAttribute{
				Required:    true,
				Description: "Object representing the registry artifact to be tagged.",
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Description: "The type of the Registry Artifact.",
						Required:    true,
						Validators: []validator.String{
							stringvalidator.OneOf(
								ArtifactTypeRegistryModule,
								ArtifactTypeRegistryProvider,
								ArtifactTypeRegistryComponent,
							),
						},
					},
					"id": schema.StringAttribute{
						Description: "The external ID of the registry artifact.",
						Required:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
				},
			},
		},
	}
}

// Create implements [resource.Resource].
func (r *resourceTFERegistryArtifactTag) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan modelRegistryArtifactTag
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := plan.Artifact.ID.ValueString()
	atype := plan.Artifact.Type.ValueString()

	tflog.Debug(ctx, fmt.Sprintf("Adding tags for registry artifact type %s with ID %s", atype, id))

	err := r.updateTagBindings(ctx, atype, id, modelTagsToTFETagBindings(plan.Tags))
	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
			resp.Diagnostics.AddError(
				"Registry Artifact Not Found",
				fmt.Sprintf("The registry artifact with ID %s was not found. Please verify that the artifact ID is correct and that it exists.", id),
			)
			return
		}
		resp.Diagnostics.AddError(
			"Error Adding Tags to Registry Artifact",
			fmt.Sprintf("An error was encountered when adding tags to registry artifact %q: %s", id, err),
		)
		return

	}

	// Generate a deterministic composite ID from the artifact type and artifact ID.
	// This uniquely identifies the resource and allows it to be reconstructed during import.
	// Similar to the pattern used by other association resources, e.g. resource_stack_variable_set.go
	plan.ID = types.StringValue(fmt.Sprintf("%s/%s", atype, id))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func modelTagsToTFETagBindings(tags []modelTag) []*tfe.TagBinding {
	out := make([]*tfe.TagBinding, 0, len(tags))

	for _, tag := range tags {
		// Skip keys that are not null or empty at apply time.
		if tag.Key.IsNull() || tag.Key.ValueString() == "" {
			continue
		}

		k := tag.Key.ValueString()
		v := tag.Value.ValueString()

		out = append(out, &tfe.TagBinding{
			Key:   k,
			Value: v,
		})
	}
	return out
}

// Read implements [resource.Resource].
func (r *resourceTFERegistryArtifactTag) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state modelRegistryArtifactTag
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.Artifact.ID.ValueString()
	atype := state.Artifact.Type.ValueString()

	bindings, err := r.listTagBindings(ctx, atype, id)
	if errors.Is(err, tfe.ErrResourceNotFound) {
		tflog.Debug(ctx, fmt.Sprintf("Registry %s : %s no longer exists", atype, id))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Error reading tag bindings for registry artifact", err.Error())
		return
	}

	state.Tags = modelFromTFETagBindings(bindings)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// modelFromTFETagBindings builds an array of modelTags from an array of tfe.TagBindings value.
func modelFromTFETagBindings(tags []*tfe.TagBinding) []modelTag {
	out := make([]modelTag, 0, len(tags))

	for _, binding := range tags {
		if binding == nil {
			continue
		}
		out = append(out, modelTag{
			Key:   types.StringValue(binding.Key),
			Value: types.StringValue(binding.Value),
		})
	}
	return out
}

// listTagBindings fetches the current tag bindings for a registry artifact by
// the artifact type. Returns ErrResourceNotFound if the artifact no longer exists.
func (r *resourceTFERegistryArtifactTag) listTagBindings(ctx context.Context, artifactType, artifactID string) ([]*tfe.TagBinding, error) {
	switch artifactType {
	case ArtifactTypeRegistryModule:
		return r.config.Client.RegistryModules.ListTagBindings(ctx, artifactID)
	case ArtifactTypeRegistryProvider:
		return r.config.Client.RegistryProviders.ListTagBindings(ctx, artifactID)
	case ArtifactTypeRegistryComponent:
		return r.config.Client.RegistryComponents.ListTagBindings(ctx, artifactID)
	default:
		return nil, fmt.Errorf("unsupported artifact type: %s", artifactType)
	}
}

// updateTagBindings replaces all tag bindings on a registry artifact by
// the artifact type. Pass an empty slice to clear all tags.
func (r *resourceTFERegistryArtifactTag) updateTagBindings(ctx context.Context, artifactType, artifactID string, tags []*tfe.TagBinding) error {
	switch artifactType {
	case ArtifactTypeRegistryModule:
		rm, err := r.config.Client.RegistryModules.Read(ctx, tfe.RegistryModuleID{ID: artifactID})
		if err != nil {
			return err
		}
		_, err = r.config.Client.RegistryModules.Update(ctx, tfe.RegistryModuleID{
			Name:         rm.Name,
			RegistryName: rm.RegistryName,
			Namespace:    rm.Namespace,
			Organization: rm.Organization.Name,
			Provider:     rm.Provider,
		}, tfe.RegistryModuleUpdateOptions{TagBindings: tags})
		return err
	case ArtifactTypeRegistryProvider:
		rp, err := r.config.Client.RegistryProviders.Read(ctx, tfe.RegistryProviderID{ID: artifactID}, nil)
		if err != nil {
			return err
		}
		_, err = r.config.Client.RegistryProviders.Update(ctx, tfe.RegistryProviderID{
			Name:             rp.Name,
			RegistryName:     rp.RegistryName,
			Namespace:        rp.Namespace,
			OrganizationName: rp.Organization.Name,
		}, &tfe.RegistryProviderUpdateOptions{TagBindings: tags})
		return err
	case ArtifactTypeRegistryComponent:
		_, err := r.config.Client.RegistryComponents.Update(ctx, artifactID, &tfe.RegistryComponentUpdateOptions{TagBindings: tags})
		return err
	default:
		return fmt.Errorf("unsupported artifact type: %s", artifactType)
	}
}

// Update implements [resource.Resource].
func (r *resourceTFERegistryArtifactTag) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan modelRegistryArtifactTag

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := plan.Artifact.ID.ValueString()
	atype := plan.Artifact.Type.ValueString()

	tflog.Debug(ctx, fmt.Sprintf("Updating tags for registry artifact type %s with ID %s", atype, id))

	err := r.updateTagBindings(ctx, atype, id, modelTagsToTFETagBindings(plan.Tags))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Tags on Registry Artifact",
			fmt.Sprintf("An error was encountered when updating tags on registry artifact %q: %s", id, err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete implements [resource.Resource].
func (r *resourceTFERegistryArtifactTag) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Read the Terraform state into the model
	var state modelRegistryArtifactTag
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	atype := state.Artifact.Type.ValueString()
	id := state.Artifact.ID.ValueString()

	tflog.Debug(ctx, fmt.Sprintf("Removing all tags from registry artifact type %s with ID %s", atype, id))

	// Clear all tags by passing an empty slice to the update API.
	err := r.updateTagBindings(ctx, atype, id, []*tfe.TagBinding{})
	if errors.Is(err, tfe.ErrResourceNotFound) {
		// Artifact is already deleted, nothing to do.
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Removing Tags from Registry Artifact",
			fmt.Sprintf("An error was encountered when removing tags from registry artifact %q: %s", id, err),
		)
		return
	}
}

// ImportState implements [resource.ResourceWithImportState].
func (r *resourceTFERegistryArtifactTag) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID Format",
			fmt.Sprintf("Expected format: <artifact_type>/<artifact_id>. Got: %q", req.ID),
		)
		return
	}

	artifactType := parts[0]
	artifactID := parts[1]

	validTypes := []string{ArtifactTypeRegistryModule, ArtifactTypeRegistryProvider, ArtifactTypeRegistryComponent}
	valid := false
	for _, t := range validTypes {
		if artifactType == t {
			valid = true
			break
		}
	}
	if !valid {
		resp.Diagnostics.AddError(
			"Invalid Artifact Type",
			fmt.Sprintf("Artifact type %q is not valid. Must be one of: %v", artifactType, validTypes),
		)
		return
	}

	state := modelRegistryArtifactTag{
		ID: types.StringValue(req.ID),
		Artifact: modelArtifact{
			Type: types.StringValue(artifactType),
			ID:   types.StringValue(artifactID),
		},
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
