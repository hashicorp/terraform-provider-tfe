package tfe

import (
	"os"
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

// skips a test if the test requires a paid feature, and this flag
// SKIP_PAID is set.
func skipIfFreeOnly(t *testing.T) {
	skip := os.Getenv("SKIP_PAID") == "1"
	if skip {
		t.Skip("Skipping test that requires a paid feature. Remove 'SKIP_PAID=1' if you want to run this test")
	}
}
