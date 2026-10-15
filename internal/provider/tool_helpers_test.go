// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testTFPolicyPluginURL returns a sample download URL for a linux build.
func testTFPolicyPluginURL(arch string) string {
	return "https://example.com/tfpolicy-plugin/tfpolicy-plugin_linux_" + arch + ".zip"
}

func testToolArchitecture(os, arch string) ToolArchitecture {
	return ToolArchitecture{
		URL:  types.StringValue(testTFPolicyPluginURL(arch)),
		Sha:  types.StringValue(strings.Repeat("a", 64)),
		OS:   types.StringValue(os),
		Arch: types.StringValue(arch),
	}
}

func TestIsToolVersionID(t *testing.T) {
	cases := map[string]bool{
		"tool-abc123": true,
		"tool-":       false,
		"":            false,
		"1.2.3":       false,
		"tools-abc":   false,
		"TOOL-abc":    false,
	}
	for id, want := range cases {
		t.Run(id, func(t *testing.T) {
			assert.Equal(t, want, isToolVersionID(id))
		})
	}
}

func TestToolArchitectureNestedObject(t *testing.T) {
	object := toolArchitectureNestedObject()
	require.Len(t, object.Attributes, 4)

	assert.Equal(t, ObjectTypeForArchitectures(), object.Type(), "schema must match the shared architecture object type")

	for _, name := range []string{"url", "sha", "os", "arch"} {
		attribute, ok := object.Attributes[name].(schema.StringAttribute)
		require.True(t, ok, "%s must be a string attribute", name)
		assert.True(t, attribute.IsRequired(), "%s must be required", name)
		assert.NotEmpty(t, attribute.Validators, "%s must be validated", name)
	}

	ctx := context.Background()
	validate := func(name, value string) bool {
		attribute := object.Attributes[name].(schema.StringAttribute)
		for _, v := range attribute.Validators {
			resp := &validator.StringResponse{}
			v.ValidateString(ctx, validator.StringRequest{ConfigValue: types.StringValue(value)}, resp)
			if resp.Diagnostics.HasError() {
				return false
			}
		}
		return true
	}

	validSHA := strings.Repeat("a1", 32)
	cases := []struct {
		attribute string
		value     string
		valid     bool
	}{
		{"url", "https://example.com/tfpolicy-plugin/tfpolicy-plugin_linux_amd64.zip", true},
		{"url", "http://example.com/tfpolicy-plugin/tfpolicy-plugin_linux_amd64.zip", true},
		{"url", "ftp://example.com/tfpolicy-plugin/tfpolicy-plugin_linux_amd64.zip", false},
		{"sha", validSHA, true},
		{"sha", strings.ToUpper(validSHA), false},
		{"sha", validSHA[:63], false},
		{"os", "linux", true},
		{"os", "darwin", false},
		{"arch", "amd64", true},
		{"arch", "arm64", true},
		{"arch", "s390x", true},
		{"arch", "386", false},
	}
	for _, tc := range cases {
		t.Run(tc.attribute+"="+tc.value, func(t *testing.T) {
			assert.Equal(t, tc.valid, validate(tc.attribute, tc.value))
		})
	}
}

func TestToolArchitecturesSetRoundTrip(t *testing.T) {
	ctx := context.Background()
	archs := []ToolArchitecture{
		testToolArchitecture("linux", "amd64"),
		testToolArchitecture("linux", "arm64"),
	}

	set, diags := toolArchitecturesToSet(ctx, archs)
	require.False(t, diags.HasError(), "%v", diags)
	assert.Len(t, set.Elements(), 2)
	assert.Equal(t, ObjectTypeForArchitectures(), set.ElementType(ctx))

	decoded, diags := toolArchitecturesFromSet(ctx, set)
	require.False(t, diags.HasError(), "%v", diags)
	assert.ElementsMatch(t, archs, decoded)
}

func TestToolArchitecturesToSet_Empty(t *testing.T) {
	set, diags := toolArchitecturesToSet(context.Background(), []ToolArchitecture{})
	require.False(t, diags.HasError(), "%v", diags)
	assert.False(t, set.IsNull())
	assert.Empty(t, set.Elements())
}

