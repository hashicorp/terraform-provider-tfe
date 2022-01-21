package tfe

import (
	"fmt"

	tfe "github.com/hashicorp/go-tfe"
)

// fetchTerraformVersionID returns a Terraform Version ID for the given Terraform version number
func fetchTerraformVersionID(version string, client *tfe.Client) (string, error) {
	versions, err := client.Admin.TerraformVersions.List(ctx, tfe.AdminTerraformVersionsListOptions{
		Filter: &version,
	})
	if err != nil {
		return "", fmt.Errorf("error reading Terraform versions: %w", err)
	}

	if len(versions.Items) == 0 {
		return "", fmt.Errorf("terraform version not found")
	}

	return versions.Items[0].ID, nil
}
