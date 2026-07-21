// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"math/rand"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/go-tfe/v2/api/models"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccTFETeamToken_basic(t *testing.T) {
	var token models.AuthenticationTokensable
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFETeamTokenDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamToken_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETeamTokenExists(
						"tfe_team_token.foobar", &token),
				),
			},
		},
	})
}

func TestAccTFETeamToken_multiple_team_tokens(t *testing.T) {
	skipUnlessBeta(t)
	var token models.AuthenticationTokensable
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFETeamTokenDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamToken_withMultipleTokens(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETeamTokenExists(
						"tfe_team_token.multi_token_1", &token),
					testAccCheckTFETeamTokenExists(
						"tfe_team_token.multi_token_2", &token),
					testAccCheckTFETeamTokenExists(
						"tfe_team_token.legacy", &token),
				),
			},
		},
	})
}

// TestAccTFETeamToken_createWithDescription is a regression test for a bug in the go-tfe
// v2 migration: createTeamToken's non-legacy branch (used whenever "description" is set)
// calls api.AuthenticationTokens().ById(teamID).Post(), which targets
// POST /authentication-tokens/{team_id}. Atlas only wires GET/DELETE for
// /authentication-tokens/{id} (see config/routes/v2.rb); the real "create a team token
// with a description" endpoint is the plural, team-nested
// POST /teams/{team_id}/authentication-tokens, which has no generated builder in the
// installed go-tfe/v2 client at all (see the openapi-atlas-verification skill for the
// full spec/route analysis). Isolated from TestAccTFETeamToken_multiple_team_tokens,
// which also exercises the unrelated legacy-token array/object response-shape bug in the
// same config.
//
// This test currently fails with a 404 against the wrong URL. It is expected to start
// passing once the create-with-description path calls a working endpoint (either an
// upstream go-tfe/v2 fix once /teams/{team_id}/authentication-tokens is promoted out of
// internal-beta, or a v1 fallback per the go-tfe-v2-migration skill's "Missing or
// unusable v2 coverage").
func TestAccTFETeamToken_createWithDescription(t *testing.T) {
	skipUnlessBeta(t)
	var token models.AuthenticationTokensable
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	description := fmt.Sprintf("tst-terraform-%d-token", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFETeamTokenDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamToken_createWithDescription(rInt, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETeamTokenExists(
						"tfe_team_token.described", &token),
					resource.TestCheckResourceAttr(
						"tfe_team_token.described", "description", description),
				),
			},
		},
	})
}

func TestAccTFETeamToken_existsWithoutForce(t *testing.T) {
	var token models.AuthenticationTokensable
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFETeamTokenDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamToken_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETeamTokenExists(
						"tfe_team_token.foobar", &token),
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
	var token models.AuthenticationTokensable
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFETeamTokenDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamToken_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETeamTokenExists(
						"tfe_team_token.foobar", &token),
				),
			},

			{
				Config: testAccTFETeamToken_existsWithForce(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETeamTokenExists(
						"tfe_team_token.regenerated", &token),
				),
			},
		},
	})
}

func TestAccTFETeamToken_invalidWithForceGenerateAndDescription(t *testing.T) {
	skipUnlessBeta(t)
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFETeamTokenDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFETeamToken_WithForceGenerateAndDescription(rInt),
				ExpectError: regexp.MustCompile(`"force_regenerate" cannot be specified when "description"`),
			},
		},
	})
}

func TestAccTFETeamToken_withBlankExpiry(t *testing.T) {
	skipUnlessBeta(t)
	var token models.AuthenticationTokensable
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFETeamTokenDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamToken_withBlankExpiry(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETeamTokenExists(
						"tfe_team_token.foobar", &token),
					// When expired_at is not provided, API sets default (24 months)
					// We now read this value from the API response
					resource.TestCheckResourceAttrSet(
						"tfe_team_token.foobar", "expired_at"),
				),
			},
		},
	})
}

