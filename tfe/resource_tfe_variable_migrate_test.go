// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tfe

import (
	"context"
	"reflect"
	"testing"

	tfe "github.com/hashicorp/go-tfe"
)

func testResourceTfeVariableStateDataV0() map[string]interface{} {
	return map[string]interface{}{
		"workspace_id": "hashicorp/a-workspace",
	}
}

func testResourceTfeVariableStateDataV1() map[string]interface{} {
	return map[string]interface{}{
		"workspace_id": "ws-123",
	}
}

func TestResourceTfeVariableStateUpgradeV0(t *testing.T) {
	client := testTfeClient(t, testClientOptions{defaultWorkspaceID: "ws-123"})
	name := "a-workspace"

	client.Workspaces.Create(context.TODO(), "hashicorp", tfe.WorkspaceCreateOptions{
		Name: &name,
	})

	expected := testResourceTfeVariableStateDataV1()
	actual, err := resourceTfeVariableStateUpgradeV0(context.Background(), testResourceTfeVariableStateDataV0(), ConfiguredClient{
		Client: client,
	})
	if err != nil {
		t.Fatalf("error migrating state: %s", err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("\n\nexpected:\n\n%#v\n\ngot:\n\n%#v\n\n", expected, actual)
	}
}
