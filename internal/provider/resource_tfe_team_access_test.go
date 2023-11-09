// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccTFETeamAccess_admin(t *testing.T) {
	tmAccess := &tfe.TeamAccess{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	expectedPermissions := map[string]interface{}{
		"runs":              tfe.RunsPermissionApply,
		"variables":         tfe.VariablesPermissionWrite,
		"state_versions":    tfe.StateVersionsPermissionWrite,
		"sentinel_mocks":    tfe.SentinelMocksPermissionRead,
		"workspace_locking": true,
		"run_tasks":         true,
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFETeamAccessDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamAccess_admin(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETeamAccessExists(
						"tfe_team_access.foobar", tmAccess),
					testAccCheckTFETeamAccessAttributesAccessIs(tmAccess, tfe.AccessAdmin),
					testAccCheckTFETeamAccessAttributesPermissionsAre(tmAccess, expectedPermissions),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "access", "admin"),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "permissions.0.runs", "apply"),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "permissions.0.variables", "write"),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "permissions.0.state_versions", "write"),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "permissions.0.sentinel_mocks", "read"),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "permissions.0.workspace_locking", "true"),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "permissions.0.run_tasks", "true"),
				),
			},
		},
	})
}

func TestAccTFETeamAccess_write(t *testing.T) {
	tmAccess := &tfe.TeamAccess{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	expectedPermissions := map[string]interface{}{
		"runs":              tfe.RunsPermissionApply,
		"variables":         tfe.VariablesPermissionWrite,
		"state_versions":    tfe.StateVersionsPermissionWrite,
		"sentinel_mocks":    tfe.SentinelMocksPermissionRead,
		"workspace_locking": true,
		"run_tasks":         false,
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFETeamAccessDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamAccess_write(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETeamAccessExists(
						"tfe_team_access.foobar", tmAccess),
					testAccCheckTFETeamAccessAttributesAccessIs(tmAccess, tfe.AccessWrite),
					testAccCheckTFETeamAccessAttributesPermissionsAre(tmAccess, expectedPermissions),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "access", "write"),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "permissions.0.runs", "apply"),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "permissions.0.variables", "write"),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "permissions.0.state_versions", "write"),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "permissions.0.sentinel_mocks", "read"),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "permissions.0.workspace_locking", "true"),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "permissions.0.run_tasks", "false"),
				),
			},
		},
	})
}

func TestAccTFETeamAccess_custom(t *testing.T) {
	tmAccess := &tfe.TeamAccess{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFETeamAccessDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamAccess_custom(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETeamAccessExists(
						"tfe_team_access.foobar", tmAccess),
					testAccCheckTFETeamAccessAttributesAccessIs(tmAccess, tfe.AccessCustom),
					testAccCheckTFETeamAccessAttributesPermissionsAre(
						tmAccess,
						map[string]interface{}{
							"runs":              tfe.RunsPermissionApply,
							"variables":         tfe.VariablesPermissionRead,
							"state_versions":    tfe.StateVersionsPermissionReadOutputs,
							"sentinel_mocks":    tfe.SentinelMocksPermissionNone,
							"workspace_locking": false,
							"run_tasks":         false,
						},
					),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "access", "custom"),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "permissions.0.runs", "apply"),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "permissions.0.variables", "read"),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "permissions.0.state_versions", "read-outputs"),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "permissions.0.sentinel_mocks", "none"),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "permissions.0.workspace_locking", "false"),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "permissions.0.run_tasks", "false"),
				),
			},
		},
	})
}

