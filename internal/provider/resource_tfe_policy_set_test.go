// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/go-cty/cty"
	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func TestAccTFEPolicySet_basic(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	policySet := &tfe.PolicySet{}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySet_basic(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetAttributes(policySet),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "name", "tst-terraform"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "description", "Policy Set"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "global", "false"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "agent_enabled", "false"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "policy_ids.#", "1"),
				),
			},
		},
	})
}

func TestAccTFEPolicySet_pinnedPolicyRuntimeVersion(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	sha := genSentinelSha(t, "secret", "data")
	version := genSafeRandomSentinelVersion()

	adminClient := tfeClient
	if !enterpriseEnabled() {
		adminClient = testAdminClient(t, versionMaintenanceAdmin)
	}

	opts := tfe.AdminSentinelVersionCreateOptions{
		Version: version,
		SHA:     sha,
		URL:     "https://hashicorp.com",
	}

	tool, err := adminClient.Admin.SentinelVersions.Create(context.Background(), opts)
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	policySet := &tfe.PolicySet{}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySet_pinnedPolicyRuntimeVersion(org.Name, tool.Version),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetAttributes(policySet),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "name", "tst-terraform"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "description", "Policy Set"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "global", "false"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "agent_enabled", "true"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "policy_tool_version", version),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "policy_ids.#", "1"),
				),
			},
		},
	})
}

func TestAccTFEPolicySetOPA_basic(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	sha := genSentinelSha(t, "secret", "data")
	version := genSafeRandomOPAVersion()

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	policySet := &tfe.PolicySet{}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySetOPA_basic(org.Name, version, sha),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetAttributes(policySet),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "name", "tst-terraform"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "kind", "opa"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "overridable", "true"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "agent_enabled", "true"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "policy_tool_version", version),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "description", "Policy Set"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "global", "false"),
				),
			},
		},
	})
}

func TestAccTFEPolicySetTFPolicy_basic(t *testing.T) {
	orgName := "tst-" + randomString(t)
	policySet := &tfe.PolicySet{}
	deletedPolicySet := &tfe.PolicySet{}
	resourceName := "tfe_policy_set.foobar-tfpolicy"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCreateBusinessOrganizationNamed(t, orgName)
		},
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySetTFPolicy_basic(orgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists(resourceName, policySet),
					testAccCheckTFEPolicySetKind(policySet, tfe.TFPolicy),
					testAccCheckTFEPolicySetNotOverridable(policySet),
					resource.TestCheckResourceAttr(resourceName, "name", "tst-terraform-tfpolicy"),
					resource.TestCheckResourceAttr(resourceName, "description", "TFPolicy Policy Set"),
					resource.TestCheckResourceAttr(resourceName, "kind", "tfpolicy"),
					resource.TestCheckResourceAttr(resourceName, "global", "false"),
					resource.TestCheckResourceAttr(resourceName, "overridable", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "agent_enabled"),
					resource.TestCheckResourceAttr(resourceName, "policy_tool_version", "latest"),
				),
			},
			{
				Config: testAccTFEPolicySetTFPolicy_updated(orgName, true),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPtr(resourceName, "id", &policySet.ID),
					testAccCheckTFEPolicySetExists(resourceName, policySet),
					testAccCheckTFEPolicySetKind(policySet, tfe.TFPolicy),
					testAccCheckTFEPolicySetNotOverridable(policySet),
					resource.TestCheckResourceAttr(resourceName, "name", "tst-terraform-tfpolicy-updated"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated TFPolicy Policy Set"),
					resource.TestCheckResourceAttr(resourceName, "global", "true"),
					resource.TestCheckResourceAttr(resourceName, "overridable", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTFEPolicySetTFPolicy_updated(orgName, false),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists(resourceName, policySet),
					testAccCheckTFEPolicySetKind(policySet, tfe.TFPolicy),
					resource.TestCheckResourceAttr(resourceName, "global", "false"),
				),
			},
			{
				Config: testAccTFEPolicySetTFPolicy_updated(orgName, false),
				PreConfig: func() {
					*deletedPolicySet = *policySet
					testAccDeletePolicySetOutOfBand(t, policySet.ID)
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists(resourceName, policySet),
					testAccCheckTFEPolicySetKind(policySet, tfe.TFPolicy),
					testAccCheckTFEPolicySetRecreated(deletedPolicySet, policySet),
				),
			},
		},
	})
}

func testAccCreateOrganizationNamed(t *testing.T, orgName string) *tfe.Organization {
	t.Helper()

	client, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createOrganization(t, client, tfe.OrganizationCreateOptions{
		Name:  tfe.String(orgName),
		Email: tfe.String(fmt.Sprintf("%s@tfe.local", orgName)),
	})
	t.Cleanup(orgCleanup)

	return org
}

func testAccCreateBusinessOrganizationNamed(t *testing.T, orgName string) {
	t.Helper()

	org := testAccCreateOrganizationNamed(t, orgName)
	newSubscriptionUpdater(org).WithBusinessPlan().Update(t)
}

func testAccDeletePolicySetOutOfBand(t *testing.T, policySetID string) {
	t.Helper()

	client, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	if err := client.PolicySets.Delete(ctx, policySetID); err != nil {
		t.Fatalf("error deleting policy set %s: %v", policySetID, err)
	}

	// Wait for the deletion to be visible so the next refresh does not race replica lag.
	err = retry.RetryContext(ctx, 30*time.Second, func() *retry.RetryError {
		_, err := client.PolicySets.Read(ctx, policySetID)
		if errors.Is(err, tfe.ErrResourceNotFound) {
			return nil
		}
		if err != nil {
			return retry.NonRetryableError(err)
		}
		return retry.RetryableError(fmt.Errorf("policy set %s still exists", policySetID))
	})
	if err != nil {
		t.Fatal(err)
	}
}

func testAccCheckTFEPolicySetNotOverridable(policySet *tfe.PolicySet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if policySet.Overridable != nil && *policySet.Overridable {
			return fmt.Errorf("expected TFPolicy policy set %s not to be overridable", policySet.ID)
		}

		return nil
	}
}

func TestAccTFEPolicySetTFPolicy_unsupportedAttributes(t *testing.T) {
	for _, attribute := range []string{"agent_enabled", "overridable"} {
		for _, value := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/%t", attribute, value), func(t *testing.T) {
				resource.Test(t, resource.TestCase{
					PreCheck:                 func() { testAccPreCheck(t) },
					ProtoV6ProviderFactories: testAccMuxedProviders,
					Steps: []resource.TestStep{
						{
							Config: fmt.Sprintf(`
resource "tfe_policy_set" "test" {
  name         = "tst-tfpolicy-invalid"
  organization = "unused-for-plan-validation"
  kind         = "tfpolicy"
  %s = %t
}`, attribute, value),
							PlanOnly:    true,
							ExpectError: regexp.MustCompile(attribute + ` is not supported when kind is "tfpolicy"`),
						},
					},
				})
			})
		}
	}
}