func TestToolArchitecturesFromSet_NullOrUnknown(t *testing.T) {
	ctx := context.Background()
	for name, set := range map[string]types.Set{
		"null":    types.SetNull(ObjectTypeForArchitectures()),
		"unknown": types.SetUnknown(ObjectTypeForArchitectures()),
	} {
		t.Run(name, func(t *testing.T) {
			archs, diags := toolArchitecturesFromSet(ctx, set)
			assert.False(t, diags.HasError())
			assert.Nil(t, archs)
		})
	}
}

func TestToolArchitecturesFromSet_UnknownField(t *testing.T) {
	ctx := context.Background()
	element := types.ObjectValueMust(ObjectTypeForArchitectures().AttrTypes, map[string]attr.Value{
		"url":  types.StringUnknown(),
		"sha":  types.StringValue(strings.Repeat("a", 64)),
		"os":   types.StringValue("linux"),
		"arch": types.StringValue("amd64"),
	})
	set := types.SetValueMust(ObjectTypeForArchitectures(), []attr.Value{element})

	archs, diags := toolArchitecturesFromSet(ctx, set)
	require.False(t, diags.HasError(), "%v", diags)
	require.Len(t, archs, 1)
	assert.True(t, archs[0].URL.IsUnknown())
}

func TestValidateUniqueToolArchitectures(t *testing.T) {
	unknownArch := testToolArchitecture("linux", "amd64")
	unknownArch.Arch = types.StringUnknown()
	nullOS := testToolArchitecture("linux", "amd64")
	nullOS.OS = types.StringNull()

	cases := map[string]struct {
		archs      []ToolArchitecture
		wantErrors int
	}{
		"empty":           {nil, 0},
		"unique pairs":    {[]ToolArchitecture{testToolArchitecture("linux", "amd64"), testToolArchitecture("linux", "arm64"), testToolArchitecture("linux", "s390x")}, 0},
		"duplicate pair":  {[]ToolArchitecture{testToolArchitecture("linux", "amd64"), testToolArchitecture("linux", "amd64")}, 1},
		"repeated pair":   {[]ToolArchitecture{testToolArchitecture("linux", "amd64"), testToolArchitecture("linux", "amd64"), testToolArchitecture("linux", "amd64")}, 1},
		"unknown skipped": {[]ToolArchitecture{testToolArchitecture("linux", "amd64"), unknownArch}, 0},
		"null skipped":    {[]ToolArchitecture{testToolArchitecture("linux", "amd64"), nullOS}, 0},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			diags := validateUniqueToolArchitectures(tc.archs)
			assert.Equal(t, tc.wantErrors, diags.ErrorsCount())
			for _, d := range diags.Errors() {
				assert.Equal(t, "Duplicate Architecture", d.Summary())
				assert.Contains(t, d.Detail(), "linux/amd64")
			}
		})
	}
}

func TestValidateToolVersionDeprecation(t *testing.T) {
	cases := map[string]struct {
		deprecated types.Bool
		reason     types.String
		wantError  bool
	}{
		"not deprecated, null reason":      {types.BoolValue(false), types.StringNull(), false},
		"not deprecated, empty reason":     {types.BoolValue(false), types.StringValue(""), false},
		"not deprecated, reason set":       {types.BoolValue(false), types.StringValue("old"), true},
		"not deprecated, whitespace":       {types.BoolValue(false), types.StringValue(" "), true},
		"deprecated, reason set":           {types.BoolValue(true), types.StringValue("old"), false},
		"deprecated, null reason":          {types.BoolValue(true), types.StringNull(), false},
		"deprecated, empty reason":         {types.BoolValue(true), types.StringValue(""), true},
		"deprecated, whitespace reason":    {types.BoolValue(true), types.StringValue(" \t"), true},
		"deprecated, unknown reason":       {types.BoolValue(true), types.StringUnknown(), false},
		"unknown deprecated, empty reason": {types.BoolUnknown(), types.StringValue(""), false},
		"null deprecated, reason set":      {types.BoolNull(), types.StringValue("old"), false},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			diags := validateToolVersionDeprecation(tc.deprecated, tc.reason)
			assert.Equal(t, tc.wantError, diags.HasError(), "%v", diags)
			for _, d := range diags.Errors() {
				assert.Equal(t, "Invalid Deprecation Reason", d.Summary())
			}
		})
	}
}

