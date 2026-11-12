// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
)

// testAccFrameworkMigrationSteps returns test steps that prove a resource
// migrated from SDKv2 to the plugin framework behaves the same for existing
// users (docs/framework-migration.md):
//
//  1. Apply config with the released SDKv2 provider at sdkVersion and record
//     the state of every resource of resourceType.
//  2. Plan and apply the same config with this repository's provider. The
//     plan must be empty, and the state of every resource of resourceType
//     (attribute values, schema version, identity) must equal what step 1
//     recorded. Other resource types in the config are not compared, since
//     their state may legitimately differ between provider versions.
//
// sdkVersion is a released hashicorp/tfe version in which the resource was
// still implemented with SDKv2. Pin the latest release at the time of the
// migration and never change it afterward. To exercise state upgraders, add
// another call pinned to an older release that wrote the older schema
// version (its config must be valid for that release).
//
// Append further steps (updates, imports) to the returned slice as needed.
func testAccFrameworkMigrationSteps(t *testing.T, resourceType, sdkVersion, config string, opts ...frameworkMigrationOption) []resource.TestStep {
	t.Helper()

	cfg := frameworkMigrationConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}

	snapshot := &resourceStateSnapshot{
		resourceType:     resourceType,
		ignoreAttributes: cfg.ignoreAttributes,
	}

	return []resource.TestStep{
		{
			ExternalProviders: map[string]resource.ExternalProvider{
				"tfe": {
					Source:            "hashicorp/tfe",
					VersionConstraint: sdkVersion,
				},
			},
			Config:            config,
			ConfigStateChecks: []statecheck.StateCheck{snapshot.record()},
		},
		{
			ProtoV6ProviderFactories: testAccMuxedProviders,
			Config:                   config,
			ConfigPlanChecks: resource.ConfigPlanChecks{
				PreApply: []plancheck.PlanCheck{
					plancheck.ExpectEmptyPlan(),
				},
			},
			ConfigStateChecks: []statecheck.StateCheck{snapshot.compare()},
		},
	}
}

type frameworkMigrationConfig struct {
	ignoreAttributes []string
}

type frameworkMigrationOption func(*frameworkMigrationConfig)

// withIgnoredAttributes excludes top-level attributes from the state
// comparison. Legitimate uses:
//   - the value changes between two reads of the same remote object (for
//     example a timestamp the API updates on read)
//   - the attribute was added after sdkVersion (only when pinning an older
//     release to exercise state upgraders)
//
// Comment every use. Never use it to hide a difference introduced by the
// migration.
func withIgnoredAttributes(names ...string) frameworkMigrationOption {
	return func(c *frameworkMigrationConfig) {
		c.ignoreAttributes = append(c.ignoreAttributes, names...)
	}
}

// resourceStateSnapshot records the state of every resource of one type in
// one test step and compares it with the same resources in a later step.
type resourceStateSnapshot struct {
	resourceType     string
	ignoreAttributes []string

	recorded map[string]*tfjson.StateResource
}

func (s *resourceStateSnapshot) record() statecheck.StateCheck {
	return stateCheckFunc(func(state *tfjson.State) error {
		resources, err := stateResourcesOfType(state, s.resourceType)
		if err != nil {
			return err
		}
		s.recorded = resources
		return nil
	})
}

func (s *resourceStateSnapshot) compare() statecheck.StateCheck {
	return stateCheckFunc(func(state *tfjson.State) error {
		if s.recorded == nil {
			return fmt.Errorf("%s: no state was recorded in the SDKv2 step", s.resourceType)
		}
		current, err := stateResourcesOfType(state, s.resourceType)
		if err != nil {
			return err
		}

		addresses := make([]string, 0, len(s.recorded))
		for addr := range s.recorded {
			addresses = append(addresses, addr)
		}
		for addr := range current {
			if _, ok := s.recorded[addr]; !ok {
				addresses = append(addresses, addr)
			}
		}
		sort.Strings(addresses)

		var diffs []string
		for _, addr := range addresses {
			before, after := s.recorded[addr], current[addr]
			switch {
			case before == nil:
				diffs = append(diffs, addr+": not in SDKv2 state")
			case after == nil:
				diffs = append(diffs, addr+": missing after migration")
			default:
				for _, d := range diffStateResources(before, after, s.ignoreAttributes) {
					diffs = append(diffs, addr+": "+d)
				}
			}
		}
		if len(diffs) > 0 {
			return fmt.Errorf("state written by the framework implementation differs from SDKv2 state:\n  %s", strings.Join(diffs, "\n  "))
		}
		return nil
	})
}

