// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// NOTE: This is a legacy resource and should be migrated to the Plugin
// Framework if substantial modifications are planned. See
// docs/new-resources.md if planning to use this code as boilerplate for
// a new resource.

package provider

import (
	"context"
	"fmt"
	"log"
	"strings"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &terraformVersionResource{}
	_ resource.ResourceWithConfigure   = &terraformVersionResource{}
	_ resource.ResourceWithImportState = &terraformVersionResource{}
)

type terraformVersionResource struct {
	config ConfiguredClient
}

func NewTerraformVersionResource() resource.Resource {
	return &terraformVersionResource{}
}

func (r *terraformVersionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "tfe_terraform_version"
}

type modelAdminTerraformVersion struct {
	ID               types.String `tfsdk:"id"`
	Version          types.String `tfsdk:"version"`
	URL              types.String `tfsdk:"url"`
	Sha              types.String `tfsdk:"sha"`
	Official         types.Bool   `tfsdk:"official"`
	Enabled          types.Bool   `tfsdk:"enabled"`
	Beta             types.Bool   `tfsdk:"beta"`
	Deprecated       types.Bool   `tfsdk:"deprecated"`
	DeprecatedReason types.String `tfsdk:"deprecated_reason"`
	Archs            types.Set    `tfsdk:"archs"`
}

type modelArch struct {
	URL  types.String `tfsdk:"url"`
	Sha  types.String `tfsdk:"sha"`
	OS   types.String `tfsdk:"os"`
	Arch types.String `tfsdk:"arch"`
}

func (r *terraformVersionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"version": schema.StringAttribute{
				Required: true,
			},
			"url": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"sha": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"official": schema.BoolAttribute{
				Optional: true,
			},
			"enabled": schema.BoolAttribute{
				Optional: true,
			},
			"beta": schema.BoolAttribute{
				Optional: true,
			},
			"deprecated": schema.BoolAttribute{
				Optional: true,
			},
			"deprecated_reason": schema.StringAttribute{
				Optional: true,
			},
			"archs": schema.SetNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"url": schema.StringAttribute{
							Required: true,
						},
						"sha": schema.StringAttribute{
							Required: true,
						},
						"os": schema.StringAttribute{
							Required: true,
						},
						"arch": schema.StringAttribute{
							Required: true,
						},
					},
				},
				Computed: true,
				Optional: true,
			},
		},
	}
}

func (r *terraformVersionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	fmt.Print("[DEBUG] Configuring terraformVersionResource\n")

	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(ConfiguredClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Provider Data Type",
			fmt.Sprintf("Expected ConfiguredClient, got: %T", req.ProviderData),
		)
		return
	}

	r.config = client
}

func (d *terraformVersionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var tfVersion modelAdminTerraformVersion
	fmt.Print("[DEBUG] Creating new Terraform version resource\n")
	resp.Diagnostics.Append(req.Plan.Get(ctx, &tfVersion)...)

	if resp.Diagnostics.HasError() {
		return
	}

	opts := tfe.AdminTerraformVersionCreateOptions{
		Version:          tfe.String(tfVersion.Version.ValueString()),
		URL:              stringOrNil(tfVersion.URL.ValueString()),
		Sha:              tfe.String(tfVersion.Sha.ValueString()),
		Official:         tfe.Bool(tfVersion.Official.ValueBool()),
		Enabled:          tfe.Bool(tfVersion.Enabled.ValueBool()),
		Beta:             tfe.Bool(tfVersion.Beta.ValueBool()),
		Deprecated:       tfe.Bool(tfVersion.Deprecated.ValueBool()),
		DeprecatedReason: tfe.String(tfVersion.DeprecatedReason.ValueString()),
		Archs: func() []*tfe.ToolVersionArchitecture {
			archs, diags := newConvertToToolVersionArchitectures(ctx, tfVersion.Archs)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return nil
			}
			return archs
		}(),
	}

	log.Printf("[DEBUG] Create new Terraform version: %s", *opts.Version)
	v, err := d.config.Client.Admin.TerraformVersions.Create(ctx, opts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Terraform version",
			fmt.Sprintf("Could not create Terraform version %s: %v", *opts.Version, err),
		)
		return
	}

	resp.State.SetAttribute(ctx, path.Root("id"), v.ID)
}

