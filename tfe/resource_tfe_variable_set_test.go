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

func TestAccTFEVariableSet_basic(t *testing.T) {
	variableSet := &tfe.VariableSet{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEVariableSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEVariableSet_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEVariableSetEExists(
						"tfe_variable_set.foobar", variableSet),
					testAccCheckTFEVariableSetAttributes(variableSet),
					resource.TestCheckResourceAttr(
						"tfe_variable_set.foobar", "name", "foobar"),
					resource.TestCheckResourceAttr(
						"tfe_variable_set.foobar", "description", "a test variable set"),
					resource.TestCheckResourceAttr(
						"tfe_variable_set.foobar", "global", "false"),
				),
			},
		},
	})
}

func TestAccTFEVariableSet_update(t *testing.T) {
	variable := &tfe.VariableSet{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEVariableSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEVariableSet_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEVariableSetEExists(
						"tfe_variable_set.foobar", variableSet),
					testAccCheckTFEVariableSetAttributes(variableSet),
					resource.TestCheckResourceAttr(
						"tfe_variable_set.foobar", "name", "foobar"),
					resource.TestCheckResourceAttr(
						"tfe_variable_set.foobar", "description", "a test variable set"),
					resource.TestCheckResourceAttr(
						"tfe_variable_set.foobar", "global", "false"),
				),
			},

			{
				Config: testAccTFEVariableSet_update(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEVariableSetEExists(
						"tfe_variable_set.foobar", variableSet),
					testAccCheckTFEVariableSetAttributesUpdate(variableSet),
					resource.TestCheckResourceAttr(
						"tfe_variable_set.foobar", "name", "variable_set_test_updated"),
					resource.TestCheckResourceAttr(
						"tfe_variable_set.foobar", "description", "another description"),
					resource.TestCheckResourceAttr(
						"tfe_variable_set.foobar", "global", "true"),
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
		CheckDestroy: testAccCheckTFEVariableSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEVariableSet_basic(rInt),
			},

			{
				ResourceName:        "tfe_variable_set.foobar",
				ImportState:         true,
				ImportStateIdPrefix: fmt.Sprintf("tst-terraform-%d/workspace-test/", rInt),
				ImportStateVerify:   true,
			},
		},
	})
}

func testAccCheckTFEVariableSetExists(
	n string, variableSet *tfe.VariableSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		tfeClient := testAccProvider.Meta().(*tfe.Client)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		vs, err := tfeClient.VariableSets.Read(ctx, rs.Primary.ID, nil)
		if err != nil {
			return err
		}

		*variableSet = *vs

		return nil
	}
}

func testAccCheckTFEVariableSetAttributes(
	variableSet *tfe.VariableSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if variableSet.name != "variable_set_test" {
			return fmt.Errorf("Bad name: %s", variableSet.Name)
		}
		if variableSet.description != "a test variable set" {
			return fmt.Errorf("Bad description: %s", variableSet.Description)
		}
		if variableSet.global != "false" {
			return fmt.Errorf("Bad global: %s", variableSet.Global)
		}
	}
}

func testAccCheckTFEVariableSetAttributesUpdate(
	variableSet *tfe.VariableSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if variableSet.name != "variable_set_test_updated" {
			return fmt.Errorf("Bad name: %s", variableSet.Name)
		}
		if variableSet.description != "another description" {
			return fmt.Errorf("Bad description: %s", variableSet.Description)
		}
		if variableSet.global != "true" {
			return fmt.Errorf("Bad global: %s", variableSet.Global)
		}
	}
}

func testAccCheckTFEVariableSetDestroy(rInt int) string {
	tfeClient := testAccProvider.Meta().(*tfe.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_variable_set" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := tfeClient.VariableSets.Read(ctx, rs.Primary.ID, nil)
		if err == nil {
			return fmt.Errorf("Variable Set %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccTFEVariableSet_basic(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name = "tft-terraform-%d"
	email = "admin@company.com"
}

resource "tfe_variable_set" "foobar" {
  name         = "variable_set_test"
	description  = "a test variable set"
	global       = false
	organizaiton = tfe_organizatoin.foobar.id
}`, rInt)
}

func testAccTFEVariable_update(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_variable_set" "foobar" {
  name         = "variable_set_test_updated"
	description  = "another description"
	global       = true
	organizaiton = tfe_organizatoin.foobar.id
}`, rInt)
}
