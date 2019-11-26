package tfe

import (
	"context"
	"fmt"
	"strings"

	tfe "github.com/hashicorp/go-tfe"
)

// fetchWorkspaceExternalID returns the external id for a workspace
// when given a workspace id of the form ORGANIZATION_AME/WORKSPACE_NAME
func fetchWorkspaceExternalID(id string, client *tfe.Client) (string, error) {
	orgName, wsName, err := unpackWorkspaceID(id)
	if err != nil {
		return "", fmt.Errorf("Error unpacking workspace ID: %v", err)
	}

	workspace, err := client.Workspaces.Read(ctx, orgName, wsName)
	if err != nil {
		return "", fmt.Errorf("Error reading configuration of workspace %s: %v", id, err)
	}

	return workspace.ID, nil
}

type workspaceIDReader interface {
	ReadByID(context.Context, string) (*tfe.Workspace, error)
}

// fetchWorkspaceHumanID returns the human readable id "org/workspace"
// when given a workspace external id
func fetchWorkspaceHumanID(id string, r workspaceIDReader) (string, error) {
	workspace, err := r.ReadByID(ctx, id)
	if err != nil {
		return "", fmt.Errorf("Error reading configuration of workspace %s: %v", id, err)
	}

	humanID, err := packWorkspaceID(workspace)
	if err != nil {
		return "", fmt.Errorf("Error creating human ID for workspace %s: %v", id, err)
	}

	return humanID, nil
}

func packWorkspaceID(w *tfe.Workspace) (id string, err error) {
	if w.Organization == nil {
		return "", fmt.Errorf("no organization in workspace response")
	}
	return w.Organization.Name + "/" + w.Name, nil
}

func unpackWorkspaceID(id string) (organization, name string, err error) {
	// Support the old ID format for backwards compatibitily.
	if s := strings.SplitN(id, "|", 2); len(s) == 2 {
		return s[1], s[0], nil
	}

	s := strings.SplitN(id, "/", 2)
	if len(s) != 2 {
		return "", "", fmt.Errorf(
			"invalid workspace ID format: %s (expected <ORGANIZATION>/<WORKSPACE>)", id)
	}

	return s[0], s[1], nil
}
