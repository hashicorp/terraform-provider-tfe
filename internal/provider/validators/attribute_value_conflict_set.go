// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validators

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type attributeValueConflictSetValidator struct {
	attributeName     string
	conflictingValues []string
}

func (v attributeValueConflictSetValidator) Description(ctx context.Context) string {
	return fmt.Sprintf("Ensures the attribute is not set if %s is one of %v", v.attributeName, v.conflictingValues)
}

func (v attributeValueConflictSetValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v attributeValueConflictSetValidator) ValidateSet(ctx context.Context, req validator.SetRequest, resp *validator.SetResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var attributeValue types.String
	diags := req.Config.GetAttribute(ctx, path.Root(v.attributeName), &attributeValue)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	for _, conflictingValue := range v.conflictingValues {
		if attributeValue.ValueString() == conflictingValue {
			resp.Diagnostics.AddError(
				"Invalid Attribute Value",
				fmt.Sprintf("The attribute '%s' cannot be set when '%s' is '%s'", req.Path, v.attributeName, conflictingValue),
			)

			return
		}
	}
}

func AttributeValueConflictSetValidator(attributeName string, conflictingValues []string) validator.Set {
	return attributeValueConflictSetValidator{attributeName: attributeName, conflictingValues: conflictingValues}
}
