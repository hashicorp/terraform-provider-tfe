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

func TestAccTFERegistryProviderVersionPlatformDataSource_basic(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFERegistryProviderVersionPlatformDataSourceConfig(orgName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tfe_registry_provider_version_platform.foobar", "id"),
					resource.TestCheckResourceAttr("data.tfe_registry_provider_version_platform.foobar", "organization", orgName),
					resource.TestCheckResourceAttr("data.tfe_registry_provider_version_platform.foobar", "registry_name", "private"),
					resource.TestCheckResourceAttr("data.tfe_registry_provider_version_platform.foobar", "namespace", orgName),
					resource.TestCheckResourceAttr("data.tfe_registry_provider_version_platform.foobar", "name", "example"),
					resource.TestCheckResourceAttr("data.tfe_registry_provider_version_platform.foobar", "version", "1.0.0"),
					resource.TestCheckResourceAttr("data.tfe_registry_provider_version_platform.foobar", "os_arch", "linux_amd64"),
					resource.TestCheckResourceAttrSet("data.tfe_registry_provider_version_platform.foobar", "filename"),
					resource.TestCheckResourceAttrSet("data.tfe_registry_provider_version_platform.foobar", "shasum"),
					resource.TestCheckResourceAttrSet("data.tfe_registry_provider_version_platform.foobar", "registry_provider_id"),
				),
			},
		},
	})
}

func testAccTFERegistryProviderVersionPlatformDataSourceConfig(orgName string) string {
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

resource "tfe_registry_provider_version_platform" "foobar" {
  organization = tfe_organization.foobar.name
  name         = tfe_registry_provider.foobar.name
  version      = tfe_registry_provider_version.foobar.version
  os_arch      = "linux_amd64"
  filename     = "https://releases.hashicorp.com/terraform-provider-null/3.2.1/terraform-provider-null_3.2.1_linux_amd64.zip"
}

data "tfe_registry_provider_version_platform" "foobar" {
  organization = tfe_organization.foobar.name
  name         = tfe_registry_provider.foobar.name
  version      = tfe_registry_provider_version.foobar.version
  os_arch      = tfe_registry_provider_version_platform.foobar.os_arch
}
`, orgName, testGPGKeyArmor)
}
