// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package validators

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// enumMemberValidator validates that a string is a recognized member of an API
// enum by delegating to a parser generated from the API contract.
type enumMemberValidator struct {
	parse    func(string) (any, error)
	enumName string
}

func (v enumMemberValidator) Description(_ context.Context) string {
	return fmt.Sprintf("value must be a recognized %s", v.enumName)
}

func (v enumMemberValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v enumMemberValidator) ValidateString(_ context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	value := request.ConfigValue.ValueString()

	// The generated parser follows the Kiota convention of returning a nil
	// result for values outside the enum. Relying on it keeps this validation
	// aligned with the API contract, so newly introduced values are accepted
	// automatically without changing the provider.
	parsed, err := v.parse(value)
	if err == nil && parsed != nil {
		return
	}

	response.Diagnostics.AddAttributeError(
		request.Path,
		"Invalid Attribute Value",
		fmt.Sprintf("Attribute %s value %q is not a recognized %s.", request.Path, value, v.enumName),
	)
}

// EnumMember returns a validator that accepts only values recognized by the
// given API enum parser. This lets configuration validation track the API
// contract: when the generated client learns a new value, it becomes valid here
// with no provider change. enumName is used in diagnostics to describe the enum.
func EnumMember(parse func(string) (any, error), enumName string) validator.String {
	return enumMemberValidator{parse: parse, enumName: enumName}
}