func TestAccTFEPolicySetTFPolicy_unsupportedDeferredAttributes(t *testing.T) {
	for _, attribute := range []string{"agent_enabled", "overridable"} {
		for _, value := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/%t", attribute, value), func(t *testing.T) {
				resource.Test(t, resource.TestCase{
					PreCheck:                 func() { testAccPreCheck(t) },
					ProtoV6ProviderFactories: testAccMuxedProviders,
					TerraformVersionChecks: []tfversion.TerraformVersionCheck{
						tfversion.SkipBelow(tfversion.Version1_4_0),
					},
					Steps: []resource.TestStep{
						{
							// terraform_data.output is unknown during plan, so validation is deferred to apply.
							Config: fmt.Sprintf(`
resource "terraform_data" "test" {
  input = %t
}

resource "tfe_policy_set" "test" {
  name         = "tst-tfpolicy-invalid"
  organization = "unused-for-apply-validation"
  kind         = "tfpolicy"
  %s = terraform_data.test.output
}`, value, attribute),
							ExpectError: regexp.MustCompile(attribute + ` is not supported when kind is "tfpolicy"`),
						},
					},
				})
			})
		}
	}
}

func TestAccTFEPolicySetTFPolicy_managedToolVersion(t *testing.T) {
	orgName := "tst-" + randomString(t)
	policySet := &tfe.PolicySet{}
	resourceName := "tfe_policy_set.foobar-tfpolicy"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCreateBusinessOrganizationNamed(t, orgName)
		},
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySetTFPolicy_managedToolVersion(orgName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists(resourceName, policySet),
					testAccCheckTFEPolicySetKind(policySet, tfe.TFPolicy),
					resource.TestCheckResourceAttr(resourceName, "kind", "tfpolicy"),
					resource.TestCheckResourceAttr(resourceName, "policy_tool_version", "managed"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTFEPolicySetTFPolicy_workspaceIDs(t *testing.T) {
	orgName := "tst-" + randomString(t)
	policySet := &tfe.PolicySet{}
	resourceName := "tfe_policy_set.foobar-tfpolicy"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCreateBusinessOrganizationNamed(t, orgName)
		},
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySetTFPolicy_workspaceIDs(orgName, "tfe_workspace.foo.id"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists(resourceName, policySet),
					testAccCheckTFEPolicySetKind(policySet, tfe.TFPolicy),
					testAccCheckTFEPolicySetWorkspaceCount(policySet, 1),
					resource.TestCheckResourceAttr(resourceName, "global", "false"),
					resource.TestCheckResourceAttr(resourceName, "workspace_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "workspace_ids.*", "tfe_workspace.foo", "id"),
				),
			},
			{
				Config: testAccTFEPolicySetTFPolicy_workspaceIDs(orgName, "tfe_workspace.foo.id, tfe_workspace.bar.id"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPtr(resourceName, "id", &policySet.ID),
					testAccCheckTFEPolicySetExists(resourceName, policySet),
					testAccCheckTFEPolicySetWorkspaceCount(policySet, 2),
					resource.TestCheckResourceAttr(resourceName, "workspace_ids.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "workspace_ids.*", "tfe_workspace.foo", "id"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "workspace_ids.*", "tfe_workspace.bar", "id"),
				),
			},
			{
				Config: testAccTFEPolicySetTFPolicy_workspaceIDs(orgName, "tfe_workspace.bar.id"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists(resourceName, policySet),
					testAccCheckTFEPolicySetWorkspaceCount(policySet, 1),
					resource.TestCheckResourceAttr(resourceName, "workspace_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "workspace_ids.*", "tfe_workspace.bar", "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTFEPolicySetTFPolicy_vcs(t *testing.T) {
	orgName := "tst-" + randomString(t)
	policySet := &tfe.PolicySet{}
	resourceName := "tfe_policy_set.foobar-tfpolicy"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			if envGithubToken == "" {
				t.Skip("Please set GITHUB_TOKEN to run this test")
			}
			if envGithubPolicySetIdentifier == "" {
				t.Skip("Please set GITHUB_POLICY_SET_IDENTIFIER to run this test")
			}
			if envGithubPolicySetPath == "" {
				t.Skip("Please set GITHUB_POLICY_SET_PATH to run this test")
			}
			testAccCreateBusinessOrganizationNamed(t, orgName)
		},
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySetTFPolicy_vcs(orgName, `"policies/**/*.hcl"`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists(resourceName, policySet),
					testAccCheckTFEPolicySetKind(policySet, tfe.TFPolicy),
					resource.TestCheckResourceAttr(resourceName, "kind", "tfpolicy"),
					resource.TestCheckResourceAttr(resourceName, "vcs_repo.0.identifier", envGithubPolicySetIdentifier),
					resource.TestCheckResourceAttr(resourceName, "vcs_repo.0.branch", "main"),
					resource.TestCheckResourceAttr(resourceName, "policies_path", envGithubPolicySetPath),
					resource.TestCheckResourceAttr(resourceName, "policy_update_patterns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "policy_update_patterns.0", "policies/**/*.hcl"),
				),
			},
			{
				Config: testAccTFEPolicySetTFPolicy_vcs(orgName, `"policies/**/*.hcl", "policy-1/**"`),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPtr(resourceName, "id", &policySet.ID),
					testAccCheckTFEPolicySetExists(resourceName, policySet),
					testAccCheckTFEPolicySetKind(policySet, tfe.TFPolicy),
					resource.TestCheckResourceAttr(resourceName, "policy_update_patterns.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "policy_update_patterns.0", "policies/**/*.hcl"),
					resource.TestCheckResourceAttr(resourceName, "policy_update_patterns.1", "policy-1/**"),
				),
			},
		},
	})
}

func testAccCheckTFEPolicySetWorkspaceCount(policySet *tfe.PolicySet, expected int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if len(policySet.Workspaces) != expected {
			return fmt.Errorf("expected policy set %s to have %d workspaces, got %d", policySet.ID, expected, len(policySet.Workspaces))
		}

		return nil
	}
}

func TestValidateTFPolicySetAttributes(t *testing.T) {
	testCases := map[string]struct {
		kind        tfe.PolicyKind
		config      cty.Value
		expectedErr string
	}{
		"sentinel allows overridable": {
			kind:   tfe.Sentinel,
			config: tfPolicySetRawConfig(cty.NullVal(cty.Bool), cty.True),
		},
		"opa allows overridable": {
			kind:   tfe.OPA,
			config: tfPolicySetRawConfig(cty.NullVal(cty.Bool), cty.True),
		},
		"opa allows agent_enabled": {
			kind:   tfe.OPA,
			config: tfPolicySetRawConfig(cty.False, cty.NullVal(cty.Bool)),
		},
		"sentinel allows agent_enabled": {
			kind:   tfe.Sentinel,
			config: tfPolicySetRawConfig(cty.False, cty.NullVal(cty.Bool)),
		},
		"tfpolicy allows omitted attributes": {
			kind:   tfe.TFPolicy,
			config: tfPolicySetRawConfig(cty.NullVal(cty.Bool), cty.NullVal(cty.Bool)),
		},
		"tfpolicy rejects overridable false": {
			kind:        tfe.TFPolicy,
			config:      tfPolicySetRawConfig(cty.NullVal(cty.Bool), cty.False),
			expectedErr: "overridable",
		},
		"tfpolicy rejects overridable true": {
			kind:        tfe.TFPolicy,
			config:      tfPolicySetRawConfig(cty.NullVal(cty.Bool), cty.True),
			expectedErr: "overridable",
		},
		"tfpolicy rejects agent_enabled false": {
			kind:        tfe.TFPolicy,
			config:      tfPolicySetRawConfig(cty.False, cty.NullVal(cty.Bool)),
			expectedErr: "agent_enabled",
		},
		"tfpolicy rejects agent_enabled true": {
			kind:        tfe.TFPolicy,
			config:      tfPolicySetRawConfig(cty.True, cty.NullVal(cty.Bool)),
			expectedErr: "agent_enabled",
		},
		"tfpolicy defers unknown attribute": {
			kind:   tfe.TFPolicy,
			config: tfPolicySetRawConfig(cty.NullVal(cty.Bool), cty.UnknownVal(cty.Bool)),
		},
		"tfpolicy defers unknown config": {
			kind:   tfe.TFPolicy,
			config: cty.DynamicVal,
		},
		"tfpolicy allows null config": {
			kind:   tfe.TFPolicy,
			config: cty.NullVal(cty.DynamicPseudoType),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			err := validateTFPolicySetAttributes(tc.config, tc.kind)

			if tc.expectedErr == "" {
				if err != nil {
					t.Fatalf("expected no error, got: %v", err)
				}
				return
			}

			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tc.expectedErr)
			}
			if !strings.Contains(err.Error(), tc.expectedErr) {
				t.Fatalf("expected error containing %q, got: %v", tc.expectedErr, err)
			}
		})
	}
}

