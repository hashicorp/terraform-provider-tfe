// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"regexp"
	"testing"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestTFEWorkspaceRunTask_stagesSupport(t *testing.T) {
	testCases := map[string]struct {
		isCloud      bool
		tfeVer       string
		expectResult bool
	}{
		"when HCP Terraform":                 {true, "", true},
		"when HCP Terraform but TFE version": {true, "v202402-2", true}, // Technically this shouldn't happen, but just in case
		"when Enterprise < v202208-3":        {false, "", false},
		"when Enterprise v202402":            {false, "v202402-2", false},
		"when Enterprise v202404-1":          {false, "v202404-1", true},
		"when Enterprise v202408-1":          {false, "v202408-1", true},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			resolver := &staticCapabilityResolver{}
			resolver.SetIsCloud(testCase.isCloud)
			resolver.SetRemoteTFEVersion(testCase.tfeVer)

			subject := resourceWorkspaceRunTask{
				config:       ConfiguredClient{Organization: "Mock", Client: &tfe.Client{}},
				capabilities: resolver,
			}

			actual := subject.supportsStagesProperty()
			if actual != testCase.expectResult {
				t.Fatalf("expected supportsStagesProperty to be %t, got %t", testCase.expectResult, actual)
			}
		})
	}
}

func TestAccTFEWorkspaceRunTask_validateSchemaAttributes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFEWorkspaceRunTask_attributes("bad_level", string(tfe.PostPlan), fmt.Sprintf("[%q]", tfe.PostPlan)),
				ExpectError: regexp.MustCompile(`enforcement_level value must be one of:`),
			},
			{
				Config:      testAccTFEWorkspaceRunTask_attributes(string(tfe.Advisory), "bad_stage", fmt.Sprintf("[%q]", tfe.PostPlan)),
				ExpectError: regexp.MustCompile(`stage value must be one of:`),
			},
			{
				Config:      testAccTFEWorkspaceRunTask_attributes(string(tfe.Advisory), string(tfe.PostPlan), `"not an array"`),
				ExpectError: regexp.MustCompile(`Inappropriate value for attribute "stages"`),
			},
			{
				Config:      testAccTFEWorkspaceRunTask_attributes(string(tfe.Advisory), string(tfe.PostPlan), `["bad_stage"]`),
				ExpectError: regexp.MustCompile(`stages\[0\] value must be one of:`),
			},
		},
	})
}

func TestAccTFEWorkspaceRunTask_create_stages_attr(t *testing.T) {
	skipUnlessRunTasksDefined(t)

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	workspaceTask := &tfe.WorkspaceRunTask{}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEWorkspaceRunTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEWorkspaceRunTask_basic_stages_attr(org.Name, runTasksURL()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceRunTaskExists("tfe_workspace_run_task.foobar", workspaceTask),
					resource.TestCheckResourceAttr("tfe_workspace_run_task.foobar", "enforcement_level", "advisory"),
					resource.TestCheckResourceAttr("tfe_workspace_run_task.foobar", "stages.#", "2"),
					resource.TestCheckResourceAttr("tfe_workspace_run_task.foobar", "stages.0", "post_plan"),
					resource.TestCheckResourceAttr("tfe_workspace_run_task.foobar", "stages.1", "pre_plan"),
				),
			},
			{
				Config: testAccTFEWorkspaceRunTask_update_stages_attr(org.Name, runTasksURL()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tfe_workspace_run_task.foobar", "enforcement_level", "mandatory"),
					resource.TestCheckResourceAttr("tfe_workspace_run_task.foobar", "stages.#", "2"),
					resource.TestCheckResourceAttr("tfe_workspace_run_task.foobar", "stages.0", "pre_apply"),
					resource.TestCheckResourceAttr("tfe_workspace_run_task.foobar", "stages.1", "post_apply"),
				),
			},
		},
	})
}

