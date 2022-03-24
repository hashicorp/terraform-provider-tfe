package tfe

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
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFERegistryModuleExists(
						"tfe_registry_module.foobar", orgName, registryModule),
					testAccCheckTFERegistryModuleAttributes(registryModule, orgName),
					resource.TestCheckResourceAttr(
						"tfe_registry_module.foobar", "organization", orgName),
					resource.TestCheckResourceAttr(
						"tfe_registry_module.foobar", "name", getRegistryModuleName()),
					resource.TestCheckResourceAttr(
						"tfe_registry_module.foobar", "module_provider", getRegistryModuleProvider()),
					resource.TestCheckResourceAttr(
						"tfe_registry_module.foobar", "vcs_repo.0.display_identifier", GITHUB_REGISTRY_MODULE_IDENTIFIER),
					resource.TestCheckResourceAttr(
						"tfe_registry_module.foobar", "vcs_repo.0.identifier", GITHUB_REGISTRY_MODULE_IDENTIFIER),
					resource.TestCheckResourceAttrSet(
						"tfe_registry_module.foobar", "vcs_repo.0.oauth_token_id"),
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
				Config:      testAccTFERegistryModule_emptyVCSRepo(rInt, GITHUB_TOKEN),
				ExpectError: regexp.MustCompile(`Missing required argument`),
			},
		},
	})
}

func TestAccTFERegistryModuleImport(t *testing.T) {
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

func testAccCheckTFERegistryModuleExists(n, orgName string, registryModule *tfe.RegistryModule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		tfeClient := testAccProvider.Meta().(*tfe.Client)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		rmID := tfe.RegistryModuleID{
			Organization: orgName,
			Name:         getRegistryModuleName(),
			Provider:     getRegistryModuleProvider(),
		}

		rm, err := tfeClient.RegistryModules.Read(ctx, rmID)
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

func testAccCheckTFERegistryModuleAttributes(registryModule *tfe.RegistryModule, orgName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if registryModule.Name != getRegistryModuleName() {
			return fmt.Errorf("Bad name: %s", registryModule.Name)
		}

		if registryModule.Provider != getRegistryModuleProvider() {
			return fmt.Errorf("Bad module_provider: %s", registryModule.Provider)
		}

		if registryModule.Organization.Name != orgName {
			return fmt.Errorf("Bad organization: %v", registryModule.Organization.Name)
		}

		if registryModule.VCSRepo == nil {
			return fmt.Errorf("Bad VCS repo: %v", registryModule.VCSRepo)
		}

		if registryModule.VCSRepo.DisplayIdentifier != GITHUB_REGISTRY_MODULE_IDENTIFIER {
			return fmt.Errorf("Bad VCS repo display identifier: %v", registryModule.VCSRepo.DisplayIdentifier)
		}

		if registryModule.VCSRepo.Identifier != GITHUB_REGISTRY_MODULE_IDENTIFIER {
			return fmt.Errorf("Bad VCS repo identifier: %v", registryModule.VCSRepo.Identifier)
		}

		if registryModule.VCSRepo.OAuthTokenID == "" {
			return fmt.Errorf("Bad VCS repo oauth token id: %v", registryModule.VCSRepo)
		}

		return nil
	}
}

func testAccCheckTFERegistryModuleDestroy(s *terraform.State) error {
	tfeClient := testAccProvider.Meta().(*tfe.Client)

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

		rmID := tfe.RegistryModuleID{
			Organization: organization,
			Name:         name,
			Provider:     module_provider,
		}
		_, err := tfeClient.RegistryModules.Read(ctx, rmID)
		if err == nil {
			return fmt.Errorf("Registry module %s still exists", id)
		}
	}

	return nil
}

func testAccPreCheckTFERegistryModule(t *testing.T) {
	if GITHUB_TOKEN == "" {
		t.Skip("Please set GITHUB_TOKEN to run this test")
	}
	if GITHUB_REGISTRY_MODULE_IDENTIFIER == "" {
		t.Skip("Please set GITHUB_REGISTRY_MODULE_IDENTIFIER to run this test")
	}
}

func getRegistryModuleRepository() string {
	if GITHUB_REGISTRY_MODULE_IDENTIFIER == "" {
		return GITHUB_REGISTRY_MODULE_IDENTIFIER
	}
	return strings.Split(GITHUB_REGISTRY_MODULE_IDENTIFIER, "/")[1]
}
func getRegistryModuleName() string {
	if GITHUB_REGISTRY_MODULE_IDENTIFIER == "" {
		return GITHUB_REGISTRY_MODULE_IDENTIFIER
	}
	return strings.SplitN(getRegistryModuleRepository(), "-", 3)[2]
}

func getRegistryModuleProvider() string {
	if GITHUB_REGISTRY_MODULE_IDENTIFIER == "" {
		return GITHUB_REGISTRY_MODULE_IDENTIFIER
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
		GITHUB_TOKEN,
		GITHUB_REGISTRY_MODULE_IDENTIFIER,
		GITHUB_REGISTRY_MODULE_IDENTIFIER)
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
