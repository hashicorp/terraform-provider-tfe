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
					// site_auditor_group_scim_id and site_auditor_group_display_name
					// are deliberately not asserted here: TestCheckResourceAttrSet
					// fails on an empty value, and both are empty whenever no group
					// is linked — which is the case on a bare instance, and always
					// on Terraform Enterprise releases older than
					// minTFEVersionSiteAuditor. They are covered with real values by
					// the "SCIM settings site auditor group" subtest in
					// resource_tfe_scim_settings_test.go.
				),
			},
		},
	})
}

func testAccTFESCIMSettingsDataSourceConfig_basic() string {
	return `data "tfe_scim_settings" "foobar"{}`
}
