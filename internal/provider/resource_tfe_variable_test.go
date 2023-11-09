// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
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
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEVariableDestroy,
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
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEVariableDestroy,
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
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEVariableDestroy,
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
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEVariableDestroy,
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

func TestAccTFEVariable_readable_value(t *testing.T) {
	variable := &tfe.Variable{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	variableValue1 := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	variableValue2 := variableValue1 + 42

	// Test that downstream resources may depend on both the value and readableValue of a non-sensitive variable
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEVariableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEVariable_readablevalue(rInt, variableValue1, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEVariableExists(
						"tfe_variable.foobar", variable),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "value", fmt.Sprintf("%d", variableValue1)),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "sensitive", "false"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.downstream-readable", "name", fmt.Sprintf("downstream-readable-%d", variableValue1)),
					resource.TestCheckResourceAttr(
						"tfe_workspace.downstream-nonreadable", "name", fmt.Sprintf("downstream-nonreadable-%d", variableValue1)),
				),
			},
			{
				Config: testAccTFEVariable_readablevalue(rInt, variableValue2, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEVariableExists(
						"tfe_variable.foobar", variable),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "value", fmt.Sprintf("%d", variableValue2)),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "sensitive", "false"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.downstream-readable", "name", fmt.Sprintf("downstream-readable-%d", variableValue2)),
					resource.TestCheckResourceAttr(
						"tfe_workspace.downstream-nonreadable", "name", fmt.Sprintf("downstream-nonreadable-%d", variableValue2)),
				),
			},
		},
	})
}

func TestAccTFEVariable_readable_value_becomes_sensitive(t *testing.T) {
	variable := &tfe.Variable{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	variableValue1 := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	variableValue2 := variableValue1 + 42

	// Test that if an insensitive variable becomes sensitive, downstream resources depending on the readableValue
	// may no longer access it, but that the value may still be used directly
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEVariableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEVariable_readablevalue(rInt, variableValue1, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEVariableExists(
						"tfe_variable.foobar", variable),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "value", fmt.Sprintf("%d", variableValue1)),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "sensitive", "false"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.downstream-readable", "name", fmt.Sprintf("downstream-readable-%d", variableValue1)),
					resource.TestCheckResourceAttr(
						"tfe_workspace.downstream-nonreadable", "name", fmt.Sprintf("downstream-nonreadable-%d", variableValue1)),
				),
			},
			{
				Config:      testAccTFEVariable_readablevalue(rInt, variableValue2, true),
				ExpectError: regexp.MustCompile(`tfe_variable.foobar.readable_value is null`),
			},
		},
	})
}

func TestAccTFEVariable_varset_readable_value(t *testing.T) {
	variable := &tfe.VariableSetVariable{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	variableValue1 := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	variableValue2 := variableValue1 + 42

	// Test that downstream resources may depend on both the value and readableValue of a non-sensitive variable
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEVariableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEVariable_varset_readablevalue(rInt, variableValue1, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEVariableSetVariableExists(
						"tfe_variable.foobar", variable),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "value", fmt.Sprintf("%d", variableValue1)),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "sensitive", "false"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.downstream-readable", "name", fmt.Sprintf("downstream-readable-%d", variableValue1)),
					resource.TestCheckResourceAttr(
						"tfe_workspace.downstream-nonreadable", "name", fmt.Sprintf("downstream-nonreadable-%d", variableValue1)),
				),
			},
			{
				Config: testAccTFEVariable_varset_readablevalue(rInt, variableValue2, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEVariableSetVariableExists(
						"tfe_variable.foobar", variable),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "value", fmt.Sprintf("%d", variableValue2)),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "sensitive", "false"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.downstream-readable", "name", fmt.Sprintf("downstream-readable-%d", variableValue2)),
					resource.TestCheckResourceAttr(
						"tfe_workspace.downstream-nonreadable", "name", fmt.Sprintf("downstream-nonreadable-%d", variableValue2)),
				),
			},
		},
	})
}

func TestAccTFEVariable_varset_readable_value_becomes_sensitive(t *testing.T) {
	variable := &tfe.VariableSetVariable{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	variableValue1 := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	variableValue2 := variableValue1 + 42

	// Test that if an insensitive variable becomes sensitive, downstream resources depending on the readableValue
	// may no longer access it, but that the value may still be used directly
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEVariableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEVariable_varset_readablevalue(rInt, variableValue1, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEVariableSetVariableExists(
						"tfe_variable.foobar", variable),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "value", fmt.Sprintf("%d", variableValue1)),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "sensitive", "false"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.downstream-readable", "name", fmt.Sprintf("downstream-readable-%d", variableValue1)),
					resource.TestCheckResourceAttr(
						"tfe_workspace.downstream-nonreadable", "name", fmt.Sprintf("downstream-nonreadable-%d", variableValue1)),
				),
			},
			{
				Config:      testAccTFEVariable_varset_readablevalue(rInt, variableValue2, true),
				ExpectError: regexp.MustCompile(`tfe_variable.foobar.readable_value is null`),
			},
		},
	})
}

