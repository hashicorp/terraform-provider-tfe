// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"time"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// modelTFERegistryGPGKey maps the resource or data source schema data to a
// struct.
type modelTFERegistryGPGKey struct {
	ID           types.String `tfsdk:"id"`
	Organization types.String `tfsdk:"organization"`
	ASCIIArmor   types.String `tfsdk:"ascii_armor"`
	CreatedAt    types.String `tfsdk:"created_at"`
	UpdatedAt    types.String `tfsdk:"updated_at"`
}

// modelFromTFEVGPGKey builds a modelTFERegistryGPGKey struct from a
// tfe.GPGKey value.
func modelFromTFEVGPGKey(v *tfe.GPGKey) modelTFERegistryGPGKey {
	return modelTFERegistryGPGKey{
		ID:           types.StringValue(v.KeyID),
		Organization: types.StringValue(v.Namespace),
		ASCIIArmor:   types.StringValue(v.AsciiArmor),
		CreatedAt:    types.StringValue(v.CreatedAt.Format(time.RFC3339)),
		UpdatedAt:    types.StringValue(v.UpdatedAt.Format(time.RFC3339)),
	}
}
