// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package planmodifiers

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-tfe/internal/provider/helpers"
)

var _ planmodifier.String = &replaceForWriteOnlyStringValue{}

func NewReplaceForWriteOnlyStringValue(attributeWriteOnly string) planmodifier.String {
	return &replaceForWriteOnlyStringValue{
		attributeWriteOnly: attributeWriteOnly,
	}
}

// replaceForWriteOnlyStringValue is a plan modifier that will cause a resource
// to be replaced if the value of a write-only attribute has changed.
//
// For this to work, the write-only attribute must be added to private state
// using WriteOnlyValueStore.SetPriorValue() after creating or updating the value.
type replaceForWriteOnlyStringValue struct {
	attributeWriteOnly string
}

func (p *replaceForWriteOnlyStringValue) Description(ctx context.Context) string {
	return "The resource will be replaced when the value of `%s` has changed"
}

func (p *replaceForWriteOnlyStringValue) MarkdownDescription(ctx context.Context) string {
	return p.Description(ctx)
}

func (p *replaceForWriteOnlyStringValue) PlanModifyString(ctx context.Context, request planmodifier.StringRequest, response *planmodifier.StringResponse) {
	// This plan modifier can be used to trigger a resource replacement when the
	// value of a write-only attribute has changed.
	//
	// Write-only argument values cannot produce a Terraform plan difference on
	// their own. The prior state value for a write-only argument will always be
	// null, and the planned state value will also be null.
	//
	// The one exception to this case is if the write-only argument is added to
	// requires_replace during Plan Modification, in that case, the write-only
	// argument will always cause a diff/trigger a resource recreation.
	var writeOnlyValue types.String
	diags := request.Config.GetAttribute(ctx, path.Root(p.attributeWriteOnly), &writeOnlyValue)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	writeOnlyValueExists := !writeOnlyValue.IsNull()
	writeOnlyValueStore := helpers.NewWriteOnlyValueStore(request.Private, p.attributeWriteOnly)
	priorWriteOnlyValueExists, diags := writeOnlyValueStore.PriorValueExists(ctx)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	if !writeOnlyValueExists && priorWriteOnlyValueExists {
		tflog.Debug(ctx, fmt.Sprintf("Replacing resource because the write-only `%s` attribute has been removed", p.attributeWriteOnly))
		response.RequiresReplace = true
		return
	}

	if !writeOnlyValueExists {
		return
	}

	// Now are dealing with a write-only attribute that has a value set.
	if !priorWriteOnlyValueExists {
		tflog.Debug(ctx, fmt.Sprintf("Replacing resource because the write-only `%s` attribute has been newly added to a resource in state", p.attributeWriteOnly))
		response.RequiresReplace = true
		return
	}

	matches, diags := writeOnlyValueStore.MatchesPriorValue(ctx, writeOnlyValue)
	response.Diagnostics.Append(diags...)

	if !matches {
		tflog.Debug(ctx, fmt.Sprintf("Replacing resource because the value of the write-only `%s` attribute has changed", p.attributeWriteOnly))
		response.RequiresReplace = true
	}
}
