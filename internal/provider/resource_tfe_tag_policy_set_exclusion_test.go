// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"math/rand"
	"regexp"
	"testing"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccTFETagPolicySetExclusion_keyValueTag(t *testing.T) {
	skipUnlessBeta(t)

	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createOrganization(t, tfeClient, tfe.OrganizationCreateOptions{Name: tfe.String("tst-" + randomString(t)), Email: tfe.String(fmt.Sprintf("%s@tfe.local", randomString(t)))})
	t.Cleanup(orgCleanup)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFETagPolicySetExclusionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETagPolicySetExclusion_keyValueTag(org.Name, rInt),
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
	skipUnlessBeta(t)

	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createOrganization(t, tfeClient, tfe.OrganizationCreateOptions{Name: tfe.String("tst-" + randomString(t)), Email: tfe.String(fmt.Sprintf("%s@tfe.local", randomString(t)))})
	t.Cleanup(orgCleanup)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFETagPolicySetExclusionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETagPolicySetExclusion_keyOnlyTag(org.Name, rInt),
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

func TestAccTFETagPolicySetExclusion_incorrectImportSyntax(t *testing.T) {
	skipUnlessBeta(t)

	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createOrganization(t, tfeClient, tfe.OrganizationCreateOptions{Name: tfe.String("tst-" + randomString(t)), Email: tfe.String(fmt.Sprintf("%s@tfe.local", randomString(t)))})
	t.Cleanup(orgCleanup)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETagPolicySetExclusion_keyValueTag(org.Name, rInt),
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
		global       = true
	}

	resource "tfe_tag_policy_set_exclusion" "test" {
		policy_set_id = tfe_policy_set.test.id
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
		global       = true
	}

	resource "tfe_tag_policy_set_exclusion" "test" {
		policy_set_id = tfe_policy_set.test.id
		key           = "team"
	}`,
		rInt, orgName, rInt, orgName)
}
