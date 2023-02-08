package tfe

import (
	"fmt"
	"math/rand"
	"regexp"
	"testing"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccTFEWorkspaceRun_createWithDefaultParams(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	organization, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	parentWorkspace, childWorkspace := setupWorkspacesWithConfig(t, tfeClient, rInt, organization.Name, "test-fixtures/basic-config")
	runForParentWorkspace := &tfe.Run{}
	runForChildWorkspace := &tfe.Run{}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEWorkspaceRun_createWithDefaults(organization.Name, parentWorkspace.Name, childWorkspace.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceRunExistWithExpectedStatus("tfe_workspace_run.ws_run_parent", runForParentWorkspace, tfe.RunApplied),
					testAccCheckTFEWorkspaceRunExistWithExpectedStatus("tfe_workspace_run.ws_run_child", runForChildWorkspace, tfe.RunApplied),
					resource.TestCheckResourceAttrWith("tfe_workspace_run.ws_run_parent", "id", func(value string) error {
						if value != runForParentWorkspace.ID {
							return fmt.Errorf("run ID for ws_run_parent should be %s but was %s", runForParentWorkspace.ID, value)
						}
						return nil
					}),
					resource.TestCheckResourceAttrWith("tfe_workspace_run.ws_run_child", "id", func(value string) error {
						if value != runForChildWorkspace.ID {
							return fmt.Errorf("run ID for ws_run_child should be %s but was %s", runForChildWorkspace.ID, value)
						}
						return nil
					}),
				),
			},
		},
	})
}

func TestAccTFEWorkspaceRun_createAndDestroyRuns(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	parentWorkspace, childWorkspace := setupWorkspacesWithConfig(t, tfeClient, rInt, org.Name, "test-fixtures/basic-config")

	runForParentWorkspace := &tfe.Run{}
	runForChildWorkspace := &tfe.Run{}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckTFEWorkspaceRunDestroy(parentWorkspace.ID),
			testAccCheckTFEWorkspaceRunDestroy(childWorkspace.ID),
		),
		Steps: []resource.TestStep{
			{
				Config: testAccTFEWorkspaceRun_createAndDestroyRuns(org.Name, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceRunExistWithExpectedStatus("tfe_workspace_run.ws_run_parent", runForParentWorkspace, tfe.RunApplied),
					testAccCheckTFEWorkspaceRunExistWithExpectedStatus("tfe_workspace_run.ws_run_child", runForChildWorkspace, tfe.RunApplied),
					resource.TestCheckResourceAttrWith("tfe_workspace_run.ws_run_parent", "id", func(value string) error {
						if value != runForParentWorkspace.ID {
							return fmt.Errorf("run ID for ws_run_parent should be %s but was %s", runForParentWorkspace.ID, value)
						}
						return nil
					}),
					resource.TestCheckResourceAttrWith("tfe_workspace_run.ws_run_child", "id", func(value string) error {
						if value != runForChildWorkspace.ID {
							return fmt.Errorf("run ID for ws_run_child should be %s but was %s", runForChildWorkspace.ID, value)
						}
						return nil
					}),
				),
			},
		},
	})
}

func TestAccTFEWorkspaceRun_invalidParams(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	organization, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	invalidCases := []struct {
		Config      string
		ExpectError *regexp.Regexp
	}{
		{
			Config:      testAccTFEWorkspaceRun_noApplyOrDestroyBlockProvided(organization.Name, rInt),
			ExpectError: regexp.MustCompile("\"apply\": one of `apply,destroy` must be specified"),
		},
		{
			Config:      testAccTFEWorkspaceRun_noWorkspaceProvided(organization.Name),
			ExpectError: regexp.MustCompile(`The argument "workspace" is required, but no definition was found`),
		},
	}

	for _, invalidCase := range invalidCases {
		resource.Test(t, resource.TestCase{
			PreCheck: func() {
				testAccPreCheck(t)
			},
			Providers: testAccProviders,
			Steps: []resource.TestStep{
				{
					Config:      invalidCase.Config,
					ExpectError: invalidCase.ExpectError,
				},
			},
		})
	}
}

func TestAccTFEWorkspaceRun_WhenRunErrors(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	parentWorkspace, _ := setupWorkspacesWithConfig(t, tfeClient, rInt, org.Name, "test-fixtures/config-with-error-during-plan")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFEWorkspaceRun_WhenRunErrors(org.Name, parentWorkspace.Name),
				ExpectError: regexp.MustCompile(`run errored during plan`),
			},
		},
	})
}

func setupWorkspacesWithConfig(t *testing.T, tfeClient *tfe.Client, rInt int, orgName string, configPath string) (*tfe.Workspace, *tfe.Workspace) {
	parentWorkspace := &tfe.Workspace{}
	childWorkspace := &tfe.Workspace{}

	// create workspace outside of the config, to allow for testing destroy runs prior to deleting the workspace
	for _, wsName := range []string{fmt.Sprintf("tst-terraform-%d-parent", rInt), fmt.Sprintf("tst-terraform-%d-child", rInt)} {
		ws, err := tfeClient.Workspaces.Create(ctx, orgName, tfe.WorkspaceCreateOptions{Name: tfe.String(wsName)})
		if err != nil {
			t.Fatal(err)
		}
		_ = createAndUploadConfigurationVersion(t, ws, tfeClient, configPath)

		if wsName == fmt.Sprintf("tst-terraform-%d-parent", rInt) {
			*parentWorkspace = *ws
		} else {
			*childWorkspace = *ws
		}
	}

	t.Cleanup(func() {
		if err := tfeClient.Workspaces.DeleteByID(ctx, parentWorkspace.ID); err != nil {
			t.Errorf("Error destroying Workspace! WARNING: Dangling resources\n"+
				"may exist! The full error is show below:\n\n"+
				"Workspace:%s\nError: %s", parentWorkspace.ID, err)
		}
	})
	t.Cleanup(func() {
		if err := tfeClient.Workspaces.DeleteByID(ctx, childWorkspace.ID); err != nil {
			t.Errorf("Error destroying Workspace! WARNING: Dangling resources\n"+
				"may exist! The full error is show below:\n\n"+
				"Workspace:%s\nError: %s", childWorkspace.ID, err)
		}
	})

	return parentWorkspace, childWorkspace
}

