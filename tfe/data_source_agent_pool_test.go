// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tfe

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccTFEAgentPoolDataSource_basic(t *testing.T) {
	skipIfEnterprise(t)

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEAgentPoolDataSourceConfig(org.Name, rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tfe_agent_pool.foobar", "id"),
					resource.TestCheckResourceAttr(
						"data.tfe_agent_pool.foobar", "name", fmt.Sprintf("agent-pool-test-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.tfe_agent_pool.foobar", "organization", org.Name),
				),
			},
		},
	})
}

func testAccTFEAgentPoolDataSourceConfig(organization string, rInt int) string {
	return fmt.Sprintf(`
resource "tfe_agent_pool" "foobar" {
  name                  = "agent-pool-test-%d"
  organization          = "%s"
}

data "tfe_agent_pool" "foobar" {
  name         = tfe_agent_pool.foobar.name
  organization = "%s"
}`, rInt, organization, organization)
}
