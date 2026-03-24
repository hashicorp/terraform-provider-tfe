// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccTFEOrgMaxTokenTTLPolicy_basic(t *testing.T) {
	skipUnlessBeta(t)
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrgMaxTokenTTLPolicy_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"tfe_org_max_token_ttl_policy.foobar", "enabled", "true"),
					resource.TestCheckResourceAttr(
						"tfe_org_max_token_ttl_policy.foobar", "org_token_max_ttl", "0.5h"),
					resource.TestCheckResourceAttr(
						"tfe_org_max_token_ttl_policy.foobar", "user_token_max_ttl", "2.5d"),
					resource.TestCheckResourceAttr(
						"tfe_org_max_token_ttl_policy.foobar", "team_token_max_ttl", "3w"),
					resource.TestCheckResourceAttr(
						"tfe_org_max_token_ttl_policy.foobar", "audit_trail_token_max_ttl", "6mo"),
				),
			},
		},
	})
}

func TestAccTFEOrgMaxTokenTTLPolicy_update(t *testing.T) {
	skipUnlessBeta(t)
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrgMaxTokenTTLPolicy_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"tfe_org_max_token_ttl_policy.foobar", "enabled", "true"),
					resource.TestCheckResourceAttr(
						"tfe_org_max_token_ttl_policy.foobar", "org_token_max_ttl", "0.5h"),
					resource.TestCheckResourceAttr(
						"tfe_org_max_token_ttl_policy.foobar", "user_token_max_ttl", "2.5d"),
					resource.TestCheckResourceAttr(
						"tfe_org_max_token_ttl_policy.foobar", "team_token_max_ttl", "3w"),
					resource.TestCheckResourceAttr(
						"tfe_org_max_token_ttl_policy.foobar", "audit_trail_token_max_ttl", "6mo"),
				),
			},
			{
				Config: testAccTFEOrgMaxTokenTTLPolicy_updated(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"tfe_org_max_token_ttl_policy.foobar", "enabled", "true"),
					resource.TestCheckResourceAttr(
						"tfe_org_max_token_ttl_policy.foobar", "org_token_max_ttl", "12h"),
					resource.TestCheckResourceAttr(
						"tfe_org_max_token_ttl_policy.foobar", "user_token_max_ttl", "10d"),
					resource.TestCheckResourceAttr(
						"tfe_org_max_token_ttl_policy.foobar", "team_token_max_ttl", "5w"),
					resource.TestCheckResourceAttr(
						"tfe_org_max_token_ttl_policy.foobar", "audit_trail_token_max_ttl", "12mo"),
				),
			},
		},
	})
}

func TestAccTFEOrgMaxTokenTTLPolicy_disabled(t *testing.T) {
	skipUnlessBeta(t)
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrgMaxTokenTTLPolicy_disabled(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"tfe_org_max_token_ttl_policy.foobar", "enabled", "false"),
				),
			},
		},
	})
}

func TestAccTFEOrgMaxTokenTTLPolicy_import(t *testing.T) {
	skipUnlessBeta(t)
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrgMaxTokenTTLPolicy_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"tfe_org_max_token_ttl_policy.foobar", "enabled", "true"),
				),
			},
			{
				ResourceName:      "tfe_org_max_token_ttl_policy.foobar",
				ImportState:       true,
				ImportStateId:     orgName,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTFEOrgMaxTokenTTLPolicy_defaultOrg(t *testing.T) {
	skipUnlessBeta(t)
	orgName, _ := setupDefaultOrganization(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrgMaxTokenTTLPolicy_defaultOrg(orgName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"tfe_org_max_token_ttl_policy.foobar", "enabled", "true"),
					resource.TestCheckResourceAttr(
						"tfe_org_max_token_ttl_policy.foobar", "organization", orgName),
					resource.TestCheckResourceAttr(
						"tfe_org_max_token_ttl_policy.foobar", "org_token_max_ttl", "1d"),
				),
			},
		},
	})
}

func testAccCheckTFEOrgMaxTokenTTLPolicyExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set for %s", n)
		}

		return nil
	}
}

func testAccTFEOrgMaxTokenTTLPolicy_basic(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
}

resource "tfe_org_max_token_ttl_policy" "foobar" {
  organization              = tfe_organization.foobar.name
  enabled                   = true
  org_token_max_ttl         = "0.5h"
  user_token_max_ttl        = "2.5d"
  team_token_max_ttl        = "3w"
  audit_trail_token_max_ttl = "6mo"
}
`, rInt)
}

func testAccTFEOrgMaxTokenTTLPolicy_updated(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
}

resource "tfe_org_max_token_ttl_policy" "foobar" {
  organization              = tfe_organization.foobar.name
  enabled                   = true
  org_token_max_ttl         = "12h"
  user_token_max_ttl        = "10d"
  team_token_max_ttl        = "5w"
  audit_trail_token_max_ttl = "12mo"
}
`, rInt)
}

func testAccTFEOrgMaxTokenTTLPolicy_disabled(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
}

resource "tfe_org_max_token_ttl_policy" "foobar" {
  organization = tfe_organization.foobar.name
  enabled      = false
}
`, rInt)
}

func testAccTFEOrgMaxTokenTTLPolicy_defaultOrg(orgName string) string {
	return fmt.Sprintf(`
resource "tfe_org_max_token_ttl_policy" "foobar" {
  organization       = "%s"
  enabled            = true
  org_token_max_ttl  = "1d"
}
`, orgName)
}
