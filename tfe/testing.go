package tfe

import (
	"testing"

	tfe "github.com/hashicorp/go-tfe"
)

func testTfeClient(t *testing.T) *tfe.Client {
	config := &tfe.Config{
		Token: "not-a-token",
	}

	client, err := tfe.NewClient(config)
	if err != nil {
		t.Fatalf("error creating tfe client: %v", err)
	}

	client.Workspaces = newMockWorkspaces()

	return client
}
