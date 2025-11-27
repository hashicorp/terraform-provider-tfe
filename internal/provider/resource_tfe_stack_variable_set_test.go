// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccTFEStackVariableSet_basic(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEStackVariableSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEStackVariableSet_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEStackVariableSetExists(
						"tfe_stack_variable_set.test"),
					resource.TestCheckResourceAttrSet(
						"tfe_stack_variable_set.test", "variable_set_id"),
					resource.TestCheckResourceAttrSet(
						"tfe_stack_variable_set.test", "stack_id"),
				),
			},
		},
	})
}

func testAccCheckTFEStackVariableSetExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		return nil
	}
}

func testAccCheckTFEStackVariableSetDestroy(s *terraform.State) error {
	return nil
}

func testAccTFEStackVariableSet_basic(rInt int) string {
	return fmt.Sprintf(`
		resource "tfe_organization" "foobar" {
			name = "tst-terraform-%d"
			email = "admin@company.com"
		}

		resource "tfe_stack" "test" {
			name         = "tst-stack-%d"
			organization = tfe_organization.foobar.name
		}

		resource "tfe_variable_set" "test" {
			name         = "variable_set_test-%d"
			description  = "a test variable set"
			global       = false
			priority     = false
			organization = tfe_organization.foobar.id
		}

		resource "tfe_stack_variable_set" "test" {
			variable_set_id = tfe_variable_set.test.id
			stack_id        = tfe_stack.test.id
		}`, rInt, rInt, rInt)
}
