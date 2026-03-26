// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTFETeamDataSource_basic(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamDataSourceConfig_basic(rInt, org.Name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.tfe_team.foobar", "name", fmt.Sprintf("team-test-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.tfe_team.foobar", "organization", org.Name),
					resource.TestCheckResourceAttr(
						"data.tfe_team.foobar", "organization_access.0.manage_policies", "true"),
					resource.TestCheckResourceAttr(
						"data.tfe_team.foobar", "organization_access.0.manage_policy_overrides", "true"),
					resource.TestCheckResourceAttr(
						"data.tfe_team.foobar", "organization_access.0.delegate_policy_overrides", "true"),
					resource.TestCheckResourceAttr(
						"data.tfe_team.foobar", "organization_access.0.manage_workspaces", "true"),
					resource.TestCheckResourceAttr(
						"data.tfe_team.foobar", "organization_access.0.manage_vcs_settings", "true"),
					resource.TestCheckResourceAttr(
						"data.tfe_team.foobar", "organization_access.0.manage_providers", "true"),
					resource.TestCheckResourceAttr(
						"data.tfe_team.foobar", "organization_access.0.manage_modules", "true"),
					resource.TestCheckResourceAttr(
						"data.tfe_team.foobar", "organization_access.0.manage_run_tasks", "true"),
					resource.TestCheckResourceAttr(
						"data.tfe_team.foobar", "organization_access.0.manage_projects", "true"),
					resource.TestCheckResourceAttr(
						"data.tfe_team.foobar", "organization_access.0.read_projects", "true"),
					resource.TestCheckResourceAttr(
						"data.tfe_team.foobar", "organization_access.0.read_workspaces", "true"),
					resource.TestCheckResourceAttr(
						"data.tfe_team.foobar", "organization_access.0.manage_membership", "true"),
					resource.TestCheckResourceAttr(
						"data.tfe_team.foobar", "organization_access.0.manage_teams", "true"),
					resource.TestCheckResourceAttr(
						"data.tfe_team.foobar", "organization_access.0.manage_organization_access", "true"),
					resource.TestCheckResourceAttr(
						"data.tfe_team.foobar", "organization_access.0.access_secret_teams", "true"),
					resource.TestCheckResourceAttr(
						"data.tfe_team.foobar", "organization_access.0.manage_agent_pools", "true"),
					resource.TestCheckResourceAttrSet("data.tfe_team.foobar", "id"),
				),
			},
		},
	})
}

func TestAccTFETeamDataSource_ssoTeamId(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	testSsoTeamID := fmt.Sprintf("sso-team-id-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamDataSourceConfig_ssoTeamId(rInt, org.Name, testSsoTeamID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.tfe_team.sso_team", "name", fmt.Sprintf("team-test-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.tfe_team.sso_team", "organization", org.Name),
					resource.TestCheckResourceAttrSet("data.tfe_team.sso_team", "id"),
					resource.TestCheckResourceAttr(
						"data.tfe_team.sso_team", "sso_team_id", testSsoTeamID),
				),
			},
		},
	})
}

func TestFlattenTeamOrganizationAccess(t *testing.T) {
	organizationAccess := flattenTeamOrganizationAccess(&tfe.OrganizationAccess{
		ManagePolicies:           true,
		ManagePolicyOverrides:    true,
		DelegatePolicyOverrides:  true,
		ManageWorkspaces:         true,
		ManageVCSSettings:        true,
		ManageProviders:          true,
		ManageModules:            true,
		ManageRunTasks:           true,
		ManageProjects:           true,
		ReadProjects:             true,
		ReadWorkspaces:           true,
		ManageMembership:         true,
		ManageTeams:              true,
		ManageOrganizationAccess: true,
		AccessSecretTeams:        true,
		ManageAgentPools:         true,
	})

	expected := []map[string]bool{{
		"manage_policies":            true,
		"manage_policy_overrides":    true,
		"delegate_policy_overrides":  true,
		"manage_workspaces":          true,
		"manage_vcs_settings":        true,
		"manage_providers":           true,
		"manage_modules":             true,
		"manage_run_tasks":           true,
		"manage_projects":            true,
		"read_projects":              true,
		"read_workspaces":            true,
		"manage_membership":          true,
		"manage_teams":               true,
		"manage_organization_access": true,
		"access_secret_teams":        true,
		"manage_agent_pools":         true,
	}}

	if !reflect.DeepEqual(organizationAccess, expected) {
		t.Fatalf("expected organization access %#v, got %#v", expected, organizationAccess)
	}
}

func testAccTFETeamDataSourceConfig_basic(rInt int, organization string) string {
	return fmt.Sprintf(`
resource "tfe_team" "foobar" {
  name         = "team-test-%d"
  organization = "%s"

  organization_access {
    manage_policies            = true
    manage_policy_overrides    = true
    delegate_policy_overrides  = true
    manage_workspaces          = true
    manage_vcs_settings        = true
    manage_providers           = true
    manage_modules             = true
    manage_run_tasks           = true
    manage_projects            = true
    read_projects              = true
    read_workspaces            = true
    manage_membership          = true
    manage_teams               = true
    manage_organization_access = true
    access_secret_teams        = true
    manage_agent_pools         = true
  }
}

data "tfe_team" "foobar" {
  name         = tfe_team.foobar.name
  organization = "%s"
}`, rInt, organization, organization)
}

func testAccTFETeamDataSourceConfig_ssoTeamId(rInt int, organization string, ssoTeamID string) string {
	return fmt.Sprintf(`
resource "tfe_team" "sso_team" {
  name         = "team-test-%d"
  organization = "%s"
  sso_team_id  = "%s"
}

data "tfe_team" "sso_team" {
  name         = tfe_team.sso_team.name
  organization = tfe_team.sso_team.organization
}`, rInt, organization, ssoTeamID)
}
