package tfe

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccTFEStateOutputs(t *testing.T) {
	skipIfFreeOnly(t)

	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	expectedOutput := "hello world"
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)
	wsName := fmt.Sprintf("workspace-test-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEStateOutputs_defaultOutputs(rInt, expectedOutput),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "name", orgName),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "name", wsName),
					resource.TestCheckOutput(
						"foo", expectedOutput),
				),
			},
			{
				Config: testAccTFEStateOutputs_dataSource(rInt, orgName, wsName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "name", fmt.Sprintf("tst-terraform-%d", rInt)),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "name", fmt.Sprintf("workspace-test-%d", rInt)),
					resource.TestCheckOutput(
						"states", expectedOutput),
				),
			},
		},
	})
}

func testAccTFEStateOutputs_defaultOutputs(rInt int, output string) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name                  = "workspace-test-%d"
  organization          = tfe_organization.foobar.name
}

output "foo" {
	value = "%s"
}`, rInt, rInt, output)
}

func testAccTFEStateOutputs_dataSource(rInt int, org, workspace string) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name                  = "workspace-test-%d"
  organization          = tfe_organization.foobar.name
}

data "tfe_state_outputs" "foobar" {
  organization = "%s"
  workspace = "%s"
}

output "states" {
	// this references the 'output "foo"' in the testAccTFEStateOutputs_defaultOutputs config
  value = data.tfe_state_outputs.foobar.values.foo
}`, rInt, rInt, org, workspace)
}
