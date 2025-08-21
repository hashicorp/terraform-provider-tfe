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

	var stateURL, planURL, stateSHA, planSHA types.String
	// Get values from state and plan
	req.State.GetAttribute(ctx, path.Root("url"), &stateURL)
	req.Plan.GetAttribute(ctx, path.Root("url"), &planURL)
	req.State.GetAttribute(ctx, path.Root("sha"), &stateSHA)
	req.Plan.GetAttribute(ctx, path.Root("sha"), &planSHA)

	// Check if values are changing
	urlChanged := !stateURL.Equal(planURL)
	shaChanged := !stateSHA.Equal(planSHA)

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

	// Check if AMD64 is already in the plan
	for i, arch := range planArchsList {
		if arch.Arch.ValueString() == "amd64" {
			// If URL or SHA is changing, update the AMD64 arch to match
			if urlChanged {
				arch.URL = planURL
			}
			if shaChanged {
				arch.Sha = planSHA
			}
			// Update the plan architecture list with the modified AMD64 arch
			planArchsList[i] = arch

			// Update the plan with the modified AMD64 arch
			archObjectType := ObjectTypeForArchitectures()
			attrValues := make([]attr.Value, len(planArchsList))

			for i, arch := range planArchsList {
				attrValues[i] = types.ObjectValueMust(
					archObjectType.AttrTypes,
					map[string]attr.Value{
						"url":  arch.URL,
						"sha":  arch.Sha,
						"os":   arch.OS,
						"arch": arch.Arch,
					},
				)
			}

			resp.PlanValue = types.SetValueMust(archObjectType, attrValues)
			return
		}
	}
}

// ValidateToolVersion provides common validation for tool version resources
func ValidateToolVersion(ctx context.Context, url, sha types.String, archs types.Set, resourceType string) diag.Diagnostics {
	var diags diag.Diagnostics

	urlPresent := !url.IsNull() && !url.IsUnknown()
	shaPresent := !sha.IsNull() && !sha.IsUnknown()

	// If URL or SHA is not set, we will rely on the archs attribute
	if !urlPresent || !shaPresent {
		return diags
	}

	// Check if archs is present
	if !archs.IsNull() && !archs.IsUnknown() {
		// Extract archs
		var archsList []ToolArchitecture
		archDiags := archs.ElementsAs(ctx, &archsList, false)
		if archDiags.HasError() {
			diags.Append(archDiags...)
			return diags
		}

		// Check for AMD64 architecture
		var hasAMD64 bool
		for _, arch := range archsList {
			if arch.Arch.ValueString() == "amd64" {
				hasAMD64 = true

				// If URL and SHA are set at top level, check they match AMD64 arch
				// Check URL matches
				if urlPresent && url.ValueString() != arch.URL.ValueString() {
					diags.AddError(
						fmt.Sprintf("Inconsistent %s URL values", resourceType),
						fmt.Sprintf("Top-level URL (%s) doesn't match AMD64 architecture URL (%s)",
							url.ValueString(), arch.URL.ValueString()),
					)
				}

				// Check SHA matches
				if shaPresent && sha.ValueString() != arch.Sha.ValueString() {
					diags.AddError(
						fmt.Sprintf("Inconsistent %s SHA values", resourceType),
						fmt.Sprintf("Top-level SHA (%s) doesn't match AMD64 architecture SHA (%s)",
							sha.ValueString(), arch.Sha.ValueString()),
					)
				}

				break
			}
		}

		// If top-level URL/SHA are set and no AMD64 arch found, add error
		if !hasAMD64 && (!url.IsNull() || !sha.IsNull()) {
			diags.AddError(
				fmt.Sprintf("Missing AMD64 architecture in %s", resourceType),
				"When specifying both top-level URL/SHA and archs, an AMD64 architecture entry must be included",
			)
		}
	}

	return diags
}
