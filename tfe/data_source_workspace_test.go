package tfe

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccTFEWorkspaceDataSource_basic(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEWorkspaceDataSourceConfig(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tfe_workspace.foobar", "id"),
					resource.TestCheckResourceAttr(
						"data.tfe_workspace.foobar", "name", fmt.Sprintf("workspace-test-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.tfe_workspace.foobar", "organization", orgName),
					resource.TestCheckResourceAttr(
						"data.tfe_workspace.foobar", "description", "provider-testing"),
					resource.TestCheckResourceAttr(
						"data.tfe_workspace.foobar", "allow_destroy_plan", "false"),
					resource.TestCheckResourceAttr(
						"data.tfe_workspace.foobar", "auto_apply", "true"),
					resource.TestCheckResourceAttr(
						"data.tfe_workspace.foobar", "file_triggers_enabled", "true"),
					resource.TestCheckResourceAttr(
						"data.tfe_workspace.foobar", "policy_check_failures", "0"),
					resource.TestCheckResourceAttr(
						"data.tfe_workspace.foobar", "queue_all_runs", "false"),
					resource.TestCheckResourceAttr(
						"data.tfe_workspace.foobar", "resource_count", "0"),
					resource.TestCheckResourceAttr(
						"data.tfe_workspace.foobar", "run_failures", "0"),
					resource.TestCheckResourceAttr(
						"data.tfe_workspace.foobar", "runs_count", "0"),
					resource.TestCheckResourceAttr(
						"data.tfe_workspace.foobar", "speculative_enabled", "true"),
					resource.TestCheckResourceAttr(
						"data.tfe_workspace.foobar", "structured_run_output_enabled", "true"),
					resource.TestCheckResourceAttr(
						"data.tfe_workspace.foobar", "tag_names.0", "modules"),
					resource.TestCheckResourceAttr(
						"data.tfe_workspace.foobar", "tag_names.1", "shared"),
					resource.TestCheckResourceAttr(
						"data.tfe_workspace.foobar", "terraform_version", "0.11.1"),
					resource.TestCheckResourceAttr(
						"data.tfe_workspace.foobar", "trigger_prefixes.#", "2"),
					resource.TestCheckResourceAttr(
						"data.tfe_workspace.foobar", "trigger_prefixes.0", "/modules"),
					resource.TestCheckResourceAttr(
						"data.tfe_workspace.foobar", "trigger_prefixes.1", "/shared"),
					resource.TestCheckResourceAttr(
						"data.tfe_workspace.foobar", "working_directory", "terraform/test"),
					resource.TestCheckOutput("foobar", "foo"),
				),
			},
		},
	})
}

func testAccTFEWorkspaceDataSourceConfig(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name                  = "workspace-test-%d"
  organization          = tfe_organization.foobar.id
  description           = "provider-testing"
  allow_destroy_plan    = false
  auto_apply            = true
  file_triggers_enabled = true
  queue_all_runs        = false
  speculative_enabled   = true
  tag_names             = ["modules", "shared"]
  terraform_version     = "0.11.1"
  trigger_prefixes      = ["/modules", "/shared"]
  working_directory     = "terraform/test"
}

resource "tfe_variable" "foobar" {
	workspace_id = tfe_workspace.foobar.id
	category = "terraform"
	key = "foo"
	value = "bar"
}

data "tfe_workspace" "foobar" {
  name         = tfe_workspace.foobar.name
  organization = tfe_workspace.foobar.organization
}

output "foobar" {
	value = data.tfe_workspace.foobar.variables[0]["name"]
}`, rInt, rInt)
}
