// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/hashicorp/go-tfe"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccTFEAgentPoolDataSource_basic(t *testing.T) {
	skipIfEnterprise(t)

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
				Config: testAccTFEAgentPoolDataSourceConfig(org.Name, rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tfe_agent_pool.foobar", "id"),
					resource.TestCheckResourceAttr(
						"data.tfe_agent_pool.foobar", "name", fmt.Sprintf("agent-pool-test-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.tfe_agent_pool.foobar", "organization", org.Name),
					resource.TestCheckResourceAttr(
						"data.tfe_agent_pool.foobar", "organization_scoped", "true"),
				),
			},
		},
	})
}

func TestAccTFEAgentPoolDataSource_allowed_workspaces(t *testing.T) {
	skipIfEnterprise(t)

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	ws, err := tfeClient.Workspaces.Create(ctx, org.Name, tfe.WorkspaceCreateOptions{
		Name: tfe.String(fmt.Sprintf("tst-workspace-test-%d", rInt)),
	})
	if err != nil {
		t.Fatal(err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEAgentPoolDataSourceAllowedWorkspacesConfig(org.Name, rInt, ws.ID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tfe_agent_pool.foobar", "id"),
					resource.TestCheckResourceAttr(
						"data.tfe_agent_pool.foobar", "name", fmt.Sprintf("agent-pool-test-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.tfe_agent_pool.foobar", "organization", org.Name),
					resource.TestCheckResourceAttr(
						"data.tfe_agent_pool.foobar", "organization_scoped", "false"),
					resource.TestCheckResourceAttr(
						"data.tfe_agent_pool.foobar", "allowed_workspace_ids.0", ws.ID),
				),
			},
		},
	})
}

func testAccTFEAgentPoolDataSourceConfig(organization string, rInt int) string {
	return fmt.Sprintf(`
resource "tfe_agent_pool" "foobar" {
  name                  = "agent-pool-test-%d"
  organization          = "%s"
}

data "tfe_agent_pool" "foobar" {
  name         = tfe_agent_pool.foobar.name
  organization = "%s"
}`, rInt, organization, organization)
}

func testAccTFEAgentPoolDataSourceAllowedWorkspacesConfig(organization string, rInt int, workspaceID string) string {
	return fmt.Sprintf(`
resource "tfe_agent_pool" "foobar" {
  name                  = "agent-pool-test-%d"
  organization          = "%s"
  organization_scoped   = false
}

resource "tfe_agent_pool_allowed_workspaces" "foobar" {
	agent_pool_id = tfe_agent_pool.foobar.id
  allowed_workspace_ids = ["%s"]
}

data "tfe_agent_pool" "foobar" {
  name         = tfe_agent_pool.foobar.name
  organization = "%s"
}`, rInt, organization, workspaceID, organization)
}
