// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccTFERegistryModule_vcs(t *testing.T) {
	registryModule := &tfe.RegistryModule{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	expectedRegistryModuleAttributes := &tfe.RegistryModule{
		Name:         getRegistryModuleName(),
		Provider:     getRegistryModuleProvider(),
		RegistryName: tfe.PrivateRegistry,
		Namespace:    orgName,
		Organization: &tfe.Organization{Name: orgName},
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTFERegistryModule(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFERegistryModuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFERegistryModule_vcs(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFERegistryModuleExists(
						"tfe_registry_module.foobar",
						tfe.RegistryModuleID{
							Organization: orgName,
							Name:         expectedRegistryModuleAttributes.Name,
							Provider:     expectedRegistryModuleAttributes.Provider,
							RegistryName: expectedRegistryModuleAttributes.RegistryName,
							Namespace:    orgName,
						}, registryModule),
					testAccCheckTFERegistryModuleAttributes(registryModule, expectedRegistryModuleAttributes),
					testAccCheckTFERegistryModuleVCSAttributes(registryModule),
					resource.TestCheckResourceAttr(
						"tfe_registry_module.foobar", "organization", orgName),
					resource.TestCheckResourceAttr(
						"tfe_registry_module.foobar", "name", expectedRegistryModuleAttributes.Name),
					resource.TestCheckResourceAttr(
						"tfe_registry_module.foobar", "module_provider", expectedRegistryModuleAttributes.Provider),
					resource.TestCheckResourceAttr(
						"tfe_registry_module.foobar", "namespace", expectedRegistryModuleAttributes.Namespace),
					resource.TestCheckResourceAttr(
						"tfe_registry_module.foobar", "registry_name", string(expectedRegistryModuleAttributes.RegistryName)),
					resource.TestCheckResourceAttr(
						"tfe_registry_module.foobar", "vcs_repo.0.display_identifier", envGithubRegistryModuleIdentifer),
					resource.TestCheckResourceAttr(
						"tfe_registry_module.foobar", "vcs_repo.0.identifier", envGithubRegistryModuleIdentifer),
					resource.TestCheckResourceAttrSet(
						"tfe_registry_module.foobar", "vcs_repo.0.oauth_token_id"),
				),
			},
		},
	})
}

func TestAccTFERegistryModule_GitHubApp(t *testing.T) {
	registryModule := &tfe.RegistryModule{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	expectedRegistryModuleAttributes := &tfe.RegistryModule{
		Name:         getRegistryModuleName(),
		Provider:     getRegistryModuleProvider(),
		RegistryName: tfe.PrivateRegistry,
		Namespace:    orgName,
		Organization: &tfe.Organization{Name: orgName},
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTFERegistryModule(t)
			testAccGHAInstallationPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFERegistryModuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFERegistryModule_GitHubApp(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFERegistryModuleExists(
						"tfe_registry_module.foobar",
						tfe.RegistryModuleID{
							Organization: orgName,
							Name:         expectedRegistryModuleAttributes.Name,
							Provider:     expectedRegistryModuleAttributes.Provider,
							RegistryName: expectedRegistryModuleAttributes.RegistryName,
							Namespace:    orgName,
						}, registryModule),
					testAccCheckTFERegistryModuleAttributes(registryModule, expectedRegistryModuleAttributes),
					testAccCheckTFERegistryModuleVCSAttributes(registryModule),
					resource.TestCheckResourceAttr(
						"tfe_registry_module.foobar", "organization", orgName),
					resource.TestCheckResourceAttr(
						"tfe_registry_module.foobar", "name", expectedRegistryModuleAttributes.Name),
					resource.TestCheckResourceAttr(
						"tfe_registry_module.foobar", "module_provider", expectedRegistryModuleAttributes.Provider),
					resource.TestCheckResourceAttr(
						"tfe_registry_module.foobar", "namespace", expectedRegistryModuleAttributes.Namespace),
					resource.TestCheckResourceAttr(
						"tfe_registry_module.foobar", "registry_name", string(expectedRegistryModuleAttributes.RegistryName)),
					resource.TestCheckResourceAttr(
						"tfe_registry_module.foobar", "vcs_repo.0.display_identifier", envGithubRegistryModuleIdentifer),
					resource.TestCheckResourceAttr(
						"tfe_registry_module.foobar", "vcs_repo.0.identifier", envGithubRegistryModuleIdentifer),
					resource.TestCheckResourceAttrSet(
						"tfe_registry_module.foobar", "vcs_repo.0.github_app_installation_id"),
				),
			},
		},
	})
}

