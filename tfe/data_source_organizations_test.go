package tfe

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccTFEOrganizationsDataSource_basic(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrganizationsDataSourceConfig_basic(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					// names attribute
					resource.TestCheckResourceAttr(
						"data.tfe_organizations.foobarbaz", "names.#", "3"), // 3 organizations created
				),
			},
		},
	})
}

func testAccTFEOrganizationsDataSourceConfig_basic(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foo" {
  name  = "tst-terraform-foo-%d"
  email = "admin@company.com"
}

resource "tfe_organization" "bar" {
  name  = "tst-terraform-bar-%d"
  email = "admin@company.com"
}

resource "tfe_organization" "baz" {
  name  = "tst-terraform-baz-%d"
  email = "admin@company.com"
}

data "tfe_organizations" "foobarbaz" {
}`, rInt, rInt, rInt)
}
