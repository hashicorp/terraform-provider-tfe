// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTFEProjectDataSource_basic(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEProjectDataSourceConfig(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.tfe_project.foobar", "name", fmt.Sprintf("project-test-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.tfe_project.foobar", "description", "project description"),
					resource.TestCheckResourceAttr(
						"data.tfe_project.foobar", "organization", orgName),
					resource.TestCheckResourceAttrSet("data.tfe_project.foobar", "id"),
					resource.TestCheckResourceAttr(
						"data.tfe_project.foobar", "workspace_ids.#", "1"),
					resource.TestCheckResourceAttr(
						"data.tfe_project.foobar", "workspace_names.#", "1"),
				),
			},
		},
	})
}

func TestAccTFEProjectDataSource_caseInsensitive(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEProjectDataSourceConfigCaseInsensitive(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.tfe_project.foobar", "name", fmt.Sprintf("PROJECT-TEST-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.tfe_project.foobar", "organization", orgName),
					resource.TestCheckResourceAttrSet("data.tfe_project.foobar", "id"),
				),
			},
		},
	})
}

func TestAccTFEProjectDataSource_basicWithAutoDestroy(t *testing.T) {
	skipIfEnterprise(t)
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}
	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEProjectDataSourceConfigWithAutoDestroy(rInt, org.Name, "3d"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.tfe_project.foobar", "name", fmt.Sprintf("project-test-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.tfe_project.foobar", "auto_destroy_activity_duration", "3d"),
				),
			},
		},
	})
}

func testAccTFEProjectDataSourceConfig(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_project" "foobar" {
  name         = "project-test-%d"
  description  = "project description"
  organization = tfe_organization.foobar.id
}

data "tfe_project" "foobar" {
  name         = tfe_project.foobar.name
  organization = tfe_project.foobar.organization
  # Read the data source after creating the workspace, so counts match
  depends_on = [
	tfe_workspace.foobar
  ]
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test-%d"
  organization = tfe_organization.foobar.id
  project_id  = tfe_project.foobar.id
}`, rInt, rInt, rInt)
}

func testAccTFEProjectDataSourceConfigCaseInsensitive(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_project" "foobar" {
  name         = "project-test-%d"
  organization = tfe_organization.foobar.id
}

data "tfe_project" "foobar" {
  name         = "PROJECT-TEST-%d"
  organization = tfe_project.foobar.organization
  # Read the data source after creating the project
  depends_on = [
    tfe_project.foobar
  ]
}`, rInt, rInt, rInt)
}

func testAccTFEProjectDataSourceConfigWithAutoDestroy(rInt int, orgName string, duration string) string {
	return fmt.Sprintf(`
resource "tfe_project" "foobar" {
  name         = "project-test-%d"
  description  = "project description"
  organization = "%s"
  auto_destroy_activity_duration = "%s"
}

data "tfe_project" "foobar" {
  name         = tfe_project.foobar.name
  organization = tfe_project.foobar.organization
}`, rInt, orgName, duration)
}
