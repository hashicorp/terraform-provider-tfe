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
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTFEOrganizationDataSource_basic(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	org := &tfe.Organization{}
	orgName := fmt.Sprintf("tst-terraform-foo-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrganizationDataSourceConfig_basic(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTFEOrganizationExists("tfe_organization.foo", org),
					// check resource attrs
					resource.TestCheckResourceAttr("tfe_organization.foo", "name", orgName),
					resource.TestCheckResourceAttr("tfe_organization.foo", "email", "admin@company.com"),

					// check data attrs
					resource.TestCheckResourceAttr("data.tfe_organization.foo", "name", orgName),
					resource.TestCheckResourceAttr("data.tfe_organization.foo", "email", "admin@company.com"),
				),
			},
		},
	})
}

func TestAccTFEOrganizationDataSource_defaultProject(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	org := &tfe.Organization{}

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrganizationDataSourceConfig_basic(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTFEOrganizationExists("tfe_organization.foo", org),
					resource.TestCheckResourceAttrWith("tfe_organization.foo", "default_project_id", func(value string) error {
						if value != org.DefaultProject.ID {
							return fmt.Errorf("default project ID should be %s but was %s", org.DefaultProject.ID, value)
						}
						return nil
					}),
				),
			},
		},
	})
}

// The data source will use the default org name from provider config if omitted.
func TestAccTFEOrganizationDataSource_defaultOrganization(t *testing.T) {
	defaultOrgName, _ := setupDefaultOrganization(t)
	providers := providerWithDefaultOrganization(defaultOrgName)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: providers,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrganizationDataSourceConfig_noName(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// check data attrs
					resource.TestCheckResourceAttr("data.tfe_organization.foo", "name", defaultOrgName),
					resource.TestMatchResourceAttr("data.tfe_organization.foo", "email", regexp.MustCompile(`@tfe\.local\z`)),
				),
			},
		},
	})
}

func testAccTFEOrganizationDataSourceConfig_basic(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foo" {
  name  = "tst-terraform-foo-%d"
  email = "admin@company.com"
}

data "tfe_organization" "foo" {
  name  = tfe_organization.foo.name
	depends_on = [tfe_organization.foo]
}`, rInt)
}

func testAccTFEOrganizationDataSourceConfig_noName() string {
	return `
data "tfe_organization" "foo" {
}`
}