func TestAccTFETeamToken_withValidExpiry(t *testing.T) {
	var token models.AuthenticationTokensable
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	expiredAt := "2051-04-11T23:15:59Z"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFETeamTokenDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamToken_withValidExpiry(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETeamTokenExists(
						"tfe_team_token.expiry", &token),
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
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFETeamTokenDestroy,
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
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFETeamTokenDestroy,
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

func TestAccTFETeamToken_importByTokenID(t *testing.T) {
	skipUnlessBeta(t)
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFETeamTokenDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamToken_withMultipleTokens(rInt),
			},
			{
				ResourceName:            "tfe_team_token.multi_token_1",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"token"},
			},
			{
				ResourceName:            "tfe_team_token.multi_token_2",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"token"},
			},
			{
				ResourceName:            "tfe_team_token.legacy",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"token"},
			},
		},
	})
}

func TestAccTFETeamToken_withNonexistentTeam(t *testing.T) {
	conf := `
resource "tfe_team_token" "invalid" {
  team_id    = "invalid"
}`
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFETeamTokenDestroy,
		Steps: []resource.TestStep{
			{
				Config: conf,
				// Terraform's CLI wraps long diagnostic text at a column width, and the
				// wrap point (a literal newline replacing a space) shifts based on the
				// surrounding text's exact length. Match word-by-word with \s+ instead of
				// literal spaces so this doesn't depend on exactly where the wrap lands.
				ExpectError: regexp.MustCompile(`(?s)team\s+does\s+not\s+exist\s+or\s+version\s+of\s+Terraform\s+Enterprise\s+does\s+not\s+support\s+multiple\s+team\s+tokens\s+with\s+descriptions`),
			},
		},
	})
}

func testAccCheckTFETeamTokenExists(
	n string, token *models.AuthenticationTokensable) resource.TestCheckFunc { //nolint:gocritic // token is an output param the caller reads after Check runs; AuthenticationTokensable must stay addressable to be settable
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		var tt models.AuthenticationTokensEnvelopeable
		var err error
		if isTokenID(rs.Primary.ID) {
			tt, err = testAccConfiguredClient.ClientV2.API.AuthenticationTokens().ById(rs.Primary.ID).Get(ctx, nil)
		} else {
			tt, err = testAccConfiguredClient.ClientV2.API.Teams().ById(rs.Primary.ID).AuthenticationToken().Get(ctx, nil)
		}

		if err != nil {
			return err
		}

		if tt == nil || tt.GetData() == nil {
			return fmt.Errorf("Team token not found")
		}

		*token = tt.GetData()

		return nil
	}
}

func testAccCheckTFETeamTokenDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_team_token" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		var err error
		if isTokenID(rs.Primary.ID) {
			_, err = testAccConfiguredClient.ClientV2.API.AuthenticationTokens().ById(rs.Primary.ID).Get(ctx, nil)
		} else {
			_, err = testAccConfiguredClient.ClientV2.API.Teams().ById(rs.Primary.ID).AuthenticationToken().Get(ctx, nil)
		}
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

func testAccTFETeamToken_createWithDescription(rInt int, description string) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_team_token" "described" {
  team_id     = tfe_team.foobar.id
  description = "%s"
}`, rInt, description)
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
  # expired_at not provided - API will set default to 24 months
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

func testAccTFETeamToken_withMultipleTokens(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = tfe_organization.foobar.id
}


resource "tfe_team_token" "multi_token_1" {
  team_id     = tfe_team.foobar.id
  description = "tst-terraform-%d-token-1"
  expired_at  = "2051-04-11T23:15:59Z"
}

resource "tfe_team_token" "multi_token_2" {
  team_id    = tfe_team.foobar.id
  description = "tst-terraform-%d-token-2"
}

resource "tfe_team_token" "legacy" {
  team_id    = tfe_team.foobar.id
}`, rInt, rInt, rInt)
}

func testAccTFETeamToken_WithForceGenerateAndDescription(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = tfe_organization.foobar.id
}


resource "tfe_team_token" "invalid" {
  team_id     = tfe_team.foobar.id
  description = "tst-terraform-%d-token"
  force_regenerate = true
}`, rInt, rInt)
}
