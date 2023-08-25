// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"testing"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccTFETeamOrganizationMembers_create_update(t *testing.T) {
	organizationMemberships := &[]tfe.OrganizationMembership{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFETeamOrganizationMembersDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamOrganizationMembers_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETeamOrganizationMembersExists("tfe_team_organization_members.foobar", organizationMemberships),
					testAccCheckTFETeamOrganizationMembersCount(3, organizationMemberships),
					testAccCheckTFETeamOrganizationMembersAttributes(organizationMemberships),
				),
			},
			{
				Config: testAccTFETeamOrganizationMembers_deletedMembership(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETeamOrganizationMembersExists("tfe_team_organization_members.foobar", organizationMemberships),
					testAccCheckTFETeamOrganizationMembersCount(2, organizationMemberships),
				),
			},
		},
	})
}

func TestAccTFETeamOrganizationMembers_import(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFETeamOrganizationMembersDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamOrganizationMembers_basic(rInt),
			},

			{
				ResourceName:      "tfe_team_organization_members.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckTFETeamOrganizationMembersExists(resourceName string, organizationMemberships *[]tfe.OrganizationMembership) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(ConfiguredClient)
		*organizationMemberships = []tfe.OrganizationMembership{}

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		// Resource ID equals the team ID
		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		orgMemberships, err := config.Client.TeamMembers.ListOrganizationMemberships(ctx, rs.Primary.ID)
		if err != nil && !errors.Is(err, tfe.ErrResourceNotFound) {
			return fmt.Errorf("error while listing organization memberships: %w", err)
		}

		if len(orgMemberships) == 0 {
			return fmt.Errorf("Memberships for team %s do not exist", rs.Primary.ID)
		}

		for _, om := range orgMemberships {
			*organizationMemberships = append(*organizationMemberships, *om)
		}

		return nil
	}
}

func testAccCheckTFETeamOrganizationMembersAttributes(organizationMemberships *[]tfe.OrganizationMembership) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Sort slice of organization membership IDs depending on their email
		sort.SliceStable(*organizationMemberships, func(i, j int) bool {
			return (*organizationMemberships)[i].Email < (*organizationMemberships)[j].Email
		})

		if (*organizationMemberships)[0].Email != "bar@foobar.com" {
			return fmt.Errorf("Bad email: expect: bar@foobar.com, got: %q", (*organizationMemberships)[0].Email)
		}

		if (*organizationMemberships)[1].Email != "foo@foobar.com" {
			return fmt.Errorf("Bad email: expect: foo@foobar.com, got: %q", (*organizationMemberships)[1].Email)
		}

		if (*organizationMemberships)[2].Email != "leberkassemme@foobar.com" {
			return fmt.Errorf("Bad email: expect: leberkassemme@foobar.com, got: %q", (*organizationMemberships)[2].Email)
		}
		return nil
	}
}

func testAccCheckTFETeamOrganizationMembersCount(expected int, organizationMemberships *[]tfe.OrganizationMembership) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if len(*organizationMemberships) != expected {
			return fmt.Errorf("expected %d memberships, got %d", expected, len(*organizationMemberships))
		}
		return nil
	}
}

func testAccCheckTFETeamOrganizationMembersDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(ConfiguredClient)

	for _, rs := range s.RootModule().Resources {
		// Continue if current resource is not a "tfe_team_organization_members" resource
		if rs.Type != "tfe_team_organization_members" {
			continue
		}

		// The resource ID equals the team ID
		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		organizationMemberships, err := config.Client.TeamMembers.ListOrganizationMemberships(ctx, rs.Primary.ID)
		if err != nil && !errors.Is(err, tfe.ErrResourceNotFound) {
			return fmt.Errorf("error while listing organization memberships: %w", err)
		}

		// No organization memberships should exist for the team
		if len(organizationMemberships) > 0 {
			return fmt.Errorf("Organization memberships %v for team %s still exist", organizationMemberships, rs.Primary.ID)
		}
	}

	return nil
}

func testAccTFETeamOrganizationMembers_deletedMembership(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_team" "foobar" {
  name         = "team-test-%d"
  organization = tfe_organization.foobar.id
}

resource "tfe_organization_membership" "foo" {
  organization = tfe_organization.foobar.id
  email = "foo@foobar.com"
}

resource "tfe_organization_membership" "bar" {
  organization = tfe_organization.foobar.id
  email = "bar@foobar.com"
}

resource "tfe_team_organization_members" "foobar" {
  team_id  = tfe_team.foobar.id
  organization_membership_ids = [
	tfe_organization_membership.foo.id,
	tfe_organization_membership.bar.id,
  ]
}`, rInt, rInt)
}

func testAccTFETeamOrganizationMembers_basic(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_team" "foobar" {
  name         = "team-test-%d"
  organization = tfe_organization.foobar.id
}

resource "tfe_organization_membership" "foo" {
  organization = tfe_organization.foobar.id
  email = "foo@foobar.com"
}

resource "tfe_organization_membership" "bar" {
  organization = tfe_organization.foobar.id
  email = "bar@foobar.com"
}

resource "tfe_organization_membership" "leberkassemme" {
  organization = tfe_organization.foobar.id
  email = "leberkassemme@foobar.com"
}

resource "tfe_team_organization_members" "foobar" {
  team_id  = tfe_team.foobar.id
  organization_membership_ids = [
	tfe_organization_membership.foo.id,
	tfe_organization_membership.bar.id,
	tfe_organization_membership.leberkassemme.id,
  ]
}`, rInt, rInt)
}
