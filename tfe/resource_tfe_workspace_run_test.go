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

func TestAccTFEWorkspaceRun_create(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	organization, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	runForParentWorkspace := &tfe.Run{}
	runForChildWorkspace := &tfe.Run{}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccGithubPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEWorkspaceRun_create(organization.Name, rInt),
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

func TestAccTFEWorkspaceRun_createWithDefaults(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	organization, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	runForParentWorkspace := &tfe.Run{}
	runForChildWorkspace := &tfe.Run{}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccGithubPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEWorkspaceRun_createWithDefaults(organization.Name, rInt),
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

	oauthClient, err := tfeClient.OAuthClients.Create(ctx, org.Name, tfe.OAuthClientCreateOptions{
		APIURL:          tfe.String("https://api.github.com"),
		HTTPURL:         tfe.String("https://github.com"),
		ServiceProvider: tfe.ServiceProvider(tfe.ServiceProviderGithub),
		OAuthToken:      tfe.String(envGithubToken),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(oauthClient.OAuthTokens) == 0 {
		t.Fatal("Expected OAuthClient to have OAuthTokens, 0 found")
	}

	t.Cleanup(func() {
		if err := tfeClient.OAuthClients.Delete(ctx, oauthClient.ID); err != nil {
			t.Errorf("Error destroying Oauth client! WARNING: Dangling resources\n"+
				"may exist! The full error is show below:\n\n"+
				"Oauth client:%s\nError: %s", oauthClient.ID, err)
		}
	})

	workspaceA := &tfe.Workspace{}
	workspaceB := &tfe.Workspace{}

	// create workspace outside of the config, to allow for testing check destroy runs prior to deleting the workspace
	for _, wsName := range []string{fmt.Sprintf("tst-terraform-%d-A", rInt), fmt.Sprintf("tst-terraform-%d-B", rInt)} {
		ws, err := tfeClient.Workspaces.Create(ctx, org.Name, tfe.WorkspaceCreateOptions{
			Name:         tfe.String(wsName),
			QueueAllRuns: tfe.Bool(false),
			VCSRepo: &tfe.VCSRepoOptions{
				Branch:       tfe.String(envGithubWorkspaceBranch),
				Identifier:   tfe.String(envGithubWorkspaceIdentifier),
				OAuthTokenID: tfe.String(oauthClient.OAuthTokens[0].ID),
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		if wsName == fmt.Sprintf("tst-terraform-%d-A", rInt) {
			*workspaceA = *ws
		} else {
			*workspaceB = *ws
		}
	}

	t.Cleanup(func() {
		if err := tfeClient.Workspaces.DeleteByID(ctx, workspaceA.ID); err != nil {
			t.Errorf("Error destroying Workspace! WARNING: Dangling resources\n"+
				"may exist! The full error is show below:\n\n"+
				"Workspace:%s\nError: %s", workspaceA.ID, err)
		}
	})
	t.Cleanup(func() {
		if err := tfeClient.Workspaces.DeleteByID(ctx, workspaceB.ID); err != nil {
			t.Errorf("Error destroying Workspace! WARNING: Dangling resources\n"+
				"may exist! The full error is show below:\n\n"+
				"Workspace:%s\nError: %s", workspaceB.ID, err)
		}
	})

	runForWorkspaceRootA := &tfe.Run{}
	runForWorkspaceB := &tfe.Run{}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccGithubPreCheck(t)
		},
		Providers: testAccProviders,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckTFEWorkspaceRunDestroy(workspaceA.ID),
			testAccCheckTFEWorkspaceRunDestroy(workspaceB.ID),
		),
		Steps: []resource.TestStep{
			{
				Config: testAccTFEWorkspaceRun_createAndDestroyRuns(org.Name, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceRunExistWithExpectedStatus("tfe_workspace_run.ws_run_root_A", runForWorkspaceRootA, tfe.RunApplied),
					testAccCheckTFEWorkspaceRunExistWithExpectedStatus("tfe_workspace_run.ws_run_B", runForWorkspaceB, tfe.RunApplied),
					resource.TestCheckResourceAttrWith("tfe_workspace_run.ws_run_root_A", "id", func(value string) error {
						if value != runForWorkspaceRootA.ID {
							return fmt.Errorf("run ID for ws_run_root_A should be %s but was %s", runForWorkspaceRootA.ID, value)
						}
						return nil
					}),
					resource.TestCheckResourceAttrWith("tfe_workspace_run.ws_run_B", "id", func(value string) error {
						if value != runForWorkspaceB.ID {
							return fmt.Errorf("run ID for ws_run_B should be %s but was %s", runForWorkspaceB.ID, value)
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
			Config:      testAccTFEWorkspaceRun_noOrganizationProvided(organization.Name, rInt),
			ExpectError: regexp.MustCompile(`The argument "organization" is required, but no definition was found`),
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
				testAccGithubPreCheck(t)
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

	workspace, err := tfeClient.Workspaces.Create(ctx, org.Name, tfe.WorkspaceCreateOptions{Name: tfe.String(fmt.Sprintf("tst-terraform-%d-A", rInt))})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := tfeClient.Workspaces.DeleteByID(ctx, workspace.ID); err != nil {
			t.Errorf("Error destroying Workspace! WARNING: Dangling resources\n"+
				"may exist! The full error is show below:\n\n"+
				"Workspace:%s\nError: %s", workspace.ID, err)
		}
	})

	_ = createAndUploadConfigurationVersion(t, workspace, tfeClient, "test-fixtures/config-with-error-during-plan")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccGithubPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFEWorkspaceRun_WhenRunErrors(org.Name, workspace.Name),
				ExpectError: regexp.MustCompile(`run errored during plan`),
			},
		},
	})
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

func testAccTFEWorkspaceRun_create(orgName string, rInt int) string {
	return fmt.Sprintf(`
	resource "tfe_oauth_client" "test" {
		organization     = "%s"
		api_url          = "https://api.github.com"
		http_url         = "https://github.com"
		oauth_token      = "%s"
		service_provider = "github"
	}

	resource "tfe_workspace" "parent" {
		name                 = "tst-terraform-%d-parent"
		organization         = "%s"
		queue_all_runs       = false
		vcs_repo {
			branch             = "%s"
			identifier         = "%s"
			oauth_token_id     = tfe_oauth_client.test.oauth_token_id
		}
	}

	resource "tfe_workspace" "child" {
		name                 = "tst-terraform-%d-child"
		organization         = "%s"
		queue_all_runs       = false
		vcs_repo {
			branch             = "%s"
			identifier         = "%s"
			oauth_token_id     = tfe_oauth_client.test.oauth_token_id
		}
	}

	resource "tfe_workspace_run" "ws_run_parent" {
		organization = "%s"
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

	resource "tfe_workspace_run" "ws_run_child" {
		organization = "%s"
		workspace    = tfe_workspace.child.name
		depends_on   = [tfe_workspace_run.ws_run_parent]

		apply {
			manual_confirm = false
			retry = true
		}

		destroy {
			manual_confirm = false
			retry = true
		}
	}`, orgName, envGithubToken, rInt, orgName, envGithubWorkspaceBranch, envGithubWorkspaceIdentifier, rInt, orgName, envGithubWorkspaceBranch, envGithubWorkspaceIdentifier, orgName, orgName)
}

func testAccTFEWorkspaceRun_createWithDefaults(orgName string, rInt int) string {
	return fmt.Sprintf(`
	resource "tfe_oauth_client" "test" {
		organization     = "%s"
		api_url          = "https://api.github.com"
		http_url         = "https://github.com"
		oauth_token      = "%s"
		service_provider = "github"
	}

	resource "tfe_workspace" "parent" {
		name                 = "tst-terraform-%d-parent"
		organization         = "%s"
		queue_all_runs       = false
		vcs_repo {
			branch             = "%s"
			identifier         = "%s"
			oauth_token_id     = tfe_oauth_client.test.oauth_token_id
		}
	}

	resource "tfe_workspace" "child" {
		name                 = "tst-terraform-%d-child"
		organization         = "%s"
		queue_all_runs       = false
		vcs_repo {
			branch             = "%s"
			identifier         = "%s"
			oauth_token_id     = tfe_oauth_client.test.oauth_token_id
		}
	}

	resource "tfe_workspace_run" "ws_run_parent" {
		organization = "%s"
		workspace    = tfe_workspace.parent.name

		apply {}

		destroy {}
	}

	resource "tfe_workspace_run" "ws_run_child" {
		organization = "%s"
		workspace    = tfe_workspace.child.name
		depends_on   = [tfe_workspace_run.ws_run_parent]

		apply {}

		destroy {}
	}`, orgName, envGithubToken, rInt, orgName, envGithubWorkspaceBranch, envGithubWorkspaceIdentifier, rInt, orgName, envGithubWorkspaceBranch, envGithubWorkspaceIdentifier, orgName, orgName)
}

func testAccTFEWorkspaceRun_createAndDestroyRuns(orgName string, rInt int) string {
	return fmt.Sprintf(`
	data "tfe_workspace" "root_A" {
		name                 = "tst-terraform-%d-A"
		organization         = "%s"
	}

	data "tfe_workspace" "B_depends_on_A" {
		name                 = "tst-terraform-%d-B"
		organization         = "%s"
	}

	resource "tfe_workspace_run" "ws_run_root_A" {
		organization = "%s"
		workspace    = data.tfe_workspace.root_A.name

		apply {
			manual_confirm = false
			retry = true
		}

		destroy {
			manual_confirm = false
			retry = true
		}
	}

	resource "tfe_workspace_run" "ws_run_B" {
		organization = "%s"
		workspace    = data.tfe_workspace.B_depends_on_A.name
		depends_on   = [tfe_workspace_run.ws_run_root_A]

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
	resource "tfe_oauth_client" "test" {
		organization     = "%s"
		api_url          = "https://api.github.com"
		http_url         = "https://github.com"
		oauth_token      = "%s"
		service_provider = "github"
	}

	resource "tfe_workspace" "root_A" {
		name                 = "tst-terraform-%d-A"
		organization         = "%s"
		queue_all_runs       = false
		vcs_repo {
			branch             = "%s"
			identifier         = "%s"
			oauth_token_id     = tfe_oauth_client.test.oauth_token_id
		}
	}

	resource "tfe_workspace_run" "ws_run_root_A" {
		organization = "%s"
		workspace    = tfe_workspace.root_A.name
	}
`, orgName, envGithubToken, rInt, orgName, envGithubWorkspaceBranch, envGithubWorkspaceIdentifier, orgName)
}

func testAccTFEWorkspaceRun_noOrganizationProvided(orgName string, rInt int) string {
	return fmt.Sprintf(`
	resource "tfe_oauth_client" "test" {
		organization     = "%s"
		api_url          = "https://api.github.com"
		http_url         = "https://github.com"
		oauth_token      = "%s"
		service_provider = "github"
	}

	resource "tfe_workspace" "root_A" {
		name                 = "tst-terraform-%d-A"
		organization         = "%s"
		queue_all_runs       = false
		vcs_repo {
			branch             = "%s"
			identifier         = "%s"
			oauth_token_id     = tfe_oauth_client.test.oauth_token_id
		}
	}

	resource "tfe_workspace_run" "ws_run_root_A" {
		workspace    = tfe_workspace.root_A.name

		apply {
			manual_confirm = false
			retry = true
		}

		destroy {
			manual_confirm = false
			retry = true
		}
	}
`, orgName, envGithubToken, rInt, orgName, envGithubWorkspaceBranch, envGithubWorkspaceIdentifier)
}

func testAccTFEWorkspaceRun_noWorkspaceProvided(orgName string) string {
	return fmt.Sprintf(`
	resource "tfe_workspace_run" "ws_run_root_A" {
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
	resource "tfe_workspace_run" "ws_run_root" {
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
