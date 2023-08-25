// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"math/rand"
	"regexp"
	"testing"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccTFEOrganizationToken_basic(t *testing.T) {
	token := &tfe.OrganizationToken{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEOrganizationTokenDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrganizationToken_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOrganizationTokenExists(
						"tfe_organization_token.foobar", token),
					resource.TestCheckResourceAttr(
						"tfe_organization_token.foobar", "organization", orgName),
				),
			},
		},
	})
}

func TestAccTFEOrganizationToken_existsWithoutForce(t *testing.T) {
	token := &tfe.OrganizationToken{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEOrganizationTokenDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrganizationToken_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOrganizationTokenExists(
						"tfe_organization_token.foobar", token),
					resource.TestCheckResourceAttr(
						"tfe_organization_token.foobar", "organization", orgName),
				),
			},

			{
				Config:      testAccTFEOrganizationToken_existsWithoutForce(rInt),
				ExpectError: regexp.MustCompile(`token already exists`),
			},
		},
	})
}

func TestAccTFEOrganizationToken_existsWithForce(t *testing.T) {
	token := &tfe.OrganizationToken{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEOrganizationTokenDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrganizationToken_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOrganizationTokenExists(
						"tfe_organization_token.foobar", token),
					resource.TestCheckResourceAttr(
						"tfe_organization_token.foobar", "organization", orgName),
				),
			},

			{
				Config: testAccTFEOrganizationToken_existsWithForce(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOrganizationTokenExists(
						"tfe_organization_token.regenerated", token),
					resource.TestCheckResourceAttr(
						"tfe_organization_token.regenerated", "organization", orgName),
				),
			},
		},
	})
}

func TestAccTFEOrganizationToken_withBlankExpiry(t *testing.T) {
	token := &tfe.OrganizationToken{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	expiredAt := ""

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEOrganizationTokenDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrganizationToken_withBlankExpiry(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOrganizationTokenExists(
						"tfe_organization_token.foobar", token),
					resource.TestCheckResourceAttr(
						"tfe_organization_token.foobar", "expired_at", expiredAt),
				),
			},
		},
	})
}

func TestAccTFEOrganizationToken_withValidExpiry(t *testing.T) {
	token := &tfe.OrganizationToken{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	expiredAt := "2051-04-11T23:15:59Z"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEOrganizationTokenDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrganizationToken_withValidExpiry(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOrganizationTokenExists(
						"tfe_organization_token.expiry", token),
					resource.TestCheckResourceAttr(
						"tfe_organization_token.expiry", "expired_at", expiredAt),
				),
			},
		},
	})
}

func TestAccTFEOrganizationToken_withInvalidExpiry(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEOrganizationTokenDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFEOrganizationToken_withInvalidExpiry(rInt),
				ExpectError: regexp.MustCompile(`must be a valid date or time, provided in iso8601 format`),
			},
		},
	})
}

func TestAccTFEOrganizationToken_import(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEOrganizationTokenDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrganizationToken_basic(rInt),
			},

			{
				ResourceName:            "tfe_organization_token.foobar",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"token"},
			},
		},
	})
}

func testAccCheckTFEOrganizationTokenExists(
	n string, token *tfe.OrganizationToken) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(ConfiguredClient)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		ot, err := config.Client.OrganizationTokens.Read(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}

		if ot == nil {
			return fmt.Errorf("OrganizationToken not found")
		}

		*token = *ot

		return nil
	}
}

func testAccCheckTFEOrganizationTokenDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(ConfiguredClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_organization_token" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := config.Client.OrganizationTokens.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("OrganizationToken %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccTFEOrganizationToken_basic(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_organization_token" "foobar" {
  organization = tfe_organization.foobar.id
}`, rInt)
}

// NOTE: This config is invalid because you cannot manage multiple tokens for
// one org. It is expected to always error.
func testAccTFEOrganizationToken_existsWithoutForce(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_organization_token" "foobar" {
  organization = tfe_organization.foobar.id
}

resource "tfe_organization_token" "error" {
  organization = tfe_organization.foobar.id
}`, rInt)
}

// NOTE: This config is invalid because you cannot manage multiple tokens for
// one org. It can run without error _once_ due to the presence of
// force_regenerate, but is expected to error on any subsequent run.
func testAccTFEOrganizationToken_existsWithForce(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_organization_token" "foobar" {
  organization = tfe_organization.foobar.id
}

resource "tfe_organization_token" "regenerated" {
  organization     = tfe_organization.foobar.id
  force_regenerate = true
}`, rInt)
}

func testAccTFEOrganizationToken_withBlankExpiry(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_organization_token" "foobar" {
  organization = tfe_organization.foobar.id
  expired_at = ""
}`, rInt)
}

func testAccTFEOrganizationToken_withValidExpiry(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_organization_token" "expiry" {
  organization  = tfe_organization.foobar.id
  expired_at 	= "2051-04-11T23:15:59Z"
}`, rInt)
}

func testAccTFEOrganizationToken_withInvalidExpiry(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_organization_token" "expiry" {
  organization  = tfe_organization.foobar.id
  expired_at 	= "2000-04-11"
}`, rInt)
}
