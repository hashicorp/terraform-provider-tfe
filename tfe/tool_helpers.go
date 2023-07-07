// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tfe

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
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

// staticInt64 returns a static Int64 value default handler.
//
// Use staticInt64 if a static default value for a Int64 should be set.
func staticInt64(defaultVal int64) defaults.Int64 {
	return staticInt64Default{
		defaultVal: defaultVal,
	}
}

// staticStringDefault is static value default handler that
// sets a value on a string attribute.
type staticInt64Default struct {
	defaultVal int64
}

// Description returns a human-readable description of the default value handler.
func (d staticInt64Default) Description(_ context.Context) string {
	return fmt.Sprintf("value defaults to %d", d.defaultVal)
}

// MarkdownDescription returns a markdown description of the default value handler.
func (d staticInt64Default) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("value defaults to `%d`", d.defaultVal)
}

// DefaultInt64 implements the static default value logic.
func (d staticInt64Default) DefaultInt64(_ context.Context, req defaults.Int64Request, resp *defaults.Int64Response) {
	resp.PlanValue = types.Int64Value(d.defaultVal)
}
