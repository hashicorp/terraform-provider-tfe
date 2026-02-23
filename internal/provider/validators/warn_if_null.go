// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package validators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type warnIfNullValidator struct {
	message string
}

func (v warnIfNullValidator) Description(ctx context.Context) string {
	return "Warns if the attribute value is null"
}

func (v warnIfNullValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v warnIfNullValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() {
		resp.Diagnostics.AddWarning(
			v.message
		)
	}
}

func WarnIfNull(message string) validator.String {
	return warnIfNullValidator{
		message: message,
	}
}
