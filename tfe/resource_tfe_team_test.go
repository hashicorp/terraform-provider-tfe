package tfe

import (
	"fmt"
	"math/rand"
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
						"tfe_team.foobar", "organization_access.0.manage_vcs_settings", "false"),
					resource.TestCheckResourceAttr(
						"tfe_team.foobar", "organization_access.0.manage_providers", "false"),
					resource.TestCheckResourceAttr(
						"tfe_team.foobar", "organization_access.0.manage_modules", "false"),
					resource.TestCheckResourceAttr(
						"tfe_team.foobar", "organization_access.0.manage_run_tasks", "false"),
					resource.TestCheckResourceAttr(
						"tfe_team.foobar", "sso_team_id", "changed-sso-id"),
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
						"tfe_team.foobar", "sso_team_id", ""),
				),
			},
		},
	})
}

func TestAccTFETeam_import(t *testing.T) {
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

func testAccCheckTFETeamExists(
	n string, team *tfe.Team) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		tfeClient := testAccProvider.Meta().(*tfe.Client)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		t, err := tfeClient.Teams.Read(ctx, rs.Primary.ID)
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

		if team.SSOTeamID != "changed-sso-id" {
			return fmt.Errorf("Bad SSO Team ID: %s", team.SSOTeamID)
		}

		return nil
	}
}

func testAccCheckTFETeamDestroy(s *terraform.State) error {
	tfeClient := testAccProvider.Meta().(*tfe.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_team" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := tfeClient.Teams.Read(ctx, rs.Primary.ID)
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
