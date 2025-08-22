// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/assert"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTFENoCodeModule_basic(t *testing.T) {
	skipUnlessBeta(t)
	nocodeModule := &tfe.RegistryNoCodeModule{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFENoCodeModuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFENoCodeModule_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFENoCodeModuleExists(
						"tfe_no_code_module.foobar", nocodeModule),
					resource.TestCheckResourceAttr(
						"tfe_no_code_module.foobar", "enabled", "true"),
					resource.TestCheckResourceAttr(
						"tfe_no_code_module.foobar", "organization", fmt.Sprintf("tst-terraform-%d", rInt)),
				),
			},
		},
	})
}

func TestAccTFENoCodeModule_with_variable_options(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatalf("error getting client %v", err)
	}
	org, cleanup := createBusinessOrganization(t, tfeClient)
	defer cleanup()
	providers := muxedProvidersWithDefaultOrganization(org.Name)
	cfg := testAccTFENoCodeModule_with_variable_options(org.Name)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: providers,
		CheckDestroy:             testAccCheckTFENoCodeModuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: cfg,
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						n := "tfe_no_code_module.sensitive"
						rs, ok := s.RootModule().Resources[n]
						if !ok || rs.Primary.ID == "" {
							return fmt.Errorf("Not found: %s", n)
						}

						opts := &tfe.RegistryNoCodeModuleReadOptions{
							Include: []tfe.RegistryNoCodeModuleIncludeOpt{tfe.RegistryNoCodeIncludeVariableOptions},
						}
						nocodeModule, err := testAccConfiguredClient.Client.RegistryNoCodeModules.Read(ctx, rs.Primary.ID, opts)
						if err != nil {
							return fmt.Errorf("unable to read nocodeModule with ID %s", rs.Primary.ID)
						}

						if !nocodeModule.Enabled {
							return fmt.Errorf("Bad 'enabled' attribute: %t", nocodeModule.Enabled)
						}

						if len(nocodeModule.VariableOptions) == 0 {
							return fmt.Errorf("Bad 'variable_options' attribute: %v", nocodeModule.VariableOptions)
						}

						for _, vo := range nocodeModule.VariableOptions {
							if vo.VariableName == "min_lower" {
								if len(vo.Options) != 5 {
									return fmt.Errorf("Bad 'min_lower' attribute options: %v", nocodeModule.VariableOptions)
								}
							}
						}

						return nil
					},
				),
			},
		},
	})
}

func TestAccTFENoCodeModule_with_version_pin(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatalf("error getting client %v", err)
	}
	org, cleanup := createBusinessOrganization(t, tfeClient)
	defer cleanup()
	providers := muxedProvidersWithDefaultOrganization(org.Name)
	cfg := testAccTFENoCodeModule_with_version_pin(org.Name)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: providers,
		CheckDestroy:             testAccCheckTFENoCodeModuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: cfg,
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						n := "tfe_no_code_module.sensitive"
						rs, ok := s.RootModule().Resources[n]
						if !ok || rs.Primary.ID == "" {
							return fmt.Errorf("Not found: %s", n)
						}

						opts := &tfe.RegistryNoCodeModuleReadOptions{
							Include: []tfe.RegistryNoCodeModuleIncludeOpt{tfe.RegistryNoCodeIncludeVariableOptions},
						}
						nocodeModule, err := testAccConfiguredClient.Client.RegistryNoCodeModules.Read(ctx, rs.Primary.ID, opts)
						if err != nil {
							return fmt.Errorf("unable to read nocodeModule with ID %s", rs.Primary.ID)
						}

						if !nocodeModule.Enabled {
							return fmt.Errorf("Bad 'enabled' attribute: %t", nocodeModule.Enabled)
						}

						if nocodeModule.VersionPin != "1.1.0" {
							return fmt.Errorf("Bad 'version_pin' attribute: %s", nocodeModule.VersionPin)
						}

						return nil
					},
				),
			},
		},
	})
}

func TestAccTFENoCodeModule_update(t *testing.T) {
	skipUnlessBeta(t)
	nocodeModule := &tfe.RegistryNoCodeModule{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFENoCodeModuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFENoCodeModule_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFENoCodeModuleExists(
						"tfe_no_code_module.foobar", nocodeModule),
					resource.TestCheckResourceAttr(
						"tfe_no_code_module.foobar", "enabled", "true"),
				),
			},
			{
				Config: testAccTFENoCodeModule_update(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFENoCodeModuleExists(
						"tfe_no_code_module.foobar", nocodeModule),
					resource.TestCheckResourceAttr(
						"tfe_no_code_module.foobar", "enabled", "true"),
				),
			},
		},
	})
}

