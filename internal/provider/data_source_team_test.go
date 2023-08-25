// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccTFETeamDataSource_basic(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamDataSourceConfig_basic(rInt, org.Name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.tfe_team.foobar", "name", fmt.Sprintf("team-test-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.tfe_team.foobar", "organization", org.Name),
					resource.TestCheckResourceAttrSet("data.tfe_team.foobar", "id"),
				),
			},
		},
	})
}

func TestAccTFETeamDataSource_ssoTeamId(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	testSsoTeamID := fmt.Sprintf("sso-team-id-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamDataSourceConfig_ssoTeamId(rInt, org.Name, testSsoTeamID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.tfe_team.sso_team", "name", fmt.Sprintf("team-test-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.tfe_team.sso_team", "organization", org.Name),
					resource.TestCheckResourceAttrSet("data.tfe_team.sso_team", "id"),
					resource.TestCheckResourceAttr(
						"data.tfe_team.sso_team", "sso_team_id", testSsoTeamID),
				),
			},
		},
	})
}

func testAccTFETeamDataSourceConfig_basic(rInt int, organization string) string {
	return fmt.Sprintf(`
resource "tfe_team" "foobar" {
  name         = "team-test-%d"
  organization = "%s"
}

data "tfe_team" "foobar" {
  name         = tfe_team.foobar.name
  organization = "%s"
}`, rInt, organization, organization)
}

func testAccTFETeamDataSourceConfig_ssoTeamId(rInt int, organization string, ssoTeamID string) string {
	return fmt.Sprintf(`
resource "tfe_team" "sso_team" {
  name         = "team-test-%d"
  organization = "%s"
  sso_team_id  = "%s"
}

data "tfe_team" "sso_team" {
  name         = tfe_team.sso_team.name
  organization = tfe_team.sso_team.organization
}`, rInt, organization, ssoTeamID)
}
