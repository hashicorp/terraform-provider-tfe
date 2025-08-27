// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
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

func (r *terraformVersionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"version": schema.StringAttribute{
				Required: true,
			},
			"url": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					SyncTopLevelURLSHAWithAMD64(),
				},
			},
			"sha": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					SyncTopLevelURLSHAWithAMD64(),
				},
			},
			"official": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
			"beta": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"deprecated": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
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
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
					PreserveAMD64ArchsOnChange(),
				},
			},
		},
	}
}

func (r *terraformVersionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	tflog.Debug(ctx, "Configuring Terraform Version Resource")

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

func (r *terraformVersionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var tfVersion modelAdminTerraformVersion
	tflog.Debug(ctx, "Creating Terraform version resource")
	resp.Diagnostics.Append(req.Plan.Get(ctx, &tfVersion)...)

	tflog.Debug(ctx, "Creating Terraform version resource", map[string]interface{}{
		"version":  tfVersion.Version.ValueString(),
		"url":      tfVersion.URL.ValueString(),
		"sha":      tfVersion.Sha.ValueString(),
		"official": tfVersion.Official.ValueBool(),
		"enabled":  tfVersion.Enabled.ValueBool(),
		"beta":     tfVersion.Beta.ValueBool(),

		"deprecated":        tfVersion.Deprecated.ValueBool(),
		"deprecated_reason": tfVersion.DeprecatedReason.ValueString(),
		"archs":             tfVersion.Archs.ElementsAs(ctx, nil, false),
	})

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
			archs, diags := convertToToolVersionArchitectures(ctx, tfVersion.Archs)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return nil
			}
			return archs
		}(),
	}
	tflog.Debug(ctx, "Creating Terraform version", map[string]interface{}{
		"version": tfVersion.Version.ValueString(),
	})

	v, err := r.config.Client.Admin.TerraformVersions.Create(ctx, opts)
	if err != nil {
		tflog.Debug(ctx, "Error creating Terraform version", map[string]interface{}{
			"error": err.Error(),
		})
		resp.Diagnostics.AddError(
			"Error creating Terraform version",
			fmt.Sprintf("Could not create Terraform version %s: %v", *opts.Version, err),
		)
		return
	}

	tfVersion.ID = types.StringValue(v.ID)
	tfVersion.Version = types.StringValue(v.Version)
	tfVersion.Official = types.BoolValue(v.Official)
	tfVersion.Enabled = types.BoolValue(v.Enabled)
	tfVersion.Beta = types.BoolValue(v.Beta)
	tfVersion.Deprecated = types.BoolValue(v.Deprecated)
	if v.DeprecatedReason != nil {
		tfVersion.DeprecatedReason = types.StringValue(*v.DeprecatedReason)
	} else {
		tfVersion.DeprecatedReason = types.StringNull()
	}
	if v.URL == "" {
		tfVersion.URL = types.StringNull()
	} else {
		tfVersion.URL = types.StringValue(v.URL)
	}
	if v.Sha == "" {
		tfVersion.Sha = types.StringNull()
	} else {
		tfVersion.Sha = types.StringValue(v.Sha)
	}
	tfVersion.Archs = convertAPIArchsToFrameworkSet(v.Archs)

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &tfVersion)...)
}

func (r *terraformVersionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var tfVersion modelAdminTerraformVersion
	resp.Diagnostics.Append(req.State.Get(ctx, &tfVersion)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Read configuration of Terraform version", map[string]interface{}{
		"id": tfVersion.ID.ValueString()})

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

	tfVersion.ID = types.StringValue(v.ID)
	tfVersion.Version = types.StringValue(v.Version)
	tfVersion.Official = types.BoolValue(v.Official)
	tfVersion.Enabled = types.BoolValue(v.Enabled)
	tfVersion.Beta = types.BoolValue(v.Beta)
	tfVersion.Deprecated = types.BoolValue(v.Deprecated)
	if v.DeprecatedReason != nil {
		tfVersion.DeprecatedReason = types.StringValue(*v.DeprecatedReason)
	} else {
		tfVersion.DeprecatedReason = types.StringNull()
	}
	if v.URL == "" {
		tfVersion.URL = types.StringNull()
	} else {
		tfVersion.URL = types.StringValue(v.URL)
	}
	if v.Sha == "" {
		tfVersion.Sha = types.StringNull()
	} else {
		tfVersion.Sha = types.StringValue(v.Sha)
	}
	tfVersion.Archs = convertAPIArchsToFrameworkSet(v.Archs)

	resp.Diagnostics.Append(resp.State.Set(ctx, &tfVersion)...)
}

