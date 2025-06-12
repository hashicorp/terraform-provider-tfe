// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccTFEWorkspaceSettings_basic(t *testing.T) {
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
					resource.TestCheckResourceAttr(
						"tfe_workspace_settings.foobar", "auto_apply", "false"),
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

func TestAccTFEWorkspaceSettings_stateSharing(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, cleanupOrg := createBusinessOrganization(t, tfeClient)
	t.Cleanup(cleanupOrg)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			// Start with local execution
			{
				Config: testAccTFEWorkspaceSettingsUnknownIDRemoteState(org.Name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"tfe_workspace_settings.foobar", "id"),
					resource.TestCheckResourceAttrSet(
						"tfe_workspace_settings.foobar", "workspace_id",
					),
				),
			},
		},
	})
}

func TestAccTFEWorkspaceSettings_basicOptions(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, cleanupOrg := createBusinessOrganization(t, tfeClient)
	t.Cleanup(cleanupOrg)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			// Start with local execution
			{
				Config: testAccTFEWorkspaceSettings_options(org.Name, "initial", true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"tfe_workspace_settings.foobar", "id"),
					resource.TestCheckResourceAttrSet(
						"tfe_workspace_settings.foobar", "workspace_id",
					),
					resource.TestCheckResourceAttr(
						"tfe_workspace_settings.foobar", "description", "initial"),
					resource.TestCheckResourceAttr(
						"tfe_workspace_settings.foobar", "auto_apply", "true"),
					resource.TestCheckResourceAttr(
						"tfe_workspace_settings.foobar", "assessments_enabled", "true"),
				),
			},
			{
				Config: testAccTFEWorkspaceSettings_options(org.Name, "updated", false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"tfe_workspace_settings.foobar", "id"),
					resource.TestCheckResourceAttrSet(
						"tfe_workspace_settings.foobar", "workspace_id",
					),
					resource.TestCheckResourceAttr(
						"tfe_workspace_settings.foobar", "description", "updated"),
					resource.TestCheckResourceAttr(
						"tfe_workspace_settings.foobar", "auto_apply", "false"),
					resource.TestCheckResourceAttr(
						"tfe_workspace_settings.foobar", "assessments_enabled", "false"),
				),
			},
		},
	})
}

func TestAccTFEWorkspaceSettingsRemoteState(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, cleanupOrg := createBusinessOrganization(t, tfeClient)
	t.Cleanup(cleanupOrg)

	ws := createTempWorkspace(t, tfeClient, org.Name)
	ws2 := createTempWorkspace(t, tfeClient, org.Name)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEWorkspaceSettingsDestroy,
		Steps: []resource.TestStep{
			// Have remote state consumer ids
			{
				Config: testAccTFEWorkspaceSettingsRemoteState(ws.ID, ws2.ID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"tfe_workspace_settings.foobar", "id"),
					resource.TestCheckResourceAttrSet(
						"tfe_workspace_settings.foobar", "workspace_id"),
					resource.TestCheckResourceAttr(
						"tfe_workspace_settings.foobar", "global_remote_state", "false"),
					resource.TestCheckResourceAttr(
						"tfe_workspace_settings.foobar", "remote_state_consumer_ids.0", ws2.ID),
					resource.TestCheckResourceAttr(
						"tfe_workspace_settings.foobar", "remote_state_consumer_ids.#", "1"),
				),
			},
			// Unset remote state consumer ids and set global remote state
			{
				Config: testAccTFEWorkspaceSettingsRemoteState_Global(ws.ID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"tfe_workspace_settings.foobar", "id"),
					resource.TestCheckResourceAttrSet(
						"tfe_workspace_settings.foobar", "workspace_id"),
					resource.TestCheckResourceAttr(
						"tfe_workspace_settings.foobar", "global_remote_state", "true"),
					resource.TestCheckResourceAttr(
						"tfe_workspace_settings.foobar", "remote_state_consumer_ids.#", "0"),
				),
			},
			// Unset execution mode
			{
				Config:      testAccTFEWorkspaceSettingsRemoteState_GlobalConflict(ws.ID, ws2.ID),
				ExpectError: regexp.MustCompile("If global_remote_state is true, remote_state_consumer_ids must not be set"),
			},
		},
	})
}