func tfPolicySetRawConfig(agentEnabled, overridable cty.Value) cty.Value {
	return cty.ObjectVal(map[string]cty.Value{
		"agent_enabled": agentEnabled,
		"overridable":   overridable,
	})
}

func TestAccTFEPolicySet_updateOverridable(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}
	sha := genSentinelSha(t, "secret", "data")
	version := genSafeRandomOPAVersion()

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	policySet := &tfe.PolicySet{}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySetOPA_basic(org.Name, version, sha),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetAttributes(policySet),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "name", "tst-terraform"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "description", "Policy Set"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "kind", "opa"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "agent_enabled", "true"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "global", "false"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "overridable", "true"),
				),
			},

			{
				Config: testAccTFEPolicySetOPA_overridable(org.Name, version, sha),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.foobar", policySet),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "name", "tst-terraform-overridable"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "global", "false"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "kind", "opa"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "agent_enabled", "true"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "workspace_ids.#", "1"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "overridable", "false"),
				),
			},
		},
	})
}

func TestAccTFEPolicySet_update(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	policySet := &tfe.PolicySet{}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySet_basic(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetAttributes(policySet),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "name", "tst-terraform"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "description", "Policy Set"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "global", "false"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "policy_ids.#", "1"),
				),
			},

			{
				Config: testAccTFEPolicySet_populated(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetPopulated(policySet, org.Name),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "name", "terraform-populated"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "global", "false"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "kind", "sentinel"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "policy_ids.#", "1"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "workspace_ids.#", "1"),
				),
			},
		},
	})
}

func TestAccTFEPolicySet_updateEmpty(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	policySet := &tfe.PolicySet{}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySet_basic(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetAttributes(policySet),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "name", "tst-terraform"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "description", "Policy Set"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "global", "false"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "policy_ids.#", "1"),
				),
			},

			{
				Config: testAccTFEPolicySet_empty(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetAttributes(policySet),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "name", "tst-terraform"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "description", "Policy Set"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "global", "false"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "policy_ids.#", "0"),
				),
			},
		},
	})
}

func TestAccTFEPolicySet_updatePopulated(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	policySet := &tfe.PolicySet{}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySet_populated(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetPopulated(policySet, org.Name),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "name", "terraform-populated"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "global", "false"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "policy_ids.#", "1"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "workspace_ids.#", "1"),
				),
			},

			{
				Config: testAccTFEPolicySet_updatePopulated(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetPopulatedUpdated(policySet, org.Name),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "name", "terraform-populated-updated"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "global", "false"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "kind", "sentinel"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "policy_ids.#", "1"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "workspace_ids.#", "1"),
				),
			},
		},
	})
}

func TestAccTFEPolicySet_updateToGlobal(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	policySet := &tfe.PolicySet{}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySet_populated(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetPopulated(policySet, org.Name),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "name", "terraform-populated"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "global", "false"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "policy_ids.#", "1"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "workspace_ids.#", "1"),
				),
			},

			{
				Config: testAccTFEPolicySet_global(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetGlobal(policySet),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "name", "terraform-global"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "global", "true"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "kind", "sentinel"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "policy_ids.#", "1"),
				),
			},
		},
	})
}

func TestAccTFEPolicySet_updateToWorkspace(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	policySet := &tfe.PolicySet{}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySet_global(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetGlobal(policySet),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "name", "terraform-global"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "global", "true"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "policy_ids.#", "1"),
				),
			},

			{
				Config: testAccTFEPolicySet_populated(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetPopulated(policySet, org.Name),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "name", "terraform-populated"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "global", "false"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "policy_ids.#", "1"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "workspace_ids.#", "1"),
				),
			},
		},
	})
}

