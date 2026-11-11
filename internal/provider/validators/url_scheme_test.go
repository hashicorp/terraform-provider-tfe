// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package validators

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestURLWithSchemes(t *testing.T) {
	cases := map[string]struct {
		schemes   []string
		value     types.String
		wantError bool
	}{
		"null is allowed": {
			schemes: []string{"https", "http"},
			value:   types.StringNull(),
		},
		"unknown is allowed": {
			schemes: []string{"https", "http"},
			value:   types.StringUnknown(),
		},
		"empty string is allowed": {
			schemes: []string{"https", "http", "mailto"},
			value:   types.StringValue(""),
		},
		"https is allowed": {
			schemes: []string{"https", "http"},
			value:   types.StringValue("https://example.com/support"),
		},
		"http is allowed": {
			schemes: []string{"https", "http"},
			value:   types.StringValue("http://example.com/support"),
		},
		"mailto is allowed when listed": {
			schemes: []string{"https", "http", "mailto"},
			value:   types.StringValue("mailto:support@example.com"),
		},
		"scheme is case-insensitive": {
			schemes: []string{"https"},
			value:   types.StringValue("HTTPS://example.com"),
		},
		"ftp is rejected": {
			schemes:   []string{"https", "http"},
			value:     types.StringValue("ftp://example.com"),
			wantError: true,
		},
		"mailto rejected when not listed": {
			schemes:   []string{"https", "http"},
			value:     types.StringValue("mailto:support@example.com"),
			wantError: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			req := validator.StringRequest{ConfigValue: tc.value}
			resp := &validator.StringResponse{}

			URLWithSchemes(tc.schemes...).ValidateString(context.Background(), req, resp)

			hasError := resp.Diagnostics.HasError()
			if hasError != tc.wantError {
				t.Fatalf("expected error=%v, got error=%v (%s)", tc.wantError, hasError, resp.Diagnostics)
			}
		})
	}
}
