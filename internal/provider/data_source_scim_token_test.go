// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccTFESCIMTokenDataSource_omnibus is the single test function for all SCIM
// token data source acceptance tests.
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
func TestAccTFESCIMTokenDataSource_omnibus(t *testing.T) {
	skipIfCloud(t)

	t.Run("basic read by id", func(t *testing.T) {
		description := "tf-acc-test-scim-token-ds-" + randomString(t)

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			CheckDestroy:             testAccTFESCIMTokenDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccTFESCIMTokenDataSourceConfig_basic(description),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("data.tfe_scim_token.test", "id"),
						resource.TestCheckResourceAttr("data.tfe_scim_token.test", "description", description),
						resource.TestCheckResourceAttrSet("data.tfe_scim_token.test", "expired_at"),
						resource.TestCheckResourceAttrSet("data.tfe_scim_token.test", "created_at"),
						resource.TestCheckNoResourceAttr("data.tfe_scim_token.test", "last_used_at"),
					),
				},
			},
		})
	})

	t.Run("read with explicit expired_at", func(t *testing.T) {
		description := "tf-acc-test-scim-token-ds-exp-" + randomString(t)
		expiredAt := time.Now().UTC().Add(180 * 24 * time.Hour).Truncate(time.Second).Format(time.RFC3339)

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			CheckDestroy:             testAccTFESCIMTokenDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccTFESCIMTokenDataSourceConfig_withExpiredAt(description, expiredAt),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("data.tfe_scim_token.test", "id"),
						resource.TestCheckResourceAttr("data.tfe_scim_token.test", "description", description),
						resource.TestCheckResourceAttr("data.tfe_scim_token.test", "expired_at", expiredAt),
					),
				},
			},
		})
	})

	t.Run("invalid id is rejected at validate time", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			Steps: []resource.TestStep{
				{
					Config:      testAccTFESCIMTokenDataSourceConfigInvalidID(),
					ExpectError: regexp.MustCompile("must be a valid SCIM token ID starting with 'at-'"),
				},
			},
		})
	})
}

// testAccTFESCIMTokenDataSourceConfig_basic returns a config that enables
// SAML + SCIM, creates a SCIM token, and reads it back via the tfe_scim_token
// data source using the resource's id.
func testAccTFESCIMTokenDataSourceConfig_basic(description string) string {
	return fmt.Sprintf(`
%s

resource "tfe_scim_settings" "enable_scim" {
    depends_on = [tfe_saml_settings.enable_saml]
}

resource "tfe_scim_token" "test" {
    description = "%s"
    depends_on  = [tfe_scim_settings.enable_scim]
}

data "tfe_scim_token" "test" {
    id         = tfe_scim_token.test.id
    depends_on = [tfe_scim_token.test]
}
`, testAccTFESCIMSettings_enableSAMLWithProviderType(scimTestSAMLSetting), description)
}

// testAccTFESCIMTokenDataSourceConfig_withExpiredAt returns a config that
// enables SAML + SCIM, creates a SCIM token with an explicit expiry, and reads
// it back via the tfe_scim_token data source.
func testAccTFESCIMTokenDataSourceConfig_withExpiredAt(description, expiredAt string) string {
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

data "tfe_scim_token" "test" {
    id         = tfe_scim_token.test.id
    depends_on = [tfe_scim_token.test]
}
`, testAccTFESCIMSettings_enableSAMLWithProviderType(scimTestSAMLSetting), description, expiredAt)
}

// testAccTFESCIMTokenDataSourceConfigInvalidID returns a config with an
// invalid SCIM token ID format to validate schema-level ID checks.
func testAccTFESCIMTokenDataSourceConfigInvalidID() string {
	return fmt.Sprintf(`
%s

data "tfe_scim_token" "test" {
	id = "invalid-token-id"
}
`, testAccTFESCIMSettings_enableSAMLWithProviderType(scimTestSAMLSetting))
}
