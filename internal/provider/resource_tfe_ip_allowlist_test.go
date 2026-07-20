// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	tfev2 "github.com/hashicorp/go-tfe/v2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/hashicorp/terraform-provider-tfe/internal/client"
)

func getClientV2UsingEnv(t *testing.T) *tfev2.Client {
	t.Helper()

	hostname := client.DefaultHostname
	if os.Getenv("TFE_HOSTNAME") != "" {
		hostname = os.Getenv("TFE_HOSTNAME")
	}
	token := os.Getenv("TFE_TOKEN")

	providerClient, err := client.GetClient(hostname, token, defaultSSLSkipVerify)
	if err != nil {
		t.Fatalf("error getting v2 client: %s", err)
	}
	return providerClient.TFEClientV2
}

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
		CheckDestroy:             testAccCheckTFEIPAllowlistDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: testAccTFEIPAllowlist_basic(org.Name, rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTFEIPAllowlistExists(t, "tfe_ip_allowlist.foobar"),
					resource.TestCheckResourceAttr(
						"tfe_ip_allowlist.foobar", "name", fmt.Sprintf("allowlist-%d", rInt)),
					resource.TestCheckResourceAttr(
						"tfe_ip_allowlist.foobar", "enforcement_scope", "organization"),
					resource.TestCheckResourceAttr(
						"tfe_ip_allowlist.foobar", "cidr_range.#", "1"),
				),
			},
			{
				Config: testAccTFEIPAllowlist_update(org.Name, rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTFEIPAllowlistExists(t, "tfe_ip_allowlist.foobar"),
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
					testAccCheckTFEIPAllowlistExists(t, "tfe_ip_allowlist.foobar"),
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

func testAccCheckTFEIPAllowlistExists(t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		clientV2 := getClientV2UsingEnv(t)
		_, err := clientV2.API.CidrRangeLists().ByCidr_range_list_id(rs.Primary.ID).Get(context.Background(), nil)
		if err != nil {
			return fmt.Errorf("error reading IP allowlist %s: %w", rs.Primary.ID, err)
		}
		return nil
	}
}

func testAccCheckTFEIPAllowlistDestroy(t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		clientV2 := getClientV2UsingEnv(t)

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

func testAccTFEIPAllowlist_basic(orgName string, rInt int) string {
	return fmt.Sprintf(`
resource "tfe_ip_allowlist" "foobar" {
  organization      = "%s"
  name              = "allowlist-%d"
  description       = "a test allowlist"
  enforcement_scope = "organization"

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
  enforcement_scope = "organization"

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
  enforcement_scope = "organization"

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
