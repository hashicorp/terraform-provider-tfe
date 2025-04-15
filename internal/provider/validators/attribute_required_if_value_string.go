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

type attributeRequiredIfValueStringValidator struct {
	attributeName  string
	requiredValues []string
}

func (v attributeRequiredIfValueStringValidator) Description(ctx context.Context) string {
	return fmt.Sprintf("Ensures the attribute is required if '%s' is one of %v", v.attributeName, v.requiredValues)
}

func (v attributeRequiredIfValueStringValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v attributeRequiredIfValueStringValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	var attributeValue types.String
	diags := req.Config.GetAttribute(ctx, path.Root(v.attributeName), &attributeValue)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	for _, requiredValue := range v.requiredValues {
		if attributeValue.ValueString() == requiredValue && (req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown()) {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Missing Required Attribute",
				fmt.Sprintf("The attribute '%s' is required when '%s' is '%s'", req.Path, v.attributeName, requiredValue),
			)
			return
		}
	}
}

func AttributeRequiredIfValueString(attributeName string, requiredValues []string) validator.String {
	return attributeRequiredIfValueStringValidator{
		attributeName:  attributeName,
		requiredValues: requiredValues,
	}
}
