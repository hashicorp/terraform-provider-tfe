package tfe

import (
	"fmt"

	tfe "github.com/hashicorp/go-tfe"
)

// fetchTerraformVersionID returns a Terraform Version ID for the given Terraform version number
func fetchTerraformVersionID(version string, client *tfe.Client) (string, error) {
	versionID := ""
	pageNum := 0

found:
	for {
		versions, err := client.Admin.TerraformVersions.List(ctx, tfe.AdminTerraformVersionsListOptions{
			ListOptions: tfe.ListOptions{
				PageNumber: pageNum,
				PageSize:   20,
			},
		})
		if err != nil {
			return "", fmt.Errorf("error reading Terraform versions: %w", err)
		}

		// we've run out of versions to search
		if len(versions.Items) == 0 {
			break
		}

		for _, v := range versions.Items {
			if v.Version == version {
				versionID = v.ID
				break found
			}
		}

		pageNum++
	}

	if versionID == "" {
		return "", fmt.Errorf("terraform version not found")
	}

	return versionID, nil
}
