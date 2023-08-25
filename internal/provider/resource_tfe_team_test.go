// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"math/rand"
	"regexp"
	"testing"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccTFETeam_basic(t *testing.T) {
	team := &tfe.Team{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFETeamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeam_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETeamExists(
						"tfe_team.foobar", team),
					testAccCheckTFETeamAttributes_basic(team),
					resource.TestCheckResourceAttr(
						"tfe_team.foobar", "name", "team-test"),
				),
			},
		},
	})
}

func TestAccTFETeam_full(t *testing.T) {
	team := &tfe.Team{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFETeamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeam_full(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETeamExists(
						"tfe_team.foobar", team),
					testAccCheckTFETeamAttributes_full(team),
					resource.TestCheckResourceAttr(
						"tfe_team.foobar", "name", "team-test"),
					resource.TestCheckResourceAttr(
						"tfe_team.foobar", "visibility", "organization"),
					resource.TestCheckResourceAttr(
						"tfe_team.foobar", "organization_access.0.manage_policies", "true"),
					resource.TestCheckResourceAttr(
						"tfe_team.foobar", "organization_access.0.manage_policy_overrides", "true"),
					resource.TestCheckResourceAttr(
						"tfe_team.foobar", "organization_access.0.manage_workspaces", "true"),
					resource.TestCheckResourceAttr(
						"tfe_team.foobar", "organization_access.0.manage_vcs_settings", "true"),
					resource.TestCheckResourceAttr(
						"tfe_team.foobar", "organization_access.0.manage_providers", "true"),
					resource.TestCheckResourceAttr(
						"tfe_team.foobar", "organization_access.0.manage_modules", "true"),
					resource.TestCheckResourceAttr(
						"tfe_team.foobar", "organization_access.0.manage_run_tasks", "true"),
					resource.TestCheckResourceAttr(
						"tfe_team.foobar", "organization_access.0.manage_projects", "true"),
					resource.TestCheckResourceAttr(
						"tfe_team.foobar", "organization_access.0.read_projects", "true"),
					resource.TestCheckResourceAttr(
						"tfe_team.foobar", "organization_access.0.read_workspaces", "true"),
					resource.TestCheckResourceAttr(
						"tfe_team.foobar", "organization_access.0.manage_membership", "true"),
				),
			},
		},
	})
}

func TestAccTFETeam_full_update(t *testing.T) {
	team := &tfe.Team{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFETeamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeam_full(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETeamExists(
						"tfe_team.foobar", team),
					testAccCheckTFETeamAttributes_full(team),
					resource.TestCheckResourceAttr(
						"tfe_team.foobar", "name", "team-test"),
					resource.TestCheckResourceAttr(
						"tfe_team.foobar", "visibility", "organization"),
					resource.TestCheckResourceAttr(
						"tfe_team.foobar", "organization_access.0.manage_policies", "true"),
					resource.TestCheckResourceAttr(
						"tfe_team.foobar", "organization_access.0.manage_policy_overrides", "true"),
					resource.TestCheckResourceAttr(
						"tfe_team.foobar", "organization_access.0.manage_workspaces", "true"),
					resource.TestCheckResourceAttr(
						"tfe_team.foobar", "organization_access.0.manage_vcs_settings", "true"),
					resource.TestCheckResourceAttr(
						"tfe_team.foobar", "organization_access.0.manage_providers", "true"),
					resource.TestCheckResourceAttr(
						"tfe_team.foobar", "organization_access.0.manage_modules", "true"),
					resource.TestCheckResourceAttr(
						"tfe_team.foobar", "organization_access.0.manage_run_tasks", "true"),
					resource.TestCheckResourceAttr(
						"tfe_team.foobar", "organization_access.0.read_projects", "true"),
					resource.TestCheckResourceAttr(
						"tfe_team.foobar", "organization_access.0.manage_projects", "true"),
					resource.TestCheckResourceAttr(
						"tfe_team.foobar", "organization_access.0.read_workspaces", "true"),
					resource.TestCheckResourceAttr(
						"tfe_team.foobar", "organization_access.0.manage_membership", "true"),
				),
			},
			{
				Config: testAccTFETeam_full_update(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETeamExists(
						"tfe_team.foobar", team),
					testAccCheckTFETeamAttributes_full_update(team),
					resource.TestCheckResourceAttr(
						"tfe_team.foobar", "name", "team-test-1"),
					resource.TestCheckResourceAttr(
						"tfe_team.foobar", "visibility", "secret"),
					resource.TestCheckResourceAttr(
						"tfe_team.foobar", "organization_access.0.manage_policies", "false"),
					resource.TestCheckResourceAttr(
						"tfe_team.foobar", "organization_access.0.manage_policy_overrides", "false"),
					resource.TestCheckResourceAttr(
						"tfe_team.foobar", "organization_access.0.manage_workspaces", "false"),
					resource.TestCheckResourceAttr(
						"tfe_team.foobar", "organization_access.0.manage_providers", "false"),
					resource.TestCheckResourceAttr(
						"tfe_team.foobar", "organization_access.0.manage_modules", "false"),
					resource.TestCheckResourceAttr(
						"tfe_team.foobar", "organization_access.0.manage_run_tasks", "false"),
					resource.TestCheckResourceAttr(
						"tfe_team.foobar", "organization_access.0.manage_projects", "false"),
					resource.TestCheckResourceAttr(
						"tfe_team.foobar", "organization_access.0.read_workspaces", "false"),
					resource.TestCheckResourceAttr(
						"tfe_team.foobar", "organization_access.0.read_projects", "false"),
					resource.TestCheckResourceAttr(
						"tfe_team.foobar", "sso_team_id", "changed-sso-id"),
					resource.TestCheckResourceAttr(
						"tfe_team.foobar", "organization_access.0.manage_membership", "false"),
				),
			},
			{
				Config: testAccTFETeam_full_update_clear(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETeamExists(
						"tfe_team.foobar", team),
					resource.TestCheckResourceAttr(
						"tfe_team.foobar", "name", "team-test-1"),
					resource.TestCheckResourceAttr(
						"tfe_team.foobar", "visibility", "secret"),
					resource.TestCheckResourceAttr(
						"tfe_team.foobar", "organization_access.0.manage_policies", "false"),
					resource.TestCheckResourceAttr(
						"tfe_team.foobar", "organization_access.0.manage_policy_overrides", "false"),
					resource.TestCheckResourceAttr(
						"tfe_team.foobar", "organization_access.0.manage_workspaces", "false"),
					resource.TestCheckResourceAttr(
						"tfe_team.foobar", "organization_access.0.manage_vcs_settings", "false"),
					resource.TestCheckResourceAttr(
						"tfe_team.foobar", "organization_access.0.manage_providers", "false"),
					resource.TestCheckResourceAttr(
						"tfe_team.foobar", "organization_access.0.manage_modules", "false"),
					resource.TestCheckResourceAttr(
						"tfe_team.foobar", "organization_access.0.manage_run_tasks", "false"),
					resource.TestCheckResourceAttr(
						"tfe_team.foobar", "organization_access.0.manage_projects", "false"),
					resource.TestCheckResourceAttr(
						"tfe_team.foobar", "organization_access.0.read_workspaces", "false"),
					resource.TestCheckResourceAttr(
						"tfe_team.foobar", "organization_access.0.read_projects", "false"),
					resource.TestCheckResourceAttr(
						"tfe_team.foobar", "sso_team_id", ""),
					resource.TestCheckResourceAttr(
						"tfe_team.foobar", "organization_access.0.manage_membership", "false"),
				),
			},
		},
	})
}

