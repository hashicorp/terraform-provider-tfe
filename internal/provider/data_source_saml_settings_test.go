// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// FLAKE ALERT: SAML settings are a singleton resource shared by the entire TFE
// instance, and any test touching them is at high risk to flake.
// This test is fine, because it only checks that the attributes have SOME
// value. Testing for any _particular_ value would not be viable, because
// `resource_tfe_saml_settings_test.go` exists. See that file for more color on
// this.
func TestAccTFESAMLSettingsDataSource_basic(t *testing.T) {
	resourceAddress := "data.tfe_saml_settings.foobar"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFESAMLSettingsDataSourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceAddress, "id"),
					resource.TestCheckResourceAttrSet(resourceAddress, "enabled"),
					resource.TestCheckResourceAttrSet(resourceAddress, "debug"),
					resource.TestCheckResourceAttrSet(resourceAddress, "team_management_enabled"),
					resource.TestCheckResourceAttrSet(resourceAddress, "authn_requests_signed"),
					resource.TestCheckResourceAttrSet(resourceAddress, "want_assertions_signed"),
					resource.TestCheckResourceAttrSet(resourceAddress, "attr_username"),
					resource.TestCheckResourceAttrSet(resourceAddress, "attr_groups"),
					resource.TestCheckResourceAttrSet(resourceAddress, "attr_site_admin"),
					resource.TestCheckResourceAttrSet(resourceAddress, "site_admin_role"),
					resource.TestCheckResourceAttrSet(resourceAddress, "sso_api_token_session_timeout"),
					resource.TestCheckResourceAttrSet(resourceAddress, "acs_consumer_url"),
					resource.TestCheckResourceAttrSet(resourceAddress, "metadata_url"),
				),
			},
		},
	},
	)
}

func testAccTFESAMLSettingsDataSourceConfig_basic() string {
	return `data "tfe_saml_settings" "foobar" {}`
}
