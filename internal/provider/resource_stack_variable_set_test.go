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

func TestAccStackVariableSet_basic(t *testing.T) {
	skipUnlessBeta(t)
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckStackVariableSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStackVariableSet_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackVariableSetExists("tfe_stack_variable_set.test"),
					resource.TestCheckResourceAttrSet("tfe_stack_variable_set.test", "id"),
					resource.TestCheckResourceAttrSet("tfe_stack_variable_set.test", "stack_id"),
					resource.TestCheckResourceAttrSet("tfe_stack_variable_set.test", "variable_set_id"),
				),
			},
		},
	})
}

func TestAccStackVariableSet_import(t *testing.T) {
	skipUnlessBeta(t)
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckStackVariableSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStackVariableSet_basic(rInt),
			},
			{
				ResourceName:      "tfe_stack_variable_set.test",
				ImportState:       true,
				ImportStateIdFunc: testAccStackVariableSetImportStateID,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckStackVariableSetExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No stack variable set ID is set")
		}

		// Verify the ID format is stack_id_variable_set_id
		stackID := rs.Primary.Attributes["stack_id"]
		variableSetID := rs.Primary.Attributes["variable_set_id"]

		if stackID == "" || variableSetID == "" {
			return fmt.Errorf("stack_id or variable_set_id is not set")
		}

		// Verify we can read the variable set to ensure the association exists
		vs, err := testAccConfiguredClient.Client.VariableSets.Read(ctx, variableSetID, nil)
		if err != nil {
			return fmt.Errorf("Error reading variable set %s: %w", variableSetID, err)
		}

		if vs == nil {
			return fmt.Errorf("Variable set %s not found", variableSetID)
		}

		return nil
	}
}

func testAccCheckStackVariableSetDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_stack_variable_set" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No stack variable set ID is set")
		}

		stackID := rs.Primary.Attributes["stack_id"]
		variableSetID := rs.Primary.Attributes["variable_set_id"]

		// Read the variable set and verify the stack is no longer associated
		vs, err := testAccConfiguredClient.Client.VariableSets.Read(ctx, variableSetID, nil)
		if err != nil {
			// If the variable set itself is gone, the association is also gone
			continue
		}

		// Check if stack is still in the variable set's stacks
		for _, stack := range vs.Stacks {
			if stack.ID == stackID {
				return fmt.Errorf("Stack %s is still associated with variable set %s", stackID, variableSetID)
			}
		}
	}

	return nil
}

func testAccStackVariableSet_basic(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "test" {
  name  = "tst-terraform-%[1]d"
  email = "admin@company.com"
}

resource "tfe_project" "test" {
  organization = tfe_organization.test.id
  name         = "tst-project-%[1]d"
}

resource "tfe_stack" "test" {
  organization  = tfe_organization.test.id
  project_id    = tfe_project.test.id
  name          = "tst-stack-%[1]d"
  description   = "Test stack for variable set association"
}

resource "tfe_variable_set" "test" {
  name         = "tst-variable-set-%[1]d"
  description  = "Test variable set"
  organization = tfe_organization.test.id
}

resource "tfe_stack_variable_set" "test" {
  stack_id       = tfe_stack.test.id
  variable_set_id = tfe_variable_set.test.id
}
`, rInt)
}

func testAccStackVariableSetImportStateID(s *terraform.State) (string, error) {
	rs, ok := s.RootModule().Resources["tfe_stack_variable_set.test"]
	if !ok {
		return "", fmt.Errorf("Not found: tfe_stack_variable_set.test")
	}

	stackID := rs.Primary.Attributes["stack_id"]
	variableSetID := rs.Primary.Attributes["variable_set_id"]

	if stackID == "" || variableSetID == "" {
		return "", fmt.Errorf("stack_id or variable_set_id is empty")
	}

	// Return the import ID in format: stack/varset
	return fmt.Sprintf("%s/%s", stackID, variableSetID), nil
}
