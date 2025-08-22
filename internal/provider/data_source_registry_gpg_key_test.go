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

func TestAccTFERegistryGPGKeyDataSource_basic(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFERegistryGPGKeyDataSourceConfig(orgName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.tfe_registry_gpg_key.foobar", "organization", orgName),
					resource.TestCheckResourceAttrSet("data.tfe_registry_gpg_key.foobar", "id"),
					resource.TestCheckResourceAttrSet("data.tfe_registry_gpg_key.foobar", "ascii_armor"),
					resource.TestCheckResourceAttrSet("data.tfe_registry_gpg_key.foobar", "created_at"),
					resource.TestCheckResourceAttrSet("data.tfe_registry_gpg_key.foobar", "updated_at")),
			},
		},
	})
}

func testAccTFERegistryGPGKeyDataSourceConfig(orgName string) string {
	return fmt.Sprintf(`
%s

data "tfe_registry_gpg_key" "foobar" {
  organization = tfe_organization.foobar.name

  id = tfe_registry_gpg_key.foobar.id
}
`, testAccTFERegistryGPGKeyResourceConfig(orgName))
}