func TestAccTFENoCodeModule_update_variable_options(t *testing.T) {
	skipUnlessBeta(t)
	nocodeModule := &tfe.RegistryNoCodeModule{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	regionOptions := `"us-east-1", "us-west-1", "eu-west-2"`
	updatedRegionOptions := `"eu-east-1", "eu-west-1", "us-west-2"`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFENoCodeModuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFENoCodeModule_with_options(rInt, regionOptions),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFENoCodeModuleExists(
						"tfe_no_code_module.foobar", nocodeModule),
					resource.TestCheckResourceAttr(
						"tfe_no_code_module.foobar", "enabled", "true"),
					func(s *terraform.State) error {
						if len(nocodeModule.VariableOptions) == 0 {
							return fmt.Errorf("Bad 'variable_options' attribute: %v", nocodeModule.VariableOptions)
						}

						for _, vo := range nocodeModule.VariableOptions {
							if vo.VariableName == "region" {
								assert.ElementsMatch(t, []string{"us-east-1", "us-west-1", "eu-west-2"}, vo.Options)
							}
						}
						return nil
					},
				),
			},
			{
				Config: testAccTFENoCodeModule_with_options(rInt, updatedRegionOptions),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFENoCodeModuleExists(
						"tfe_no_code_module.foobar", nocodeModule),
					resource.TestCheckResourceAttr(
						"tfe_no_code_module.foobar", "enabled", "true"),
					func(s *terraform.State) error {
						if len(nocodeModule.VariableOptions) == 0 {
							return fmt.Errorf("Bad 'variable_options' attribute: %v", nocodeModule.VariableOptions)
						}

						for _, vo := range nocodeModule.VariableOptions {
							if vo.VariableName == "region" {
								assert.ElementsMatch(t, []string{"eu-east-1", "eu-west-1", "us-west-2"}, vo.Options)
							}
						}
						return nil
					},
				),
			},
		},
	})
}

func TestAccTFENoCodeModule_delete(t *testing.T) {
	skipUnlessBeta(t)
	nocodeModule := &tfe.RegistryNoCodeModule{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFENoCodeModuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFENoCodeModule_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFENoCodeModuleExists(
						"tfe_no_code_module.foobar", nocodeModule),
					resource.TestCheckResourceAttr(
						"tfe_no_code_module.foobar", "enabled", "true"),
				),
			},
			{
				Config: testAccTFENoCodeModule_update(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFENoCodeModuleExists(
						"tfe_no_code_module.foobar", nocodeModule),
					resource.TestCheckResourceAttr(
						"tfe_no_code_module.foobar", "enabled", "true"),
				),
			},
		},
	})
}

func TestAccTFENoCodeModule_import(t *testing.T) {
	skipUnlessBeta(t)
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	nocodeModule := &tfe.RegistryNoCodeModule{}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFENoCodeModuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFENoCodeModule_basic(rInt),
				Check: testAccCheckTFENoCodeModuleExists(
					"tfe_no_code_module.foobar", nocodeModule),
			},

			{
				ResourceName:      "tfe_no_code_module.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "tfe_no_code_module.foobar",
				ImportState:       true,
				ImportStateId:     nocodeModule.ID,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccTFENoCodeModule_update(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
	name  = "tst-terraform-%d"
	email = "admin@company.com"
}
	
resource "tfe_registry_module" "foobar" {
	organization    = tfe_organization.foobar.id
	module_provider = "my_provider"
	name            = "test_module"
}
	
resource "tfe_no_code_module" "foobar" {
	organization = tfe_organization.foobar.id
	registry_module = tfe_registry_module.foobar.id
}`, rInt)
}

func testAccTFENoCodeModule_basic(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
name  = "tst-terraform-%d"
email = "admin@company.com"
}

resource "tfe_registry_module" "foobar" {
	organization    = tfe_organization.foobar.id
	module_provider = "my_provider"
	name            = "test_module"
}

resource "tfe_no_code_module" "foobar" {
	organization = tfe_organization.foobar.id
	registry_module = tfe_registry_module.foobar.id
}
`, rInt)
}

