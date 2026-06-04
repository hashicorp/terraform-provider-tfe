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

// TestAccTFESCIMGroupsDataSource_omnibus is the single test function for all
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
func TestAccTFESCIMGroupsDataSource_omnibus(t *testing.T) {
	skipIfCloud(t)

	t.Run("validation: config-level argument rules", func(t *testing.T) {
		lengthErr := regexp.MustCompile(`(?s)Invalid Attribute Value Length|at least 1`)
		whitespaceErr := regexp.MustCompile(`(?s)Invalid Attribute Value Match|non-whitespace`)

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			Steps: []resource.TestStep{
				{
					Config:      testAccTFESCIMGroupsDataSource_noArgs(),
					ExpectError: regexp.MustCompile(`(?s)Missing required argument|argument "name" is required`),
					PlanOnly:    true,
				},
				{
					Config:      testAccTFESCIMGroupsDataSource_emptyName(),
					ExpectError: lengthErr,
					PlanOnly:    true,
				},
				{
					Config:      testAccTFESCIMGroupsDataSource_whitespaceName(),
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

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			CheckDestroy:             testAccTFESCIMSettingsDestroy,
			Steps: []resource.TestStep{
				// Step 0: turn SCIM on and grab a token for the SCIM API.
				{
					Config: testAccTFESCIMGroupsDataSource_setup(tokenDescription),
					Check:  captureSCIMTokenValue("tfe_scim_token.this", &scimToken),
				},
				// name → no matches: empty groups, null group_id/group_name.
				{
					Config: testAccTFESCIMGroupsDataSource_byName(tokenDescription, missingName),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.tfe_scim_groups.test", "id", "name/"+missingName),
						resource.TestCheckResourceAttr("data.tfe_scim_groups.test", "groups.#", "0"),
						resource.TestCheckNoResourceAttr("data.tfe_scim_groups.test", "group_id"),
						resource.TestCheckNoResourceAttr("data.tfe_scim_groups.test", "group_name"),
					),
				},
				// name → single matching group.
				{
					PreConfig: func() {
						requireToken()
						createSCIMGroup(t, singleName, scimToken)
					},
					Config: testAccTFESCIMGroupsDataSource_byName(tokenDescription, singleName),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.tfe_scim_groups.test", "id", "name/"+singleName),
						resource.TestCheckResourceAttr("data.tfe_scim_groups.test", "name", singleName),
						resource.TestCheckResourceAttr("data.tfe_scim_groups.test", "groups.#", "1"),
						resource.TestCheckResourceAttr("data.tfe_scim_groups.test", "group_name", singleName),
						resource.TestCheckResourceAttrSet("data.tfe_scim_groups.test", "group_id"),
						resource.TestCheckResourceAttr("data.tfe_scim_groups.test", "groups.0.name", singleName),
						resource.TestCheckResourceAttrSet("data.tfe_scim_groups.test", "groups.0.id"),
					),
				},
				// name → groups whose name starts with the prefix but don't equal it exactly are ignored.
				{
					PreConfig: func() {
						requireToken()
						createSCIMGroup(t, fuzzyPrefix+"-bar", scimToken)
						createSCIMGroup(t, fuzzyPrefix+"-baz", scimToken)
					},
					Config: testAccTFESCIMGroupsDataSource_byName(tokenDescription, fuzzyPrefix),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.tfe_scim_groups.test", "id", "name/"+fuzzyPrefix),
						resource.TestCheckResourceAttr("data.tfe_scim_groups.test", "groups.#", "0"),
						resource.TestCheckNoResourceAttr("data.tfe_scim_groups.test", "group_id"),
						resource.TestCheckNoResourceAttr("data.tfe_scim_groups.test", "group_name"),
					),
				},
				// name → matches case-insensitively.
				{
					PreConfig: func() {
						requireToken()
						createSCIMGroup(t, caseActual, scimToken)
					},
					Config: testAccTFESCIMGroupsDataSource_byName(tokenDescription, caseQueried),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.tfe_scim_groups.test", "id", "name/"+caseQueried),
						resource.TestCheckResourceAttr("data.tfe_scim_groups.test", "name", caseQueried),
						resource.TestCheckResourceAttr("data.tfe_scim_groups.test", "groups.#", "1"),
						resource.TestCheckResourceAttr("data.tfe_scim_groups.test", "group_name", caseActual),
						resource.TestCheckResourceAttr("data.tfe_scim_groups.test", "groups.0.name", caseActual),
						resource.TestCheckResourceAttrSet("data.tfe_scim_groups.test", "group_id"),
					),
				},
				// name → returns only the exact match when fuzzy siblings exist.
				{
					PreConfig: func() {
						requireToken()
						createSCIMGroup(t, exactPrefix, scimToken)
						createSCIMGroup(t, exactSibling, scimToken)
					},
					Config: testAccTFESCIMGroupsDataSource_byName(tokenDescription, exactPrefix),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.tfe_scim_groups.test", "id", "name/"+exactPrefix),
						resource.TestCheckResourceAttr("data.tfe_scim_groups.test", "groups.#", "1"),
						resource.TestCheckResourceAttr("data.tfe_scim_groups.test", "group_name", exactPrefix),
						resource.TestCheckResourceAttrSet("data.tfe_scim_groups.test", "group_id"),
						resource.TestCheckResourceAttr("data.tfe_scim_groups.test", "groups.0.name", exactPrefix),
					),
				},
			},
		})
	})
}

// testAccTFESCIMGroupsDataSource_noArgs returns a data source block with
// `name` unset, which must fail because `name` is required.
func testAccTFESCIMGroupsDataSource_noArgs() string {
	return `data "tfe_scim_groups" "test" {}`
}

// testAccTFESCIMGroupsDataSource_emptyName returns a data source block with
// an empty `name`, which must fail the LengthAtLeast(1) validator.
func testAccTFESCIMGroupsDataSource_emptyName() string {
	return `
data "tfe_scim_groups" "test" {
    name = ""
}
`
}

// testAccTFESCIMGroupsDataSource_whitespaceName returns a data source block
// with a whitespace-only `name`, which must fail the non-whitespace validator.
func testAccTFESCIMGroupsDataSource_whitespaceName() string {
	return `
data "tfe_scim_groups" "test" {
    name = "   "
}
`
}

// testAccTFESCIMGroupsDataSource_setup enables SAML + SCIM and creates a
// SCIM token so tests can use it to push groups through the SCIM API.
func testAccTFESCIMGroupsDataSource_setup(tokenDescription string) string {
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

// testAccTFESCIMGroupsDataSource_byName reads the data source by `name`,
// keeping the setup resources in place.
func testAccTFESCIMGroupsDataSource_byName(tokenDescription, name string) string {
	return fmt.Sprintf(`
%s

data "tfe_scim_groups" "test" {
    name       = "%s"
    depends_on = [tfe_scim_token.this]
}
`, testAccTFESCIMGroupsDataSource_setup(tokenDescription), name)
}
