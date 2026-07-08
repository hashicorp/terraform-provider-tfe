// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func testResourceTfeWorkspaceStateDataV0() map[string]interface{} {
	return map[string]interface{}{
		"id":          "hashicorp/test",
		"external_id": "ws-123",
	}
}

func testResourceTfeWorkspaceStateDataV1() map[string]interface{} {
	v0 := testResourceTfeWorkspaceStateDataV0()
	return map[string]interface{}{
		"id":          v0["external_id"],
		"external_id": v0["external_id"],
	}
}

func TestResourceTfeWorkspaceStateUpgradeV0(t *testing.T) {
	expected := testResourceTfeWorkspaceStateDataV1()
	actual, err := resourceTfeWorkspaceStateUpgradeV0(context.Background(), testResourceTfeWorkspaceStateDataV0(), nil)
	if err != nil {
		t.Fatalf("error migrating state: %s", err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("\n\nexpected:\n\n%#v\n\ngot:\n\n%#v\n\n", expected, actual)
	}
}

func TestResourceTFEWorkspace_UpgradeStateV0_FrameworkPath(t *testing.T) {
	ctx := context.Background()
	r := &resourceTFEWorkspaceFramework{}

	upgrader, ok := r.UpgradeState(ctx)[0]
	if !ok {
		t.Fatal("expected v0 upgrader")
	}
	if upgrader.PriorSchema == nil {
		t.Fatal("expected prior schema for v0 upgrader")
	}

	oldState := tfsdk.State{Schema: *upgrader.PriorSchema}
	oldData := modelWorkspaceV0{
		ID:              types.StringValue("hashicorp/workspace-test"),
		ExternalID:      types.StringValue("ws-123"),
		Name:            types.StringValue("workspace-test"),
		Organization:    types.StringValue("hashicorp"),
		TriggerPrefixes: types.ListNull(types.StringType),
		VCSRepo:         types.ListNull(types.ObjectType{AttrTypes: map[string]attr.Type{"identifier": types.StringType, "branch": types.StringType, "ingress_submodules": types.BoolType, "oauth_token_id": types.StringType, "github_app_installation_id": types.StringType}}),
	}
	if diags := oldState.Set(ctx, oldData); diags.HasError() {
		t.Fatalf("failed setting old state: %v", diags)
	}

	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)

	resp := &fwresource.UpgradeStateResponse{
		State: tfsdk.State{Schema: schemaResp.Schema},
	}
	req := fwresource.UpgradeStateRequest{State: &oldState}

	upgrader.StateUpgrader(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
	}

	var newData modelWorkspace
	if diags := resp.State.Get(ctx, &newData); diags.HasError() {
		t.Fatalf("failed reading upgraded state: %v", diags)
	}

	if newData.ID.ValueString() != "ws-123" {
		t.Fatalf("expected id to be ws-123, got %q", newData.ID.ValueString())
	}
}
