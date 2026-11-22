// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Regression test for hashicorp/terraform-provider-tfe#2296.
//
// Identities written by v0 of the tfe_registry_module identity schema record
// empty "namespace" and "registry_name" values because the create path builds
// the identity from the pre-ingest API response. The v0 -> v1 upgrader must
// backfill those values from the stored organization so that subsequent reads
// do not report an "Unexpected Identity Change".
func TestRegistryModuleIdentityUpgradeV0ToV1BackfillsNamespaceAndRegistryName(t *testing.T) {
	ctx := context.Background()
	r := resourceTFERegistryModule()

	identity := r.Identity
	if identity.Version != 1 {
		t.Fatalf("expected identity schema version 1, got %d", identity.Version)
	}

	upgrader := identity.IdentityUpgraders[0]
	if upgrader.Version != 0 {
		t.Fatalf("expected a version 0 identity upgrader, got version %d", upgrader.Version)
	}

	stored := map[string]interface{}{
		"id":              "mod-123",
		"hostname":        "app.terraform.io",
		"organization":    "example-org",
		"namespace":       "",
		"registry_name":   "",
		"name":            "example",
		"module_provider": "aws",
	}

	got, err := upgrader.Upgrade(ctx, stored, nil)
	if err != nil {
		t.Fatalf("unexpected identity upgrade error: %v", err)
	}

	for key, want := range map[string]string{
		"id":              "mod-123",
		"hostname":        "app.terraform.io",
		"organization":    "example-org",
		"namespace":       "example-org",
		"registry_name":   "private",
		"name":            "example",
		"module_provider": "aws",
	} {
		if value, ok := got[key].(string); !ok || value != want {
			t.Fatalf("expected %s to be %q, got %#v", key, want, got[key])
		}
	}
}

func TestRegistryModuleIdentityUpgradeV0ToV1PreservesPopulatedValues(t *testing.T) {
	ctx := context.Background()
	r := resourceTFERegistryModule()

	upgrader := r.Identity.IdentityUpgraders[0]
	stored := map[string]interface{}{
		"id":              "mod-123",
		"hostname":        "app.terraform.io",
		"organization":    "example-org",
		"namespace":       "other-org",
		"registry_name":   "public",
		"name":            "example",
		"module_provider": "aws",
	}

	got, err := upgrader.Upgrade(ctx, stored, nil)
	if err != nil {
		t.Fatalf("unexpected identity upgrade error: %v", err)
	}

	if got["namespace"] != "other-org" || got["registry_name"] != "public" {
		t.Fatalf("expected populated identity values to be preserved, got %#v", got)
	}
}

func TestRegistryModuleIdentityUpgradeV0ToV1HandlesNilIdentity(t *testing.T) {
	ctx := context.Background()
	r := resourceTFERegistryModule()

	upgrader := r.Identity.IdentityUpgraders[0]
	got, err := upgrader.Upgrade(ctx, nil, nil)
	if err != nil {
		t.Fatalf("unexpected identity upgrade error: %v", err)
	}
	if got != nil {
		t.Fatalf("expected a nil identity to remain nil, got %#v", got)
	}
}

// Ensure the resource's identity schema still exposes all import attributes.
func TestRegistryModuleIdentitySchemaHasRequiredAttributes(t *testing.T) {
	r := resourceTFERegistryModule()
	for _, attribute := range []string{"id", "hostname", "organization", "registry_name", "namespace", "name", "module_provider"} {
		if _, ok := r.Identity.SchemaMap()[attribute]; !ok {
			t.Fatalf("expected identity schema to contain %q", attribute)
		}
	}
	_ = schema.TypeString // keep the import explicit for the schema package
}
