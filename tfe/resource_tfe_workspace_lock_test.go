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
					testAccCheckTFEWorkspaceAttributes(workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspace_lock.foobar", "id", workspace.ID),
					resource.TestCheckResourceAttr(
						"tfe_workspace_lock.foobar", "workspace_id", workspace.ID),
					resource.TestCheckResourceAttr(
						"tfe_workspace_lock.foobar", "reason", "test"),
				),
			},
		},
	})
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
