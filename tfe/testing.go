package tfe

import (
	"testing"

	tfe "github.com/hashicorp/go-tfe"
)

// testTfeClient creates a mock client that creates workspaces with their ID
// set to workspaceID.
func testTfeClient(t *testing.T, workspaceID string) *tfe.Client {
	config := &tfe.Config{
		Token: "not-a-token",
	}

	client, err := tfe.NewClient(config)
	if err != nil {
		t.Fatalf("error creating tfe client: %v", err)
	}

	client.Workspaces = newMockWorkspaces(workspaceID)

	return client
}
