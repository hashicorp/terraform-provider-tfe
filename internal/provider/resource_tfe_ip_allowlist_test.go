// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccTFEIPAllowlist_basic(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEIPAllowlistDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccTFEIPAllowlist_basic(org.Name, rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTFEIPAllowlistExists("tfe_ip_allowlist.foobar"),
					resource.TestCheckResourceAttr(
						"tfe_ip_allowlist.foobar", "name", fmt.Sprintf("allowlist-%d", rInt)),
					resource.TestCheckResourceAttr(
						"tfe_ip_allowlist.foobar", "enforcement_scope", "all_agent_pools"),
					resource.TestCheckResourceAttr(
						"tfe_ip_allowlist.foobar", "cidr_range.#", "1"),
				),
			},
			{
				Config: testAccTFEIPAllowlist_update(org.Name, rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTFEIPAllowlistExists("tfe_ip_allowlist.foobar"),
					resource.TestCheckResourceAttr(
						"tfe_ip_allowlist.foobar", "description", "updated description"),
					resource.TestCheckResourceAttr(
						"tfe_ip_allowlist.foobar", "cidr_range.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"tfe_ip_allowlist.foobar", "cidr_range.*", map[string]string{
							"range":       "192.168.1.0/24",
							"description": "vpn",
							"enabled":     "false",
						}),
				),
			},
			{
				// Toggling `enabled` on a range must not drop the descriptions of
				// any range (regression test for the SetNestedAttribute computed
				// default dropping sibling descriptions).
				Config: testAccTFEIPAllowlist_toggleEnabled(org.Name, rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTFEIPAllowlistExists("tfe_ip_allowlist.foobar"),
					resource.TestCheckResourceAttr(
						"tfe_ip_allowlist.foobar", "cidr_range.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"tfe_ip_allowlist.foobar", "cidr_range.*", map[string]string{
							"range":       "10.0.0.0/24",
							"description": "office",
							"enabled":     "true",
						}),
					resource.TestCheckTypeSetElemNestedAttrs(
						"tfe_ip_allowlist.foobar", "cidr_range.*", map[string]string{
							"range":       "192.168.1.0/24",
							"description": "vpn",
							"enabled":     "true",
						}),
				),
			},
			{
				ResourceName:      "tfe_ip_allowlist.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// selected_agent_pools scope exercises the agent-pools relationship that is
// written as part of the create and update requests. The first step creates the
// list assigned to a single agent pool; the second step reassigns it to a
// different pool, which drives the Update path that sets the agent-pools
// relationship in the PATCH body.
func TestAccTFEIPAllowlist_selectedAgentPools(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEIPAllowlistDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccTFEIPAllowlist_selectedAgentPools(org.Name, rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTFEIPAllowlistExists("tfe_ip_allowlist.foobar"),
					resource.TestCheckResourceAttr(
						"tfe_ip_allowlist.foobar", "enforcement_scope", "selected_agent_pools"),
					resource.TestCheckResourceAttr(
						"tfe_ip_allowlist.foobar", "agent_pool_ids.#", "1"),
					resource.TestCheckResourceAttrPair(
						"tfe_ip_allowlist.foobar", "agent_pool_ids.0",
						"tfe_agent_pool.foobar", "id"),
				),
			},
			{
				Config: testAccTFEIPAllowlist_selectedAgentPoolsUpdate(org.Name, rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTFEIPAllowlistExists("tfe_ip_allowlist.foobar"),
					resource.TestCheckResourceAttr(
						"tfe_ip_allowlist.foobar", "enforcement_scope", "selected_agent_pools"),
					resource.TestCheckResourceAttr(
						"tfe_ip_allowlist.foobar", "agent_pool_ids.#", "1"),
					resource.TestCheckResourceAttrPair(
						"tfe_ip_allowlist.foobar", "agent_pool_ids.0",
						"tfe_agent_pool.barbaz", "id"),
				),
			},
			{
				ResourceName:      "tfe_ip_allowlist.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckTFEIPAllowlistExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		clientV2, err := getClientV2UsingEnv()
		if err != nil {
			return err
		}
		_, err = clientV2.API.CidrRangeLists().ByCidr_range_list_id(rs.Primary.ID).Get(context.Background(), nil)
		if err != nil {
			return fmt.Errorf("error reading IP allowlist %s: %w", rs.Primary.ID, err)
		}
		return nil
	}
}

func testAccCheckTFEIPAllowlistDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		clientV2, err := getClientV2UsingEnv()
		if err != nil {
			return err
		}

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "tfe_ip_allowlist" {
				continue
			}
			if rs.Primary.ID == "" {
				continue
			}

			_, err := clientV2.API.CidrRangeLists().ByCidr_range_list_id(rs.Primary.ID).Get(context.Background(), nil)
			if err == nil {
				return fmt.Errorf("IP allowlist %s still exists", rs.Primary.ID)
			}
			if !isV2ResourceNotFound(err) {
				return fmt.Errorf("unexpected error checking IP allowlist %s: %w", rs.Primary.ID, err)
			}
		}
		return nil
	}
}