func TestValidateToolArchitectures(t *testing.T) {
	withField := func(field string, value types.String) ToolArchitecture {
		arch := testToolArchitecture("linux", "amd64")
		switch field {
		case "url":
			arch.URL = value
		case "sha":
			arch.Sha = value
		case "os":
			arch.OS = value
		case "arch":
			arch.Arch = value
		}
		return arch
	}

	cases := map[string]struct {
		archs   []ToolArchitecture
		summary string
	}{
		"valid builds":     {[]ToolArchitecture{testToolArchitecture("linux", "amd64"), testToolArchitecture("linux", "arm64"), testToolArchitecture("linux", "s390x")}, ""},
		"no builds":        {nil, "Invalid Architectures"},
		"unknown url":      {[]ToolArchitecture{withField("url", types.StringUnknown())}, "Invalid Architecture"},
		"null sha":         {[]ToolArchitecture{withField("sha", types.StringNull())}, "Invalid Architecture"},
		"non-http url":     {[]ToolArchitecture{withField("url", types.StringValue("ftp://example.com/tfpolicy-plugin/tfpolicy-plugin_linux_amd64.zip"))}, "Invalid Architecture"},
		"uppercase sha":    {[]ToolArchitecture{withField("sha", types.StringValue(strings.Repeat("A", 64)))}, "Invalid Architecture"},
		"unsupported os":   {[]ToolArchitecture{withField("os", types.StringValue("darwin"))}, "Invalid Architecture"},
		"unsupported arch": {[]ToolArchitecture{withField("arch", types.StringValue("386"))}, "Invalid Architecture"},
		"duplicate pair":   {[]ToolArchitecture{testToolArchitecture("linux", "amd64"), testToolArchitecture("linux", "amd64")}, "Duplicate Architecture"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			diags := validateToolArchitectures(tc.archs)
			if tc.summary == "" {
				assert.False(t, diags.HasError(), "%v", diags)
				return
			}
			require.Equal(t, 1, diags.ErrorsCount(), "%v", diags)
			assert.Equal(t, tc.summary, diags.Errors()[0].Summary())
		})
	}
}

func TestValidateToolVersionRequest(t *testing.T) {
	known := types.BoolValue(false)

	diags := validateToolVersionRequest(types.StringValue("1.0.0"), known, known, known, known)
	assert.False(t, diags.HasError(), "%v", diags)

	for name, version := range map[string]types.String{
		"null":    types.StringNull(),
		"unknown": types.StringUnknown(),
		"blank":   types.StringValue("  "),
	} {
		t.Run("version "+name, func(t *testing.T) {
			diags := validateToolVersionRequest(version, known, known, known, known)
			require.Equal(t, 1, diags.ErrorsCount(), "%v", diags)
			assert.Equal(t, "Invalid Version", diags.Errors()[0].Summary())
		})
	}

	t.Run("unknown and null flags", func(t *testing.T) {
		diags := validateToolVersionRequest(types.StringValue("1.0.0"), types.BoolUnknown(), known, types.BoolNull(), known)
		require.Equal(t, 2, diags.ErrorsCount(), "%v", diags)
		assert.Contains(t, diags.Errors()[0].Detail(), "official")
		assert.Contains(t, diags.Errors()[1].Detail(), "beta")
	})
}

func TestPreserveEmptyDeprecatedReason(t *testing.T) {
	current := types.StringNull()
	cases := map[string]struct {
		deprecated types.Bool
		configured types.String
		want       types.String
	}{
		"not deprecated, configured empty": {types.BoolValue(false), types.StringValue(""), types.StringValue("")},
		"not deprecated, configured null":  {types.BoolValue(false), types.StringNull(), current},
		"deprecated, configured empty":     {types.BoolValue(true), types.StringValue(""), current},
		"unknown deprecated":               {types.BoolUnknown(), types.StringValue(""), current},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, preserveEmptyDeprecatedReason(tc.deprecated, current, tc.configured))
		})
	}

	t.Run("API reason kept", func(t *testing.T) {
		apiReason := types.StringValue("server default")
		assert.Equal(t, apiReason, preserveEmptyDeprecatedReason(types.BoolValue(true), apiReason, types.StringNull()))
	})
}
