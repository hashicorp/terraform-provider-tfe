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

func TestAccTFEWorkspaceRunTaskDataSource_basic(t *testing.T) {
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
				Config: testAccTFEWorkspaceRunTaskDataSourceConfig(org.Name, rInt, runTasksURL()),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.tfe_workspace_run_task.foobar", "enforcement_level", "advisory"),
					resource.TestCheckResourceAttrSet("data.tfe_workspace_run_task.foobar", "stage"),
					resource.TestCheckResourceAttrSet("data.tfe_workspace_run_task.foobar", "id"),
					resource.TestCheckResourceAttrSet("data.tfe_workspace_run_task.foobar", "task_id"),
					resource.TestCheckResourceAttrSet("data.tfe_workspace_run_task.foobar", "workspace_id"),
				),
			},
		},
	})
}

func testAccTFEWorkspaceRunTaskDataSourceConfig(orgName string, rInt int, runTaskURL string) string {
	return fmt.Sprintf(`
locals {
    organization_name = "%s"
}

resource "tfe_organization_run_task" "foobar" {
	organization = local.organization_name
	url          = "%s"
	name         = "foobar-task-%d"
}

resource "tfe_workspace" "foobar" {
	name         = "workspace-test-%d"
	organization = local.organization_name
}

resource "tfe_workspace_run_task" "foobar" {
	workspace_id      = resource.tfe_workspace.foobar.id
	task_id           = resource.tfe_organization_run_task.foobar.id
	enforcement_level = "advisory"
}

data "tfe_workspace_run_task" "foobar" {
	workspace_id      = resource.tfe_workspace.foobar.id
	task_id           = resource.tfe_organization_run_task.foobar.id
	depends_on = [tfe_workspace_run_task.foobar]
}`, orgName, runTaskURL, rInt, rInt)
}
