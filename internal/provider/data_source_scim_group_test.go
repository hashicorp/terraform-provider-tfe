// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccTFESCIMGroupDataSource_omnibus is the single test function for all
// SCIM groups data source acceptance tests.
//
// FLAKE ALERT: SCIM settings are a singleton resource shared by the entire TFE
// instance. Every sub-test here enables SCIM (via an inline tfe_scim_settings
// block) as a prerequisite. Keeping all cases in one function with no
// t.Parallel call prevents concurrent tests from racing over that singleton.
//
// FLAKE ALERT (dual-singleton): This suite also contends with
// resource_tfe_saml_settings_test.go for the SAML singleton. Both singletons
// must be treated as exclusive resources: do not run SCIM and SAML acceptance
// tests concurrently.
//
// Keep this test name matching the SCIM acceptance-test prefix used by the
// skip regex in ci.yml (currently TestAccTFESCIM), or update that regex.
func TestAccTFESCIMGroupDataSource_omnibus(t *testing.T) {
	skipIfCloud(t)

	t.Run("validation: config-level argument rules", func(t *testing.T) {
		lengthErr := regexp.MustCompile(`(?s)Invalid Attribute Value Length|at least 1`)
		whitespaceErr := regexp.MustCompile(`(?s)Invalid Attribute Value Match|non-whitespace`)

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			Steps: []resource.TestStep{
				{
					Config:      testAccTFESCIMGroupDataSourceNoArgs(),
					ExpectError: regexp.MustCompile(`(?s)Missing required argument|argument "name" is required`),
					PlanOnly:    true,
				},
				{
					Config:      testAccTFESCIMGroupDataSourceEmptyName(),
					ExpectError: lengthErr,
					PlanOnly:    true,
				},
				{
					Config:      testAccTFESCIMGroupDataSourceWhitespaceName(),
					ExpectError: whitespaceErr,
					PlanOnly:    true,
				},
			},
		})
	})

	t.Run("lifecycle: name", func(t *testing.T) {
		// Per-scenario unique prefixes so SCIM groups created in one step
		// can't interfere with another step's checks.
		rand := randomString(t)
		missingName := "tf-acc-scim-grp-name-missing-" + rand
		singleName := "tf-acc-scim-grp-name-single-" + rand
		fuzzyPrefix := "tf-acc-scim-grp-name-fuzzy-" + rand
		caseActual := "tf-acc-scim-grp-name-CaseSensitive-" + rand
		caseQueried := strings.ToUpper(caseActual)
		exactPrefix := "tf-acc-scim-grp-name-exact-" + rand
		exactSibling := exactPrefix + "-sibling"
		tokenDescription := "scim groups name lifecycle " + rand

		var scimToken string
		requireToken := func() {
			if scimToken == "" {
				t.Fatal("captured SCIM token value is empty")
			}
		}

		// Captured group ID returned by the SCIM API, asserted against the
		// data source's `id` attribute.
		var scimGroupID string

		notFoundErr := regexp.MustCompile(`(?s)SCIM group not found|No SCIM group found`)

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			CheckDestroy:             testAccTFESCIMSettingsDestroy,
			Steps: []resource.TestStep{
				// Step 0: turn SCIM on and grab a token for the SCIM API.
				{
					Config: testAccTFESCIMGroupDataSourceSetup(tokenDescription),
					Check:  captureSCIMTokenValue("tfe_scim_token.this", &scimToken),
				},
				// name → no match: errors.
				{
					Config:      testAccTFESCIMGroupDataSourceByName(tokenDescription, missingName),
					ExpectError: notFoundErr,
				},
				// name → single matching group.
				{
					PreConfig: func() {
						requireToken()
						scimGroupID = createSCIMGroup(t, singleName, scimToken)
					},
					Config: testAccTFESCIMGroupDataSourceByName(tokenDescription, singleName),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.tfe_scim_group.test", "name", singleName),
						resource.TestCheckResourceAttrPtr("data.tfe_scim_group.test", "id", &scimGroupID),
					),
				},
				// name → groups whose name starts with the prefix but don't equal it exactly are ignored (no match → errors).
				{
					PreConfig: func() {
						requireToken()
						createSCIMGroup(t, fuzzyPrefix+"-bar", scimToken)
						createSCIMGroup(t, fuzzyPrefix+"-baz", scimToken)
					},
					Config:      testAccTFESCIMGroupDataSourceByName(tokenDescription, fuzzyPrefix),
					ExpectError: notFoundErr,
				},
				// name → matches case-insensitively.
				{
					PreConfig: func() {
						requireToken()
						scimGroupID = createSCIMGroup(t, caseActual, scimToken)
					},
					Config: testAccTFESCIMGroupDataSourceByName(tokenDescription, caseQueried),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.tfe_scim_group.test", "name", caseQueried),
						resource.TestCheckResourceAttrPtr("data.tfe_scim_group.test", "id", &scimGroupID),
					),
				},
				// name → returns only the exact match when fuzzy siblings exist.
				{
					PreConfig: func() {
						requireToken()
						scimGroupID = createSCIMGroup(t, exactPrefix, scimToken)
						createSCIMGroup(t, exactSibling, scimToken)
					},
					Config: testAccTFESCIMGroupDataSourceByName(tokenDescription, exactPrefix),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.tfe_scim_group.test", "name", exactPrefix),
						resource.TestCheckResourceAttrPtr("data.tfe_scim_group.test", "id", &scimGroupID),
					),
				},
			},
		})
	})
}

// testAccTFESCIMGroupDataSourceNoArgs returns a data source block with
// `name` unset, which must fail because `name` is required.
func testAccTFESCIMGroupDataSourceNoArgs() string {
	return `data "tfe_scim_group" "test" {}`
}

// testAccTFESCIMGroupDataSourceEmptyName returns a data source block with
// an empty `name`, which must fail the LengthAtLeast(1) validator.
func testAccTFESCIMGroupDataSourceEmptyName() string {
	return `
data "tfe_scim_group" "test" {
    name = ""
}
`
}

// testAccTFESCIMGroupDataSourceWhitespaceName returns a data source block
// with a whitespace-only `name`, which must fail the non-whitespace validator.
func testAccTFESCIMGroupDataSourceWhitespaceName() string {
	return `
data "tfe_scim_group" "test" {
    name = "   "
}
`
}

// testAccTFESCIMGroupDataSourceSetup enables SAML + SCIM and creates a
// SCIM token so tests can use it to push groups through the SCIM API.
func testAccTFESCIMGroupDataSourceSetup(tokenDescription string) string {
	return fmt.Sprintf(`
%s

resource "tfe_scim_settings" "enable_scim" {
    depends_on = [tfe_saml_settings.enable_saml]
}

resource "tfe_scim_token" "this" {
    description = "%s"
    depends_on  = [tfe_scim_settings.enable_scim]
}
`, testAccTFESCIMSettings_enableSAMLWithProviderType(scimTestSAMLSetting), tokenDescription)
}

// testAccTFESCIMGroupDataSourceByName reads the data source by `name`,
// keeping the setup resources in place.
func testAccTFESCIMGroupDataSourceByName(tokenDescription, name string) string {
	return fmt.Sprintf(`
%s

data "tfe_scim_group" "test" {
    name       = "%s"
    depends_on = [tfe_scim_token.this]
}
`, testAccTFESCIMGroupDataSourceSetup(tokenDescription), name)
}
