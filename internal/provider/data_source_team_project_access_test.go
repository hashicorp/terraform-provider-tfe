// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

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

func TestAccTFETeamProjectCustomAccessDataSource_basic(t *testing.T) {
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
				Config: testAccTFETeamProjectCustomAccessDataSourceConfig(org.Name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tfe_team_project_access.foobar_custom", "id"),
					resource.TestCheckResourceAttrSet("data.tfe_team_project_access.foobar_custom", "team_id"),
					resource.TestCheckResourceAttrSet("data.tfe_team_project_access.foobar_custom", "project_id"),
					resource.TestCheckResourceAttr(
						"data.tfe_team_project_access.foobar_custom", "access", "custom"),
					resource.TestCheckResourceAttr(
						"data.tfe_team_project_access.foobar_custom", "project_access.0.settings", "delete"),
					resource.TestCheckResourceAttr(
						"data.tfe_team_project_access.foobar_custom", "project_access.0.teams", "manage"),
					resource.TestCheckResourceAttr(
						"data.tfe_team_project_access.foobar_custom", "workspace_access.0.state_versions", "write"),
					resource.TestCheckResourceAttr(
						"data.tfe_team_project_access.foobar_custom", "workspace_access.0.sentinel_mocks", "read"),
					resource.TestCheckResourceAttr(
						"data.tfe_team_project_access.foobar_custom", "workspace_access.0.runs", "apply"),
					resource.TestCheckResourceAttr(
						"data.tfe_team_project_access.foobar_custom", "workspace_access.0.variables", "write"),
					resource.TestCheckResourceAttr(
						"data.tfe_team_project_access.foobar_custom", "workspace_access.0.create", "true"),
					resource.TestCheckResourceAttr(
						"data.tfe_team_project_access.foobar_custom", "workspace_access.0.locking", "true"),
					resource.TestCheckResourceAttr(
						"data.tfe_team_project_access.foobar_custom", "workspace_access.0.move", "true"),
					resource.TestCheckResourceAttr(
						"data.tfe_team_project_access.foobar_custom", "workspace_access.0.delete", "false"),
					resource.TestCheckResourceAttr(
						"data.tfe_team_project_access.foobar_custom", "workspace_access.0.run_tasks", "false"),
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

func testAccTFETeamProjectCustomAccessDataSourceConfig(organization string) string {
	return fmt.Sprintf(`
resource "tfe_team" "foobar_custom" {
  name         = "team-test2"
  organization = "%s"
}

resource "tfe_project" "foobar_custom" {
  name         = "projecttest2"
  organization = "%s"
}

resource "tfe_team_project_access" "foobar_custom" {
  access       = "custom"
  team_id      = tfe_team.foobar_custom.id
  project_id   = tfe_project.foobar_custom.id
  project_access {
    settings = "delete"
    teams    = "manage"
  }
  workspace_access {
    state_versions = "write"
    sentinel_mocks = "read"
		runs					 = "apply"
    variables      = "write"
    create         = true
    locking        = true
    move           = true
    delete         = false
    run_tasks      = false
  }
}

data "tfe_team_project_access" "foobar_custom" {
  team_id      = tfe_team.foobar_custom.id
  project_id   = tfe_project.foobar_custom.id
  depends_on   = [tfe_team_project_access.foobar_custom]
}`, organization, organization)
}
