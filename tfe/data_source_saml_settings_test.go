// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tfe

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

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
					resource.TestCheckResourceAttrSet(resourceAddress, "idp_cert"),
					resource.TestCheckResourceAttrSet(resourceAddress, "old_idp_cert"),
					resource.TestCheckResourceAttrSet(resourceAddress, "slo_endpoint_url"),
					resource.TestCheckResourceAttrSet(resourceAddress, "sso_endpoint_url"),
					resource.TestCheckResourceAttrSet(resourceAddress, "attr_username"),
					resource.TestCheckResourceAttrSet(resourceAddress, "attr_groups"),
					resource.TestCheckResourceAttrSet(resourceAddress, "attr_site_admin"),
					resource.TestCheckResourceAttrSet(resourceAddress, "site_admin_role"),
					resource.TestCheckResourceAttrSet(resourceAddress, "sso_api_token_session_timeout"),
					resource.TestCheckResourceAttrSet(resourceAddress, "acs_consumer_url"),
					resource.TestCheckResourceAttrSet(resourceAddress, "metadata_url"),
					resource.TestCheckResourceAttrSet(resourceAddress, "certificate"),
				),
			},
		},
	},
	)
}

func testAccTFESAMLSettingsDataSourceConfig_basic() string {
	return fmt.Sprintf(`data "tfe_saml_settings" "foobar" {}`)
}
