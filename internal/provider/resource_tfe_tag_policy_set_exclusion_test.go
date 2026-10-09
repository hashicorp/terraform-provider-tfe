// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"math/rand"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccTFETagPolicySetExclusion_keyValueTag(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := "tst-" + randomString(t)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCreateOrganizationNamed(t, orgName)
		},
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFETagPolicySetExclusionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETagPolicySetExclusion_keyValueTag(orgName, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETagPolicySetExclusionExists("tfe_tag_policy_set_exclusion.test"),
					resource.TestCheckResourceAttr("tfe_tag_policy_set_exclusion.test", "key", "env"),
					resource.TestCheckResourceAttr("tfe_tag_policy_set_exclusion.test", "value", "staging"),
					resource.TestCheckResourceAttrSet("tfe_tag_policy_set_exclusion.test", "policy_set_id"),
				),
			},
			{
				ResourceName: "tfe_tag_policy_set_exclusion.test",
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["tfe_tag_policy_set_exclusion.test"]
					if !ok {
						return "", fmt.Errorf("resource not found")
					}
					return fmt.Sprintf("%s/env/staging", rs.Primary.Attributes["policy_set_id"]), nil
				},
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTFETagPolicySetExclusion_keyOnlyTag(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := "tst-" + randomString(t)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCreateOrganizationNamed(t, orgName)
		},
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFETagPolicySetExclusionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETagPolicySetExclusion_keyOnlyTag(orgName, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETagPolicySetExclusionExists("tfe_tag_policy_set_exclusion.test"),
					resource.TestCheckResourceAttr("tfe_tag_policy_set_exclusion.test", "key", "team"),
					resource.TestCheckNoResourceAttr("tfe_tag_policy_set_exclusion.test", "value"),
					resource.TestCheckResourceAttrSet("tfe_tag_policy_set_exclusion.test", "policy_set_id"),
				),
			},
			{
				ResourceName: "tfe_tag_policy_set_exclusion.test",
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["tfe_tag_policy_set_exclusion.test"]
					if !ok {
						return "", fmt.Errorf("resource not found")
					}
					return fmt.Sprintf("%s/team", rs.Primary.Attributes["policy_set_id"]), nil
				},
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTFETagPolicySetExclusion_tfPolicy(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := "tst-" + randomString(t)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCreateBusinessOrganizationNamed(t, orgName)
		},
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFETagPolicySetExclusionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETagPolicySetExclusion_tfPolicy(orgName, rInt, "any"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETagPolicySetExclusionExists("tfe_tag_policy_set_exclusion.test"),
					resource.TestCheckResourceAttr("tfe_policy_set.test", "kind", "tfpolicy"),
					resource.TestCheckResourceAttr("tfe_policy_set.test", "tag_match_logic", "any"),
					resource.TestCheckResourceAttr("tfe_policy_set.test", "global", "true"),
					resource.TestCheckResourceAttr("tfe_tag_policy_set_exclusion.test", "key", "env"),
					resource.TestCheckResourceAttr("tfe_tag_policy_set_exclusion.test", "value", "staging"),
					resource.TestCheckResourceAttrPair(
						"tfe_tag_policy_set_exclusion.test", "policy_set_id", "tfe_policy_set.test", "id"),
				),
			},
			{
				ResourceName: "tfe_tag_policy_set_exclusion.test",
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["tfe_tag_policy_set_exclusion.test"]
					if !ok {
						return "", fmt.Errorf("resource not found")
					}
					return fmt.Sprintf("%s/env/staging", rs.Primary.Attributes["policy_set_id"]), nil
				},
				ImportStateVerify: true,
			},
			{
				Config: testAccTFETagPolicySetExclusion_tfPolicy(orgName, rInt, "all"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("tfe_policy_set.test", plancheck.ResourceActionUpdate),
						plancheck.ExpectResourceAction("tfe_tag_policy_set_exclusion.test", plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETagPolicySetExclusionExists("tfe_tag_policy_set_exclusion.test"),
					resource.TestCheckResourceAttr("tfe_policy_set.test", "tag_match_logic", "all"),
				),
			},
		},
	})
}

func TestAccTFETagPolicySetExclusion_incorrectImportSyntax(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := "tst-" + randomString(t)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCreateOrganizationNamed(t, orgName)
		},
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETagPolicySetExclusion_keyValueTag(orgName, rInt),
			},
			{
				ResourceName:  "tfe_tag_policy_set_exclusion.test",
				ImportState:   true,
				ImportStateId: "not-a-polset/env/staging",
				ExpectError:   regexp.MustCompile(`Invalid Policy Set ID`),
			},
		},
	})
}

func testAccCheckTFETagPolicySetExclusionExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		id := rs.Primary.ID
		if id == "" {
			return fmt.Errorf("no ID is set")
		}

		policySetID := rs.Primary.Attributes["policy_set_id"]
		if policySetID == "" {
			return fmt.Errorf("no policy set id set")
		}

		key := rs.Primary.Attributes["key"]
		if key == "" {
			return fmt.Errorf("no tag key set")
		}

		value := rs.Primary.Attributes["value"]

		policySet, err := testAccConfiguredClient.Client.PolicySets.Read(ctx, policySetID)
		if err != nil {
			return fmt.Errorf("error reading policy set %s: %w", policySetID, err)
		}

		for _, ts := range policySet.TagSelectors {
			if ts.Key == key && ts.IsExclude {
				if value == "" && ts.Value == nil {
					return nil
				}
				if ts.Value != nil && *ts.Value == value {
					return nil
				}
			}
		}

		return fmt.Errorf("tag exclusion (key=%s, value=%s) not found in policy set (%s)", key, value, policySetID)
	}
}

func testAccCheckTFETagPolicySetExclusionDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_tag_policy_set_exclusion" {
			continue
		}

		policySetID := rs.Primary.Attributes["policy_set_id"]
		key := rs.Primary.Attributes["key"]
		value := rs.Primary.Attributes["value"]

		policySet, err := testAccConfiguredClient.Client.PolicySets.Read(ctx, policySetID)
		if err != nil {
			// Policy set itself was destroyed, so tag exclusion is definitely gone
			continue
		}

		for _, ts := range policySet.TagSelectors {
			if ts.Key == key && ts.IsExclude {
				if value == "" && ts.Value == nil {
					return fmt.Errorf("tag exclusion (key=%s) still exists in policy set %s", key, policySetID)
				}
				if ts.Value != nil && *ts.Value == value {
					return fmt.Errorf("tag exclusion (key=%s, value=%s) still exists in policy set %s", key, value, policySetID)
				}
			}
		}
	}

	return nil
}

func testAccTFETagPolicySetExclusion_keyValueTag(orgName string, rInt int) string {
	return fmt.Sprintf(`
	resource "tfe_workspace" "test" {
		name         = "tst-workspace-%d"
		organization = "%s"
		tags = {
			env = "staging"
		}
	}

	resource "tfe_policy_set" "test" {
		name         = "tst-policy-set-%d"
		description  = "Policy Set"
		organization = "%s"
		tag_match_logic = "any"
		global       = true
	}

	resource "tfe_tag_policy_set_exclusion" "test" {
		policy_set_id = tfe_policy_set.test.id
		depends_on    = [tfe_workspace.test]
		key           = "env"
		value         = "staging"
	}`,
		rInt, orgName, rInt, orgName)
}

func testAccTFETagPolicySetExclusion_keyOnlyTag(orgName string, rInt int) string {
	return fmt.Sprintf(`
	resource "tfe_workspace" "test" {
		name         = "tst-workspace-%d"
		organization = "%s"
		tags = {
			team = ""
		}
	}

	resource "tfe_policy_set" "test" {
		name         = "tst-policy-set-%d"
		description  = "Policy Set"
		organization = "%s"
		tag_match_logic = "any"
		global       = true
	}

	resource "tfe_tag_policy_set_exclusion" "test" {
		policy_set_id = tfe_policy_set.test.id
		depends_on    = [tfe_workspace.test]
		key           = "team"
	}`,
		rInt, orgName, rInt, orgName)
}

func testAccTFETagPolicySetExclusion_tfPolicy(orgName string, rInt int, tagMatchLogic string) string {
	return fmt.Sprintf(`
	resource "tfe_workspace" "test" {
		name         = "tst-workspace-%d"
		organization = "%s"
		tags = {
			env = "staging"
		}
	}

	resource "tfe_policy_set" "test" {
		name         = "tst-tfpolicy-set-%d"
		description  = "TFPolicy Policy Set"
		organization = "%s"
		tag_match_logic = "%s"
		kind         = "tfpolicy"
		global       = true
	}

	resource "tfe_tag_policy_set_exclusion" "test" {
		policy_set_id = tfe_policy_set.test.id
		depends_on    = [tfe_workspace.test]
		key           = "env"
		value         = "staging"
	}`,
		rInt, orgName, rInt, orgName, tagMatchLogic)
}

func TestAccTFETagPolicySetExclusion_tfPolicyNonGlobal(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := "tst-" + randomString(t)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCreateBusinessOrganizationNamed(t, orgName)
		},
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFETagPolicySetExclusionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETagPolicySetExclusion_tfPolicyNonGlobal(orgName, rInt),
				// Terraform wraps long diagnostics, so allow any whitespace between words.
				ExpectError: regexp.MustCompile(`Tag-based\s+exclusions\s+are\s+not\s+allowed\s+on\s+non-global\s+policy\s+sets`),
			},
		},
	})
}

func testAccTFETagPolicySetExclusion_tfPolicyNonGlobal(orgName string, rInt int) string {
	return fmt.Sprintf(`
	resource "tfe_policy_set" "test" {
		name            = "tst-tfpolicy-set-%d"
		description     = "TFPolicy Policy Set"
		organization    = "%s"
		kind            = "tfpolicy"
		global          = false
		tag_match_logic = "any"
	}

	resource "tfe_tag_policy_set_exclusion" "test" {
		policy_set_id = tfe_policy_set.test.id
		key           = "env"
		value         = "staging"
	}`, rInt, orgName)
}
