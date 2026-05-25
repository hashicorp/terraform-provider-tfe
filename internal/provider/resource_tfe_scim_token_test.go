// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"errors"
	"fmt"
	"regexp"
	"testing"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestAccTFESCIMToken_omnibus is the single test function for all SCIM token
// acceptance tests.
//
// FLAKE ALERT: SCIM settings are a singleton resource shared by the entire TFE
// instance. Every sub-test here enables SCIM (via an inline tfe_scim_settings
// block) as a prerequisite. Running all cases inside one function — without
// calling t.Parallel in any sub-test — prevents concurrent tests from racing
// over the same singleton state.
//
// FLAKE ALERT (dual-singleton): This suite also contends with
// resource_tfe_saml_settings_test.go for the SAML singleton. Both singletons
// must be treated as exclusive resources: do not run SCIM and SAML acceptance
// tests concurrently.
//
// Should this test name ever change, you will also need to update the regex in ci.yml.
func TestAccTFESCIMToken_omnibus(t *testing.T) {
	skipIfCloud(t)

	t.Run("basic create read delete", func(t *testing.T) {
		description := "tf-acc-test-scim-token-" + randomString(t)

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			CheckDestroy:             testAccTFESCIMTokenDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccTFESCIMToken_basic(description),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("tfe_scim_token.test", "id"),
						resource.TestCheckResourceAttr("tfe_scim_token.test", "description", description),
						// token is only returned on create
						resource.TestCheckResourceAttrSet("tfe_scim_token.test", "token"),
						// expired_at defaults to 365 days when not set
						resource.TestCheckResourceAttrSet("tfe_scim_token.test", "expired_at"),
						resource.TestCheckResourceAttrSet("tfe_scim_token.test", "created_at"),
					),
				},
				// Re-apply must be a no-op.
				{
					Config:   testAccTFESCIMToken_basic(description),
					PlanOnly: true,
				},
			},
		})
	})

	t.Run("explicit expired_at is preserved across reads", func(t *testing.T) {
		description := "tf-acc-test-scim-token-exp-" + randomString(t)
		// stay under the API's 365-day max
		expiredAt := time.Now().UTC().Add(364 * 24 * time.Hour).Truncate(time.Second).Format(time.RFC3339)

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			CheckDestroy:             testAccTFESCIMTokenDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccTFESCIMToken_withExpiredAt(description, expiredAt),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("tfe_scim_token.test", "id"),
						resource.TestCheckResourceAttr("tfe_scim_token.test", "description", description),
						resource.TestCheckResourceAttrSet("tfe_scim_token.test", "token"),
						resource.TestCheckResourceAttr("tfe_scim_token.test", "expired_at", expiredAt),
					),
				},
				// no perpetual diff
				{
					Config:   testAccTFESCIMToken_withExpiredAt(description, expiredAt),
					PlanOnly: true,
				},
			},
		})
	})

	t.Run("description change triggers resource replacement", func(t *testing.T) {
		descriptionA := "tf-acc-test-scim-token-a-" + randomString(t)
		descriptionB := "tf-acc-test-scim-token-b-" + randomString(t)

		var firstTokenID string

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			CheckDestroy:             testAccTFESCIMTokenDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccTFESCIMToken_basic(descriptionA),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("tfe_scim_token.test", "description", descriptionA),
						func(s *terraform.State) error {
							rs, ok := s.RootModule().Resources["tfe_scim_token.test"]
							if !ok {
								return fmt.Errorf("tfe_scim_token.test not found in state")
							}
							firstTokenID = rs.Primary.ID
							return nil
						},
					),
				},
				// description change => RequiresReplace, ID must change
				{
					Config: testAccTFESCIMToken_basic(descriptionB),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("tfe_scim_token.test", "description", descriptionB),
						resource.TestCheckResourceAttrSet("tfe_scim_token.test", "token"),
						func(s *terraform.State) error {
							rs, ok := s.RootModule().Resources["tfe_scim_token.test"]
							if !ok {
								return fmt.Errorf("tfe_scim_token.test not found in state")
							}
							if rs.Primary.ID == firstTokenID {
								return fmt.Errorf("expected resource ID to change after description update (RequiresReplace), but it remained %s", firstTokenID)
							}
							return nil
						},
					),
				},
			},
		})
	})

	t.Run("import sets token to null", func(t *testing.T) {
		description := "tf-acc-test-scim-token-import-" + randomString(t)

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			CheckDestroy:             testAccTFESCIMTokenDestroy,
			Steps: []resource.TestStep{
				// Create a token to import.
				{
					Config: testAccTFESCIMToken_basic(description),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("tfe_scim_token.test", "token"),
					),
				},
				// Import by ID. The Read endpoint never returns the token value, so
				// ignore it; everything else should round-trip.
				{
					ResourceName:            "tfe_scim_token.test",
					ImportState:             true,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: []string{"token"},
				},
				// token is null post-import; confirm no perpetual diff
				{
					Config:   testAccTFESCIMToken_basic(description),
					PlanOnly: true,
				},
			},
		})
	})

	t.Run("import with invalid id is rejected", func(t *testing.T) {
		description := "tf-acc-test-scim-token-badimp-" + randomString(t)

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			CheckDestroy:             testAccTFESCIMTokenDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccTFESCIMToken_basic(description),
				},
				// IDs must start with "at-"
				{
					ResourceName:  "tfe_scim_token.test",
					ImportState:   true,
					ImportStateId: "not-a-valid-token-id",
					ExpectError:   regexp.MustCompile(`Unexpected Import Identifier`),
				},
			},
		})
	})

	t.Run("missing description is rejected at validate time", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			Steps: []resource.TestStep{
				{
					Config:      testAccTFESCIMToken_missingDescription(),
					ExpectError: regexp.MustCompile(`(?s)"description" is required|Missing required argument`),
					PlanOnly:    true,
				},
			},
		})
	})

	t.Run("invalid expired_at is rejected at create time", func(t *testing.T) {
		description := "tf-acc-test-scim-token-badexp-" + randomString(t)

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			CheckDestroy:             testAccTFESCIMTokenDestroy,
			Steps: []resource.TestStep{
				{
					Config:      testAccTFESCIMToken_withExpiredAt(description, "not-a-timestamp"),
					ExpectError: regexp.MustCompile(`must be a valid date or time`),
				},
			},
		})
	})

	t.Run("expired_at change triggers resource replacement", func(t *testing.T) {
		description := "tf-acc-test-scim-token-reexp-" + randomString(t)
		expiredAtA := time.Now().UTC().Add(30 * 24 * time.Hour).Truncate(time.Second).Format(time.RFC3339)
		expiredAtB := time.Now().UTC().Add(60 * 24 * time.Hour).Truncate(time.Second).Format(time.RFC3339)

		var firstTokenID string

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			CheckDestroy:             testAccTFESCIMTokenDestroy,
			Steps: []resource.TestStep{
				// 30-day token
				{
					Config: testAccTFESCIMToken_withExpiredAt(description, expiredAtA),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("tfe_scim_token.test", "expired_at", expiredAtA),
						func(s *terraform.State) error {
							rs, ok := s.RootModule().Resources["tfe_scim_token.test"]
							if !ok {
								return fmt.Errorf("tfe_scim_token.test not found in state")
							}
							firstTokenID = rs.Primary.ID
							return nil
						},
					),
				},
				// bump to 60 days => RequiresReplace, ID must change
				{
					Config: testAccTFESCIMToken_withExpiredAt(description, expiredAtB),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("tfe_scim_token.test", "expired_at", expiredAtB),
						resource.TestCheckResourceAttrSet("tfe_scim_token.test", "token"),
						func(s *terraform.State) error {
							rs, ok := s.RootModule().Resources["tfe_scim_token.test"]
							if !ok {
								return fmt.Errorf("tfe_scim_token.test not found in state")
							}
							if rs.Primary.ID == firstTokenID {
								return fmt.Errorf("expected resource ID to change after expired_at update (RequiresReplace), but it remained %s", firstTokenID)
							}
							return nil
						},
					),
				},
			},
		})
	})

	t.Run("removing expired_at triggers resource replacement", func(t *testing.T) {
		description := "tf-acc-test-scim-token-rmexp-" + randomString(t)
		expiredAt := time.Now().UTC().Add(30 * 24 * time.Hour).Truncate(time.Second).Format(time.RFC3339)

		var firstTokenID string

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			CheckDestroy:             testAccTFESCIMTokenDestroy,
			Steps: []resource.TestStep{
				// Create with explicit expiry; Create writes a private-state marker.
				{
					Config: testAccTFESCIMToken_withExpiredAt(description, expiredAt),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("tfe_scim_token.test", "expired_at", expiredAt),
						func(s *terraform.State) error {
							rs, ok := s.RootModule().Resources["tfe_scim_token.test"]
							if !ok {
								return fmt.Errorf("tfe_scim_token.test not found in state")
							}
							firstTokenID = rs.Primary.ID
							return nil
						},
					),
				},
				// Drop expired_at: marker + null config => replace.
				{
					Config: testAccTFESCIMToken_basic(description),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("tfe_scim_token.test", "expired_at"),
						resource.TestCheckResourceAttrSet("tfe_scim_token.test", "token"),
						func(s *terraform.State) error {
							rs, ok := s.RootModule().Resources["tfe_scim_token.test"]
							if !ok {
								return fmt.Errorf("tfe_scim_token.test not found in state")
							}
							if rs.Primary.ID == firstTokenID {
								return fmt.Errorf("expected resource ID to change after removing expired_at (RequiresReplace), but it remained %s", firstTokenID)
							}
							return nil
						},
					),
				},
				// Marker is cleared after the replacement; re-apply must be a no-op.
				{
					Config:   testAccTFESCIMToken_basic(description),
					PlanOnly: true,
				},
			},
		})
	})

	t.Run("token deleted out-of-band is detected and re-created", func(t *testing.T) {
		description := "tf-acc-test-scim-token-drift-" + randomString(t)

		var tokenID string

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			CheckDestroy:             testAccTFESCIMTokenDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccTFESCIMToken_basic(description),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("tfe_scim_token.test", "id"),
						func(s *terraform.State) error {
							rs, ok := s.RootModule().Resources["tfe_scim_token.test"]
							if !ok {
								return fmt.Errorf("tfe_scim_token.test not found in state")
							}
							tokenID = rs.Primary.ID
							return nil
						},
					),
				},
				// Delete out-of-band: Read returns 404, state is cleared, then Create re-establishes.
				{
					PreConfig: func() {
						err := testAccConfiguredClient.Client.Admin.Settings.SCIM.Tokens.Delete(ctx, tokenID)
						if err != nil && !errors.Is(err, tfe.ErrResourceNotFound) {
							t.Fatalf("delete SCIM token out-of-band: %v", err)
						}
					},
					Config: testAccTFESCIMToken_basic(description),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("tfe_scim_token.test", "id"),
						resource.TestCheckResourceAttr("tfe_scim_token.test", "description", description),
						resource.TestCheckResourceAttrSet("tfe_scim_token.test", "token"),
					),
				},
			},
		})
	})
}

