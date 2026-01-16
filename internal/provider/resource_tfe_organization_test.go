// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccTFEOrganization_basic(t *testing.T) {
	org := &tfe.Organization{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEOrganizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrganization_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOrganizationExists(
						"tfe_organization.foobar", org),
					testAccCheckTFEOrganizationAttributesBasic(org, orgName),
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "name", orgName),
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "email", "admin@company.com"),
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "collaborator_auth_policy", "password"),
				),
			},
		},
	})
}

func TestAccTFEOrganization_full(t *testing.T) {
	org := &tfe.Organization{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEOrganizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrganization_full(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOrganizationExists(
						"tfe_organization.foobar", org),
					testAccCheckTFEOrganizationAttributesFull(org, orgName),
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "name", orgName),
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "email", "admin@company.com"),
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "session_timeout_minutes", "30"),
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "session_remember_minutes", "30"),
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "collaborator_auth_policy", "password"),
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "owners_team_saml_role_id", "owners"),
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "cost_estimation_enabled", "false"),
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "send_passing_statuses_for_untriggered_speculative_plans", "false"),
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "aggregated_commit_status_enabled", "true"),
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "assessments_enforced", "false"),
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "allow_force_delete_workspaces", "false"),
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "speculative_plan_management_enabled", "true"),
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "user_tokens_enabled", "true"),
				),
			},
		},
	})
}

func TestAccTFEOrganization_defaultProject(t *testing.T) {
	org := &tfe.Organization{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEOrganizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrganization_full(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOrganizationExists(
						"tfe_organization.foobar", org),
					resource.TestCheckResourceAttrWith("tfe_organization.foobar", "default_project_id", func(value string) error {
						if value == "" {
							return errors.New("default project ID not exposed")
						}
						return nil
					}),
				),
			},
		},
	})
}

func TestAccTFEOrganization_update_costEstimation(t *testing.T) {
	t.Skip("Skipping this test until the SDK can support importing resources before applying a configuration")

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	// First update
	costEstimationEnabled1 := true
	assessmentsEnforced1 := true
	allowForceDeleteWorkspaces1 := true

	// Second update
	costEstimationEnabled2 := false
	assessmentsEnforced2 := false
	allowForceDeleteWorkspaces2 := false
	updatedName := org.Name + "_foobar"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEOrganizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrganization_update(org.Name, org.Email, costEstimationEnabled1, assessmentsEnforced1, allowForceDeleteWorkspaces1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOrganizationExists(
						"tfe_organization.foobar", org),
					testAccCheckTFEOrganizationAttributesUpdated(org, org.Name, costEstimationEnabled1),
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "name", org.Name),
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "email", "admin-updated@company.com"),
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "session_timeout_minutes", "3600"),
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "session_remember_minutes", "3600"),
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "collaborator_auth_policy", "password"),
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "owners_team_saml_role_id", "owners"),
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "cost_estimation_enabled", strconv.FormatBool(costEstimationEnabled1)),
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "send_passing_statuses_for_untriggered_speculative_plans", "false"),
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "aggregated_commit_status_enabled", "true"),
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "assessments_enforced", strconv.FormatBool(assessmentsEnforced1)),
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "allow_force_delete_workspaces", strconv.FormatBool(allowForceDeleteWorkspaces1)),
				),
			},

			{
				Config: testAccTFEOrganization_update(updatedName, org.Email, costEstimationEnabled2, assessmentsEnforced2, allowForceDeleteWorkspaces2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOrganizationExists(
						"tfe_organization.foobar", org),
					testAccCheckTFEOrganizationAttributesUpdated(org, updatedName, costEstimationEnabled2),
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "name", updatedName),
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "email", "admin-updated@company.com"),
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "session_timeout_minutes", "3600"),
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "session_remember_minutes", "3600"),
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "collaborator_auth_policy", "password"),
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "owners_team_saml_role_id", "owners"),
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "cost_estimation_enabled", strconv.FormatBool(costEstimationEnabled2)),
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "assessments_enforced", strconv.FormatBool(assessmentsEnforced2)),
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "allow_force_delete_workspaces", strconv.FormatBool(allowForceDeleteWorkspaces2)),
				),
			},
		},
	})
}

