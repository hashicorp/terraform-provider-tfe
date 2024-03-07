// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccTFERegistryGPGKeysDataSource_basic(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFERegistryGPGKeysDataSourceConfig(orgName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.tfe_registry_gpg_keys.all", "organization", orgName),
					resource.TestCheckResourceAttr(
						"data.tfe_registry_gpg_keys.all", "keys.#", "1"),
					resource.TestCheckResourceAttrSet(
						"data.tfe_registry_gpg_keys.all", "keys.0.id"),
					resource.TestCheckResourceAttr(
						"data.tfe_registry_gpg_keys.all", "keys.0.organization", orgName),
					resource.TestCheckResourceAttrSet(
						"data.tfe_registry_gpg_keys.all", "keys.0.ascii_armor"),
				),
			},
		},
	})
}

func TestAccTFERegistryGPGKeysDataSource_basicNoKeys(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFERegistryGPGKeysDataSourceConfig_noKeys(orgName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.tfe_registry_gpg_keys.all", "organization", orgName),
					resource.TestCheckResourceAttr(
						"data.tfe_registry_gpg_keys.all", "keys.#", "0"),
				),
			},
		},
	})
}

func testAccTFERegistryGPGKeysDataSourceConfig(orgName string) string {
	return fmt.Sprintf(`
%s

data "tfe_registry_gpg_keys" "all" {
  organization = tfe_organization.foobar.name

  depends_on = [tfe_registry_gpg_key.foobar]
}
`, testAccTFERegistryGPGKeyResourceConfig(orgName))
}

func testAccTFERegistryGPGKeysDataSourceConfig_noKeys(orgName string) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "%s"
  email = "admin@tfe.local"
}

data "tfe_registry_gpg_keys" "all" {
  organization = tfe_organization.foobar.name
}
`, orgName)
}
