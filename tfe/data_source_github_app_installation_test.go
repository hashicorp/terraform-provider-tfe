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
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccTFEGHAInstallationDataSourcePreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEGHAInstallationDataSourceConfig(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"tfe_oauth_client.test", "api_url",
						"data.tfe_oauth_client.client", "api_url"),
					resource.TestCheckResourceAttrPair(
						"tfe_oauth_client.test", "http_url",
						"data.tfe_oauth_client.client", "http_url"),
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
				Config: testAccTFEGHAInstallationDataSourceConfig(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"tfe_oauth_client.test", "api_url",
						"data.tfe_oauth_client.client", "api_url"),
					resource.TestCheckResourceAttrPair(
						"tfe_oauth_client.test", "http_url",
						"data.tfe_oauth_client.client", "http_url"),
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
				Config: testAccTFEGHAInstallationDataSourceConfig(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"tfe_oauth_client.test", "api_url",
						"data.tfe_oauth_client.client", "api_url"),
					resource.TestCheckResourceAttrPair(
						"tfe_oauth_client.test", "http_url",
						"data.tfe_oauth_client.client", "http_url"),
				),
			},
		},
	})
}

func testAccTFEGHAInstallationDataSourceConfig(rInt int) string {
	return fmt.Sprintf(`
data "tfe_github_app_installation" "gha_installation" {
	id = "%s"
	name = "name-%d"
	installation_id = %d
}
`, envGithubAppInstallationId, rInt, rInt)
}