func TestAccTFERegistryModule_emptyVCSRepo(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckTFERegistryModule(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFERegistryModuleDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFERegistryModule_emptyVCSRepo(rInt, envGithubToken),
				ExpectError: regexp.MustCompile(`Missing required argument`),
			},
		},
	})
}

func TestAccTFERegistryModule_nonVCSPrivateRegistryModuleWithoutRegistryName(t *testing.T) {
	registryModule := &tfe.RegistryModule{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	expectedRegistryModuleAttributes := &tfe.RegistryModule{
		Name:         "test_module",
		Provider:     "my_provider",
		RegistryName: tfe.PrivateRegistry,
		Namespace:    orgName,
		Organization: &tfe.Organization{Name: orgName},
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFERegistryModuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFERegistryModule_privateRMWithoutRegistryName(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFERegistryModuleExists(
						"tfe_registry_module.foobar",
						tfe.RegistryModuleID{
							Organization: orgName,
							Name:         expectedRegistryModuleAttributes.Name,
							Provider:     expectedRegistryModuleAttributes.Provider,
							RegistryName: expectedRegistryModuleAttributes.RegistryName,
							Namespace:    expectedRegistryModuleAttributes.Namespace,
						}, registryModule),
					testAccCheckTFERegistryModuleAttributes(registryModule, expectedRegistryModuleAttributes),
					resource.TestCheckResourceAttr(
						"tfe_registry_module.foobar", "organization", orgName),
					resource.TestCheckResourceAttr(
						"tfe_registry_module.foobar", "name", expectedRegistryModuleAttributes.Name),
					resource.TestCheckResourceAttr(
						"tfe_registry_module.foobar", "module_provider", expectedRegistryModuleAttributes.Provider),
					resource.TestCheckResourceAttr(
						"tfe_registry_module.foobar", "namespace", expectedRegistryModuleAttributes.Namespace),
					resource.TestCheckResourceAttr(
						"tfe_registry_module.foobar", "registry_name", string(expectedRegistryModuleAttributes.RegistryName)),
				),
			},
		},
	})
}

func TestAccTFERegistryModule_nonVCSPrivateRegistryModuleWithRegistryName(t *testing.T) {
	registryModule := &tfe.RegistryModule{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	expectedRegistryModuleAttributes := &tfe.RegistryModule{
		Name:         "another_test_module",
		Provider:     "my_provider",
		RegistryName: tfe.PrivateRegistry,
		Namespace:    orgName,
		Organization: &tfe.Organization{Name: orgName},
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFERegistryModuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFERegistryModule_privateRMWithRegistryName(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFERegistryModuleExists(
						"tfe_registry_module.foobar",
						tfe.RegistryModuleID{
							Organization: orgName,
							Name:         expectedRegistryModuleAttributes.Name,
							Provider:     expectedRegistryModuleAttributes.Provider,
							RegistryName: expectedRegistryModuleAttributes.RegistryName,
							Namespace:    expectedRegistryModuleAttributes.Namespace,
						}, registryModule),
					testAccCheckTFERegistryModuleAttributes(registryModule, expectedRegistryModuleAttributes),
					resource.TestCheckResourceAttr(
						"tfe_registry_module.foobar", "organization", orgName),
					resource.TestCheckResourceAttr(
						"tfe_registry_module.foobar", "name", expectedRegistryModuleAttributes.Name),
					resource.TestCheckResourceAttr(
						"tfe_registry_module.foobar", "module_provider", expectedRegistryModuleAttributes.Provider),
					resource.TestCheckResourceAttr(
						"tfe_registry_module.foobar", "namespace", expectedRegistryModuleAttributes.Namespace),
					resource.TestCheckResourceAttr(
						"tfe_registry_module.foobar", "registry_name", string(expectedRegistryModuleAttributes.RegistryName)),
				),
			},
		},
	})
}