func TestAccTFETeam_import_byId(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFETeamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeam_basic(rInt),
			},

			{
				ResourceName:        "tfe_team.foobar",
				ImportState:         true,
				ImportStateIdPrefix: fmt.Sprintf("tst-terraform-%d/", rInt),
				ImportStateVerify:   true,
			},
		},
	})
}

func TestAccTFETeam_import_byId_doesNotExist(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFETeamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeam_basic(rInt),
			},

			{
				ResourceName:  "tfe_team.foobar",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("tst-terraform-%d/team-1234567891234567", rInt),
				ExpectError:   regexp.MustCompile("no team found with name or ID team-1234567891234567 in organization"),
			},
		},
	})
}

func TestAccTFETeam_import_byName(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFETeamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeam_basic(rInt),
			},

			{
				ResourceName:      "tfe_team.foobar",
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("tst-terraform-%d/team-test", rInt),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTFETeam_import_missingOrg(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFETeamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeam_basic(rInt),
			},

			{
				ResourceName:  "tfe_team.foobar",
				ImportState:   true,
				ImportStateId: "wrongOrg/team-test",
				ExpectError:   regexp.MustCompile("no team found with name or ID .* in organization wrongOrg"),
			},
		},
	})
}

func TestAccTFETeam_import_missingTeam(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFETeamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeam_basic(rInt),
			},

			{
				ResourceName:  "tfe_team.foobar",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("tst-terraform-%d/wrongTeam", rInt),
				ExpectError:   regexp.MustCompile("no team found with name or ID wrongTeam"),
			},
		},
	})
}

func TestAccTFETeam_import_teamNameWithSpaces(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFETeamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeam_withSpaces(rInt),
			},

			{
				ResourceName:      "tfe_team.foobar",
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("tst-terraform-%d/team name with spaces", rInt),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTFETeam_import_teamNameWithSlashes(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFETeamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeam_withSlashes(rInt),
			},

			{
				ResourceName:      "tfe_team.foobar",
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("tst-terraform-%d/team/name/with/slashes", rInt),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTFETeam_import_teamNameWhichLooksLikeID(t *testing.T) {
	// Check that we can import a team with a name which looks like a team ID

	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFETeamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeam_withIDLikeName(rInt),
			},

			{
				ResourceName:      "tfe_team.foobar",
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("tst-terraform-%d/team-aaaabbbbcccc", rInt),
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckTFETeamExists(
	n string, team *tfe.Team) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(ConfiguredClient)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		t, err := config.Client.Teams.Read(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}

		if t == nil {
			return fmt.Errorf("Team not found")
		}

		*team = *t

		return nil
	}
}

func testAccCheckTFETeamAttributes_basic(
	team *tfe.Team) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if team.Name != "team-test" {
			return fmt.Errorf("Bad name: %s", team.Name)
		}
		return nil
	}
}

