package tfe

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccTFEOrganizationDataSource_basic(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	org := &tfe.Organization{}
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)
	fmt.Println(orgName)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrganizationDataSourceConfig_basic(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTFEOrganizationExists(
						"tfe_organization.foobar", org),
					// names attribute
					resource.TestCheckResourceAttr(
						"data.tfe_organization.foobar", "name", orgName),
					resource.TestCheckResourceAttr(
						"data.tfe_organization.foobar", "email", "admin@company.com"),
				),
			},
		},
	})
}

func testAccTFEOrganizationDataSourceConfig_basic(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

data "tfe_organization" "foobar" {
	name = "tst-terraform-%d"
}`, rInt, rInt)
}