// diffStateResources compares schema version, attribute values, and identity
// of two state entries for the same resource. Values come from
// `terraform show -json`, so null and zero values ("" / 0 / false / []) are
// distinguished.
func diffStateResources(before, after *tfjson.StateResource, ignoreAttributes []string) []string {
	var diffs []string

	if before.SchemaVersion != after.SchemaVersion {
		diffs = append(diffs, fmt.Sprintf("schema_version: %d -> %d", before.SchemaVersion, after.SchemaVersion))
	}

	ignored := make(map[string]bool, len(ignoreAttributes))
	for _, name := range ignoreAttributes {
		ignored[name] = true
	}
	diffs = append(diffs, diffValueMaps("attribute", before.AttributeValues, after.AttributeValues, ignored)...)

	beforeIDVersion, afterIDVersion := derefUint64(before.IdentitySchemaVersion), derefUint64(after.IdentitySchemaVersion)
	if beforeIDVersion != afterIDVersion {
		diffs = append(diffs, fmt.Sprintf("identity_schema_version: %s -> %s", beforeIDVersion, afterIDVersion))
	}
	diffs = append(diffs, diffValueMaps("identity", before.IdentityValues, after.IdentityValues, nil)...)

	return diffs
}

func diffValueMaps(kind string, before, after map[string]any, ignored map[string]bool) []string {
	keys := make(map[string]bool, len(before)+len(after))
	for k := range before {
		keys[k] = true
	}
	for k := range after {
		keys[k] = true
	}
	sorted := make([]string, 0, len(keys))
	for k := range keys {
		sorted = append(sorted, k)
	}
	sort.Strings(sorted)

	var diffs []string
	for _, k := range sorted {
		if ignored[k] {
			continue
		}
		b, inBefore := before[k]
		a, inAfter := after[k]
		switch {
		case !inBefore:
			diffs = append(diffs, fmt.Sprintf("%s %q: absent -> %#v", kind, k, a))
		case !inAfter:
			diffs = append(diffs, fmt.Sprintf("%s %q: %#v -> absent", kind, k, b))
		case !reflect.DeepEqual(b, a):
			diffs = append(diffs, fmt.Sprintf("%s %q: %#v -> %#v", kind, k, b, a))
		}
	}
	return diffs
}

func derefUint64(v *uint64) string {
	if v == nil {
		return "none"
	}
	return fmt.Sprintf("%d", *v)
}

// stateResourcesOfType returns the root module's managed resources of the
// given type, keyed by address. It is an error if there are none.
func stateResourcesOfType(state *tfjson.State, resourceType string) (map[string]*tfjson.StateResource, error) {
	resources := map[string]*tfjson.StateResource{}
	if state != nil && state.Values != nil && state.Values.RootModule != nil {
		for _, r := range state.Values.RootModule.Resources {
			if r.Mode == tfjson.ManagedResourceMode && r.Type == resourceType {
				resources[r.Address] = r
			}
		}
	}
	if len(resources) == 0 {
		return nil, fmt.Errorf("no %s resources found in state", resourceType)
	}
	return resources, nil
}

type stateCheckFunc func(*tfjson.State) error

func (f stateCheckFunc) CheckState(_ context.Context, req statecheck.CheckStateRequest, resp *statecheck.CheckStateResponse) {
	resp.Error = f(req.State)
}
