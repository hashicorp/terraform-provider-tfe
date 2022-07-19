package tfe

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccTFEWorkspaceRunTask_create(t *testing.T) {
	skipUnlessRunTasksDefined(t)
	skipIfFreeOnly(t) // Run Tasks requires TFE or a TFC paid/trial subscription

	workspaceTask := &tfe.WorkspaceRunTask{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEWorkspaceRunTaskDestroy,
		Steps: []resource.TestStep{
			testCheckCreateOrgWithRunTasks(orgName),
			{
				Config: testAccTFEWorkspaceRunTask_basic(orgName, runTasksURL()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceRunTaskExists("tfe_workspace_run_task.foobar", workspaceTask),
					resource.TestCheckResourceAttr("tfe_workspace_run_task.foobar", "enforcement_level", "advisory"),
				),
			},
			{
				Config: testAccTFEWorkspaceRunTask_update(orgName, runTasksURL()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tfe_workspace_run_task.foobar", "enforcement_level", "mandatory"),
				),
			},
		},
	})
}

func TestAccTFEWorkspaceRunTask_beta_create(t *testing.T) {
	skipUnlessRunTasksDefined(t)
	skipUnlessBeta(t)
	skipIfFreeOnly(t) // Run Tasks requires TFE or a TFC paid/trial subscription

	workspaceTask := &tfe.WorkspaceRunTask{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEWorkspaceRunTaskDestroy,
		Steps: []resource.TestStep{
			testCheckCreateOrgWithRunTasks(orgName),
			{
				Config: testAccTFEWorkspaceRunTask_beta_basic(orgName, runTasksURL()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceRunTaskExists("tfe_workspace_run_task.foobar", workspaceTask),
					resource.TestCheckResourceAttr("tfe_workspace_run_task.foobar", "enforcement_level", "advisory"),
					resource.TestCheckResourceAttr("tfe_workspace_run_task.foobar", "stage", "post_plan"),
				),
			},
			{
				Config: testAccTFEWorkspaceRunTask_beta_update(orgName, runTasksURL()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tfe_workspace_run_task.foobar", "enforcement_level", "mandatory"),
					resource.TestCheckResourceAttr("tfe_workspace_run_task.foobar", "stage", "pre_plan"),
				),
			},
		},
	})
}

func TestAccTFEWorkspaceRunTask_import(t *testing.T) {
	skipUnlessRunTasksDefined(t)
	skipIfFreeOnly(t) // Run Tasks requires TFE or a TFC paid/trial subscription

	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFETeamAccessDestroy,
		Steps: []resource.TestStep{
			testCheckCreateOrgWithRunTasks(orgName),
			{
				Config: testAccTFEWorkspaceRunTask_basic(orgName, runTasksURL()),
			},
			{
				ResourceName:      "tfe_workspace_run_task.foobar",
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("tst-terraform-%d/workspace-test/foobar-task", rInt),
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckTFEWorkspaceRunTaskExists(n string, runTask *tfe.WorkspaceRunTask) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		tfeClient := testAccProvider.Meta().(*tfe.Client)

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

		rt, err := tfeClient.WorkspaceRunTasks.Read(ctx, rs.Primary.Attributes["workspace_id"], rs.Primary.ID)
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
	tfeClient := testAccProvider.Meta().(*tfe.Client)

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

		_, err := tfeClient.WorkspaceRunTasks.Read(ctx, rs.Primary.Attributes["workspace_id"], rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Workspace Run Tasks %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccTFEWorkspaceRunTask_basic(orgName, runTaskURL string) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "%s"
  email = "admin@company.com"
}

resource "tfe_organization_run_task" "foobar" {
  organization = tfe_organization.foobar.id
  url          = "%s"
  name         = "foobar-task"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_workspace_run_task" "foobar" {
  workspace_id      = resource.tfe_workspace.foobar.id
  task_id           = resource.tfe_organization_run_task.foobar.id
  enforcement_level = "advisory"
}
`, orgName, runTaskURL)
}

func testAccTFEWorkspaceRunTask_update(orgName, runTaskURL string) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "%s"
  email = "admin@company.com"
}

resource "tfe_organization_run_task" "foobar" {
  organization = tfe_organization.foobar.id
  url          = "%s"
  name         = "foobar-task"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_workspace_run_task" "foobar" {
  workspace_id      = resource.tfe_workspace.foobar.id
  task_id           = resource.tfe_organization_run_task.foobar.id
  enforcement_level = "mandatory"
}
`, orgName, runTaskURL)
}

func testAccTFEWorkspaceRunTask_beta_basic(orgName, runTaskURL string) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "%s"
  email = "admin@company.com"
}

resource "tfe_organization_run_task" "foobar" {
  organization = tfe_organization.foobar.id
  url          = "%s"
  name         = "foobar-task"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_workspace_run_task" "foobar" {
  workspace_id      = resource.tfe_workspace.foobar.id
  task_id           = resource.tfe_organization_run_task.foobar.id
  enforcement_level = "advisory"
  stage             = "post_plan"
}
`, orgName, runTaskURL)
}

func testAccTFEWorkspaceRunTask_beta_update(orgName, runTaskURL string) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "%s"
  email = "admin@company.com"
}

resource "tfe_organization_run_task" "foobar" {
  organization = tfe_organization.foobar.id
  url          = "%s"
  name         = "foobar-task"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_workspace_run_task" "foobar" {
  workspace_id      = resource.tfe_workspace.foobar.id
  task_id           = resource.tfe_organization_run_task.foobar.id
  enforcement_level = "mandatory"
  stage             = "pre_plan"
}
`, orgName, runTaskURL)
}
