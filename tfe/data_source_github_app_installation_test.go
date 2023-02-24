// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tfe

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"math/rand"
	"testing"
	"time"
)

func testAccTFEGHAInstallationDataSourcePreCheck(t *testing.T) {
	testAccPreCheck(t)
	if envGithubAppInstallationId == "" {
		t.Skip("Please set GITHUB_APP_INSTALLATION_ID to run this test")
	}
}

func TestAccTFEGHAInstallationDataSource_findByID(t *testing.T) {
	//rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccTFEGHAInstallationDataSourcePreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEGHAInstallationDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.tfe_github_app_installation.gha_installation", "installation_id", envGithubAppInstallationId),
					resource.TestCheckResourceAttrSet("data.tfe_github_app_installation.gha_installation", "id"),
					resource.TestCheckResourceAttrSet("data.tfe_github_app_installation.gha_installation", "name"),
				),
			},
		},
	})
}

func TestAccTFEGHAInstallationDataSource_findByName(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccTFEGHAInstallationDataSourcePreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEGHAInstallationDataSourceConfig_findByName(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					//resource.TestCheckResourceAttrSet("data.tfe_github_app_installation.gha_installation", "installation_id"),
					//resource.TestCheckResourceAttrSet("data.tfe_github_app_installation.gha_installation", "id"),
					resource.TestCheckResourceAttr("data.tfe_github_app_installation.gha_installation", "name", fmt.Sprintf("name-%d", rInt)),
				),
			},
		},
	})
}

func TestAccTFEGHAInstallationDataSource_findByInstallationID(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccTFEGHAInstallationDataSourcePreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEGHAInstallationDataSourceConfig_findByInstallationID(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.tfe_github_app_installation.gha_installation", "installation_id", envGithubAppInstallationId),
					//resource.TestCheckResourceAttrSet("data.tfe_github_app_installation.gha_installation", "id"),
					//resource.TestCheckResourceAttrSet("data.tfe_github_app_installation.gha_installation", "name"),
				),
			},
		},
	})
}

func testAccTFEGHAInstallationDataSourceConfig() string {
	return fmt.Sprintf(`
data "tfe_github_app_installation" "gha_installation" {
	id = "%s"
}
`, envGithubAppInstallationId)
}

func testAccTFEGHAInstallationDataSourceConfig_findByName(rInt int) string {
	return fmt.Sprintf(`
data "tfe_github_app_installation" "gha_installation" {
	name = "name-%d"
}
`, rInt)
}

func testAccTFEGHAInstallationDataSourceConfig_findByInstallationID(rInt int) string {
	return fmt.Sprintf(`
data "tfe_github_app_installation" "gha_installation" {
	installation_id = %d
}
`, rInt)
}
