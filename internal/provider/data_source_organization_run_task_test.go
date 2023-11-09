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

func TestAccTFEOrganizationRunTaskDataSource_basic(t *testing.T) {
	skipUnlessRunTasksDefined(t)

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
				Config: testAccTFEOrganizationRunTaskDataSourceConfig(org.Name, rInt, runTasksURL()),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.tfe_organization_run_task.foobar", "name", fmt.Sprintf("foobar-task-%d", rInt)),
					resource.TestCheckResourceAttr("data.tfe_organization_run_task.foobar", "url", runTasksURL()),
					resource.TestCheckResourceAttr("data.tfe_organization_run_task.foobar", "category", "task"),
					resource.TestCheckResourceAttr("data.tfe_organization_run_task.foobar", "enabled", "false"),
					resource.TestCheckResourceAttr("data.tfe_organization_run_task.foobar", "description", "a description"),
					resource.TestCheckResourceAttrSet("data.tfe_organization_run_task.foobar", "id"),
					resource.TestCheckResourceAttrSet("data.tfe_organization_run_task.foobar", "organization"),
				),
			},
		},
	})
}

func testAccTFEOrganizationRunTaskDataSourceConfig(orgName string, rInt int, runTaskURL string) string {
	return fmt.Sprintf(`
locals {
    organization_name = "%s"
}

resource "tfe_organization_run_task" "foobar" {
	organization = local.organization_name
	url          = "%s"
	name         = "foobar-task-%d"
	hmac_key     = "Password1"
	enabled      = false
	description = "a description"
}

data "tfe_organization_run_task" "foobar" {
	organization      = local.organization_name
	name              = "foobar-task-%d"
	depends_on = [tfe_organization_run_task.foobar]
}`, orgName, runTaskURL, rInt, rInt)
}
