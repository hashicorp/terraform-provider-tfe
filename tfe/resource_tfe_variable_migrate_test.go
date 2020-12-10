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
	client := testTfeClient(t)
	name := "a-workspace"
	client.Workspaces.Create(nil, "hashicorp", tfe.WorkspaceCreateOptions{
		ID:   "ws-123",
		Name: &name,
	})

	expected := testResourceTfeVariableStateDataV1()
	actual, err := resourceTfeVariableStateUpgradeV0(context.Background(), testResourceTfeVariableStateDataV0(), client)
	if err != nil {
		t.Fatalf("error migrating state: %s", err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("\n\nexpected:\n\n%#v\n\ngot:\n\n%#v\n\n", expected, actual)
	}
}
