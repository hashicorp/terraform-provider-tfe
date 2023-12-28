// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &resourceTFERegistryProvider{}
var _ resource.ResourceWithConfigure = &resourceTFERegistryProvider{}
var _ resource.ResourceWithImportState = &resourceTFERegistryProvider{}
var _ resource.ResourceWithModifyPlan = &resourceTFERegistryProvider{}
var _ resource.ResourceWithValidateConfig = &resourceTFERegistryProvider{}

func NewRegistryProviderResource() resource.Resource {
	return &resourceTFERegistryProvider{}
}

// resourceTFERegistryProvider implements the tfe_registry_provider resource type
type resourceTFERegistryProvider struct {
	config ConfiguredClient
}

type modelTFERegistryProvider struct {
	ID           types.String `tfsdk:"id"`
	Organization types.String `tfsdk:"organization"`
	RegistryName types.String `tfsdk:"registry_name"`
	Namespace    types.String `tfsdk:"namespace"`
	Name         types.String `tfsdk:"name"`
	CreatedAt    types.String `tfsdk:"created_at"`
	UpdatedAt    types.String `tfsdk:"updated_at"`
}

func modelFromTFERegistryProvider(v *tfe.RegistryProvider) modelTFERegistryProvider {
	return modelTFERegistryProvider{
		ID:           types.StringValue(v.ID),
		Organization: types.StringValue(v.Organization.Name),
		RegistryName: types.StringValue(string(v.RegistryName)),
		Namespace:    types.StringValue(v.Namespace),
		Name:         types.StringValue(v.Name),
		CreatedAt:    types.StringValue(v.CreatedAt),
		UpdatedAt:    types.StringValue(v.UpdatedAt),
	}
}

func (r *resourceTFERegistryProvider) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_registry_provider"
}

func (r *resourceTFERegistryProvider) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages public and private providers in the private registry.",
		Version:     1,

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID of the provider.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"organization": schema.StringAttribute{
				Description: "Name of the organization. If omitted, organization must be defined in the provider config.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"registry_name": schema.StringAttribute{
				Description: "Whether this is a publicly maintained provider or private. Must be either `public` or `private`.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("private"),
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(tfe.PrivateRegistry),
						string(tfe.PublicRegistry),
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"namespace": schema.StringAttribute{
				Description: "The namespace of the provider. For private providers this is the same as the oraganization.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the provider.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"created_at": schema.StringAttribute{
				Description: "The time when the provider was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Description: "The time when the provider was last updated.",
				Computed:    true,
			},
		},
	}
}

// Configure implements resource.ResourceWithConfigure
func (r *resourceTFERegistryProvider) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
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

func (r *resourceTFERegistryProvider) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config modelTFERegistryProvider

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if config.RegistryName.ValueString() == "public" && (config.Namespace.IsNull() || config.Namespace.IsUnknown()) {
		resp.Diagnostics.AddAttributeError(
			path.Root("namespace"),
			"Missing Attribute Configuration",
			"Expected namespace to be configured when registry_name is \"public\".",
		)
	} else if (config.RegistryName.IsNull() || config.RegistryName.ValueString() == "private") && !config.Namespace.IsNull() && !config.Namespace.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("namespace"),
			"Invalid Attribute Combination",
			"The namespace attribute cannot be configured when registry_name is \"private\".",
		)
	}
}

func (r *resourceTFERegistryProvider) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	modifyPlanForDefaultOrganizationChange(ctx, r.config.Organization, req.State, req.Config, req.Plan, resp)
}

func (r *resourceTFERegistryProvider) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan modelTFERegistryProvider

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var organization string
	resp.Diagnostics.Append(r.config.dataOrDefaultOrganization(ctx, req.Plan, &organization)...)

	if resp.Diagnostics.HasError() {
		return
	}

	registryName := plan.RegistryName.ValueString()

	var namespace string
	if registryName == "private" {
		namespace = organization
	} else {
		namespace = plan.Namespace.ValueString()
	}

	options := tfe.RegistryProviderCreateOptions{
		Type:         "registry-providers",
		Name:         plan.Name.ValueString(),
		Namespace:    namespace,
		RegistryName: tfe.RegistryName(registryName),
	}

	tflog.Debug(ctx, "Creating private registry provider")
	provider, err := r.config.Client.RegistryProviders.Create(ctx, organization, options)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create private registry provider", err.Error())
		return
	}

	result := modelFromTFERegistryProvider(provider)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *resourceTFERegistryProvider) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state modelTFERegistryProvider

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var organization string
	resp.Diagnostics.Append(r.config.dataOrDefaultOrganization(ctx, req.State, &organization)...)

	if resp.Diagnostics.HasError() {
		return
	}

	registryName := state.RegistryName.ValueString()

	providerID := tfe.RegistryProviderID{
		OrganizationName: organization,
		RegistryName:     tfe.RegistryName(registryName),
		Namespace:        state.Namespace.ValueString(),
		Name:             state.Name.ValueString(),
	}

	options := tfe.RegistryProviderReadOptions{}

	tflog.Debug(ctx, "Reading private registry provider")
	provider, err := r.config.Client.RegistryProviders.Read(ctx, providerID, &options)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read private registry provider", err.Error())
		return
	}

	result := modelFromTFERegistryProvider(provider)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *resourceTFERegistryProvider) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// If the resource does not support modification and should always be recreated on
	// configuration value updates, the Update logic can be left empty and ensure all
	// configurable schema attributes implement the resource.RequiresReplace()
	// attribute plan modifier.
	resp.Diagnostics.AddError("Update not supported", "The update operation is not supported on this resource. This is a bug in the provider.")
}

func (r *resourceTFERegistryProvider) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state modelTFERegistryProvider

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	registryName := state.RegistryName.ValueString()

	providerID := tfe.RegistryProviderID{
		OrganizationName: state.Organization.ValueString(),
		RegistryName:     tfe.RegistryName(registryName),
		Namespace:        state.Namespace.ValueString(),
		Name:             state.Name.ValueString(),
	}

	tflog.Debug(ctx, "Deleting private registry provider")
	err := r.config.Client.RegistryProviders.Delete(ctx, providerID)
	if err != nil {
		resp.Diagnostics.AddError("Unable to delete private registry provider", err.Error())
		return
	}
}

func (r *resourceTFERegistryProvider) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	s := strings.SplitN(req.ID, "/", 4)
	if len(s) != 4 {
		resp.Diagnostics.AddError(
			"Error importing variable",
			fmt.Sprintf("Invalid variable import format: %s (expected <ORGANIZATION>/<REGISTRY NAME>/<NAMESPACE>/<PROVIDER NAME>)", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization"), s[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("registry_name"), s[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("namespace"), s[2])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), s[3])...)
}
