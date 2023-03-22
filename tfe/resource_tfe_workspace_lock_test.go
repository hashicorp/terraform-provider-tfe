// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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

func TestAccTFEWorkspaceLock_basic(t *testing.T) {
	workspace := &tfe.Workspace{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEWorkspace_lock(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace, testAccProvider),
					testAccCheckTFEWorkspaceLocked(workspace, true),
					testAccCheckTFEWorkspaceLockedID(
						"tfe_workspace.foobar", "tfe_workspace_lock.foobar"),
					resource.TestCheckResourceAttr(
						"tfe_workspace_lock.foobar", "reason", "test"),
				),
			},
			{
				Config: testAccTFEWorkspace_unlock(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace, testAccProvider),
					testAccCheckTFEWorkspaceLocked(workspace, false),
				),
			},
		},
	})
}

func testAccCheckTFEWorkspaceLocked(
	workspace *tfe.Workspace, locked bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if workspace.Locked != locked {
			return fmt.Errorf("Expected workspaces lock status to be %t but got %t name", locked, workspace.Locked)
		}
		return nil
	}
}

func testAccCheckTFEWorkspaceLockedID(ws, lock string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[ws]
		if !ok {
			return fmt.Errorf("Not found: %s", ws)
		}

		wID := rs.Primary.ID

		if wID == "" {
			return fmt.Errorf("No ID is set")
		}

		rs, ok = s.RootModule().Resources[lock]
		if !ok {
			return fmt.Errorf("Not found: %s", ws)
		}

		if wID != rs.Primary.ID {
			return fmt.Errorf("expected lock ID to be workspace ID %s but got %s", wID, rs.Primary.ID)
		}
		if wID != rs.Primary.Attributes["workspace_id"] {
			return fmt.Errorf("expected lock workspace ID to be workspace ID %s but got %s", wID, rs.Primary.Attributes["workspace_id"])
		}

		return nil
	}
}

func testAccTFEWorkspace_lock(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name               = "workspace-test"
  organization       = tfe_organization.foobar.id
  description        = "My favorite workspace!"
  allow_destroy_plan = false
  auto_apply         = true
  tag_names          = ["fav", "test"]
}

resource "tfe_workspace_lock" "foobar" {
  workspace_id = tfe_workspace.foobar.id
  reason       = "test"
}
`, rInt)
}

func testAccTFEWorkspace_unlock(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name               = "workspace-test"
  organization       = tfe_organization.foobar.id
  description        = "My favorite workspace!"
  allow_destroy_plan = false
  auto_apply         = true
  tag_names          = ["fav", "test"]
}
`, rInt)
}
