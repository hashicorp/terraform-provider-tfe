// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccTFEOrganizationsDataSource_basic(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrganizationsDataSourceConfig_basic_resource(rInt),
			},
			{
				Config: testAccTFEOrganizationsDataSourceConfig_basic_data(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTFEOrganizationHasNames("data.tfe_organizations.foobarbaz", []string{
						fmt.Sprintf("tst-terraform-foo-%d", rInt),
						fmt.Sprintf("tst-terraform-bar-%d", rInt),
						fmt.Sprintf("tst-terraform-baz-%d", rInt),
					}),
					testAccCheckTFEOrganizationHasIDs("data.tfe_organizations.foobarbaz", []string{
						fmt.Sprintf("tst-terraform-foo-%d", rInt),
						fmt.Sprintf("tst-terraform-bar-%d", rInt),
						fmt.Sprintf("tst-terraform-baz-%d", rInt),
					}),
				),
			},
		},
	})
}

func testAccCheckTFEOrganizationHasNames(dataOrg string, orgNames []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		org, ok := s.RootModule().Resources[dataOrg]
		if !ok {
			return fmt.Errorf("Data organization '%s' not found.", dataOrg)
		}
		numOrgsStr := org.Primary.Attributes["names.#"]
		numOrgs, _ := strconv.Atoi(numOrgsStr)

		if numOrgs < len(orgNames) {
			return fmt.Errorf("expected at least %d organizations, but found %d.", len(orgNames), numOrgs)
		}

		allOrgsMap := map[string]struct{}{}
		for i := 0; i < numOrgs; i++ {
			orgName := org.Primary.Attributes[fmt.Sprintf("names.%d", i)]
			allOrgsMap[orgName] = struct{}{}
		}

		for _, orgName := range orgNames {
			_, ok := allOrgsMap[orgName]
			if !ok {
				return fmt.Errorf("expected to find organization name %s, but did not.", orgName)
			}
		}

		return nil
	}
}

func testAccCheckTFEOrganizationHasIDs(dataOrg string, orgNames []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		org, ok := s.RootModule().Resources[dataOrg]
		if !ok {
			return fmt.Errorf("Organization not found: %s.", dataOrg)
		}

		for _, orgName := range orgNames {
			id := fmt.Sprintf("ids.%s", orgName)
			_, ok := org.Primary.Attributes[id]
			if !ok {
				return fmt.Errorf("expected to find organization id %s, but did not.", id)
			}
		}

		return nil
	}
}

func testAccTFEOrganizationsDataSourceConfig_basic_resource(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foo" {
  name  = "tst-terraform-foo-%d"
  email = "admin@company.com"
}

resource "tfe_organization" "bar" {
  name  = "tst-terraform-bar-%d"
  email = "admin@company.com"
}

resource "tfe_organization" "baz" {
  name  = "tst-terraform-baz-%d"
  email = "admin@company.com"
}`, rInt, rInt, rInt)
}

func testAccTFEOrganizationsDataSourceConfig_basic_data() string {
	return `data "tfe_organizations" "foobarbaz" {}`
}
