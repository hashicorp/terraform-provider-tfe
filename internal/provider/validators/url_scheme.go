// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package validators

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// urlSchemeValidator validates that a string is either empty or a URL whose
// scheme is one of the permitted schemes. An empty string is always allowed so
// that an override can be cleared. This mirrors the server-side validation in
// Terraform Enterprise, which parses the value with a URL parser and checks the
// scheme against an allow-list.
type urlSchemeValidator struct {
	allowedSchemes []string
}

func (v urlSchemeValidator) description() string {
	return fmt.Sprintf("value must be empty or a URL using one of the following schemes: %s", strings.Join(v.allowedSchemes, ", "))
}

func (v urlSchemeValidator) Description(_ context.Context) string {
	return v.description()
}

func (v urlSchemeValidator) MarkdownDescription(_ context.Context) string {
	return v.description()
}

func (v urlSchemeValidator) ValidateString(_ context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	value := request.ConfigValue.ValueString()
	// An empty string is permitted so the override can be cleared.
	if value == "" {
		return
	}

	parsed, err := url.Parse(value)
	if err != nil {
		response.Diagnostics.AddAttributeError(
			request.Path,
			"Invalid URL",
			fmt.Sprintf("Value %q is not a valid URL: %s", value, err),
		)
		return
	}

	for _, scheme := range v.allowedSchemes {
		if strings.EqualFold(parsed.Scheme, scheme) {
			return
		}
	}

	response.Diagnostics.AddAttributeError(
		request.Path,
		"Invalid URL scheme",
		fmt.Sprintf("Value %q must use one of the following schemes: %s.", value, strings.Join(v.allowedSchemes, ", ")),
	)
}

// URLWithSchemes returns a validator ensuring a string is empty or a URL using
// one of the given schemes (for example "https", "http", "mailto").
func URLWithSchemes(schemes ...string) validator.String {
	return urlSchemeValidator{allowedSchemes: schemes}
}
