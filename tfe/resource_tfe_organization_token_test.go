package tfe

import (
	"fmt"
	"testing"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccTFEOrganizationToken_basic(t *testing.T) {
	token := &tfe.OrganizationToken{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEOrganizationTokenDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccTFEOrganizationToken_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOrganizationTokenExists(
						"tfe_organization_token.foobar", token),
					resource.TestCheckResourceAttr(
						"tfe_organization_token.foobar", "organization", "terraform-test"),
				),
			},
		},
	})
}

func testAccCheckTFEOrganizationTokenExists(
	n string, token *tfe.OrganizationToken) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		tfeClient := testAccProvider.Meta().(*tfe.Client)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		ot, err := tfeClient.OrganizationTokens.Read(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}

		if ot == nil {
			return fmt.Errorf("OrganizationToken not found")
		}

		*token = *ot

		return nil
	}
}

func testAccCheckTFEOrganizationTokenDestroy(s *terraform.State) error {
	tfeClient := testAccProvider.Meta().(*tfe.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_organization_token" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := tfeClient.OrganizationTokens.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("OrganizationToken %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

const testAccTFEOrganizationToken_basic = `
resource "tfe_organization" "foobar" {
  name = "terraform-test"
	email = "admin@company.com"
}

resource "tfe_organization_token" "foobar" {
  organization = "${tfe_organization.foobar.id}"
}`
