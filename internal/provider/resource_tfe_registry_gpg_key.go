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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &resourceTFERegistryGPGKey{}
var _ resource.ResourceWithConfigure = &resourceTFERegistryGPGKey{}
var _ resource.ResourceWithImportState = &resourceTFERegistryGPGKey{}
var _ resource.ResourceWithModifyPlan = &resourceTFERegistryGPGKey{}

func NewRegistryGPGKeyResource() resource.Resource {
	return &resourceTFERegistryGPGKey{}
}

// resourceTFERegistryGPGKey implements the tfe_registry_gpg_key resource type
type resourceTFERegistryGPGKey struct {
	config ConfiguredClient
}

func (r *resourceTFERegistryGPGKey) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_registry_gpg_key"
}

func (r *resourceTFERegistryGPGKey) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	modifyPlanForDefaultOrganizationChange(ctx, r.config.Organization, req, resp)
}

func (r *resourceTFERegistryGPGKey) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a public key of the GPG key pair used to sign releases of private providers in the private registry.",
		Version:     1,

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID of the GPG key.",
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
			"ascii_armor": schema.StringAttribute{
				Description: "ASCII-armored representation of the GPG key.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"created_at": schema.StringAttribute{
				Description: "The time when the GPG key was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Description: "The time when the GPG key was last updated.",
				Computed:    true,
			},
		},
	}
}

// Configure implements resource.ResourceWithConfigure
func (r *resourceTFERegistryGPGKey) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *resourceTFERegistryGPGKey) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan modelTFERegistryGPGKey

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

	options := tfe.GPGKeyCreateOptions{
		Type:       "gpg-keys",
		Namespace:  organization,
		AsciiArmor: plan.ASCIIArmor.ValueString(),
	}

	tflog.Debug(ctx, "Creating private registry GPG key")
	key, err := r.config.Client.GPGKeys.Create(ctx, "private", options)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create private registry GPG key", err.Error())
		return
	}

	result := modelFromTFEVGPGKey(key)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *resourceTFERegistryGPGKey) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state modelTFERegistryGPGKey

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

	keyID := tfe.GPGKeyID{
		RegistryName: "private",
		Namespace:    organization,
		KeyID:        state.ID.ValueString(),
	}

	tflog.Debug(ctx, "Reading private registry GPG key")
	key, err := r.config.Client.GPGKeys.Read(ctx, keyID)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read private registry GPG key", err.Error())
		return
	}

	result := modelFromTFEVGPGKey(key)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *resourceTFERegistryGPGKey) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// If the resource does not support modification and should always be recreated on
	// configuration value updates, the Update logic can be left empty and ensure all
	// configurable schema attributes implement the resource.RequiresReplace()
	// attribute plan modifier.
	resp.Diagnostics.AddError("Update not supported", "The update operation is not supported on this resource. This is a bug in the provider.")
}

func (r *resourceTFERegistryGPGKey) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state modelTFERegistryGPGKey

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	keyID := tfe.GPGKeyID{
		RegistryName: "private",
		Namespace:    state.Organization.ValueString(),
		KeyID:        state.ID.ValueString(),
	}

	tflog.Debug(ctx, "Deleting private registry GPG key")
	err := r.config.Client.GPGKeys.Delete(ctx, keyID)
	if err != nil {
		resp.Diagnostics.AddError("Unable to delete private registry GPG key", err.Error())
		return
	}
}

func (r *resourceTFERegistryGPGKey) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	s := strings.SplitN(req.ID, "/", 2)
	if len(s) != 2 {
		resp.Diagnostics.AddError(
			"Error importing variable",
			fmt.Sprintf("Invalid variable import format: %s (expected <ORGANIZATION>/<KEY ID>)", req.ID),
		)
		return
	}
	org := s[0]
	id := s[1]

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization"), org)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
