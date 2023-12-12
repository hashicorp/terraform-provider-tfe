// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccTFEWorkspaceSettings(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, cleanupOrg := createBusinessOrganization(t, tfeClient)
	t.Cleanup(cleanupOrg)

	ws := createTempWorkspace(t, tfeClient, org.Name)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEWorkspaceSettingsDestroy,
		Steps: []resource.TestStep{
			// Start with local execution
			{
				Config: testAccTFEWorkspaceSettings_basic(ws.ID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"tfe_workspace_settings.foobar", "id"),
					resource.TestCheckResourceAttrSet(
						"tfe_workspace_settings.foobar", "workspace_id"),
					resource.TestCheckResourceAttr(
						"tfe_workspace_settings.foobar", "execution_mode", "local"),
					resource.TestCheckNoResourceAttr(
						"tfe_workspace_settings.foobar", "agent_pool_id"),
					resource.TestCheckResourceAttr(
						"tfe_workspace_settings.foobar", "overwrites.0.execution_mode", "true"),
					resource.TestCheckResourceAttr(
						"tfe_workspace_settings.foobar", "overwrites.0.agent_pool", "true"),
					resource.TestCheckResourceAttr(
						"tfe_workspace_settings.foobar", "overwrites.#", "1"),
				),
			},
			// Change to agent pool
			{
				Config: testAccTFEWorkspaceSettings_updateExecutionMode(org.Name, ws.ID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"tfe_workspace_settings.foobar", "id"),
					resource.TestCheckResourceAttrSet(
						"tfe_workspace_settings.foobar", "workspace_id"),
					resource.TestCheckResourceAttr(
						"tfe_workspace_settings.foobar", "execution_mode", "agent"),
					resource.TestCheckResourceAttrSet(
						"tfe_workspace_settings.foobar", "agent_pool_id"),
					resource.TestCheckResourceAttr(
						"tfe_workspace_settings.foobar", "overwrites.0.execution_mode", "true"),
					resource.TestCheckResourceAttr(
						"tfe_workspace_settings.foobar", "overwrites.0.agent_pool", "true"),
					resource.TestCheckResourceAttr(
						"tfe_workspace_settings.foobar", "overwrites.#", "1"),
				),
			},
			// Unset execution mode
			{
				Config: testAccTFEWorkspaceSettings_unsetExecutionMode(org.Name, ws.ID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"tfe_workspace_settings.foobar", "id"),
					resource.TestCheckResourceAttrSet(
						"tfe_workspace_settings.foobar", "workspace_id"),
					resource.TestCheckResourceAttr(
						"tfe_workspace_settings.foobar", "execution_mode", "remote"),
					resource.TestCheckNoResourceAttr(
						"tfe_workspace_settings.foobar", "agent_pool_id"),
					resource.TestCheckResourceAttr(
						"tfe_workspace_settings.foobar", "overwrites.0.execution_mode", "false"),
					resource.TestCheckResourceAttr(
						"tfe_workspace_settings.foobar", "overwrites.0.agent_pool", "false"),
					resource.TestCheckResourceAttr(
						"tfe_workspace_settings.foobar", "overwrites.#", "1"),
				),
			},
		},
	})
}

func TestAccTFEWorkspaceSettingsImport(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, cleanupOrg := createBusinessOrganization(t, tfeClient)
	t.Cleanup(cleanupOrg)

	ws := createTempWorkspace(t, tfeClient, org.Name)

	_, err = tfeClient.Workspaces.UpdateByID(ctx, ws.ID, tfe.WorkspaceUpdateOptions{
		ExecutionMode: tfe.String("local"),
	})
	if err != nil {
		t.Fatal(err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEWorkspaceSettingsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEWorkspaceSettings_basic(ws.ID),
			},
			{
				ResourceName:      "tfe_workspace_settings.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTFEWorkspaceSettingsImport_ByName(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, cleanupOrg := createBusinessOrganization(t, tfeClient)
	t.Cleanup(cleanupOrg)

	ws := createTempWorkspace(t, tfeClient, org.Name)

	_, err = tfeClient.Workspaces.UpdateByID(ctx, ws.ID, tfe.WorkspaceUpdateOptions{
		ExecutionMode: tfe.String("local"),
	})
	if err != nil {
		t.Fatal(err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEOrganizationMembershipDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEWorkspaceSettings_basic(ws.ID),
			},
			{
				ResourceName:      "tfe_workspace_settings.foobar",
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s/%s", org.Name, ws.Name),
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckTFEWorkspaceSettingsDestroy(s *terraform.State) error {
	return testAccCheckTFEWorkspaceSettingsDestroyProvider(testAccProvider)(s)
}

func testAccCheckTFEWorkspaceSettingsDestroyProvider(p *schema.Provider) func(s *terraform.State) error {
	return func(s *terraform.State) error {
		tfeClient, err := getClientUsingEnv()
		if err != nil {
			return err
		}

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "tfe_workspace_settings" {
				continue
			}

			if rs.Primary.ID == "" {
				return fmt.Errorf("No instance ID is set")
			}

			ws, err := tfeClient.Workspaces.ReadByID(ctx, rs.Primary.ID)
			if err != nil {
				return fmt.Errorf("Workspace %s does not exist", rs.Primary.ID)
			}

			if ws.ExecutionMode != "remote" {
				return fmt.Errorf("expected execution mode to be remote after destroy, but was %s", ws.ExecutionMode)
			}

			if ws.AgentPool != nil {
				return errors.New("expected agent pool to be nil after destroy, but wasn't")
			}
		}

		return nil
	}
}

func testAccTFEWorkspaceSettings_basic(workspaceID string) string {
	return fmt.Sprintf(`
resource "tfe_workspace_settings" "foobar" {
	workspace_id   = "%s"
	execution_mode = "local"
}
`, workspaceID)
}

func testAccTFEWorkspaceSettings_updateExecutionMode(orgName, workspaceID string) string {
	return fmt.Sprintf(`
resource "tfe_agent_pool" "mypool" {
	name = "test-pool-default"
	organization = "%s"
}

resource "tfe_workspace_settings" "foobar" {
	workspace_id   = "%s"
	execution_mode = "agent"
	agent_pool_id  = tfe_agent_pool.mypool.id
}
`, orgName, workspaceID)
}

func testAccTFEWorkspaceSettings_unsetExecutionMode(orgName, workspaceID string) string {
	return fmt.Sprintf(`
resource "tfe_agent_pool" "mypool" {
	name = "test-pool-default"
	organization = "%s"
}

resource "tfe_workspace_settings" "foobar" {
	workspace_id   = "%s"
}
`, orgName, workspaceID)
}
