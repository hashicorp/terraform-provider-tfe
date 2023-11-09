// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccTFEOrganizationModuleSharing_basic(t *testing.T) {
	skipIfCloud(t)

	rInt1 := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	rInt2 := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	rInt3 := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	// Destroying a module sharing relationship is effectively updating
	// the module sharing resource consumers to be an empty array
	// We've omitted CheckDestroy since verifying the module sharing
	// has been deleted requires the organizations to exist (and they are destroyed
	// prior to CheckDestroy being executed)
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrganizationModuleSharing_basic(rInt1, rInt2, rInt3),
				Check: resource.ComposeAggregateTestCheckFunc(
					// organization attribute */
					resource.TestCheckResourceAttr(
						"tfe_organization_module_sharing.producer", "organization", fmt.Sprintf("tst-terraform-%d", rInt1)),

					// module_consumers attribute
					resource.TestCheckResourceAttr(
						"tfe_organization_module_sharing.producer", "module_consumers.#", "2"),
					resource.TestCheckResourceAttrSet(
						"tfe_organization_module_sharing.producer", "module_consumers.0"),
					resource.TestCheckResourceAttrSet(
						"tfe_organization_module_sharing.producer", "module_consumers.1"),
				),
			},
		},
	})
}

func TestAccTFEOrganizationModuleSharing_emptyOrg(t *testing.T) {
	skipIfCloud(t)

	rInt1 := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	rInt2 := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrganizationModuleSharing_emptyOrg(rInt1, rInt2),
				Check: resource.ComposeAggregateTestCheckFunc(
					// organization attribute */
					resource.TestCheckResourceAttr(
						"tfe_organization_module_sharing.foo", "organization", fmt.Sprintf("tst-terraform-%d", rInt1)),

					// module_consumers attribute
					// even though we've provided an empty string,
					// we'll have two entries here since they are ignored
					resource.TestCheckResourceAttr(
						"tfe_organization_module_sharing.foo", "module_consumers.#", "2"),
					resource.TestCheckResourceAttr(
						"tfe_organization_module_sharing.foo", "module_consumers.0", ""),
					resource.TestCheckResourceAttr(
						"tfe_organization_module_sharing.foo", "module_consumers.1", fmt.Sprintf("tst-terraform-%d", rInt2)),
				),
			},
		},
	})
}

func TestAccTFEOrganizationModuleSharing_stopSharing(t *testing.T) {
	skipIfCloud(t)

	rInt1 := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	// This test will serve as a proxy for CheckDestroy, since
	// setting a module_consumers to an empty array of
	// "destroys" the resource
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrganizationModuleSharing_stopSharing(rInt1),
				Check: resource.ComposeAggregateTestCheckFunc(
					// organization attribute */
					resource.TestCheckResourceAttr(
						"tfe_organization_module_sharing.foo", "organization", fmt.Sprintf("tst-terraform-%d", rInt1)),

					// module_consumers attribute
					resource.TestCheckResourceAttr(
						"tfe_organization_module_sharing.foo", "module_consumers.#", "0"),
				),
			},
		},
	})
}

func testAccTFEOrganizationModuleSharing_basic(rInt1 int, rInt2 int, rInt3 int) string {
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

resource "tfe_organization_module_sharing" "producer" {
  organization = tfe_organization.foobar.id
  module_consumers = [tfe_organization.foo.id, tfe_organization.bar.id]
}`, rInt1, rInt2, rInt3)
}

func testAccTFEOrganizationModuleSharing_emptyOrg(rInt1 int, rInt2 int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foo" {
  name = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_organization" "bar" {
  name = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_organization_module_sharing" "foo" {
  organization = tfe_organization.foo.id
  module_consumers = ["", tfe_organization.bar.id]
}`, rInt1, rInt2)
}

func testAccTFEOrganizationModuleSharing_stopSharing(rInt1 int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foo" {
  name = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_organization_module_sharing" "foo" {
  organization = tfe_organization.foo.id
  module_consumers = []
}`, rInt1)
}
