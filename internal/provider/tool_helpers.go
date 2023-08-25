// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"

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
