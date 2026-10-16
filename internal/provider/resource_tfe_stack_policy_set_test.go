// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"testing"
	"time"

	tfev2 "github.com/hashicorp/go-tfe/v2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccTFEStackPolicySet_basic(t *testing.T) {
	skipUnlessStackPolicySetsSupported(t)

	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	parents := testAccTFEStackPolicySetParentsConfig(rInt, "tfpolicy", false)
	var second testAccStackPolicySetIDs

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEStackPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: parents + testAccTFEStackPolicySetConfig("tfe_stack.first.id", "tfe_policy_set.test.id"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEStackPolicySetID("tfe_stack_policy_set.test"),
					testAccCheckTFEStackPolicySetAttached("tfe_stack.first", "tfe_policy_set.test", true),
				),
			},
			{
				ResourceName:      "tfe_stack_policy_set.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:  "tfe_stack_policy_set.test",
				ImportState:   true,
				ImportStateId: "invalid-import-id",
				ExpectError:   regexp.MustCompile("Invalid Import ID Format"),
			},
			{
				// Valid format, but this stack is not attached to the policy set.
				ResourceName:      "tfe_stack_policy_set.test",
				ImportState:       true,
				ImportStateIdFunc: testAccTFEStackPolicySetImportID("tfe_stack.second", "tfe_policy_set.test"),
				ExpectError:       regexp.MustCompile("Cannot import non-existent remote object"),
			},
			{
				Config: parents + testAccTFEStackPolicySetConfig("tfe_stack.second.id", "tfe_policy_set.test.id"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("tfe_stack_policy_set.test", plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEStackPolicySetAttached("tfe_stack.first", "tfe_policy_set.test", false),
					testAccCheckTFEStackPolicySetAttached("tfe_stack.second", "tfe_policy_set.test", true),
					testAccCaptureStackPolicySetIDs("tfe_stack.second", "tfe_policy_set.test", &second),
				),
			},
			{
				// Detaching outside Terraform is detected as drift and re-created.
				Config:    parents + testAccTFEStackPolicySetConfig("tfe_stack.second.id", "tfe_policy_set.test.id"),
				PreConfig: testAccTFEStackPolicySetDetach(t, &second),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("tfe_stack_policy_set.test", plancheck.ResourceActionCreate),
					},
				},
				Check: testAccCheckTFEStackPolicySetAttached("tfe_stack.second", "tfe_policy_set.test", true),
			},
		},
	})
}

func TestAccTFEStackPolicySet_notFound(t *testing.T) {
	skipUnlessStackPolicySetsSupported(t)

	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	parents := testAccTFEStackPolicySetParentsConfig(rInt, "tfpolicy", false)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEStackPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config:      parents + testAccTFEStackPolicySetConfig(`"st-doesnotexist"`, "tfe_policy_set.test.id"),
				ExpectError: regexp.MustCompile("Stack Not Found"),
			},
			{
				Config:      parents + testAccTFEStackPolicySetConfig("tfe_stack.first.id", `"polset-doesnotexist"`),
				ExpectError: regexp.MustCompile("Error Creating Stack Policy Set Association"),
			},
		},
	})
}

func TestAccTFEStackPolicySet_invalidPolicySet(t *testing.T) {
	skipUnlessStackPolicySetsSupported(t)

	cases := map[string]struct {
		kind   string
		global bool
		err    *regexp.Regexp
	}{
		"sentinel kind": {kind: "sentinel", err: regexp.MustCompile(`kind\s+tfpolicy`)},
		"opa kind":      {kind: "opa", err: regexp.MustCompile(`kind\s+tfpolicy`)},
		"global":        {kind: "tfpolicy", global: true, err: regexp.MustCompile(`marked\s+as\s+global`)},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
			parents := testAccTFEStackPolicySetParentsConfig(rInt, tc.kind, tc.global)

			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { testAccPreCheck(t) },
				ProtoV6ProviderFactories: testAccMuxedProviders,
				CheckDestroy:             testAccCheckTFEStackPolicySetDestroy,
				Steps: []resource.TestStep{
					{
						Config:      parents + testAccTFEStackPolicySetConfig("tfe_stack.first.id", "tfe_policy_set.test.id"),
						ExpectError: tc.err,
					},
				},
			})
		})
	}
}

func TestAccTFEStackPolicySet_tagScopeConflict(t *testing.T) {
	skipUnlessStackPolicySetsSupported(t)

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
		CheckDestroy:             testAccCheckTFEStackPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: parents,
			},
			{
				Config:      parents + testAccTFEStackPolicySetConfig("tfe_stack.first.id", "tfe_policy_set.tagged.id"),
				ExpectError: regexp.MustCompile(`Tag-based\s+scoping\s+cannot\s+be\s+combined`),
			},
		},
	})
}

func TestAccTFEStackPolicySet_excludedConflict(t *testing.T) {
	skipUnlessStackPolicySetsSupported(t)

	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	parents := testAccTFEStackPolicySetParentsConfig(rInt, "tfpolicy", false) +
		testAccTFEStackPolicySetExclusionConfig("tfe_stack.first.id", "tfe_policy_set.test.id")

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
				Config:      parents + testAccTFEStackPolicySetConfig("tfe_stack.first.id", "tfe_policy_set.test.id"),
				ExpectError: regexp.MustCompile(`is\s+excluded\s+from\s+this\s+policy\s+set`),
			},
		},
	})
}

func testAccCheckTFEStackPolicySetID(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		want := rs.Primary.Attributes["stack_id"] + "/" + rs.Primary.Attributes["policy_set_id"]
		if rs.Primary.ID != want {
			return fmt.Errorf("expected ID %q, got %q", want, rs.Primary.ID)
		}
		return nil
	}
}

