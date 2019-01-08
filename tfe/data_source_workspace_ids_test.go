package tfe

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccTFEWorkspaceIDsDataSource_basic(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEWorkspaceIDsDataSourceConfig_basic(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.tfe_workspace_ids.foobar", "names.#", "2"),
					resource.TestCheckResourceAttr(
						"data.tfe_workspace_ids.foobar", "names.0", fmt.Sprintf("workspace-foo-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.tfe_workspace_ids.foobar", "names.1", fmt.Sprintf("workspace-bar-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.tfe_workspace_ids.foobar", "organization", fmt.Sprintf("terraform-test-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.tfe_workspace_ids.foobar", "ids.%", "2"),
					resource.TestCheckResourceAttr(
						"data.tfe_workspace_ids.foobar",
						fmt.Sprintf("ids.workspace-foo-%d", rInt),
						fmt.Sprintf("terraform-test-%d/workspace-foo-%d", rInt, rInt),
					),
					resource.TestCheckResourceAttr(
						"data.tfe_workspace_ids.foobar",
						fmt.Sprintf("ids.workspace-bar-%d", rInt),
						fmt.Sprintf("terraform-test-%d/workspace-bar-%d", rInt, rInt),
					),
					resource.TestCheckResourceAttr(
						"data.tfe_workspace_ids.foobar", "external_ids.%", "2"),
					resource.TestCheckResourceAttrSet(
						"data.tfe_workspace_ids.foobar", fmt.Sprintf("external_ids.workspace-foo-%d", rInt)),
					resource.TestCheckResourceAttrSet(
						"data.tfe_workspace_ids.foobar", fmt.Sprintf("external_ids.workspace-bar-%d", rInt)),
					resource.TestCheckResourceAttrSet("data.tfe_workspace_ids.foobar", "id"),
				),
			},
		},
	})
}

func TestAccTFEWorkspaceIDsDataSource_wildcard(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEWorkspaceIDsDataSourceConfig_wildcard(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.tfe_workspace_ids.foobar", "names.#", "1"),
					resource.TestCheckResourceAttr(
						"data.tfe_workspace_ids.foobar", "names.0", "*"),
					resource.TestCheckResourceAttr(
						"data.tfe_workspace_ids.foobar", "organization", fmt.Sprintf("terraform-test-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.tfe_workspace_ids.foobar", "ids.%", "3"),
					resource.TestCheckResourceAttr(
						"data.tfe_workspace_ids.foobar",
						fmt.Sprintf("ids.workspace-foo-%d", rInt),
						fmt.Sprintf("terraform-test-%d/workspace-foo-%d", rInt, rInt),
					),
					resource.TestCheckResourceAttr(
						"data.tfe_workspace_ids.foobar",
						fmt.Sprintf("ids.workspace-bar-%d", rInt),
						fmt.Sprintf("terraform-test-%d/workspace-bar-%d", rInt, rInt),
					),
					resource.TestCheckResourceAttr(
						"data.tfe_workspace_ids.foobar",
						fmt.Sprintf("ids.workspace-dummy-%d", rInt),
						fmt.Sprintf("terraform-test-%d/workspace-dummy-%d", rInt, rInt),
					),
					resource.TestCheckResourceAttr(
						"data.tfe_workspace_ids.foobar", "external_ids.%", "3"),
					resource.TestCheckResourceAttrSet(
						"data.tfe_workspace_ids.foobar", fmt.Sprintf("external_ids.workspace-foo-%d", rInt)),
					resource.TestCheckResourceAttrSet(
						"data.tfe_workspace_ids.foobar", fmt.Sprintf("external_ids.workspace-bar-%d", rInt)),
					resource.TestCheckResourceAttrSet(
						"data.tfe_workspace_ids.foobar", fmt.Sprintf("external_ids.workspace-dummy-%d", rInt)),
					resource.TestCheckResourceAttrSet("data.tfe_workspace_ids.foobar", "id"),
				),
			},
		},
	})
}

func testAccTFEWorkspaceIDsDataSourceConfig_basic(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "terraform-test-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foo" {
  name         = "workspace-foo-%d"
  organization = "${tfe_organization.foobar.id}"
}

resource "tfe_workspace" "bar" {
  name         = "workspace-bar-%d"
  organization = "${tfe_organization.foobar.id}"
}

resource "tfe_workspace" "dummy" {
  name         = "workspace-dummy-%d"
  organization = "${tfe_organization.foobar.id}"
}

data "tfe_workspace_ids" "foobar" {
  names        = ["${tfe_workspace.foo.name}", "${tfe_workspace.bar.name}"]
  organization = "${tfe_organization.foobar.name}"
}`, rInt, rInt, rInt, rInt)
}

func testAccTFEWorkspaceIDsDataSourceConfig_wildcard(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "terraform-test-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foo" {
  name         = "workspace-foo-%d"
  organization = "${tfe_organization.foobar.id}"
}

resource "tfe_workspace" "bar" {
  name         = "workspace-bar-%d"
  organization = "${tfe_organization.foobar.id}"
}

resource "tfe_workspace" "dummy" {
  name         = "workspace-dummy-%d"
  organization = "${tfe_organization.foobar.id}"
}

data "tfe_workspace_ids" "foobar" {
  names        = ["*"]
  organization = "${tfe_workspace.dummy.organization}"
}`, rInt, rInt, rInt, rInt)
}
