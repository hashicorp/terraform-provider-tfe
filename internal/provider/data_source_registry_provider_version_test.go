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

func TestAccTFERegistryProviderVersionDataSource_basic(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFERegistryProviderVersionDataSourceConfig(orgName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tfe_registry_provider_version.foobar", "id"),
					resource.TestCheckResourceAttr("data.tfe_registry_provider_version.foobar", "organization", orgName),
					resource.TestCheckResourceAttr("data.tfe_registry_provider_version.foobar", "registry_name", "private"),
					resource.TestCheckResourceAttr("data.tfe_registry_provider_version.foobar", "namespace", orgName),
					resource.TestCheckResourceAttr("data.tfe_registry_provider_version.foobar", "name", "example"),
					resource.TestCheckResourceAttr("data.tfe_registry_provider_version.foobar", "version", "1.0.0"),
					resource.TestCheckResourceAttrSet("data.tfe_registry_provider_version.foobar", "key_id"),
					resource.TestCheckResourceAttr("data.tfe_registry_provider_version.foobar", "protocols.#", "1"),
					resource.TestCheckResourceAttr("data.tfe_registry_provider_version.foobar", "protocols.0", "5.0"),
					resource.TestCheckResourceAttrSet("data.tfe_registry_provider_version.foobar", "created_at"),
					resource.TestCheckResourceAttrSet("data.tfe_registry_provider_version.foobar", "updated_at"),
				),
			},
		},
	})
}

func TestAccTFERegistryProviderVersionDataSource_withPlatforms(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFERegistryProviderVersionDataSourceConfig_withPlatforms(orgName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tfe_registry_provider_version.foobar", "id"),
					resource.TestCheckResourceAttr("data.tfe_registry_provider_version.foobar", "organization", orgName),
					resource.TestCheckResourceAttr("data.tfe_registry_provider_version.foobar", "name", "example"),
					resource.TestCheckResourceAttr("data.tfe_registry_provider_version.foobar", "version", "1.0.0"),
					resource.TestCheckResourceAttr("data.tfe_registry_provider_version.foobar", "platforms.#", "2"),
					resource.TestCheckResourceAttrSet("data.tfe_registry_provider_version.foobar", "platforms.0.id"),
					resource.TestCheckResourceAttrSet("data.tfe_registry_provider_version.foobar", "platforms.0.os_arch"),
					resource.TestCheckResourceAttrSet("data.tfe_registry_provider_version.foobar", "platforms.0.filename"),
					resource.TestCheckResourceAttrSet("data.tfe_registry_provider_version.foobar", "platforms.0.shasum"),
				),
			},
		},
	})
}

func testAccTFERegistryProviderVersionDataSourceConfig(orgName string) string {
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

resource "tfe_registry_provider_version" "foobar" {
  organization = tfe_organization.foobar.name
  name         = tfe_registry_provider.foobar.name
  version      = "1.0.0"
  key_id       = tfe_registry_gpg_key.foobar.id
  protocols    = ["5.0"]
}

data "tfe_registry_provider_version" "foobar" {
  organization = tfe_organization.foobar.name
  name         = tfe_registry_provider.foobar.name
  version      = tfe_registry_provider_version.foobar.version
}
`, orgName, testGPGKeyArmor)
}

func testAccTFERegistryProviderVersionDataSourceConfig_withPlatforms(orgName string) string {
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

resource "tfe_registry_provider_version" "foobar" {
  organization = tfe_organization.foobar.name
  name         = tfe_registry_provider.foobar.name
  version      = "1.0.0"
  key_id       = tfe_registry_gpg_key.foobar.id
  protocols    = ["5.0"]
}

resource "tfe_registry_provider_version_platform" "linux_amd64" {
  organization = tfe_organization.foobar.name
  name         = tfe_registry_provider.foobar.name
  version      = tfe_registry_provider_version.foobar.version
  os_arch      = "linux_amd64"
  filename     = "https://releases.hashicorp.com/terraform-provider-null/3.2.1/terraform-provider-null_3.2.1_linux_amd64.zip"
}

resource "tfe_registry_provider_version_platform" "darwin_amd64" {
  organization = tfe_organization.foobar.name
  name         = tfe_registry_provider.foobar.name
  version      = tfe_registry_provider_version.foobar.version
  os_arch      = "darwin_amd64"
  filename     = "https://releases.hashicorp.com/terraform-provider-null/3.2.1/terraform-provider-null_3.2.1_darwin_amd64.zip"
}

data "tfe_registry_provider_version" "foobar" {
  organization = tfe_organization.foobar.name
  name         = tfe_registry_provider.foobar.name
  version      = tfe_registry_provider_version.foobar.version

  depends_on = [
    tfe_registry_provider_version_platform.linux_amd64,
    tfe_registry_provider_version_platform.darwin_amd64,
  ]
}
`, orgName, testGPGKeyArmor)
}