func (r *terraformVersionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var tfVersion modelAdminTerraformVersion
	resp.Diagnostics.Append(req.State.Get(ctx, &tfVersion)...)
	if resp.Diagnostics.HasError() {
		return
	}

	v, err := r.config.Client.Admin.TerraformVersions.Read(ctx, tfVersion.ID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading Terraform version",
			fmt.Sprintf("Could not read Terraform version %s: %v", tfVersion.ID.ValueString(), err),
		)
		return
	}

	// Update state with values from the API
	tfVersion.Version = types.StringValue(v.Version)
	tfVersion.URL = types.StringValue(v.URL)
	tfVersion.Sha = types.StringValue(v.Sha)
	tfVersion.Official = types.BoolValue(v.Official)
	tfVersion.Enabled = types.BoolValue(v.Enabled)
	tfVersion.Beta = types.BoolValue(v.Beta)
	tfVersion.Deprecated = types.BoolValue(v.Deprecated)
	if v.DeprecatedReason != nil {
		tfVersion.DeprecatedReason = types.StringValue(*v.DeprecatedReason)
	} else {
		tfVersion.DeprecatedReason = types.StringNull()
	}

	// Convert archs
	if v.Archs != nil && len(v.Archs) > 0 {
		// Logic to convert API archs to framework type
		// This depends on your model definitions
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &tfVersion)...)
}

func (d *terraformVersionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var tfVersion modelAdminTerraformVersion
	resp.Diagnostics.Append(req.Plan.Get(ctx, &tfVersion)...)

	if resp.Diagnostics.HasError() {
		return
	}

	opts := tfe.AdminTerraformVersionUpdateOptions{
		Version:          tfe.String(tfVersion.Version.ValueString()),
		URL:              stringOrNil(tfVersion.URL.ValueString()),
		Sha:              tfe.String(tfVersion.Sha.ValueString()),
		Official:         tfe.Bool(tfVersion.Official.ValueBool()),
		Enabled:          tfe.Bool(tfVersion.Enabled.ValueBool()),
		Beta:             tfe.Bool(tfVersion.Beta.ValueBool()),
		Deprecated:       tfe.Bool(tfVersion.Deprecated.ValueBool()),
		DeprecatedReason: tfe.String(tfVersion.DeprecatedReason.ValueString()),
		Archs: func() []*tfe.ToolVersionArchitecture {
			archs, diags := newConvertToToolVersionArchitectures(ctx, tfVersion.Archs)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return nil
			}
			return archs
		}(),
	}

	log.Printf("[DEBUG] Update Terraform version configuration for ID: %s", tfVersion.ID.ValueString())
	v, err := d.config.Client.Admin.TerraformVersions.Update(ctx, tfVersion.ID.ValueString(), opts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Terraform version",
			fmt.Sprintf("Could not update Terraform version %s: %v", tfVersion.ID.ValueString(), err),
		)
		return
	}

	resp.State.SetAttribute(ctx, path.Root("id"), v.ID)
}

func (d *terraformVersionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var tfVersion modelAdminTerraformVersion
	resp.Diagnostics.Append(req.State.Get(ctx, &tfVersion)...)
	if resp.Diagnostics.HasError() {
		return
	}
	log.Printf("[DEBUG] Delete Terraform version with ID: %s", tfVersion.ID.ValueString())
	err := d.config.Client.Admin.TerraformVersions.Delete(ctx, tfVersion.ID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			log.Printf("[DEBUG] Terraform version %s not found, skipping deletion", tfVersion.ID.ValueString())
			return
		}
		resp.Diagnostics.AddError(
			"Error deleting Terraform version",
			fmt.Sprintf("Could not delete Terraform version %s: %v", tfVersion.ID.ValueString(), err),
		)
		return
	}
	log.Printf("[DEBUG] Successfully deleted Terraform version with ID: %s", tfVersion.ID.ValueString())
	resp.State.RemoveResource(ctx)
}

func (d *terraformVersionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// The ID is expected to be the version ID, so we can directly set it
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	if resp.Diagnostics.HasError() {
		return
	}

	log.Printf("[DEBUG] Importing Terraform version with ID: %s", req.ID)
}
