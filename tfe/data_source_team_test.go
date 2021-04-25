package tfe

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccTFETeamDataSource_basic(t *testing.T) {
	skipIfFreeOnly(t)

	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamDataSourceConfig(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.tfe_team.foobar", "name", fmt.Sprintf("team-test-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.tfe_team.foobar", "organization", orgName),
					resource.TestCheckResourceAttrSet("data.tfe_team.foobar", "id"),
				),
			},
		},
	})
}

func testAccTFETeamDataSourceConfig(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_team" "foobar" {
  name         = "team-test-%d"
  organization = tfe_organization.foobar.id
}

data "tfe_team" "foobar" {
  name         = tfe_team.foobar.name
  organization = tfe_team.foobar.organization
}`, rInt, rInt)
}
