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

func TestAccTFENoCodeModuleDataSource_public(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFENoCodeModuleDataSourceConfig_public(rInt),
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

// func TestAccTFENoCodeModuleDataSource_private(t *testing.T) {
// 	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
// 	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

// 	resource.Test(t, resource.TestCase{
// 		PreCheck:                 func() { testAccPreCheck(t) },
// 		ProtoV5ProviderFactories: testAccMuxedProviders,
// 		Steps: []resource.TestStep{
// 			{
// 				Config: testAccTFENoCodeModuleDataSourceConfig_private(orgName),
// 				Check: resource.ComposeAggregateTestCheckFunc(
// 					resource.TestCheckResourceAttrSet("tfe_registry_provider.foobar", "id"),
// 					resource.TestCheckResourceAttr("tfe_registry_provider.foobar", "organization", orgName),
// 					resource.TestCheckResourceAttr("tfe_registry_provider.foobar", "registry_name", "private"),
// 					resource.TestCheckResourceAttr("tfe_registry_provider.foobar", "namespace", orgName),
// 					resource.TestCheckResourceAttr("tfe_registry_provider.foobar", "name", "example"),
// 					resource.TestCheckResourceAttrSet("tfe_registry_provider.foobar", "created_at"),
// 					resource.TestCheckResourceAttrSet("tfe_registry_provider.foobar", "updated_at"),
// 				),
// 			},
// 		},
// 	})
// }

// need:
//
//	:organization_name
//	:registry_name
//	:namespace
//	:name
//	:provider
//
// # namespace = tfe_organization.foobar.name  namespace is same as org
// name?
func testAccTFENoCodeModuleDataSourceConfig_public(rInt int) string {
	return fmt.Sprintf(`
%s

data "no_code_module" "foobar" {
  organization = tfe_organization.foobar.name
  registry_name = "public"
  name          = tfe_no_code_module.foobar.name
  provider = tfe_registry_provider.foobar.name
}
`, testAccTFENoCodeModule_basic(rInt))
}

// func testAccTFENoCodeModuleDataSourceConfig_private(orgName string) string {
// 	return fmt.Sprintf(`
// %s

// data "tfe_registry_provider" "foobar" {
//   organization = tfe_organization.foobar.name

//   name = tfe_registry_provider.foobar.name
// }
// `, testAccTFENoCodeModuleResourceConfig_private(orgName))
// }
