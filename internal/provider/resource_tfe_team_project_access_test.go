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

func TestAccTFETeamProjectAccess(t *testing.T) {
	tmAccess := &tfe.TeamProjectAccess{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	for _, access := range []tfe.TeamProjectAccessType{tfe.TeamProjectAccessAdmin, tfe.TeamProjectAccessMaintain, tfe.TeamProjectAccessWrite, tfe.TeamProjectAccessRead} {
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

func TestAccTFETeamProjectCustomAccess(t *testing.T) {
	tmAccess := &tfe.TeamProjectAccess{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	access := tfe.TeamProjectAccessCustom

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFETeamProjectAccessDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamProjectCustomAccess(rInt, access),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETeamProjectAccessExists(
						"tfe_team_project_access.custom_foobar", tmAccess),
					testAccCheckTFETeamProjectAccessAttributesAccessIs(tmAccess, access),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "access", string(access)),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "project_access.0.settings", "delete"),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "project_access.0.teams", "manage"),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "workspace_access.0.state_versions", "write"),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "workspace_access.0.sentinel_mocks", "read"),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "workspace_access.0.runs", "read"),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "workspace_access.0.variables", "write"),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "workspace_access.0.create", "true"),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "workspace_access.0.locking", "true"),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "workspace_access.0.move", "true"),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "workspace_access.0.delete", "false"),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "workspace_access.0.run_tasks", "false"),
				),
			},
		},
	})
}
func TestAccTFETeamProjectAccess_import(t *testing.T) {
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

func TestAccTFETeamProjectCustomAccess_import(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	tmAccess := &tfe.TeamProjectAccess{}
	access := tfe.TeamProjectAccessCustom

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFETeamProjectAccessDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamProjectCustomAccess(rInt, access),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETeamProjectAccessExists(
						"tfe_team_project_access.custom_foobar", tmAccess),
					testAccCheckTFETeamProjectAccessAttributesAccessIs(tmAccess, access),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "access", string(access)),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "project_access.0.settings", "delete"),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "project_access.0.teams", "manage"),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "workspace_access.0.state_versions", "write"),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "workspace_access.0.sentinel_mocks", "read"),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "workspace_access.0.runs", "read"),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "workspace_access.0.variables", "write"),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "workspace_access.0.create", "true"),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "workspace_access.0.locking", "true"),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "workspace_access.0.move", "true"),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "workspace_access.0.delete", "false"),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "workspace_access.0.run_tasks", "false"),
				),
			},
			{
				ResourceName:      "tfe_team_project_access.custom_foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTFETeamProjectCustomAccess_full_update(t *testing.T) {
	tmAccess := &tfe.TeamProjectAccess{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	access := tfe.TeamProjectAccessCustom

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamProjectCustomAccess(rInt, access),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETeamProjectAccessExists(
						"tfe_team_project_access.custom_foobar", tmAccess),
					testAccCheckTFETeamProjectAccessAttributesAccessIs(tmAccess, access),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "access", string(access)),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "project_access.0.settings", "delete"),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "project_access.0.teams", "manage"),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "workspace_access.0.state_versions", "write"),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "workspace_access.0.sentinel_mocks", "read"),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "workspace_access.0.runs", "read"),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "workspace_access.0.variables", "write"),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "workspace_access.0.create", "true"),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "workspace_access.0.locking", "true"),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "workspace_access.0.move", "true"),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "workspace_access.0.delete", "false"),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "workspace_access.0.run_tasks", "false"),
				),
			},
			{
				Config: testAccTFETeamProjectCustomAccess_full_update(rInt, access),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETeamProjectAccessExists(
						"tfe_team_project_access.custom_foobar", tmAccess),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "access", string(access)),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "project_access.0.settings", "read"),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "project_access.0.teams", "none"),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "workspace_access.0.state_versions", "read"),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "workspace_access.0.sentinel_mocks", "none"),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "workspace_access.0.runs", "apply"),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "workspace_access.0.variables", "read"),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "workspace_access.0.create", "false"),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "workspace_access.0.locking", "false"),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "workspace_access.0.move", "false"),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "workspace_access.0.delete", "true"),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "workspace_access.0.run_tasks", "true"),
				),
			},
		},
	})
}

