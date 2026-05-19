// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTFESCIMSettingsDataSource_basic(t *testing.T) {
	skipIfCloud(t)

	resourceAddress := "data.tfe_scim_settings.foobar"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFESCIMSettingsDataSourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceAddress, "id"),
					resource.TestCheckResourceAttrSet(resourceAddress, "enabled"),
					resource.TestCheckResourceAttrSet(resourceAddress, "paused"),
				),
			},
		},
	})
}

func testAccTFESCIMSettingsDataSourceConfig_basic() string {
	return `data "tfe_scim_settings" "foobar"{}`
}
