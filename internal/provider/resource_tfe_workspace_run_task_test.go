// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccTFEWorkspaceRunTask_create(t *testing.T) {
	skipUnlessRunTasksDefined(t)

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	workspaceTask := &tfe.WorkspaceRunTask{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEWorkspaceRunTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEWorkspaceRunTask_basic(org.Name, runTasksURL()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceRunTaskExists("tfe_workspace_run_task.foobar", workspaceTask),
					resource.TestCheckResourceAttr("tfe_workspace_run_task.foobar", "enforcement_level", "advisory"),
				),
			},
			{
				Config: testAccTFEWorkspaceRunTask_update(org.Name, runTasksURL()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tfe_workspace_run_task.foobar", "enforcement_level", "mandatory"),
				),
			},
		},
	})
}

func TestAccTFEWorkspaceRunTask_beta_create(t *testing.T) {
	skipUnlessRunTasksDefined(t)
	skipUnlessBeta(t)

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	workspaceTask := &tfe.WorkspaceRunTask{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEWorkspaceRunTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEWorkspaceRunTask_beta_basic(org.Name, runTasksURL()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceRunTaskExists("tfe_workspace_run_task.foobar", workspaceTask),
					resource.TestCheckResourceAttr("tfe_workspace_run_task.foobar", "enforcement_level", "advisory"),
					resource.TestCheckResourceAttr("tfe_workspace_run_task.foobar", "stage", "post_plan"),
				),
			},
			{
				Config: testAccTFEWorkspaceRunTask_beta_update(org.Name, runTasksURL()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tfe_workspace_run_task.foobar", "enforcement_level", "mandatory"),
					resource.TestCheckResourceAttr("tfe_workspace_run_task.foobar", "stage", "pre_plan"),
				),
			},
		},
	})
}

func TestAccTFEWorkspaceRunTask_import(t *testing.T) {
	skipUnlessRunTasksDefined(t)

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFETeamAccessDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEWorkspaceRunTask_basic(org.Name, runTasksURL()),
			},
			{
				ResourceName:      "tfe_workspace_run_task.foobar",
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s/workspace-test/foobar-task", org.Name),
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckTFEWorkspaceRunTaskExists(n string, runTask *tfe.WorkspaceRunTask) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(ConfiguredClient)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		if rs.Primary.Attributes["workspace_id"] == "" {
			return fmt.Errorf("No Workspace ID is set")
		}

		rt, err := config.Client.WorkspaceRunTasks.Read(ctx, rs.Primary.Attributes["workspace_id"], rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error reading Workspace Run Task: %w", err)
		}

		if rt == nil {
			return fmt.Errorf("Workspace Run Task not found")
		}

		*runTask = *rt

		return nil
	}
}

func testAccCheckTFEWorkspaceRunTaskDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(ConfiguredClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_workspace_run_task" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}
		if rs.Primary.Attributes["workspace_id"] == "" {
			return fmt.Errorf("No Workspace ID is set")
		}

		_, err := config.Client.WorkspaceRunTasks.Read(ctx, rs.Primary.Attributes["workspace_id"], rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Workspace Run Tasks %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccTFEWorkspaceRunTask_basic(orgName, runTaskURL string) string {
	return fmt.Sprintf(`
locals {
    organization_name = "%s"
}
resource "tfe_organization_run_task" "foobar" {
  organization = local.organization_name
  url          = "%s"
  name         = "foobar-task"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = local.organization_name
}

resource "tfe_workspace_run_task" "foobar" {
  workspace_id      = resource.tfe_workspace.foobar.id
  task_id           = resource.tfe_organization_run_task.foobar.id
  enforcement_level = "advisory"
}
`, orgName, runTaskURL)
}

func testAccTFEWorkspaceRunTask_update(orgName, runTaskURL string) string {
	return fmt.Sprintf(`
locals {
    organization_name = "%s"
}

resource "tfe_organization_run_task" "foobar" {
  organization = local.organization_name
  url          = "%s"
  name         = "foobar-task"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = local.organization_name
}

resource "tfe_workspace_run_task" "foobar" {
  workspace_id      = resource.tfe_workspace.foobar.id
  task_id           = resource.tfe_organization_run_task.foobar.id
  enforcement_level = "mandatory"
}
`, orgName, runTaskURL)
}

func testAccTFEWorkspaceRunTask_beta_basic(orgName, runTaskURL string) string {
	return fmt.Sprintf(`
locals {
    organization_name = "%s"
}

resource "tfe_organization_run_task" "foobar" {
  organization = local.organization_name
  url          = "%s"
  name         = "foobar-task"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_workspace_run_task" "foobar" {
  workspace_id      = resource.tfe_workspace.foobar.id
  task_id           = resource.tfe_organization_run_task.foobar.id
  enforcement_level = "advisory"
  stage             = "post_plan"
}
`, orgName, runTaskURL)
}

func testAccTFEWorkspaceRunTask_beta_update(orgName, runTaskURL string) string {
	return fmt.Sprintf(`
locals {
    organization_name = "%s"
}

resource "tfe_organization_run_task" "foobar" {
  organization = local.organization_name
  url          = "%s"
  name         = "foobar-task"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = local.organization_name
}

resource "tfe_workspace_run_task" "foobar" {
  workspace_id      = resource.tfe_workspace.foobar.id
  task_id           = resource.tfe_organization_run_task.foobar.id
  enforcement_level = "mandatory"
  stage             = "pre_plan"
}
`, orgName, runTaskURL)
}
