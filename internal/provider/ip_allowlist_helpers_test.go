// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	tfev2models "github.com/hashicorp/go-tfe/v2/api/models"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	customValidators "github.com/hashicorp/terraform-provider-tfe/internal/provider/validators"
)

func TestEnforcementScopeRoundTrip(t *testing.T) {
	scopes := []string{
		ipAllowlistScopeOrganization,
		ipAllowlistScopeAllAgentPools,
		ipAllowlistScopeSelectedAgentPools,
	}

	for _, scope := range scopes {
		v2Scope, err := enforcementScopeToV2(scope)
		if err != nil {
			t.Fatalf("enforcementScopeToV2(%q) unexpected error: %s", scope, err)
		}
		if got := enforcementScopeFromV2(&v2Scope); got != scope {
			t.Errorf("round trip for %q returned %q", scope, got)
		}
	}
}

func TestEnforcementScopeToV2Invalid(t *testing.T) {
	if _, err := enforcementScopeToV2("nope"); err == nil {
		t.Fatal("expected error for invalid enforcement_scope, got nil")
	}
}

func TestEnforcementScopeFromV2Nil(t *testing.T) {
	if got := enforcementScopeFromV2(nil); got != "" {
		t.Errorf("expected empty string for nil scope, got %q", got)
	}
}

func TestStringSliceDifference(t *testing.T) {
	cases := []struct {
		name string
		a    []string
		b    []string
		want []string
	}{
		{"disjoint", []string{"a", "b"}, []string{"c"}, []string{"a", "b"}},
		{"subset", []string{"a", "b"}, []string{"a", "b", "c"}, nil},
		{"partial", []string{"a", "b", "c"}, []string{"b"}, []string{"a", "c"}},
		{"empty a", nil, []string{"a"}, nil},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := stringSliceDifference(tc.a, tc.b)
			if len(got) != len(tc.want) {
				t.Fatalf("stringSliceDifference(%v, %v) = %v, want %v", tc.a, tc.b, got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("stringSliceDifference(%v, %v) = %v, want %v", tc.a, tc.b, got, tc.want)
				}
			}
		})
	}
}

func TestCidrRangesRelationship(t *testing.T) {
	ranges := []modelTFECIDRRange{
		{
			Range:       types.StringValue("10.0.0.0/24"),
			Description: types.StringValue("office"),
			Enabled:     types.BoolValue(true),
		},
		{
			Range:       types.StringValue("192.168.1.0/24"),
			Description: types.StringNull(),
			Enabled:     types.BoolValue(false),
		},
	}

	rel := cidrRangesRelationship(ranges)
	if rel == nil {
		t.Fatal("cidrRangesRelationship returned nil")
	}

	cidrRanges := rel.GetCidrRanges()
	if cidrRanges == nil {
		t.Fatal("expected cidr-ranges relationship to be set")
	}

	data := cidrRanges.GetData()
	if len(data) != len(ranges) {
		t.Fatalf("expected %d embedded ranges, got %d", len(ranges), len(data))
	}

	// Index the embedded ranges by CIDR to make assertions order-independent.
	byRange := make(map[string]tfev2models.CidrRangesable, len(data))
	for _, d := range data {
		attrs := d.GetAttributes()
		if attrs == nil || attrs.GetRangeEscaped() == nil {
			t.Fatal("embedded range missing attributes/range")
		}
		if d.GetTypeEscaped() == nil || *d.GetTypeEscaped() != tfev2models.CIDRRANGES_CIDRRANGES_TYPE {
			t.Errorf("embedded range has unexpected type %v", d.GetTypeEscaped())
		}
		byRange[*attrs.GetRangeEscaped()] = d
	}

	office := byRange["10.0.0.0/24"]
	if office == nil {
		t.Fatal("expected 10.0.0.0/24 to be embedded")
	}
	if desc := office.GetAttributes().GetDescription(); desc == nil || *desc != "office" {
		t.Errorf("expected description \"office\", got %v", desc)
	}
	if enabled := office.GetAttributes().GetEnabled(); enabled == nil || !*enabled {
		t.Errorf("expected 10.0.0.0/24 enabled=true, got %v", enabled)
	}

	vpn := byRange["192.168.1.0/24"]
	if vpn == nil {
		t.Fatal("expected 192.168.1.0/24 to be embedded")
	}
	if desc := vpn.GetAttributes().GetDescription(); desc != nil {
		t.Errorf("expected nil description for null description, got %q", *desc)
	}
	if enabled := vpn.GetAttributes().GetEnabled(); enabled == nil || *enabled {
		t.Errorf("expected 192.168.1.0/24 enabled=false, got %v", enabled)
	}
}

func TestIsV2ResourceNotFound(t *testing.T) {
	notFound := tfev2models.NewErrors()
	notFound.SetStatusCode(404)
	if !isV2ResourceNotFound(notFound) {
		t.Error("expected 404 Errors to be reported as not found")
	}

	serverErr := tfev2models.NewErrors()
	serverErr.SetStatusCode(500)
	if isV2ResourceNotFound(serverErr) {
		t.Error("expected 500 Errors to not be reported as not found")
	}

	if isV2ResourceNotFound(nil) {
		t.Error("expected nil error to not be reported as not found")
	}
}

func TestIsIPv4CIDRValidator(t *testing.T) {
	cases := []struct {
		name      string
		value     types.String
		expectErr bool
	}{
		{"valid ipv4 cidr", types.StringValue("10.0.0.0/24"), false},
		{"valid single host", types.StringValue("192.168.1.1/32"), false},
		{"not a cidr", types.StringValue("10.0.0.0"), true},
		{"garbage", types.StringValue("not-a-cidr"), true},
		{"ipv6 cidr rejected", types.StringValue("2001:db8::/32"), true},
		{"null skipped", types.StringNull(), false},
		{"unknown skipped", types.StringUnknown(), false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := validator.StringRequest{
				Path:        path.Root("range"),
				ConfigValue: tc.value,
			}
			resp := &validator.StringResponse{}
			customValidators.IsIPv4CIDR().ValidateString(context.Background(), req, resp)

			if tc.expectErr && !resp.Diagnostics.HasError() {
				t.Errorf("expected validation error for %q, got none", tc.value.ValueString())
			}
			if !tc.expectErr && resp.Diagnostics.HasError() {
				t.Errorf("unexpected validation error for %q: %v", tc.value.ValueString(), resp.Diagnostics)
			}
		})
	}
}
