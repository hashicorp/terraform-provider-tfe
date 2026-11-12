// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"sort"
	"testing"
)

// TestSDKv2Inventory guards the SDKv2 -> plugin framework migration.
//
// The lists below are every resource and data source still served by the
// SDKv2 provider (provider.go). They may only shrink:
//
//   - New resources and data sources must be written with the plugin
//     framework (docs/new-resources.md), so adding to the SDKv2 maps fails.
//   - Migrating one to the framework removes it from the SDKv2 maps; remove it
//     from the list here in the same PR (docs/framework-migration.md).
var (
	sdkv2Resources = []string{
		"tfe_admin_organization_settings",
		"tfe_agent_pool",
		"tfe_agent_pool_allowed_projects",
		"tfe_agent_pool_allowed_workspaces",
		"tfe_agent_pool_excluded_workspaces",
		"tfe_agent_token",
		"tfe_no_code_module",
		"tfe_oauth_client",
		"tfe_organization",
		"tfe_organization_membership",
		"tfe_organization_module_sharing",
		"tfe_organization_token",
		"tfe_policy",
		"tfe_policy_set",
		"tfe_project_oauth_client",
		"tfe_project_policy_set",
		"tfe_project_variable_set",
		"tfe_registry_module",
		"tfe_run_trigger",
		"tfe_sentinel_policy",
		"tfe_team",
		"tfe_team_access",
		"tfe_team_member",
		"tfe_team_members",
		"tfe_team_organization_member",
		"tfe_team_organization_members",
		"tfe_team_project_access",
		"tfe_variable_set",
		"tfe_workspace",
		"tfe_workspace_policy_set",
		"tfe_workspace_policy_set_exclusion",
		"tfe_workspace_run",
		"tfe_workspace_variable_set",
	}

	sdkv2DataSources = []string{
		"tfe_agent_pool",
		"tfe_github_app_installation",
		"tfe_oauth_client",
		"tfe_organization",
		"tfe_organization_members",
		"tfe_organization_membership",
		"tfe_organization_tags",
		"tfe_organizations",
		"tfe_policy_set",
		"tfe_slug",
		"tfe_ssh_key",
		"tfe_team",
		"tfe_team_access",
		"tfe_team_project_access",
		"tfe_teams",
		"tfe_variable_set",
		"tfe_workspace",
		"tfe_workspace_ids",
	}
)

func TestSDKv2Inventory(t *testing.T) {
	p := Provider()

	resources := make([]string, 0, len(p.ResourcesMap))
	for name := range p.ResourcesMap {
		resources = append(resources, name)
	}
	dataSources := make([]string, 0, len(p.DataSourcesMap))
	for name := range p.DataSourcesMap {
		dataSources = append(dataSources, name)
	}

	checkSDKv2Inventory(t, "resource", sdkv2Resources, resources)
	checkSDKv2Inventory(t, "data source", sdkv2DataSources, dataSources)
}

func checkSDKv2Inventory(t *testing.T, kind string, expected, actual []string) {
	t.Helper()

	added, removed := diffNames(expected, actual)
	for _, name := range added {
		t.Errorf("%s %s was added to the SDKv2 provider; new resources and data sources must use the plugin framework (see docs/new-resources.md)", kind, name)
	}
	for _, name := range removed {
		t.Errorf("%s %s is no longer served by the SDKv2 provider; remove it from the list in sdkv2_inventory_test.go (see docs/framework-migration.md)", kind, name)
	}
}

// diffNames returns names present in actual but not expected (added) and in
// expected but not actual (removed), both sorted.
func diffNames(expected, actual []string) (added, removed []string) {
	want := make(map[string]bool, len(expected))
	for _, n := range expected {
		want[n] = true
	}
	got := make(map[string]bool, len(actual))
	for _, n := range actual {
		got[n] = true
		if !want[n] {
			added = append(added, n)
		}
	}
	for _, n := range expected {
		if !got[n] {
			removed = append(removed, n)
		}
	}
	sort.Strings(added)
	sort.Strings(removed)
	return added, removed
}
