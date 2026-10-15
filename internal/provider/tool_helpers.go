// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"regexp"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"strings"

	tfe "github.com/hashicorp/go-tfe"
	sdkValidation "github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	customValidators "github.com/hashicorp/terraform-provider-tfe/internal/provider/validators"
)

// fetchTerraformVersionID returns a Terraform Version ID for the given Terraform version number
func fetchTerraformVersionID(version string, client *tfe.Client) (string, error) {
	versions, err := client.Admin.TerraformVersions.List(ctx, &tfe.AdminTerraformVersionsListOptions{
		Filter: version,
	})
	if err != nil {
		return "", fmt.Errorf("error reading Terraform versions: %w", err)
	}

	// filter[version] returns 1 item or 0, if however
	// the number of versions returned is greater than 1,
	// we can assume the API doesn't support the filter[version] query param
	// and so we'll use a fallback search mechanism
	switch len(versions.Items) {
	case 0:
		return "", fmt.Errorf("terraform version not found")
	case 1:
		return versions.Items[0].ID, nil
	default:
		options := &tfe.AdminTerraformVersionsListOptions{}
		for {
			for _, v := range versions.Items {
				if v.Version == version {
					return v.ID, nil
				}
			}

			if versions.CurrentPage >= versions.TotalPages {
				break
			}

			options.PageNumber = versions.NextPage

			versions, err = client.Admin.TerraformVersions.List(ctx, options)
			if err != nil {
				return "", fmt.Errorf("error reading Terraform Versions: %w", err)
			}
		}
	}

	return "", fmt.Errorf("terraform version not found")
}

// fetchSentinelVersionID returns a Sentinel Version ID for the given Sentinel version number
func fetchSentinelVersionID(version string, client *tfe.Client) (string, error) {
	versions, err := client.Admin.SentinelVersions.List(ctx, &tfe.AdminSentinelVersionsListOptions{
		Filter: version,
	})
	if err != nil {
		return "", fmt.Errorf("error reading Sentinel versions: %w", err)
	}

	// filter[version] returns 1 item or 0, if however
	// the number of versions returned is greater than 1,
	// we can assume the API doesn't support the filter[version] query param
	// and so we'll use a fallback search mechanism
	switch len(versions.Items) {
	case 0:
		return "", fmt.Errorf("sentinel version not found")
	case 1:
		return versions.Items[0].ID, nil
	default:
		options := &tfe.AdminSentinelVersionsListOptions{}
		for {
			for _, v := range versions.Items {
				if v.Version == version {
					return v.ID, nil
				}
			}

			if versions.CurrentPage >= versions.TotalPages {
				break
			}

			options.PageNumber = versions.NextPage

			versions, err = client.Admin.SentinelVersions.List(ctx, options)
			if err != nil {
				return "", fmt.Errorf("error reading Sentinel Versions: %w", err)
			}
		}
	}

	return "", fmt.Errorf("sentinel version not found")
}

// fetchOPAVersionID returns a OPA Version ID for the given OPA version number
func fetchOPAVersionID(version string, client *tfe.Client) (string, error) {
	versions, err := client.Admin.OPAVersions.List(ctx, &tfe.AdminOPAVersionsListOptions{
		Filter: version,
	})
	if err != nil {
		return "", fmt.Errorf("error reading OPA versions: %w", err)
	}

	// filter[version] returns 1 item or 0, if however
	// the number of versions returned is greater than 1,
	// we can assume the API doesn't support the filter[version] query param
	// and so we'll use a fallback search mechanism
	switch len(versions.Items) {
	case 0:
		return "", fmt.Errorf("OPA version not found")
	case 1:
		return versions.Items[0].ID, nil
	default:
		options := &tfe.AdminOPAVersionsListOptions{}
		for {
			for _, v := range versions.Items {
				if v.Version == version {
					return v.ID, nil
				}
			}

			if versions.CurrentPage >= versions.TotalPages {
				break
			}

			options.PageNumber = versions.NextPage

			versions, err = client.Admin.OPAVersions.List(ctx, options)
			if err != nil {
				return "", fmt.Errorf("error reading OPA Versions: %w", err)
			}
		}
	}

	return "", fmt.Errorf("OPA version not found")
}

func stringOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// ToolArchitecture represents the common architecture structure for all tool versions
type ToolArchitecture struct {
	URL  types.String `tfsdk:"url"`
	Sha  types.String `tfsdk:"sha"`
	OS   types.String `tfsdk:"os"`
	Arch types.String `tfsdk:"arch"`
}

// ObjectTypeForArchitectures returns the standard object type definition for architecture objects
func ObjectTypeForArchitectures() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"url":  types.StringType,
			"sha":  types.StringType,
			"os":   types.StringType,
			"arch": types.StringType,
		},
	}
}

var toolVersionSHARegexp = regexp.MustCompile(`^[a-f0-9]{64}$`)

var (
	toolVersionSupportedOS    = []string{"linux"}
	toolVersionSupportedArchs = []string{"amd64", "arm64", "s390x"}
)

// isToolVersionID reports whether id is a non-empty tool-* external ID.
func isToolVersionID(id string) bool {
	suffix, ok := strings.CutPrefix(id, "tool-")
	return ok && suffix != ""
}

// toolArchitectureNestedObject returns the schema for a single tool version binary build.
func toolArchitectureNestedObject() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				Description: "HTTP or HTTPS download URL for this binary build.",
				Required:    true,
				Validators:  []validator.String{customValidators.IsURLWithHTTPorHTTPS()},
			},
			"sha": schema.StringAttribute{
				Description: "SHA-256 checksum (64 lowercase hexadecimal characters) of this binary build.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(toolVersionSHARegexp, "must be a SHA-256 checksum"),
				},
			},
			"os": schema.StringAttribute{
				Description: "Operating system for this build: " + strings.Join(toolVersionSupportedOS, ", ") + ".",
				Required:    true,
				Validators:  []validator.String{stringvalidator.OneOf(toolVersionSupportedOS...)},
			},
			"arch": schema.StringAttribute{
				Description: "CPU architecture for this build: " + strings.Join(toolVersionSupportedArchs, ", ") + ".",
				Required:    true,
				Validators:  []validator.String{stringvalidator.OneOf(toolVersionSupportedArchs...)},
			},
		},
	}
}

// toolArchitecturesFromSet decodes an archs set. Null or unknown sets return no builds.
func toolArchitecturesFromSet(ctx context.Context, set types.Set) ([]ToolArchitecture, diag.Diagnostics) {
	if set.IsNull() || set.IsUnknown() {
		return nil, nil
	}
	var archs []ToolArchitecture
	diags := set.ElementsAs(ctx, &archs, true)
	return archs, diags
}

// toolArchitecturesToSet encodes builds as an archs set.
func toolArchitecturesToSet(ctx context.Context, archs []ToolArchitecture) (types.Set, diag.Diagnostics) {
	return types.SetValueFrom(ctx, ObjectTypeForArchitectures(), archs)
}

// validateUniqueToolArchitectures rejects builds that repeat a known os/arch pair.
func validateUniqueToolArchitectures(archs []ToolArchitecture) diag.Diagnostics {
	var diags diag.Diagnostics
	seen := make(map[string]bool, len(archs))
	for _, arch := range archs {
		if arch.OS.IsNull() || arch.OS.IsUnknown() || arch.Arch.IsNull() || arch.Arch.IsUnknown() {
			continue
		}
		key := arch.OS.ValueString() + "/" + arch.Arch.ValueString()
		if seen[key] {
			diags.AddAttributeError(path.Root("archs"), "Duplicate Architecture",
				fmt.Sprintf("The %s build is configured more than once. Each operating system/CPU architecture pair must be unique.", key))
		}
		seen[key] = true
	}
	return diags
}

