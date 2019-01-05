package tfe

import (
	"fmt"
	"testing"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccTFETeamMembers_basic(t *testing.T) {
	users := []*tfe.User{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFETeamMembersDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccTFETeamMembers_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETeamMembersExists(
						"tfe_team_members.foobar", &users),
					testAccCheckTFETeamMembersAttributes(&users),
					resource.TestCheckResourceAttr(
						"tfe_team_members.foobar", "usernames.#", "2"),
					resource.TestCheckResourceAttr(
						"tfe_team_members.foobar", "usernames.3672628397", "admin"),
					resource.TestCheckResourceAttr(
						"tfe_team_members.foobar", "usernames.4078738388", "sander"),
				),
			},
		},
	})
}

func TestAccTFETeamMembers_update(t *testing.T) {
	users := []*tfe.User{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFETeamMembersDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccTFETeamMembers_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETeamMembersExists(
						"tfe_team_members.foobar", &users),
					testAccCheckTFETeamMembersAttributes(&users),
					resource.TestCheckResourceAttr(
						"tfe_team_members.foobar", "usernames.#", "2"),
					resource.TestCheckResourceAttr(
						"tfe_team_members.foobar", "usernames.3672628397", "admin"),
					resource.TestCheckResourceAttr(
						"tfe_team_members.foobar", "usernames.4078738388", "sander"),
				),
			},

			resource.TestStep{
				Config: testAccTFETeamMembers_update,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETeamMembersExists(
						"tfe_team_members.foobar", &users),
					testAccCheckTFETeamMembersAttributesUpdate(&users),
					resource.TestCheckResourceAttr(
						"tfe_team_members.foobar", "usernames.#", "2"),
					resource.TestCheckResourceAttr(
						"tfe_team_members.foobar", "usernames.1348969918", "ryan"),
					resource.TestCheckResourceAttr(
						"tfe_team_members.foobar", "usernames.3672628397", "admin"),
				),
			},
		},
	})
}

func TestAccTFETeamMembers_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFETeamMembersDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccTFETeamMembers_basic,
			},

			resource.TestStep{
				ResourceName:      "tfe_team_members.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckTFETeamMembersExists(
	n string, users *[]*tfe.User) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		tfeClient := testAccProvider.Meta().(*tfe.Client)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		us, err := tfeClient.TeamMembers.List(ctx, rs.Primary.ID)
		if err != nil && err != tfe.ErrResourceNotFound {
			return err
		}

		if len(us) != 2 {
			return fmt.Errorf("Users not found: %#+v", us[0])
		}

		*users = us

		return nil
	}
}

func testAccCheckTFETeamMembersAttributes(
	users *[]*tfe.User) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		found := 0
		for _, user := range *users {
			switch user.Username {
			case "admin", "sander":
				found++
			}
		}

		if found != 2 {
			return fmt.Errorf("Bad users: %#+v", *users)
		}

		return nil
	}
}

func testAccCheckTFETeamMembersAttributesUpdate(
	users *[]*tfe.User) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		found := 0
		for _, user := range *users {
			switch user.Username {
			case "admin", "ryan":
				found++
			}
		}

		if found != 2 {
			return fmt.Errorf("Bad users: %#+v", *users)
		}

		return nil
	}
}

func testAccCheckTFETeamMembersDestroy(s *terraform.State) error {
	tfeClient := testAccProvider.Meta().(*tfe.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_team_members" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		users, err := tfeClient.TeamMembers.List(ctx, rs.Primary.ID)
		if err != nil && err != tfe.ErrResourceNotFound {
			return err
		}

		if len(users) != 0 {
			return fmt.Errorf("Users still exist")
		}
	}

	return nil
}

const testAccTFETeamMembers_basic = `
resource "tfe_organization" "foobar" {
  name  = "terraform-test"
  email = "admin@company.com"
}

resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = "${tfe_organization.foobar.id}"
}

resource "tfe_team_members" "foobar" {
  team_id   = "${tfe_team.foobar.id}"
  usernames = ["admin", "sander"]
}`

const testAccTFETeamMembers_update = `
resource "tfe_organization" "foobar" {
  name  = "terraform-test"
  email = "admin@company.com"
}

resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = "${tfe_organization.foobar.id}"
}

resource "tfe_team_members" "foobar" {
  team_id   = "${tfe_team.foobar.id}"
  usernames = ["admin", "ryan"]
}`
