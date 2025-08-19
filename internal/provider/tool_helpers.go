// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

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
