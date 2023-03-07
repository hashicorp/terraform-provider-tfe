// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tfe

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func testAccTFEGHAInstallationDataSourcePreCheck(t *testing.T) {
	testAccPreCheck(t)
	fmt.Printf("envGithubAppInstallationID is %s\n", envGithubAppInstallationID)
	if envGithubAppInstallationID == "" {
		t.Skip("Please set GITHUB_APP_INSTALLATION_ID to run this test")
	}
}

func TestAccTFEGHAInstallationDataSource_findByID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccTFEGHAInstallationDataSourcePreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEGHAInstallationDataSourceConfig(envGithubAppInstallationID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.tfe_github_app_installation.gha_installation", "id", envGithubAppInstallationID),
					resource.TestCheckResourceAttrSet("data.tfe_github_app_installation.gha_installation", "installation_id"),
					resource.TestCheckResourceAttrSet("data.tfe_github_app_installation.gha_installation", "name"),
				),
			},
		},
	})
}

func TestAccTFEGHAInstallationDataSource_findByName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccTFEGHAInstallationDataSourcePreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEGHAInstallationDataSourceConfig_findByName("installation-name"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.tfe_github_app_installation.gha_installation", "id", envGithubAppInstallationID),
					resource.TestCheckResourceAttrSet("data.tfe_github_app_installation.gha_installation", "installation_id"),
					resource.TestCheckResourceAttrSet("data.tfe_github_app_installation.gha_installation", "name"),
				),
			},
		},
	})
}

func TestAccTFEGHAInstallationDataSource_findByInstallationID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccTFEGHAInstallationDataSourcePreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEGHAInstallationDataSourceConfig_findByInstallationID(12345),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.tfe_github_app_installation.gha_installation", "id", envGithubAppInstallationID),
					resource.TestCheckResourceAttrSet("data.tfe_github_app_installation.gha_installation", "installation_id"),
					resource.TestCheckResourceAttrSet("data.tfe_github_app_installation.gha_installation", "name"),
				),
			},
		},
	})
}

func testAccTFEGHAInstallationDataSourceConfig(envGithubAppInstallationID string) string {
	return fmt.Sprintf(`
data "tfe_github_app_installation" "gha_installation" {
	id = "%s"
}
`, envGithubAppInstallationID)
}

func testAccTFEGHAInstallationDataSourceConfig_findByName(name string) string {
	return fmt.Sprintf(`
data "tfe_github_app_installation" "gha_installation" {
	name = "%s"
}
`, name)
}

func testAccTFEGHAInstallationDataSourceConfig_findByInstallationID(installationID int) string {
	return fmt.Sprintf(`
data "tfe_github_app_installation" "gha_installation" {
	installation_id = %d
}
`, installationID)
}
