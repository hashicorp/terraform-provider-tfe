// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTFERegistryProviderVersionsDataSource_basic(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFERegistryProviderVersionsDataSourceConfig(orgName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tfe_registry_provider_versions.foobar", "id"),
					resource.TestCheckResourceAttr("data.tfe_registry_provider_versions.foobar", "organization", orgName),
					resource.TestCheckResourceAttr("data.tfe_registry_provider_versions.foobar", "registry_name", "private"),
					resource.TestCheckResourceAttr("data.tfe_registry_provider_versions.foobar", "namespace", orgName),
					resource.TestCheckResourceAttr("data.tfe_registry_provider_versions.foobar", "name", "example"),
					resource.TestCheckResourceAttr("data.tfe_registry_provider_versions.foobar", "versions.#", "2"),
					// Check that both versions exist (order may vary)
					resource.TestCheckTypeSetElemNestedAttrs("data.tfe_registry_provider_versions.foobar", "versions.*", map[string]string{
						"version": "1.0.0",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("data.tfe_registry_provider_versions.foobar", "versions.*", map[string]string{
						"version": "2.0.0",
					}),
				),
			},
		},
	})
}

func testAccTFERegistryProviderVersionsDataSourceConfig(orgName string) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "%s"
  email = "admin@tfe.local"
}

resource "tfe_registry_provider" "foobar" {
  organization = tfe_organization.foobar.name
  name         = "example"
}

resource "tfe_registry_gpg_key" "foobar" {
  organization = tfe_organization.foobar.name
  ascii_armor  = <<-EOT
%s
EOT
}

resource "tfe_registry_provider_version" "v1" {
  organization = tfe_organization.foobar.name
  name         = tfe_registry_provider.foobar.name
  version      = "1.0.0"
  key_id       = tfe_registry_gpg_key.foobar.id
  protocols    = ["5.0"]
}

resource "tfe_registry_provider_version" "v2" {
  organization = tfe_organization.foobar.name
  name         = tfe_registry_provider.foobar.name
  version      = "2.0.0"
  key_id       = tfe_registry_gpg_key.foobar.id
  protocols    = ["5.0", "6.0"]
}

data "tfe_registry_provider_versions" "foobar" {
  organization = tfe_organization.foobar.name
  name         = tfe_registry_provider.foobar.name

  depends_on = [
    tfe_registry_provider_version.v1,
    tfe_registry_provider_version.v2
  ]
}
`, orgName, testGPGKeyArmor)
}