func (r *terraformVersionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var tfVersion modelAdminTerraformVersion
	resp.Diagnostics.Append(req.Plan.Get(ctx, &tfVersion)...)

	var state modelAdminTerraformVersion
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Use the ID from the state
	tfVersion.ID = state.ID

	tflog.Debug(ctx, "Updating Terraform version resource", map[string]interface{}{
		"id": tfVersion.ID.ValueString(),
	})

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
			archs, diags := convertToToolVersionArchitectures(ctx, tfVersion.Archs)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return nil
			}
			return archs
		}(),
	}

	tflog.Debug(ctx, "Updating Terraform version", map[string]interface{}{
		"id": tfVersion.ID.ValueString()})

	v, err := r.config.Client.Admin.TerraformVersions.Update(ctx, tfVersion.ID.ValueString(), opts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Terraform version",
			fmt.Sprintf("Could not update Terraform version %s: %v", tfVersion.ID.ValueString(), err),
		)
		return
	}

	tfVersion.ID = types.StringValue(v.ID)
	tfVersion.Version = types.StringValue(v.Version)
	tfVersion.Official = types.BoolValue(v.Official)
	tfVersion.Enabled = types.BoolValue(v.Enabled)
	tfVersion.Beta = types.BoolValue(v.Beta)
	tfVersion.Deprecated = types.BoolValue(v.Deprecated)
	if v.DeprecatedReason != nil {
		tfVersion.DeprecatedReason = types.StringValue(*v.DeprecatedReason)
	} else {
		tfVersion.DeprecatedReason = types.StringNull()
	}
	if v.URL == "" {
		tfVersion.URL = types.StringNull()
	} else {
		tfVersion.URL = types.StringValue(v.URL)
	}
	if v.Sha == "" {
		tfVersion.Sha = types.StringNull()
	} else {
		tfVersion.Sha = types.StringValue(v.Sha)
	}
	tfVersion.Archs = convertAPIArchsToFrameworkSet(v.Archs)

	resp.Diagnostics.Append(resp.State.Set(ctx, &tfVersion)...)
}

func (r *terraformVersionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var tfVersion modelAdminTerraformVersion
	resp.Diagnostics.Append(req.State.Get(ctx, &tfVersion)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Deleting Terraform version", map[string]interface{}{
		"id": tfVersion.ID.ValueString(),
	})

	err := r.config.Client.Admin.TerraformVersions.Delete(ctx, tfVersion.ID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			tflog.Debug(ctx, "Terraform version not found, skipping deletion", map[string]interface{}{
				"id": tfVersion.ID.ValueString(),
			})
			return
		}
		resp.Diagnostics.AddError(
			"Error deleting Terraform version",
			fmt.Sprintf("Could not delete Terraform version %s: %v", tfVersion.ID.ValueString(), err),
		)
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *terraformVersionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var id string
	// Splitting by '-' and checking if the first elem is equal to tool
	// determines if the string is a tool version ID
	s := strings.Split(req.ID, "-")
	if s[0] != "tool" {
		versionID, err := fetchTerraformVersionID(req.ID, r.config.Client)
		tflog.Debug(ctx, "Importing Terraform version", map[string]interface{}{
			"version_id": versionID,
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Importing Terraform Version",
				fmt.Sprintf("error retrieving terraform version %s: %v", req.ID, err),
			)
			return
		}

		id = versionID
	} else {
		id = req.ID
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