// validateToolVersionDeprecation rejects a reason on a non-deprecated version and a blank
// reason on a deprecated one. Null or unknown values are left for the API or plan to resolve.
func validateToolVersionDeprecation(deprecated types.Bool, reason types.String) diag.Diagnostics {
	var diags diag.Diagnostics
	if deprecated.IsNull() || deprecated.IsUnknown() || reason.IsNull() || reason.IsUnknown() {
		return diags
	}
	value := reason.ValueString()
	if !deprecated.ValueBool() && value != "" {
		diags.AddAttributeError(path.Root("deprecated_reason"), "Invalid Deprecation Reason",
			"When deprecated is false, omit deprecated_reason or set it to null or an empty string.")
	}
	if deprecated.ValueBool() && strings.TrimSpace(value) == "" {
		diags.AddAttributeError(path.Root("deprecated_reason"), "Invalid Deprecation Reason",
			"When deprecated is true, provide a non-blank reason or omit it to use the server's update default.")
	}
	return diags
}

// validateToolArchitectures checks builds before they are sent to the API: at least one
// build, each fully known and supported, and no repeated os/arch pair.
func validateToolArchitectures(archs []ToolArchitecture) diag.Diagnostics {
	var diags diag.Diagnostics
	if len(archs) == 0 {
		diags.AddAttributeError(path.Root("archs"), "Invalid Architectures", "archs must contain at least one known binary build.")
		return diags
	}
	for _, arch := range archs {
		if !validToolArchitecture(arch) {
			diags.AddAttributeError(path.Root("archs"), "Invalid Architecture",
				fmt.Sprintf("Each build requires a known HTTP or HTTPS URL, a SHA-256 checksum, an os of %s, and an arch of %s.",
					strings.Join(toolVersionSupportedOS, ", "), strings.Join(toolVersionSupportedArchs, ", ")))
			return diags
		}
	}
	diags.Append(validateUniqueToolArchitectures(archs)...)
	return diags
}

func validToolArchitecture(arch ToolArchitecture) bool {
	for _, value := range []types.String{arch.URL, arch.Sha, arch.OS, arch.Arch} {
		if value.IsNull() || value.IsUnknown() {
			return false
		}
	}
	_, urlErrors := sdkValidation.IsURLWithHTTPorHTTPS(arch.URL.ValueString(), "url")
	return len(urlErrors) == 0 &&
		toolVersionSHARegexp.MatchString(arch.Sha.ValueString()) &&
		slices.Contains(toolVersionSupportedOS, arch.OS.ValueString()) &&
		slices.Contains(toolVersionSupportedArchs, arch.Arch.ValueString())
}

// validateToolVersionRequest checks that the version and release flags are known before applying.
func validateToolVersionRequest(version types.String, official, enabled, beta, deprecated types.Bool) diag.Diagnostics {
	var diags diag.Diagnostics
	if version.IsNull() || version.IsUnknown() || strings.TrimSpace(version.ValueString()) == "" {
		diags.AddAttributeError(path.Root("version"), "Invalid Version", "version must be known and non-blank.")
	}
	flags := []struct {
		name  string
		value types.Bool
	}{{"official", official}, {"enabled", enabled}, {"beta", beta}, {"deprecated", deprecated}}
	for _, flag := range flags {
		if flag.value.IsNull() || flag.value.IsUnknown() {
			diags.AddAttributeError(path.Root(flag.name), "Invalid Release Flag", flag.name+" must be known before applying.")
		}
	}
	return diags
}

// preserveEmptyDeprecatedReason keeps a configured empty reason on a non-deprecated version,
// because the API returns null in that case.
func preserveEmptyDeprecatedReason(deprecated types.Bool, current, configured types.String) types.String {
	if deprecated.Equal(types.BoolValue(false)) && configured.Equal(types.StringValue("")) {
		return configured
	}
	return current
}

// convertAPIArchsToFrameworkSet converts API architecture objects to a Framework Set value
func convertAPIArchsToFrameworkSet(apiArchs []*tfe.ToolVersionArchitecture) types.Set {
	archObjectType := ObjectTypeForArchitectures()

	// Return empty set rather than null set
	if len(apiArchs) == 0 {
		return types.SetValueMust(archObjectType, []attr.Value{})
	}

	// Rest of function remains the same
	archValues := make([]attr.Value, len(apiArchs))
	for i, arch := range apiArchs {
		archValues[i] = types.ObjectValueMust(
			archObjectType.AttrTypes,
			map[string]attr.Value{
				"url":  types.StringValue(arch.URL),
				"sha":  types.StringValue(arch.Sha),
				"os":   types.StringValue(arch.OS),
				"arch": types.StringValue(arch.Arch),
			},
		)
	}

	return types.SetValueMust(archObjectType, archValues)
}

