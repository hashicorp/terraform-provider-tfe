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

func TestAccTFEStackPolicySetExclusion_basic(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	parents := testAccTFEStackPolicySetParentsConfig(rInt, "tfpolicy", false)
	var second testAccStackPolicySetIDs

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEStackPolicySetExclusionDestroy,
		Steps: []resource.TestStep{
			{
				Config: parents + testAccTFEStackPolicySetExclusionConfig("tfe_stack.first.id", "tfe_policy_set.test.id"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEStackPolicySetID("tfe_stack_policy_set_exclusion.test"),
					testAccCheckTFEStackPolicySetExcluded("tfe_stack.first", "tfe_policy_set.test", true),
				),
			},
			{
				ResourceName:      "tfe_stack_policy_set_exclusion.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:  "tfe_stack_policy_set_exclusion.test",
				ImportState:   true,
				ImportStateId: "invalid-import-id",
				ExpectError:   regexp.MustCompile("Invalid Import ID Format"),
			},
			{
				// Valid format, but this stack is not excluded from the policy set.
				ResourceName:      "tfe_stack_policy_set_exclusion.test",
				ImportState:       true,
				ImportStateIdFunc: testAccTFEStackPolicySetImportID("tfe_stack.second", "tfe_policy_set.test"),
				ExpectError:       regexp.MustCompile("Cannot import non-existent remote object"),
			},
			{
				Config: parents + testAccTFEStackPolicySetExclusionConfig("tfe_stack.second.id", "tfe_policy_set.test.id"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("tfe_stack_policy_set_exclusion.test", plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEStackPolicySetExcluded("tfe_stack.first", "tfe_policy_set.test", false),
					testAccCheckTFEStackPolicySetExcluded("tfe_stack.second", "tfe_policy_set.test", true),
					testAccCaptureStackPolicySetIDs("tfe_stack.second", "tfe_policy_set.test", &second),
				),
			},
			{
				// Removing the exclusion outside Terraform is detected as drift and re-created.
				Config:    parents + testAccTFEStackPolicySetExclusionConfig("tfe_stack.second.id", "tfe_policy_set.test.id"),
				PreConfig: testAccTFEStackPolicySetExclusionDetach(t, &second),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("tfe_stack_policy_set_exclusion.test", plancheck.ResourceActionCreate),
					},
				},
				Check: testAccCheckTFEStackPolicySetExcluded("tfe_stack.second", "tfe_policy_set.test", true),
			},
		},
	})
}

// Unlike inclusions, exclusions are allowed on global policy sets.
func TestAccTFEStackPolicySetExclusion_globalPolicySet(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	parents := testAccTFEStackPolicySetParentsConfig(rInt, "tfpolicy", true)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEStackPolicySetExclusionDestroy,
		Steps: []resource.TestStep{
			{
				Config: parents + testAccTFEStackPolicySetExclusionConfig("tfe_stack.first.id", "tfe_policy_set.test.id"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tfe_policy_set.test", "global", "true"),
					testAccCheckTFEStackPolicySetExcluded("tfe_stack.first", "tfe_policy_set.test", true),
				),
			},
			{
				ResourceName:      "tfe_stack_policy_set_exclusion.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTFEStackPolicySetExclusion_notFound(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	parents := testAccTFEStackPolicySetParentsConfig(rInt, "tfpolicy", false)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEStackPolicySetExclusionDestroy,
		Steps: []resource.TestStep{
			{
				Config:      parents + testAccTFEStackPolicySetExclusionConfig(`"st-doesnotexist"`, "tfe_policy_set.test.id"),
				ExpectError: regexp.MustCompile("Stack Not Found"),
			},
			{
				Config:      parents + testAccTFEStackPolicySetExclusionConfig("tfe_stack.first.id", `"polset-doesnotexist"`),
				ExpectError: regexp.MustCompile("Error Creating Stack Policy Set Exclusion"),
			},
		},
	})
}

func TestAccTFEStackPolicySetExclusion_invalidPolicySet(t *testing.T) {
	for _, kind := range []string{"sentinel", "opa"} {
		t.Run(kind+" kind", func(t *testing.T) {
			rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
			parents := testAccTFEStackPolicySetParentsConfig(rInt, kind, false)

			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { testAccPreCheck(t) },
				ProtoV6ProviderFactories: testAccMuxedProviders,
				CheckDestroy:             testAccCheckTFEStackPolicySetExclusionDestroy,
				Steps: []resource.TestStep{
					{
						Config:      parents + testAccTFEStackPolicySetExclusionConfig("tfe_stack.first.id", "tfe_policy_set.test.id"),
						ExpectError: regexp.MustCompile(`kind\s+tfpolicy`),
					},
				},
			})
		})
	}
}

