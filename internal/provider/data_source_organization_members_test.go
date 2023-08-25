// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccTFEOrganizationMembersDataSource_basic(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	options := tfe.OrganizationMembershipCreateOptions{
		Email: tfe.String("invited_user@company.com"),
	}
	membership := createOrganizationMembership(t, tfeClient, org.Name, options)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrganizationMembersDataSourceConfig(org.Name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.tfe_organization_members.all_members", "organization", org.Name),
					resource.TestCheckResourceAttr(
						"data.tfe_organization_members.all_members", "id", org.Name),

					resource.TestCheckResourceAttr(
						"data.tfe_organization_members.all_members", "members.#", "1"),
					resource.TestCheckResourceAttrSet(
						"data.tfe_organization_members.all_members", "members.0.organization_membership_id"),
					resource.TestCheckResourceAttrSet(
						"data.tfe_organization_members.all_members", "members.0.user_id"),

					resource.TestCheckResourceAttr(
						"data.tfe_organization_members.all_members", "members_waiting.#", "1"),
					resource.TestCheckResourceAttr(
						"data.tfe_organization_members.all_members", "members_waiting.0.organization_membership_id", membership.ID),
					resource.TestCheckResourceAttr(
						"data.tfe_organization_members.all_members", "members_waiting.0.user_id", membership.User.ID),
				),
			},
		},
	})
}

func testAccTFEOrganizationMembersDataSourceConfig(orgName string) string {
	return fmt.Sprintf(`
data "tfe_organization_members" "all_members" {
  organization = "%s"
}`, orgName)
}
