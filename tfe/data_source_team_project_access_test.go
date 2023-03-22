// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tfe

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccTFETeamProjectAccessDataSource_basic(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamProjectAccessDataSourceConfig(org.Name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.tfe_team_project_access.foobar", "access", "read"),
					resource.TestCheckResourceAttrSet("data.tfe_team_project_access.foobar", "id"),
					resource.TestCheckResourceAttrSet("data.tfe_team_project_access.foobar", "team_id"),
					resource.TestCheckResourceAttrSet("data.tfe_team_project_access.foobar", "project_id"),
				),
			},
		},
	})
}

func testAccTFETeamProjectAccessDataSourceConfig(organization string) string {
	return fmt.Sprintf(`
resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = "%s"
}

resource "tfe_project" "foobar" {
  name         = "projecttest"
  organization = "%s"
}

resource "tfe_team_project_access" "foobar" {
  access       = "read"
  team_id      = tfe_team.foobar.id
  project_id   = tfe_project.foobar.id
}

data "tfe_team_project_access" "foobar" {
  team_id      = tfe_team.foobar.id
  project_id   = tfe_project.foobar.id
  depends_on = [tfe_team_project_access.foobar]
}`, organization, organization)
}
