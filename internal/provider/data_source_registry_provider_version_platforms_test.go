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

func TestAccTFERegistryProviderVersionPlatformsDataSource_basic(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFERegistryProviderVersionPlatformsDataSourceConfig(orgName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tfe_registry_provider_version_platforms.foobar", "id"),
					resource.TestCheckResourceAttr("data.tfe_registry_provider_version_platforms.foobar", "organization", orgName),
					resource.TestCheckResourceAttr("data.tfe_registry_provider_version_platforms.foobar", "registry_name", "private"),
					resource.TestCheckResourceAttr("data.tfe_registry_provider_version_platforms.foobar", "namespace", orgName),
					resource.TestCheckResourceAttr("data.tfe_registry_provider_version_platforms.foobar", "name", "example"),
					resource.TestCheckResourceAttr("data.tfe_registry_provider_version_platforms.foobar", "version", "1.0.0"),
					resource.TestCheckResourceAttr("data.tfe_registry_provider_version_platforms.foobar", "platforms.#", "2"),
					// Check that both platforms exist (order may vary)
					resource.TestCheckTypeSetElemNestedAttrs("data.tfe_registry_provider_version_platforms.foobar", "platforms.*", map[string]string{
						"os_arch": "linux_amd64",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("data.tfe_registry_provider_version_platforms.foobar", "platforms.*", map[string]string{
						"os_arch": "darwin_arm64",
					}),
				),
			},
		},
	})
}

func testAccTFERegistryProviderVersionPlatformsDataSourceConfig(orgName string) string {
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

resource "tfe_registry_provider_version_platform" "linux" {
  organization = tfe_organization.foobar.name
  name         = tfe_registry_provider.foobar.name
  version      = tfe_registry_provider_version.foobar.version
  os_arch      = "linux_amd64"
  filename     = "https://releases.hashicorp.com/terraform-provider-null/3.2.1/terraform-provider-null_3.2.1_linux_amd64.zip"
}

resource "tfe_registry_provider_version_platform" "darwin_arm" {
  organization = tfe_organization.foobar.name
  name         = tfe_registry_provider.foobar.name
  version      = tfe_registry_provider_version.foobar.version
  os_arch      = "darwin_arm64"
  filename     = "https://releases.hashicorp.com/terraform-provider-null/3.2.1/terraform-provider-null_3.2.1_darwin_arm64.zip"
}

data "tfe_registry_provider_version_platforms" "foobar" {
  organization = tfe_organization.foobar.name
  name         = tfe_registry_provider.foobar.name
  version      = tfe_registry_provider_version.foobar.version

  depends_on = [
    tfe_registry_provider_version_platform.linux,
    tfe_registry_provider_version_platform.darwin_arm
  ]
}
`, orgName, testGPGKeyArmor)
}
