// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &sentinelVersionResource{}
	_ resource.ResourceWithConfigure   = &sentinelVersionResource{}
	_ resource.ResourceWithImportState = &sentinelVersionResource{}
)

type sentinelVersionResource struct {
	config ConfiguredClient
}

func NewsentinelVersionResource() resource.Resource {
	return &sentinelVersionResource{}
}

func (r *sentinelVersionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "tfe_sentinel_version"
}

type modelAdminSentinelVersion struct {
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

type modelsentinelArch struct {
	URL  types.String `tfsdk:"url"`
	SHA  types.String `tfsdk:"sha"`
	OS   types.String `tfsdk:"os"`
	Arch types.String `tfsdk:"arch"`
}

func (r *sentinelVersionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
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

func (r *sentinelVersionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	tflog.Debug(ctx, "Configuring sentinel Version Resource")

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

func (d *sentinelVersionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var opaVersion modelAdminSentinelVersion
	tflog.Debug(ctx, "Creating sentinel version resource")
	resp.Diagnostics.Append(req.Plan.Get(ctx, &opaVersion)...)

	tflog.Debug(ctx, "Creating sentinel version resource", map[string]interface{}{
		"version":  opaVersion.Version.ValueString(),
		"url":      opaVersion.URL.ValueString(),
		"SHA":      opaVersion.SHA.ValueString(),
		"official": opaVersion.Official.ValueBool(),
		"enabled":  opaVersion.Enabled.ValueBool(),
		"beta":     opaVersion.Beta.ValueBool(),

		"deprecated":        opaVersion.Deprecated.ValueBool(),
		"deprecated_reason": opaVersion.DeprecatedReason.ValueString(),
		"archs":             opaVersion.Archs.ElementsAs(ctx, nil, false),
	})

	if resp.Diagnostics.HasError() {
		return
	}

	opts := tfe.AdminSentinelVersionCreateOptions{
		Version:          opaVersion.Version.ValueString(),
		URL:              opaVersion.URL.ValueString(),
		SHA:              opaVersion.SHA.ValueString(),
		Official:         tfe.Bool(opaVersion.Official.ValueBool()),
		Enabled:          tfe.Bool(opaVersion.Enabled.ValueBool()),
		Beta:             tfe.Bool(opaVersion.Beta.ValueBool()),
		Deprecated:       tfe.Bool(opaVersion.Deprecated.ValueBool()),
		DeprecatedReason: tfe.String(opaVersion.DeprecatedReason.ValueString()),
		Archs: func() []*tfe.ToolVersionArchitecture {
			archs, diags := newConvertToToolVersionArchitectures(ctx, opaVersion.Archs)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return nil
			}
			return archs
		}(),
	}
	tflog.Debug(ctx, "Creating sentinel version", map[string]interface{}{
		"version": opaVersion.Version.ValueString(),
	})

	v, err := d.config.Client.Admin.SentinelVersions.Create(ctx, opts)
	if err != nil {
		tflog.Debug(ctx, "Error creating sentinel version", map[string]interface{}{
			"error": err.Error(),
		})
		resp.Diagnostics.AddError(
			"Error creating sentinel version",
			fmt.Sprintf("Could not create sentinel version %s: %v", opts.Version, err),
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

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &opaVersion)...)
}

func (r *sentinelVersionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var opaVersion modelAdminSentinelVersion
	resp.Diagnostics.Append(req.State.Get(ctx, &opaVersion)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Read configuration of sentinel version", map[string]interface{}{
		"id": opaVersion.ID.ValueString()})

	v, err := r.config.Client.Admin.SentinelVersions.Read(ctx, opaVersion.ID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading sentinel version",
			fmt.Sprintf("Could not read sentinel version %s: %v", opaVersion.ID.ValueString(), err),
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

	// Convert archs
	if len(v.Archs) > 0 {
		archs := make([]modelsentinelArch, len(v.Archs))
		for i, arch := range v.Archs {
			archs[i] = modelsentinelArch{
				URL:  types.StringValue(arch.URL),
				SHA:  types.StringValue(arch.Sha),
				OS:   types.StringValue(arch.OS),
				Arch: types.StringValue(arch.Arch),
			}
		}
		archValues := make([]attr.Value, len(archs))
		for i, arch := range archs {
			archValues[i] = types.ObjectValueMust(
				map[string]attr.Type{
					"url":  types.StringType,
					"sha":  types.StringType,
					"os":   types.StringType,
					"arch": types.StringType,
				},
				map[string]attr.Value{
					"url":  arch.URL,
					"sha":  arch.SHA,
					"os":   arch.OS,
					"arch": arch.Arch,
				},
			)
		}
		opaVersion.Archs = types.SetValueMust(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"url":  types.StringType,
				"sha":  types.StringType,
				"os":   types.StringType,
				"arch": types.StringType,
			},
		}, archValues)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &opaVersion)...)
}

