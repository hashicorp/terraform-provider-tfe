// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// AttrGettable is a small enabler for helper functions that need to read one
// attribute of a Configuration, Plan, or State.
type AttrGettable interface {
	GetAttribute(ctx context.Context, path path.Path, target interface{}) diag.Diagnostics
}

// dataOrDefaultOrganization returns the value of the "organization" attribute
// from the Config/Plan/State data, defaulting to the provier configuration.
// If neither is set, an error is returned.
func (c *ConfiguredClient) dataOrDefaultOrganization(ctx context.Context, data AttrGettable, target *string) diag.Diagnostics {
	schemaPath := path.Root("organization")

	var organization types.String
	diags := data.GetAttribute(ctx, schemaPath, &organization)
	if diags.HasError() {
		return diags
	}

	if !organization.IsNull() && !organization.IsUnknown() {
		*target = organization.ValueString()
	} else if c.Organization == "" {
		diags.AddAttributeError(schemaPath, "No organization was specified on the resource or provider", "")
	} else {
		*target = c.Organization
	}

	return diags
}
