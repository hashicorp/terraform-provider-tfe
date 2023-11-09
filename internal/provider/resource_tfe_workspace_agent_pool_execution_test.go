// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"errors"
	"fmt"
	"log"
	"testing"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccTFEWorkspaceExecutionMode_create_update(t *testing.T) {
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
				Config: testAccTFEWorkspaceExecutionMode_basic(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccTFECheckWorkspaceAgentPoolAttached("tfe_workspace.workspace", "tfe_agent_pool.pool"),
					resource.TestCheckResourceAttr("tfe_agent_pool.pool", "organization_scoped", "false"),
				),
			},
			{
				Config: testAccTFEWorkspaceExecutionMode_update(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccTFECheckWorkspaceAgentPoolAttached("tfe_workspace.workspace", "tfe_agent_pool.pool"),
					resource.TestCheckResourceAttr("tfe_agent_pool.pool", "organization_scoped", "false"),
				),
			},
			{
				Config: testAccTFEWorkspaceExecutionMode_destroy(org.Name),
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

		// Read state file for workspace
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

		// Read state file for agent pool
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

		ap, ok := s.RootModule().Resources[pool]
		if !ok {
			return fmt.Errorf("Resource not found: %s", pool)
		}

		// Resource ID equals the Workspace ID
		if ws.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		// Resource ID equals the Agent Pool ID
		if ap.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		workspace, err := config.Client.Workspaces.ReadByID(ctx, ws.Primary.ID)
		if err != nil && !errors.Is(err, tfe.ErrResourceNotFound) {
			return fmt.Errorf("error fetching workspace: %w", err)
		}

		if workspace.AgentPoolID != "" {
			return fmt.Errorf("error detaching agent pool %s: %w", ap.Primary.ID, err)
		}

		pool, err := config.Client.AgentPools.Read(ctx, ap.Primary.ID)
		if err != nil && !errors.Is(err, tfe.ErrResourceNotFound) {
			log.Printf("[DEBUB] THERE IS SOMETHING WRONG WITH THIS AGENT POOL %s", pool.ID)
			return fmt.Errorf("error fetching agent pool: %w", err)
		}

		return nil
	}
}

func testAccTFEWorkspaceExecutionMode_basic(organization string) string {
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
resource "tfe_workspace_execution_mode" "attach"{
	workspace_id = tfe_workspace.workspace.id
	agent_pool_id = tfe_agent_pool_allowed_workspaces.permit.agent_pool_id
	execution_mode = "agent"
	depends_on = [tfe_agent_pool_allowed_workspaces.permit]
}`, organization, organization)
}

func testAccTFEWorkspaceExecutionMode_update(organization string) string {
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

resource "tfe_workspace_execution_mode" "attach"{
	workspace_id = tfe_workspace.workspace.id
	agent_pool_id = tfe_agent_pool_allowed_workspaces.permit.agent_pool_id
	execution_mode = "agent"
	depends_on = [tfe_agent_pool_allowed_workspaces.permit]
}`, organization, organization)
}

func testAccTFEWorkspaceExecutionMode_destroy(organization string) string {
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

	resource "tfe_workspace_execution_mode" "attach"{
		workspace_id = tfe_workspace.workspace.id
		agent_pool_id = tfe_agent_pool_allowed_workspaces.permit.agent_pool_id
		execution_mode = "agent"
		depends_on = [tfe_agent_pool_allowed_workspaces.permit]
	}`, organization, organization)
}
