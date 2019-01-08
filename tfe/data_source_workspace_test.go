package tfe

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccTFEWorkspaceDataSource_basic(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEWorkspaceDataSourceConfig(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.tfe_workspace.foobar", "name", fmt.Sprintf("workspace-test-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.tfe_workspace.foobar", "organization", fmt.Sprintf("terraform-test-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.tfe_workspace.foobar", "auto_apply", "true"),
					resource.TestCheckResourceAttr(
						"data.tfe_workspace.foobar", "queue_all_runs", "false"),
					resource.TestCheckResourceAttr(
						"data.tfe_workspace.foobar", "terraform_version", "0.11.1"),
					resource.TestCheckResourceAttr(
						"data.tfe_workspace.foobar", "working_directory", "terraform/test"),
					resource.TestCheckResourceAttrSet("data.tfe_workspace.foobar", "id"),
					resource.TestCheckResourceAttrSet("data.tfe_workspace.foobar", "external_id"),
				),
			},
		},
	})
}

func testAccTFEWorkspaceDataSourceConfig(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "terraform-test-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name              = "workspace-test-%d"
  organization      = "${tfe_organization.foobar.id}"
  auto_apply        = true
  queue_all_runs    = false
  terraform_version = "0.11.1"
  working_directory = "terraform/test"
}

data "tfe_workspace" "foobar" {
  name         = "${tfe_workspace.foobar.name}"
  organization = "${tfe_workspace.foobar.organization}"
}`, rInt, rInt)
}
