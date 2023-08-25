// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccTFEWorkspaceVariableSet_basic(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEWorkspaceVariableSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEWorkspaceVariableSet_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceVariableSetExists(
						"tfe_workspace_variable_set.test"),
				),
			},
			{
				ResourceName:      "tfe_workspace_variable_set.test",
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("tst-terraform-%d/tst-terraform-%d/variable_set_test", rInt, rInt),
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckTFEWorkspaceVariableSetExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(ConfiguredClient)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		id := rs.Primary.ID

		if id == "" {
			return fmt.Errorf("No ID is set")
		}
		wID, vSID, err := decodeWorkspaceVariableSetID(id)
		if err != nil {
			return fmt.Errorf("error decoding ID (%s): %w", id, err)
		}

		vS, err := config.Client.VariableSets.Read(ctx, vSID, &tfe.VariableSetReadOptions{
			Include: &[]tfe.VariableSetIncludeOpt{tfe.VariableSetWorkspaces},
		})
		if err != nil {
			return fmt.Errorf("error reading variable set %s: %w", vSID, err)
		}
		for _, workspace := range vS.Workspaces {
			if workspace.ID == wID {
				return nil
			}
		}

		return fmt.Errorf("Workspace (%s) is not attached to variable set (%s).", wID, vSID)
	}
}

func testAccCheckTFEWorkspaceVariableSetDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(ConfiguredClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_variable_set" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := config.Client.VariableSets.Read(ctx, rs.Primary.ID, nil)
		if err == nil {
			return fmt.Errorf("Variable Set %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccTFEWorkspaceVariableSet_base(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "test" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "test" {
  name         = "tst-terraform-%d"
  organization = tfe_organization.test.id
  auto_apply   = true
  tag_names    = ["test"]
}

resource "tfe_variable_set" "test" {
  name         = "variable_set_test"
  description  = "a test variable set"
  global       = false
  organization = tfe_organization.test.id
}
`, rInt, rInt)
}

func testAccTFEWorkspaceVariableSet_basic(rInt int) string {
	return testAccTFEWorkspaceVariableSet_base(rInt) + `
resource "tfe_workspace_variable_set" "test" {
  variable_set_id = tfe_variable_set.test.id
  workspace_id    = tfe_workspace.test.id
}`
}

func decodeWorkspaceVariableSetID(id string) (string, string, error) {
	idParts := strings.Split(id, "_")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return "", "", fmt.Errorf("expected ID in the form of workspace-id_variable-set-id, given: %q", id)
	}
	return idParts[0], idParts[1], nil
}
