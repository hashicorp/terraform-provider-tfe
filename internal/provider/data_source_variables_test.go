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

func TestAccTFEVariablesDataSource_basic(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEVariablesDataSourceConfig_basic(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					// variables attribute
					resource.TestCheckResourceAttrSet("data.tfe_variables.workspace_foobar", "id"),
					resource.TestCheckResourceAttrSet("data.tfe_variables.variable_set_foobar", "id"),
					resource.TestCheckOutput("workspace_variables", "foo"),
					resource.TestCheckOutput("workspace_env", "foo"),
					resource.TestCheckOutput("workspace_terraform", "foo"),
					resource.TestCheckOutput("variable_set_variables", "vfoo"),
					resource.TestCheckOutput("variable_set_env", "vfoo"),
					resource.TestCheckOutput("variable_set_terraform", "vfoo"),
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

resource "tfe_variable_set" "foobar" {
  name = "varset-foo-%d"
	organization = tfe_organization.foobar.id
}

resource "tfe_variable" "terrbar" {
	key          = "foo"
	value        = "bar"
	category     = "terraform"
	workspace_id = tfe_workspace.foobar.id
}

resource "tfe_variable" "envbar" {
	key          = "foo"
	value        = "bar"
	category     = "env"
	workspace_id = tfe_workspace.foobar.id
}

resource "tfe_variable" "terrfoo" {
	key          = "vfoo"
	value        = "bar"
	category     = "terraform"
	variable_set_id = tfe_variable_set.foobar.id
}

resource "tfe_variable" "envfoo" {
	key          = "vfoo"
	value        = "bar"
	category     = "env"
	variable_set_id = tfe_variable_set.foobar.id
}

data "tfe_variables" "workspace_foobar" {
	workspace_id = tfe_workspace.foobar.id
	depends_on = [
		tfe_variable.terrbar,
		tfe_variable.envbar
  ]
}

data "tfe_variables" "variable_set_foobar" {
	variable_set_id = tfe_variable_set.foobar.id
	depends_on = [
		tfe_variable.terrfoo,
		tfe_variable.envfoo
  ]
}

output "workspace_variables" {
	value = data.tfe_variables.workspace_foobar.variables[0]["name"]
}

output "workspace_env" {
	value = data.tfe_variables.workspace_foobar.env[0]["name"]
}

output "workspace_terraform" {
	value = data.tfe_variables.workspace_foobar.terraform[0]["name"]
}

output "variable_set_variables" {
	value = data.tfe_variables.variable_set_foobar.variables[0]["name"]
}

output "variable_set_env" {
	value = data.tfe_variables.variable_set_foobar.terraform[0]["name"]
}

output "variable_set_terraform" {
	value = data.tfe_variables.variable_set_foobar.env[0]["name"]
}`, rInt, rInt, rInt)
}
