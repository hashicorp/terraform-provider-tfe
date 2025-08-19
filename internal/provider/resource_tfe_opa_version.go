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
	"strings"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &OPAVersionResource{}
	_ resource.ResourceWithConfigure   = &OPAVersionResource{}
	_ resource.ResourceWithImportState = &OPAVersionResource{}
)

type OPAVersionResource struct {
	config ConfiguredClient
}

func NewOPAVersionResource() resource.Resource {
	return &OPAVersionResource{}
}

func (r *OPAVersionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "tfe_opa_version"
}

type modelAdminOPAVersion struct {
	ID               types.String `tfsdk:"id"`
	Version          types.String `tfsdk:"version"`
	URL              types.String `tfsdk:"url"`
	SHA              types.String `tfsdk:"sha"`
	Official         types.Bool   `tfsdk:"official"`
	Enabled          types.Bool   `tfsdk:"enabled"`
	Beta             types.Bool   `tfsdk:"beta"`
	Deprecated       types.Bool   `tfsdk:"deprecated"`
	DeprecatedReason types.String `tfsdk:"deprecated_reason"`
	Archs            types.Set    `tfsdk:"archs"`
}

func (r *OPAVersionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
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
						"sha": schema.StringAttribute{ // Ensure lowercase
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

func (r *OPAVersionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	tflog.Debug(ctx, "Configuring OPA Version Resource")

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

func (r *OPAVersionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var opaVersion modelAdminOPAVersion
	tflog.Debug(ctx, "Creating OPA version resource")
	resp.Diagnostics.Append(req.Plan.Get(ctx, &opaVersion)...)

	tflog.Debug(ctx, "Creating OPA version resource", map[string]interface{}{
		"version":           opaVersion.Version.ValueString(),
		"url":               opaVersion.URL.ValueString(),
		"SHA":               opaVersion.SHA.ValueString(),
		"official":          opaVersion.Official.ValueBool(),
		"enabled":           opaVersion.Enabled.ValueBool(),
		"beta":              opaVersion.Beta.ValueBool(),
		"deprecated":        opaVersion.Deprecated.ValueBool(),
		"deprecated_reason": opaVersion.DeprecatedReason.ValueString(),
		"archs":             opaVersion.Archs.ElementsAs(ctx, nil, false),
	})

	if resp.Diagnostics.HasError() {
		return
	}

	opts := tfe.AdminOPAVersionCreateOptions{
		Version:          opaVersion.Version.ValueString(),
		URL:              opaVersion.URL.ValueString(),
		SHA:              opaVersion.SHA.ValueString(),
		Official:         tfe.Bool(opaVersion.Official.ValueBool()),
		Enabled:          tfe.Bool(opaVersion.Enabled.ValueBool()),
		Beta:             tfe.Bool(opaVersion.Beta.ValueBool()),
		Deprecated:       tfe.Bool(opaVersion.Deprecated.ValueBool()),
		DeprecatedReason: tfe.String(opaVersion.DeprecatedReason.ValueString()),
		Archs: func() []*tfe.ToolVersionArchitecture {
			archs, diags := convertToToolVersionArchitectures(ctx, opaVersion.Archs)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return nil
			}
			return archs
		}(),
	}
	tflog.Debug(ctx, "Creating OPA version", map[string]interface{}{
		"version": opaVersion.Version.ValueString(),
	})

	v, err := r.config.Client.Admin.OPAVersions.Create(ctx, opts)
	if err != nil {
		tflog.Debug(ctx, "Error creating OPA version", map[string]interface{}{
			"error": err.Error(),
		})
		resp.Diagnostics.AddError(
			"Error creating OPA version",
			fmt.Sprintf("Could not create OPA version %s: %v", opts.Version, err),
		)
		return
	}

	opaVersion.ID = types.StringValue(v.ID)

	// ensure there are no unknown values
	if v.URL == "" {
		opaVersion.URL = types.StringNull()
	} else {
		opaVersion.URL = types.StringValue(v.URL)
	}
	if v.SHA == "" {
		opaVersion.SHA = types.StringNull()
	} else {
		opaVersion.SHA = types.StringValue(v.SHA)
	}

	opaVersion.Archs = convertAPIArchsToFrameworkSet(v.Archs)

	resp.Diagnostics.Append(resp.State.Set(ctx, &opaVersion)...)
}

func (r *OPAVersionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var opaVersion modelAdminOPAVersion
	resp.Diagnostics.Append(req.State.Get(ctx, &opaVersion)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Read configuration of OPA version", map[string]interface{}{
		"id": opaVersion.ID.ValueString()})

	v, err := r.config.Client.Admin.OPAVersions.Read(ctx, opaVersion.ID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading OPA version",
			fmt.Sprintf("Could not read OPA version %s: %v", opaVersion.ID.ValueString(), err),
		)
		return
	}

	opaVersion.ID = types.StringValue(v.ID)
	opaVersion.Official = types.BoolValue(v.Official)
	opaVersion.Enabled = types.BoolValue(v.Enabled)
	opaVersion.Beta = types.BoolValue(v.Beta)
	opaVersion.Deprecated = types.BoolValue(v.Deprecated)
	if v.DeprecatedReason != nil {
		opaVersion.DeprecatedReason = types.StringValue(*v.DeprecatedReason)
	} else {
		opaVersion.DeprecatedReason = types.StringNull()
	}
	if v.URL == "" {
		opaVersion.URL = types.StringNull()
	} else {
		opaVersion.URL = types.StringValue(v.URL)
	}
	if v.SHA == "" {
		opaVersion.SHA = types.StringNull()
	} else {
		opaVersion.SHA = types.StringValue(v.SHA)
	}

	opaVersion.Archs = convertAPIArchsToFrameworkSet(v.Archs)

	resp.Diagnostics.Append(resp.State.Set(ctx, &opaVersion)...)
}

