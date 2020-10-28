package tfe

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
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
					testAccCheckTFEOrganizationAttributes(org, orgName),
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

func TestAccTFEOrganization_update(t *testing.T) {
	org := &tfe.Organization{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	rInt1 := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
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
					testAccCheckTFEOrganizationAttributes(org, orgName),
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "name", orgName),
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "email", "admin@company.com"),
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "collaborator_auth_policy", "password"),
				),
			},

			{
				Config: testAccTFEOrganization_update(rInt1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOrganizationExists(
						"tfe_organization.foobar", org),
					testAccCheckTFEOrganizationAttributesUpdated(org, fmt.Sprintf("tst-terraform-%d", rInt1)),
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
				),
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

func testAccCheckTFEOrganizationAttributes(
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

func testAccCheckTFEOrganizationAttributesUpdated(
	org *tfe.Organization, expectedOrgName string) resource.TestCheckFunc {
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

func testAccTFEOrganization_update(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name                     = "tst-terraform-%d"
  email                    = "admin-updated@company.com"
  session_timeout_minutes  = 3600
  session_remember_minutes = 3600
  owners_team_saml_role_id = "owners"
}`, rInt)
}