func TestAccTFERegistryModule_publicRegistryModule(t *testing.T) {
	registryModule := &tfe.RegistryModule{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	expectedRegistryModuleAttributes := &tfe.RegistryModule{
		Name:         "vpc",
		Provider:     "aws",
		RegistryName: tfe.PublicRegistry,
		Namespace:    "terraform-aws-modules",
		Organization: &tfe.Organization{Name: orgName},
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFERegistryModuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFERegistryModule_publicRM(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFERegistryModuleExists(
						"tfe_registry_module.foobar",
						tfe.RegistryModuleID{
							Organization: orgName,
							Name:         expectedRegistryModuleAttributes.Name,
							Provider:     expectedRegistryModuleAttributes.Provider,
							RegistryName: expectedRegistryModuleAttributes.RegistryName,
							Namespace:    expectedRegistryModuleAttributes.Namespace,
						}, registryModule),
					testAccCheckTFERegistryModuleAttributes(registryModule, expectedRegistryModuleAttributes),
					resource.TestCheckResourceAttr(
						"tfe_registry_module.foobar", "organization", orgName),
					resource.TestCheckResourceAttr(
						"tfe_registry_module.foobar", "name", expectedRegistryModuleAttributes.Name),
					resource.TestCheckResourceAttr(
						"tfe_registry_module.foobar", "module_provider", expectedRegistryModuleAttributes.Provider),
					resource.TestCheckResourceAttr(
						"tfe_registry_module.foobar", "namespace", expectedRegistryModuleAttributes.Namespace),
					resource.TestCheckResourceAttr(
						"tfe_registry_module.foobar", "registry_name", string(expectedRegistryModuleAttributes.RegistryName)),
				),
			},
		},
	})
}

func TestAccTFERegistryModule_noCodeModule(t *testing.T) {
	skipIfEnterprise(t)

	registryModule := &tfe.RegistryModule{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	expectedRegistryModuleAttributes := &tfe.RegistryModule{
		Name:         "vpc",
		Provider:     "aws",
		RegistryName: tfe.PublicRegistry,
		Namespace:    "terraform-aws-modules",
		Organization: &tfe.Organization{Name: orgName},
		NoCode:       true,
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFERegistryModuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFERegistryModule_NoCode(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFERegistryModuleExists(
						"tfe_registry_module.foobar",
						tfe.RegistryModuleID{
							Organization: orgName,
							Name:         expectedRegistryModuleAttributes.Name,
							Provider:     expectedRegistryModuleAttributes.Provider,
							RegistryName: expectedRegistryModuleAttributes.RegistryName,
							Namespace:    expectedRegistryModuleAttributes.Namespace,
						}, registryModule),
					testAccCheckTFERegistryModuleAttributes(registryModule, expectedRegistryModuleAttributes),
					resource.TestCheckResourceAttr(
						"tfe_registry_module.foobar", "organization", orgName),
					resource.TestCheckResourceAttr(
						"tfe_registry_module.foobar", "name", expectedRegistryModuleAttributes.Name),
					resource.TestCheckResourceAttr(
						"tfe_registry_module.foobar", "module_provider", expectedRegistryModuleAttributes.Provider),
					resource.TestCheckResourceAttr(
						"tfe_registry_module.foobar", "namespace", expectedRegistryModuleAttributes.Namespace),
					resource.TestCheckResourceAttr(
						"tfe_registry_module.foobar", "registry_name", string(expectedRegistryModuleAttributes.RegistryName)),
					resource.TestCheckResourceAttr(
						"tfe_registry_module.foobar", "no_code", fmt.Sprint(expectedRegistryModuleAttributes.NoCode)),
				),
			},
		},
	})
}

func TestAccTFERegistryModuleImport_vcsPrivateRMDeprecatedFormat(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckTFERegistryModule(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFERegistryModuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFERegistryModule_vcs(rInt),
			},
			{
				ResourceName:        "tfe_registry_module.foobar",
				ImportState:         true,
				ImportStateIdPrefix: fmt.Sprintf("tst-terraform-%d/%v/%v/", rInt, getRegistryModuleName(), getRegistryModuleProvider()),
				ImportStateVerify:   true,
			},
		},
	})
}

