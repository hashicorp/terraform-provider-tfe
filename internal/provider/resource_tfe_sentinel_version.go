// Copyright IBM Corp. 2018, 2025
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

func (r *sentinelVersionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages Sentinel versions available on HCP Terraform and Terraform Enterprise.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:   "The ID of the Sentinel version.",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"version": schema.StringAttribute{
				Description: "A semantic version string in N.N.N or N.N.N-bundleName format.",
				Required:    true,
			},
			"url": schema.StringAttribute{
				Description: "The URL where a ZIP-compressed 64-bit Linux binary of this version can be downloaded. Soon to be deprecated in favor of the archs attribute.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					SyncTopLevelURLSHAWithAMD64(),
				},
			},
			"sha": schema.StringAttribute{
				Description: "The SHA-256 checksum of the compressed Sentinel binary. Soon to be deprecated in favor of the archs attribute.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					SyncTopLevelURLSHAWithAMD64(),
				},
			},
			"official": schema.BoolAttribute{
				Description: "Whether or not this is an official release of Sentinel. Defaults to false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether or not this version of Sentinel is enabled for use in HCP Terraform and Terraform Enterprise. Defaults to true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"beta": schema.BoolAttribute{
				Description: "Whether or not this version of Sentinel is beta pre-release. Defaults to false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"deprecated": schema.BoolAttribute{
				Description: "Whether or not this version of Sentinel is deprecated. Defaults to false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"deprecated_reason": schema.StringAttribute{
				Description: "Additional context about why a version of Sentinel is deprecated. Defaults to null unless deprecated is true.",
				Optional:    true,
			},
			"archs": schema.SetNestedAttribute{
				Description: "A list of architecture-specific binaries for this Sentinel version. When specifying architecture-specific binaries, the top-level url and sha attributes are deprecated and should not be used. If both top-level url and sha are specified, an archs entry for the amd64 architecture must also be included, and its url and sha values must match the top-level values.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"url": schema.StringAttribute{
							Description: "The URL where a ZIP-compressed binary of this version can be downloaded.",
							Required:    true,
						},
						"sha": schema.StringAttribute{
							Description: "The SHA-256 checksum of the compressed binary.",
							Required:    true,
						},
						"os": schema.StringAttribute{
							Description: "The operating system for which this binary is intended.",
							Required:    true,
						},
						"arch": schema.StringAttribute{
							Description: "The architecture for which this binary is intended.",
							Required:    true,
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

func (r *sentinelVersionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var sentinelVersion modelAdminSentinelVersion
	tflog.Debug(ctx, "Creating sentinel version resource")
	resp.Diagnostics.Append(req.Plan.Get(ctx, &sentinelVersion)...)

	tflog.Debug(ctx, "Creating sentinel version resource", map[string]interface{}{
		"version":  sentinelVersion.Version.ValueString(),
		"url":      sentinelVersion.URL.ValueString(),
		"SHA":      sentinelVersion.SHA.ValueString(),
		"official": sentinelVersion.Official.ValueBool(),
		"enabled":  sentinelVersion.Enabled.ValueBool(),
		"beta":     sentinelVersion.Beta.ValueBool(),

		"deprecated":        sentinelVersion.Deprecated.ValueBool(),
		"deprecated_reason": sentinelVersion.DeprecatedReason.ValueString(),
		"archs":             sentinelVersion.Archs.ElementsAs(ctx, nil, false),
	})

	if resp.Diagnostics.HasError() {
		return
	}

	opts := tfe.AdminSentinelVersionCreateOptions{
		Version:          sentinelVersion.Version.ValueString(),
		URL:              sentinelVersion.URL.ValueString(),
		SHA:              sentinelVersion.SHA.ValueString(),
		Official:         tfe.Bool(sentinelVersion.Official.ValueBool()),
		Enabled:          tfe.Bool(sentinelVersion.Enabled.ValueBool()),
		Beta:             tfe.Bool(sentinelVersion.Beta.ValueBool()),
		Deprecated:       tfe.Bool(sentinelVersion.Deprecated.ValueBool()),
		DeprecatedReason: tfe.String(sentinelVersion.DeprecatedReason.ValueString()),
		Archs: func() []*tfe.ToolVersionArchitecture {
			archs, diags := convertToToolVersionArchitectures(ctx, sentinelVersion.Archs)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return nil
			}
			return archs
		}(),
	}
	tflog.Debug(ctx, "Creating sentinel version", map[string]interface{}{
		"version": sentinelVersion.Version.ValueString(),
	})

	v, err := r.config.Client.Admin.SentinelVersions.Create(ctx, opts)
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

	sentinelVersion.ID = types.StringValue(v.ID)
	sentinelVersion.Version = types.StringValue(v.Version)
	sentinelVersion.Official = types.BoolValue(v.Official)
	sentinelVersion.Enabled = types.BoolValue(v.Enabled)
	sentinelVersion.Beta = types.BoolValue(v.Beta)
	sentinelVersion.Deprecated = types.BoolValue(v.Deprecated)
	if v.DeprecatedReason != nil {
		sentinelVersion.DeprecatedReason = types.StringValue(*v.DeprecatedReason)
	} else {
		sentinelVersion.DeprecatedReason = types.StringNull()
	}
	if v.URL == "" {
		sentinelVersion.URL = types.StringNull()
	} else {
		sentinelVersion.URL = types.StringValue(v.URL)
	}
	if v.SHA == "" {
		sentinelVersion.SHA = types.StringNull()
	} else {
		sentinelVersion.SHA = types.StringValue(v.SHA)
	}
	sentinelVersion.Archs = convertAPIArchsToFrameworkSet(v.Archs)

	resp.Diagnostics.Append(resp.State.Set(ctx, &sentinelVersion)...)
}

