// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	tfe "github.com/hashicorp/go-tfe/v2"
	"github.com/hashicorp/go-tfe/v2/api/admin"
	"github.com/hashicorp/go-tfe/v2/api/models"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                   = &TFPolicyVersionResource{}
	_ resource.ResourceWithConfigure      = &TFPolicyVersionResource{}
	_ resource.ResourceWithImportState    = &TFPolicyVersionResource{}
	_ resource.ResourceWithValidateConfig = &TFPolicyVersionResource{}
	_ resource.ResourceWithModifyPlan     = &TFPolicyVersionResource{}
)

// minTFEVersionTFPolicyVersions is the first Terraform Enterprise release with
// the admin Terraform policy plugin versions API.
const minTFEVersionTFPolicyVersions = "2.1.0"

type TFPolicyVersionResource struct {
	config ConfiguredClient
}

func NewTFPolicyVersionResource() resource.Resource {
	return &TFPolicyVersionResource{}
}

type modelAdminTFPolicyVersion struct {
	ID               types.String `tfsdk:"id"`
	Version          types.String `tfsdk:"version"`
	Official         types.Bool   `tfsdk:"official"`
	Enabled          types.Bool   `tfsdk:"enabled"`
	Beta             types.Bool   `tfsdk:"beta"`
	Deprecated       types.Bool   `tfsdk:"deprecated"`
	DeprecatedReason types.String `tfsdk:"deprecated_reason"`
	Archs            types.Set    `tfsdk:"archs"`
}

func (r *TFPolicyVersionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tfpolicy_version"
}

func (r *TFPolicyVersionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages Terraform policy plugin versions on Terraform Enterprise " + minTFEVersionTFPolicyVersions + " or later. Requires administrator access. " +
			"Configure binary builds using archs; deprecated top-level URL and checksum fields are not exposed. " +
			"Versions and builds can be updated in place. Official or in-use versions cannot be deleted, and in-use versions cannot be disabled.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The external tool-* ID of the Terraform policy plugin version.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"version": schema.StringAttribute{
				Description: "The unique semantic version of this Terraform policy plugin release.",
				Required:    true,
				Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"official": schema.BoolAttribute{
				Description: "Whether this is an official HashiCorp release. Defaults to false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether this version can be selected for use. Defaults to true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"beta": schema.BoolAttribute{
				Description: "Whether this version is a beta/pre-release. Defaults to false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"deprecated": schema.BoolAttribute{
				Description: "Whether this version is deprecated. Defaults to false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"deprecated_reason": schema.StringAttribute{
				Description: "Explanation for deprecation. Creating a deprecated version requires a non-blank reason. " +
					"On update, omitting the reason lets the server supply a default. " +
					"When deprecated is false, omit this field or use null or an empty string.",
				Optional: true,
				Computed: true,
			},
			"archs": schema.SetNestedAttribute{
				Description: "The complete set of binary builds. Changes replace the full set. At least one build is required, " +
					"and each operating system/CPU architecture pair must be unique.",
				Required:     true,
				Validators:   []validator.Set{setvalidator.SizeAtLeast(1)},
				NestedObject: toolArchitectureNestedObject(),
			},
		},
	}
}

func (r *TFPolicyVersionResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config modelAdminTFPolicyVersion
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// deprecated defaults to false, so an omitted value is validated as false.
	deprecated := config.Deprecated
	if deprecated.IsNull() {
		deprecated = types.BoolValue(false)
	}
	resp.Diagnostics.Append(validateToolVersionDeprecation(deprecated, config.DeprecatedReason)...)
	archs, diags := toolArchitecturesFromSet(ctx, config.Archs)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}
	resp.Diagnostics.Append(validateUniqueToolArchitectures(archs)...)
}

func (r *TFPolicyVersionResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}
	var config, plan modelAdminTFPolicyVersion
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() || !config.DeprecatedReason.IsNull() {
		return
	}
	if plan.Deprecated.Equal(types.BoolValue(false)) {
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("deprecated_reason"), types.StringNull())...)
		return
	}
	if !plan.Deprecated.Equal(types.BoolValue(true)) || req.State.Raw.IsNull() {
		return
	}
	var prior modelAdminTFPolicyVersion
	resp.Diagnostics.Append(req.State.Get(ctx, &prior)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// A generated reason may change when deprecating or renaming a release.
	if !prior.Deprecated.Equal(plan.Deprecated) || !prior.Version.Equal(plan.Version) {
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("deprecated_reason"), types.StringUnknown())...)
	}
}