func testAccTFENoCodeModule_with_options(rInt int, regionOpts string) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
name  = "tst-terraform-%d"
email = "admin@company.com"
}

resource "tfe_registry_module" "foobar" {
	organization    = tfe_organization.foobar.id
	module_provider = "my_provider"
	name            = "test_module"
}

resource "tfe_no_code_module" "foobar" {
	organization = tfe_organization.foobar.id
	registry_module = tfe_registry_module.foobar.id

	variable_options {
		name    = "ami"
		type    = "string"
		options = [ "ami-0", "ami-1", "ami-2" ]
	}

	variable_options {
		name    = "region"
		type    = "string"
		options = [%s]
	}
}
`, rInt, regionOpts)
}

func testAccCheckTFENoCodeModuleDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_no_code_module" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := testAccConfiguredClient.Client.RegistryNoCodeModules.Read(ctx, rs.Primary.ID, nil)
		if err == nil {
			return fmt.Errorf("Project %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckTFENoCodeModuleExists(n string, nocodeModule *tfe.RegistryNoCodeModule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		opts := &tfe.RegistryNoCodeModuleReadOptions{
			Include: []tfe.RegistryNoCodeModuleIncludeOpt{tfe.RegistryNoCodeIncludeVariableOptions},
		}
		p, err := testAccConfiguredClient.Client.RegistryNoCodeModules.Read(ctx, rs.Primary.ID, opts)
		if err != nil {
			return fmt.Errorf("unable to read nocodeModule with ID %s", nocodeModule.ID)
		}

		*nocodeModule = *p

		return nil
	}
}

func testAccCheckTFENoCodeModuleVariableOptions(
	nocodeModule *tfe.RegistryNoCodeModule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !nocodeModule.Enabled {
			return fmt.Errorf("Bad 'enabled' attribute: %t", nocodeModule.Enabled)
		}

		if len(nocodeModule.VariableOptions) == 0 {
			return fmt.Errorf("Bad 'variable_options' attribute: %v", nocodeModule.VariableOptions)
		}

		for _, vo := range nocodeModule.VariableOptions {
			if vo.VariableName == "region" {
				if len(vo.Options) != 3 {
					return fmt.Errorf("Bad 'variable_options' attribute: %v", nocodeModule.VariableOptions)
				}
			}
		}

		return nil
	}
}

func testAccTFENoCodeModule_with_variable_options(org string) string {
	return fmt.Sprintf(`
	locals {
		organization_name = "%s"
		identifier         = "%s"
	}

	resource "tfe_oauth_client" "github" {
		organization     = local.organization_name
		api_url          = "https://api.github.com"
		http_url         = "https://github.com"
		oauth_token      = "%s"
		service_provider = "github"
	}

	resource "tfe_registry_module" "sensitive" {
		vcs_repo {
			display_identifier = local.identifier
			identifier         = local.identifier
			oauth_token_id     = tfe_oauth_client.github.oauth_token_id
		}
	}

	resource "tfe_no_code_module" "sensitive" {
	organization    = local.organization_name
	registry_module = tfe_registry_module.sensitive.id
	version_pin     = "1.1.0"

	variable_options {
			name    = "min_lower"
			type    = "number"
			options = [ "1", "2", "3", "4", "5" ]
	}
}
`, org, envGithubRegistryModuleIdentifer, envGithubToken)
}

func testAccTFENoCodeModule_with_version_pin(org string) string {
	return fmt.Sprintf(`
	locals {
		organization_name = "%s"
		identifier         = "%s"
	}

	resource "tfe_oauth_client" "github" {
		organization     = local.organization_name
		api_url          = "https://api.github.com"
		http_url         = "https://github.com"
		oauth_token      = "%s"
		service_provider = "github"
	}

	resource "tfe_registry_module" "sensitive" {
		vcs_repo {
			display_identifier = local.identifier
			identifier         = local.identifier
			oauth_token_id     = tfe_oauth_client.github.oauth_token_id
		}
	}

	resource "tfe_no_code_module" "sensitive" {
	organization    = local.organization_name
	registry_module = tfe_registry_module.sensitive.id
	version_pin     = "1.1.0"
}
`, org, envGithubRegistryModuleIdentifer, envGithubToken)
}
