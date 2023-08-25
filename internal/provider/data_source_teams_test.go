// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccTFETeamsDataSource_basic(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	org, orgCleanup := createOrganization(t, tfeClient, tfe.OrganizationCreateOptions{
		Name:  tfe.String(fmt.Sprintf("tst-terraform-%d", rInt)),
		Email: tfe.String(fmt.Sprintf("%s@tfe.local", randomString(t))),
	})
	t.Cleanup(orgCleanup)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamsDataSourceConfig_basic_resource(rInt, org.Name),
			},
			{
				Config: testAccTFETeamsDataSourceConfig_basic_data(org.Name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTFETeamsHasNames("data.tfe_teams.foobar", []string{
						fmt.Sprintf("team-foo-%d", rInt),
						fmt.Sprintf("team-bar-%d", rInt),
					}),
					testAccCheckTFETeamsHasIDs("data.tfe_teams.foobar", []string{
						fmt.Sprintf("team-foo-%d", rInt),
						fmt.Sprintf("team-bar-%d", rInt),
					}),
				),
			},
		},
	})
}

func testAccCheckTFETeamsHasNames(teamsData string, teamNames []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		teams, ok := s.RootModule().Resources[teamsData]
		if !ok {
			return fmt.Errorf("Teams data '%s' not found.", teamsData)
		}
		numTeamsStr := teams.Primary.Attributes["names.#"]
		numTeams, _ := strconv.Atoi(numTeamsStr)

		if numTeams < len(teamNames) {
			return fmt.Errorf("expected %d organizations, but found %d.", len(teamNames), numTeams)
		}

		teamsMap := map[string]struct{}{}
		for i := 0; i < numTeams; i++ {
			teamName := teams.Primary.Attributes[fmt.Sprintf("names.%d", i)]
			teamsMap[teamName] = struct{}{}
		}

		for _, teamName := range teamNames {
			_, ok := teamsMap[teamName]
			if !ok {
				return fmt.Errorf("expected to find team name %s, but did not.", teamName)
			}
		}

		return nil
	}
}

func testAccCheckTFETeamsHasIDs(teamsData string, teamNames []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		teams, ok := s.RootModule().Resources[teamsData]
		if !ok {
			return fmt.Errorf("Teams data '%s' not found.", teamsData)
		}

		for _, teamName := range teamNames {
			id := fmt.Sprintf("ids.%s", teamName)
			_, ok := teams.Primary.Attributes[id]
			if !ok {
				return fmt.Errorf("expected to find team id %s, but did not.", id)
			}
		}

		return nil
	}
}

func testAccTFETeamsDataSourceConfig_basic_resource(rInt int, organization string) string {
	return fmt.Sprintf(`
resource "tfe_team" "foo" {
  name         = "team-foo-%d"
  organization = "%s"
}

resource "tfe_team" "bar" {
	name         = "team-bar-%d"
	organization = "%s"
}`, rInt, organization, rInt, organization)
}

func testAccTFETeamsDataSourceConfig_basic_data(organization string) string {
	return fmt.Sprintf(`
	data "tfe_teams" "foobar" {
		organization = "%s"
	}`, organization)
}
