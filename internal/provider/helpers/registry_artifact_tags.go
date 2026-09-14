// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package helpers

import (
	"context"
	"fmt"
	"sort"

	"github.com/hashicorp/go-tfe"
)

// Registry artifact types supported for tag bindings.
const (
	ArtifactTypeRegistryModule    string = "registry-module"
	ArtifactTypeRegistryProvider  string = "registry-provider"
	ArtifactTypeRegistryComponent string = "registry-component"
)

// tagBindingService is satisfied by any go-tfe registry service that can list the
// tag bindings of an artifact.
type tagBindingService interface {
	ListTagBindings(ctx context.Context, id string) ([]*tfe.TagBinding, error)
}

// artifactTagBindingServices maps each supported artifact type to the go-tfe
// service that manages its tag bindings. This is the single source of truth for
// which artifact types support tagging.
var artifactTagBindingServices = map[string]func(*tfe.Client) tagBindingService{
	ArtifactTypeRegistryModule:    func(c *tfe.Client) tagBindingService { return c.RegistryModules },
	ArtifactTypeRegistryProvider:  func(c *tfe.Client) tagBindingService { return c.RegistryProviders },
	ArtifactTypeRegistryComponent: func(c *tfe.Client) tagBindingService { return c.RegistryComponents },
}

// RegistryArtifactTypes returns the sorted list of artifact types that support tag
// bindings, so callers do not need to hardcode the list.
func RegistryArtifactTypes() []string {
	out := make([]string, 0, len(artifactTagBindingServices))
	for artifactType := range artifactTagBindingServices {
		out = append(out, artifactType)
	}
	sort.Strings(out)
	return out
}

// ListRegistryArtifactTagBindings fetches the current tag bindings for a registry
// artifact by dispatching on the artifact type. Returns tfe.ErrResourceNotFound if
// the artifact no longer exists.
func ListRegistryArtifactTagBindings(ctx context.Context, client *tfe.Client, artifactType, artifactID string) ([]*tfe.TagBinding, error) {
	newSvc, ok := artifactTagBindingServices[artifactType]
	if !ok {
		return nil, fmt.Errorf("unsupported artifact type: %s", artifactType)
	}
	return newSvc(client).ListTagBindings(ctx, artifactID)
}