// convertToToolVersionArchitectures converts Framework types.Set to API architecture objects
func convertToToolVersionArchitectures(ctx context.Context, archs types.Set) ([]*tfe.ToolVersionArchitecture, diag.Diagnostics) {
	if archs.IsNull() || archs.IsUnknown() {
		return nil, nil
	}

	var diags diag.Diagnostics
	var archModels []ToolArchitecture

	diags.Append(archs.ElementsAs(ctx, &archModels, false)...)
	if diags.HasError() {
		return nil, diags
	}

	result := make([]*tfe.ToolVersionArchitecture, 0, len(archModels))
	for _, model := range archModels {
		result = append(result, &tfe.ToolVersionArchitecture{
			URL:  model.URL.ValueString(),
			Sha:  model.Sha.ValueString(),
			OS:   model.OS.ValueString(),
			Arch: model.Arch.ValueString(),
		})
	}

	return result, nil
}

// PreserveAMD64ArchsOnChange creates a plan modifier that preserves AMD64 architecture entries
// when top-level URL or SHA changes, to be used across all tool version resources
func PreserveAMD64ArchsOnChange() planmodifier.Set {
	return &preserveAMD64ArchsModifier{}
}

type preserveAMD64ArchsModifier struct{}

func (m *preserveAMD64ArchsModifier) Description(ctx context.Context) string {
	return "Preserves AMD64 architecture entries when top-level URL or SHA changes"
}

func (m *preserveAMD64ArchsModifier) MarkdownDescription(ctx context.Context) string {
	return "Preserves AMD64 architecture entries when top-level URL or SHA changes"
}

// PlanModifySet modifies the plan to ensure AMD64 architecture entries are preserved
func (m *preserveAMD64ArchsModifier) PlanModifySet(ctx context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
	// Skip if we're destroying the resource or no state
	if req.Plan.Raw.IsNull() || req.State.Raw.IsNull() || req.StateValue.IsNull() {
		return
	}

	// skip if archs is set in the config
	if !req.ConfigValue.IsUnknown() && !req.ConfigValue.IsNull() {
		return
	}

	var configURL, stateURL, configSHA, stateSHA types.String

	req.Config.GetAttribute(ctx, path.Root("url"), &configURL)
	req.State.GetAttribute(ctx, path.Root("url"), &stateURL)
	req.Config.GetAttribute(ctx, path.Root("sha"), &configSHA)
	req.State.GetAttribute(ctx, path.Root("sha"), &stateSHA)

	urlChanged := !configURL.Equal(stateURL)
	shaChanged := !configSHA.Equal(stateSHA)

	// If neither URL nor SHA is changing, do nothing
	if !urlChanged && !shaChanged {
		tflog.Debug(ctx, "No changes to URL or SHA, skipping AMD64 architecture preservation")
		return
	}

	stateArchs := req.StateValue
	planArchs := req.PlanValue

	var stateArchsList []ToolArchitecture
	diags := stateArchs.ElementsAs(ctx, &stateArchsList, false)
	if diags.HasError() {
		return
	}

	var planArchsList []ToolArchitecture
	diags = planArchs.ElementsAs(ctx, &planArchsList, false)
	if diags.HasError() {
		tflog.Debug(ctx, "Error extracting plan architectures", map[string]interface{}{
			"diagnostics": diags,
		})
		return
	}

	tflog.Debug(ctx, "plan architectures", map[string]interface{}{
		"planArchsList":  planArchsList,
		"stateArchsList": stateArchsList,
		"urlChanged":     urlChanged,
		"shaChanged":     shaChanged,
		"stateURL":       configURL,
		"planURL":        stateURL,
		"stateSHA":       configSHA,
		"planSHA":        stateSHA,
	})

	for _, arch := range planArchsList {
		if arch.Arch.ValueString() == "amd64" {
			tflog.Debug(ctx, "Found AMD64 architecture in plan", map[string]interface{}{
				"url":  arch.URL.ValueString(),
				"sha":  arch.Sha.ValueString(),
				"os":   arch.OS.ValueString(),
				"arch": arch.Arch.ValueString(),
			})

			if urlChanged {
				arch.URL = configURL
			}
			if shaChanged {
				arch.Sha = configSHA
			}

			// Update the plan with the modified AMD64 arch
			archObjectType := ObjectTypeForArchitectures()
			attrValue := types.ObjectValueMust(
				archObjectType.AttrTypes,
				map[string]attr.Value{
					"url":  arch.URL,
					"sha":  arch.Sha,
					"os":   arch.OS,
					"arch": arch.Arch,
				},
			)
			tflog.Debug(ctx, "Updating AMD64 architecture in plan", map[string]interface{}{
				"url":  arch.URL.ValueString(),
				"sha":  arch.Sha.ValueString(),
				"os":   arch.OS.ValueString(),
				"arch": arch.Arch.ValueString(),
			})

			resp.PlanValue = types.SetValueMust(archObjectType, []attr.Value{attrValue})
			return
		}
	}
}

