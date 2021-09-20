package tfe

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccTFEVariablesDataSource_basic(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEVariablesDataSourceConfig_basic(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					// variables attribute
					resource.TestCheckResourceAttrSet("data.tfe_variables.foobar", "id"),
					resource.TestCheckOutput("foobar", "foo"),
				),
			},
		},
	},
	)
}

func testAccTFEVariablesDataSourceConfig_basic(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "org-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-foo-%d"
  organization = tfe_organization.foobar.id
}

resource "tfe_variable" "foobar" {
	key          = "foo"
	value        = "bar"
	category     = "terraform"
	workspace_id = tfe_workspace.foobar.id
}

data "tfe_variables" "foobar" {
	workspace_id = tfe_workspace.foobar.id
	depends_on = [
    tfe_variable.foobar
  ]
}

output "foobar" {
	value = data.tfe_variables.foobar.variables[0]["name"]
}`, rInt, rInt)
}
