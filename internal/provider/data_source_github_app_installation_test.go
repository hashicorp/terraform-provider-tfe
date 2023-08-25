// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func testAccTFEGHAInstallationDataSourcePreCheck(t *testing.T) {
	testAccPreCheck(t)
	if envGithubAppInstallationName == "" {
		t.Skip("Please set GITHUB_APP_INSTALLATION_NAME to run this test")
	}
}

func TestAccTFEGHAInstallationDataSource_findByName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccTFEGHAInstallationDataSourcePreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEGHAInstallationDataSourceConfig_findByName(envGithubAppInstallationName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.tfe_github_app_installation.gha_installation", "name", envGithubAppInstallationName),
					resource.TestCheckResourceAttrSet("data.tfe_github_app_installation.gha_installation", "installation_id"),
					resource.TestCheckResourceAttrSet("data.tfe_github_app_installation.gha_installation", "id"),
				),
			},
		},
	})
}

func testAccTFEGHAInstallationDataSourceConfig_findByName(name string) string {
	return fmt.Sprintf(`
data "tfe_github_app_installation" "gha_installation" {
	name = "%s"
}
`, name)
}
