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

func TestAccTFETeamProjectAccess_admin(t *testing.T) {
	skipUnlessBeta(t)

	tmAccess := &tfe.TeamProjectAccess{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	for _, access := range []tfe.TeamProjectAccessType{tfe.TeamProjectAccessAdmin, tfe.TeamProjectAccessRead} {
		resource.Test(t, resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckTFETeamProjectAccessDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccTFETeamProjectAccess(rInt, access),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckTFETeamProjectAccessExists(
							"tfe_team_project_access.foobar", tmAccess),
						testAccCheckTFETeamProjectAccessAttributesAccessIs(tmAccess, access),
						resource.TestCheckResourceAttr("tfe_team_project_access.foobar", "access", string(access)),
					),
				},
			},
		})
	}
}

func TestAccTFETeamProjectAccess_import(t *testing.T) {
	skipUnlessBeta(t)

	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFETeamProjectAccessDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamProjectAccess(rInt, tfe.TeamProjectAccessAdmin),
			},
			{
				ResourceName:      "tfe_team_project_access.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckTFETeamProjectAccessExists(
	n string, tmAccess *tfe.TeamProjectAccess) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(ConfiguredClient)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no instance ID is set")
		}

		ta, err := config.Client.TeamProjectAccess.Read(ctx, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error reading team project access %s: %w", rs.Primary.ID, err)
		}

		if ta == nil {
			return fmt.Errorf("TeamAccess not found")
		}

		*tmAccess = *ta

		return nil
	}
}

func testAccCheckTFETeamProjectAccessAttributesAccessIs(tmAccess *tfe.TeamProjectAccess, access tfe.TeamProjectAccessType) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if tmAccess.Access != access {
			return fmt.Errorf("Bad access: %s", tmAccess.Access)
		}
		return nil
	}
}

func testAccCheckTFETeamProjectAccessDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(ConfiguredClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_team_project_access" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := config.Client.TeamProjectAccess.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Team project access %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccTFETeamProjectAccess(rInt int, access tfe.TeamProjectAccessType) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_project" "foobar" {
  name         = "projecttest"
  organization = tfe_organization.foobar.id
}

resource "tfe_team_project_access" "foobar" {
  access       = "%s"
  team_id      = tfe_team.foobar.id
  project_id   = tfe_project.foobar.id
}`, rInt, access)
}
