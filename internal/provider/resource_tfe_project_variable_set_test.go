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

func TestAccTFEProjectVariableSet_basic(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	// Make an organization
	org, orgCleanup := createOrganization(t, tfeClient, tfe.OrganizationCreateOptions{
		Name:  tfe.String(fmt.Sprintf("tst-terraform-%d", rInt)),
		Email: tfe.String(fmt.Sprintf("%s@tfe.local", randomString(t))),
	})
	t.Cleanup(orgCleanup)

	// Make a project
	prj := createProject(t, tfeClient, org.Name, tfe.ProjectCreateOptions{
		Name: randomString(t),
	})

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEProjectVariableSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEProjectVariableSet_basic(org.Name, prj.ID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEProjectVariableSetExists(
						"tfe_project_variable_set.test"),
				),
			},
			{
				ResourceName:      "tfe_project_variable_set.test",
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s/%s/variable_set_test", org.Name, prj.ID),
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckTFEProjectVariableSetExists(n string) resource.TestCheckFunc {
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
		prjID, vSID, err := decodeProjectVariableSetID(id)
		if err != nil {
			return fmt.Errorf("error decoding ID (%s): %w", id, err)
		}

		vS, err := config.Client.VariableSets.Read(ctx, vSID, &tfe.VariableSetReadOptions{
			Include: &[]tfe.VariableSetIncludeOpt{tfe.VariableSetProjects},
		})
		if err != nil {
			return fmt.Errorf("error reading variable set %s: %w", vSID, err)
		}
		for _, project := range vS.Projects {
			if project.ID == prjID {
				return nil
			}
		}

		return fmt.Errorf("Project (%s) is not attached to variable set (%s).", prjID, vSID)
	}
}

func testAccCheckTFEProjectVariableSetDestroy(s *terraform.State) error {
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

func testAccTFEProjectVariableSet_base(orgName string) string {
	return fmt.Sprintf(`
resource "tfe_variable_set" "test" {
  name         = "variable_set_test"
  description  = "a test variable set"
  global       = false
  organization = "%s"
}
`, orgName)
}

func testAccTFEProjectVariableSet_basic(orgName string, prjID string) string {
	return testAccTFEProjectVariableSet_base(orgName) + fmt.Sprintf(`
resource "tfe_project_variable_set" "test" {
  variable_set_id = tfe_variable_set.test.id
  project_id      = "%s"
}
`, prjID)
}

func decodeProjectVariableSetID(id string) (string, string, error) {
	idParts := strings.Split(id, "_")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return "", "", fmt.Errorf("expected ID in the form of project-id_variable-set-id, given: %q", id)
	}
	return idParts[0], idParts[1], nil
}
