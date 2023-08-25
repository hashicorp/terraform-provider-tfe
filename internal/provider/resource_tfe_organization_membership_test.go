// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"math/rand"
	"regexp"
	"testing"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccTFEOrganizationMembership_basic(t *testing.T) {
	mem := &tfe.OrganizationMembership{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEOrganizationMembershipDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrganizationMembership_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOrganizationMembershipExists(
						"tfe_organization_membership.foobar", mem),
					testAccCheckTFEOrganizationMembershipAttributes(mem, orgName),
					resource.TestCheckResourceAttr(
						"tfe_organization_membership.foobar", "email", "example@hashicorp.com"),
					resource.TestCheckResourceAttr(
						"tfe_organization_membership.foobar", "organization", orgName),
					resource.TestCheckResourceAttrSet("tfe_organization_membership.foobar", "user_id"),
					resource.TestCheckResourceAttr(
						"tfe_organization_membership.foobar", "username", ""),
				),
			},
		},
	})
}

func TestAccTFEOrganizationMembershipImport_ByID(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEOrganizationMembershipDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrganizationMembership_basic(rInt),
			},
			{
				ResourceName:      "tfe_organization_membership.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTFEOrganizationMembershipImport_ByEmail(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	orgName := fmt.Sprintf("tst-terraform-%d", rInt)
	email := "testuser@hashicorp.com"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEOrganizationMembershipDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrganizationMembership_nameAndEmail(orgName, email),
			},
			{
				ResourceName:      "tfe_organization_membership.foobar",
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s/%s", orgName, email),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTFEOrganizationMembershipImport_invalidImportId(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	orgName := fmt.Sprintf("tst-terraform-%d", rInt)
	email := "testuser@hashicorp.com"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEOrganizationMembershipDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrganizationMembership_nameAndEmail(orgName, email),
			},
			{
				ResourceName:  "tfe_organization_membership.foobar",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s/%s/someOtherString", orgName, email),
				ExpectError:   regexp.MustCompile(fmt.Sprintf("error retrieving user with email %s/someOtherString from organization %s", email, orgName)),
			},
			{
				ResourceName:  "tfe_organization_membership.foobar",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("invalid-org-%d/%s", rInt, email),
				ExpectError:   regexp.MustCompile(fmt.Sprintf("error retrieving user with email %s from organization invalid-org-%d", email, rInt)),
			},
			{
				ResourceName:  "tfe_organization_membership.foobar",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s/invalidEmail", orgName),
				ExpectError:   regexp.MustCompile(fmt.Sprintf("error retrieving user with email invalidEmail from organization %s", orgName)),
			},
		},
	})
}

func testAccCheckTFEOrganizationMembershipExists(
	n string, membership *tfe.OrganizationMembership) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(ConfiguredClient)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		options := tfe.OrganizationMembershipReadOptions{
			Include: []tfe.OrgMembershipIncludeOpt{tfe.OrgMembershipUser},
		}

		m, err := config.Client.OrganizationMemberships.ReadWithOptions(ctx, rs.Primary.ID, options)
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
	membership *tfe.OrganizationMembership, expectedOrgName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if membership.Organization.Name != expectedOrgName {
			return fmt.Errorf("Bad organization: %s", membership.Organization.Name)
		}
		if membership.User.Email != "example@hashicorp.com" {
			return fmt.Errorf("Bad email: %s", membership.User.Email)
		}
		if membership.User.ID == "" {
			return fmt.Errorf("Bad user ID: %s", membership.User.ID)
		}

		return nil
	}
}

func testAccCheckTFEOrganizationMembershipDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(ConfiguredClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_organization_membership" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := config.Client.OrganizationMemberships.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Membership %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccTFEOrganizationMembership_basic(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_organization_membership" "foobar" {
  email        = "example@hashicorp.com"
  organization = tfe_organization.foobar.id
}`, rInt)
}

func testAccTFEOrganizationMembership_nameAndEmail(orgName string, email string) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "%s"
  email = "admin@company.com"
}

resource "tfe_organization_membership" "foobar" {
  email        = "%s"
  organization = tfe_organization.foobar.id
}`, orgName, email)
}
