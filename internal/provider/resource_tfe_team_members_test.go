// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccTFETeamMembers_basic(t *testing.T) {
	t.Skip("Skipping, due to current testing limitations; namely, an organization membership must first be confirmed.")
	users := []*tfe.User{}
	tfeUser1Hash := hashSchemaString(envTFEUser1)
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			if envTFEUser1 == "" {
				t.Skip("Please set TFE_USER1 to run this test")
			}
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFETeamMembersDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamMembers_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETeamMembersExists(
						"tfe_team_members.foobar", &users),
					testAccCheckTFETeamMembersAttributes(&users, []string{"admin", envTFEUser1}),
					resource.TestCheckResourceAttr(
						"tfe_team_members.foobar", "usernames.#", "2"),
					resource.TestCheckResourceAttr(
						"tfe_team_members.foobar", "usernames.3672628397", "admin"),
					resource.TestCheckResourceAttr(
						"tfe_team_members.foobar", fmt.Sprintf("usernames.%d", tfeUser1Hash), envTFEUser1),
				),
			},
		},
	})
}

func TestAccTFETeamMembers_update(t *testing.T) {
	t.Skip("Skipping, due to current testing limitations; namely, an organization membership must first be confirmed.")
	users := []*tfe.User{}
	tfeUser1Hash := hashSchemaString(envTFEUser1)
	tfeUser2Hash := hashSchemaString(envTFEUser2)
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			if envTFEUser1 == "" {
				t.Skip("Please set TFE_USER1 to run this test")
			}
			if envTFEUser2 == "" {
				t.Skip("Please set TFE_USER2 to run this test")
			}
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFETeamMembersDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamMembers_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETeamMembersExists(
						"tfe_team_members.foobar", &users),
					testAccCheckTFETeamMembersAttributes(&users, []string{"admin", envTFEUser1}),
					resource.TestCheckResourceAttr(
						"tfe_team_members.foobar", "usernames.#", "2"),
					resource.TestCheckResourceAttr(
						"tfe_team_members.foobar", "usernames.3672628397", "admin"),
					resource.TestCheckResourceAttr(
						"tfe_team_members.foobar", fmt.Sprintf("usernames.%d", tfeUser1Hash), envTFEUser1),
				),
			},

			{
				Config: testAccTFETeamMembers_update(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETeamMembersExists(
						"tfe_team_members.foobar", &users),
					testAccCheckTFETeamMembersAttributes(&users, []string{"admin", envTFEUser2}),
					resource.TestCheckResourceAttr(
						"tfe_team_members.foobar", "usernames.#", "2"),
					resource.TestCheckResourceAttr(
						"tfe_team_members.foobar", fmt.Sprintf("usernames.%d", tfeUser2Hash), envTFEUser2),
					resource.TestCheckResourceAttr(
						"tfe_team_members.foobar", "usernames.3672628397", "admin"),
				),
			},
		},
	})
}

func TestAccTFETeamMembers_import(t *testing.T) {
	t.Skip("Skipping, due to current testing limitations; namely, an organization membership must first be confirmed.")
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			if envTFEUser1 == "" {
				t.Skip("Please set TFE_USER1 to run this test")
			}
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFETeamMembersDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamMembers_basic(rInt),
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
		config := testAccProvider.Meta().(ConfiguredClient)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		us, err := config.Client.TeamMembers.List(ctx, rs.Primary.ID)
		if err != nil && !errors.Is(err, tfe.ErrResourceNotFound) {
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
	config := testAccProvider.Meta().(ConfiguredClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_team_members" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		users, err := config.Client.TeamMembers.List(ctx, rs.Primary.ID)
		if err != nil && !errors.Is(err, tfe.ErrResourceNotFound) {
			return err
		}

		if len(users) != 0 {
			return fmt.Errorf("Users still exist")
		}
	}

	return nil
}

func testAccTFETeamMembers_basic(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_team_members" "foobar" {
  team_id   = tfe_team.foobar.id
  usernames = ["%s"]
}`, rInt, envTFEUser1)
}

func testAccTFETeamMembers_update(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_team_members" "foobar" {
  team_id   = tfe_team.foobar.id
  usernames = ["%s", "%s"]
}`, rInt, envTFEUser1, envTFEUser2)
}
