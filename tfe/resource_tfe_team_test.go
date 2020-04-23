package tfe

import (
	"fmt"
	"testing"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccTFETeam_basic(t *testing.T) {
	team := &tfe.Team{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFETeamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeam_basic,
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

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFETeamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeam_full,
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
						"tfe_team.foobar", "organization_access.0.manage_workspaces", "true"),
					resource.TestCheckResourceAttr(
						"tfe_team.foobar", "organization_access.0.manage_vcs_settings", "true"),
				),
			},
		},
	})
}

func TestAccTFETeam_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFETeamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeam_basic,
			},

			{
				ResourceName:        "tfe_team.foobar",
				ImportState:         true,
				ImportStateIdPrefix: "tst-terraform/",
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

const testAccTFETeam_basic = `
resource "tfe_organization" "foobar" {
  name  = "tst-terraform"
  email = "admin@company.com"
}

resource "tfe_team" "foobar" {
  name         = "team-test"
	organization = "${tfe_organization.foobar.id}"
}`

const testAccTFETeam_full = `
resource "tfe_organization" "foobar" {
  name  = "tst-terraform"
  email = "admin@company.com"
}

resource "tfe_team" "foobar" {
  name         = "team-test"
	organization = "${tfe_organization.foobar.id}"

	visibility = "organization"
	
	organization_access {
		manage_policies = true
		manage_workspaces = true
		manage_vcs_settings = true
	}
}`
