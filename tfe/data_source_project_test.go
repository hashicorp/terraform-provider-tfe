// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tfe

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccTFEProjectDataSource_basic(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEProjectDataSourceConfig(org.Name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.tfe_project.foobar", "name", "projecttest"),
					resource.TestCheckResourceAttrSet("data.tfe_project.foobar", "id"),
				),
			},
		},
	})
}

func testAccTFEProjectDataSourceConfig(organization string) string {
	return fmt.Sprintf(`
resource "tfe_project" "foobar" {
  name         = "projecttest"
  organization = "%s"
}

data "tfe_project" "foobar" {
  name         = tfe_project.foobar.name
  organization = "%s"
}`, organization, organization)
}