func (r *TFPolicyVersionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(ConfiguredClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data Type", fmt.Sprintf("Expected ConfiguredClient, got: %T", req.ProviderData))
		return
	}
	if client.ClientV2 == nil {
		resp.Diagnostics.AddError("Missing API Client", "Terraform policy plugin versions require the go-tfe v2 client.")
		return
	}
	r.config = client
}

func (r *TFPolicyVersionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan modelAdminTFPolicyVersion
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !r.supportsTFPolicyVersions(&resp.Diagnostics) {
		return
	}
	attributes, diags := tfpolicyVersionAttributes(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if plan.Deprecated.ValueBool() && (plan.DeprecatedReason.IsNull() || plan.DeprecatedReason.IsUnknown()) {
		resp.Diagnostics.AddAttributeError(path.Root("deprecated_reason"), "Deprecation Reason Required",
			"Creating a deprecated Terraform policy plugin version requires a non-blank deprecated_reason.")
	}
	if resp.Diagnostics.HasError() {
		return
	}
	data := models.NewAdminTfpolicyVersionCreate()
	data.SetAttributes(attributes)
	data.SetTypeEscaped(ptr(models.TFPOLICYVERSIONS_ADMINTFPOLICYVERSIONS_TYPE))
	body := models.NewAdminTfpolicyVersionCreateEnvelope()
	body.SetData(data)
	result, err := r.config.ClientV2.API.Admin().TfpolicyVersions().Post(ctx, body, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Terraform Policy Plugin Version", apiErrorDetail(err))
		return
	}
	state, err := tfpolicyVersionModel(ctx, result, "")
	if err != nil {
		resp.Diagnostics.AddError("Invalid Terraform Policy Plugin Version Response", err.Error())
		return
	}
	state.DeprecatedReason = preserveEmptyDeprecatedReason(state.Deprecated, state.DeprecatedReason, plan.DeprecatedReason)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *TFPolicyVersionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state modelAdminTFPolicyVersion
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := r.config.ClientV2.API.Admin().TfpolicyVersions().ById(state.ID.ValueString()).Get(ctx, nil)
	if errors.Is(err, tfe.ErrNotFound) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Terraform Policy Plugin Version", apiErrorDetail(err))
		return
	}
	updated, err := tfpolicyVersionModel(ctx, result, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid Terraform Policy Plugin Version Response", err.Error())
		return
	}
	updated.DeprecatedReason = preserveEmptyDeprecatedReason(updated.Deprecated, updated.DeprecatedReason, state.DeprecatedReason)
	resp.Diagnostics.Append(resp.State.Set(ctx, &updated)...)
}

func (r *TFPolicyVersionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, prior modelAdminTFPolicyVersion
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &prior)...)
	if resp.Diagnostics.HasError() {
		return
	}
	attributes, diags := tfpolicyVersionAttributes(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if plan.Archs.Equal(prior.Archs) {
		attributes.SetArchs(nil)
	}
	data := models.NewAdminTfpolicyVersions()
	data.SetAttributes(attributes)
	data.SetTypeEscaped(ptr(models.TFPOLICYVERSIONS_ADMINTFPOLICYVERSIONS_TYPE))
	body := models.NewAdminTfpolicyVersionsEnvelope()
	body.SetData(data)
	result, err := r.config.ClientV2.API.Admin().TfpolicyVersions().ById(prior.ID.ValueString()).Patch(ctx, body, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Terraform Policy Plugin Version", apiErrorDetail(err))
		return
	}
	state, err := tfpolicyVersionModel(ctx, result, prior.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid Terraform Policy Plugin Version Response", err.Error())
		return
	}
	state.DeprecatedReason = preserveEmptyDeprecatedReason(state.Deprecated, state.DeprecatedReason, plan.DeprecatedReason)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *TFPolicyVersionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state modelAdminTFPolicyVersion
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.config.ClientV2.API.Admin().TfpolicyVersions().ById(state.ID.ValueString()).Delete(ctx, nil)
	if err != nil && !errors.Is(err, tfe.ErrNotFound) {
		resp.Diagnostics.AddError("Error Deleting Terraform Policy Plugin Version", apiErrorDetail(err))
		return
	}
	resp.State.RemoveResource(ctx)
}

