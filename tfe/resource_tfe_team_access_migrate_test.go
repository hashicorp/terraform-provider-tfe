package tfe

func testResourceTfeTeamAccessStateDataV0() map[string]interface{} {
	return map[string]interface{}{
		"workspace_id": "hashicorp/workspace",
	}
}

func testResourceTfeTeamAccessStateDataV1() map[string]interface{} {
	return map[string]interface{}{
		"workspace_id": "ws-123",
	}
}

//TODO: how would I test this since I have to pass a tfe.Client as the meta? I don't want to mock out the entire client.

// func TestResourceTfeTeamAccessStateUpgradeV0(t *testing.T) {
// 	expected := testResourceTfeTeamAccessStateDataV1()
// 	actual, err := resourceTfeTeamAccessStateUpgradeV0(testResourceTfeTeamAccessStateDataV0(), nil)
// 	if err != nil {
// 		t.Fatalf("error migrating state: %s", err)
// 	}
//
// 	if !reflect.DeepEqual(expected, actual) {
// 		t.Fatalf("\n\nexpected:\n\n%#v\n\ngot:\n\n%#v\n\n", expected, actual)
// 	}
// }
