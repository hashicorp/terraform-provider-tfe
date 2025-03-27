// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTFEIPRangesDataSource_basic(t *testing.T) {
	skipIfEnterprise(t)
	ipRegex := regexp.MustCompile(`^([\d]{1,3}\.){3}[\d]{1,3}/[\d]{1,3}$`)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEIPRangesDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tfe_ip_ranges.ips", "id"),
					resource.TestCheckResourceAttrSet("data.tfe_ip_ranges.ips", "api.0"),
					resource.TestMatchResourceAttr("data.tfe_ip_ranges.ips", "api.0", ipRegex),
					resource.TestCheckResourceAttrSet("data.tfe_ip_ranges.ips", "notifications.0"),
					resource.TestMatchResourceAttr("data.tfe_ip_ranges.ips", "notifications.0", ipRegex),
					resource.TestCheckResourceAttrSet("data.tfe_ip_ranges.ips", "sentinel.0"),
					resource.TestMatchResourceAttr("data.tfe_ip_ranges.ips", "sentinel.0", ipRegex),
					resource.TestCheckResourceAttrSet("data.tfe_ip_ranges.ips", "vcs.0"),
					resource.TestMatchResourceAttr("data.tfe_ip_ranges.ips", "vcs.0", ipRegex),
				),
			},
		},
	})
}

func testAccTFEIPRangesDataSourceConfig() string {
	return `data "tfe_ip_ranges" "ips" {}`
}
