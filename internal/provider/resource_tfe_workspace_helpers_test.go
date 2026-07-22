// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/hashicorp/jsonapi"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestWorkspaceIDRegexp(t *testing.T) {
	if !workspaceIDRegexp.MatchString("ws-abcDEF1234567890") {
		t.Fatal("expected workspaceIDRegexp to match valid workspace id")
	}
	if workspaceIDRegexp.MatchString("not-a-workspace-id") {
		t.Fatal("expected workspaceIDRegexp to reject invalid workspace id")
	}
}

func TestValidTagName(t *testing.T) {
	if !validTagName("tag_name-1") {
		t.Fatal("expected validTagName to accept valid tag")
	}
	if validTagName("-bad") {
		t.Fatal("expected validTagName to reject invalid tag")
	}
}

func TestHasConfiguredTriggerConflict(t *testing.T) {
	prefixes := types.ListValueMust(types.StringType, []attr.Value{})
	patterns := types.ListValueMust(types.StringType, []attr.Value{})
	if !hasConfiguredTriggerConflict(prefixes, patterns) {
		t.Fatal("expected configured empty lists to conflict")
	}

	if hasConfiguredTriggerConflict(types.ListNull(types.StringType), patterns) {
		t.Fatal("expected null prefixes to not conflict")
	}
}

func TestFlattenAutoDestroyAt(t *testing.T) {
	v, err := flattenAutoDestroyAt(jsonapi.NullableAttr[time.Time]{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != nil {
		t.Fatalf("expected nil for unspecified value, got %v", *v)
	}

	now := time.Now().UTC().Truncate(time.Second)
	v, err = flattenAutoDestroyAt(jsonapi.NewNullableAttrWithValue(now))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v == nil || *v != now.Format(time.RFC3339) {
		t.Fatalf("unexpected flatten result: %#v", v)
	}
}

type safeDeleteMockWorkspaces struct {
	*mockWorkspaces
	err error
}

func (m *safeDeleteMockWorkspaces) SafeDeleteByID(ctx context.Context, workspaceID string) error {
	return m.err
}

func TestSafeWorkspaceDelete(t *testing.T) {
	client := testTfeClient(t, testClientOptions{})
	client.Workspaces = &safeDeleteMockWorkspaces{mockWorkspaces: newMockWorkspaces(testClientOptions{}), err: nil}

	if err := safeWorkspaceDelete(context.Background(), ConfiguredClient{Client: client}, "ws-any"); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	expected := errors.New("boom")
	client.Workspaces = &safeDeleteMockWorkspaces{mockWorkspaces: newMockWorkspaces(testClientOptions{}), err: expected}
	if err := safeWorkspaceDelete(context.Background(), ConfiguredClient{Client: client}, "ws-any"); !errors.Is(err, expected) {
		t.Fatalf("expected %v, got %v", expected, err)
	}
}

func TestWorkspaceDeleteErrorHelpers(t *testing.T) {
	if err := errWorkspaceSafeDeleteWithPermission("ws-123", errors.New("conflict test")); err == nil {
		t.Fatal("expected wrapped conflict error")
	}
	if err := errWorkspaceResourceCountCheck("ws-123", 2); err == nil {
		t.Fatal("expected error for managed resources")
	}
	if err := errWorkspaceResourceCountCheck("ws-123", 0); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestExpandWorkspaceTagBindings(t *testing.T) {
	m := types.MapValueMust(types.StringType, map[string]attr.Value{
		"a": types.StringValue("1"),
		"b": types.StringValue("2"),
	})
	bindings := expandWorkspaceTagBindings(m)
	if len(bindings) != 2 {
		t.Fatalf("expected 2 bindings, got %d", len(bindings))
	}
}

func TestExpandWorkspaceStringList(t *testing.T) {
	d := &diag.Diagnostics{}
	vals, ok := expandWorkspaceStringList(context.Background(), types.ListNull(types.StringType), d)
	if ok || vals != nil {
		t.Fatal("expected null list to return not configured")
	}

	l := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("x"), types.StringValue("y")})
	vals, ok = expandWorkspaceStringList(context.Background(), l, d)
	if !ok || len(vals) != 2 || vals[0] != "x" || vals[1] != "y" {
		t.Fatalf("unexpected list expansion: %#v", vals)
	}
}

func TestExpandWorkspaceTagNames(t *testing.T) {
	d := &diag.Diagnostics{}
	set := types.SetValueMust(types.StringType, []attr.Value{types.StringValue("alpha")})
	tags, ok := expandWorkspaceTagNames(context.Background(), set, d)
	if !ok || len(tags) != 1 || tags[0].Name != "alpha" {
		t.Fatalf("unexpected tag expansion: %#v", tags)
	}
}

func TestExpandWorkspaceVCSRepoOptions(t *testing.T) {
	ctx := context.Background()
	d := &diag.Diagnostics{}
	objType := types.ObjectType{AttrTypes: map[string]attr.Type{
		"identifier":                 types.StringType,
		"branch":                     types.StringType,
		"ingress_submodules":         types.BoolType,
		"oauth_token_id":             types.StringType,
		"tags_regex":                 types.StringType,
		"github_app_installation_id": types.StringType,
	}}

	list, listDiags := types.ListValueFrom(ctx, objType, []modelWorkspaceVCSRepo{{
		Identifier:              types.StringValue("owner/repo"),
		Branch:                  types.StringValue("main"),
		IngressSubmodules:       types.BoolValue(true),
		OAuthTokenID:            types.StringValue("ot-123"),
		TagsRegex:               types.StringValue("v.*"),
		GithubAppInstallationID: types.StringValue("gha-123"),
	}})
	d.Append(listDiags...)

	createOpts := expandWorkspaceVCSRepoOptions(ctx, list, d, false)
	if createOpts == nil || *createOpts.Identifier != "owner/repo" || *createOpts.Branch != "main" {
		t.Fatalf("unexpected create vcs opts: %#v", createOpts)
	}

	updateOpts := expandWorkspaceVCSRepoOptions(ctx, list, d, true)
	if updateOpts == nil || *updateOpts.OAuthTokenID != "ot-123" || *updateOpts.GHAInstallationID != "gha-123" {
		t.Fatalf("unexpected update vcs opts: %#v", updateOpts)
	}
}

func TestSimpleConversionHelpers(t *testing.T) {
	if v := stringToFramework(""); !v.IsNull() {
		t.Fatal("expected null string")
	}
	if v := stringToFramework("x"); v.ValueString() != "x" {
		t.Fatal("expected string value")
	}

	l := stringSliceToList([]string{"a", "b"})
	if l.IsNull() || len(l.Elements()) != 2 {
		t.Fatal("unexpected list conversion")
	}

	s := stringSliceToSet([]string{"a", "b"})
	if s.IsNull() || len(s.Elements()) != 2 {
		t.Fatal("unexpected set conversion")
	}

	m := mapTypeFromStringMap(map[string]interface{}{"k": "v"})
	if m.IsNull() || len(m.Elements()) != 1 {
		t.Fatal("unexpected map conversion")
	}
	back := mapFromStringMapType(m)
	if back["k"] != "v" {
		t.Fatalf("unexpected map round trip: %#v", back)
	}

	if !boolValueOrDefault(types.BoolNull(), true) {
		t.Fatal("expected default bool")
	}
	if stringValueOrDefault(types.StringNull(), "d") != "d" {
		t.Fatal("expected default string")
	}
}
