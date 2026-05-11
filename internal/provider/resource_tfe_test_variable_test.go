// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"math/rand"
	"regexp"
	"testing"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func TestAccTFETestVariable_basic(t *testing.T) {
	variable := &tfe.Variable{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccGithubPreCheck(t)
		},
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFETestVariableDestroy,
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

func TestAccTFETestVariable_valueWriteOnly(t *testing.T) {
	variable := &tfe.Variable{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	variableValue1 := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	variableValue2 := variableValue1 + 42
	versionOne, versionTwo := 1, 2

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccGithubPreCheck(t)
		},
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(version.Must(version.NewVersion("1.11.0"))),
		},
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFETestVariableDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFETestVariable_valueAndValueWO(rInt, variableValue1),
				ExpectError: regexp.MustCompile(`Attribute "value" cannot be specified when "value_wo" is specified`),
			},
			{
				Config:      testAccTFETestVariable_valueWOOnly(rInt, variableValue1),
				ExpectError: regexp.MustCompile(`Attribute "value_wo_version" must be specified when "value_wo" is specified`),
			},
			{
				Config:      testAccTFETestVariable_versionOnly(rInt),
				ExpectError: regexp.MustCompile(`Attribute "value_wo" must be specified when "value_wo_version" is specified`),
			},
			{
				Config:      testAccTFETestVariable_valueWithVersion(rInt),
				ExpectError: regexp.MustCompile(`Attribute "value" cannot be specified when "value_wo_version" is specified`),
			},
			{
				Config: testAccTFETestVariable_valueWriteOnly(rInt, variableValue1, versionOne),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETestVariableExists(
						"tfe_test_variable.foobar", variable),
					resource.TestCheckNoResourceAttr(
						"tfe_test_variable.foobar", "value_wo"),
					resource.TestCheckResourceAttr(
						"tfe_test_variable.foobar", "value_wo_version", fmt.Sprintf("%d", versionOne)),
					resource.TestCheckResourceAttr(
						"tfe_test_variable.foobar", "sensitive", "false"),
				),
			},
			{
				Config: testAccTFETestVariable_valueWriteOnly(rInt, variableValue2, versionTwo),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETestVariableExists(
						"tfe_test_variable.foobar", variable),
					resource.TestCheckNoResourceAttr(
						"tfe_test_variable.foobar", "value_wo"),
					resource.TestCheckResourceAttr(
						"tfe_test_variable.foobar", "value_wo_version", fmt.Sprintf("%d", versionTwo)),
					resource.TestCheckResourceAttr(
						"tfe_test_variable.foobar", "sensitive", "false"),
				),
			},
		},
	})
}

func TestAccTFETestVariable_updateValueWriteOnlyToValue(t *testing.T) {
	variable := &tfe.Variable{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	variableValue := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	versionOne := 1

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccGithubPreCheck(t)
		},
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(version.Must(version.NewVersion("1.11.0"))),
		},
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFETestVariableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETestVariable_valueWriteOnly(rInt, variableValue, versionOne),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETestVariableExists(
						"tfe_test_variable.foobar", variable),
					resource.TestCheckResourceAttr(
						"tfe_test_variable.foobar", "value", ""),
					resource.TestCheckNoResourceAttr(
						"tfe_test_variable.foobar", "value_wo"), // write-only values are never persisted in state
					resource.TestCheckResourceAttr(
						"tfe_test_variable.foobar", "value_wo_version", fmt.Sprintf("%d", versionOne)),
				),
			},
			{
				Config: testAccTFETestVariable_test_variable(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETestVariableExists(
						"tfe_test_variable.foobar", variable),
					resource.TestCheckResourceAttr(
						"tfe_test_variable.foobar", "value", "value_test"),
					resource.TestCheckNoResourceAttr(
						"tfe_test_variable.foobar", "value_wo"),
					resource.TestCheckNoResourceAttr(
						"tfe_test_variable.foobar", "value_wo_version"),
				),
			},
		},
	})
}