func TestAccTFEPolicySet_vcs(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	policySet := &tfe.PolicySet{}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			if envGithubToken == "" {
				t.Skip("Please set GITHUB_TOKEN to run this test")
			}
			if envGithubPolicySetIdentifier == "" {
				t.Skip("Please set GITHUB_POLICY_SET_IDENTIFIER to run this test")
			}
			if envGithubPolicySetBranch == "" {
				t.Skip("Please set GITHUB_POLICY_SET_BRANCH to run this test")
			}
			if envGithubPolicySetPath == "" {
				t.Skip("Please set GITHUB_POLICY_SET_PATH to run this test")
			}
		},
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySet_vcs(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetAttributes(policySet),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "name", "tst-terraform"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "description", "Policy Set"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "global", "false"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "vcs_repo.0.identifier", envGithubPolicySetIdentifier),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "vcs_repo.0.branch", "main"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "vcs_repo.0.ingress_submodules", "true"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "policies_path", envGithubPolicySetPath),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "policy_update_patterns.#", "2"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "policy_update_patterns.0", "**/*.sentinel"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "policy_update_patterns.1", "policies/**/*.hcl"),
				),
			},
		},
	})
}

func TestAccTFEPolicySet_GithubApp(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	policySet := &tfe.PolicySet{}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccGHAInstallationPreCheck(t)
			if envGithubToken == "" {
				t.Skip("Please set GITHUB_TOKEN to run this test")
			}
			if envGithubPolicySetIdentifier == "" {
				t.Skip("Please set GITHUB_POLICY_SET_IDENTIFIER to run this test")
			}
			if envGithubPolicySetBranch == "" {
				t.Skip("Please set GITHUB_POLICY_SET_BRANCH to run this test")
			}
			if envGithubPolicySetPath == "" {
				t.Skip("Please set GITHUB_POLICY_SET_PATH to run this test")
			}
		},
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySet_GithubApp(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetAttributes(policySet),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "name", "tst-terraform"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "description", "Policy Set"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "global", "false"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "vcs_repo.0.identifier", envGithubPolicySetIdentifier),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "vcs_repo.0.branch", "main"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "vcs_repo.0.ingress_submodules", "true"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "policies_path", envGithubPolicySetPath),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "policy_update_patterns.#", "2"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "policy_update_patterns.0", "**/*.sentinel"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "policy_update_patterns.1", "policies/**/*.hcl"),
				),
			},
		},
	})
}

func TestAccTFEPolicySet_updateVCSBranch(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	policySet := &tfe.PolicySet{}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			if envGithubToken == "" {
				t.Skip("Please set GITHUB_TOKEN to run this test")
			}
			if envGithubPolicySetIdentifier == "" {
				t.Skip("Please set GITHUB_POLICY_SET_IDENTIFIER to run this test")
			}
			if envGithubPolicySetBranch == "" {
				t.Skip("Please set GITHUB_POLICY_SET_BRANCH to run this test")
			}
			if envGithubPolicySetPath == "" {
				t.Skip("Please set GITHUB_POLICY_SET_PATH to run this test")
			}
		},
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySet_vcs(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetAttributes(policySet),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "name", "tst-terraform"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "description", "Policy Set"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "global", "false"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "vcs_repo.0.identifier", envGithubPolicySetIdentifier),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "vcs_repo.0.branch", "main"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "vcs_repo.0.ingress_submodules", "true"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "policies_path", envGithubPolicySetPath),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "policy_update_patterns.#", "2"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "policy_update_patterns.0", "**/*.sentinel"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "policy_update_patterns.1", "policies/**/*.hcl"),
				),
			},

			{
				Config: testAccTFEPolicySet_updateVCSBranch(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetAttributes(policySet),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "name", "tst-terraform"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "description", "Policy Set"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "global", "false"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "vcs_repo.0.identifier", envGithubPolicySetIdentifier),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "vcs_repo.0.branch", envGithubPolicySetBranch),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "vcs_repo.0.ingress_submodules", "true"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "policies_path", envGithubPolicySetPath),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "policy_update_patterns.#", "2"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "policy_update_patterns.0", "**/*.sentinel"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "policy_update_patterns.1", "policies/**/*.hcl"),
				),
			},
		},
	})
}

func TestAccTFEPolicySet_versionedSlug(t *testing.T) {
	skipIfUnitTest(t)

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	policySet := &tfe.PolicySet{}
	checksum, err := hashPolicies(testFixtureVersionFiles)
	if err != nil {
		t.Fatalf("Unable to generate checksum for policies %v", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySet_versionSlug(org.Name, testFixtureVersionFiles),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetAttributes(policySet),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "name", "tst-terraform"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "description", "Policy Set"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "slug.%", "2"),
					resource.TestCheckResourceAttrSet(
						"tfe_policy_set.foobar", "slug.source_path"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "slug.source_path", testFixtureVersionFiles),
					resource.TestCheckResourceAttrSet(
						"tfe_policy_set.foobar", "slug.id"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "slug.id", checksum),
				),
			},
		},
	})
}

func TestAccTFEPolicySet_versionedSlugUpdate(t *testing.T) {
	skipIfUnitTest(t)

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	policySet := &tfe.PolicySet{}

	originalChecksum, err := hashPolicies(testFixtureVersionFiles)
	if err != nil {
		t.Fatalf("Unable to generate checksum for policies %v", err)
	}

	newFile := fmt.Sprintf("%s/newfile.test.sentinel", testFixtureVersionFiles)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySet_versionSlug(org.Name, testFixtureVersionFiles),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetAttributes(policySet),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "name", "tst-terraform"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "description", "Policy Set"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "global", "false"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "slug.id", originalChecksum),
				),
			},
			{
				PreConfig: func() {
					err = os.WriteFile(newFile, []byte("main = rule { true }"), 0o755)
					if err != nil {
						t.Fatalf("error writing to file %s", newFile)
					}
					t.Cleanup(func() {
						os.Remove(newFile)
					})
				},
				Config: testAccTFEPolicySet_versionSlug(org.Name, testFixtureVersionFiles),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.foobar", policySet),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "name", "tst-terraform"),
					testAccCheckTFEPolicySetVersionValidateChecksum("tfe_policy_set.foobar", testFixtureVersionFiles),
				),
			},
		},
	})
}

func TestAccTFEPolicySet_versionedNoConflicts(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFEPolicySet_versionsConflict(org.Name, testFixtureVersionFiles),
				ExpectError: regexp.MustCompile(`Conflicting configuration`),
			},
		},
	})
}

func testAccCheckTFEPolicySetVersionValidateChecksum(n string, sourcePath string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		newChecksum, err := hashPolicies(sourcePath)
		if err != nil {
			return fmt.Errorf("unable to generate checksum for policies %w", err)
		}

		if rs.Primary.Attributes["slug.id"] != newChecksum {
			return fmt.Errorf("the new checksum for the policies contents did not match")
		}

		return nil
	}
}

func TestAccTFEPolicySet_invalidName(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFEPolicySet_invalidName(org.Name),
				ExpectError: regexp.MustCompile(`can only include letters, numbers, -, and _.`),
			},
		},
	})
}

