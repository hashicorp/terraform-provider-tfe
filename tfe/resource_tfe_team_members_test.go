package tfe

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccTFETeamMembers_basic(t *testing.T) {
	users := []*tfe.User{}
	username := os.Getenv("TFE_USER1")
	usernameHash := hashSchemaString(username)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCheckEnvVariables(t, []string{"TFE_USER1"})
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFETeamMembersDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamMembers_basic([]string{"admin", username}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETeamMembersExists(
						"tfe_team_members.foobar", &users),
					testAccCheckTFETeamMembersAttributes(&users, []string{"admin", username}),
					resource.TestCheckResourceAttr(
						"tfe_team_members.foobar", "usernames.#", "2"),
					resource.TestCheckResourceAttr(
						"tfe_team_members.foobar", "usernames.3672628397", "admin"),
					resource.TestCheckResourceAttr(
						"tfe_team_members.foobar", fmt.Sprintf("usernames.%d", usernameHash), username),
				),
			},
		},
	})
}

func TestAccTFETeamMembers_update(t *testing.T) {
	users := []*tfe.User{}
	username := os.Getenv("TFE_USER1")
	usernameHash := hashSchemaString(username)
	secondUsername := os.Getenv("TFE_USER2")
	secondUsernameHash := hashSchemaString(secondUsername)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCheckEnvVariables(t, []string{"TFE_USER1", "TFE_USER2"})
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFETeamMembersDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamMembers_basic([]string{"admin", username}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETeamMembersExists(
						"tfe_team_members.foobar", &users),
					testAccCheckTFETeamMembersAttributes(&users, []string{"admin", username}),
					resource.TestCheckResourceAttr(
						"tfe_team_members.foobar", "usernames.#", "2"),
					resource.TestCheckResourceAttr(
						"tfe_team_members.foobar", "usernames.3672628397", "admin"),
					resource.TestCheckResourceAttr(
						"tfe_team_members.foobar", fmt.Sprintf("usernames.%d", usernameHash), username),
				),
			},

			{
				Config: testAccTFETeamMembers_basic([]string{"admin", secondUsername}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETeamMembersExists(
						"tfe_team_members.foobar", &users),
					testAccCheckTFETeamMembersAttributes(&users, []string{"admin", secondUsername}),
					resource.TestCheckResourceAttr(
						"tfe_team_members.foobar", "usernames.#", "2"),
					resource.TestCheckResourceAttr(
						"tfe_team_members.foobar", fmt.Sprintf("usernames.%d", secondUsernameHash), secondUsername),
					resource.TestCheckResourceAttr(
						"tfe_team_members.foobar", "usernames.3672628397", "admin"),
				),
			},
		},
	})
}

func TestAccTFETeamMembers_import(t *testing.T) {
	username := os.Getenv("TFE_USER1")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCheckEnvVariables(t, []string{"TFE_USER1"})
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFETeamMembersDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamMembers_basic([]string{"admin", username}),
			},

			{
				ResourceName:      "tfe_team_members.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func hashSchemaString(username string) int {
	return schema.HashSchema(&schema.Schema{Type: schema.TypeString})(username)

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
	users *[]*tfe.User, expectedUsernames []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		usernames := usernamesFromTFEUsers(*users)
		if !reflect.DeepEqual(usernames, expectedUsernames) {
			return fmt.Errorf("Expected usernames: %q, Given: %q",
				expectedUsernames, usernames)
		}

		return nil
	}
}

func usernamesFromTFEUsers(users []*tfe.User) []string {
	usernames := make([]string, len(users), len(users))
	for i, user := range users {
		usernames[i] = user.Username
	}
	return usernames
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

func testAccTFETeamMembers_basic(usernames []string) string {
	return fmt.Sprintf(`
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
  usernames = ["%s"]
}`, strings.Join(usernames, `", "`))
}
