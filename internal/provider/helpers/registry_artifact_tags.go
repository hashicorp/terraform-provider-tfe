// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package helpers

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-tfe"
)

// Registry artifact types supported for tag bindings.
const (
	ArtifactTypeRegistryModule    string = "registry-module"
	ArtifactTypeRegistryProvider  string = "registry-provider"
	ArtifactTypeRegistryComponent string = "registry-component"
)

// ListRegistryArtifactTagBindings fetches the current tag bindings for a registry
// artifact by dispatching on the artifact type. Returns tfe.ErrResourceNotFound if
// the artifact no longer exists.
func ListRegistryArtifactTagBindings(ctx context.Context, client *tfe.Client, artifactType, artifactID string) ([]*tfe.TagBinding, error) {
	switch artifactType {
	case ArtifactTypeRegistryModule:
		return client.RegistryModules.ListTagBindings(ctx, artifactID)
	case ArtifactTypeRegistryProvider:
		return client.RegistryProviders.ListTagBindings(ctx, artifactID)
	case ArtifactTypeRegistryComponent:
		return client.RegistryComponents.ListTagBindings(ctx, artifactID)
	default:
		return nil, fmt.Errorf("unsupported artifact type: %s", artifactType)
	}
}
