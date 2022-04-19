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
				Config: testAccTFETeamDataSourceConfig_basic(rInt),
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

func TestAccTFETeamDataSource_ssoTeamId(t *testing.T) {
	skipIfFreeOnly(t)

	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)
	testSsoTeamId := fmt.Sprintf("sso-team-id-%d", rInt)
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamDataSourceConfig_ssoTeamId(rInt, testSsoTeamId),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.tfe_team.sso_team", "name", fmt.Sprintf("team-test-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.tfe_team.sso_team", "organization", orgName),
					resource.TestCheckResourceAttrSet("data.tfe_team.sso_team", "id"),
					resource.TestCheckResourceAttr(
						"data.tfe_team.sso_team", "sso_team_id", testSsoTeamId),
				),
			},
		},
	})
}

func testAccTFETeamDataSourceConfig_basic(rInt int) string {
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

func testAccTFETeamDataSourceConfig_ssoTeamId(rInt int, ssoTeamId string) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_team" "sso_team" {
  name         = "team-test-%d"
  organization = tfe_organization.foobar.id
  sso_team_id  = "%s"
}

data "tfe_team" "sso_team" {
  name         = tfe_team.sso_team.name
  organization = tfe_team.sso_team.organization
}`, rInt, rInt, ssoTeamId)
}
