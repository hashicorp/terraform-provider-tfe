// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccTFETeamAccessDataSource_basic(t *testing.T) {
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
				Config: testAccTFETeamAccessDataSourceConfig(org.Name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.tfe_team_access.foobar", "access", "write"),
					resource.TestCheckResourceAttr(
						"data.tfe_team_access.foobar", "permissions.0.runs", "apply"),
					resource.TestCheckResourceAttr(
						"data.tfe_team_access.foobar", "permissions.0.variables", "write"),
					resource.TestCheckResourceAttr(
						"data.tfe_team_access.foobar", "permissions.0.state_versions", "write"),
					resource.TestCheckResourceAttr(
						"data.tfe_team_access.foobar", "permissions.0.sentinel_mocks", "read"),
					resource.TestCheckResourceAttr(
						"data.tfe_team_access.foobar", "permissions.0.workspace_locking", "true"),
					resource.TestCheckResourceAttr(
						"data.tfe_team_access.foobar", "permissions.0.run_tasks", "false"),
					resource.TestCheckResourceAttrSet("data.tfe_team_access.foobar", "id"),
					resource.TestCheckResourceAttrSet("data.tfe_team_access.foobar", "team_id"),
					resource.TestCheckResourceAttrSet("data.tfe_team_access.foobar", "workspace_id"),
				),
			},
		},
	})
}

func testAccTFETeamAccessDataSourceConfig(organization string) string {
	return fmt.Sprintf(`
resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = "%s"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = "%s"
}

resource "tfe_team_access" "foobar" {
  access       = "write"
  team_id      = tfe_team.foobar.id
  workspace_id = tfe_workspace.foobar.id
}

data "tfe_team_access" "foobar" {
  team_id      = tfe_team.foobar.id
  workspace_id = tfe_team_access.foobar.workspace_id
}`, organization, organization)
}
