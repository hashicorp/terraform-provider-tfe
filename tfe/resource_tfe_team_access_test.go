package tfe

import (
	"fmt"
	"testing"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccTFETeamAccess_write(t *testing.T) {
	tmAccess := &tfe.TeamAccess{}
	expectedPermissions := map[string]interface{}{
		"runs":              tfe.RunsPermissionApply,
		"variables":         tfe.VariablesPermissionWrite,
		"state_versions":    tfe.StateVersionsPermissionWrite,
		"sentinel_mocks":    tfe.SentinelMocksPermissionRead,
		"workspace_locking": true,
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFETeamAccessDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamAccess_write,
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
				),
			},
		},
	})
}

func TestAccTFETeamAccess_custom(t *testing.T) {
	tmAccess := &tfe.TeamAccess{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFETeamAccessDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamAccess_custom,
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
						},
					),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "access", "custom"),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "permissions.0.runs", "apply"),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "permissions.0.variables", "read"),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "permissions.0.state_versions", "read-outputs"),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "permissions.0.sentinel_mocks", "none"),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "permissions.0.workspace_locking", "false"),
				),
			},
		},
	})
}

func TestAccTFETeamAccess_updateToCustom(t *testing.T) {
	tmAccess := &tfe.TeamAccess{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFETeamAccessDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamAccess_write,
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
						},
					),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "access", "write"),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "permissions.0.runs", "apply"),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "permissions.0.variables", "write"),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "permissions.0.state_versions", "write"),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "permissions.0.sentinel_mocks", "read"),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "permissions.0.workspace_locking", "true"),
				),
			},
			{
				Config: testAccTFETeamAccess_custom,
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
						},
					),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "access", "custom"),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "permissions.0.runs", "apply"),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "permissions.0.variables", "read"),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "permissions.0.state_versions", "read-outputs"),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "permissions.0.sentinel_mocks", "none"),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "permissions.0.workspace_locking", "false"),
				),
			},
		},
	})
}

func TestAccTFETeamAccess_updateFromCustom(t *testing.T) {
	tmAccess := &tfe.TeamAccess{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFETeamAccessDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamAccess_custom,
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
						},
					),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "access", "custom"),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "permissions.0.runs", "apply"),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "permissions.0.variables", "read"),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "permissions.0.state_versions", "read-outputs"),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "permissions.0.sentinel_mocks", "none"),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "permissions.0.workspace_locking", "false"),
				),
			},
			{
				Config: testAccTFETeamAccess_plan,
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
						},
					),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "access", "plan"),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "permissions.0.runs", "plan"),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "permissions.0.variables", "read"),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "permissions.0.state_versions", "read"),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "permissions.0.sentinel_mocks", "none"),
					resource.TestCheckResourceAttr("tfe_team_access.foobar", "permissions.0.workspace_locking", "false"),
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
			{
				Config: testAccTFETeamAccess_write,
			},

			{
				ResourceName:        "tfe_team_access.foobar",
				ImportState:         true,
				ImportStateIdPrefix: "tst-terraform/workspace-test/",
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

const testAccTFETeamAccess_write = `
resource "tfe_organization" "foobar" {
  name  = "tst-terraform"
  email = "admin@company.com"
}

resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = "${tfe_organization.foobar.id}"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = "${tfe_organization.foobar.id}"
}

resource "tfe_team_access" "foobar" {
  access       = "write"
  team_id      = "${tfe_team.foobar.id}"
  workspace_id = "${tfe_workspace.foobar.id}"
}`

const testAccTFETeamAccess_plan = `
resource "tfe_organization" "foobar" {
  name  = "tst-terraform"
  email = "admin@company.com"
}

resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = "${tfe_organization.foobar.id}"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = "${tfe_organization.foobar.id}"
}

resource "tfe_team_access" "foobar" {
  access       = "plan"
  team_id      = "${tfe_team.foobar.id}"
  workspace_id = "${tfe_workspace.foobar.id}"
}`

const testAccTFETeamAccess_custom = `
resource "tfe_organization" "foobar" {
  name  = "tst-terraform"
  email = "admin@company.com"
}

resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = "${tfe_organization.foobar.id}"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = "${tfe_organization.foobar.id}"
}

resource "tfe_team_access" "foobar" {
  permissions {
    runs = "apply"
    variables = "read"
    state_versions = "read-outputs"
    sentinel_mocks = "none"
    workspace_locking = false
  }
  team_id      = "${tfe_team.foobar.id}"
  workspace_id = "${tfe_workspace.foobar.id}"
}`
