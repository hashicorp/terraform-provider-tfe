// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// FLAKE ALERT: SMTP settings are a singleton resource shared by the entire TFE
// instance, and any test touching them is at high risk to flake.
// This test is fine, because it only checks that the attributes have SOME
// value. Testing for any _particular_ value would not be viable, because
// `resource_tfe_smtp_settings_test.go` exists. See that file for more color on
// this.
func TestAccTFESMTPSettingsDataSource_basic(t *testing.T) {
	resourceAddress := "data.tfe_smtp_settings.foobar"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFESMTPSettingsDataSourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceAddress, "id"),
					resource.TestCheckResourceAttrSet(resourceAddress, "enabled"),
					// Note: host, port, and auth may be empty when SMTP is disabled
					// so we don't check for them here
				),
			},
		},
	},
	)
}

func testAccTFESMTPSettingsDataSourceConfig_basic() string {
	return `data "tfe_smtp_settings" "foobar" {}`
}