func TestAccTFERegistryModuleImport_vcsPrivateRMRecommendedFormat(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckTFERegistryModule(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFERegistryModuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFERegistryModule_vcs(rInt),
			},
			{
				ResourceName:        "tfe_registry_module.foobar",
				ImportState:         true,
				ImportStateIdPrefix: fmt.Sprintf("tst-terraform-%d/%v/tst-terraform-%d/%v/%v/", rInt, "private", rInt, getRegistryModuleName(), getRegistryModuleProvider()),
				ImportStateVerify:   true,
			},
		},
	})
}

func TestAccTFERegistryModuleImport_nonVCSPrivateRM(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFERegistryModuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFERegistryModule_privateRMWithRegistryName(rInt),
			},
			{
				ResourceName:        "tfe_registry_module.foobar",
				ImportState:         true,
				ImportStateIdPrefix: fmt.Sprintf("tst-terraform-%d/%v/tst-terraform-%d/%v/%v/", rInt, "private", rInt, "another_test_module", "my_provider"),
				ImportStateVerify:   true,
			},
		},
	})
}

func TestAccTFERegistryModuleImport_publicRM(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFERegistryModuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFERegistryModule_publicRM(rInt),
			},
			{
				ResourceName:        "tfe_registry_module.foobar",
				ImportState:         true,
				ImportStateIdPrefix: fmt.Sprintf("tst-terraform-%d/%v/%v/%v/%v/", rInt, "public", "terraform-aws-modules", "vpc", "aws"),
				ImportStateVerify:   true,
			},
		},
	})
}

func TestAccTFERegistryModule_invalidWithBothVCSRepoAndModuleProvider(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFERegistryModule_invalidWithBothVCSRepoAndModuleProvider(),
				ExpectError: regexp.MustCompile("\"module_provider\": only one of `module_provider,vcs_repo` can be specified,\nbut `module_provider,vcs_repo` were specified."),
			},
		},
	})
}

func TestAccTFERegistryModule_invalidRegistryName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFERegistryModule_invalidRegistryName(),
				ExpectError: regexp.MustCompile(`invalid value for registry-name. It must be either "private" or "public"`),
			},
		},
	})
}

func TestAccTFERegistryModule_invalidWithModuleProviderAndNoName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFERegistryModule_invalidWithModuleProviderAndNoName(),
				ExpectError: regexp.MustCompile("\"module_provider\": all of `module_provider,name,organization` must be\nspecified"),
			},
		},
	})
}

func TestAccTFERegistryModule_invalidWithModuleProviderAndNoOrganization(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFERegistryModule_invalidWithModuleProviderAndNoOrganization(),
				ExpectError: regexp.MustCompile("\"module_provider\": all of `module_provider,name,organization` must be\nspecified"),
			},
		},
	})
}

func TestAccTFERegistryModule_invalidWithNamespaceAndNoRegistryName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFERegistryModule_invalidWithNamespaceAndNoRegistryName(),
				ExpectError: regexp.MustCompile("\"namespace\": all of `namespace,registry_name` must be specified"),
			},
		},
	})
}

func TestAccTFERegistryModule_invalidWithRegistryNameAndNoModuleProvider(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFERegistryModule_invalidWithRegistryNameAndNoModuleProvider(),
				ExpectError: regexp.MustCompile("\"registry_name\": all of `module_provider,registry_name` must be specified"),
			},
		},
	})
}
func testAccCheckTFERegistryModuleExists(n string, rmID tfe.RegistryModuleID, registryModule *tfe.RegistryModule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(ConfiguredClient)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		rm, err := config.Client.RegistryModules.Read(ctx, rmID)
		if err != nil {
			return err
		}

		if rm.ID != rs.Primary.ID {
			return fmt.Errorf("Not found: %s", n)
		}

		*registryModule = *rm

		return nil
	}
}