func TestAccTFETeamAccess_updateToCustom(t *testing.T) {
	tmAccess := &tfe.TeamAccess{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFETeamAccessDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamAccess_write(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETeamAccessExists(
						"tfe_team_access.foobar", tmAccess),
					testAccCheckTFETeamAccessAttributesAccessIs(tmAccess, tfe.AccessWrite),
					testAccCheckTFETeamAccessAttributesPermissionsAre(
						tmAccess,
						map[string]interface{}{
							"runs":              tfe.RunsPermissionApply,
							"variables":         tfe.VariablesPermissionWrite,
							"state_versions":    tfe.StateVersionsPermissionWrite,
							"sentinel_mocks":    tfe.SentinelMocksPermissionRead,
							"workspace_locking": true,
							"run_tasks":         false,
						},
					),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "access", "write"),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "permissions.0.runs", "apply"),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "permissions.0.variables", "write"),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "permissions.0.state_versions", "write"),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "permissions.0.sentinel_mocks", "read"),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "permissions.0.workspace_locking", "true"),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "permissions.0.run_tasks", "false"),
				),
			},
			{
				Config: testAccTFETeamAccess_custom(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETeamAccessExists(
						"tfe_team_access.foobar", tmAccess),
					testAccCheckTFETeamAccessAttributesAccessIs(tmAccess, tfe.AccessCustom),
					testAccCheckTFETeamAccessAttributesPermissionsAre(
						tmAccess,
						map[string]interface{}{
							"runs":              tfe.RunsPermissionApply,
							"variables":         tfe.VariablesPermissionRead,
							"state_versions":    tfe.StateVersionsPermissionReadOutputs,
							"sentinel_mocks":    tfe.SentinelMocksPermissionNone,
							"workspace_locking": false,
							"run_tasks":         false,
						},
					),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "access", "custom"),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "permissions.0.runs", "apply"),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "permissions.0.variables", "read"),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "permissions.0.state_versions", "read-outputs"),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "permissions.0.sentinel_mocks", "none"),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "permissions.0.workspace_locking", "false"),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "permissions.0.run_tasks", "false"),
				),
			},
		},
	})
}

func TestAccTFETeamAccess_updateFromCustom(t *testing.T) {
	tmAccess := &tfe.TeamAccess{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFETeamAccessDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamAccess_custom(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETeamAccessExists(
						"tfe_team_access.foobar", tmAccess),
					testAccCheckTFETeamAccessAttributesAccessIs(tmAccess, tfe.AccessCustom),
					testAccCheckTFETeamAccessAttributesPermissionsAre(
						tmAccess,
						map[string]interface{}{
							"runs":              tfe.RunsPermissionApply,
							"variables":         tfe.VariablesPermissionRead,
							"state_versions":    tfe.StateVersionsPermissionReadOutputs,
							"sentinel_mocks":    tfe.SentinelMocksPermissionNone,
							"workspace_locking": false,
							"run_tasks":         false,
						},
					),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "access", "custom"),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "permissions.0.runs", "apply"),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "permissions.0.variables", "read"),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "permissions.0.state_versions", "read-outputs"),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "permissions.0.sentinel_mocks", "none"),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "permissions.0.workspace_locking", "false"),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "permissions.0.run_tasks", "false"),
				),
			},
			{
				Config: testAccTFETeamAccess_plan(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETeamAccessExists(
						"tfe_team_access.foobar", tmAccess),
					testAccCheckTFETeamAccessAttributesAccessIs(tmAccess, tfe.AccessPlan),
					testAccCheckTFETeamAccessAttributesPermissionsAre(
						tmAccess,
						map[string]interface{}{
							"runs":              tfe.RunsPermissionPlan,
							"variables":         tfe.VariablesPermissionRead,
							"state_versions":    tfe.StateVersionsPermissionRead,
							"sentinel_mocks":    tfe.SentinelMocksPermissionNone,
							"workspace_locking": false,
							"run_tasks":         false,
						},
					),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "access", "plan"),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "permissions.0.runs", "plan"),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "permissions.0.variables", "read"),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "permissions.0.state_versions", "read"),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "permissions.0.sentinel_mocks", "none"),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "permissions.0.workspace_locking", "false"),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "permissions.0.run_tasks", "false"),
				),
			},
		},
	})
}

