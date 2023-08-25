// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccTFEOrganizationMembershipDataSource_basic(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrganizationMembershipDataSourceConfig(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.tfe_organization_membership.foobar", "email", "example@hashicorp.com"),
					resource.TestCheckResourceAttr(
						"data.tfe_organization_membership.foobar", "username", ""),
					resource.TestCheckResourceAttr(
						"data.tfe_organization_membership.foobar", "organization", orgName),
					resource.TestCheckResourceAttrSet("data.tfe_organization_membership.foobar", "user_id"),
				),
			},
			{
				Config: testAccTFEOrganizationMembershipDataSourceConfigWithOrgMemberID(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.tfe_organization_membership.foobar", "email", "example@hashicorp.com"),
					resource.TestCheckResourceAttr(
						"data.tfe_organization_membership.foobar", "username", ""),
					resource.TestCheckResourceAttr(
						"data.tfe_organization_membership.foobar", "organization", orgName),
					resource.TestCheckResourceAttrSet("data.tfe_organization_membership.foobar", "user_id"),
				),
			},
		},
	})
}

func TestAccTFEOrganizationMembershipDataSource_findByName(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("CI Seeded users are not in the hashicorp org in CI")
	}

	// This test requires a user that exists in a TFC organization called "hashicorp".
	// Our CI instance has a default organization "hashicorp" and prepopulates it
	// with users (i.e TFE_USER1, etc) since we are unable to create users via the API.
	// In order to run this against your own organization, simply modify the organization
	// attribute in the test step config and set TFE_USER1 to the desired user you want to fetch.
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			if envTFEUser1 == "" {
				t.Skip("Please set TFE_USER1 to run this test")
			}
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrganizationMembershipDataSourceSearchUsername(envTFEUser1),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"data.tfe_organization_membership.foobar", "email"),
					resource.TestCheckResourceAttr(
						"data.tfe_organization_membership.foobar", "username", envTFEUser1),
					resource.TestCheckResourceAttr(
						"data.tfe_organization_membership.foobar", "organization", "hashicorp"),
					resource.TestCheckResourceAttrSet("data.tfe_organization_membership.foobar", "user_id"),
				),
			},
		},
	})
}

func TestAccTFEOrganizationMembershipDataSource_missingParams(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFEOrganizationMembershipDataSourceMissingParams(rInt),
				ExpectError: regexp.MustCompile("`email,organization_membership_id,username` must be specified"),
			},
		},
	})
}

func testAccTFEOrganizationMembershipDataSourceConfig(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_organization_membership" "foobar" {
  email        = "example@hashicorp.com"
  organization = tfe_organization.foobar.id
}

data "tfe_organization_membership" "foobar" {
  email        = tfe_organization_membership.foobar.email
  organization = tfe_organization.foobar.name
}`, rInt)
}

func testAccTFEOrganizationMembershipDataSourceSearchUsername(username string) string {
	return fmt.Sprintf(`
data "tfe_organization_membership" "foobar" {
  username     = "%s"
  organization = "hashicorp"
}`, username)
}

func testAccTFEOrganizationMembershipDataSourceMissingParams(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_organization_membership" "foobar" {
  email        = "example@hashicorp.com"
  organization = tfe_organization.foobar.id
}

data "tfe_organization_membership" "foobar" {
  organization = tfe_organization.foobar.name
}`, rInt)
}

func testAccTFEOrganizationMembershipDataSourceConfigWithOrgMemberID(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_organization_membership" "foobar" {
  email        = "example@hashicorp.com"
  organization = tfe_organization.foobar.id
}

data "tfe_organization_membership" "foobar" {
  organization 								= tfe_organization.foobar.name
  organization_membership_id  = tfe_organization_membership.foobar.id
}`, rInt)
}
