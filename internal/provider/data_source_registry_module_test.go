// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTFERegistryModuleDataSource_basicPrivate(t *testing.T) {
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
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFERegistryModuleDataSourceBasic(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.tfe_registry_module.test", "organization", orgName),
					resource.TestCheckResourceAttr("data.tfe_registry_module.test", "namespace", orgName),
					resource.TestCheckResourceAttr("data.tfe_registry_module.test", "name", expectedRegistryModuleAttributes.Name),
					resource.TestCheckResourceAttr("data.tfe_registry_module.test", "registry_name", string(expectedRegistryModuleAttributes.RegistryName)),
					resource.TestCheckResourceAttr("data.tfe_registry_module.test", "module_provider", expectedRegistryModuleAttributes.Provider),
					resource.TestCheckResourceAttr("data.tfe_registry_module.test", "no_code", "false"),
					resource.TestCheckResourceAttr("data.tfe_registry_module.test", "permissions.0.can_delete", "true"),
					resource.TestCheckResourceAttr("data.tfe_registry_module.test", "permissions.0.can_retry", "true"),
					resource.TestCheckResourceAttr("data.tfe_registry_module.test", "permissions.0.can_resync", "true"),
					resource.TestCheckResourceAttr("data.tfe_registry_module.test", "publishing_mechanism", "git_tag"),
					resource.TestCheckResourceAttr("data.tfe_registry_module.test", "status", "pending"),
					resource.TestCheckResourceAttr("data.tfe_registry_module.test", "vcs_repo.0.display_identifier", envGithubRegistryModuleIdentifer),
					resource.TestCheckResourceAttr("data.tfe_registry_module.test", "vcs_repo.0.identifier", envGithubRegistryModuleIdentifer),
					resource.TestCheckResourceAttrSet("data.tfe_registry_module.test", "vcs_repo.0.oauth_token_id"),
					resource.TestCheckResourceAttr("data.tfe_registry_module.test", "vcs_repo.0.branch", ""),
					resource.TestCheckResourceAttr("data.tfe_registry_module.test", "vcs_repo.0.tags", "true"),
				),
			},
		},
	})
}

func TestAccTFERegistryModuleDataSource_basicNoCodePublic(t *testing.T) {
	skipIfEnterprise(t)

	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)
	dsName := "data.tfe_registry_module.test"
	rsName := "tfe_no_code_module.foobar"

	expectedRegistryModuleAttributes := &tfe.RegistryModule{
		Name:         "vpc",
		Provider:     "aws",
		RegistryName: tfe.PublicRegistry,
		Namespace:    "terraform-aws-modules",
		Organization: &tfe.Organization{Name: orgName},
	}

	ncms := fmt.Sprintf("%s/%s/%s/%s", expectedRegistryModuleAttributes.RegistryName, expectedRegistryModuleAttributes.Namespace, expectedRegistryModuleAttributes.Name, expectedRegistryModuleAttributes.Provider)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFERegistryModuleDataSourceBasic_noCodePublic(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dsName, "no_code_module_id", rsName, "id"),
					resource.TestCheckResourceAttr("data.tfe_registry_module.test", "organization", orgName),
					resource.TestCheckResourceAttr("data.tfe_registry_module.test", "namespace", expectedRegistryModuleAttributes.Namespace),
					resource.TestCheckResourceAttr("data.tfe_registry_module.test", "name", expectedRegistryModuleAttributes.Name),
					resource.TestCheckResourceAttr("data.tfe_registry_module.test", "registry_name", string(expectedRegistryModuleAttributes.RegistryName)),
					resource.TestCheckResourceAttr("data.tfe_registry_module.test", "module_provider", expectedRegistryModuleAttributes.Provider),
					resource.TestCheckResourceAttr("data.tfe_registry_module.test", "no_code_module_source", ncms),
				),
			},
		},
	})
}

func TestAccTFERegistryModuleDataSource_basicNoCodePrivate_VCSDependent(t *testing.T) {
	skipIfEnterprise(t)

	dsName := "data.tfe_registry_module.test"
	rsName := "tfe_no_code_module.foobar"
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	expectedRegistryModuleAttributes := &tfe.RegistryModule{
		Name:         getRegistryModuleName(),
		Provider:     getRegistryModuleProvider(),
		RegistryName: tfe.PrivateRegistry,
		Namespace:    orgName,
		Organization: &tfe.Organization{Name: orgName},
	}
	ncms := fmt.Sprintf("%s/%s/%s/%s", expectedRegistryModuleAttributes.RegistryName, expectedRegistryModuleAttributes.Namespace, expectedRegistryModuleAttributes.Name, expectedRegistryModuleAttributes.Provider)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFERegistryModuleDataSourceBasic_noCodePrivate(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dsName, "no_code_module_id", rsName, "id"),
					resource.TestCheckResourceAttr("data.tfe_registry_module.test", "no_code_module_source", ncms),
				),
			},
		},
	})
}

func testAccTFERegistryModuleDataSourceBasic(rInt int) string {
	return fmt.Sprintf(`
%s
data "tfe_registry_module" "test" {
  organization    = tfe_registry_module.foobar.organization
  name            = tfe_registry_module.foobar.name
  module_provider = tfe_registry_module.foobar.module_provider
}`, testAccTFERegistryModule_vcsBasic(rInt))
}

func testAccTFERegistryModuleDataSourceBasic_noCodePublic(rInt int) string {
	return fmt.Sprintf(`
%s
data "tfe_registry_module" "test" {
  organization    = tfe_no_code_module.foobar.organization
  name            = tfe_registry_module.foobar.name
  module_provider = tfe_registry_module.foobar.module_provider
  registry_name   = tfe_registry_module.foobar.registry_name
  namespace       = tfe_registry_module.foobar.namespace
}`, testAccTFERegistryModule_NoCode(rInt))
}

func testAccTFERegistryModuleDataSourceBasic_noCodePrivate(rInt int) string {
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
 }
}

resource "tfe_no_code_module" "foobar" {
  organization    = tfe_organization.foobar.id
  registry_module = tfe_registry_module.foobar.id
} 

data "tfe_registry_module" "test" {
  organization    = tfe_no_code_module.foobar.organization
  name            = tfe_registry_module.foobar.name
  module_provider = tfe_registry_module.foobar.module_provider
  registry_name   = tfe_registry_module.foobar.registry_name
  namespace       = tfe_registry_module.foobar.namespace
}
`,
		rInt,
		envGithubToken,
		envGithubRegistryModuleIdentifer,
		envGithubRegistryModuleIdentifer)
}
