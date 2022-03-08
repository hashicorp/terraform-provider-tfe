package tfe

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccTFEVariable_basic(t *testing.T) {
	variable := &tfe.Variable{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEVariableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEVariable_basic(rInt),
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

func TestAccTFEVariable_basic_variable_set(t *testing.T) {
	variable := &tfe.VariableSetVariable{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEVariableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEVariable_basic_variable_set(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEVariableSetVariableExists(
						"tfe_variable.foobar", variable),
					testAccCheckTFEVariableiSetVariableAttributes(variable),
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
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEVariableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEVariable_basic(rInt),
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
				Config: testAccTFEVariable_update(rInt),
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

func TestAccTFEVariable_update_key_sensitive(t *testing.T) {
	first := &tfe.Variable{}
	second := &tfe.Variable{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEVariableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEVariable_update(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEVariableExists(
						"tfe_variable.foobar", first),
					testAccCheckTFEVariableAttributesUpdate(first),
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
			{
				Config: testAccTFEVariable_update_key_sensitive(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEVariableExists(
						"tfe_variable.foobar", second),
					testAccCheckTFEVariableAttributesUpdate_key_sensitive(second),
					testAccCheckTFEVariableIDsNotEqual(first, second),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "key", "key_updated_2"),
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
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEVariableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEVariable_basic(rInt),
			},

			{
				ResourceName:        "tfe_variable.foobar",
				ImportState:         true,
				ImportStateIdPrefix: fmt.Sprintf("tst-terraform-%d/workspace-test/", rInt),
				ImportStateVerify:   true,
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
		ws, err := tfeClient.Workspaces.ReadByID(ctx, wsID)
		if err != nil {
			return fmt.Errorf(
				"Error retrieving workspace %s: %v", wsID, err)
		}

		v, err := tfeClient.Variables.Read(ctx, ws.ID, rs.Primary.ID)
		if err != nil {
			return err
		}

		*variable = *v

		return nil
	}
}

func testAccCheckTFEVariableSetVariableExists(
	n string, variable *tfe.VariableSetVariable) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		tfeClient := testAccProvider.Meta().(*tfe.Client)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		vsID := rs.Primary.Attributes["variable_set_id"]
		vs, err := tfeClient.VariableSets.Read(ctx, vsID, nil)
		if err != nil {
			return fmt.Errorf(
				"Error retrieving variable set %s: %v", vsID, err)
		}

		v, err := tfeClient.VariableSetVariables.Read(ctx, vs.ID, rs.Primary.ID)
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

func testAccCheckTFEVariableiSetVariableAttributes(
	variable *tfe.VariableSetVariable) resource.TestCheckFunc {
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

func testAccCheckTFEVariableAttributesUpdate_key_sensitive(
	variable *tfe.Variable) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if variable.Key != "key_updated_2" {
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

func testAccCheckTFEVariableIDsNotEqual(
	a, b *tfe.Variable) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if a.ID == b.ID {
			return fmt.Errorf("Variables should not have same ID: %s, %s", a.ID, b.ID)
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

func testAccTFEVariable_basic(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_variable" "foobar" {
  key          = "key_test"
  value        = "value_test"
  description  = "some description"
  category     = "env"
  workspace_id = tfe_workspace.foobar.id
}`, rInt)
}

func testAccTFEVariable_basic_variable_set(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_variable_set" "foobar" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_variable" "foobar" {
  key          = "key_test"
  value        = "value_test"
  description  = "some description"
  category     = "env"
  variable_set_id = tfe_varaible_set.foobar.id
}`, rInt)
}

func testAccTFEVariable_update(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_variable" "foobar" {
  key          = "key_updated"
  value        = "value_updated"
  description  = "another description"
  category     = "terraform"
  hcl          = true
  sensitive    = true
  workspace_id = tfe_workspace.foobar.id
}`, rInt)
}

func testAccTFEVariable_update_key_sensitive(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_variable" "foobar" {
  key          = "key_updated_2"
  value        = "value_updated"
  description  = "another description"
  category     = "terraform"
  hcl          = true
  sensitive    = true
  workspace_id = tfe_workspace.foobar.id
}`, rInt)
}
