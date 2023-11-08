// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"errors"
	"fmt"
	"testing"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccTFEWorkspaceAgentPoolExecution_create_update(t *testing.T) {
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
				Config: testAccTFEWorkspaceAgentPoolExecution_basic(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccTFECheckWorkspaceAgentPoolAttached("tfe_workspace.workspace", "tfe_agent_pool.pool"),
					resource.TestCheckResourceAttr("tfe_agent_pool.pool", "organization_scoped", "false"),
				),
			},
			{
				Config: testAccTFEWorkspaceAgentPoolExecution_update(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccTFECheckWorkspaceAgentPoolAttached("tfe_workspace.workspace", "tfe_agent_pool.pool"),
					resource.TestCheckResourceAttr("tfe_agent_pool.pool", "organization_scoped", "false"),
				),
			},
			{
				Config: testAccTFEWorkspaceAgentPoolExecution_destroy(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccTFECheckWorkspaceAgentPoolNotDetached("tfe_workspace.workspace", "tfe_agent_pool.pool"),
				),
			},
		},
	})
}

func testAccTFECheckWorkspaceAgentPoolAttached(workspace string, pool string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(ConfiguredClient)

		ws, ok := s.RootModule().Resources[workspace]
		if !ok {
			return fmt.Errorf("Resource not found: %s", workspace)
		}

		// Resource ID equals the Workspace ID
		if ws.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		workspace, err := config.Client.Workspaces.ReadByID(ctx, ws.Primary.ID)
		if err != nil && !errors.Is(err, tfe.ErrResourceNotFound) {
			return fmt.Errorf("error fetching workspace: %w", err)
		}

		ap, ok := s.RootModule().Resources[pool]
		if !ok {
			return fmt.Errorf("Resource not found: %s", pool)
		}

		// Resource ID equals the Agent Pool ID
		if ap.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		if workspace.AgentPoolID != ap.Primary.ID {
			return fmt.Errorf("error attaching agent pool %s: %w", ap.Primary.ID, err)
		}

		return nil
	}
}

func testAccTFECheckWorkspaceAgentPoolNotDetached(workspace string, pool string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(ConfiguredClient)

		ws, ok := s.RootModule().Resources[workspace]
		if !ok {
			return fmt.Errorf("Resource not found: %s", workspace)
		}

		// Resource ID equals the Workspace ID
		if ws.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		workspace, err := config.Client.Workspaces.ReadByID(ctx, ws.Primary.ID)
		if err != nil && !errors.Is(err, tfe.ErrResourceNotFound) {
			return fmt.Errorf("error fetching workspace: %w", err)
		}

		ap, ok := s.RootModule().Resources[pool]
		if !ok {
			return fmt.Errorf("Resource not found: %s", pool)
		}

		// Resource ID equals the Agent Pool ID
		if ap.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		if workspace.AgentPoolID != "" {
			return fmt.Errorf("error detaching agent pool %s: %w", ap.Primary.ID, err)
		}

		return nil
	}
}

func testAccTFEWorkspaceAgentPoolExecution_basic(organization string) string {
	return fmt.Sprintf(`
resource "tfe_workspace" "workspace" {
  name = "test-workspace"
  organization = "%s"
}

resource "tfe_agent_pool" "pool" {
  name         = "agent-pool-updated"
  organization = "%s"
  organization_scoped = false
}

resource "tfe_agent_pool_allowed_workspaces" "permit"{
  agent_pool_id 		= tfe_agent_pool.pool.id
  allowed_workspace_ids = [
		tfe_workspace.workspace.id
		]
}
resource "tfe_workspace_agent_pool_execution" "attach"{
	workspace_id = tfe_workspace.workspace.id
	agent_pool_id = tfe_agent_pool.pool.id
	depends_on = [tfe_agent_pool_allowed_workspaces.permit]
}`, organization, organization)
}

func testAccTFEWorkspaceAgentPoolExecution_update(organization string) string {
	return fmt.Sprintf(`
resource "tfe_workspace" "workspace" {
  name = "test-workspace"
  organization = "%s"
}

resource "tfe_agent_pool" "pool" {
  name         = "agent-pool-updated"
  organization = "%s"
  organization_scoped = false
}

resource "tfe_agent_pool_allowed_workspaces" "permit"{
  agent_pool_id 		= tfe_agent_pool.pool.id
  allowed_workspace_ids = [
		tfe_workspace.workspace.id
		]
}

resource "tfe_workspace_agent_pool_execution" "attach"{
	workspace_id = tfe_workspace.workspace.id
	agent_pool_id = tfe_agent_pool.pool.id
	depends_on = [tfe_agent_pool_allowed_workspaces.permit]
}`, organization, organization)
}

func testAccTFEWorkspaceAgentPoolExecution_destroy(organization string) string {
	return fmt.Sprintf(`
	resource "tfe_workspace" "workspace" {
		name = "test-workspace"
		organization = "%s"
	}

	resource "tfe_agent_pool" "pool" {
		name         = "agent-pool-updated"
		organization = "%s"
		organization_scoped = false
	}

	resource "tfe_agent_pool_allowed_workspaces" "permit"{
		agent_pool_id 		= tfe_agent_pool.pool.id
		allowed_workspace_ids = [
			tfe_workspace.workspace.id
			]
	}

	resource "tfe_workspace_agent_pool_execution" "attach"{
		workspace_id = tfe_workspace.workspace.id
		agent_pool_id = tfe_agent_pool.pool.id
		depends_on = [tfe_agent_pool_allowed_workspaces.permit]
	}`, organization, organization)
}
