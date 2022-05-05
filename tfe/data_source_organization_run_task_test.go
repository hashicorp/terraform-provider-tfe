package tfe

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccTFEOrganizationRunTaskDataSource_basic(t *testing.T) {
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
				Config: testAccTFEOrganizationRunTaskDataSourceConfig(orgName, rInt, runTasksUrl()),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.tfe_organization_run_task.foobar", "name", fmt.Sprintf("foobar-task-%d", rInt)),
					resource.TestCheckResourceAttr("data.tfe_organization_run_task.foobar", "url", runTasksUrl()),
					resource.TestCheckResourceAttr("data.tfe_organization_run_task.foobar", "category", "task"),
					resource.TestCheckResourceAttr("data.tfe_organization_run_task.foobar", "enabled", "false"),
					resource.TestCheckResourceAttrSet("data.tfe_organization_run_task.foobar", "id"),
					resource.TestCheckResourceAttrSet("data.tfe_organization_run_task.foobar", "organization"),
				),
			},
		},
	})
}

func testAccTFEOrganizationRunTaskDataSourceConfig(orgName string, rInt int, runTaskUrl string) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
	name  = "%s"
	email = "admin@company.com"
}

resource "tfe_organization_run_task" "foobar" {
	organization = tfe_organization.foobar.id
	url          = "%s"
	name         = "foobar-task-%d"
	hmac_key     = "Password1"
	enabled      = false
}

data "tfe_organization_run_task" "foobar" {
	organization      = resource.tfe_organization.foobar.id
	name              = "foobar-task-%d"
	depends_on = [tfe_organization_run_task.foobar]
}`, orgName, runTaskUrl, rInt, rInt)
}
