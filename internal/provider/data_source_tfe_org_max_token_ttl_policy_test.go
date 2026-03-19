// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTFEOrgMaxTokenTTLPolicyDataSource_basic(t *testing.T) {
	skipUnlessBeta(t)
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrgMaxTokenTTLPolicyDataSourceConfig_basic(orgName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.tfe_org_max_token_ttl_policy.test", "organization", orgName),
					resource.TestCheckResourceAttrSet("data.tfe_org_max_token_ttl_policy.test", "enabled"),
					resource.TestCheckResourceAttrSet("data.tfe_org_max_token_ttl_policy.test", "org_token_max_ttl_ms"),
					resource.TestCheckResourceAttrSet("data.tfe_org_max_token_ttl_policy.test", "team_token_max_ttl_ms"),
					resource.TestCheckResourceAttrSet("data.tfe_org_max_token_ttl_policy.test", "audit_trail_token_max_ttl_ms"),
					resource.TestCheckResourceAttrSet("data.tfe_org_max_token_ttl_policy.test", "user_token_max_ttl_ms"),
				),
			},
		},
	})
}

func TestAccTFEOrgMaxTokenTTLPolicyDataSource_withResource(t *testing.T) {
	skipUnlessBeta(t)
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrgMaxTokenTTLPolicyDataSourceConfig_withResource(orgName),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Check resource
					resource.TestCheckResourceAttr("tfe_org_max_token_ttl_policy.test", "enabled", "true"),
					resource.TestCheckResourceAttr("tfe_org_max_token_ttl_policy.test", "org_token_max_ttl", "30d"),
					resource.TestCheckResourceAttr("tfe_org_max_token_ttl_policy.test", "team_token_max_ttl", "7d"),
					resource.TestCheckResourceAttr("tfe_org_max_token_ttl_policy.test", "user_token_max_ttl", "24h"),
					resource.TestCheckResourceAttr("tfe_org_max_token_ttl_policy.test", "audit_trail_token_max_ttl", "90d"),
					// Check data source returns milliseconds
					resource.TestCheckResourceAttr("data.tfe_org_max_token_ttl_policy.test", "enabled", "true"),
					resource.TestCheckResourceAttr("data.tfe_org_max_token_ttl_policy.test", "org_token_max_ttl_ms", "2592000000"),         // 30 days
					resource.TestCheckResourceAttr("data.tfe_org_max_token_ttl_policy.test", "team_token_max_ttl_ms", "604800000"),         // 7 days
					resource.TestCheckResourceAttr("data.tfe_org_max_token_ttl_policy.test", "user_token_max_ttl_ms", "86400000"),          // 24 hours
					resource.TestCheckResourceAttr("data.tfe_org_max_token_ttl_policy.test", "audit_trail_token_max_ttl_ms", "7776000000"), // 90 days
				),
			},
		},
	})
}

func testAccTFEOrgMaxTokenTTLPolicyDataSourceConfig_basic(orgName string) string {
	return fmt.Sprintf(`
resource "tfe_organization" "test" {
  name  = "%s"
}

data "tfe_org_max_token_ttl_policy" "test" {
  organization = tfe_organization.test.name
}
`, orgName)
}

func testAccTFEOrgMaxTokenTTLPolicyDataSourceConfig_withResource(orgName string) string {
	return fmt.Sprintf(`
resource "tfe_organization" "test" {
  name  = "%s"
}

resource "tfe_org_max_token_ttl_policy" "test" {
  organization              = tfe_organization.test.name
  enabled                   = true
  org_token_max_ttl         = "30d"
  team_token_max_ttl        = "7d"
  user_token_max_ttl        = "24h"
  audit_trail_token_max_ttl = "90d"
}

data "tfe_org_max_token_ttl_policy" "test" {
  organization = tfe_org_max_token_ttl_policy.test.organization
  depends_on   = [tfe_org_max_token_ttl_policy.test]
}
`, orgName)
}
