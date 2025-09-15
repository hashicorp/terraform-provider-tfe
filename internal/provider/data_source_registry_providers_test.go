// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccTFERegistryProvidersDataSource_all(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFERegistryProvidersDataSourceConfig_all(orgName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.tfe_registry_providers.foobar", "organization", orgName),
					testAccTFERegistryProvidersDataSourceEntries("data.tfe_registry_providers.foobar", []string{
						"hashicorp/tfe",
						fmt.Sprintf("%s/foobar", orgName),
					}),
				),
			},
		},
	})
}

func TestAccTFERegistryProvidersDataSource_public(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFERegistryProvidersDataSourceConfig_public(orgName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.tfe_registry_providers.foobar", "organization", orgName),
					testAccTFERegistryProvidersDataSourceEntries("data.tfe_registry_providers.foobar", []string{
						"hashicorp/tfe",
					}),
				),
			},
		},
	})
}

func TestAccTFERegistryProvidersDataSource_private(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFERegistryProvidersDataSourceConfig_private(orgName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.tfe_registry_providers.foobar", "organization", orgName),
					testAccTFERegistryProvidersDataSourceEntries("data.tfe_registry_providers.foobar", []string{
						fmt.Sprintf("%s/foobar", orgName),
					}),
				),
			},
		},
	})
}

func TestAccTFERegistryProvidersDataSource_filtered(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFERegistryProvidersDataSourceConfig_filtered(orgName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.tfe_registry_providers.foobar", "organization", orgName),
					testAccTFERegistryProvidersDataSourceEntries("data.tfe_registry_providers.foobar", []string{
						fmt.Sprintf("%s/foobar", orgName),
					}),
				),
			},
		},
	})
}

func testAccTFERegistryProvidersDataSourceEntries(resourceName string, providers []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		providerDataSource, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("data source '%s' not found.", resourceName)
		}
		numProvidersStr := providerDataSource.Primary.Attributes["providers.#"]
		numProviders, _ := strconv.Atoi(numProvidersStr)

		if numProviders != len(providers) {
			return fmt.Errorf("expected %d providers, but found %d.", len(providers), numProviders)
		}

		allProvidersMap := map[string]struct{}{}
		for i := 0; i < numProviders; i++ {
			providerNamespace := providerDataSource.Primary.Attributes[fmt.Sprintf("providers.%d.namespace", i)]
			providerName := providerDataSource.Primary.Attributes[fmt.Sprintf("providers.%d.name", i)]
			allProvidersMap[fmt.Sprintf("%s/%s", providerNamespace, providerName)] = struct{}{}
		}

		for _, provider := range providers {
			if _, ok := allProvidersMap[provider]; !ok {
				return fmt.Errorf("expected provider '%s' not found.", provider)
			}
		}

		return nil
	}
}

func testAccTFERegistryProvidersDataSourceConfig_resources(orgName string) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "%s"
  email = "admin@tfe.local"
}

resource "tfe_registry_provider" "public" {
  organization  = tfe_organization.foobar.name
  registry_name = "public"
  namespace     = "hashicorp"
  name          = "tfe"
}

resource "tfe_registry_provider" "private" {
  organization  = tfe_organization.foobar.name
  registry_name = "private"
  name 	        = "foobar"
}
`, orgName)
}

func testAccTFERegistryProvidersDataSourceConfig_all(orgName string) string {
	return fmt.Sprintf(`
%s

data "tfe_registry_providers" "foobar" {
  organization = tfe_organization.foobar.name

  depends_on = [tfe_registry_provider.public, tfe_registry_provider.private]
}
`, testAccTFERegistryProvidersDataSourceConfig_resources(orgName))
}

func testAccTFERegistryProvidersDataSourceConfig_public(orgName string) string {
	return fmt.Sprintf(`
%s

data "tfe_registry_providers" "foobar" {
  organization  = tfe_organization.foobar.name
  registry_name = "public"

  depends_on = [tfe_registry_provider.public, tfe_registry_provider.private]
}
`, testAccTFERegistryProvidersDataSourceConfig_resources(orgName))
}

func testAccTFERegistryProvidersDataSourceConfig_private(orgName string) string {
	return fmt.Sprintf(`
%s

data "tfe_registry_providers" "foobar" {
  organization  = tfe_organization.foobar.name
  registry_name = "private"

  depends_on = [tfe_registry_provider.public, tfe_registry_provider.private]
}
`, testAccTFERegistryProvidersDataSourceConfig_resources(orgName))
}

func testAccTFERegistryProvidersDataSourceConfig_filtered(orgName string) string {
	return fmt.Sprintf(`
%s

data "tfe_registry_providers" "foobar" {
  organization = tfe_organization.foobar.name
  search       = "foo"

  depends_on = [tfe_registry_provider.public, tfe_registry_provider.private]
}
`, testAccTFERegistryProvidersDataSourceConfig_resources(orgName))
}
