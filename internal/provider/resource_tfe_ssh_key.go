// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-tfe/internal/provider/helpers"
	"github.com/hashicorp/terraform-provider-tfe/internal/provider/planmodifiers"
)

var (
	_ resource.Resource              = &resourceTFESSHKey{}
	_ resource.ResourceWithConfigure = &resourceTFESSHKey{}
)

func NewSSHKey() resource.Resource {
	return &resourceTFESSHKey{}
}

type resourceTFESSHKey struct {
	config ConfiguredClient
}

type modelTFESSHKey struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Organization types.String `tfsdk:"organization"`
	Key          types.String `tfsdk:"key"`
	KeyWO        types.String `tfsdk:"key_wo"`
}

func modelFromTFESSHKey(organization string, sshKey *tfe.SSHKey, lastValue types.String, isWriteOnly bool) *modelTFESSHKey {
	m := &modelTFESSHKey{
		ID:           types.StringValue(sshKey.ID),
		Name:         types.StringValue(sshKey.Name),
		Organization: types.StringValue(organization),
		Key:          lastValue,
	}

	if isWriteOnly {
		m.Key = types.StringNull()
	}

	return m
}

// Configure implements resource.ResourceWithConfigure
func (r *resourceTFESSHKey) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *resourceTFESSHKey) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ssh_key"
}

// Schema implements resource.Resource
func (r *resourceTFESSHKey) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Service-generated ID for the SSH key.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"name": schema.StringAttribute{
				Description: "The name of the SSH key.",
				Required:    true,
			},

			"organization": schema.StringAttribute{
				Description: "The name of the organization.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			"key": schema.StringAttribute{
				Description: "The text of the SSH private key",
				Optional:    true,
				Sensitive:   true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("key_wo")),
					stringvalidator.PreferWriteOnlyAttribute(path.MatchRoot("key_wo")),
					stringvalidator.AtLeastOneOf(path.MatchRoot("key"), path.MatchRoot("key_wo")),
				},
			},

			"key_wo": schema.StringAttribute{
				Description: "The text of the SSH private key, guaranteed not to be written to state.",
				Optional:    true,
				WriteOnly:   true,
				Sensitive:   true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("key")),
				},
				PlanModifiers: []planmodifier.String{
					planmodifiers.NewReplaceForWriteOnlyStringValue("key_wo"),
				},
			},
		},
	}
}

// Create implements resource.Resource
func (r *resourceTFESSHKey) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Load the plan and config into the model
	var plan, config modelTFESSHKey
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Determine the organization
	var organization string
	resp.Diagnostics.Append(r.config.dataOrDefaultOrganization(ctx, req.Config, &organization)...)
	if resp.Diagnostics.HasError() {
		return
	}

	options := tfe.SSHKeyCreateOptions{
		Name: plan.Name.ValueStringPointer(),
	}

	// Set Value from `value_wo` if set, otherwise use the normal value
	isWriteOnly := !config.KeyWO.IsNull()
	if isWriteOnly {
		options.Value = config.KeyWO.ValueStringPointer()
	} else {
		options.Value = plan.Key.ValueStringPointer()
	}

	tflog.Debug(ctx, fmt.Sprintf("Create new SSH key for organization: %s", organization))
	sshKey, err := r.config.Client.SSHKeys.Create(ctx, organization, options)
	if err != nil {
		resp.Diagnostics.AddError("Error creating SSH key", err.Error())
		return
	}

	// Load the response data into the model
	result := modelFromTFESSHKey(organization, sshKey, plan.Key, isWriteOnly)

	// Write the hashed private key to the state if it was provided
	if !config.KeyWO.IsNull() {
		store := r.writeOnlyValueStore(resp.Private)
		resp.Diagnostics.Append(store.SetPriorValue(ctx, config.KeyWO)...)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Update state
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

// Read implements resource.Resource
func (r *resourceTFESSHKey) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Load the plan into the model
	var state modelTFESSHKey
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Determine the organization
	var organization string
	resp.Diagnostics.Append(r.config.dataOrDefaultOrganization(ctx, req.State, &organization)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	tflog.Debug(ctx, fmt.Sprintf("Read SSH key %s for organization: %s", id, organization))
	sshKey, err := r.config.Client.SSHKeys.Read(ctx, id)
	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
			tflog.Debug(ctx, fmt.Sprintf("SSH key %s no longer exists", id))
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("Error reading SSH key", err.Error())
		return
	}

	// Check if the parameter is write-only
	isWriteOnly, diags := r.writeOnlyValueStore(resp.Private).PriorValueExists(ctx)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	// Load the response data into the model
	result := modelFromTFESSHKey(organization, sshKey, state.Key, isWriteOnly)

	// Update state
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

// Update implements resource.Resource
func (r *resourceTFESSHKey) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Load the plan and config into the model
	var plan, config modelTFESSHKey
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Determine the organization
	var organization string
	resp.Diagnostics.Append(r.config.dataOrDefaultOrganization(ctx, req.Config, &organization)...)
	if resp.Diagnostics.HasError() {
		return
	}

	options := tfe.SSHKeyUpdateOptions{
		Name: plan.Name.ValueStringPointer(),
	}

	id := plan.ID.ValueString()

	tflog.Debug(ctx, fmt.Sprintf("Update SSH key %s for organization: %s", id, organization))
	sshKey, err := r.config.Client.SSHKeys.Update(ctx, id, options)
	if err != nil {
		resp.Diagnostics.AddError("Error updating SSH key", err.Error())
		return
	}

	// Load the response data into the model
	result := modelFromTFESSHKey(organization, sshKey, plan.Key, !config.KeyWO.IsNull())

	// Update state
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

// Delete implements resource.Resource
func (r *resourceTFESSHKey) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Load the plan into the model
	var state modelTFESSHKey
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Determine the organization
	var organization string
	resp.Diagnostics.Append(r.config.dataOrDefaultOrganization(ctx, req.State, &organization)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	tflog.Debug(ctx, fmt.Sprintf("Delete SSH key %s for organization: %s", id, organization))
	err := r.config.Client.SSHKeys.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
			tflog.Debug(ctx, fmt.Sprintf("SSH key %s no longer exists", id))
			// The resource is implicitly deleted from state on return
			return
		}

		resp.Diagnostics.AddError("Error deleting SSH key", err.Error())
		return
	}
}

func (r *resourceTFESSHKey) writeOnlyValueStore(private helpers.PrivateState) *helpers.WriteOnlyValueStore {
	return helpers.NewWriteOnlyValueStore(private, "key_wo")
}
