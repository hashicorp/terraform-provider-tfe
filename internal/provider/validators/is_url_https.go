// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	oldValidation "github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

type isURLWithHTTPorHTTPSValidator struct{}

func (v isURLWithHTTPorHTTPSValidator) Description(_ context.Context) string {
	return "string is a valid HTTP or HTTPS URL"
}

func (v isURLWithHTTPorHTTPSValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v isURLWithHTTPorHTTPSValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	value := request.ConfigValue.ValueString()

	if _, errs := oldValidation.IsURLWithHTTPorHTTPS(value, value); errs != nil {
		for _, err := range errs {
			response.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
				request.Path,
				"Invalid Attribute Value",
				err.Error(),
			))
		}
	}
}
func IsURLWithHTTPorHTTPS() validator.String {
	return isURLWithHTTPorHTTPSValidator{}
}
