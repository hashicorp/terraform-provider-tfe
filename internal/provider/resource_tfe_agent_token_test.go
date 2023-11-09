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

func TestAccTFEAgentToken_basic(t *testing.T) {
	skipIfEnterprise(t)

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	agentToken := &tfe.AgentToken{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEAgentTokenDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEAgentToken_basic(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEAgentTokenExists(
						"tfe_agent_token.foobar", agentToken),
					testAccCheckTFEAgentTokenAttributes(agentToken),
					resource.TestCheckResourceAttr(
						"tfe_agent_token.foobar", "description", "agent-token-test"),
				),
			},
		},
	})
}

func testAccCheckTFEAgentTokenExists(
	n string, agentToken *tfe.AgentToken) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(ConfiguredClient)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no instance ID is set")
		}

		sk, err := config.Client.AgentTokens.Read(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}

		if sk == nil {
			return fmt.Errorf("agent token not found")
		}

		*agentToken = *sk

		return nil
	}
}

func testAccCheckTFEAgentTokenAttributes(
	agentToken *tfe.AgentToken) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if agentToken.Description != "agent-token-test" {
			return fmt.Errorf("bad name: %s", agentToken.Description)
		}
		return nil
	}
}

func testAccCheckTFEAgentTokenDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(ConfiguredClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_agent_token" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no instance ID is set")
		}

		_, err := config.Client.AgentTokens.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("agent token %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccTFEAgentToken_basic(organization string) string {
	return fmt.Sprintf(`
resource "tfe_agent_pool" "foobar" {
  name         = "agent-pool-test"
  organization = "%s"
}

resource "tfe_agent_token" "foobar" {
	agent_pool_id = tfe_agent_pool.foobar.id
	description   = "agent-token-test"
}`, organization)
}
