// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccCurrentUserDataSource_basic(t *testing.T) {
	resourceAddress := "data.tfe_current_user.test"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCurrentUserDataSourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceAddress, "id"),
					resource.TestCheckResourceAttrSet(resourceAddress, "username"),
					resource.TestCheckResourceAttrSet(resourceAddress, "email"),
					resource.TestCheckResourceAttrSet(resourceAddress, "is_service_account"),
				),
			},
		},
	})
}

func testAccCurrentUserDataSourceConfig_basic() string {
	return `data "tfe_current_user" "test" {}`
}
