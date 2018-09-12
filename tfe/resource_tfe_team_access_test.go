package tfe

import (
	"fmt"
	"testing"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccTFETeamAccess_basic(t *testing.T) {
	tmAccess := &tfe.TeamAccess{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFETeamAccessDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccTFETeamAccess_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETeamAccessExists(
						"tfe_team_access.foobar", tmAccess),
					testAccCheckTFETeamAccessAttributes(tmAccess),
					resource.TestCheckResourceAttr(
						"tfe_team_access.foobar", "access", "write"),
				),
			},
		},
	})
}

func TestAccTFETeamAccess_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFETeamAccessDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccTFETeamAccess_basic,
			},

			resource.TestStep{
				ResourceName:        "tfe_team_access.foobar",
				ImportState:         true,
				ImportStateIdPrefix: "terraform-test/workspace-test/",
				ImportStateVerify:   true,
			},
		},
	})
}

func testAccCheckTFETeamAccessExists(
	n string, tmAccess *tfe.TeamAccess) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		tfeClient := testAccProvider.Meta().(*tfe.Client)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		ta, err := tfeClient.TeamAccess.Read(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}

		if ta == nil {
			return fmt.Errorf("TeamAccess not found")
		}

		*tmAccess = *ta

		return nil
	}
}

func testAccCheckTFETeamAccessAttributes(
	tmAccess *tfe.TeamAccess) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if tmAccess.Access != tfe.AccessWrite {
			return fmt.Errorf("Bad access: %s", tmAccess.Access)
		}
		return nil
	}
}

func testAccCheckTFETeamAccessDestroy(s *terraform.State) error {
	tfeClient := testAccProvider.Meta().(*tfe.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_team_access" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := tfeClient.TeamAccess.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Team access %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

const testAccTFETeamAccess_basic = `
resource "tfe_organization" "foobar" {
  name = "terraform-test"
  email = "admin@company.com"
}

resource "tfe_team" "foobar" {
  name = "team-test"
  organization = "${tfe_organization.foobar.id}"
}

resource "tfe_workspace" "foobar" {
  name = "workspace-test"
  organization = "${tfe_organization.foobar.id}"
}

resource "tfe_team_access" "foobar" {
  access = "write"
  team_id = "${tfe_team.foobar.id}"
  workspace_id = "${tfe_workspace.foobar.id}"
}`
