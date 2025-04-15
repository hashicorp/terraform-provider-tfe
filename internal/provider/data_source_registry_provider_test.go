// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTFERegistryProviderDataSource_public(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFERegistryProviderDataSourceConfig_public(orgName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tfe_registry_provider.foobar", "id"),
					resource.TestCheckResourceAttr("tfe_registry_provider.foobar", "organization", orgName),
					resource.TestCheckResourceAttr("tfe_registry_provider.foobar", "registry_name", "public"),
					resource.TestCheckResourceAttr("tfe_registry_provider.foobar", "namespace", "hashicorp"),
					resource.TestCheckResourceAttr("tfe_registry_provider.foobar", "name", "aws"),
					resource.TestCheckResourceAttrSet("tfe_registry_provider.foobar", "created_at"),
					resource.TestCheckResourceAttrSet("tfe_registry_provider.foobar", "updated_at"),
				),
			},
		},
	})
}

func TestAccTFERegistryProviderDataSource_private(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFERegistryProviderDataSourceConfig_private(orgName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tfe_registry_provider.foobar", "id"),
					resource.TestCheckResourceAttr("tfe_registry_provider.foobar", "organization", orgName),
					resource.TestCheckResourceAttr("tfe_registry_provider.foobar", "registry_name", "private"),
					resource.TestCheckResourceAttr("tfe_registry_provider.foobar", "namespace", orgName),
					resource.TestCheckResourceAttr("tfe_registry_provider.foobar", "name", "example"),
					resource.TestCheckResourceAttrSet("tfe_registry_provider.foobar", "created_at"),
					resource.TestCheckResourceAttrSet("tfe_registry_provider.foobar", "updated_at"),
				),
			},
		},
	})
}

func testAccTFERegistryProviderDataSourceConfig_public(orgName string) string {
	return fmt.Sprintf(`
%s

data "tfe_registry_provider" "foobar" {
  organization = tfe_organization.foobar.name

  registry_name = "public"
  namespace     = tfe_registry_provider.foobar.namespace
  name          = tfe_registry_provider.foobar.name
}
`, testAccTFERegistryProviderResourceConfig_public(orgName))
}

func testAccTFERegistryProviderDataSourceConfig_private(orgName string) string {
	return fmt.Sprintf(`
%s

data "tfe_registry_provider" "foobar" {
  organization = tfe_organization.foobar.name

  name = tfe_registry_provider.foobar.name
}
`, testAccTFERegistryProviderResourceConfig_private(orgName))
}
