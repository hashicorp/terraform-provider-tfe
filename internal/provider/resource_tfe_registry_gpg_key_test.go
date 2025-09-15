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

func TestAccTFERegistryGPGKeyResource_basic(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFERegistryGPGKeyResourceConfig(orgName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tfe_registry_gpg_key.foobar", "id", testGPGKeyID),
					resource.TestCheckResourceAttr("tfe_registry_gpg_key.foobar", "organization", orgName),
					resource.TestCheckResourceAttr("tfe_registry_gpg_key.foobar", "ascii_armor", testGPGKeyArmor+"\n"),
					resource.TestCheckResourceAttrSet("tfe_registry_gpg_key.foobar", "created_at"),
					resource.TestCheckResourceAttrSet("tfe_registry_gpg_key.foobar", "updated_at"),
				),
			},
		},
	})
}

func testAccTFERegistryGPGKeyResourceConfig(orgName string) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "%s"
  email = "admin@tfe.local"
}

resource "tfe_registry_gpg_key" "foobar" {
  organization = tfe_organization.foobar.name

  ascii_armor = <<ASCII_ARMOR
%s
ASCII_ARMOR
}
`, orgName, testGPGKeyArmor)
}

const testGPGKeyID string = "13DFECCA3B58CE4A"

const testGPGKeyArmor string = `
-----BEGIN PGP PUBLIC KEY BLOCK-----

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
=bBSS
-----END PGP PUBLIC KEY BLOCK-----
`
