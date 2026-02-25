// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package planmodifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

type warnIfNullOnCreateModifier struct {
	message string
}

func (m warnIfNullOnCreateModifier) Description(ctx context.Context) string {
	return "Warning that the attribute value is null during resource creation"
}

func (m warnIfNullOnCreateModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m warnIfNullOnCreateModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.State.Raw.IsNull() && req.ConfigValue.IsNull() {
		resp.Diagnostics.AddWarning(
			m.message,
			"",
		)
	}
}

func WarnIfNullOnCreate(message string) planmodifier.String {
	return warnIfNullOnCreateModifier{
		message: message,
	}
}
