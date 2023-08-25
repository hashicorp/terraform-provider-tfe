// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"math/rand"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccTFEAdminOrganizationSettings_basic(t *testing.T) {
	skipIfCloud(t)

	rInt1 := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	rInt2 := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	rInt3 := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testConfigTFEAdminOrganizationSettings_basic(rInt1, rInt2, rInt3),
				Check: resource.ComposeAggregateTestCheckFunc(
					// organization attribute */
					resource.TestCheckResourceAttr(
						"tfe_admin_organization_settings.settings", "organization", fmt.Sprintf("tst-terraform-%d", rInt1)),
					resource.TestCheckResourceAttr(
						"tfe_admin_organization_settings.settings", "global_module_sharing", "false"),
					resource.TestCheckResourceAttr(
						"tfe_admin_organization_settings.settings", "access_beta_tools", "true"),

					// module_consumers attribute
					resource.TestCheckResourceAttr(
						"tfe_admin_organization_settings.settings", "module_sharing_consumer_organizations.#", "2"),
					resource.TestCheckResourceAttrSet(
						"tfe_admin_organization_settings.settings", "module_sharing_consumer_organizations.0"),
					resource.TestCheckResourceAttrSet(
						"tfe_admin_organization_settings.settings", "module_sharing_consumer_organizations.1"),
				),
			},
			{
				Config:      testConfigTFEAdminOrganizationSettings_conflict(rInt1, rInt2),
				ExpectError: regexp.MustCompile(`global_module_sharing cannot be true if module_sharing_consumer_organizations are set`),
			},
			{
				PreConfig: deleteOrganization(fmt.Sprintf("tst-terraform-%d", rInt1)),
				Config:    testConfigTFEAdminOrganizationSettings_basic(rInt1, rInt2, rInt3),
				Check: resource.ComposeAggregateTestCheckFunc(
					// organization attribute */
					resource.TestCheckResourceAttr(
						"tfe_admin_organization_settings.settings", "organization", fmt.Sprintf("tst-terraform-%d", rInt1)),
					resource.TestCheckResourceAttr(
						"tfe_admin_organization_settings.settings", "global_module_sharing", "false"),
					resource.TestCheckResourceAttr(
						"tfe_admin_organization_settings.settings", "access_beta_tools", "true"),

					// module_consumers attribute
					resource.TestCheckResourceAttr(
						"tfe_admin_organization_settings.settings", "module_sharing_consumer_organizations.#", "2"),
					resource.TestCheckResourceAttrSet(
						"tfe_admin_organization_settings.settings", "module_sharing_consumer_organizations.0"),
					resource.TestCheckResourceAttrSet(
						"tfe_admin_organization_settings.settings", "module_sharing_consumer_organizations.1"),
				),
			},
		},
	})
}

func testConfigTFEAdminOrganizationSettings_basic(rInt1, rInt2, rInt3 int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
	name = "tst-terraform-%d"
	email = "admin@company.com"
}

resource "tfe_organization" "foo" {
	name = "tst-terraform-%d"
	email = "admin@company.com"
}

resource "tfe_organization" "bar" {
	name = "tst-terraform-%d"
	email = "admin@company.com"
}

resource "tfe_admin_organization_settings" "settings" {
	organization = tfe_organization.foobar.id
	global_module_sharing = false
	access_beta_tools = true

	module_sharing_consumer_organizations = [tfe_organization.foo.id, tfe_organization.bar.id]
}`, rInt1, rInt2, rInt3)
}

func testConfigTFEAdminOrganizationSettings_conflict(rInt1, rInt2 int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
	name = "tst-terraform-%d"
	email = "admin@company.com"
}

resource "tfe_organization" "foo" {
	name = "tst-terraform-%d"
	email = "admin@company.com"
}

resource "tfe_admin_organization_settings" "settings" {
	organization = tfe_organization.foobar.id
	global_module_sharing = true
	module_sharing_consumer_organizations = [tfe_organization.foo.id]
}`, rInt1, rInt2)
}

func deleteOrganization(name string) func() {
	return func() {
		client := testAccProvider.Meta().(ConfiguredClient).Client
		client.Organizations.Delete(context.Background(), name)
	}
}