func testAccCheckTFEWorkspaceRunExistWithExpectedStatus(n string, run *tfe.Run, expectedStatus tfe.RunStatus) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(ConfiguredClient)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		runData, err := config.Client.Runs.Read(ctx, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Unable to read run, %w", err)
		}

		if runData == nil {
			return fmt.Errorf("Run not found")
		}

		if runData.Status != expectedStatus {
			return fmt.Errorf("Expected run status to be %s, but got %s", expectedStatus, runData.Status)
		}

		if runData.IsDestroy {
			return fmt.Errorf("Expected run to create resources")
		}

		*run = *runData

		return nil
	}
}

func testAccCheckTFEWorkspaceRunDestroy(workspaceID string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(ConfiguredClient)

		runList, err := config.Client.Runs.List(ctx, workspaceID, &tfe.RunListOptions{
			Operation: "destroy",
			Status:    string(tfe.RunApplied),
		})
		if err != nil {
			return fmt.Errorf("Unable to find destroy run, %w", err)
		}

		if runList.TotalCount == 0 {
			return fmt.Errorf("Destroy run not found")
		}

		if len(runList.Items) > 1 {
			return fmt.Errorf("Expected only 1 destroy run but found %d", runList.TotalCount)
		}

		return nil
	}
}

func testAccTFEWorkspaceRun_createWithDefaults(orgName string, parentWorkspaceName string, childWorkspaceName string) string {
	return fmt.Sprintf(`
	resource "tfe_workspace_run" "ws_run_parent" {
		organization = "%s"
		workspace    = "%s"

		apply {}

		destroy {}
	}

	resource "tfe_workspace_run" "ws_run_child" {
		organization = "%s"
		workspace    = "%s"
		depends_on   = [tfe_workspace_run.ws_run_parent]

		apply {}

		destroy {}
	}`, orgName, parentWorkspaceName, orgName, childWorkspaceName)
}

func testAccTFEWorkspaceRun_createAndDestroyRuns(orgName string, rInt int) string {
	return fmt.Sprintf(`
	data "tfe_workspace" "parent" {
		name                 = "tst-terraform-%d-parent"
		organization         = "%s"
	}

	data "tfe_workspace" "child_depends_on_parent" {
		name                 = "tst-terraform-%d-child"
		organization         = "%s"
	}

	resource "tfe_workspace_run" "ws_run_parent" {
		organization = "%s"
		workspace    = data.tfe_workspace.parent.name

		apply {
			manual_confirm = false
			retry = true
		}

		destroy {
			manual_confirm = false
			retry = true
		}
	}

	resource "tfe_workspace_run" "ws_run_child" {
		organization = "%s"
		workspace    = data.tfe_workspace.child_depends_on_parent.name
		depends_on   = [tfe_workspace_run.ws_run_parent]

		apply {
			manual_confirm = false
			retry = true
		}

		destroy {
			manual_confirm = false
			retry = true
		}
	}`, rInt, orgName, rInt, orgName, orgName, orgName)
}

func testAccTFEWorkspaceRun_noApplyOrDestroyBlockProvided(orgName string, rInt int) string {
	return fmt.Sprintf(`
	resource "tfe_workspace" "parent" {
		name                 = "tst-terraform-%d-parent"
		organization         = "%s"
	}

	resource "tfe_workspace_run" "ws_run_parent" {
		organization = "%s"
		workspace    = tfe_workspace.parent.name
	}
`, rInt, orgName, orgName)
}

func testAccTFEWorkspaceRun_noOrganizationProvided(orgName string, rInt int) string {
	return fmt.Sprintf(`
	resource "tfe_workspace" "parent" {
		name                 = "tst-terraform-%d-parent"
		organization         = "%s"
	}

	resource "tfe_workspace_run" "ws_run_parent" {
		workspace    = tfe_workspace.parent.name

		apply {
			manual_confirm = false
			retry = true
		}

		destroy {
			manual_confirm = false
			retry = true
		}
	}
`, rInt, orgName)
}

func testAccTFEWorkspaceRun_noWorkspaceProvided(orgName string) string {
	return fmt.Sprintf(`
	resource "tfe_workspace_run" "ws_run_parent" {
		organization = "%s"

		apply {
			manual_confirm = false
			retry = true
		}

		destroy {
			manual_confirm = false
			retry = true
		}
	}
`, orgName)
}

func testAccTFEWorkspaceRun_WhenRunErrors(orgName string, workspaceName string) string {
	return fmt.Sprintf(`
	resource "tfe_workspace_run" "ws_run_parent" {
		organization = "%s"
		workspace    = "%s"

		apply {
			manual_confirm = false
			retry = false
		}

		destroy {
			manual_confirm = false
			retry = false
		}
	}
`, orgName, workspaceName)
}