func (r *OPAVersionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var opaVersion modelAdminOPAVersion
	resp.Diagnostics.Append(req.Plan.Get(ctx, &opaVersion)...)

	var state modelAdminOPAVersion
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Use the ID from the state
	opaVersion.ID = state.ID

	tflog.Debug(ctx, "Updating OPA version resource", map[string]interface{}{
		"id": opaVersion.ID.ValueString(),
	})

	opts := tfe.AdminOPAVersionUpdateOptions{
		Version:          tfe.String(opaVersion.Version.ValueString()),
		URL:              stringOrNil(opaVersion.URL.ValueString()),
		SHA:              tfe.String(opaVersion.SHA.ValueString()),
		Official:         tfe.Bool(opaVersion.Official.ValueBool()),
		Enabled:          tfe.Bool(opaVersion.Enabled.ValueBool()),
		Beta:             tfe.Bool(opaVersion.Beta.ValueBool()),
		Deprecated:       tfe.Bool(opaVersion.Deprecated.ValueBool()),
		DeprecatedReason: tfe.String(opaVersion.DeprecatedReason.ValueString()),
		Archs: func() []*tfe.ToolVersionArchitecture {
			archs, diags := convertToToolVersionArchitectures(ctx, opaVersion.Archs)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return nil
			}
			return archs
		}(),
	}

	tflog.Debug(ctx, "Updating OPA version", map[string]interface{}{
		"id": opaVersion.ID.ValueString()})
	v, err := r.config.Client.Admin.OPAVersions.Update(ctx, opaVersion.ID.ValueString(), opts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating OPA version",
			fmt.Sprintf("Could not update OPA version %s: %v", opaVersion.ID.ValueString(), err),
		)
		return
	}

	// Set ID and other attributes
	opaVersion.ID = types.StringValue(v.ID)
	opaVersion.Version = types.StringValue(v.Version)

	// IMPORTANT: Set explicit values for URL and SHA
	if v.URL != "" {
		opaVersion.URL = types.StringValue(v.URL)
	} else {
		opaVersion.URL = types.StringNull()
	}

	if v.SHA != "" {
		opaVersion.SHA = types.StringValue(v.SHA)
	} else {
		opaVersion.SHA = types.StringNull()
	}

	opaVersion.Official = types.BoolValue(v.Official)
	opaVersion.Enabled = types.BoolValue(v.Enabled)
	opaVersion.Beta = types.BoolValue(v.Beta)
	opaVersion.Deprecated = types.BoolValue(v.Deprecated)
	if v.DeprecatedReason != nil {
		opaVersion.DeprecatedReason = types.StringValue(*v.DeprecatedReason)
	} else {
		opaVersion.DeprecatedReason = types.StringNull()
	}

	opaVersion.Archs = convertAPIArchsToFrameworkSet(v.Archs)

	resp.Diagnostics.Append(resp.State.Set(ctx, &opaVersion)...)
}

func (r *OPAVersionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var opaVersion modelAdminOPAVersion
	resp.Diagnostics.Append(req.State.Get(ctx, &opaVersion)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Deleting OPA version", map[string]interface{}{
		"id": opaVersion.ID.ValueString(),
	})

	err := r.config.Client.Admin.OPAVersions.Delete(ctx, opaVersion.ID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			tflog.Debug(ctx, "OPA version not found, skipping deletion", map[string]interface{}{
				"id": opaVersion.ID.ValueString(),
			})
			return
		}
		resp.Diagnostics.AddError(
			"Error deleting OPA version",
			fmt.Sprintf("Could not delete OPA version %s: %v", opaVersion.ID.ValueString(), err),
		)
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *OPAVersionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Splitting by '-' and checking if the first elem is equal to tool
	// determines if the string is a tool version ID
	s := strings.Split(req.ID, "-")
	if s[0] != "tool" {
		versionID, err := fetchOPAVersionID(req.ID, r.config.Client)
		tflog.Debug(ctx, "Importing OPA version", map[string]interface{}{
			"version_id": versionID,
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Importing OPA Version",
				fmt.Sprintf("error retrieving OPA version %s: %v", req.ID, err),
			)
			return
		}

		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), versionID)...)
	}
}
