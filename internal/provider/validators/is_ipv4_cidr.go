// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package validators

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type isIPv4CIDRValidator struct{}

func (v isIPv4CIDRValidator) Description(_ context.Context) string {
	return "Validates that a string is a canonical IPv4 CIDR range with no host bits set (e.g. 10.0.0.0/24)"
}

func (v isIPv4CIDRValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v isIPv4CIDRValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	value := request.ConfigValue.ValueString()

	ip, ipNet, err := net.ParseCIDR(value)
	if err != nil {
		response.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
			request.Path,
			"Invalid Attribute Value",
			fmt.Sprintf("%q is not a valid CIDR range: %s", value, err),
		))
		return
	}

	// HCP Terraform only supports IPv4 CIDR ranges for IP allowlists. Reject
	// IPv6 ranges, including IPv4-mapped IPv6 forms (e.g. "::ffff:192.0.2.1/120"),
	// which net.ParseCIDR accepts and whose parsed IP still reports To4() != nil.
	// Requiring a dotted-quad address with no colons keeps the value canonical
	// IPv4.
	if strings.Contains(value, ":") || ip.To4() == nil {
		response.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
			request.Path,
			"Invalid Attribute Value",
			fmt.Sprintf("%q is not a valid IPv4 CIDR range; only IPv4 ranges are supported", value),
		))
		return
	}

	// The API normalizes CIDR ranges by masking host bits (for example,
	// "10.0.0.5/16" becomes "10.0.0.0/16"), which would otherwise produce a
	// permanent diff. Reject non-canonical ranges (those with host bits set) so
	// the configured value matches what the API stores and returns.
	if !ipNet.IP.Equal(ip) {
		response.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
			request.Path,
			"Invalid Attribute Value",
			fmt.Sprintf("%q has host bits set; use the canonical network address %q instead", value, ipNet.String()),
		))
	}
}

// IsIPv4CIDR returns a validator which ensures that a string attribute is a
// valid IPv4 CIDR range.
func IsIPv4CIDR() validator.String {
	return isIPv4CIDRValidator{}
}
