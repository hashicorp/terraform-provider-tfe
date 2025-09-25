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

func TestAccTFEAgentPoolAllowedProjects_create_update(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	allowedProjectsIDs := &[]string{}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEAgentPoolAllowedProjects_basic(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEAgentPoolAllowedProjectExists("tfe_agent_pool.foobar", allowedProjectsIDs),
					testAccCheckTFEAgentPoolAllowedProjectsCount(2, allowedProjectsIDs),
				),
			},
			{
				Config: testAccTFEAgentPoolAllowedProjects_update(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEAgentPoolAllowedProjectExists("tfe_agent_pool.foobar", allowedProjectsIDs),
					testAccCheckTFEAgentPoolAllowedProjectsCount(1, allowedProjectsIDs),
				),
			},
			{
				Config: testAccTFEAgentPoolAllowedProjects_destroy(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEAgentPoolAllowedProjectsNotExists("tfe_agent_pool.foobar"),
				),
			},
		},
	})
}

func testAccCheckTFEAgentPoolAllowedProjectExists(resourceName string, allowedProjects *[]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		*allowedProjects = []string{}

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

		if len(agentPool.AllowedProjects) == 0 {
			return fmt.Errorf("Allowed Projects for agent pool %s do not exist", rs.Primary.ID)
		}

		for _, project := range agentPool.AllowedProjects {
			*allowedProjects = append(*allowedProjects, project.ID)
		}

		return nil
	}
}

func testAccCheckTFEAgentPoolAllowedProjectsNotExists(resourceName string) resource.TestCheckFunc {
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

		if len(agentPool.AllowedProjects) > 0 {
			return fmt.Errorf("Allowed Projects for agent pool %s exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckTFEAgentPoolAllowedProjectsCount(expected int, allowedProjects *[]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if len(*allowedProjects) != expected {
			return fmt.Errorf("expected %d allowed projects, got %d", expected, len(*allowedProjects))
		}
		return nil
	}
}

func TestAccTFEAgentPoolAllowedProjects_import(t *testing.T) {
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
				Config: testAccTFEAgentPoolAllowedProjects_basic(org.Name),
			},
			{
				ResourceName:      "tfe_agent_pool_allowed_projects.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccTFEAgentPoolAllowedProjects_destroy(organization string) string {
	return fmt.Sprintf(`
resource "tfe_project" "foobar" {
  name = "foobar"
  organization = "%s"
}

resource "tfe_project" "test-project" {
  name = "test-project"
  organization = "%s"
}

resource "tfe_agent_pool" "foobar" {
  name         = "agent-pool-updated"
  organization = "%s"
  organization_scoped = false
}`, organization, organization, organization)
}

func testAccTFEAgentPoolAllowedProjects_update(organization string) string {
	return fmt.Sprintf(`
resource "tfe_project" "foobar" {
  name = "foobar"
  organization = "%s"
}

resource "tfe_project" "test-project" {
  name = "test-project"
  organization = "%s"
}

resource "tfe_agent_pool" "foobar" {
  name         = "agent-pool-updated"
  organization = "%s"
  organization_scoped = false
}

resource "tfe_agent_pool_allowed_projects" "foobar"{
  agent_pool_id 		= tfe_agent_pool.foobar.id
  allowed_project_ids = [tfe_project.foobar.id]
}`, organization, organization, organization)
}

func testAccTFEAgentPoolAllowedProjects_basic(organization string) string {
	return fmt.Sprintf(`
resource "tfe_project" "foobar" {
  name = "foobar"
  organization = "%s"
}

resource "tfe_project" "test-project" {
  name = "test-project"
  organization = "%s"
}

resource "tfe_agent_pool" "foobar" {
  name         = "agent-pool-updated"
  organization = "%s"
  organization_scoped = false
}

resource "tfe_agent_pool_allowed_projects" "foobar"{
  agent_pool_id 		= tfe_agent_pool.foobar.id
  allowed_project_ids = [
	tfe_project.foobar.id,
	tfe_project.test-project.id
   ]
}`, organization, organization, organization)
}
