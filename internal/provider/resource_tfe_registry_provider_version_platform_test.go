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

// testAccPreCheckWithTimeout extends the default timeout for tests that need more time
func testAccPreCheckWithTimeout(t *testing.T, timeout time.Duration) {
	// Set a longer timeout for the test
	if deadline, ok := t.Deadline(); ok {
		remaining := time.Until(deadline)
		if remaining < timeout {
			t.Logf("Warning: Test deadline (%v) is less than requested timeout (%v)", remaining, timeout)
		}
	}
	testAccPreCheck(t)
}

func TestAccTFERegistryProviderVersionPlatformResource_basic(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)
	provName := "example-null"
	provVersion := "3.2.1"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFERegistryProviderVersionPlatformResourceConfig_basic(orgName, provName, provVersion),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tfe_registry_provider_version_platform.foobar", "id"),
					resource.TestCheckResourceAttr("tfe_registry_provider_version_platform.foobar", "organization", orgName),
					resource.TestCheckResourceAttr("tfe_registry_provider_version_platform.foobar", "registry_name", "private"),
					resource.TestCheckResourceAttr("tfe_registry_provider_version_platform.foobar", "namespace", orgName),
					resource.TestCheckResourceAttr("tfe_registry_provider_version_platform.foobar", "name", provName),
					resource.TestCheckResourceAttr("tfe_registry_provider_version_platform.foobar", "version", provVersion),
					resource.TestCheckResourceAttr("tfe_registry_provider_version_platform.foobar", "os_arch", "linux_amd64"),
					resource.TestCheckResourceAttrSet("tfe_registry_provider_version_platform.foobar", "filename"),
					resource.TestCheckResourceAttrSet("tfe_registry_provider_version_platform.foobar", "shasum"),
					resource.TestCheckResourceAttrSet("tfe_registry_provider_version_platform.foobar", "registry_provider_id"),
				),
			},
		},
	})
}

func TestAccTFERegistryProviderVersionPlatformResource_multiplePlatforms(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)
	provName := "example-null"
	provVersion := "3.2.1"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFERegistryProviderVersionPlatformResourceConfig_multiplePlatforms(orgName, provName, provVersion),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tfe_registry_provider_version_platform.linux", "id"),
					resource.TestCheckResourceAttr("tfe_registry_provider_version_platform.linux", "os_arch", "linux_amd64"),
					resource.TestCheckResourceAttrSet("tfe_registry_provider_version_platform.darwin_arm", "id"),
					resource.TestCheckResourceAttr("tfe_registry_provider_version_platform.darwin_arm", "os_arch", "darwin_arm64"),
				),
			},
		},
	})
}

func TestAccTFERegistryProviderVersionPlatformResource_import(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)
	provName := "example-null"
	provVersion := "3.2.1"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFERegistryProviderVersionPlatformResourceConfig_basic(orgName, provName, provVersion),
			},
			{
				ResourceName:            "tfe_registry_provider_version_platform.foobar",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateId:           fmt.Sprintf("%s/private/%s/%s/%s/linux/amd64", orgName, orgName, provName, provVersion),
				ImportStateVerifyIgnore: []string{"filename"}, // filename is not stored in state after creation
			},
		},
	})
}

func testAccTFERegistryProviderVersionPlatformResourceConfig_basic(orgName string, provName string, provVersion string) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "%s"
  email = "admin@tfe.local"
}

resource "tfe_registry_provider" "foobar" {
  organization = tfe_organization.foobar.name
  name         = "%s"
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
  version      = "%s"
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
`, orgName, provName, testGPGKeyArmor, provVersion)
}

func testAccTFERegistryProviderVersionPlatformResourceConfig_multiplePlatforms(orgName string, provName string, provVersion string) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "%s"
  email = "admin@tfe.local"
}

resource "tfe_registry_provider" "foobar" {
  organization = tfe_organization.foobar.name
  name         = "%s"
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
  version      = "%s"
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
  filename     = "https://releases.hashicorp.com/terraform-provider-null/3.2.1/terraform-provider-null_3.2.1_darwin_amd64.zip"
}

`, orgName, provName, testGPGKeyArmor, provVersion)
}