func TestAccTFEPolicySetImport(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySet_populated(org.Name),
			},

			{
				ResourceName:      "tfe_policy_set.foobar",
				ImportState:       true,
				ImportStateVerify: true,
				// Note: We ignore the optional fields below, since the old API endpoints send empty values
				// and the results may vary depending on the API version
				ImportStateVerifyIgnore: []string{"kind", "overridable"},
			},
		},
	})
}

func TestAccTFEPolicySet_importByIdentity(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySet_populated(org.Name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectIdentity("tfe_policy_set.foobar", map[string]knownvalue.Check{
						"id":       knownvalue.NotNull(),
						"hostname": knownvalue.StringExact(os.Getenv("TFE_HOSTNAME")),
					}),
				},
			},

			{
				ResourceName:    "tfe_policy_set.foobar",
				ImportState:     true,
				ImportStateKind: resource.ImportBlockWithResourceIdentity,
				// Note: We ignore the optional fields below, since the old API endpoints send empty values
				// and the results may vary depending on the API version
				ImportStateVerifyIgnore: []string{"kind", "overridable"},
			},
		},
	})
}

func testAccCheckTFEPolicySetExists(n string, policySet *tfe.PolicySet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		ps, err := testAccConfiguredClient.Client.PolicySets.Read(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}

		if ps.ID != rs.Primary.ID {
			return fmt.Errorf("PolicySet not found")
		}

		*policySet = *ps

		return nil
	}
}

func testAccCheckTFEPolicySetAttributes(policySet *tfe.PolicySet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if policySet.Name != "tst-terraform" {
			return fmt.Errorf("Bad name: %s", policySet.Name)
		}

		if policySet.Description != "Policy Set" {
			return fmt.Errorf("Bad description: %s", policySet.Description)
		}

		if policySet.Global {
			return fmt.Errorf("Bad value for global: %v", policySet.Global)
		}

		return nil
	}
}

func testAccCheckTFEPolicySetPopulated(policySet *tfe.PolicySet, orgName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if policySet.Name != "terraform-populated" {
			return fmt.Errorf("Bad name: %s", policySet.Name)
		}

		if policySet.Global {
			return fmt.Errorf("Bad value for global: %v", policySet.Global)
		}

		if len(policySet.Policies) != 1 {
			return fmt.Errorf("Wrong number of policies: %v", len(policySet.Policies))
		}

		policyID := policySet.Policies[0].ID
		policy, _ := testAccConfiguredClient.Client.Policies.Read(ctx, policyID)
		if policy.Name != "policy-foo" {
			return fmt.Errorf("Wrong member policy: %v", policy.Name)
		}

		if len(policySet.Workspaces) != 1 {
			return fmt.Errorf("Wrong number of workspaces: %v", len(policySet.Workspaces))
		}

		workspaceID := policySet.Workspaces[0].ID
		workspace, _ := testAccConfiguredClient.Client.Workspaces.Read(ctx, orgName, "workspace-foo")
		if workspace.ID != workspaceID {
			return fmt.Errorf("Wrong member workspace: %v", workspace.Name)
		}

		return nil
	}
}

func testAccCheckTFEPolicySetPopulatedUpdated(policySet *tfe.PolicySet, orgName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if policySet.Name != "terraform-populated-updated" {
			return fmt.Errorf("Bad name: %s", policySet.Name)
		}

		if policySet.Global {
			return fmt.Errorf("Bad value for global: %v", policySet.Global)
		}

		if len(policySet.Policies) != 1 {
			return fmt.Errorf("Wrong number of policies: %v", len(policySet.Policies))
		}

		policyID := policySet.Policies[0].ID
		policy, _ := testAccConfiguredClient.Client.Policies.Read(ctx, policyID)
		if policy.Name != "policy-bar" {
			return fmt.Errorf("Wrong member policy: %v", policy.Name)
		}

		if len(policySet.Workspaces) != 1 {
			return fmt.Errorf("Wrong number of workspaces: %v", len(policySet.Workspaces))
		}

		workspaceID := policySet.Workspaces[0].ID
		workspace, _ := testAccConfiguredClient.Client.Workspaces.Read(ctx, orgName, "workspace-bar")
		if workspace.ID != workspaceID {
			return fmt.Errorf("Wrong member workspace: %v", workspace.Name)
		}

		return nil
	}
}

func testAccCheckTFEPolicySetGlobal(policySet *tfe.PolicySet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if policySet.Name != "terraform-global" {
			return fmt.Errorf("Bad name: %s", policySet.Name)
		}

		if !policySet.Global {
			return fmt.Errorf("Bad value for global: %v", policySet.Global)
		}

		if len(policySet.Policies) != 1 {
			return fmt.Errorf("Wrong number of policies: %v", len(policySet.Policies))
		}

		policyID := policySet.Policies[0].ID
		policy, _ := testAccConfiguredClient.Client.Policies.Read(ctx, policyID)
		if policy.Name != "policy-foo" {
			return fmt.Errorf("Wrong member policy: %v", policy.Name)
		}

		// No workspaces are returned for global policy sets
		if len(policySet.Workspaces) != 0 {
			return fmt.Errorf("Wrong number of workspaces: %v", len(policySet.Workspaces))
		}

		return nil
	}
}

func testAccCheckTFEPolicySetKind(policySet *tfe.PolicySet, kind tfe.PolicyKind) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if policySet.Kind != kind {
			return fmt.Errorf("Bad kind: expected %q, got %q", kind, policySet.Kind)
		}

		return nil
	}
}

func testAccCheckTFEPolicySetRecreated(before, after *tfe.PolicySet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before.ID == after.ID {
			return fmt.Errorf("Expected policy set to be recreated, but ID %q was reused", before.ID)
		}

		return nil
	}
}

func testAccCheckTFEPolicySetDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_policy_set" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := testAccConfiguredClient.Client.PolicySets.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Sentinel policy %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccTFEPolicySet_basic(organization string) string {
	return fmt.Sprintf(`
resource "tfe_sentinel_policy" "foo" {
  name         = "policy-foo"
  policy       = "main = rule { true }"
  organization = "%s"
}

resource "tfe_policy_set" "foobar" {
  name         = "tst-terraform"
  description  = "Policy Set"
  organization = "%s"
  policy_ids   = [tfe_sentinel_policy.foo.id]
}`, organization, organization)
}

func testAccTFEPolicySet_pinnedPolicyRuntimeVersion(organization string, version string) string {
	return fmt.Sprintf(`
resource "tfe_sentinel_policy" "foo" {
  name         = "policy-foo"
  policy       = "main = rule { true }"
  organization = "%s"
}

resource "tfe_policy_set" "foobar" {
  name         = "tst-terraform"
  description  = "Policy Set"
  organization = "%s"
  agent_enabled = true
  policy_tool_version = "%s"
  policy_ids   = [tfe_sentinel_policy.foo.id]
}`, organization, organization, version)
}