func TestAccTFETeamAccess_import(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFETeamAccessDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamAccess_write(rInt),
			},

			{
				ResourceName:        "tfe_team_access.foobar",
				ImportState:         true,
				ImportStateIdPrefix: fmt.Sprintf("tst-terraform-%d/workspace-test/", rInt),
				ImportStateVerify:   true,
			},
		},
	})
}

func testAccCheckTFETeamAccessExists(
	n string, tmAccess *tfe.TeamAccess) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(ConfiguredClient)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no instance ID is set")
		}

		ta, err := config.Client.TeamAccess.Read(ctx, rs.Primary.ID)
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

func testAccCheckTFETeamAccessAttributesAccessIs(tmAccess *tfe.TeamAccess, access tfe.AccessType) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if tmAccess.Access != access {
			return fmt.Errorf("Bad access: %s", tmAccess.Access)
		}
		return nil
	}
}

func testAccCheckTFETeamAccessAttributesPermissionsAre(tmAccess *tfe.TeamAccess, expectedPermissions map[string]interface{}) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if tmAccess.Runs != expectedPermissions["runs"].(tfe.RunsPermissionType) {
			return fmt.Errorf("Bad runs permission: Expected %s, Received %s", expectedPermissions["runs"], tmAccess.Runs)
		}
		if tmAccess.Variables != expectedPermissions["variables"].(tfe.VariablesPermissionType) {
			return fmt.Errorf("Bad variables permission: Expected %s, Received %s", expectedPermissions["variables"], tmAccess.Variables)
		}
		if tmAccess.StateVersions != expectedPermissions["state_versions"].(tfe.StateVersionsPermissionType) {
			return fmt.Errorf("Bad state-versions permission: Expected %s, Received %s", expectedPermissions["state_versions"], tmAccess.StateVersions)
		}
		if tmAccess.SentinelMocks != expectedPermissions["sentinel_mocks"].(tfe.SentinelMocksPermissionType) {
			return fmt.Errorf("Bad sentinel-mocks permission: Expected %s, Received %s", expectedPermissions["sentinel_mocks"], tmAccess.SentinelMocks)
		}
		if tmAccess.WorkspaceLocking != expectedPermissions["workspace_locking"].(bool) {
			return fmt.Errorf("Bad workspace-locking permission: Expected %s, Received %t", expectedPermissions["workspace_locking"], tmAccess.WorkspaceLocking)
		}
		if tmAccess.RunTasks != expectedPermissions["run_tasks"].(bool) {
			return fmt.Errorf("Bad run_tasks permission: Expected %s, Received %t", expectedPermissions["run_tasks"], tmAccess.RunTasks)
		}
		return nil
	}
}

func testAccCheckTFETeamAccessDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(ConfiguredClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_team_access" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := config.Client.TeamAccess.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Team access %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccTFETeamAccess_admin(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_team_access" "foobar" {
  access       = "admin"
  team_id      = tfe_team.foobar.id
  workspace_id = tfe_workspace.foobar.id
}`, rInt)
}

func testAccTFETeamAccess_write(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_team_access" "foobar" {
  access       = "write"
  team_id      = tfe_team.foobar.id
  workspace_id = tfe_workspace.foobar.id
}`, rInt)
}

func testAccTFETeamAccess_plan(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_team_access" "foobar" {
  access       = "plan"
  team_id      = tfe_team.foobar.id
  workspace_id = tfe_workspace.foobar.id
}`, rInt)
}

func testAccTFETeamAccess_custom(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_team_access" "foobar" {
  permissions {
    runs = "apply"
    variables = "read"
    state_versions = "read-outputs"
    sentinel_mocks = "none"
    workspace_locking = false
    run_tasks = false
  }
  team_id      = tfe_team.foobar.id
  workspace_id = tfe_workspace.foobar.id
}`, rInt)
}
