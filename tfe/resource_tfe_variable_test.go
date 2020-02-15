package tfe

import (
	"fmt"
	"testing"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccTFEVariable_basic(t *testing.T) {
	variable := &tfe.Variable{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEVariableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEVariable_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEVariableExists(
						"tfe_variable.foobar", variable),
					testAccCheckTFEVariableAttributes(variable),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "key", "key_test"),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "value", "value_test"),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "description", "some description"),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "category", "env"),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "hcl", "false"),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "sensitive", "false"),
				),
			},
		},
	})
}

func TestAccTFEVariable_update(t *testing.T) {
	variable := &tfe.Variable{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEVariableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEVariable_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEVariableExists(
						"tfe_variable.foobar", variable),
					testAccCheckTFEVariableAttributes(variable),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "key", "key_test"),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "value", "value_test"),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "description", "some description"),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "category", "env"),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "hcl", "false"),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "sensitive", "false"),
				),
			},

			{
				Config: testAccTFEVariable_update,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEVariableExists(
						"tfe_variable.foobar", variable),
					testAccCheckTFEVariableAttributesUpdate(variable),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "key", "key_updated"),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "value", "value_updated"),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "description", "another description"),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "category", "terraform"),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "hcl", "true"),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "sensitive", "true"),
				),
			},
		},
	})
}

func TestAccTFEVariable_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEVariableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEVariable_basic,
			},

			{
				ResourceName:            "tfe_variable.foobar",
				ImportState:             true,
				ImportStateIdPrefix:     "tst-terraform/workspace-test/",
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"overwrite"},
			},
		},
	})
}

func testAccCheckTFEVariableExists(
	n string, variable *tfe.Variable) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		tfeClient := testAccProvider.Meta().(*tfe.Client)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		wsID := rs.Primary.Attributes["workspace_id"]
		organization, workspace, err := unpackWorkspaceID(wsID)
		if err != nil {
			return fmt.Errorf("Unable to unpack workspace ID: %s", wsID)
		}

		ws, err := tfeClient.Workspaces.Read(ctx, organization, workspace)
		if err != nil {
			return fmt.Errorf("Unable to retreive workspace: %s", err)
		}

		v, err := tfeClient.Variables.Read(ctx, ws.ID, rs.Primary.ID)
		if err != nil {
			return err
		}

		*variable = *v

		return nil
	}
}

func testAccCheckTFEVariableAttributes(
	variable *tfe.Variable) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if variable.Key != "key_test" {
			return fmt.Errorf("Bad key: %s", variable.Key)
		}

		if variable.Value != "value_test" {
			return fmt.Errorf("Bad value: %s", variable.Value)
		}

		if variable.Description != "some description" {
			return fmt.Errorf("Bad description: %s", variable.Description)
		}

		if variable.Category != tfe.CategoryEnv {
			return fmt.Errorf("Bad category: %s", variable.Category)
		}

		if variable.HCL != false {
			return fmt.Errorf("Bad HCL: %t", variable.HCL)
		}

		if variable.Sensitive != false {
			return fmt.Errorf("Bad sensitive: %t", variable.Sensitive)
		}

		return nil
	}
}

func testAccCheckTFEVariableAttributesUpdate(
	variable *tfe.Variable) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if variable.Key != "key_updated" {
			return fmt.Errorf("Bad key: %s", variable.Key)
		}

		if variable.Value != "" {
			return fmt.Errorf("Bad value: %s", variable.Value)
		}

		if variable.Description != "another description" {
			return fmt.Errorf("Bad description: %s", variable.Description)
		}

		if variable.Category != tfe.CategoryTerraform {
			return fmt.Errorf("Bad category: %s", variable.Category)
		}

		if variable.HCL != true {
			return fmt.Errorf("Bad HCL: %t", variable.HCL)
		}

		if variable.Sensitive != true {
			return fmt.Errorf("Bad sensitive: %t", variable.Sensitive)
		}

		return nil
	}
}

func testAccCheckTFEVariableDestroy(s *terraform.State) error {
	tfeClient := testAccProvider.Meta().(*tfe.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_variable" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := tfeClient.Variables.Read(ctx, rs.Primary.Attributes["workspace_id"], rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Variable %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

const testAccTFEVariable_basic = `
resource "tfe_organization" "foobar" {
  name  = "tst-terraform"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = "${tfe_organization.foobar.id}"
}

resource "tfe_variable" "foobar" {
  key          = "key_test"
  value        = "value_test"
  description  = "some description"
  category     = "env"
  workspace_id = "${tfe_workspace.foobar.id}"
}`

const testAccTFEVariable_update = `
resource "tfe_organization" "foobar" {
  name  = "tst-terraform"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = "${tfe_organization.foobar.id}"
}

resource "tfe_variable" "foobar" {
  key          = "key_updated"
  value        = "value_updated"
  description  = "another description"
  category     = "terraform"
  hcl          = true
  sensitive    = true
  workspace_id = "${tfe_workspace.foobar.id}"
}`