func TestAccTFEOrganization_user_tokens_enabled(t *testing.T) {
	// this test is a bit tricky because once user tokens are disabled, we cannot use a user token to re-enable them
	// through the API.
	// Therefore, we need to create an org, generate an owners team token for that org, and then use that token
	// in the go-tfe client to test the user_tokens_enabled setting.

	org := &tfe.Organization{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	customClient, err := getClientUsingEnv()

	if err != nil {
		t.Error(err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckTFEOrganizationDestroy,
		ProtoV6ProviderFactories: muxedProvidersWithCustomClient(func() *tfe.Client { return customClient }),
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrganization_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOrganizationExists(
						"tfe_organization.foobar", org),
					testAccCheckTFEOrganizationAttributesBasic(org, orgName),
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "user_tokens_enabled", "true"),
					testAccCheckTFEOrganizationUserTokensEnabled(org, orgName, true),
				),
			},
			{
				PreConfig: func() {
					// create a client for the owners team in the org,
					// then update the custom client pointer, so it is picked up when the provider is reinitialized
					// during Config steps
					customClient, err = getOwnerTeamClientForOrg(t, orgName)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testAccTFEOrganization_userTokensEnabled(rInt, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOrganizationExists(
						"tfe_organization.foobar", org),
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "user_tokens_enabled", "false"),
					testAccCheckTFEOrganizationUserTokensEnabled(org, orgName, false),
				),
			},
			{
				Config: testAccTFEOrganization_userTokensEnabled(rInt, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOrganizationExists(
						"tfe_organization.foobar", org),
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "user_tokens_enabled", "true"),
					testAccCheckTFEOrganizationUserTokensEnabled(org, orgName, true),
				),
			},
		},
	})
}

func TestAccTFEOrganization_EnforceHYOK(t *testing.T) {
	skipUnlessHYOKEnabled(t)

	t.Skip("Skipping test until HYOK configurations can be promoted to primary through the provider. Currently," +
		"even if promotion is possible, primary configurations cannot be deleted and leaves dangling resources.")

	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)
	org := &tfe.Organization{}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEOrganizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrganization_updateEnforceHYOK(orgName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOrganizationExists(
						"tfe_organization.foobar", org),
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "enforce_hyok", "false"),
				),
			},
			{
				PreConfig: func() {
					newSubscriptionUpdater(org).WithPremiumPlan().Update(t)
				},
				Config: testAccTFEOrganization_updateEnforceHYOK(orgName, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "enforce_hyok", "true"),
				),
			},
		},
	})
}

func TestAccTFEOrganization_case(t *testing.T) {
	org := &tfe.Organization{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEOrganizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrganization_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOrganizationExists(
						"tfe_organization.foobar", org),
					testAccCheckTFEOrganizationAttributesBasic(org, orgName),
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "name", orgName),
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "email", "admin@company.com"),
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "collaborator_auth_policy", "password"),
				),
			},
			{
				Config: testAccTFEOrganization_title_case(rInt),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccTFEOrganization_import(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEOrganizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrganization_basic(rInt),
			},

			{
				ResourceName:      "tfe_organization.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func getOwnerTeamClientForOrg(t *testing.T, orgName string) (*tfe.Client, error) {
	ownersTeams, err := testAccConfiguredClient.Client.Teams.List(ctx, orgName, &tfe.TeamListOptions{
		Names: []string{"owners"},
	})
	if err != nil {
		return nil, err
	}
	if len(ownersTeams.Items) != 1 {
		return nil, fmt.Errorf("expected to find 1 owners team, found %d", len(ownersTeams.Items))
	}
	ownersTeam := ownersTeams.Items[0]

	teamToken, err := testAccConfiguredClient.Client.TeamTokens.Create(ctx, ownersTeam.ID)
	if err != nil {
		return nil, err
	}
	return getClientWithToken(teamToken.Token)
}

func testAccCheckTFEOrganizationExists(
	n string, org *tfe.Organization) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		o, err := testAccConfiguredClient.Client.Organizations.Read(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}

		if o.Name != rs.Primary.ID {
			return fmt.Errorf("Organization not found")
		}

		*org = *o

		return nil
	}
}

func testAccCheckTFEOrganizationAttributesBasic(
	org *tfe.Organization, expectedOrgName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if org.Name != expectedOrgName {
			return fmt.Errorf("Bad name: %s", org.Name)
		}

		if org.Email != "admin@company.com" {
			return fmt.Errorf("Bad email: %s", org.Email)
		}

		if org.CollaboratorAuthPolicy != tfe.AuthPolicyPassword {
			return fmt.Errorf("Bad auth policy: %s", org.CollaboratorAuthPolicy)
		}

		return nil
	}
}