// SyncTopLevelURLSHAWithAMD64 creates a plan modifier that synchronizes the top-level URL/SHA with the AMD64 architecture on updates where URL or SHA is not set in the config,
func SyncTopLevelURLSHAWithAMD64() planmodifier.String {
	return &SyncTopLevelURLSHAWithAMD64Modifier{}
}

type SyncTopLevelURLSHAWithAMD64Modifier struct{}

func (m *SyncTopLevelURLSHAWithAMD64Modifier) Description(ctx context.Context) string {
	return "Combines top-level URL/SHA with AMD64 architecture"
}

func (m *SyncTopLevelURLSHAWithAMD64Modifier) MarkdownDescription(ctx context.Context) string {
	return "Combines top-level URL/SHA with AMD64 architecture"
}

// PlanModifySet modifies the plan to combine URL/SHA with AMD64 architecture
func (m *SyncTopLevelURLSHAWithAMD64Modifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// Skip if we're destroying the resource or no state
	if req.Plan.Raw.IsNull() || req.State.Raw.IsNull() || req.StateValue.IsNull() {
		tflog.Debug(ctx, "Skipping AMD64 URL/SHA combination because state or plan is null")
		return
	}

	if !req.ConfigValue.IsUnknown() && !req.ConfigValue.IsNull() {
		tflog.Debug(ctx, "Skipping because value is set in config")
		return
	}

	var configArchs types.Set
	diags := req.Config.GetAttribute(ctx, path.Root("archs"), &configArchs)
	if diags.HasError() {
		tflog.Debug(ctx, "Error extracting config architectures", map[string]interface{}{
			"diagnostics": diags,
		})
		return
	}

	if configArchs.IsNull() && configArchs.IsUnknown() {
		tflog.Debug(ctx, "Skipping top level arch modifying because archs are NOT set in config")
		return
	}

	segments := req.Path.String()
	attributeName := segments[strings.LastIndex(segments, ".")+1:]

	var amd64Arch ToolArchitecture
	var configArchsList []ToolArchitecture
	diags = configArchs.ElementsAs(ctx, &configArchsList, false)
	if diags.HasError() {
		tflog.Debug(ctx, "Error extracting config architectures", map[string]interface{}{
			"diagnostics": diags,
		})
		return
	}

	for _, arch := range configArchsList {
		if arch.Arch.ValueString() == "amd64" {
			amd64Arch = arch
			break
		}
	}

	if amd64Arch.Arch.IsNull() || amd64Arch.Arch.IsUnknown() {
		resp.PlanValue = types.StringNull()
		return
	}

	switch attributeName {
	case "url":
		resp.PlanValue = types.StringValue(amd64Arch.URL.ValueString())
	case "sha":
		resp.PlanValue = types.StringValue(amd64Arch.Sha.ValueString())
	default:
		tflog.Debug(ctx, "Unsupported attribute for AMD64 combination", map[string]interface{}{
			"attribute": attributeName,
		})
		return
	}
}
