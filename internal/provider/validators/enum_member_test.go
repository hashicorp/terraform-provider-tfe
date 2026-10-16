// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package validators

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestEnumMember(t *testing.T) {
	// parse mimics a generated Kiota enum parser: it returns a non-nil result
	// for recognized values, a nil result for unrecognized-but-well-formed
	// values, and an error only for malformed input.
	parse := func(value string) (any, error) {
		switch value {
		case "gravatar":
			return "gravatar", nil
		case "boom":
			return nil, errors.New("parse failure")
		default:
			return nil, nil
		}
	}

	cases := map[string]struct {
		value     types.String
		wantError bool
	}{
		"null is allowed": {
			value: types.StringNull(),
		},
		"unknown is allowed": {
			value: types.StringUnknown(),
		},
		"recognized value is allowed": {
			value: types.StringValue("gravatar"),
		},
		"unrecognized value is rejected": {
			value:     types.StringValue("disney"),
			wantError: true,
		},
		"parser error is rejected": {
			value:     types.StringValue("boom"),
			wantError: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			req := validator.StringRequest{ConfigValue: tc.value}
			resp := &validator.StringResponse{}

			EnumMember(parse, "avatar source").ValidateString(context.Background(), req, resp)

			hasError := resp.Diagnostics.HasError()
			if hasError != tc.wantError {
				t.Fatalf("expected error=%v, got error=%v (%s)", tc.wantError, hasError, resp.Diagnostics)
			}
		})
	}
}