func TestAccTFEVariable_import(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEVariableDestroy,
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

// Verify that the rewritten framework version of the resource results in no
// changes when upgrading from the final sdk v2 version of the resource.
func TestAccTFEVariable_rewrite(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"tfe": {
						VersionConstraint: "0.44.1",
						Source:            "hashicorp/tfe",
					},
				},
				Config: testAccTFEVariable_everything(rInt),
				// leaving Check empty, we just care that they're the same
			},
			{
				ProtoV5ProviderFactories: testAccMuxedProviders,
				Config:                   testAccTFEVariable_everything(rInt),
				PlanOnly:                 true,
			},
		},
	})
}

func testAccCheckTFEVariableExists(
	n string, variable *tfe.Variable) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(ConfiguredClient)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		wsID := rs.Primary.Attributes["workspace_id"]
		ws, err := config.Client.Workspaces.ReadByID(ctx, wsID)
		if err != nil {
			return fmt.Errorf(
				"Error retrieving workspace %s: %w", wsID, err)
		}

		v, err := config.Client.Variables.Read(ctx, ws.ID, rs.Primary.ID)
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
		config := testAccProvider.Meta().(ConfiguredClient)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		vsID := rs.Primary.Attributes["variable_set_id"]
		vs, err := config.Client.VariableSets.Read(ctx, vsID, nil)
		if err != nil {
			return fmt.Errorf(
				"Error retrieving variable set %s: %w", vsID, err)
		}

		v, err := config.Client.VariableSetVariables.Read(ctx, vs.ID, rs.Primary.ID)
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
	config := testAccProvider.Meta().(ConfiguredClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_variable" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := config.Client.Variables.Read(ctx, rs.Primary.Attributes["workspace_id"], rs.Primary.ID)
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
  variable_set_id = tfe_variable_set.foobar.id
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

func testAccTFEVariable_everything(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_variable" "ws_env" {
  key          = "ENV_ONE"
  value        = "value_test"
  description  = "some description"
  category     = "env"
  workspace_id = tfe_workspace.foobar.id
}

resource "tfe_variable" "ws_env_sensitive" {
  key          = "ENV_SENSITIVE"
  value        = "value_test"
  description  = "some description"
  category     = "env"
  sensitive    = true
  workspace_id = tfe_workspace.foobar.id
}

resource "tfe_variable" "ws_terraform" {
  key          = "key_one"
  value        = "value_test"
  description  = "some description"
  category     = "terraform"
  workspace_id = tfe_workspace.foobar.id
}

resource "tfe_variable" "ws_terraform_hcl" {
  key          = "key_hcl"
  value        = "{ map_key = \"value\" }"
  description  = "some description"
  category     = "terraform"
  hcl          = true
  workspace_id = tfe_workspace.foobar.id
}

resource "tfe_variable" "ws_terraform_no_val" {
  key          = "key_no_val"
  # value absent, defaults to empty string
  description  = "some description"
  category     = "terraform"
  workspace_id = tfe_workspace.foobar.id
}

resource "tfe_variable_set" "foobar" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_variable" "vs_env" {
  key          = "ENV_ONE"
  value        = "value_test"
  description  = "other description"
  category     = "env"
  variable_set_id = tfe_variable_set.foobar.id
}

resource "tfe_variable" "vs_env_sensitive" {
  key          = "ENV_TWO"
  value        = "value_test"
  description  = "other description"
  category     = "env"
  sensitive    = true
  variable_set_id = tfe_variable_set.foobar.id
}

resource "tfe_variable" "vs_terraform" {
  key          = "key_whatever"
  value        = "\"hcl string\""
  description  = "other description"
  category     = "terraform"
  hcl          = true
  variable_set_id = tfe_variable_set.foobar.id
}`, rInt)
}
func testAccTFEVariable_readablevalue(rIntOrg int, rIntVariableValue int, sensitive bool) string {
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
  value        = "%d"
  description  = "some description"
  category     = "env"
  workspace_id = tfe_workspace.foobar.id
  sensitive    = %s
}

resource "tfe_workspace" "downstream-readable" {
  name         = "downstream-readable-${tfe_variable.foobar.readable_value}"
  organization = tfe_organization.foobar.id
}

resource "tfe_workspace" "downstream-nonreadable" {
  name         = "downstream-nonreadable-${tfe_variable.foobar.value}"
  organization = tfe_organization.foobar.id
}
`, rIntOrg, rIntVariableValue, strconv.FormatBool(sensitive))
}

func testAccTFEVariable_varset_readablevalue(rIntOrg int, rIntVariableValue int, sensitive bool) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_variable_set" "variable_set" {
  name         = "varset-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_variable" "foobar" {
  key          = "key_test"
  value        = "%d"
  description  = "some description"
  category     = "env"
  variable_set_id = tfe_variable_set.variable_set.id
  sensitive    = %s
}

resource "tfe_workspace" "downstream-readable" {
  name         = "downstream-readable-${tfe_variable.foobar.readable_value}"
  organization = tfe_organization.foobar.id
}

resource "tfe_workspace" "downstream-nonreadable" {
  name         = "downstream-nonreadable-${tfe_variable.foobar.value}"
  organization = tfe_organization.foobar.id
}
`, rIntOrg, rIntVariableValue, strconv.FormatBool(sensitive))
}