func TestAccTFEWorkspaceSettings_import(t *testing.T) {
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

func TestAccTFEWorkspaceSettings_importByName(t *testing.T) {
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

func TestAccTFEWorkspaceSettings_basicTags(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			// Start with local execution
			{
				Config: testAccTFEWorkspaceSettings_basicTagOne(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"tfe_workspace_settings.test", "tags.%", "1"),
					resource.TestCheckResourceAttr(
						"tfe_workspace_settings.test", "tags.keyA", "valueA"),
					resource.TestCheckResourceAttr(
						"tfe_workspace_settings.test", "effective_tags.%", "2"),
				),
			},
			{
				Config: testAccTFEWorkspaceSettings_basicTagTwo(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"tfe_workspace_settings.test", "tags.%", "2"),
					resource.TestCheckResourceAttr(
						"tfe_workspace_settings.test", "tags.keyA", "valueA"),
					resource.TestCheckResourceAttr(
						"tfe_workspace_settings.test", "tags.keyB", "valueB"),
					resource.TestCheckResourceAttr(
						"tfe_workspace_settings.test", "effective_tags.%", "3"),
				),
			},
			{
				Config: testAccTFEWorkspaceSettings_basicTagOne(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"tfe_workspace_settings.test", "tags.%", "1"),
					resource.TestCheckResourceAttr(
						"tfe_workspace_settings.test", "tags.keyA", "valueA"),
					resource.TestCheckResourceAttr(
						"tfe_workspace_settings.test", "effective_tags.%", "2"),
				),
			},
			{
				Config: testAccTFEWorkspaceSettings_basicTagZero(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"tfe_workspace_settings.test", "tags.%", "0"),
					resource.TestCheckResourceAttr(
						"tfe_workspace_settings.test", "effective_tags.%", "1"),
				),
			},
		},
	})
}

func testAccCheckTFEWorkspaceSettingsDestroy(s *terraform.State) error {
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

func testAccTFEWorkspaceSettingsUnknownIDRemoteState(orgName string) string {
	return fmt.Sprintf(`
resource "tfe_workspace" "foobar1" {
	name = "foobar1"
	organization = "%s"
}

resource "tfe_workspace" "foobar2" {
	name = "foobar2"
	organization = "%s"
}

resource "tfe_workspace_settings" "foobar" {
	workspace_id              = tfe_workspace.foobar1.id
	global_remote_state       = false
	remote_state_consumer_ids = [tfe_workspace.foobar2.id]
}
`, orgName, orgName)
}

func testAccTFEWorkspaceSettingsRemoteState(workspaceID, workspaceID2 string) string {
	return fmt.Sprintf(`
resource "tfe_workspace_settings" "foobar" {
	workspace_id              = "%s"
	global_remote_state       = false
	remote_state_consumer_ids = ["%s"]
}
`, workspaceID, workspaceID2)
}

func testAccTFEWorkspaceSettingsRemoteState_Global(workspaceID string) string {
	return fmt.Sprintf(`
resource "tfe_workspace_settings" "foobar" {
	workspace_id              = "%s"
	global_remote_state       = true
}
`, workspaceID)
}

func testAccTFEWorkspaceSettingsRemoteState_GlobalConflict(workspaceID, workspaceID2 string) string {
	return fmt.Sprintf(`
resource "tfe_workspace_settings" "foobar" {
	workspace_id              = "%s"
	global_remote_state       = true
	remote_state_consumer_ids = ["%s"]
}
`, workspaceID, workspaceID2)
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

func testAccTFEWorkspaceSettings_options(orgName string, description string, boolOptions bool) string {
	return fmt.Sprintf(`
resource "tfe_workspace" "foobar1" {
	name = "foobar1"
	organization = "%s"
}

resource "tfe_workspace_settings" "foobar" {
	workspace_id              = tfe_workspace.foobar1.id
	description               = "%s"
	auto_apply                = %t
	assessments_enabled       = %t
}
`, orgName, description, boolOptions, boolOptions)
}

func tagBase(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "test" {
  name  = "tst-tfeprovider-%d"
  email = "admin@company.com"
}

resource "tfe_project" "test" {
  organization = tfe_organization.test.name
  name = "tfe-provider-test-%d"
	tags = {
	  projectTag = "valueA"
	}
}

resource "tfe_workspace" "test" {
	name         = "tfe-provider-test-workspace-%d"
	organization = tfe_organization.test.name
    project_id   = tfe_project.test.id 
}
`, rInt, rInt, rInt)
}

func testAccTFEWorkspaceSettings_basicTagOne(rInt int) string {
	return tagBase(rInt) + `
resource "tfe_workspace_settings" "test" {
	workspace_id = tfe_workspace.test.id
	tags = {
	  keyA = "valueA"
	}
}
`
}

func testAccTFEWorkspaceSettings_basicTagTwo(rInt int) string {
	return tagBase(rInt) + `
resource "tfe_workspace_settings" "test" {
	workspace_id = tfe_workspace.test.id
	tags = {
	  keyA = "valueA"
	  keyB = "valueB"	
	}
}
`
}

func testAccTFEWorkspaceSettings_basicTagZero(rInt int) string {
	return tagBase(rInt) + `
resource "tfe_workspace_settings" "test" {
	workspace_id = tfe_workspace.test.id
	tags = {}
}
`
}
