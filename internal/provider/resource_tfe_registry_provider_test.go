// Copyright IBM Corp. 2018, 2025
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

func TestAccTFERegistryProviderResource_public(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFERegistryProviderResourceConfig_public(orgName),
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

func TestAccTFERegistryProviderResource_private(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFERegistryProviderResourceConfig_private(orgName),
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

func TestAccTFERegistryProviderResource_importByIdentity(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFERegistryProviderResourceConfig_private(orgName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectIdentity("tfe_registry_provider.foobar", map[string]knownvalue.Check{
						"id":            knownvalue.NotNull(),
						"hostname":      knownvalue.StringExact(os.Getenv("TFE_HOSTNAME")),
						"organization":  knownvalue.StringExact(orgName),
						"registry_name": knownvalue.StringExact(string(tfe.PrivateRegistry)),
						"namespace":     knownvalue.StringExact(orgName),
						"name":          knownvalue.StringExact("example"),
					}),
				},
			},
			{
				ResourceName:    "tfe_registry_provider.foobar",
				ImportState:     true,
				ImportStateKind: resource.ImportBlockWithResourceIdentity,
			},
			{
				Config: testAccTFERegistryProviderResourceConfig_public(orgName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectIdentity("tfe_registry_provider.foobar", map[string]knownvalue.Check{
						"id":            knownvalue.NotNull(),
						"hostname":      knownvalue.StringExact(os.Getenv("TFE_HOSTNAME")),
						"organization":  knownvalue.StringExact(orgName),
						"registry_name": knownvalue.StringExact(string(tfe.PublicRegistry)),
						"namespace":     knownvalue.StringExact("hashicorp"),
						"name":          knownvalue.StringExact("aws"),
					}),
				},
			},
			{
				ResourceName:    "tfe_registry_provider.foobar",
				ImportState:     true,
				ImportStateKind: resource.ImportBlockWithResourceIdentity,
			},
		},
	})
}

func testAccTFERegistryProviderResourceConfig_public(orgName string) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "%s"
  email = "admin@tfe.local"
}

resource "tfe_registry_provider" "foobar" {
  organization = tfe_organization.foobar.name

  registry_name = "public"
  namespace     = "hashicorp"
  name          = "aws"
}
`, orgName)
}

func testAccTFERegistryProviderResourceConfig_private(orgName string) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "%s"
  email = "admin@tfe.local"
}

resource "tfe_registry_provider" "foobar" {
  organization = tfe_organization.foobar.name

  name = "example"
}
`, orgName)
}
