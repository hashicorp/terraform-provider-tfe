// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccTFESentinelPolicy_basic(t *testing.T) {
	policy := &tfe.Policy{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFESentinelPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFESentinelPolicy_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFESentinelPolicyExists(
						"tfe_sentinel_policy.foobar", policy),
					testAccCheckTFESentinelPolicyAttributes(policy),
					resource.TestCheckResourceAttr(
						"tfe_sentinel_policy.foobar", "name", "policy-test"),
					resource.TestCheckResourceAttr(
						"tfe_sentinel_policy.foobar", "description", "A test policy"),
					resource.TestCheckResourceAttr(
						"tfe_sentinel_policy.foobar", "policy", "main = rule { true }"),
					resource.TestCheckResourceAttr(
						"tfe_sentinel_policy.foobar", "enforce_mode", "hard-mandatory"),
				),
			},
		},
	})
}

func TestAccTFESentinelPolicy_update(t *testing.T) {
	policy := &tfe.Policy{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFESentinelPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFESentinelPolicy_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFESentinelPolicyExists(
						"tfe_sentinel_policy.foobar", policy),
					testAccCheckTFESentinelPolicyAttributes(policy),
					resource.TestCheckResourceAttr(
						"tfe_sentinel_policy.foobar", "name", "policy-test"),
					resource.TestCheckResourceAttr(
						"tfe_sentinel_policy.foobar", "description", "A test policy"),
					resource.TestCheckResourceAttr(
						"tfe_sentinel_policy.foobar", "policy", "main = rule { true }"),
					resource.TestCheckResourceAttr(
						"tfe_sentinel_policy.foobar", "enforce_mode", "hard-mandatory"),
				),
			},

			{
				Config: testAccTFESentinelPolicy_update(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFESentinelPolicyExists(
						"tfe_sentinel_policy.foobar", policy),
					testAccCheckTFESentinelPolicyAttributesUpdated(policy),
					resource.TestCheckResourceAttr(
						"tfe_sentinel_policy.foobar", "name", "policy-test"),
					resource.TestCheckResourceAttr(
						"tfe_sentinel_policy.foobar", "description", "An updated test policy"),
					resource.TestCheckResourceAttr(
						"tfe_sentinel_policy.foobar", "policy", "main = rule { false }"),
					resource.TestCheckResourceAttr(
						"tfe_sentinel_policy.foobar", "enforce_mode", "soft-mandatory"),
				),
			},
		},
	})
}

func TestAccTFESentinelPolicy_import(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFESentinelPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFESentinelPolicy_basic(rInt),
			},

			{
				ResourceName:        "tfe_sentinel_policy.foobar",
				ImportState:         true,
				ImportStateIdPrefix: fmt.Sprintf("tst-terraform-%d/", rInt),
				ImportStateVerify:   true,
			},
		},
	})
}

func testAccCheckTFESentinelPolicyExists(
	n string, policy *tfe.Policy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(ConfiguredClient)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		p, err := config.Client.Policies.Read(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}

		if p.ID != rs.Primary.ID {
			return fmt.Errorf("SentinelPolicy not found")
		}

		*policy = *p

		return nil
	}
}

func testAccCheckTFESentinelPolicyAttributes(
	policy *tfe.Policy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if policy.Name != "policy-test" {
			return fmt.Errorf("Bad name: %s", policy.Name)
		}

		if policy.Enforce[0].Mode != "hard-mandatory" {
			return fmt.Errorf("Bad enforce mode: %s", policy.Enforce[0].Mode)
		}

		return nil
	}
}

func testAccCheckTFESentinelPolicyAttributesUpdated(
	policy *tfe.Policy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if policy.Name != "policy-test" {
			return fmt.Errorf("Bad name: %s", policy.Name)
		}

		if policy.Enforce[0].Mode != "soft-mandatory" {
			return fmt.Errorf("Bad enforce mode: %s", policy.Enforce[0].Mode)
		}

		return nil
	}
}

func testAccCheckTFESentinelPolicyDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(ConfiguredClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_sentinel_policy" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := config.Client.Policies.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Sentinel policy %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccTFESentinelPolicy_basic(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_sentinel_policy" "foobar" {
  name         = "policy-test"
  description  = "A test policy"
  organization = tfe_organization.foobar.id
  policy       = "main = rule { true }"
  enforce_mode = "hard-mandatory"
}`, rInt)
}

func testAccTFESentinelPolicy_update(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_sentinel_policy" "foobar" {
  name         = "policy-test"
  description  = "An updated test policy"
  organization = tfe_organization.foobar.id
  policy       = "main = rule { false }"
  enforce_mode = "soft-mandatory"
}`, rInt)
}
