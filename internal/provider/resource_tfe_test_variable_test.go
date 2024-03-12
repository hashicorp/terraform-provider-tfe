// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccTFEVariable_testingvariables(t *testing.T) {
	variable := &tfe.Variable{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEVariableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETestVariable_test_variable(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETestVariableExists(
						"tfe_test_variable.foobar", variable),
					testAccCheckTFEVariableAttributes(variable),
					resource.TestCheckResourceAttr(
						"tfe_test_variable.foobar", "key", "key_test"),
					resource.TestCheckResourceAttr(
						"tfe_test_variable.foobar", "value", "value_test"),
					resource.TestCheckResourceAttr(
						"tfe_test_variable.foobar", "description", "some description"),
					resource.TestCheckResourceAttr(
						"tfe_test_variable.foobar", "category", "env"),
					resource.TestCheckResourceAttr(
						"tfe_test_variable.foobar", "hcl", "false"),
					resource.TestCheckResourceAttr(
						"tfe_test_variable.foobar", "sensitive", "false"),
				),
			},
		},
	})
}

func testAccCheckTFETestVariableExists(
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
		moduleID := tfe.RegistryModuleID{
			Organization: rs.Primary.Attributes["organization"],
			Name:         rs.Primary.Attributes["module_name"],
			Provider:     rs.Primary.Attributes["module_provider"],
			Namespace:    rs.Primary.Attributes["module_namespace"],
			RegistryName: "private",
		}

		v, err := config.Client.TestVariables.Read(ctx, moduleID, rs.Primary.ID)
		if err != nil {
			return err
		}

		*variable = *v

		return nil
	}
}

func testAccTFETestVariable_test_variable(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}
  
resource "tfe_oauth_client" "foobar" {
  organization     = tfe_organization.foobar.name
  api_url          = "https://api.github.com"
  http_url         = "https://github.com"
  oauth_token      = "%s"
  service_provider = "github"
}

resource "tfe_registry_module" "foobar" {
  organization     = tfe_organization.foobar.name
  vcs_repo {
  display_identifier = "%s"
  identifier         = "%s"
  oauth_token_id     = tfe_oauth_client.foobar.oauth_token_id
  branch             = "main"
  tags				 = false
}
  test_config {
	tests_enabled = true
  }
}

resource "tfe_test_variable" "foobar" {
  key          = "key_test"
  value        = "value_test"
  description  = "some description"
  category     = "env"
  organization = tfe_organization.foobar.name
  module_name = tfe_registry_module.foobar.name
  module_provider = tfe_registry_module.foobar.module_provider
}
`,
		rInt,
		envGithubToken,
		envGithubRegistryModuleIdentifer,
		envGithubRegistryModuleIdentifer)
}