func testAccTFEPolicySetOPA_basic(organization string, version string, sha string) string {
	return fmt.Sprintf(`
resource "tfe_opa_version" "foobar" {
  version = "%s"
  url = "https://www.hashicorp.com"
  sha = "%s"
}

resource "tfe_policy_set" "foobar" {
  name         = "tst-terraform"
  description  = "Policy Set"
  organization = "%s"
  kind         = "opa"
  overridable = "true"
  agent_enabled = "true"
  policy_tool_version = "%s"
  depends_on = [tfe_opa_version.foobar]
}`, version, sha, organization, version)
}

func testAccTFEPolicySetTFPolicy_basic(organization string) string {
	return fmt.Sprintf(`
resource "tfe_policy_set" "foobar-tfpolicy" {
  name         = "tst-terraform-tfpolicy"
  description  = "TFPolicy Policy Set"
  organization = "%s"
  kind         = "tfpolicy"
}`, organization)
}

func testAccTFEPolicySetTFPolicy_updated(organization string, global bool) string {
	return fmt.Sprintf(`
resource "tfe_policy_set" "foobar-tfpolicy" {
  name         = "tst-terraform-tfpolicy-updated"
  description  = "Updated TFPolicy Policy Set"
  organization = "%s"
  kind         = "tfpolicy"
  global       = %t
}`, organization, global)
}

func testAccTFEPolicySetTFPolicy_managedToolVersion(organization string) string {
	return fmt.Sprintf(`
resource "tfe_policy_set" "foobar-tfpolicy" {
  name                = "tst-terraform-tfpolicy-managed"
  description         = "Managed TFPolicy Policy Set"
  organization        = "%s"
  kind                = "tfpolicy"
  policy_tool_version = "managed"
}`, organization)
}

func testAccTFEPolicySetTFPolicy_workspaceIDs(organization string, workspaceIDs string) string {
	return fmt.Sprintf(`
resource "tfe_workspace" "foo" {
  name         = "workspace-foo"
  organization = "%s"
}

resource "tfe_workspace" "bar" {
  name         = "workspace-bar"
  organization = "%s"
}

resource "tfe_policy_set" "foobar-tfpolicy" {
  name          = "tst-terraform-tfpolicy-workspaces"
  description   = "TFPolicy Policy Set"
  organization  = "%s"
  kind          = "tfpolicy"
  workspace_ids = [%s]
}`, organization, organization, organization, workspaceIDs)
}

func testAccTFEPolicySetTFPolicy_vcs(organization string, policyUpdatePatterns string) string {
	return fmt.Sprintf(`
resource "tfe_oauth_client" "test" {
  organization     = "%s"
  api_url          = "https://api.github.com"
  http_url         = "https://github.com"
  oauth_token      = "%s"
  service_provider = "github"
}

resource "tfe_policy_set" "foobar-tfpolicy" {
  name                   = "tst-terraform-tfpolicy-vcs"
  description            = "TFPolicy Policy Set"
  organization           = "%s"
  kind                   = "tfpolicy"
  policies_path          = "%s"
  policy_update_patterns = [%s]

  vcs_repo {
    identifier         = "%s"
    branch             = "main"
    ingress_submodules = true
    oauth_token_id     = tfe_oauth_client.test.oauth_token_id
  }
}`, organization,
		envGithubToken,
		organization,
		envGithubPolicySetPath,
		policyUpdatePatterns,
		envGithubPolicySetIdentifier,
	)
}

func testAccTFEPolicySet_empty(organization string) string {
	return fmt.Sprintf(`
 resource "tfe_policy_set" "foobar" {
  name         = "tst-terraform"
  description  = "Policy Set"
  organization = "%s"
}`, organization)
}

func testAccTFEPolicySet_populated(organization string) string {
	return fmt.Sprintf(`
locals {
    organization_name = "%s"
}

resource "tfe_sentinel_policy" "foo" {
  name         = "policy-foo"
  policy       = "main = rule { true }"
  organization = local.organization_name
}

resource "tfe_workspace" "foo" {
  name         = "workspace-foo"
  organization = local.organization_name
}

resource "tfe_policy_set" "foobar" {
  name          = "terraform-populated"
  organization = local.organization_name
  policy_ids    = [tfe_sentinel_policy.foo.id]
  workspace_ids = [tfe_workspace.foo.id]
}`, organization)
}

func testAccTFEPolicySetOPA_overridable(organization string, version string, sha string) string {
	return fmt.Sprintf(`
locals {
    organization_name = "%s"
}

resource "tfe_opa_version" "foobar" {
  version = "%s"
  url = "https://www.hashicorp.com"
  sha = "%s"
}

resource "tfe_workspace" "foo" {
  name         = "workspace-foo"
  organization = local.organization_name
}

resource "tfe_policy_set" "foobar" {
  name          = "tst-terraform-overridable"
  organization = local.organization_name
  workspace_ids = [tfe_workspace.foo.id]
  overridable = "false"
  agent_enabled = "true"
  policy_tool_version = "%s"
  kind = "opa"
}`, organization, version, sha, version)
}

func testAccTFEPolicySet_updatePopulated(organization string) string {
	return fmt.Sprintf(`
locals {
    organization_name = "%s"
}

resource "tfe_sentinel_policy" "foo" {
  name         = "policy-foo"
  policy       = "main = rule { true }"
  organization = local.organization_name
}

resource "tfe_sentinel_policy" "bar" {
  name         = "policy-bar"
  policy       = "main = rule { false }"
  organization = local.organization_name
}

resource "tfe_workspace" "foo" {
  name         = "workspace-foo"
  organization = local.organization_name
}

resource "tfe_workspace" "bar" {
  name         = "workspace-bar"
  organization = local.organization_name
}

resource "tfe_policy_set" "foobar" {
  name          = "terraform-populated-updated"
  organization = local.organization_name
  policy_ids    = [tfe_sentinel_policy.bar.id]
  workspace_ids = [tfe_workspace.bar.id]
}`, organization)
}

func testAccTFEPolicySet_global(organization string) string {
	return fmt.Sprintf(`
locals {
    organization_name = "%s"
}

resource "tfe_sentinel_policy" "foo" {
  name         = "policy-foo"
  policy       = "main = rule { true }"
  organization = local.organization_name
}

resource "tfe_workspace" "foo" {
  name         = "workspace-foo"
  organization = local.organization_name
}

resource "tfe_policy_set" "foobar" {
  name         = "terraform-global"
  organization = local.organization_name
  global       = true
  policy_ids   = [tfe_sentinel_policy.foo.id]
}`, organization)
}

