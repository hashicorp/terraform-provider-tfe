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

func TestAccTFETeamToken_basic(t *testing.T) {
	token := &tfe.TeamToken{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFETeamTokenDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamToken_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETeamTokenExists(
						"tfe_team_token.foobar", token),
				),
			},
		},
	})
}

func TestAccTFETeamToken_existsWithoutForce(t *testing.T) {
	token := &tfe.TeamToken{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFETeamTokenDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamToken_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETeamTokenExists(
						"tfe_team_token.foobar", token),
				),
			},

			{
				Config:      testAccTFETeamToken_existsWithoutForce(rInt),
				ExpectError: regexp.MustCompile(`token already exists`),
			},
		},
	})
}

func TestAccTFETeamToken_existsWithForce(t *testing.T) {
	token := &tfe.TeamToken{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFETeamTokenDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamToken_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETeamTokenExists(
						"tfe_team_token.foobar", token),
				),
			},

			{
				Config: testAccTFETeamToken_existsWithForce(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETeamTokenExists(
						"tfe_team_token.regenerated", token),
				),
			},
		},
	})
}

func TestAccTFETeamToken_withBlankExpiry(t *testing.T) {
	token := &tfe.TeamToken{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	expiredAt := ""

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFETeamTokenDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamToken_withBlankExpiry(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETeamTokenExists(
						"tfe_team_token.foobar", token),
					resource.TestCheckResourceAttr(
						"tfe_team_token.foobar", "expired_at", expiredAt),
				),
			},
		},
	})
}

func TestAccTFETeamToken_withValidExpiry(t *testing.T) {
	token := &tfe.TeamToken{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	expiredAt := "2051-04-11T23:15:59Z"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFETeamTokenDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamToken_withValidExpiry(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETeamTokenExists(
						"tfe_team_token.expiry", token),
					resource.TestCheckResourceAttr(
						"tfe_team_token.expiry", "expired_at", expiredAt),
				),
			},
		},
	})
}

func TestAccTFETeamToken_withInvalidExpiry(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFETeamTokenDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFETeamToken_withInvalidExpiry(rInt),
				ExpectError: regexp.MustCompile(`must be a valid date or time, provided in iso8601 format`),
			},
		},
	})
}

func TestAccTFETeamToken_import(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFETeamTokenDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamToken_basic(rInt),
			},

			{
				ResourceName:            "tfe_team_token.foobar",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"token"},
			},
		},
	})
}

func testAccCheckTFETeamTokenExists(
	n string, token *tfe.TeamToken) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(ConfiguredClient)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		tt, err := config.Client.TeamTokens.Read(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}

		if tt == nil {
			return fmt.Errorf("Team token not found")
		}

		*token = *tt

		return nil
	}
}

func testAccCheckTFETeamTokenDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(ConfiguredClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_team_token" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := config.Client.TeamTokens.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Team token %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccTFETeamToken_basic(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_team_token" "foobar" {
  team_id = tfe_team.foobar.id
}`, rInt)
}

// NOTE: This config is invalid because you cannot manage multiple tokens for
// one team. It is expected to always error.
func testAccTFETeamToken_existsWithoutForce(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_team_token" "foobar" {
  team_id = tfe_team.foobar.id
}

resource "tfe_team_token" "error" {
  team_id = tfe_team.foobar.id
}`, rInt)
}

// NOTE: This config is invalid because you cannot manage multiple tokens for
// one team. It can run without error _once_ due to the presence of
// force_regenerate, but is expected to error on any subsequent run.
func testAccTFETeamToken_existsWithForce(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_team_token" "foobar" {
  team_id = tfe_team.foobar.id
}

resource "tfe_team_token" "regenerated" {
  team_id          = tfe_team.foobar.id
  force_regenerate = true
}`, rInt)
}

func testAccTFETeamToken_withBlankExpiry(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_team_token" "foobar" {
  team_id = tfe_team.foobar.id
  expired_at = ""
}`, rInt)
}

func testAccTFETeamToken_withValidExpiry(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_team_token" "expiry" {
  team_id    = tfe_team.foobar.id
  expired_at = "2051-04-11T23:15:59Z"
}`, rInt)
}

func testAccTFETeamToken_withInvalidExpiry(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_team_token" "expiry" {
  team_id    = tfe_team.foobar.id
  expired_at = "2000-04-11"
}`, rInt)
}
