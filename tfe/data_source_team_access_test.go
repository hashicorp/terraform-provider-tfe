package tfe

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccTFETeamAccessDataSource_basic(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamAccessDataSourceConfig(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.tfe_team_access.foobar", "access", "write"),
					resource.TestCheckResourceAttrSet("data.tfe_team_access.foobar", "id"),
					resource.TestCheckResourceAttrSet("data.tfe_team_access.foobar", "team_id"),
					resource.TestCheckResourceAttrSet("data.tfe_team_access.foobar", "workspace_id"),
				),
			},
		},
	})
}

func testAccTFETeamAccessDataSourceConfig(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_team" "foobar" {
  name         = "team-test-%d"
  organization = "${tfe_organization.foobar.id}"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test-%d"
  organization = "${tfe_organization.foobar.id}"
}

resource "tfe_team_access" "foobar" {
  access       = "write"
  team_id      = "${tfe_team.foobar.id}"
  workspace_id = "${tfe_workspace.foobar.id}"
}

data "tfe_team_access" "foobar" {
  team_id      = "${tfe_team.foobar.id}"
  workspace_id = "${tfe_team_access.foobar.workspace_id}"
}`, rInt, rInt, rInt)
}