func TestAccTFEWorkspaceRunTask_create_stage_attr(t *testing.T) {
	skipUnlessRunTasksDefined(t)

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	workspaceTask := &tfe.WorkspaceRunTask{}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEWorkspaceRunTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEWorkspaceRunTask_basic_stage_attr(org.Name, runTasksURL()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceRunTaskExists("tfe_workspace_run_task.foobar", workspaceTask),
					resource.TestCheckResourceAttr("tfe_workspace_run_task.foobar", "enforcement_level", "advisory"),
					resource.TestCheckResourceAttr("tfe_workspace_run_task.foobar", "stage", "post_plan"),
					resource.TestCheckResourceAttr("tfe_workspace_run_task.foobar", "stages.#", "1"),
					resource.TestCheckResourceAttr("tfe_workspace_run_task.foobar", "stages.0", "post_plan"),
				),
			},
			{
				Config: testAccTFEWorkspaceRunTask_update_stage_attr(org.Name, runTasksURL()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tfe_workspace_run_task.foobar", "enforcement_level", "mandatory"),
					resource.TestCheckResourceAttr("tfe_workspace_run_task.foobar", "stage", "pre_plan"),
					resource.TestCheckResourceAttr("tfe_workspace_run_task.foobar", "stages.#", "1"),
					resource.TestCheckResourceAttr("tfe_workspace_run_task.foobar", "stages.0", "pre_plan"),
				),
			},
		},
	})
}

func TestAccTFEWorkspaceRunTask_import(t *testing.T) {
	skipUnlessRunTasksDefined(t)

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFETeamAccessDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEWorkspaceRunTask_basic_stage_attr(org.Name, runTasksURL()),
			},
			{
				ResourceName:      "tfe_workspace_run_task.foobar",
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s/workspace-test/foobar-task", org.Name),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTFEWorkspaceRunTask_Read(t *testing.T) {
	skipUnlessRunTasksDefined(t)

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	// Create test fixtures
	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)
	ws := createTempWorkspace(t, tfeClient, org.Name)
	key := runTasksHMACKey()
	task := createRunTask(t, tfeClient, org.Name, tfe.RunTaskCreateOptions{
		Name:    fmt.Sprintf("tst-task-%s", randomString(t)),
		URL:     runTasksURL(),
		HMACKey: &key,
	})

	org_tf := fmt.Sprintf(`data "tfe_organization" "orgtask" { name = %q }`, org.Name)

	create_wstask_tf := fmt.Sprintf(`
		%s
		resource "tfe_workspace_run_task" "foobar" {
			workspace_id      = %q
			task_id           = %q
			enforcement_level = "advisory"
			stage             = "post_plan"
		}
		`, org_tf, ws.ID, task.ID)

	delete_wstasks := func() {
		wstasks, err := tfeClient.WorkspaceRunTasks.List(ctx, ws.ID, nil)
		if err != nil || wstasks == nil {
			t.Fatalf("Error listing tasks: %s", err)
			return
		}
		// There shouldn't be more that 25 run tasks so we don't need to worry about pagination
		for _, wstask := range wstasks.Items {
			if wstask != nil {
				if err := tfeClient.WorkspaceRunTasks.Delete(ctx, ws.ID, wstask.ID); err != nil {
					t.Fatalf("Error deleting workspace task: %s", err)
				}
			}
		}
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFETeamAccessDestroy,
		Steps: []resource.TestStep{
			{
				Config: create_wstask_tf,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tfe_workspace_run_task.foobar", "enforcement_level", "advisory"),
				),
			},
			{
				// Delete the created workspace run task and ensure we can re-create it
				PreConfig: delete_wstasks,
				Config:    create_wstask_tf,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tfe_workspace_run_task.foobar", "enforcement_level", "advisory"),
				),
			},
			{
				// Delete the created workspace run task and ensure we can ignore it if we no longer need to manage it
				PreConfig: delete_wstasks,
				Config:    org_tf,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceRunTaskDestroy,
				),
			},
		},
	})
}

