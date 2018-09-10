package tfe

import (
	"fmt"
	"testing"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccTFEOrganization_basic(t *testing.T) {
	org := &tfe.Organization{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEOrganizationDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccTFEOrganization_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOrganizationExists(
						"tfe_organization.foobar", org),
					testAccCheckTFEOrganizationAttributes(org),
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "name", "terraform-test"),
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

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEOrganizationDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccTFEOrganization_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOrganizationExists(
						"tfe_organization.foobar", org),
					testAccCheckTFEOrganizationAttributes(org),
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "name", "terraform-test"),
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "email", "admin@company.com"),
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "collaborator_auth_policy", "password"),
				),
			},

			resource.TestStep{
				Config: testAccTFEOrganization_update,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOrganizationExists(
						"tfe_organization.foobar", org),
					testAccCheckTFEOrganizationAttributesUpdated(org),
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "name", "terraform-updated"),
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "email", "admin-updated@company.com"),
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "session_timeout_minutes", "3600"),
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "session_remember_minutes", "3600"),
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "collaborator_auth_policy", "password"),
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

func testAccCheckTFEOrganizationAttributes(
	org *tfe.Organization) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if org.Name != "terraform-test" {
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
	org *tfe.Organization) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if org.Name != "terraform-updated" {
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

const testAccTFEOrganization_basic = `
resource "tfe_organization" "foobar" {
  name = "terraform-test"
  email = "admin@company.com"
}`

const testAccTFEOrganization_update = `
resource "tfe_organization" "foobar" {
  name = "terraform-updated"
  email = "admin-updated@company.com"
  session_timeout_minutes = 3600
  session_remember_minutes = 3600
}`
