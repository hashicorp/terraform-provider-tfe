// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tfe

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccTFEProjectDataSource_basic(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("org-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEProjectDataSourceConfig(orgName, rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.tfe_project.foobar", "name", "projecttest"),
					resource.TestCheckResourceAttrSet("data.tfe_project.foobar", "id"),
					resource.TestCheckResourceAttr(
						"data.tfe_project.foobar", "workspace_ids.#", "1"),
				),
			},
		},
	})
}

func testAccTFEProjectDataSourceConfig(organization string, rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "org-%d"
  email = "admin@company.com"
}

resource "tfe_project" "foobar" {
  name         = "project-%d"
  organization = "org-%d"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-%d"
  organization = tfe_organization.foobar.id
  project_id  = tfe_project.foobar.id
}

data "tfe_project" "foobar" {
  name         = tfe_project.foobar.name
  organization = "%s"
}`, rInt, rInt, rInt, rInt, organization)
}
