// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// modelTFEProject maps the resource or data source schema data to a
// struct.
type modelTFEProject struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	Organization types.String `tfsdk:"organization"`
}

// modelFromTFEProject builds a modelTFEProject struct from a
// tfe.Project value.
func modelFromTFEProject(v *tfe.Project) modelTFEProject {
	return modelTFEProject{
		ID:           types.StringValue(v.ID),
		Name:         types.StringValue(v.Name),
		Description:  types.StringValue(v.Description),
		Organization: types.StringValue(v.Organization.Name),
	}
}