// testAccCheckTFEStackPolicySetAttached checks the API directly, so it also
// verifies detachment after the association resource has been replaced.
func testAccCheckTFEStackPolicySetAttached(stackName, policySetName string, want bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		stackID, policySetID, err := testAccTFEStackPolicySetIDs(s, stackName, policySetName)
		if err != nil {
			return err
		}
		ps, err := readPolicySetWithStacks(ctx, *testAccConfiguredClient, policySetID)
		if err != nil {
			return err
		}
		attached, err := policySetHasStack(ps, stackID, false)
		if err != nil {
			return err
		}
		if attached != want {
			return fmt.Errorf("stack %s attached to policy set %s = %t, want %t", stackID, policySetID, attached, want)
		}
		return nil
	}
}

func testAccCheckTFEStackPolicySetDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_stack_policy_set" {
			continue
		}
		stackID := rs.Primary.Attributes["stack_id"]
		policySetID := rs.Primary.Attributes["policy_set_id"]

		ps, err := readPolicySetWithStacks(ctx, *testAccConfiguredClient, policySetID)
		if errors.Is(err, tfev2.ErrNotFound) {
			// If the policy set is gone, so is the association.
			continue
		}
		if err != nil {
			return fmt.Errorf("reading policy set %s: %w", policySetID, err)
		}
		attached, err := policySetHasStack(ps, stackID, false)
		if err != nil {
			return err
		}
		if attached {
			return fmt.Errorf("stack %s is still attached to policy set %s", stackID, policySetID)
		}
	}
	return nil
}

func testAccTFEStackPolicySetImportID(stackName, policySetName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		stackID, policySetID, err := testAccTFEStackPolicySetIDs(s, stackName, policySetName)
		if err != nil {
			return "", err
		}
		return stackID + "/" + policySetID, nil
	}
}

type testAccStackPolicySetIDs struct {
	stackID     string
	policySetID string
}

// testAccCaptureStackPolicySetIDs saves IDs from state for a later step's
// PreConfig, which does not receive state.
func testAccCaptureStackPolicySetIDs(stackName, policySetName string, ids *testAccStackPolicySetIDs) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		var err error
		ids.stackID, ids.policySetID, err = testAccTFEStackPolicySetIDs(s, stackName, policySetName)
		return err
	}
}

// testAccTFEStackPolicySetDetach removes the association through the API to
// simulate a change made outside Terraform.
func testAccTFEStackPolicySetDetach(t *testing.T, ids *testAccStackPolicySetIDs) func() {
	return func() {
		if ids.stackID == "" || ids.policySetID == "" {
			t.Fatal("stack policy set IDs were not captured by a previous step")
		}
		r := &resourceStackPolicySet{config: *testAccConfiguredClient}
		if err := r.writeStackRelationship(ctx, ids.policySetID, ids.stackID, true); err != nil {
			t.Fatalf("detaching stack %s from policy set %s: %s", ids.stackID, ids.policySetID, apiErrorDetail(err))
		}
	}
}

func testAccTFEStackPolicySetIDs(s *terraform.State, stackName, policySetName string) (string, string, error) {
	stack, ok := s.RootModule().Resources[stackName]
	if !ok {
		return "", "", fmt.Errorf("Not found: %s", stackName)
	}
	policySet, ok := s.RootModule().Resources[policySetName]
	if !ok {
		return "", "", fmt.Errorf("Not found: %s", policySetName)
	}
	return stack.Primary.ID, policySet.Primary.ID, nil
}

func testAccTFEStackPolicySetParentsConfig(rInt int, kind string, global bool) string {
	return fmt.Sprintf(`
resource "tfe_organization" "test" {
  name           = "tst-terraform-%[1]d"
  email          = "admin@tfe.local"
  stacks_enabled = true
}

resource "tfe_project" "test" {
  name         = "stack-policy-set-test"
  organization = tfe_organization.test.name
}

resource "tfe_stack" "first" {
  name       = "stack-policy-set-first"
  project_id = tfe_project.test.id
}

resource "tfe_stack" "second" {
  name       = "stack-policy-set-second"
  project_id = tfe_project.test.id
}

resource "tfe_policy_set" "test" {
  name         = "stack-policy-set-test"
  organization = tfe_organization.test.name
  kind         = %[2]q
  global       = %[3]t
}
`, rInt, kind, global)
}

// testAccTFEStackPolicySetConfig takes HCL expressions, so callers can pass
// either a resource reference or a quoted literal ID.
func testAccTFEStackPolicySetConfig(stackID, policySetID string) string {
	return fmt.Sprintf(`
resource "tfe_stack_policy_set" "test" {
  stack_id      = %s
  policy_set_id = %s
}
`, stackID, policySetID)
}

// skipUnlessStackPolicySetsSupported skips on Terraform Enterprise releases
// older than minTFEVersionStackPolicySets. HCP Terraform always supports them.
func skipUnlessStackPolicySetsSupported(t *testing.T) {
	t.Helper()
	skipIfUnitTest(t)

	pc, err := getProviderClientUsingEnv()
	if err != nil {
		t.Fatalf("failed to build client: %v", err)
	}
	client := ConfiguredClient{Client: pc.TfeClient, ClientV2: pc.TFEClientV2, Hostname: pc.Hostname}
	supported, err := client.MeetsMinRemoteTFEVersion(minTFEVersionStackPolicySets)
	if err != nil {
		t.Fatalf("failed to check Terraform Enterprise version: %v", err)
	}
	if !supported {
		t.Skipf("Stack policy sets require Terraform Enterprise %s or later; this instance reports %s.",
			minTFEVersionStackPolicySets, client.RemoteTFEVersion())
	}
}