// testAccTFESCIMTokenDestroy verifies all tfe_scim_token resources have been
// removed from the backend.
func testAccTFESCIMTokenDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_scim_token" {
			continue
		}

		_, err := testAccConfiguredClient.Client.Admin.Settings.SCIM.Tokens.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("SCIM token %s still exists", rs.Primary.ID)
		}
		if !errors.Is(err, tfe.ErrResourceNotFound) {
			return fmt.Errorf("unexpected error checking SCIM token %s: %w", rs.Primary.ID, err)
		}
	}
	return nil
}

// testAccTFESCIMToken_basic returns a config with SAML + SCIM enabled and a
// tfe_scim_token using only the required description.
func testAccTFESCIMToken_basic(description string) string {
	return fmt.Sprintf(`
%s

resource "tfe_scim_settings" "enable_scim" {
    depends_on = [tfe_saml_settings.enable_saml]
}

resource "tfe_scim_token" "test" {
	description = "%s"
	depends_on  = [tfe_scim_settings.enable_scim]
}
`, testAccTFESCIMSettings_enableSAMLWithProviderType(scimTestSAMLSetting), description)
}

// testAccTFESCIMToken_withExpiredAt returns a config with an explicit
// expiration timestamp.
func testAccTFESCIMToken_withExpiredAt(description, expiredAt string) string {
	return fmt.Sprintf(`
%s

resource "tfe_scim_settings" "enable_scim" {
    depends_on = [tfe_saml_settings.enable_saml]
}

resource "tfe_scim_token" "test" {
	description = "%s"
	expired_at  = "%s"
	depends_on  = [tfe_scim_settings.enable_scim]
}
`, testAccTFESCIMSettings_enableSAMLWithProviderType(scimTestSAMLSetting), description, expiredAt)
}

// testAccTFESCIMToken_missingDescription returns a config that omits the
// required description attribute, so the framework should reject it at
// validate/plan time.
func testAccTFESCIMToken_missingDescription() string {
	return fmt.Sprintf(`
%s

resource "tfe_scim_settings" "enable_scim" {
    depends_on = [tfe_saml_settings.enable_saml]
}

resource "tfe_scim_token" "test" {
	depends_on = [tfe_scim_settings.enable_scim]
}
`, testAccTFESCIMSettings_enableSAMLWithProviderType(scimTestSAMLSetting))
}