func (r *sentinelVersionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var sentinelVersion modelAdminSentinelVersion
	resp.Diagnostics.Append(req.State.Get(ctx, &sentinelVersion)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Read configuration of sentinel version", map[string]interface{}{
		"id": sentinelVersion.ID.ValueString()})

	v, err := r.config.Client.Admin.SentinelVersions.Read(ctx, sentinelVersion.ID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading sentinel version",
			fmt.Sprintf("Could not read sentinel version %s: %v", sentinelVersion.ID.ValueString(), err),
		)
		return
	}

	sentinelVersion.ID = types.StringValue(v.ID)
	sentinelVersion.Version = types.StringValue(v.Version)
	sentinelVersion.Official = types.BoolValue(v.Official)
	sentinelVersion.Enabled = types.BoolValue(v.Enabled)
	sentinelVersion.Beta = types.BoolValue(v.Beta)
	sentinelVersion.Deprecated = types.BoolValue(v.Deprecated)
	if v.DeprecatedReason != nil {
		sentinelVersion.DeprecatedReason = types.StringValue(*v.DeprecatedReason)
	} else {
		sentinelVersion.DeprecatedReason = types.StringNull()
	}
	if v.URL == "" {
		sentinelVersion.URL = types.StringNull()
	} else {
		sentinelVersion.URL = types.StringValue(v.URL)
	}
	if v.SHA == "" {
		sentinelVersion.SHA = types.StringNull()
	} else {
		sentinelVersion.SHA = types.StringValue(v.SHA)
	}
	sentinelVersion.Archs = convertAPIArchsToFrameworkSet(v.Archs)

	resp.Diagnostics.Append(resp.State.Set(ctx, &sentinelVersion)...)
}

func (r *sentinelVersionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var sentinelVersion modelAdminSentinelVersion
	resp.Diagnostics.Append(req.Plan.Get(ctx, &sentinelVersion)...)

	var state modelAdminSentinelVersion
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sentinelVersion.ID = state.ID

	tflog.Debug(ctx, "Updating sentinel version resource", map[string]interface{}{
		"id": sentinelVersion.ID.ValueString(),
	})

	opts := tfe.AdminSentinelVersionUpdateOptions{
		Version:          tfe.String(sentinelVersion.Version.ValueString()),
		URL:              stringOrNil(sentinelVersion.URL.ValueString()),
		SHA:              stringOrNil(sentinelVersion.SHA.ValueString()),
		Official:         tfe.Bool(sentinelVersion.Official.ValueBool()),
		Enabled:          tfe.Bool(sentinelVersion.Enabled.ValueBool()),
		Beta:             tfe.Bool(sentinelVersion.Beta.ValueBool()),
		Deprecated:       tfe.Bool(sentinelVersion.Deprecated.ValueBool()),
		DeprecatedReason: tfe.String(sentinelVersion.DeprecatedReason.ValueString()),
		Archs: func() []*tfe.ToolVersionArchitecture {
			archs, diags := convertToToolVersionArchitectures(ctx, sentinelVersion.Archs)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return nil
			}
			return archs
		}(),
	}

	v, err := r.config.Client.Admin.SentinelVersions.Update(ctx, sentinelVersion.ID.ValueString(), opts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating sentinel version",
			fmt.Sprintf("Could not update sentinel version %s: %v", sentinelVersion.ID.ValueString(), err),
		)
		return
	}

	sentinelVersion.ID = types.StringValue(v.ID)
	sentinelVersion.Version = types.StringValue(v.Version)
	sentinelVersion.Official = types.BoolValue(v.Official)
	sentinelVersion.Enabled = types.BoolValue(v.Enabled)
	sentinelVersion.Beta = types.BoolValue(v.Beta)
	sentinelVersion.Deprecated = types.BoolValue(v.Deprecated)
	if v.DeprecatedReason != nil {
		sentinelVersion.DeprecatedReason = types.StringValue(*v.DeprecatedReason)
	} else {
		sentinelVersion.DeprecatedReason = types.StringNull()
	}
	if v.URL == "" {
		sentinelVersion.URL = types.StringNull()
	} else {
		sentinelVersion.URL = types.StringValue(v.URL)
	}
	if v.SHA == "" {
		sentinelVersion.SHA = types.StringNull()
	} else {
		sentinelVersion.SHA = types.StringValue(v.SHA)
	}
	sentinelVersion.Archs = convertAPIArchsToFrameworkSet(v.Archs)

	resp.Diagnostics.Append(resp.State.Set(ctx, &sentinelVersion)...)
}

func (r *sentinelVersionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var sentinelVersion modelAdminSentinelVersion
	resp.Diagnostics.Append(req.State.Get(ctx, &sentinelVersion)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Deleting sentinel version", map[string]interface{}{
		"id": sentinelVersion.ID.ValueString(),
	})

	err := r.config.Client.Admin.SentinelVersions.Delete(ctx, sentinelVersion.ID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			tflog.Debug(ctx, "sentinel version not found, skipping deletion", map[string]interface{}{
				"id": sentinelVersion.ID.ValueString(),
			})
			return
		}
		resp.Diagnostics.AddError(
			"Error deleting sentinel version",
			fmt.Sprintf("Could not delete sentinel version %s: %v", sentinelVersion.ID.ValueString(), err),
		)
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *sentinelVersionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var id string
	// Splitting by '-' and checking if the first elem is equal to tool
	// determines if the string is a tool version ID
	s := strings.Split(req.ID, "-")
	if s[0] != "tool" {
		versionID, err := fetchSentinelVersionID(req.ID, r.config.Client)
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

		id = versionID
	} else {
		id = req.ID
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