func testAccCheckTFERegistryModuleAttributes(registryModule *tfe.RegistryModule, expectedRegistryModuleAttributes *tfe.RegistryModule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if registryModule.Name != expectedRegistryModuleAttributes.Name {
			return fmt.Errorf("Bad name: %s", registryModule.Name)
		}

		if registryModule.Provider != expectedRegistryModuleAttributes.Provider {
			return fmt.Errorf("Bad module_provider: %s", registryModule.Provider)
		}

		if registryModule.Organization.Name != expectedRegistryModuleAttributes.Organization.Name {
			return fmt.Errorf("Bad organization: %v", registryModule.Organization.Name)
		}

		if registryModule.RegistryName != expectedRegistryModuleAttributes.RegistryName {
			return fmt.Errorf("Bad registry_name: %v", registryModule.RegistryName)
		}

		if registryModule.Namespace != expectedRegistryModuleAttributes.Namespace {
			return fmt.Errorf("Bad namespace: %v", registryModule.Namespace)
		}

		return nil
	}
}

func testAccCheckTFERegistryModuleVCSAttributes(registryModule *tfe.RegistryModule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if registryModule.VCSRepo == nil {
			return fmt.Errorf("Bad VCS repo: %v", registryModule.VCSRepo)
		}

		if registryModule.VCSRepo.DisplayIdentifier != envGithubRegistryModuleIdentifer {
			return fmt.Errorf("Bad VCS repo display identifier: %v", registryModule.VCSRepo.DisplayIdentifier)
		}

		if registryModule.VCSRepo.Identifier != envGithubRegistryModuleIdentifer {
			return fmt.Errorf("Bad VCS repo identifier: %v", registryModule.VCSRepo.Identifier)
		}

		switch registryModule.VCSRepo.ServiceProvider {
		case "github_app":
			if registryModule.VCSRepo.GHAInstallationID == "" {
				return fmt.Errorf("Bad VCS repo github app installation id: %v", registryModule.VCSRepo)
			}
		default:
			if registryModule.VCSRepo.OAuthTokenID == "" {
				return fmt.Errorf("Bad VCS repo oauth token id: %v", registryModule.VCSRepo)
			}
		}

		return nil
	}
}

func testAccCheckTFERegistryModuleDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(ConfiguredClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_registry_module" {
			continue
		}

		id := rs.Primary.ID
		if id == "" {
			return fmt.Errorf("No instance ID is set")
		}

		organization := rs.Primary.Attributes["organization"]
		if organization == "" {
			return fmt.Errorf("No organization is set for registry module %s", id)
		}

		name := rs.Primary.Attributes["name"]
		if name == "" {
			return fmt.Errorf("No name is set for registry module %s", id)
		}

		module_provider := rs.Primary.Attributes["module_provider"]
		if module_provider == "" {
			return fmt.Errorf("No module_provider is set for registry module %s", id)
		}

		namespace := rs.Primary.Attributes["namespace"]
		if namespace == "" {
			return fmt.Errorf("No namespace is set for registry module %s", id)
		}

		registry_name := rs.Primary.Attributes["registry_name"]
		if registry_name == "" {
			return fmt.Errorf("No registry_name is set for registry module %s", id)
		}

		rmID := tfe.RegistryModuleID{
			Organization: organization,
			Name:         name,
			Provider:     module_provider,
			Namespace:    rs.Primary.Attributes["namespace"],
			RegistryName: tfe.RegistryName(rs.Primary.Attributes["registry_name"]),
		}
		_, err := config.Client.RegistryModules.Read(ctx, rmID)
		if err == nil {
			return fmt.Errorf("Registry module %s still exists", id)
		}
	}

	return nil
}

func testAccPreCheckTFERegistryModule(t *testing.T) {
	if envGithubToken == "" {
		t.Skip("Please set GITHUB_TOKEN to run this test")
	}
	if envGithubRegistryModuleIdentifer == "" {
		t.Skip("Please set GITHUB_REGISTRY_MODULE_IDENTIFIER to run this test")
	}
}

func getRegistryModuleRepository() string {
	if envGithubRegistryModuleIdentifer == "" {
		return envGithubRegistryModuleIdentifer
	}
	return strings.Split(envGithubRegistryModuleIdentifer, "/")[1]
}
func getRegistryModuleName() string {
	if envGithubRegistryModuleIdentifer == "" {
		return envGithubRegistryModuleIdentifer
	}
	return strings.SplitN(getRegistryModuleRepository(), "-", 3)[2]
}

func getRegistryModuleProvider() string {
	if envGithubRegistryModuleIdentifer == "" {
		return envGithubRegistryModuleIdentifer
	}
	return strings.SplitN(getRegistryModuleRepository(), "-", 3)[1]
}