// These tests intentionally use the "all_agent_pools" enforcement scope. Only
// an "organization"-scoped list is treated as the org-wide IP allowlist, which
// atlas enforces against the caller's IP on every API authorization (owners are
// exempt only via the UI, not the API). Creating an "organization"-scoped list
// that does not contain the test runner's IP would immediately lock the test
// out of the org (HTTP 404 on all subsequent requests). "all_agent_pools" only
// affects agent connections, so it never gates the test's own API calls.
func testAccTFEIPAllowlist_basic(orgName string, rInt int) string {
	return fmt.Sprintf(`
resource "tfe_ip_allowlist" "foobar" {
  organization      = "%s"
  name              = "allowlist-%d"
  description       = "a test allowlist"
  enforcement_scope = "all_agent_pools"

  cidr_range = [
    {
      range       = "10.0.0.0/24"
      description = "office"
    },
  ]
}
`, orgName, rInt)
}

func testAccTFEIPAllowlist_update(orgName string, rInt int) string {
	return fmt.Sprintf(`
resource "tfe_ip_allowlist" "foobar" {
  organization      = "%s"
  name              = "allowlist-%d"
  description       = "updated description"
  enforcement_scope = "all_agent_pools"

  cidr_range = [
    {
      range       = "10.0.0.0/24"
      description = "office"
    },
    {
      range       = "192.168.1.0/24"
      description = "vpn"
      enabled     = false
    },
  ]
}
`, orgName, rInt)
}

// testAccTFEIPAllowlist_toggleEnabled is identical to _update except the
// 192.168.1.0/24 range is enabled. The descriptions are unchanged, so this
// exercises flipping `enabled` without dropping sibling descriptions.
func testAccTFEIPAllowlist_toggleEnabled(orgName string, rInt int) string {
	return fmt.Sprintf(`
resource "tfe_ip_allowlist" "foobar" {
  organization      = "%s"
  name              = "allowlist-%d"
  description       = "updated description"
  enforcement_scope = "all_agent_pools"

  cidr_range = [
    {
      range       = "10.0.0.0/24"
      description = "office"
    },
    {
      range       = "192.168.1.0/24"
      description = "vpn"
      enabled     = true
    },
  ]
}
`, orgName, rInt)
}

// testAccTFEIPAllowlist_selectedAgentPools creates a "selected_agent_pools"
// scoped allowlist assigned to a single agent pool. This scope only affects
// agent connections, so (like "all_agent_pools") it never gates the test
// runner's own API calls the way an "organization"-scoped list would.
func testAccTFEIPAllowlist_selectedAgentPools(orgName string, rInt int) string {
	return fmt.Sprintf(`
resource "tfe_agent_pool" "foobar" {
  name                = "agent-pool-foobar-%d"
  organization        = "%s"
  organization_scoped = false
}

resource "tfe_agent_pool" "barbaz" {
  name                = "agent-pool-barbaz-%d"
  organization        = "%s"
  organization_scoped = false
}

resource "tfe_ip_allowlist" "foobar" {
  organization      = "%s"
  name              = "allowlist-%d"
  description       = "a test allowlist"
  enforcement_scope = "selected_agent_pools"
  agent_pool_ids    = [tfe_agent_pool.foobar.id]

  cidr_range = [
    {
      range       = "10.0.0.0/24"
      description = "office"
    },
  ]
}
`, rInt, orgName, rInt, orgName, orgName, rInt)
}

// testAccTFEIPAllowlist_selectedAgentPoolsUpdate reassigns the allowlist from
// tfe_agent_pool.foobar to tfe_agent_pool.barbaz, exercising the Update path
// that writes the full agent-pools relationship in the PATCH body.
func testAccTFEIPAllowlist_selectedAgentPoolsUpdate(orgName string, rInt int) string {
	return fmt.Sprintf(`
resource "tfe_agent_pool" "foobar" {
  name                = "agent-pool-foobar-%d"
  organization        = "%s"
  organization_scoped = false
}

resource "tfe_agent_pool" "barbaz" {
  name                = "agent-pool-barbaz-%d"
  organization        = "%s"
  organization_scoped = false
}

resource "tfe_ip_allowlist" "foobar" {
  organization      = "%s"
  name              = "allowlist-%d"
  description       = "a test allowlist"
  enforcement_scope = "selected_agent_pools"
  agent_pool_ids    = [tfe_agent_pool.barbaz.id]

  cidr_range = [
    {
      range       = "10.0.0.0/24"
      description = "office"
    },
  ]
}
`, rInt, orgName, rInt, orgName, orgName, rInt)
}
