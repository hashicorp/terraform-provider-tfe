// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
)

func TestAccTFERegistryProviderVersionResource_basic(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFERegistryProviderVersionResourceConfig_basic(orgName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tfe_registry_provider_version.foobar", "id"),
					resource.TestCheckResourceAttr("tfe_registry_provider_version.foobar", "organization", orgName),
					resource.TestCheckResourceAttr("tfe_registry_provider_version.foobar", "registry_name", "private"),
					resource.TestCheckResourceAttr("tfe_registry_provider_version.foobar", "namespace", orgName),
					resource.TestCheckResourceAttr("tfe_registry_provider_version.foobar", "name", "example"),
					resource.TestCheckResourceAttr("tfe_registry_provider_version.foobar", "version", "1.0.0"),
					resource.TestCheckResourceAttrSet("tfe_registry_provider_version.foobar", "key_id"),
					resource.TestCheckResourceAttr("tfe_registry_provider_version.foobar", "protocols.#", "1"),
					resource.TestCheckResourceAttr("tfe_registry_provider_version.foobar", "protocols.0", "5.0"),
					resource.TestCheckResourceAttrSet("tfe_registry_provider_version.foobar", "created_at"),
					resource.TestCheckResourceAttrSet("tfe_registry_provider_version.foobar", "updated_at"),
				),
			},
		},
	})
}

func TestAccTFERegistryProviderVersionResource_multipleProtocols(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFERegistryProviderVersionResourceConfig_multipleProtocols(orgName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tfe_registry_provider_version.foobar", "id"),
					resource.TestCheckResourceAttr("tfe_registry_provider_version.foobar", "organization", orgName),
					resource.TestCheckResourceAttr("tfe_registry_provider_version.foobar", "name", "example"),
					resource.TestCheckResourceAttr("tfe_registry_provider_version.foobar", "version", "2.0.0"),
					resource.TestCheckResourceAttr("tfe_registry_provider_version.foobar", "protocols.#", "2"),
					resource.TestCheckResourceAttr("tfe_registry_provider_version.foobar", "protocols.0", "5.0"),
					resource.TestCheckResourceAttr("tfe_registry_provider_version.foobar", "protocols.1", "6.0"),
				),
			},
		},
	})
}

func TestAccTFERegistryProviderVersionResource_importByIdentity(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFERegistryProviderVersionResourceConfig_basic(orgName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectIdentity("tfe_registry_provider_version.foobar", map[string]knownvalue.Check{
						"id":            knownvalue.NotNull(),
						"hostname":      knownvalue.StringExact(os.Getenv("TFE_HOSTNAME")),
						"organization":  knownvalue.StringExact(orgName),
						"registry_name": knownvalue.StringExact(string(tfe.PrivateRegistry)),
						"namespace":     knownvalue.StringExact(orgName),
						"name":          knownvalue.StringExact("example"),
						"version":       knownvalue.StringExact("1.0.0"),
					}),
				},
			},
			{
				ResourceName:    "tfe_registry_provider_version.foobar",
				ImportState:     true,
				ImportStateKind: resource.ImportBlockWithResourceIdentity,
			},
		},
	})
}

func TestAccTFERegistryProviderVersionResource_importByID(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFERegistryProviderVersionResourceConfig_basic(orgName),
			},
			{
				ResourceName:      "tfe_registry_provider_version.foobar",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     fmt.Sprintf("%s/private/%s/example/1.0.0", orgName, orgName),
			},
		},
	})
}

func TestAccTFERegistryProviderVersionResource_withShasums(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFERegistryProviderVersionResourceConfig_withShasums(orgName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tfe_registry_provider_version.foobar", "id"),
					resource.TestCheckResourceAttr("tfe_registry_provider_version.foobar", "organization", orgName),
					resource.TestCheckResourceAttr("tfe_registry_provider_version.foobar", "name", "example"),
					resource.TestCheckResourceAttr("tfe_registry_provider_version.foobar", "version", "3.2.1"),
					resource.TestCheckResourceAttr("tfe_registry_provider_version.foobar", "shasums_uploaded", "true"),
					resource.TestCheckResourceAttr("tfe_registry_provider_version.foobar", "shasums_sig_uploaded", "true"),
				),
			},
		},
	})
}

func testAccTFERegistryProviderVersionResourceConfig_basic(orgName string) string {
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
`, orgName, testGPGKeyArmor)
}

func testAccTFERegistryProviderVersionResourceConfig_multipleProtocols(orgName string) string {
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
  version      = "2.0.0"
  key_id       = tfe_registry_gpg_key.foobar.id
  protocols    = ["5.0", "6.0"]
}
`, orgName, testGPGKeyArmor)
}

func testAccTFERegistryProviderVersionResourceConfig_withShasums(orgName string) string {
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
  organization     = tfe_organization.foobar.name
  name             = tfe_registry_provider.foobar.name
  version          = "3.2.1"
  key_id           = tfe_registry_gpg_key.foobar.id
  protocols        = ["5.0"]
  shasums_file     = "https://releases.hashicorp.com/terraform-provider-null/3.2.1/terraform-provider-null_3.2.1_SHA256SUMS"
  shasums_sig_file = "https://releases.hashicorp.com/terraform-provider-null/3.2.1/terraform-provider-null_3.2.1_SHA256SUMS.sig"
}
`, orgName, testGPGKeyArmor)
}
