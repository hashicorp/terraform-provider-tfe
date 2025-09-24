// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"errors"
	"fmt"
	"testing"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccTFEAgentPoolExcludedWorkspaces_create_update(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	excludedWorkspaceIDs := &[]string{}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEAgentPoolExcludedWorkspaces_basic(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEAgentPoolExcludedWorkspacesExists("tfe_agent_pool.foobar", excludedWorkspaceIDs),
					testAccCheckTFEAgentPoolExcludedWorkspacesCount(2, excludedWorkspaceIDs),
				),
			},
			{
				Config: testAccTFEAgentPoolExcludedWorkspaces_update(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEAgentPoolExcludedWorkspacesExists("tfe_agent_pool.foobar", excludedWorkspaceIDs),
					testAccCheckTFEAgentPoolExcludedWorkspacesCount(1, excludedWorkspaceIDs),
				),
			},
			{
				Config: testAccTFEAgentPoolExcludedWorkspaces_destroy(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEAgentPoolExcludedWorkspacesNotExists("tfe_agent_pool.foobar"),
				),
			},
		},
	})
}

func testAccCheckTFEAgentPoolExcludedWorkspacesExists(resourceName string, excludedWorkspaces *[]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		*excludedWorkspaces = []string{}

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		// Resource ID equals the Agent Pool ID
		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		agentPool, err := testAccConfiguredClient.Client.AgentPools.Read(ctx, rs.Primary.ID)
		if err != nil && !errors.Is(err, tfe.ErrResourceNotFound) {
			return fmt.Errorf("error while fetching agent pool: %w", err)
		}

		if len(agentPool.ExcludedWorkspaces) == 0 {
			return fmt.Errorf("Excluded Workspaces for agent pool %s do not exist", rs.Primary.ID)
		}

		for _, workspace := range agentPool.ExcludedWorkspaces {
			*excludedWorkspaces = append(*excludedWorkspaces, workspace.ID)
		}

		return nil
	}
}

func testAccCheckTFEAgentPoolExcludedWorkspacesNotExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		// Resource ID equals the Agent Pool ID
		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		agentPool, err := testAccConfiguredClient.Client.AgentPools.Read(ctx, rs.Primary.ID)
		if err != nil && !errors.Is(err, tfe.ErrResourceNotFound) {
			return fmt.Errorf("error while fetching agent pool: %w", err)
		}

		if len(agentPool.ExcludedWorkspaces) > 0 {
			return fmt.Errorf("Excluded Workspaces for agent pool %s exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckTFEAgentPoolExcludedWorkspacesCount(expected int, excludedWorkspaces *[]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if len(*excludedWorkspaces) != expected {
			return fmt.Errorf("expected %d excluded workspaces, got %d", expected, len(*excludedWorkspaces))
		}
		return nil
	}
}

func TestAccTFEAgentPoolExcludedWorkspaces_import(t *testing.T) {
	skipIfEnterprise(t)

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEAgentPoolExcludedWorkspaces_basic(org.Name),
			},
			{
				ResourceName:      "tfe_agent_pool_excluded_workspaces.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccTFEAgentPoolExcludedWorkspaces_destroy(organization string) string {
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

func testAccTFEAgentPoolExcludedWorkspaces_update(organization string) string {
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

resource "tfe_agent_pool_excluded_workspaces" "foobar"{
  agent_pool_id 		= tfe_agent_pool.foobar.id
  excluded_workspace_ids = [tfe_workspace.foobar.id]
}`, organization, organization, organization)
}

func testAccTFEAgentPoolExcludedWorkspaces_basic(organization string) string {
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

resource "tfe_agent_pool_excluded_workspaces" "foobar"{
  agent_pool_id 		= tfe_agent_pool.foobar.id
  excluded_workspace_ids = [
	tfe_workspace.foobar.id,
	tfe_workspace.test-workspace.id
   ]
}`, organization, organization, organization)
}