func TestAccTFEStackPolicySetExclusion_tagScopeConflict(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	parents := testAccTFEStackPolicySetParentsConfig(rInt, "tfpolicy", false) + `
resource "tfe_workspace" "tagged" {
  name         = "stack-policy-set-tagged"
  organization = tfe_organization.test.name
  tags = {
    env = "test"
  }
}

resource "tfe_policy_set" "tagged" {
  name            = "stack-policy-set-tagged"
  organization    = tfe_organization.test.name
  kind            = "tfpolicy"
  tag_match_logic = "any"
}

resource "tfe_tag_policy_set" "test" {
  policy_set_id = tfe_policy_set.tagged.id
  key           = "env"
  value         = "test"
  depends_on    = [tfe_workspace.tagged]
}
`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEStackPolicySetExclusionDestroy,
		Steps: []resource.TestStep{
			{
				Config: parents,
			},
			{
				Config:      parents + testAccTFEStackPolicySetExclusionConfig("tfe_stack.first.id", "tfe_policy_set.tagged.id"),
				ExpectError: regexp.MustCompile(`Tag-based\s+scoping\s+cannot\s+be\s+combined`),
			},
		},
	})
}

func TestAccTFEStackPolicySetExclusion_includedConflict(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	parents := testAccTFEStackPolicySetParentsConfig(rInt, "tfpolicy", false) +
		testAccTFEStackPolicySetConfig("tfe_stack.first.id", "tfe_policy_set.test.id")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckTFEStackPolicySetDestroy,
			testAccCheckTFEStackPolicySetExclusionDestroy,
		),
		Steps: []resource.TestStep{
			{
				Config: parents,
			},
			{
				Config:      parents + testAccTFEStackPolicySetExclusionConfig("tfe_stack.first.id", "tfe_policy_set.test.id"),
				ExpectError: regexp.MustCompile(`is\s+included\s+in\s+this\s+policy\s+set`),
			},
		},
	})
}

func testAccCheckTFEStackPolicySetExcluded(stackName, policySetName string, want bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		stackID, policySetID, err := testAccTFEStackPolicySetIDs(s, stackName, policySetName)
		if err != nil {
			return err
		}
		r := &resourceStackPolicySetExclusion{config: *testAccConfiguredClient}
		ps, err := r.readPolicySet(ctx, policySetID)
		if err != nil {
			return err
		}
		excluded, err := policySetHasStack(ps, stackID, true)
		if err != nil {
			return err
		}
		if excluded != want {
			return fmt.Errorf("stack %s excluded from policy set %s = %t, want %t", stackID, policySetID, excluded, want)
		}
		return nil
	}
}

func testAccCheckTFEStackPolicySetExclusionDestroy(s *terraform.State) error {
	r := &resourceStackPolicySetExclusion{config: *testAccConfiguredClient}
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_stack_policy_set_exclusion" {
			continue
		}
		stackID := rs.Primary.Attributes["stack_id"]
		policySetID := rs.Primary.Attributes["policy_set_id"]

		ps, err := r.readPolicySet(ctx, policySetID)
		if err != nil {
			// If the policy set is gone, so is the exclusion.
			continue
		}
		excluded, err := policySetHasStack(ps, stackID, true)
		if err != nil {
			return err
		}
		if excluded {
			return fmt.Errorf("stack %s is still excluded from policy set %s", stackID, policySetID)
		}
	}
	return nil
}

// testAccTFEStackPolicySetExclusionDetach removes the exclusion through the
// API to simulate a change made outside Terraform.
func testAccTFEStackPolicySetExclusionDetach(t *testing.T, ids *testAccStackPolicySetIDs) func() {
	return func() {
		if ids.stackID == "" || ids.policySetID == "" {
			t.Fatal("stack policy set IDs were not captured by a previous step")
		}
		r := &resourceStackPolicySetExclusion{config: *testAccConfiguredClient}
		if err := r.writeStackExclusionRelationship(ctx, ids.policySetID, ids.stackID, true); err != nil {
			t.Fatalf("removing exclusion of stack %s from policy set %s: %s", ids.stackID, ids.policySetID, apiErrorDetail(err))
		}
	}
}

// testAccTFEStackPolicySetExclusionConfig takes HCL expressions, so callers
// can pass either a resource reference or a quoted literal ID.
func testAccTFEStackPolicySetExclusionConfig(stackID, policySetID string) string {
	return fmt.Sprintf(`
resource "tfe_stack_policy_set_exclusion" "test" {
  stack_id      = %s
  policy_set_id = %s
}
`, stackID, policySetID)
}