func testAccTFEPolicySet_vcs(organization string) string {
	return fmt.Sprintf(`
locals {
    organization_name = "%s"
}

resource "tfe_oauth_client" "test" {
  organization     = local.organization_name
  api_url          = "https://api.github.com"
  http_url         = "https://github.com"
  oauth_token      = "%s"
  service_provider = "github"
}

resource "tfe_policy_set" "foobar" {
  name         = "tst-terraform"
  description  = "Policy Set"
  organization     = local.organization_name
  vcs_repo {
    identifier         = "%s"
    branch             = "main"
    ingress_submodules = true
    oauth_token_id     = tfe_oauth_client.test.oauth_token_id
  }

  policies_path = "%s"
}
`, organization,
		envGithubToken,
		envGithubPolicySetIdentifier,
		envGithubPolicySetPath,
	)
}

func testAccTFEPolicySet_GithubApp(organization string) string {
	return fmt.Sprintf(`
locals {
    organization_name = "%s"
}

resource "tfe_policy_set" "foobar" {
  name         = "tst-terraform"
  description  = "Policy Set"
  organization     = local.organization_name
  vcs_repo {
    identifier         = "%s"
    branch             = "main"
    ingress_submodules = true
    github_app_installation_id = "%s"
  }

  policies_path = "%s"
}
`, organization,
		envGithubPolicySetIdentifier,
		envGithubAppInstallationID,
		envGithubPolicySetPath,
	)
}

func testAccTFEPolicySet_updateVCSBranch(organization string) string {
	return fmt.Sprintf(`
locals {
    organization_name = "%s"
}

resource "tfe_oauth_client" "test" {
  organization     = local.organization_name
  api_url          = "https://api.github.com"
  http_url         = "https://github.com"
  oauth_token      = "%s"
  service_provider = "github"
}

resource "tfe_policy_set" "foobar" {
  name         = "tst-terraform"
  description  = "Policy Set"
  organization     = local.organization_name
  vcs_repo {
    identifier         = "%s"
    branch             = "%s"
    ingress_submodules = true
    oauth_token_id     = tfe_oauth_client.test.oauth_token_id
  }

  policies_path = "%s"
}
`, organization,
		envGithubToken,
		envGithubPolicySetIdentifier,
		envGithubPolicySetBranch,
		envGithubPolicySetPath,
	)
}

func testAccTFEPolicySet_invalidName(organization string) string {
	return fmt.Sprintf(`
locals {
    organization_name = "%s"
}

resource "tfe_sentinel_policy" "foo" {
  name         = "policy-foo"
  policy       = "main = rule { true }"
  organization = local.organization_name
}

resource "tfe_policy_set" "foobar" {
  name         = "not the right format"
  description  = "Policy Set"
  organization = local.organization_name
  policy_ids   = [tfe_sentinel_policy.foo.id]
}`, organization)
}

func testAccTFEPolicySet_versionSlug(organization string, sourcePath string) string {
	return fmt.Sprintf(`
data "tfe_slug" "policy" {
  source_path = "%s"
}

resource "tfe_policy_set" "foobar" {
  name         = "tst-terraform"
  description  = "Policy Set"
  organization = "%s"
  slug         = data.tfe_slug.policy
}`, sourcePath, organization)
}

func testAccTFEPolicySet_versionsConflict(organization string, sourcePath string) string {
	return fmt.Sprintf(`
data "tfe_slug" "policy" {
  source_path = "%s"
}

resource "tfe_policy_set" "foobar" {
  name         = "tst-terraform"
  description  = "Policy Set"
  organization = "%s"
	slug = data.tfe_slug.policy
  vcs_repo {
    identifier         = "foo"
    branch             = "foo"
    ingress_submodules = true
    oauth_token_id     = "id"
  }
} `, sourcePath, organization)
}

func TestAccTFEPolicySet_tagMatchLogicAll(t *testing.T) {

	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createOrganization(t, tfeClient, tfe.OrganizationCreateOptions{
		Name:  tfe.String("tst-" + randomString(t)),
		Email: tfe.String(fmt.Sprintf("%s@tfe.local", randomString(t))),
	})
	t.Cleanup(orgCleanup)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySet_tagMatchLogicWithTagSelectorAll(org.Name, rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tfe_policy_set.test", "tag_match_logic", "all"),
				),
			},
			{
				// Update tag_match_logic to "any".
				Config: testAccTFEPolicySet_tagMatchLogicWithTagSelectorAny(org.Name, rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tfe_policy_set.test", "tag_match_logic", "any"),
				),
			},
			{
				ResourceName:            "tfe_policy_set.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"overridable"},
			},
			{
				Config: testAccTFEPolicySet_tagMatchLogicNoTagSelectors(org.Name, rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tfe_policy_set.test", "tag_match_logic", ""),
				),
			},
			{
				Config: testAccTFEPolicySet_tagMatchLogicWorkspaceScope(org.Name, rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tfe_policy_set.test", "tag_match_logic", ""),
				),
			},
			{
				// Switch back from workspace scope to tag-based scoping with "any".
				// Verifies the full round-trip: tag → workspace → tag.
				Config: testAccTFEPolicySet_tagMatchLogicWithTagSelectorAny(org.Name, rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tfe_policy_set.test", "tag_match_logic", "any"),
				),
			},
			{
				// Upgrade tag_match_logic from "any" to "all" to verify the full cycle.
				Config: testAccTFEPolicySet_tagMatchLogicWithTagSelectorAll(org.Name, rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tfe_policy_set.test", "tag_match_logic", "all"),
				),
			},
		},
	})
}

func testAccTFEPolicySet_tagMatchLogicWithTagSelectorAll(organization string, rInt int) string {
	return fmt.Sprintf(`
resource "tfe_workspace" "test" {
  name         = "tst-workspace-%d"
  organization = %q
  tags = {
    env    = "prod"
    region = "us-east-1"
  }
}

resource "tfe_policy_set" "test" {
  name            = "tst-tag-match-%d"
  organization    = %q
  tag_match_logic = "all"
}

resource "tfe_tag_policy_set" "test" {
  policy_set_id = tfe_policy_set.test.id
  key           = "env"
  value         = "prod"
  depends_on    = [tfe_workspace.test]
}

resource "tfe_tag_policy_set" "test2" {
  policy_set_id = tfe_policy_set.test.id
  key           = "region"
  value         = "us-east-1"
  depends_on    = [tfe_workspace.test]
}`, rInt, organization, rInt, organization)
}

