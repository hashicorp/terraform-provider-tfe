// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tfe

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccTFENoCodeModule_basic(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}
	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	nocodeModule := &tfe.RegistryNoCodeModule{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFENoCodeModuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFENoCodeModule_basic(org),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFENoCodeModuleExists(
						"tfe_no_code_module.foobar", nocodeModule),
					resource.TestCheckResourceAttr(
						"tfe_no_code_module.foobar", "enabled", "true"),
					resource.TestCheckResourceAttr(
						"tfe_no_code_module.foobar", "organization", org.Name),
				),
			},
		},
	})
}

func TestAccTFENoCodeModule_with_variable_options(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}
	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	nocodeModule := &tfe.RegistryNoCodeModule{}

	regionOptions := `"us-east-1", "us-west-1", "eu-west-2"`

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFENoCodeModuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFENoCodeModule_with_options(org, regionOptions),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFENoCodeModuleExists(
						"tfe_no_code_module.foobar", nocodeModule),
					resource.TestCheckResourceAttr(
						"tfe_no_code_module.foobar", "enabled", "true"),
					resource.TestCheckResourceAttr(
						"tfe_no_code_module.foobar", "organization", org.Name),
					testAccCheckTFENoCodeModuleVariableOptions(nocodeModule),
				),
			},
		},
	})
}

func TestAccTFENoCodeModule_update(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}
	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	nocodeModule := &tfe.RegistryNoCodeModule{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFENoCodeModuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFENoCodeModule_basic(org),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFENoCodeModuleExists(
						"tfe_no_code_module.foobar", nocodeModule),
					resource.TestCheckResourceAttr(
						"tfe_no_code_module.foobar", "enabled", "true"),
				),
			},
			{
				Config: testAccTFENoCodeModule_update(org),
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
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}
	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	nocodeModule := &tfe.RegistryNoCodeModule{}
	regionOptions := `"us-east-1", "us-west-1", "eu-west-2"`
	updatedRegionOptions := `"eu-east-1", "eu-west-1", "us-west-2"`

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFENoCodeModuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFENoCodeModule_with_options(org, regionOptions),
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
				Config: testAccTFENoCodeModule_with_options(org, updatedRegionOptions),
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
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}
	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	nocodeModule := &tfe.RegistryNoCodeModule{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFENoCodeModuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFENoCodeModule_basic(org),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFENoCodeModuleExists(
						"tfe_no_code_module.foobar", nocodeModule),
					resource.TestCheckResourceAttr(
						"tfe_no_code_module.foobar", "enabled", "true"),
				),
			},
			{
				Config: testAccTFENoCodeModule_update(org),
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
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}
	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	nocodeModule := &tfe.RegistryNoCodeModule{}
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFENoCodeModuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFENoCodeModule_basic(org),
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

func testAccTFENoCodeModule_update(org *tfe.Organization) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
	name  = "%s"
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
}`, org.Name)
}

func testAccTFENoCodeModule_basic(org *tfe.Organization) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
name  = "%s"
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
`, org.Name)
}

func testAccTFENoCodeModule_with_options(org *tfe.Organization, regionOpts string) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
name  = "%s"
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
`, org.Name, regionOpts)
}

func testAccCheckTFENoCodeModuleDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(ConfiguredClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_no_code_module" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := config.Client.RegistryNoCodeModules.Read(ctx, rs.Primary.ID, nil)
		if err == nil {
			return fmt.Errorf("Project %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckTFENoCodeModuleExists(n string, nocodeModule *tfe.RegistryNoCodeModule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(ConfiguredClient)

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
		p, err := config.Client.RegistryNoCodeModules.Read(ctx, rs.Primary.ID, opts)
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