func testAccCheckTFEWorkspaceRunTaskExists(n string, runTask *tfe.WorkspaceRunTask) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(ConfiguredClient)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		if rs.Primary.Attributes["workspace_id"] == "" {
			return fmt.Errorf("No Workspace ID is set")
		}

		rt, err := config.Client.WorkspaceRunTasks.Read(ctx, rs.Primary.Attributes["workspace_id"], rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error reading Workspace Run Task: %w", err)
		}

		if rt == nil {
			return fmt.Errorf("Workspace Run Task not found")
		}

		*runTask = *rt

		return nil
	}
}

func testAccCheckTFEWorkspaceRunTaskDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(ConfiguredClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_workspace_run_task" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}
		if rs.Primary.Attributes["workspace_id"] == "" {
			return fmt.Errorf("No Workspace ID is set")
		}

		_, err := config.Client.WorkspaceRunTasks.Read(ctx, rs.Primary.Attributes["workspace_id"], rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Workspace Run Tasks %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccTFEWorkspaceRunTask_attributes(enforcementLevel, stage, stages string) string {
	return fmt.Sprintf(`
resource "tfe_workspace_run_task" "foobar" {
  workspace_id      = "ws-abc123"
  task_id           = "task-abc123"
  enforcement_level = "%s"
  stage             = "%s"
	stages            = %s
}
`, enforcementLevel, stage, stages)
}

func testAccTFEWorkspaceRunTask_basic_stage_attr(orgName, runTaskURL string) string {
	return fmt.Sprintf(`
locals {
    organization_name = "%s"
}
resource "tfe_organization_run_task" "foobar" {
  organization = local.organization_name
  url          = "%s"
  name         = "foobar-task"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = local.organization_name
}

resource "tfe_workspace_run_task" "foobar" {
  workspace_id      = resource.tfe_workspace.foobar.id
  task_id           = resource.tfe_organization_run_task.foobar.id
  enforcement_level = "advisory"
  stage             = "post_plan"
}
`, orgName, runTaskURL)
}

func testAccTFEWorkspaceRunTask_basic_stages_attr(orgName, runTaskURL string) string {
	return fmt.Sprintf(`
locals {
    organization_name = "%s"
}
resource "tfe_organization_run_task" "foobar" {
  organization = local.organization_name
  url          = "%s"
  name         = "foobar-task"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = local.organization_name
}

resource "tfe_workspace_run_task" "foobar" {
  workspace_id      = resource.tfe_workspace.foobar.id
  task_id           = resource.tfe_organization_run_task.foobar.id
  enforcement_level = "advisory"
	stages            = ["post_plan", "pre_plan"]
}
`, orgName, runTaskURL)
}

func testAccTFEWorkspaceRunTask_update_stage_attr(orgName, runTaskURL string) string {
	return fmt.Sprintf(`
locals {
    organization_name = "%s"
}

resource "tfe_organization_run_task" "foobar" {
  organization = local.organization_name
  url          = "%s"
  name         = "foobar-task"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = local.organization_name
}

resource "tfe_workspace_run_task" "foobar" {
  workspace_id      = resource.tfe_workspace.foobar.id
  task_id           = resource.tfe_organization_run_task.foobar.id
  enforcement_level = "mandatory"
  stage             = "pre_plan"
}
`, orgName, runTaskURL)
}

func testAccTFEWorkspaceRunTask_update_stages_attr(orgName, runTaskURL string) string {
	return fmt.Sprintf(`
locals {
    organization_name = "%s"
}

resource "tfe_organization_run_task" "foobar" {
  organization = local.organization_name
  url          = "%s"
  name         = "foobar-task"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = local.organization_name
}

resource "tfe_workspace_run_task" "foobar" {
  workspace_id      = resource.tfe_workspace.foobar.id
  task_id           = resource.tfe_organization_run_task.foobar.id
  enforcement_level = "mandatory"
  stages            = ["pre_apply", "post_apply"]
}
`, orgName, runTaskURL)
}
