package tfe

import (
	"testing"

	tfe "github.com/hashicorp/go-tfe"
)

type testClientOptions struct {
	defaultWorkspaceID           string
	remoteStateConsumersResponse string
}

// testTfeClient creates a mock client that creates workspaces with their ID
// set to workspaceID.
func testTfeClient(t *testing.T, options testClientOptions) *tfe.Client {
	config := &tfe.Config{
		Token: "not-a-token",
	}

	client, err := tfe.NewClient(config)
	if err != nil {
		t.Fatalf("error creating tfe client: %v", err)
	}

	client.Workspaces = newMockWorkspaces(options)

	return client
}
