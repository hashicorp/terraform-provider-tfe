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

func TestAccTFEVariableSetsDataSource_basic(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("org-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEVariableSetsDataSourceConfig_basic(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tfe_variable_set.foobar", "id"),
					resource.TestCheckResourceAttr(
						"data.tfe_variable_set.foobar", "name", fmt.Sprintf("varset-foo-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.tfe_variable_set.foobar", "description", "a description"),
					resource.TestCheckResourceAttr(
						"data.tfe_variable_set.foobar", "global", "false"),
					resource.TestCheckResourceAttr(
						"data.tfe_variable_set.foobar", "organization", orgName),
				),
			},
		},
	},
	)
}

func TestAccTFEVariableSetsDataSource_full(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEVariableSetsDataSourceConfig_full(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tfe_variable_set.foobar", "id"),
					resource.TestCheckResourceAttr(
						"data.tfe_variable_set.foobar", "name", fmt.Sprintf("varset-foo-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.tfe_variable_set.foobar", "workspace_ids.#", "1"),
					resource.TestCheckResourceAttr(
						"data.tfe_variable_set.foobar", "variable_ids.#", "1"),
					resource.TestCheckResourceAttr(
						"data.tfe_variable_set.foobar", "project_ids.#", "1"),
				),
			},
		},
	},
	)
}

func testAccTFEVariableSetsDataSourceConfig_basic(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "org-%d"
  email = "admin@company.com"
}

resource "tfe_variable_set" "foobar" {
  name = "varset-foo-%d"
	description = "a description"
	organization = tfe_organization.foobar.id
}

data "tfe_variable_set" "foobar" {
  name = tfe_variable_set.foobar.name
	organization = tfe_variable_set.foobar.organization
}`, rInt, rInt)
}

func testAccTFEVariableSetsDataSourceConfig_full(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "org-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-foo-%d"
  organization = tfe_organization.foobar.id
}

resource "tfe_project" "foobar" {
  name         = "project-foo-%d"
  organization = tfe_organization.foobar.id
}

resource "tfe_variable_set" "foobar" {
  name = "varset-foo-%d"
	description = "a description"
	organization = tfe_organization.foobar.id
	workspace_ids = [tfe_workspace.foobar.id]
}

resource "tfe_project_variable_set" "foobar" {
	variable_set_id = tfe_variable_set.foobar.id
	project_id = tfe_project.foobar.id
}

resource "tfe_variable" "envfoo" {
	key          = "vfoo"
	value        = "bar"
	category     = "env"
	variable_set_id = tfe_variable_set.foobar.id
}

data "tfe_variable_set" "foobar" {
  name = tfe_variable_set.foobar.name
	organization = tfe_variable_set.foobar.organization
	depends_on = [tfe_variable.envfoo]
}`, rInt, rInt, rInt, rInt)
}