func testAccCheckTFEOrganizationAttributesFull(
	org *tfe.Organization, expectedOrgName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if org.Name != expectedOrgName {
			return fmt.Errorf("Bad name: %s", org.Name)
		}

		if org.Email != "admin@company.com" {
			return fmt.Errorf("Bad email: %s", org.Email)
		}

		if org.SessionTimeout != 30 {
			return fmt.Errorf("Bad session timeout minutes: %d", org.SessionTimeout)
		}

		if org.SessionRemember != 30 {
			return fmt.Errorf("Bad session remember minutes: %d", org.SessionRemember)
		}

		if org.CollaboratorAuthPolicy != tfe.AuthPolicyPassword {
			return fmt.Errorf("Bad auth policy: %s", org.CollaboratorAuthPolicy)
		}

		if org.OwnersTeamSAMLRoleID != "owners" {
			return fmt.Errorf("Bad owners team SAML role ID: %s", org.OwnersTeamSAMLRoleID)
		}

		if org.CostEstimationEnabled != false {
			return fmt.Errorf("Bad cost-estimation-enabled: %t", org.CostEstimationEnabled)
		}

		return nil
	}
}

func testAccCheckTFEOrganizationUserTokensEnabled(
	org *tfe.Organization, expectedOrgName string, expectedUserTokensEnabled bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if org.Name != expectedOrgName {
			return fmt.Errorf("Expected org.Name to be: %s, but received: %s", expectedOrgName, org.Name)
		}

		if org.UserTokensEnabled != nil && *org.UserTokensEnabled != expectedUserTokensEnabled {
			return fmt.Errorf("Expected org.UserTokensEnabled to be: %v, but received: %v", expectedUserTokensEnabled, *org.UserTokensEnabled)
		}
		return nil
	}
}

func testAccCheckTFEOrganizationAttributesUpdated(
	org *tfe.Organization, expectedOrgName string, expectedCostEstimationEnabled bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if org.Name != expectedOrgName {
			return fmt.Errorf("Bad name: %s", org.Name)
		}

		if org.Email != "admin-updated@company.com" {
			return fmt.Errorf("Bad email: %s", org.Email)
		}

		if org.SessionTimeout != 3600 {
			return fmt.Errorf("Bad session timeout minutes: %d", org.SessionTimeout)
		}

		if org.SessionRemember != 3600 {
			return fmt.Errorf("Bad session remember minutes: %d", org.SessionRemember)
		}

		if org.CollaboratorAuthPolicy != tfe.AuthPolicyPassword {
			return fmt.Errorf("Bad auth policy: %s", org.CollaboratorAuthPolicy)
		}

		if org.OwnersTeamSAMLRoleID != "owners" {
			return fmt.Errorf("Bad owners team SAML role ID: %s", org.OwnersTeamSAMLRoleID)
		}

		if org.CostEstimationEnabled != expectedCostEstimationEnabled {
			return fmt.Errorf("Bad cost-estimation-enabled: %t", org.CostEstimationEnabled)
		}

		return nil
	}
}

func testAccCheckTFEOrganizationDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_organization" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := testAccConfiguredClient.Client.Organizations.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Organization %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccTFEOrganization_basic(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}`, rInt)
}

func testAccTFEOrganization_title_case(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "Tst-Terraform-%d"
  email = "admin@company.com"
}`, rInt)
}

func testAccTFEOrganization_full(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name                              = "tst-terraform-%d"
  email                             = "admin@company.com"
  session_timeout_minutes           = 30
  session_remember_minutes          = 30
  collaborator_auth_policy          = "password"
  owners_team_saml_role_id          = "owners"
  cost_estimation_enabled           = false
  assessments_enforced              = false
  allow_force_delete_workspaces     = false
}`, rInt)
}

func testAccTFEOrganization_update(orgName string, orgEmail string, costEstimationEnabled bool, assessmentsEnforced bool, allowForceDeleteWorkspaces bool) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name                              = "%s"
  email                             = "%s"
  session_timeout_minutes           = 3600
  session_remember_minutes          = 3600
  owners_team_saml_role_id          = "owners"
  cost_estimation_enabled           = %t
  assessments_enforced              = %t
  allow_force_delete_workspaces     = %t
}`, orgName, orgEmail, costEstimationEnabled, assessmentsEnforced, allowForceDeleteWorkspaces)
}

func testAccTFEOrganization_userTokensEnabled(rInt int, userTokensEnabled bool) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
  user_tokens_enabled = %t
}`, rInt, userTokensEnabled)
}

func testAccTFEOrganization_updateEnforceHYOK(orgName string, enforceHYOK bool) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name         = "%s"
  email        = "admin@company.com"
  enforce_hyok = %t
}`, orgName, enforceHYOK)
}