func testAccCheckTFETeamAttributes_full(
	team *tfe.Team) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if team.Name != "team-test" {
			return fmt.Errorf("Bad name: %s", team.Name)
		}

		if team.Visibility != "organization" {
			return fmt.Errorf("Bad visibility: %s", team.Visibility)
		}

		if !team.OrganizationAccess.ManagePolicies {
			return fmt.Errorf("OrganizationAccess.ManagePolicies should be true")
		}
		if !team.OrganizationAccess.ManageVCSSettings {
			return fmt.Errorf("OrganizationAccess.ManageVCSSettings should be true")
		}
		if !team.OrganizationAccess.ManageWorkspaces {
			return fmt.Errorf("OrganizationAccess.ManageWorkspaces should be true")
		}
		if !team.OrganizationAccess.ManageRunTasks {
			return fmt.Errorf("OrganizationAccess.ManageRunTasks should be true")
		}
		if !team.OrganizationAccess.ManageProjects {
			return fmt.Errorf("OrganizationAccess.ManageProjects should be true")
		}
		if !team.OrganizationAccess.ManageMembership {
			return fmt.Errorf("OrganizationAccess.ManageMembership should be true")
		}
		if team.SSOTeamID != "team-test-sso-id" {
			return fmt.Errorf("Bad SSO Team ID: %s", team.SSOTeamID)
		}

		return nil
	}
}

func testAccCheckTFETeamAttributes_full_update(
	team *tfe.Team) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if team.Name != "team-test-1" {
			return fmt.Errorf("Bad name: %s", team.Name)
		}

		if team.Visibility != "secret" {
			return fmt.Errorf("Bad visibility: %s", team.Visibility)
		}

		if team.OrganizationAccess.ManagePolicies {
			return fmt.Errorf("OrganizationAccess.ManagePolicies should be false")
		}
		if team.OrganizationAccess.ManageVCSSettings {
			return fmt.Errorf("OrganizationAccess.ManageVCSSettings should be false")
		}
		if team.OrganizationAccess.ManageWorkspaces {
			return fmt.Errorf("OrganizationAccess.ManageWorkspaces should be false")
		}
		if team.OrganizationAccess.ManageRunTasks {
			return fmt.Errorf("OrganizationAccess.ManageRunTasks should be false")
		}
		if team.OrganizationAccess.ManageProjects {
			return fmt.Errorf("OrganizationAccess.ManageProjects should be false")
		}
		if team.OrganizationAccess.ManageMembership {
			return fmt.Errorf("OrganizationAccess.ManageMembership should be false")
		}

		if team.SSOTeamID != "changed-sso-id" {
			return fmt.Errorf("Bad SSO Team ID: %s", team.SSOTeamID)
		}

		return nil
	}
}

func testAccCheckTFETeamDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(ConfiguredClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_team" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := config.Client.Teams.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Team %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccTFETeam_basic(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = tfe_organization.foobar.id
}`, rInt)
}

func testAccTFETeam_full(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = tfe_organization.foobar.id

  visibility = "organization"

  organization_access {
    manage_policies = true
    manage_policy_overrides = true
    manage_workspaces = true
    manage_vcs_settings = true
    manage_run_tasks = true
	manage_providers = true
	manage_modules = true
	manage_projects = true
	read_workspaces = true
	read_projects = true
	manage_membership = true
  }
  sso_team_id = "team-test-sso-id"
}`, rInt)
}

func testAccTFETeam_full_update(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_team" "foobar" {
  name         = "team-test-1"
  organization = tfe_organization.foobar.id

  visibility = "secret"

  organization_access {
    manage_policies = false
    manage_policy_overrides = false
    manage_workspaces = false
    manage_vcs_settings = false
    manage_run_tasks = false
	manage_providers = false
	manage_modules = false
	manage_projects = false
	read_projects = false
	read_workspaces = false
	manage_membership = false
  }

  sso_team_id = "changed-sso-id"
}`, rInt)
}

// unsets values to check that they are properly cleared
func testAccTFETeam_full_update_clear(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_team" "foobar" {
  name         = "team-test-1"
  organization = tfe_organization.foobar.id
}`, rInt)
}

func testAccTFETeam_withSpaces(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_team" "foobar" {
  name         = "team name with spaces"
  organization = tfe_organization.foobar.id
}`, rInt)
}

func testAccTFETeam_withSlashes(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_team" "foobar" {
  name         = "team/name/with/slashes"
  organization = tfe_organization.foobar.id
}`, rInt)
}

func testAccTFETeam_withIDLikeName(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_team" "foobar" {
  name         = "team-aaaabbbbcccc"
  organization = tfe_organization.foobar.id
}`, rInt)
}
