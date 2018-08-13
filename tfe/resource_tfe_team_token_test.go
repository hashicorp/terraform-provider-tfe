package tfe

import (
	"fmt"
	"regexp"
	"testing"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccTFETeamToken_basic(t *testing.T) {
	token := &tfe.TeamToken{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFETeamTokenDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccTFETeamToken_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETeamTokenExists(
						"tfe_team_token.foobar", token),
				),
			},
		},
	})
}

func TestAccTFETeamToken_existsWithoutForce(t *testing.T) {
	token := &tfe.TeamToken{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFETeamTokenDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccTFETeamToken_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETeamTokenExists(
						"tfe_team_token.foobar", token),
				),
			},

			resource.TestStep{
				Config:      testAccTFETeamToken_existsWithoutForce,
				ExpectError: regexp.MustCompile(`token already exists`),
			},
		},
	})
}

func TestAccTFETeamToken_existsWithForce(t *testing.T) {
	token := &tfe.TeamToken{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFETeamTokenDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccTFETeamToken_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETeamTokenExists(
						"tfe_team_token.foobar", token),
				),
			},

			resource.TestStep{
				Config: testAccTFETeamToken_existsWithForce,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETeamTokenExists(
						"tfe_team_token.regenerated", token),
				),
			},
		},
	})
}

func testAccCheckTFETeamTokenExists(
	n string, token *tfe.TeamToken) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		tfeClient := testAccProvider.Meta().(*tfe.Client)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		tt, err := tfeClient.TeamTokens.Read(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}

		if tt == nil {
			return fmt.Errorf("Team token not found")
		}

		*token = *tt

		return nil
	}
}

func testAccCheckTFETeamTokenDestroy(s *terraform.State) error {
	tfeClient := testAccProvider.Meta().(*tfe.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_team_token" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := tfeClient.TeamTokens.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Team token %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

const testAccTFETeamToken_basic = `
resource "tfe_organization" "foobar" {
  name = "terraform-test"
  email = "admin@company.com"
}

resource "tfe_team" "foobar" {
  name = "team-test"
  organization = "${tfe_organization.foobar.id}"
}

resource "tfe_team_token" "foobar" {
  team_id = "${tfe_team.foobar.id}"
}`

const testAccTFETeamToken_existsWithoutForce = `
resource "tfe_organization" "foobar" {
  name = "terraform-test"
  email = "admin@company.com"
}

resource "tfe_team" "foobar" {
  name = "team-test"
  organization = "${tfe_organization.foobar.id}"
}

resource "tfe_team_token" "foobar" {
  team_id = "${tfe_team.foobar.id}"
}

resource "tfe_team_token" "error" {
  team_id = "${tfe_team.foobar.id}"
}`

const testAccTFETeamToken_existsWithForce = `
resource "tfe_organization" "foobar" {
  name = "terraform-test"
  email = "admin@company.com"
}

resource "tfe_team" "foobar" {
  name = "team-test"
  organization = "${tfe_organization.foobar.id}"
}

resource "tfe_team_token" "foobar" {
  team_id = "${tfe_team.foobar.id}"
}

resource "tfe_team_token" "regenerated" {
  team_id = "${tfe_team.foobar.id}"
  force_regenerate = true
}`
