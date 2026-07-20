// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package validators

import (
	"context"
	"fmt"
	"net"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type isIPv4CIDRValidator struct{}

func (v isIPv4CIDRValidator) Description(_ context.Context) string {
	return "string is a valid IPv4 CIDR range (e.g. 10.0.0.0/24)"
}

func (v isIPv4CIDRValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v isIPv4CIDRValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	value := request.ConfigValue.ValueString()

	ip, _, err := net.ParseCIDR(value)
	if err != nil {
		response.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
			request.Path,
			"Invalid Attribute Value",
			fmt.Sprintf("%q is not a valid CIDR range: %s", value, err),
		))
		return
	}

	// HCP Terraform only supports IPv4 CIDR ranges for IP allowlists.
	if ip.To4() == nil {
		response.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
			request.Path,
			"Invalid Attribute Value",
			fmt.Sprintf("%q is not a valid IPv4 CIDR range; only IPv4 ranges are supported", value),
		))
	}
}

// IsIPv4CIDR returns a validator which ensures that a string attribute is a
// valid IPv4 CIDR range.
func IsIPv4CIDR() validator.String {
	return isIPv4CIDRValidator{}
}
