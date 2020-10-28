package tfe

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccTFEOrganizationMembershipDataSource_basic(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrganizationMembershipDataSourceConfig(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.tfe_organization_membership.foobar", "email", "example@hashicorp.com"),
					resource.TestCheckResourceAttr(
						"data.tfe_organization_membership.foobar", "organization", orgName),
					resource.TestCheckResourceAttrSet("data.tfe_organization_membership.foobar", "user_id"),
				),
			},
		},
	})
}

func testAccTFEOrganizationMembershipDataSourceConfig(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_organization_membership" "foobar" {
  email        = "example@hashicorp.com"
  organization = tfe_organization.foobar.id
}

data "tfe_organization_membership" "foobar" {
  email        = tfe_organization_membership.foobar.email
  organization = tfe_organization.foobar.name
}`, rInt)
}
