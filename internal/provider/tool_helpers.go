// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"strings"

	tfe "github.com/hashicorp/go-tfe"
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
	Sha  types.String `tfsdk:"sha"` // Standardized to lowercase field name
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
			Sha:  model.Sha.ValueString(), // Consistent lowercase field name
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

// Implement the plan modifier interface
type preserveAMD64ArchsModifier struct{}

// Description provides a plain text description of the plan modifier
func (m *preserveAMD64ArchsModifier) Description(ctx context.Context) string {
	return "Preserves AMD64 architecture entries when top-level URL or SHA changes"
}

// MarkdownDescription provides markdown documentation
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

	var configURL, planURL, configSHA, planSHA types.String

	// Get values from state and plan
	req.Config.GetAttribute(ctx, path.Root("url"), &configURL)
	req.Plan.GetAttribute(ctx, path.Root("url"), &planURL)
	req.State.GetAttribute(ctx, path.Root("sha"), &configSHA)
	req.Plan.GetAttribute(ctx, path.Root("sha"), &planSHA)

	// Check if values are changing
	urlChanged := !configURL.Equal(planURL)
	shaChanged := !configSHA.Equal(planSHA)

	// If neither URL nor SHA is changing, do nothing
	if !urlChanged && !shaChanged {
		return
	}

	// Extract state archs and plan archs
	stateArchs := req.StateValue
	planArchs := req.PlanValue

	// Extract archs from state
	var stateArchsList []ToolArchitecture
	diags := stateArchs.ElementsAs(ctx, &stateArchsList, false)
	if diags.HasError() {
		return
	}

	// we need to update the plan amd url and sha to match the top level values
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
		"planURL":        planURL,
		"stateSHA":       configSHA,
		"planSHA":        planSHA,
	})
	// Check if AMD64 is already in the plan
	for _, arch := range planArchsList {
		if arch.Arch.ValueString() == "amd64" {
			tflog.Debug(ctx, "Found AMD64 architecture in plan", map[string]interface{}{
				"url":  arch.URL.ValueString(),
				"sha":  arch.Sha.ValueString(),
				"os":   arch.OS.ValueString(),
				"arch": arch.Arch.ValueString(),
			})
			// If we found AMD64, update its URL and SHA if they are changing
			// If URL or SHA is changing, update the AMD64 arch to match
			if urlChanged {
				arch.URL = configURL
			}
			if shaChanged {
				arch.Sha = configURL
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

			resp.PlanValue = types.SetValueMust(archObjectType, []attr.Value{attrValue})
			return
		}
	}
}

// SyncTopLevelURLSHAWithAMD64 creates a plan modifier that synchronizes the top-level URL/SHA with the AMD64 architecture on updates where URL or SHA is not set in the config,
func SyncTopLevelURLSHAWithAMD64() planmodifier.String {
	return &SyncTopLevelURLSHAWithAMD64Modifier{}
}

// Implement the plan modifier interface
type SyncTopLevelURLSHAWithAMD64Modifier struct{}

// Description provides a plain text description of the plan modifier
func (m *SyncTopLevelURLSHAWithAMD64Modifier) Description(ctx context.Context) string {
	return "Combines top-level URL/SHA with AMD64 architecture"
}

// MarkdownDescription provides markdown documentation
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

	// if config archs is not set we will not modify the plan
	// get the config archs

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

	// we do not know the name of the attibute
	// we are modifying, so we will use the path to get the attribute name
	segments := req.Path.String()
	attributeName := segments[strings.LastIndex(segments, ".")+1:]

	// get the archs from the plan
	// get the amd arch from the configArchs
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
		tflog.Debug(ctx, "No AMD64 architecture found in config archs, skipping modification")
		return
	}

	// Get the value of the attributeName from the amd64 arch
	switch attributeName {
	case "url":
		// set the plan value to the URL of the AMD64 arch
		resp.PlanValue = types.StringValue(amd64Arch.URL.ValueString())
	case "sha":
		// set the plan value to the SHA of the AMD64 arch
		resp.PlanValue = types.StringValue(amd64Arch.Sha.ValueString())
	default:
		tflog.Debug(ctx, "Unsupported attribute for AMD64 combination", map[string]interface{}{
			"attribute": attributeName,
		})
		return
	}

}

//
// // ValidateToolVersion provides common validation for tool version resources
// func ValidateToolVersion(ctx context.Context, url, sha types.String, archs types.Set, resourceType string) diag.Diagnostics {
//     var diags diag.Diagnostics

//     urlPresent := !url.IsNull() && !url.IsUnknown()
//     shaPresent := !sha.IsNull() && !sha.IsUnknown()

//     // If URL or SHA is not set, we will rely on the archs attribute
//     if !urlPresent || !shaPresent {
//         return diags
//     }

//     // If archs aren't present, we can't validate against them
//     if archs.IsNull() || archs.IsUnknown() {
//         return diags
//     }

//     // Extract archs
//     var archsList []ToolArchitecture
//     archDiags := archs.ElementsAs(ctx, &archsList, false)
//     if archDiags.HasError() {
//         diags.Append(archDiags...)
//         return diags
//     }

//     // Check for AMD64 architecture
//     hasAMD64 := false
//     var amd64Arch ToolArchitecture
//     for _, arch := range archsList {
//         if arch.Arch.ValueString() == "amd64" {
//             hasAMD64 = true
//             amd64Arch = arch
//             break
//         }
//     }

//     // If top-level URL/SHA are set and no AMD64 arch found, add error
//     if !hasAMD64 {
//         diags.AddError(
//             fmt.Sprintf("Missing AMD64 architecture in %s", resourceType),
//             "When specifying both top-level URL/SHA and archs, an AMD64 architecture entry must be included",
//         )
//         return diags
//     }

//     // If URL and SHA are set at top level, check they match AMD64 arch
//     if url.ValueString() != amd64Arch.URL.ValueString() {
//         diags.AddError(
//             fmt.Sprintf("Inconsistent %s URL values", resourceType),
//             fmt.Sprintf("Top-level URL (%s) doesn't match AMD64 architecture URL (%s)",
//                 url.ValueString(), amd64Arch.URL.ValueString()),
//         )
//     }

//     // Check SHA matches
//     if sha.ValueString() != amd64Arch.Sha.ValueString() {
//         diags.AddError(
//             fmt.Sprintf("Inconsistent %s SHA values", resourceType),
//             fmt.Sprintf("Top-level SHA (%s) doesn't match AMD64 architecture SHA (%s)",
//                 sha.ValueString(), amd64Arch.Sha.ValueString()),
//         )
//     }

//     return diags
// }
