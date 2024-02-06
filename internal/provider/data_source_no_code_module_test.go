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
	skipUnlessBeta(t)
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFENoCodeModuleDataSourceConfig_public(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tfe_no_code_module.foobar", "id"),
					resource.TestCheckResourceAttr("data.tfe_no_code_module.foobar", "organization", orgName),
					resource.TestCheckResourceAttr("data.tfe_no_code_module.foobar", "enabled", "true"),
					resource.TestCheckResourceAttrSet("data.tfe_no_code_module.foobar", "registry_module_id"),
				),
			},
		},
	})
}

func testAccTFENoCodeModuleDataSourceConfig_public(rInt int) string {
	return fmt.Sprintf(`
%s

data "tfe_no_code_module" "foobar" {
		id = tfe_no_code_module.foobar.id
}
`, testAccTFENoCodeModule_basic(rInt))
}
