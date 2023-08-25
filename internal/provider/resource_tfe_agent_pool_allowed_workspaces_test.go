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

func TestAccTFEAgentPoolAllowedWorkspaces_create_update(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	allowedWorkspaceIDs := &[]string{}

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEAgentPoolAllowedWorkspaces_basic(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEAgentPoolAllowedWorkspacesExists("tfe_agent_pool.foobar", allowedWorkspaceIDs),
					testAccCheckTFEAgentPoolAllowedWorkspacesCount(2, allowedWorkspaceIDs),
				),
			},
			{
				Config: testAccTFEAgentPoolAllowedWorkspaces_update(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEAgentPoolAllowedWorkspacesExists("tfe_agent_pool.foobar", allowedWorkspaceIDs),
					testAccCheckTFEAgentPoolAllowedWorkspacesCount(1, allowedWorkspaceIDs),
				),
			},
			{
				Config: testAccTFEAgentPoolAllowedWorkspaces_destroy(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEAgentPoolAllowedWorkspacesNotExists("tfe_agent_pool.foobar"),
				),
			},
		},
	})
}

func testAccCheckTFEAgentPoolAllowedWorkspacesExists(resourceName string, allowedWorkspaces *[]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(ConfiguredClient)
		*allowedWorkspaces = []string{}

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		// Resource ID equals the Agent Pool ID
		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		agentPool, err := config.Client.AgentPools.Read(ctx, rs.Primary.ID)
		if err != nil && !errors.Is(err, tfe.ErrResourceNotFound) {
			return fmt.Errorf("error while fetching agent pool: %w", err)
		}

		if len(agentPool.AllowedWorkspaces) == 0 {
			return fmt.Errorf("Allowed Workspaces for agent pool %s do not exist", rs.Primary.ID)
		}

		for _, workspace := range agentPool.AllowedWorkspaces {
			*allowedWorkspaces = append(*allowedWorkspaces, workspace.ID)
		}

		return nil
	}
}

func testAccCheckTFEAgentPoolAllowedWorkspacesNotExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(ConfiguredClient)

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		// Resource ID equals the Agent Pool ID
		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		agentPool, err := config.Client.AgentPools.Read(ctx, rs.Primary.ID)
		if err != nil && !errors.Is(err, tfe.ErrResourceNotFound) {
			return fmt.Errorf("error while fetching agent pool: %w", err)
		}

		if len(agentPool.AllowedWorkspaces) > 0 {
			return fmt.Errorf("Allowed Workspaces for agent pool %s exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckTFEAgentPoolAllowedWorkspacesCount(expected int, allowedWorkspaces *[]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if len(*allowedWorkspaces) != expected {
			return fmt.Errorf("expected %d allowed workspaces, got %d", expected, len(*allowedWorkspaces))
		}
		return nil
	}
}

func TestAccTFEAgentPoolAllowedWorkspaces_import(t *testing.T) {
	skipIfEnterprise(t)

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
				Config: testAccTFEAgentPoolAllowedWorkspaces_basic(org.Name),
			},
			{
				ResourceName:      "tfe_agent_pool_allowed_workspaces.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccTFEAgentPoolAllowedWorkspaces_destroy(organization string) string {
	return fmt.Sprintf(`
resource "tfe_workspace" "foobar" {
  name = "foobar"
  organization = "%s"
}

resource "tfe_workspace" "test-workspace" {
  name = "test-workspace"
  organization = "%s"
}

resource "tfe_agent_pool" "foobar" {
  name         = "agent-pool-updated"
  organization = "%s"
  organization_scoped = false
}`, organization, organization, organization)
}

func testAccTFEAgentPoolAllowedWorkspaces_update(organization string) string {
	return fmt.Sprintf(`
resource "tfe_workspace" "foobar" {
  name = "foobar"
  organization = "%s"
}

resource "tfe_workspace" "test-workspace" {
  name = "test-workspace"
  organization = "%s"
}

resource "tfe_agent_pool" "foobar" {
  name         = "agent-pool-updated"
  organization = "%s"
  organization_scoped = false
}

resource "tfe_agent_pool_allowed_workspaces" "foobar"{
  agent_pool_id 		= tfe_agent_pool.foobar.id
  allowed_workspace_ids = [tfe_workspace.foobar.id]
}`, organization, organization, organization)
}

func testAccTFEAgentPoolAllowedWorkspaces_basic(organization string) string {
	return fmt.Sprintf(`
resource "tfe_workspace" "foobar" {
  name = "foobar"
  organization = "%s"
}

resource "tfe_workspace" "test-workspace" {
  name = "test-workspace"
  organization = "%s"
}

resource "tfe_agent_pool" "foobar" {
  name         = "agent-pool-updated"
  organization = "%s"
  organization_scoped = false
}

resource "tfe_agent_pool_allowed_workspaces" "foobar"{
  agent_pool_id 		= tfe_agent_pool.foobar.id
  allowed_workspace_ids = [
	tfe_workspace.foobar.id,
	tfe_workspace.test-workspace.id
   ]
}`, organization, organization, organization)
}
