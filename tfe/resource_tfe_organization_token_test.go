package tfe

import (
	"fmt"
	"math/rand"
	"regexp"
	"testing"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccTFEOrganizationToken_basic(t *testing.T) {
	token := &tfe.OrganizationToken{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEOrganizationTokenDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrganizationToken_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOrganizationTokenExists(
						"tfe_organization_token.foobar", token),
					resource.TestCheckResourceAttr(
						"tfe_organization_token.foobar", "organization", orgName),
				),
			},
		},
	})
}

func TestAccTFEOrganizationToken_existsWithoutForce(t *testing.T) {
	token := &tfe.OrganizationToken{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEOrganizationTokenDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrganizationToken_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOrganizationTokenExists(
						"tfe_organization_token.foobar", token),
					resource.TestCheckResourceAttr(
						"tfe_organization_token.foobar", "organization", orgName),
				),
			},

			{
				Config:      testAccTFEOrganizationToken_existsWithoutForce(rInt),
				ExpectError: regexp.MustCompile(`token already exists`),
			},
		},
	})
}

func TestAccTFEOrganizationToken_existsWithForce(t *testing.T) {
	token := &tfe.OrganizationToken{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEOrganizationTokenDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrganizationToken_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOrganizationTokenExists(
						"tfe_organization_token.foobar", token),
					resource.TestCheckResourceAttr(
						"tfe_organization_token.foobar", "organization", orgName),
				),
			},

			{
				Config: testAccTFEOrganizationToken_existsWithForce(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOrganizationTokenExists(
						"tfe_organization_token.regenerated", token),
					resource.TestCheckResourceAttr(
						"tfe_organization_token.regenerated", "organization", orgName),
				),
			},
		},
	})
}

func TestAccTFEOrganizationToken_import(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEOrganizationTokenDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrganizationToken_basic(rInt),
			},

			{
				ResourceName:            "tfe_organization_token.foobar",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"token"},
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

func testAccTFEOrganizationToken_basic(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_organization_token" "foobar" {
  organization = "${tfe_organization.foobar.id}"
}`, rInt)
}

func testAccTFEOrganizationToken_existsWithoutForce(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_organization_token" "foobar" {
  organization = "${tfe_organization.foobar.id}"
}

resource "tfe_organization_token" "error" {
  organization = "${tfe_organization.foobar.id}"
}`, rInt)
}

func testAccTFEOrganizationToken_existsWithForce(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_organization_token" "foobar" {
  organization = "${tfe_organization.foobar.id}"
}

resource "tfe_organization_token" "regenerated" {
  organization     = "${tfe_organization.foobar.id}"
  force_regenerate = true
}`, rInt)
}