func (r *TFPolicyVersionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id := req.ID
	suffix, isID := strings.CutPrefix(id, "tool-")
	if strings.TrimSpace(id) == "" || (isID && suffix == "") {
		resp.Diagnostics.AddError("Invalid Terraform Policy Plugin Version Import", "Provide a tool-* ID or an exact semantic version.")
		return
	}
	if !r.supportsTFPolicyVersions(&resp.Diagnostics) {
		return
	}
	if !isID {
		var err error
		id, err = r.findVersionID(ctx, id)
		if err != nil {
			resp.Diagnostics.AddError("Error Importing Terraform Policy Plugin Version", err.Error())
			return
		}
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}

// supportsTFPolicyVersions reports an error when the connected Terraform
// Enterprise predates the admin Terraform policy plugin versions API.
func (r *TFPolicyVersionResource) supportsTFPolicyVersions(d *diag.Diagnostics) bool {
	meets, err := r.config.MeetsMinRemoteTFEVersion(minTFEVersionTFPolicyVersions)
	if err != nil {
		d.AddError("Error checking minimum Terraform Enterprise version",
			fmt.Sprintf("Could not determine whether Terraform Enterprise version %s meets the minimum required version %s: %v",
				r.config.RemoteTFEVersion(), minTFEVersionTFPolicyVersions, err))
		return false
	}
	if !meets {
		d.AddError("Terraform Enterprise version does not support Terraform policy plugin versions",
			fmt.Sprintf("tfe_tfpolicy_version requires Terraform Enterprise %s or later. This instance reports %s.",
				minTFEVersionTFPolicyVersions, r.config.RemoteTFEVersion()))
		return false
	}
	return true
}

func (r *TFPolicyVersionResource) findVersionID(ctx context.Context, version string) (string, error) {
	page := int32(1)
	var id string
	for {
		result, err := r.config.ClientV2.API.Admin().TfpolicyVersions().Get(ctx, withQueryParams(
			&admin.TfpolicyVersionsRequestBuilderGetQueryParameters{Filterversion: &version, Pagenumber: &page}))
		if err != nil {
			return "", fmt.Errorf("could not list Terraform policy plugin versions: %s", apiErrorDetail(err))
		}
		if result == nil || result.GetData() == nil {
			return "", fmt.Errorf("Terraform policy plugin versions response is missing data")
		}
		for _, item := range result.GetData() {
			if item == nil || item.GetAttributes() == nil || item.GetAttributes().GetVersion() == nil {
				return "", fmt.Errorf("Terraform policy plugin versions response contains an invalid version")
			}
			if *item.GetAttributes().GetVersion() != version {
				continue
			}
			if item.GetId() == nil || !isToolVersionID(*item.GetId()) {
				return "", fmt.Errorf("Terraform policy plugin version %q has an invalid external ID", version)
			}
			if id != "" {
				return "", fmt.Errorf("multiple Terraform policy plugin versions match %q", version)
			}
			id = *item.GetId()
		}
		next := nextPageFromMeta(result.GetMeta())
		if next == nil {
			break
		}
		if *next <= page {
			return "", fmt.Errorf("Terraform policy plugin versions response has invalid pagination")
		}
		page = *next
	}
	if id == "" {
		return "", fmt.Errorf("Terraform policy plugin version %q was not found", version)
	}
	return id, nil
}

func tfpolicyVersionAttributes(ctx context.Context, plan modelAdminTFPolicyVersion) (*models.AdminTfpolicyVersions_attributes, diag.Diagnostics) {
	diags := validateToolVersionDeprecation(plan.Deprecated, plan.DeprecatedReason)
	diags.Append(validateToolVersionRequest(plan.Version, plan.Official, plan.Enabled, plan.Beta, plan.Deprecated)...)
	builds, buildDiags := tfpolicyVersionBuilds(ctx, plan.Archs)
	diags.Append(buildDiags...)
	if diags.HasError() {
		return nil, diags
	}
	attributes := models.NewAdminTfpolicyVersions_attributes()
	attributes.SetVersion(plan.Version.ValueStringPointer())
	attributes.SetArchs(builds)
	attributes.SetOfficial(plan.Official.ValueBoolPointer())
	attributes.SetEnabled(plan.Enabled.ValueBoolPointer())
	attributes.SetBeta(plan.Beta.ValueBoolPointer())
	attributes.SetDeprecated(plan.Deprecated.ValueBoolPointer())
	if !plan.DeprecatedReason.IsNull() && !plan.DeprecatedReason.IsUnknown() {
		attributes.SetDeprecatedReason(plan.DeprecatedReason.ValueStringPointer())
	}
	return attributes, diags
}

func tfpolicyVersionBuilds(ctx context.Context, value types.Set) ([]models.AdminTfpolicyVersionArchitectureable, diag.Diagnostics) {
	archs, diags := toolArchitecturesFromSet(ctx, value)
	if diags.HasError() {
		return nil, diags
	}
	diags.Append(validateToolArchitectures(archs)...)
	if diags.HasError() {
		return nil, diags
	}
	builds := make([]models.AdminTfpolicyVersionArchitectureable, 0, len(archs))
	for _, arch := range archs {
		cpu := parseEnumPtr[models.AdminTfpolicyVersionArchitecture_arch](models.ParseAdminTfpolicyVersionArchitecture_arch, arch.Arch.ValueString())
		if cpu == nil {
			diags.AddAttributeError(path.Root("archs"), "Invalid Terraform Policy Plugin Architecture",
				fmt.Sprintf("CPU architecture %q is not supported by the Terraform policy plugin versions API.", arch.Arch.ValueString()))
			return nil, diags
		}
		build := models.NewAdminTfpolicyVersionArchitecture()
		build.SetOs(ptr(models.LINUX_ADMINTFPOLICYVERSIONARCHITECTURE_OS))
		build.SetArch(cpu)
		build.SetUrl(arch.URL.ValueStringPointer())
		build.SetSha(arch.Sha.ValueStringPointer())
		builds = append(builds, build)
	}
	return builds, diags
}

func tfpolicyVersionModel(ctx context.Context, result models.AdminTfpolicyVersionsEnvelopeable, expectedID string) (modelAdminTFPolicyVersion, error) {
	var state modelAdminTFPolicyVersion
	if result == nil || result.GetData() == nil {
		return state, fmt.Errorf("API response is missing data")
	}
	data := result.GetData()
	if data.GetId() == nil || !isToolVersionID(*data.GetId()) ||
		(expectedID != "" && *data.GetId() != expectedID) {
		return state, fmt.Errorf("API response has an invalid or unexpected external ID")
	}
	if data.GetTypeEscaped() == nil || *data.GetTypeEscaped() != models.TFPOLICYVERSIONS_ADMINTFPOLICYVERSIONS_TYPE {
		return state, fmt.Errorf("API response has an invalid resource type")
	}
	a := data.GetAttributes()
	if a == nil || a.GetVersion() == nil || strings.TrimSpace(*a.GetVersion()) == "" ||
		a.GetOfficial() == nil || a.GetEnabled() == nil || a.GetBeta() == nil || a.GetDeprecated() == nil {
		return state, fmt.Errorf("API response is missing required release attributes")
	}
	archs := make([]ToolArchitecture, 0, len(a.GetArchs()))
	for _, build := range a.GetArchs() {
		if build == nil || build.GetUrl() == nil || build.GetSha() == nil || build.GetOs() == nil || build.GetArch() == nil {
			return state, fmt.Errorf("API response contains an incomplete binary build")
		}
		archs = append(archs, ToolArchitecture{
			URL: types.StringPointerValue(build.GetUrl()), Sha: types.StringPointerValue(build.GetSha()),
			OS: types.StringValue(build.GetOs().String()), Arch: types.StringValue(build.GetArch().String()),
		})
	}
	set, diags := toolArchitecturesToSet(ctx, archs)
	if diags.HasError() {
		return state, fmt.Errorf("decoding binary builds: %v", diags)
	}
	if _, diags := tfpolicyVersionBuilds(ctx, set); diags.HasError() {
		return state, fmt.Errorf("API response contains invalid binary builds: %v", diags)
	}
	state = modelAdminTFPolicyVersion{
		ID: types.StringPointerValue(data.GetId()), Version: types.StringPointerValue(a.GetVersion()),
		Official: types.BoolPointerValue(a.GetOfficial()), Enabled: types.BoolPointerValue(a.GetEnabled()),
		Beta: types.BoolPointerValue(a.GetBeta()), Deprecated: types.BoolPointerValue(a.GetDeprecated()),
		DeprecatedReason: types.StringNull(), Archs: set,
	}
	if state.Deprecated.ValueBool() {
		if a.GetDeprecatedReason() == nil || strings.TrimSpace(*a.GetDeprecatedReason()) == "" {
			return modelAdminTFPolicyVersion{}, fmt.Errorf("API response is missing the deprecation reason")
		}
		state.DeprecatedReason = types.StringPointerValue(a.GetDeprecatedReason())
	}
	return state, nil
}