func (d *sentinelVersionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var opaVersion modelAdminSentinelVersion
	resp.Diagnostics.Append(req.Plan.Get(ctx, &opaVersion)...)

	var state modelAdminSentinelVersion
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Use the ID from the state
	opaVersion.ID = state.ID

	tflog.Debug(ctx, "Updating sentinel version resource", map[string]interface{}{
		"id": opaVersion.ID.ValueString(),
	})

	opts := tfe.AdminSentinelVersionUpdateOptions{
		Version:          tfe.String(opaVersion.Version.ValueString()),
		URL:              stringOrNil(opaVersion.URL.ValueString()),
		SHA:              tfe.String(opaVersion.SHA.ValueString()),
		Official:         tfe.Bool(opaVersion.Official.ValueBool()),
		Enabled:          tfe.Bool(opaVersion.Enabled.ValueBool()),
		Beta:             tfe.Bool(opaVersion.Beta.ValueBool()),
		Deprecated:       tfe.Bool(opaVersion.Deprecated.ValueBool()),
		DeprecatedReason: tfe.String(opaVersion.DeprecatedReason.ValueString()),
		Archs: func() []*tfe.ToolVersionArchitecture {
			archs, diags := newConvertToToolVersionArchitectures(ctx, opaVersion.Archs)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return nil
			}
			return archs
		}(),
	}

	tflog.Debug(ctx, "Updating sentinel version", map[string]interface{}{
		"id": opaVersion.ID.ValueString()})
	v, err := d.config.Client.Admin.SentinelVersions.Update(ctx, opaVersion.ID.ValueString(), opts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating sentinel version",
			fmt.Sprintf("Could not update sentinel version %s: %v", opaVersion.ID.ValueString(), err),
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

	// Set remaining attributes
	opaVersion.Official = types.BoolValue(v.Official)
	opaVersion.Enabled = types.BoolValue(v.Enabled)
	opaVersion.Beta = types.BoolValue(v.Beta)
	opaVersion.Deprecated = types.BoolValue(v.Deprecated)
	if v.DeprecatedReason != nil {
		opaVersion.DeprecatedReason = types.StringValue(*v.DeprecatedReason)
	} else {
		opaVersion.DeprecatedReason = types.StringNull()
	}

	// Handle archs just like in Read method
	if len(v.Archs) > 0 {
		archs := make([]modelsentinelArch, len(v.Archs))
		for i, arch := range v.Archs {
			archs[i] = modelsentinelArch{
				URL:  types.StringValue(arch.URL),
				SHA:  types.StringValue(arch.Sha),
				OS:   types.StringValue(arch.OS),
				Arch: types.StringValue(arch.Arch),
			}
		}
		archValues := make([]attr.Value, len(archs))
		for i, arch := range archs {
			archValues[i] = types.ObjectValueMust(
				map[string]attr.Type{
					"url":  types.StringType,
					"sha":  types.StringType,
					"os":   types.StringType,
					"arch": types.StringType,
				},
				map[string]attr.Value{
					"url":  arch.URL,
					"sha":  arch.SHA,
					"os":   arch.OS,
					"arch": arch.Arch,
				},
			)
		}
		opaVersion.Archs = types.SetValueMust(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"url":  types.StringType,
				"sha":  types.StringType,
				"os":   types.StringType,
				"arch": types.StringType,
			},
		}, archValues)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &opaVersion)...)
}

func (d *sentinelVersionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var opaVersion modelAdminSentinelVersion
	resp.Diagnostics.Append(req.State.Get(ctx, &opaVersion)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Deleting sentinel version", map[string]interface{}{
		"id": opaVersion.ID.ValueString(),
	})

	err := d.config.Client.Admin.SentinelVersions.Delete(ctx, opaVersion.ID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			tflog.Debug(ctx, "sentinel version not found, skipping deletion", map[string]interface{}{
				"id": opaVersion.ID.ValueString(),
			})
			return
		}
		resp.Diagnostics.AddError(
			"Error deleting sentinel version",
			fmt.Sprintf("Could not delete sentinel version %s: %v", opaVersion.ID.ValueString(), err),
		)
		return
	}

	resp.State.RemoveResource(ctx)
}

func (d *sentinelVersionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Splitting by '-' and checking if the first elem is equal to tool
	// determines if the string is a tool version ID
	s := strings.Split(req.ID, "-")
	if s[0] != "tool" {
		versionID, err := fetchSentinelVersionID(req.ID, d.config.Client)
		tflog.Debug(ctx, "Importing sentinel version", map[string]interface{}{
			"version_id": versionID,
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Importing sentinel Version",
				fmt.Sprintf("error retrieving sentinel version %s: %v", req.ID, err),
			)
			return
		}

		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), versionID)...)
	}
}
