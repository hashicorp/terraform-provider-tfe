package tfe

import (
	"fmt"
	"testing"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccTFEOrganizationMembership_basic(t *testing.T) {
	mem := &tfe.OrganizationMembership{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEOrganizationMembershipDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrganizationMembership_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOrganizationMembershipExists(
						"tfe_organization_membership.foobar", mem),
					testAccCheckTFEOrganizationMembershipAttributes(mem),
					resource.TestCheckResourceAttr(
						"tfe_organization_membership.foobar", "email", "example@hashicorp.com"),
				),
			},
		},
	})
}

func TestAccTFEOrganizationMembership_import(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEOrganizationMembershipDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrganizationMembership_basic,
			},
			{
				ResourceName:      "tfe_organization_membership.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckTFEOrganizationMembershipExists(
	n string, membership *tfe.OrganizationMembership) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		tfeClient := testAccProvider.Meta().(*tfe.Client)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		options := tfe.OrganizationMembershipReadOptions{
			Include: "user",
		}

		m, err := tfeClient.OrganizationMemberships.ReadWithOptions(ctx, rs.Primary.ID, options)
		if err != nil {
			return err
		}

		if m == nil {
			return fmt.Errorf("Membership not found")
		}

		*membership = *m

		return nil
	}
}

func testAccCheckTFEOrganizationMembershipAttributes(
	membership *tfe.OrganizationMembership) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if membership.User.Email != "example@hashicorp.com" {
			return fmt.Errorf("Bad email: %s", membership.User.Email)
		}
		return nil
	}
}

func testAccCheckTFEOrganizationMembershipDestroy(s *terraform.State) error {
	tfeClient := testAccProvider.Meta().(*tfe.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_organization_membership" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := tfeClient.OrganizationMemberships.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Membership %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

const testAccTFEOrganizationMembership_basic = `
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-2"
  email = "admin@company.com"
}

resource "tfe_organization_membership" "foobar" {
  email        = "example@hashicorp.com"
  organization = "${tfe_organization.foobar.id}"
}`
