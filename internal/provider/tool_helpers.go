// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
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

func setUnknownTfAttrs(tfVersion modelAdminTerraformVersion, v *tfe.AdminTerraformVersion) modelAdminTerraformVersion {
	tfVersion.ID = types.StringValue(v.ID)
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

	return tfVersion
}

func convertToToolVersionArchitectures(archs interface{}) []*tfe.ToolVersionArchitecture {
	if archs == nil {
		return nil
	}

	archsList, ok := archs.([]interface{})
	if !ok {
		return nil
	}

	var convertedArchs []*tfe.ToolVersionArchitecture
	for _, arch := range archsList {
		if archMap, ok := arch.(map[string]interface{}); ok {
			convertedArchs = append(convertedArchs, &tfe.ToolVersionArchitecture{
				URL:  archMap["url"].(string),
				Sha:  archMap["sha"].(string),
				OS:   archMap["os"].(string),
				Arch: archMap["arch"].(string),
			})
		}
	}

	return convertedArchs
}

func convertToToolVersionArchitecturesMap(archs []*tfe.ToolVersionArchitecture) []map[string]interface{} {
	if len(archs) == 0 {
		return nil
	}

	archsList := make([]map[string]interface{}, len(archs))
	for i, arch := range archs {
		archsList[i] = map[string]interface{}{
			"url":  arch.URL,
			"sha":  arch.Sha,
			"os":   arch.OS,
			"arch": arch.Arch,
		}
	}

	return archsList
}

func stringOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func newConvertToToolVersionArchitectures(ctx context.Context, archs types.Set) ([]*tfe.ToolVersionArchitecture, diag.Diagnostics) {
	if archs.IsNull() || archs.IsUnknown() {
		return nil, nil
	}

	var diags diag.Diagnostics
	var archModels []modelArch
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
