package tfe

import (
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccTFEOrganization_basic(t *testing.T) {
	org := &tfe.Organization{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEOrganizationDestroy,
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
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEOrganizationDestroy,
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
				),
			},
		},
	})
}

func TestAccTFEOrganization_update_costEstimation(t *testing.T) {
	skipIfFreeOnly(t)

	org := &tfe.Organization{}

	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	// First update
	rInt1 := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	costEstimationEnabled1 := true

	// Second update
	rInt2 := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	costEstimationEnabled2 := false

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEOrganizationDestroy,
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
				Config: testAccTFEOrganization_update(rInt1, costEstimationEnabled1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOrganizationExists(
						"tfe_organization.foobar", org),
					testAccCheckTFEOrganizationAttributesUpdated(org, fmt.Sprintf("tst-terraform-%d", rInt1), costEstimationEnabled1),
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "name", fmt.Sprintf("tst-terraform-%d", rInt1)),
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
				),
			},

			{
				Config: testAccTFEOrganization_update(rInt2, costEstimationEnabled2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOrganizationExists(
						"tfe_organization.foobar", org),
					testAccCheckTFEOrganizationAttributesUpdated(org, fmt.Sprintf("tst-terraform-%d", rInt2), costEstimationEnabled2),
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "name", fmt.Sprintf("tst-terraform-%d", rInt2)),
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
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEOrganizationDestroy,
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
				Config:   testAccTFEOrganization_title_case(rInt),
				PlanOnly: true,
			},
		},
	})
}

func TestAccTFEOrganization_import(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEOrganizationDestroy,
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

func TestAccTFEOrganization_update_workspaceLimit(t *testing.T) {
	skipIfCloud(t)

	org := &tfe.Organization{}

	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	// First update
	initialWorkspaceLimit := 42

	// Second update
	updatedWorkspaceLimit := 1337

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEOrganizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrganization_workspaceLimited(rInt, initialWorkspaceLimit),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOrganizationExists(
						"tfe_organization.foobar", org),
					testAccCheckTFEOrganizationAttributesBasic(org, orgName),
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "name", orgName),
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "email", "admin@company.com"),
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "admin_settings.0.workspace_limit", "42"),
				),
			},
			{
				Config: testAccTFEOrganization_workspaceLimited(rInt, updatedWorkspaceLimit),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOrganizationExists(
						"tfe_organization.foobar", org),
					testAccCheckTFEOrganizationAttributesBasic(org, orgName),
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "name", orgName),
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "email", "admin@company.com"),
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "admin_settings.0.workspace_limit", "1337"),
				),
			},
			{
				Config: testAccTFEOrganization_workspaceNil(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOrganizationExists(
						"tfe_organization.foobar", org),
					testAccCheckTFEOrganizationAttributesBasic(org, orgName),
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "name", orgName),
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "email", "admin@company.com"),
					resource.TestCheckNoResourceAttr(
						"tfe_organization.foobar", "admin_settings.0.workspace_limit"),
				),
			},
		},
	})
}

func testAccCheckTFEOrganizationExists(
	n string, org *tfe.Organization) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		tfeClient := testAccProvider.Meta().(*tfe.Client)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		o, err := tfeClient.Organizations.Read(ctx, rs.Primary.ID)
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
	tfeClient := testAccProvider.Meta().(*tfe.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_organization" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := tfeClient.Organizations.Read(ctx, rs.Primary.ID)
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
  name                     = "tst-terraform-%d"
  email                    = "admin@company.com"
  session_timeout_minutes  = 30
  session_remember_minutes = 30
  collaborator_auth_policy = "password"
  owners_team_saml_role_id = "owners"
  cost_estimation_enabled  = false
}`, rInt)
}

func testAccTFEOrganization_update(rInt int, costEstimationEnabled bool) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name                     = "tst-terraform-%d"
  email                    = "admin-updated@company.com"
  session_timeout_minutes  = 3600
  session_remember_minutes = 3600
  owners_team_saml_role_id = "owners"
  cost_estimation_enabled  = %t
}`, rInt, costEstimationEnabled)
}

func testAccTFEOrganization_workspaceLimited(rInt int, workspaceLimit int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
  admin_settings {
    workspace_limit = %d
  }
}`, rInt, workspaceLimit)
}

func testAccTFEOrganization_workspaceNil(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}`, rInt)
}
