package tfe

func testResourceTfeVariableStateDataV0() map[string]interface{} {
	return map[string]interface{}{
		"workspace_id": "hashicorp/workspace",
	}
}

func testResourceTfeVariableStateDataV1() map[string]interface{} {
	return map[string]interface{}{
		"workspace_id": "ws-123",
	}
}

//TODO: how would I test this since I have to pass a tfe.Client as the meta? I don't want to mock out the entire client.

// func TestResourceTfeVariableStateUpgradeV0(t *testing.T) {
// 	expected := testResourceTfeVariableStateDataV1()
// 	actual, err := resourceTfeVariableStateUpgradeV0(testResourceTfeVariableStateDataV0(), nil)
// 	if err != nil {
// 		t.Fatalf("error migrating state: %s", err)
// 	}
//
// 	if !reflect.DeepEqual(expected, actual) {
// 		t.Fatalf("\n\nexpected:\n\n%#v\n\ngot:\n\n%#v\n\n", expected, actual)
// 	}
// }
