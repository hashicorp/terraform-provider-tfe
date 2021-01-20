package tfe

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccTFEIPRangesDataSource_basic(t *testing.T) {
	ipRegex := regexp.MustCompile(`^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\/([1-9]|[1-9][0-9])$`)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
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
	return fmt.Sprintf(`data "tfe_ip_ranges" "ips" {}`)
}
