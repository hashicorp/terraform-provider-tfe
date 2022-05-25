package tfe

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccTFEWorkspaceRunTaskDataSource_basic(t *testing.T) {
	skipUnlessRunTasksDefined(t)
	skipIfFreeOnly(t) // Run Tasks requires TFE or a TFC paid/trial subscription

	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			testCheckCreateOrgWithRunTasks(orgName),
			{
				Config: testAccTFEWorkspaceRunTaskDataSourceConfig(orgName, rInt, runTasksURL()),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.tfe_workspace_run_task.foobar", "enforcement_level", "advisory"),
					resource.TestCheckResourceAttrSet("data.tfe_workspace_run_task.foobar", "id"),
					resource.TestCheckResourceAttrSet("data.tfe_workspace_run_task.foobar", "task_id"),
					resource.TestCheckResourceAttrSet("data.tfe_workspace_run_task.foobar", "workspace_id"),
				),
			},
		},
	})
}

func testAccTFEWorkspaceRunTaskDataSourceConfig(orgName string, rInt int, runTaskURL string) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
	name  = "%s"
	email = "admin@company.com"
}

resource "tfe_organization_run_task" "foobar" {
	organization = tfe_organization.foobar.id
	url          = "%s"
	name         = "foobar-task-%d"
}

resource "tfe_workspace" "foobar" {
	name         = "workspace-test-%d"
	organization = tfe_organization.foobar.id
}

resource "tfe_workspace_run_task" "foobar" {
	workspace_id      = resource.tfe_workspace.foobar.id
	task_id           = resource.tfe_organization_run_task.foobar.id
	enforcement_level = "advisory"
}

data "tfe_workspace_run_task" "foobar" {
	workspace_id      = resource.tfe_workspace.foobar.id
	task_id           = resource.tfe_organization_run_task.foobar.id
	depends_on = [tfe_workspace_run_task.foobar]
}`, orgName, runTaskURL, rInt, rInt)
}