func testAccTFERegistryModule_vcs(rInt int) string {
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
 vcs_repo {
   display_identifier = "%s"
   identifier         = "%s"
   oauth_token_id     = tfe_oauth_client.foobar.oauth_token_id
 }
}`,
		rInt,
		envGithubToken,
		envGithubRegistryModuleIdentifer,
		envGithubRegistryModuleIdentifer)
}

func testAccTFERegistryModule_GitHubApp(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
 name  = "tst-terraform-%d"
 email = "admin@company.com"
}

resource "tfe_registry_module" "foobar" {
 organization    = tfe_organization.foobar.id
 vcs_repo {
   display_identifier = "%s"
   identifier         = "%s"
   github_app_installation_id = "%s"
 }
}`,
		rInt,
		envGithubRegistryModuleIdentifer,
		envGithubRegistryModuleIdentifer,
		envGithubAppInstallationID)
}

func testAccTFERegistryModule_emptyVCSRepo(rInt int, token string) string {
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
 vcs_repo {}
}`, rInt, token)
}

func testAccTFERegistryModule_privateRMWithoutRegistryName(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
 name  = "tst-terraform-%d"
 email = "admin@company.com"
}

resource "tfe_registry_module" "foobar" {
	organization    = tfe_organization.foobar.id
  module_provider = "my_provider"
  name            = "test_module"
 }`,
		rInt)
}

func testAccTFERegistryModule_privateRMWithRegistryName(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
 name  = "tst-terraform-%d"
 email = "admin@company.com"
}

resource "tfe_registry_module" "foobar" {
	organization    = tfe_organization.foobar.id
  module_provider = "my_provider"
  name            = "another_test_module"
  registry_name   = "private"
 }`,
		rInt)
}

func testAccTFERegistryModule_publicRM(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
 name  = "tst-terraform-%d"
 email = "admin@company.com"
}

resource "tfe_registry_module" "foobar" {
  organization    = tfe_organization.foobar.id
  namespace       = "terraform-aws-modules"
  module_provider = "aws"
  name            = "vpc"
  registry_name   = "public"
 }`,
		rInt)
}

func testAccTFERegistryModule_NoCode(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
 name  = "tst-terraform-%d"
 email = "admin@company.com"
}

resource "tfe_registry_module" "foobar" {
  organization    = tfe_organization.foobar.id
  namespace       = "terraform-aws-modules"
  module_provider = "aws"
  name            = "vpc"
  registry_name   = "public"
  no_code         = true
 }`,
		rInt)
}

func testAccTFERegistryModule_invalidWithBothVCSRepoAndModuleProvider() string {
	return `
resource "tfe_registry_module" "foobar" {
  module_provider = "aws"
	vcs_repo {
		display_identifier = "hashicorp/terraform-random-module"
		identifier         = "hashicorp/terraform-random-module"
		oauth_token_id     = "sample-auth-token"
	}
 }`
}

func testAccTFERegistryModule_invalidRegistryName() string {
	return `
resource "tfe_registry_module" "foobar" {
  organization    = "hashicorp"
  module_provider = "aws"
  name            = "eks"
  registry_name   = "PRIVATE"
 }`
}

func testAccTFERegistryModule_invalidWithModuleProviderAndNoName() string {
	return `
resource "tfe_registry_module" "foobar" {
  organization    = "hashicorp"
  module_provider = "aws"
  registry_name   = "private"
 }`
}

func testAccTFERegistryModule_invalidWithModuleProviderAndNoOrganization() string {
	return `
resource "tfe_registry_module" "foobar" {
  name            = "eks"
  module_provider = "aws"
  registry_name   = "private"
 }`
}

func testAccTFERegistryModule_invalidWithNamespaceAndNoRegistryName() string {
	return `
resource "tfe_registry_module" "foobar" {
  organization    = "hashicorp"
  module_provider = "aws"
  name            = "eks"
  namespace       = "terraform-aws-modules"
 }`
}

func testAccTFERegistryModule_invalidWithRegistryNameAndNoModuleProvider() string {
	return `
resource "tfe_registry_module" "foobar" {
  organization    = "hashicorp"
  name            = "eks"
  namespace       = "terraform-aws-modules"
	registry_name   = "private"
 }`
}
