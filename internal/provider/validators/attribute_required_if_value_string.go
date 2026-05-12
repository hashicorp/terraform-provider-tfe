// Copyright IBM Corp. 2018, 2026
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
		if attributeValue.ValueString() == requiredValue && req.ConfigValue.IsNull() {
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

type attributeRequiredIfValueStringUnlessOtherSetValidator struct {
	attributeName       string
	requiredValues      []string
	unlessAttributeName string
}

func (v attributeRequiredIfValueStringUnlessOtherSetValidator) Description(ctx context.Context) string {
	return fmt.Sprintf("Ensures the attribute is required if '%s' is one of %v, unless '%s' is set", v.attributeName, v.requiredValues, v.unlessAttributeName)
}

func (v attributeRequiredIfValueStringUnlessOtherSetValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v attributeRequiredIfValueStringUnlessOtherSetValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	var unlessValue types.String
	diags := req.Config.GetAttribute(ctx, path.Root(v.unlessAttributeName), &unlessValue)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !unlessValue.IsNull() {
		return
	}

	var attributeValue types.String
	diags = req.Config.GetAttribute(ctx, path.Root(v.attributeName), &attributeValue)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	for _, requiredValue := range v.requiredValues {
		if attributeValue.ValueString() == requiredValue && req.ConfigValue.IsNull() {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Missing Required Attribute",
				fmt.Sprintf("The attribute '%s' is required when '%s' is '%s'", req.Path, v.attributeName, requiredValue),
			)
			return
		}
	}
}

// AttributeRequiredIfValueStringUnlessOtherSet validates that the attribute is required when
// attributeName equals one of requiredValues, unless unlessAttributeName is also set.
func AttributeRequiredIfValueStringUnlessOtherSet(attributeName string, requiredValues []string, unlessAttributeName string) validator.String {
	return attributeRequiredIfValueStringUnlessOtherSetValidator{
		attributeName:       attributeName,
		requiredValues:      requiredValues,
		unlessAttributeName: unlessAttributeName,
	}
}
