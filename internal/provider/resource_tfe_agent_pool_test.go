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

func TestAccTFEAgentPool_basic(t *testing.T) {
	skipIfEnterprise(t)

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	agentPool := &tfe.AgentPool{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEAgentPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEAgentPool_basic(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEAgentPoolExists(
						"tfe_agent_pool.foobar", agentPool),
					testAccCheckTFEAgentPoolAttributes(agentPool),
					resource.TestCheckResourceAttr(
						"tfe_agent_pool.foobar", "name", "agent-pool-test"),
					resource.TestCheckResourceAttr(
						"tfe_agent_pool.foobar", "organization_scoped", "true"),
				),
			},
		},
	})
}

func TestAccTFEAgentPool_custom_scope(t *testing.T) {
	skipIfEnterprise(t)

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	agentPool := &tfe.AgentPool{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEAgentPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEAgentPool_custom_scope(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEAgentPoolExists(
						"tfe_agent_pool.foobar", agentPool),
					testAccCheckTFEAgentPoolAttributes(agentPool),
					resource.TestCheckResourceAttr(
						"tfe_agent_pool.foobar", "name", "agent-pool-test"),
					resource.TestCheckResourceAttr(
						"tfe_agent_pool.foobar", "organization_scoped", "false"),
				),
			},
		},
	})
}

func TestAccTFEAgentPool_update(t *testing.T) {
	skipIfEnterprise(t)

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	agentPool := &tfe.AgentPool{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEAgentPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEAgentPool_basic(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEAgentPoolExists(
						"tfe_agent_pool.foobar", agentPool),
					testAccCheckTFEAgentPoolAttributes(agentPool),
					resource.TestCheckResourceAttr(
						"tfe_agent_pool.foobar", "name", "agent-pool-test"),
					resource.TestCheckResourceAttr(
						"tfe_agent_pool.foobar", "organization_scoped", "true"),
				),
			},

			{
				Config: testAccTFEAgentPool_update(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEAgentPoolExists(
						"tfe_agent_pool.foobar", agentPool),
					testAccCheckTFEAgentPoolAttributesUpdated(agentPool),
					resource.TestCheckResourceAttr(
						"tfe_agent_pool.foobar", "name", "agent-pool-updated"),
					resource.TestCheckResourceAttr(
						"tfe_agent_pool.foobar", "organization_scoped", "false"),
				),
			},
		},
	})
}

func TestAccTFEAgentPool_import(t *testing.T) {
	skipIfEnterprise(t)

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEAgentPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEAgentPool_basic(org.Name),
			},
			{
				ResourceName:      "tfe_agent_pool.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "tfe_agent_pool.foobar",
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s/agent-pool-test", org.Name),
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckTFEAgentPoolExists(
	n string, agentPool *tfe.AgentPool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(ConfiguredClient)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no instance ID is set")
		}

		sk, err := config.Client.AgentPools.Read(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}

		if sk == nil {
			return fmt.Errorf("agent pool not found")
		}

		*agentPool = *sk

		return nil
	}
}

func testAccCheckTFEAgentPoolAttributes(
	agentPool *tfe.AgentPool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if agentPool.Name != "agent-pool-test" {
			return fmt.Errorf("bad name: %s", agentPool.Name)
		}
		return nil
	}
}

func testAccCheckTFEAgentPoolAttributesUpdated(
	agentPool *tfe.AgentPool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if agentPool.Name != "agent-pool-updated" {
			return fmt.Errorf("bad name: %s", agentPool.Name)
		}
		return nil
	}
}

func testAccCheckTFEAgentPoolDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(ConfiguredClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_agent_pool" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no instance ID is set")
		}

		_, err := config.Client.AgentPools.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("agent pool %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccTFEAgentPool_basic(organization string) string {
	return fmt.Sprintf(`
resource "tfe_agent_pool" "foobar" {
  name         = "agent-pool-test"
  organization = "%s"
}`, organization)
}

func testAccTFEAgentPool_custom_scope(organization string) string {
	return fmt.Sprintf(`
resource "tfe_agent_pool" "foobar" {
  name         = "agent-pool-test"
  organization = "%s"
  organization_scoped = false
}`, organization)
}

func testAccTFEAgentPool_update(organization string) string {
	return fmt.Sprintf(`
resource "tfe_workspace" "foobar" {
  name = "foobar"
  organization = "%s"
}

resource "tfe_agent_pool" "foobar" {
  name         = "agent-pool-updated"
  organization = "%s"
  organization_scoped = false
}`, organization, organization)
}
