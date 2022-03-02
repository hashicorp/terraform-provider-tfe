package tfe

import (
	"context"
	"fmt"
	"strings"

	tfe "github.com/hashicorp/go-tfe"
)

// fetchWorkspaceExternalID returns the external id for a workspace
// when given a workspace id of the form ORGANIZATION_AME/WORKSPACE_NAME
func fetchVariableSetExternalID(id string, client *tfe.Client) (string, error) {
	orgName, vsId, err := unpackVariableSetID(id)
	if err != nil {
		return "", fmt.Errorf("Error unpacking variable set ID: %v", err)
	}

	vs, err := client.VariableSets.Read(ctx, orgName, vsId)
	if err != nil {
		return "", fmt.Errorf("Error reading configuration of variable set %s: %v", id, err)
	}

	return vs.ID, nil
}

func unpackVariableSetID(id string) (organization, name string, err error) {
	s := strings.SplitN(id, "/", 2)
	if len(s) != 2 {
		return "", "", fmt.Errorf(
			"invalid workspace ID format: %s (expected <ORGANIZATION>/<VARUIABLE SET>)", id)
	}

	return s[0], s[1], nil
}