func TestAccTFETestVariable_update(t *testing.T) {
	variable := &tfe.Variable{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccGithubPreCheck(t)
		},
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFETestVariableDestroy,
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

			{
				Config: testAccTFETestVariable_update(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETestVariableExists(
						"tfe_test_variable.foobar", variable),
					testAccCheckTFETestVariableAttributesUpdate(variable),
					resource.TestCheckResourceAttr(
						"tfe_test_variable.foobar", "key", "key_updated"),
					resource.TestCheckResourceAttr(
						"tfe_test_variable.foobar", "value", "value_updated"),
					resource.TestCheckResourceAttr(
						"tfe_test_variable.foobar", "description", "another description"),
					resource.TestCheckResourceAttr(
						"tfe_test_variable.foobar", "category", "env"),
					resource.TestCheckResourceAttr(
						"tfe_test_variable.foobar", "hcl", "true"),
					resource.TestCheckResourceAttr(
						"tfe_test_variable.foobar", "sensitive", "true"),
				),
			},
		},
	})
}

func testAccCheckTFETestVariableExists(
	n string, variable *tfe.Variable) resource.TestCheckFunc {
	return func(s *terraform.State) error {
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
			Namespace:    rs.Primary.Attributes["organization"],
			RegistryName: "private",
		}

		v, err := testAccConfiguredClient.Client.TestVariables.Read(ctx, moduleID, rs.Primary.ID)
		if err != nil {
			return err
		}

		*variable = *v

		return nil
	}
}

func testAccCheckTFETestVariableDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_test_variable" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		moduleID := tfe.RegistryModuleID{
			Organization: rs.Primary.Attributes["organization"],
			Name:         rs.Primary.Attributes["module_name"],
			Provider:     rs.Primary.Attributes["module_provider"],
			Namespace:    rs.Primary.Attributes["organization"],
			RegistryName: "private",
		}

		_, err := testAccConfiguredClient.Client.TestVariables.Read(ctx, moduleID, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Variable %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckTFETestVariableAttributesUpdate(
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

		if variable.Category != tfe.CategoryEnv {
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

func testAccTFETestVariable_valueWriteOnly(rInt int, rIntVariableValue int, valueVersion int) string {
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
  key              = "key_test"
  value_wo         = "%d"
  value_wo_version = %d
  description      = "some description"
  category         = "env"
  organization     = tfe_organization.foobar.name
  module_name      = tfe_registry_module.foobar.name
  module_provider  = tfe_registry_module.foobar.module_provider
}
`, rInt,
		envGithubToken,
		envGithubRegistryModuleIdentifer,
		envGithubRegistryModuleIdentifer,
		rIntVariableValue,
		valueVersion)
}

func testAccTFETestVariable_valueAndValueWO(rInt int, rIntVariableValue int) string {
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
  key              = "key_test"
  value            = "%d"
  value_wo         = "%d"
  value_wo_version = 1
  description      = "some description"
  category         = "env"
  organization     = tfe_organization.foobar.name
  module_name      = tfe_registry_module.foobar.name
  module_provider  = tfe_registry_module.foobar.module_provider
}
`, rInt,
		envGithubToken,
		envGithubRegistryModuleIdentifer,
		envGithubRegistryModuleIdentifer,
		rIntVariableValue,
		rIntVariableValue)
}

func testAccTFETestVariable_valueWOOnly(rInt int, rIntVariableValue int) string {
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
  key             = "key_test"
  value_wo        = "%d"
  description     = "some description"
  category        = "env"
  organization    = tfe_organization.foobar.name
  module_name     = tfe_registry_module.foobar.name
  module_provider = tfe_registry_module.foobar.module_provider
}
`, rInt,
		envGithubToken,
		envGithubRegistryModuleIdentifer,
		envGithubRegistryModuleIdentifer,
		rIntVariableValue)
}

func testAccTFETestVariable_versionOnly(rInt int) string {
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
  key              = "key_test"
  value_wo_version = 1
  description      = "some description"
  category         = "env"
  organization     = tfe_organization.foobar.name
  module_name      = tfe_registry_module.foobar.name
  module_provider  = tfe_registry_module.foobar.module_provider
}
`, rInt,
		envGithubToken,
		envGithubRegistryModuleIdentifer,
		envGithubRegistryModuleIdentifer)
}

func testAccTFETestVariable_valueWithVersion(rInt int) string {
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
  key              = "key_test"
  value            = "value_test"
  value_wo_version = 1
  description      = "some description"
  category         = "env"
  organization     = tfe_organization.foobar.name
  module_name      = tfe_registry_module.foobar.name
  module_provider  = tfe_registry_module.foobar.module_provider
}
`, rInt,
		envGithubToken,
		envGithubRegistryModuleIdentifer,
		envGithubRegistryModuleIdentifer)
}

func testAccTFETestVariable_update(rInt int) string {
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
  key          = "key_updated"
  value        = "value_updated"
  description  = "another description"
  category     = "env"
  hcl          = true
  sensitive    = true
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
