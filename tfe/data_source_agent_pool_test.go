package tfe

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccTFEAgentPoolDataSource_basic(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEAgentPoolDataSourceConfig(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tfe_agent_pool.foobar", "id"),
					resource.TestCheckResourceAttr(
						"data.tfe_agent_pool.foobar", "name", fmt.Sprintf("agent-pool-test-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.tfe_agent_pool.foobar", "organization", fmt.Sprintf("tst-terraform-%d", rInt)),
				),
			},
		},
	})
}

func testAccTFEAgentPoolDataSourceConfig(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_agent_pool" "foobar" {
  name                  = "agent-pool-test-%d"
  organization          = tfe_organization.foobar.id
}

data "tfe_agent_pool" "foobar" {
  name         = tfe_agent_pool.foobar.name
  organization = tfe_agent_pool.foobar.organization
}`, rInt, rInt)
}