func testAccTFEPolicySet_tagMatchLogicWithTagSelectorAny(organization string, rInt int) string {
	return fmt.Sprintf(`
resource "tfe_workspace" "test" {
  name         = "tst-workspace-%d"
  organization = %q
  tags = {
    env    = "prod"
    region = "us-east-1"
  }
}

resource "tfe_policy_set" "test" {
  name            = "tst-tag-match-%d"
  organization    = %q
  tag_match_logic = "any"
}

resource "tfe_tag_policy_set" "test" {
  policy_set_id = tfe_policy_set.test.id
  key           = "env"
  value         = "prod"
  depends_on    = [tfe_workspace.test]
}

resource "tfe_tag_policy_set" "test2" {
  policy_set_id = tfe_policy_set.test.id
  key           = "region"
  value         = "us-east-1"
  depends_on    = [tfe_workspace.test]
}`, rInt, organization, rInt, organization)
}

func testAccTFEPolicySet_tagMatchLogicNoTagSelectors(organization string, rInt int) string {
	return fmt.Sprintf(`
resource "tfe_workspace" "test" {
  name         = "tst-workspace-%d"
  organization = %q
  tags = {
    env    = "prod"
    region = "us-east-1"
  }
}

resource "tfe_policy_set" "test" {
  name            = "tst-tag-match-%d"
  organization    = %q
}`, rInt, organization, rInt, organization)
}

func testAccTFEPolicySet_tagMatchLogicWorkspaceScope(organization string, rInt int) string {
	return fmt.Sprintf(`
resource "tfe_workspace" "test" {
  name         = "tst-workspace-%d"
  organization = %q
  tags = {
    env    = "prod"
    region = "us-east-1"
  }
}

resource "tfe_policy_set" "test" {
  name         = "tst-tag-match-%d"
  organization = %q
}

resource "tfe_workspace_policy_set" "test" {
  policy_set_id = tfe_policy_set.test.id
  workspace_id  = tfe_workspace.test.id
}`, rInt, organization, rInt, organization)
}

func TestAccTFEPolicySet_tagMatchLogicExclusion(t *testing.T) {

	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createOrganization(t, tfeClient, tfe.OrganizationCreateOptions{
		Name:  tfe.String("tst-" + randomString(t)),
		Email: tfe.String(fmt.Sprintf("%s@tfe.local", randomString(t))),
	})
	t.Cleanup(orgCleanup)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySet_tagMatchLogicExclusionAll(org.Name, rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tfe_policy_set.test", "tag_match_logic", "all"),
				),
			},
			{
				Config: testAccTFEPolicySet_tagMatchLogicExclusionAny(org.Name, rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tfe_policy_set.test", "tag_match_logic", "any"),
				),
			},
			{
				ResourceName:            "tfe_policy_set.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"overridable"},
			},
			{
				Config: testAccTFEPolicySet_tagMatchLogicExclusionNoTagSelectors(org.Name, rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tfe_policy_set.test", "tag_match_logic", ""),
				),
			},
			{
				Config: testAccTFEPolicySet_tagMatchLogicExclusionWorkspaceScope(org.Name, rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tfe_policy_set.test", "tag_match_logic", ""),
				),
			},
			{
				// Switch back from workspace scope to tag-based scoping with "any".
				// Verifies the full round-trip: tag → workspace → tag.
				Config: testAccTFEPolicySet_tagMatchLogicExclusionAny(org.Name, rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tfe_policy_set.test", "tag_match_logic", "any"),
				),
			},
			{
				// Upgrade tag_match_logic from "any" to "all" to verify the full cycle.
				Config: testAccTFEPolicySet_tagMatchLogicExclusionAll(org.Name, rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tfe_policy_set.test", "tag_match_logic", "all"),
				),
			},
		},
	})
}

func testAccTFEPolicySet_tagMatchLogicExclusionAll(organization string, rInt int) string {
	return fmt.Sprintf(`
resource "tfe_workspace" "test" {
  name         = "tst-workspace-%d"
  organization = %q
  tags = {
    env    = "staging"
    region = "us-west-2"
  }
}

resource "tfe_policy_set" "test" {
  name            = "tst-tag-match-excl-%d"
  organization    = %q
  global          = true
  tag_match_logic = "all"
}

resource "tfe_tag_policy_set_exclusion" "test" {
  policy_set_id = tfe_policy_set.test.id
  key           = "env"
  value         = "staging"
  depends_on    = [tfe_workspace.test]
}

resource "tfe_tag_policy_set_exclusion" "test2" {
  policy_set_id = tfe_policy_set.test.id
  key           = "region"
  value         = "us-west-2"
  depends_on    = [tfe_workspace.test]
}`, rInt, organization, rInt, organization)
}

func testAccTFEPolicySet_tagMatchLogicExclusionAny(organization string, rInt int) string {
	return fmt.Sprintf(`
resource "tfe_workspace" "test" {
  name         = "tst-workspace-%d"
  organization = %q
  tags = {
    env    = "staging"
    region = "us-west-2"
  }
}

resource "tfe_policy_set" "test" {
  name            = "tst-tag-match-excl-%d"
  organization    = %q
  global          = true
  tag_match_logic = "any"
}

resource "tfe_tag_policy_set_exclusion" "test" {
  policy_set_id = tfe_policy_set.test.id
  key           = "env"
  value         = "staging"
  depends_on    = [tfe_workspace.test]
}

resource "tfe_tag_policy_set_exclusion" "test2" {
  policy_set_id = tfe_policy_set.test.id
  key           = "region"
  value         = "us-west-2"
  depends_on    = [tfe_workspace.test]
}`, rInt, organization, rInt, organization)
}

func testAccTFEPolicySet_tagMatchLogicExclusionNoTagSelectors(organization string, rInt int) string {
	return fmt.Sprintf(`
resource "tfe_workspace" "test" {
  name         = "tst-workspace-%d"
  organization = %q
  tags = {
    env    = "staging"
    region = "us-west-2"
  }
}

resource "tfe_policy_set" "test" {
  name            = "tst-tag-match-excl-%d"
  organization    = %q
}`, rInt, organization, rInt, organization)
}

func testAccTFEPolicySet_tagMatchLogicExclusionWorkspaceScope(organization string, rInt int) string {
	return fmt.Sprintf(`
resource "tfe_workspace" "test" {
  name         = "tst-workspace-%d"
  organization = %q
  tags = {
    env    = "staging"
    region = "us-west-2"
  }
}

resource "tfe_policy_set" "test" {
  name         = "tst-tag-match-excl-%d"
  organization = %q
}

resource "tfe_workspace_policy_set_exclusion" "test" {
  policy_set_id = tfe_policy_set.test.id
  workspace_id  = tfe_workspace.test.id
}`, rInt, organization, rInt, organization)
}