func TestAccTFETeamProjectCustomAccess_partial_update(t *testing.T) {
	tmAccess := &tfe.TeamProjectAccess{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	access := tfe.TeamProjectAccessCustom

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamProjectCustomAccess(rInt, access),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETeamProjectAccessExists(
						"tfe_team_project_access.custom_foobar", tmAccess),
					testAccCheckTFETeamProjectAccessAttributesAccessIs(tmAccess, access),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "access", string(access)),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "project_access.0.settings", "delete"),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "project_access.0.teams", "manage"),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "workspace_access.0.state_versions", "write"),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "workspace_access.0.sentinel_mocks", "read"),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "workspace_access.0.variables", "write"),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "workspace_access.0.create", "true"),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "workspace_access.0.locking", "true"),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "workspace_access.0.move", "true"),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "workspace_access.0.delete", "false"),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "workspace_access.0.run_tasks", "false"),
				),
			},
			{
				Config: testAccTFETeamProjectCustomAccess_partial_update(rInt, access),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETeamProjectAccessExists(
						"tfe_team_project_access.custom_foobar", tmAccess),
					testAccCheckTFETeamProjectAccessAttributesAccessIs(tmAccess, access),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "access", string(access)),
					// changed access levels
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "project_access.0.settings", "read"),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "workspace_access.0.delete", "true"),
					// unchanged access levels
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "project_access.0.teams", "manage"),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "workspace_access.0.state_versions", "write"),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "workspace_access.0.sentinel_mocks", "read"),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "workspace_access.0.variables", "write"),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "workspace_access.0.create", "true"),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "workspace_access.0.locking", "true"),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "workspace_access.0.move", "true"),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "workspace_access.0.delete", "true"),
					resource.TestCheckResourceAttr("tfe_team_project_access.custom_foobar", "workspace_access.0.run_tasks", "false"),
				),
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

func TestAccTFETeamProjectCustomAccess_invalid_custom_access(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFETeamProjectAccessDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFETeamProjectCustomAccess_invalid_custom_config(rInt),
				ExpectError: regexp.MustCompile("you can only set workspace_access permissions with access level custom"),
			},
		},
	})
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

func testAccTFETeamProjectCustomAccess(rInt int, access tfe.TeamProjectAccessType) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar_2" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_team" "foobar_2" {
  name         = "team-test"
  organization = tfe_organization.foobar_2.id
}

resource "tfe_project" "foobar_2" {
  name         = "projecttest"
  organization = tfe_organization.foobar_2.id
}

resource "tfe_team_project_access" "custom_foobar" {
  access       = "%s"
  team_id      = tfe_team.foobar_2.id
  project_id   = tfe_project.foobar_2.id
  project_access {
    settings = "delete"
    teams    = "manage"
  }
  workspace_access {
    state_versions = "write"
    sentinel_mocks = "read"
    runs           = "read"
    variables      = "write"
    create         = true
    locking        = true
    move           = true
    delete         = false
    run_tasks      = false
  }

}`, rInt, access)
}

func testAccTFETeamProjectCustomAccess_full_update(rInt int, access tfe.TeamProjectAccessType) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar_2" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_team" "foobar_2" {
  name         = "team-test"
  organization = tfe_organization.foobar_2.id
}

resource "tfe_project" "foobar_2" {
  name         = "projecttest"
  organization = tfe_organization.foobar_2.id
}

resource "tfe_team_project_access" "custom_foobar" {
  access       = "%s"
  team_id      = tfe_team.foobar_2.id
  project_id   = tfe_project.foobar_2.id
  project_access {
    settings = "read"
    teams    = "none"
  }
  workspace_access {
    state_versions = "read"
    sentinel_mocks = "none"
    runs           = "apply"
    variables      = "read"
    create         = false
    locking        = false
    move           = false
    delete         = true
    run_tasks      = true
  }
}`, rInt, access)
}

func testAccTFETeamProjectCustomAccess_partial_update(rInt int, access tfe.TeamProjectAccessType) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar_2" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_team" "foobar_2" {
  name         = "team-test"
  organization = tfe_organization.foobar_2.id
}

resource "tfe_project" "foobar_2" {
  name         = "projecttest"
  organization = tfe_organization.foobar_2.id
}

resource "tfe_team_project_access" "custom_foobar" {
  access       = "%s"
  team_id      = tfe_team.foobar_2.id
  project_id   = tfe_project.foobar_2.id
  project_access {
    settings = "read"
  }
  workspace_access {
    delete = true
  }
}`, rInt, access)
}

func testAccTFETeamProjectCustomAccess_invalid_custom_config(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar_2" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_team" "foobar_invalid" {
  name         = "team-test"
  organization = tfe_organization.foobar_2.id
}

resource "tfe_project" "foobar_invalid" {
  name         = "projecttest"
  organization = tfe_organization.foobar_2.id
}

resource "tfe_team_project_access" "custom_invalid" {
  access       = "read"
  team_id      = tfe_team.foobar_invalid.id
  project_id   = tfe_project.foobar_invalid.id

  workspace_access {
    delete = true
  }
}`, rInt)
}
