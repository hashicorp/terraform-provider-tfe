package tfe

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"testing"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccTFEWorkspace_basic(t *testing.T) {
	workspace := &tfe.Workspace{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccTFEWorkspace_basic, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					testAccCheckTFEWorkspaceAttributes(workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "name", "workspace-test"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "auto_apply", "true"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "file_triggers_enabled", "true"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "operations", "true"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "queue_all_runs", "true"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "speculative_enabled", "true"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "trigger_prefixes.#", "0"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "working_directory", ""),
				),
			},
		},
	})
}

func TestAccTFEWorkspace_panic(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config:             fmt.Sprintf(testAccTFEWorkspace_basic, rInt),
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", &tfe.Workspace{}),
					testAccCheckTFEWorkspacePanic("tfe_workspace.foobar"),
				),
			},
		},
	})
}

func TestAccTFEWorkspace_monorepo(t *testing.T) {
	workspace := &tfe.Workspace{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccTFEWorkspace_monorepo, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					testAccCheckTFEWorkspaceMonorepoAttributes(workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "name", "workspace-monorepo"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "file_triggers_enabled", "true"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "operations", "true"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "trigger_prefixes.#", "2"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "trigger_prefixes.0", "/modules"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "trigger_prefixes.1", "/shared"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "working_directory", "/db"),
				),
			},
		},
	})
}

func TestAccTFEWorkspace_renamed(t *testing.T) {
	workspace := &tfe.Workspace{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccTFEWorkspace_basic, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					testAccCheckTFEWorkspaceAttributes(workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "name", "workspace-test"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "auto_apply", "true"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "operations", "true"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "queue_all_runs", "true"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "working_directory", ""),
				),
			},

			{
				PreConfig: testAccCheckTFEWorkspaceRename(fmt.Sprintf("tst-terraform-%d", rInt)),
				Config:    fmt.Sprintf(testAccTFEWorkspace_renamed, rInt),
				PlanOnly:  true,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					testAccCheckTFEWorkspaceAttributes(workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "name", "workspace-test"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "auto_apply", "true"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "operations", "true"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "queue_all_runs", "true"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "working_directory", ""),
				),
			},
		},
	})
}

func TestAccTFEWorkspace_update(t *testing.T) {
	workspace := &tfe.Workspace{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccTFEWorkspace_basic, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					testAccCheckTFEWorkspaceAttributes(workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "name", "workspace-test"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "auto_apply", "true"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "operations", "true"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "queue_all_runs", "true"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "working_directory", ""),
				),
			},
			{
				Config: fmt.Sprintf(testAccTFEWorkspace_update, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					testAccCheckTFEWorkspaceAttributesUpdated(workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "name", "workspace-updated"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "auto_apply", "false"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "file_triggers_enabled", "true"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "operations", "false"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "queue_all_runs", "false"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "terraform_version", "0.11.1"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "trigger_prefixes.#", "2"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "trigger_prefixes.0", "/modules"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "trigger_prefixes.1", "/shared"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "working_directory", "terraform/test"),
				),
			},
		},
	})
}

func TestAccTFEWorkspace_updateWorkingDirectory(t *testing.T) {
	workspace := &tfe.Workspace{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccTFEWorkspace_basic, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					testAccCheckTFEWorkspaceAttributes(workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "name", "workspace-test"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "auto_apply", "true"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "operations", "true"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "queue_all_runs", "true"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "working_directory", ""),
				),
			},
			{
				Config: fmt.Sprintf(testAccTFEWorkspace_updateAddWorkingDirectory, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					testAccCheckTFEWorkspaceAttributesUpdatedAddWorkingDirectory(workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "name", "workspace-updated"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "working_directory", "terraform/test"),
				),
			},
			{
				Config: fmt.Sprintf(testAccTFEWorkspace_updateRemoveWorkingDirectory, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					testAccCheckTFEWorkspaceAttributesUpdatedRemoveWorkingDirectory(workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "name", "workspace-updated"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "working_directory", ""),
				),
			},
		},
	})
}

func TestAccTFEWorkspace_updateFileTriggers(t *testing.T) {
	workspace := &tfe.Workspace{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccTFEWorkspace_basic, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "file_triggers_enabled", "true"),
				),
			},

			{
				Config: fmt.Sprintf(testAccTFEWorkspace_basicFileTriggersOff, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "file_triggers_enabled", "false"),
				),
			},
		},
	})
}

func TestAccTFEWorkspace_updateTriggerPrefixes(t *testing.T) {
	workspace := &tfe.Workspace{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccTFEWorkspace_triggerPrefixes, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "trigger_prefixes.#", "2"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "trigger_prefixes.0", "/modules"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "trigger_prefixes.1", "/shared"),
				),
			},

			{
				Config: fmt.Sprintf(testAccTFEWorkspace_updateEmptyTriggerPrefixes, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					testAccCheckTFEWorkspaceAttributes(workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "trigger_prefixes.#", "0"),
				),
			},
		},
	})
}

func TestAccTFEWorkspace_updateSpeculative(t *testing.T) {
	workspace := &tfe.Workspace{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccTFEWorkspace_basic, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "speculative_enabled", "true"),
				),
			},

			{
				Config: fmt.Sprintf(testAccTFEWorkspace_basicSpeculativeOff, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "speculative_enabled", "false"),
				),
			},
		},
	})
}

func TestAccTFEWorkspace_updateVCSRepo(t *testing.T) {
	workspace := &tfe.Workspace{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			if GITHUB_TOKEN == "" {
				t.Skip("Please set GITHUB_TOKEN to run this test")
			}
			if GITHUB_WORKSPACE_IDENTIFIER == "" {
				t.Skip("Please set GITHUB_WORKSPACE_IDENTIFIER to run this test")
			}
			if GITHUB_WORKSPACE_BRANCH == "" {
				t.Skip("Please set GITHUB_WORKSPACE_BRANCH to run this test")
			}
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccTFEWorkspace_basic, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					testAccCheckTFEWorkspaceAttributes(workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "name", "workspace-test"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "auto_apply", "true"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "operations", "true"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "queue_all_runs", "true"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "working_directory", ""),
				),
			},
			{
				Config: testAccTFEWorkspace_updateAddVCSRepo(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists("tfe_workspace.foobar", workspace),
					testAccCheckTFEWorkspaceUpdatedAddVCSRepoAttributes(workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "name", "workspace-test-add-vcs-repo"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "vcs_repo.0.identifier", GITHUB_WORKSPACE_IDENTIFIER),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "vcs_repo.0.branch", ""),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "vcs_repo.0.ingress_submodules", "false"),
				),
			},
			{
				Config: testAccTFEWorkspace_updateUpdateVCSRepoBranch(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists("tfe_workspace.foobar", workspace),
					testAccCheckTFEWorkspaceUpdatedUpdateVCSRepoBranchAttributes(workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "name", "workspace-test-update-vcs-repo-branch"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "vcs_repo.0.identifier", GITHUB_WORKSPACE_IDENTIFIER),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "vcs_repo.0.branch", GITHUB_WORKSPACE_BRANCH),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "vcs_repo.0.ingress_submodules", "false"),
				),
			},
			{
				Config: fmt.Sprintf(testAccTFEWorkspace_updateRemoveVCSRepo, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists("tfe_workspace.foobar", workspace),
					testAccCheckTFEWorkspaceUpdatedRemoveVCSRepoAttributes(workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "name", "workspace-test-remove-vcs-repo"),
					resource.TestCheckNoResourceAttr("tfe_workspace.foobar", "vcs_repo"),
				),
			},
		},
	})
}

func TestAccTFEWorkspace_sshKey(t *testing.T) {
	workspace := &tfe.Workspace{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccTFEWorkspace_basic, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					testAccCheckTFEWorkspaceAttributes(workspace),
				),
			},

			{
				Config: fmt.Sprintf(testAccTFEWorkspace_sshKey, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					testAccCheckTFEWorkspaceAttributesSSHKey(workspace),
					resource.TestCheckResourceAttrSet(
						"tfe_workspace.foobar", "ssh_key_id"),
				),
			},

			{
				Config: fmt.Sprintf(testAccTFEWorkspace_noSSHKey, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					testAccCheckTFEWorkspaceAttributes(workspace),
				),
			},
		},
	})
}

func TestAccTFEWorkspace_import(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccTFEWorkspace_basic, rInt),
			},

			{
				ResourceName:      "tfe_workspace.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTFEWorkspace_importVCSBranch(t *testing.T) {
	workspace := &tfe.Workspace{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			if GITHUB_TOKEN == "" {
				t.Skip("Please set GITHUB_TOKEN to run this test")
			}
			if GITHUB_WORKSPACE_IDENTIFIER == "" {
				t.Skip("Please set GITHUB_WORKSPACE_IDENTIFIER to run this test")
			}
			if GITHUB_WORKSPACE_BRANCH == "" {
				t.Skip("Please set GITHUB_WORKSPACE_BRANCH to run this test")
			}

		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEWorkspace_updateUpdateVCSRepoBranch(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists("tfe_workspace.foobar", workspace),
					testAccCheckTFEWorkspaceUpdatedUpdateVCSRepoBranchAttributes(workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "name", "workspace-test-update-vcs-repo-branch"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "vcs_repo.0.identifier", GITHUB_WORKSPACE_IDENTIFIER),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "vcs_repo.0.branch", GITHUB_WORKSPACE_BRANCH),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "vcs_repo.0.ingress_submodules", "false"),
				),
			},

			{
				ResourceName:      "tfe_workspace.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckTFEWorkspaceExists(
	n string, workspace *tfe.Workspace) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		tfeClient := testAccProvider.Meta().(*tfe.Client)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		// Get the workspace
		w, err := tfeClient.Workspaces.ReadByID(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}

		*workspace = *w

		return nil
	}
}

// As of 4/20/2020 there is a bug that will cause the provider to panic
// if a workspace is deleted outside of terraform. This case is handled
// by the data_workspace but not resource_workspace.
//
// This test demonstrates the bug.
//
// panic: runtime error: invalid memory address or nil pointer dereference
// resource_tfe_workspace.go:208 resourceTFEWorkspaceRead(...)
func testAccCheckTFEWorkspacePanic(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		tfeClient := testAccProvider.Meta().(*tfe.Client)

		// Grab the resource out of the state and delete it from TFC/E directly.
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		err := tfeClient.Workspaces.DeleteByID(ctx, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Could not delete %s: %v", n, err)
		}

		// Read the workspace again using the lower level resource reader
		// which will trigger the panic
		rd := &schema.ResourceData{}
		rd.SetId(rs.Primary.ID)

		err = resourceTFEWorkspaceRead(rd, testAccProvider.Meta())
		if err != nil && err != tfe.ErrResourceNotFound {
			return fmt.Errorf("Could not re-read resource directly: %v", err)
		}

		return nil
	}
}

func testAccCheckTFEWorkspaceAttributes(
	workspace *tfe.Workspace) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if workspace.Name != "workspace-test" {
			return fmt.Errorf("Bad name: %s", workspace.Name)
		}

		if workspace.AutoApply != true {
			return fmt.Errorf("Bad auto apply: %t", workspace.AutoApply)
		}

		if workspace.Operations != true {
			return fmt.Errorf("Bad operations: %t", workspace.Operations)
		}

		if workspace.QueueAllRuns != true {
			return fmt.Errorf("Bad queue all runs: %t", workspace.QueueAllRuns)
		}

		if workspace.SSHKey != nil {
			return fmt.Errorf("Bad SSH key: %v", workspace.SSHKey)
		}

		if workspace.WorkingDirectory != "" {
			return fmt.Errorf("Bad working directory: %s", workspace.WorkingDirectory)
		}

		if len(workspace.TriggerPrefixes) != 0 {
			return fmt.Errorf("Bad trigger prefixes: %s", workspace.TriggerPrefixes)
		}

		return nil
	}
}

func testAccCheckTFEWorkspaceMonorepoAttributes(
	workspace *tfe.Workspace) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if workspace.Name != "workspace-monorepo" {
			return fmt.Errorf("Bad name: %s", workspace.Name)
		}

		if workspace.FileTriggersEnabled != true {
			return fmt.Errorf("Bad file triggers enabled: %t", workspace.FileTriggersEnabled)
		}

		triggerPrefixes := []string{"/modules", "/shared"}
		if len(workspace.TriggerPrefixes) != len(triggerPrefixes) {
			return fmt.Errorf("Bad trigger prefixes length: %d", len(workspace.TriggerPrefixes))
		}
		for i := range triggerPrefixes {
			if workspace.TriggerPrefixes[i] != triggerPrefixes[i] {
				return fmt.Errorf("Bad trigger prefixes %v", workspace.TriggerPrefixes)
			}
		}

		if workspace.WorkingDirectory != "/db" {
			return fmt.Errorf("Bad working directory: %s", workspace.WorkingDirectory)
		}

		return nil
	}
}

func testAccCheckTFEWorkspaceRename(orgName string) func() {
	return func() {
		tfeClient := testAccProvider.Meta().(*tfe.Client)

		w, err := tfeClient.Workspaces.Update(
			context.Background(),
			orgName,
			"workspace-test",
			tfe.WorkspaceUpdateOptions{Name: tfe.String("renamed-out-of-band")},
		)
		if err != nil {
			log.Fatalf("Could not rename the workspace out of band: %v", err)
		}

		if w.Name != "renamed-out-of-band" {
			log.Fatalf("Failed to rename the workspace out of band: %v", err)
		}
	}
}

func testAccCheckTFEWorkspaceAttributesUpdated(
	workspace *tfe.Workspace) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if workspace.Name != "workspace-updated" {
			return fmt.Errorf("Bad name: %s", workspace.Name)
		}

		if workspace.AutoApply != false {
			return fmt.Errorf("Bad auto apply: %t", workspace.AutoApply)
		}

		if workspace.Operations != false {
			return fmt.Errorf("Bad operations: %t", workspace.Operations)
		}

		if workspace.QueueAllRuns != false {
			return fmt.Errorf("Bad queue all runs: %t", workspace.QueueAllRuns)
		}

		if workspace.TerraformVersion != "0.11.1" {
			return fmt.Errorf("Bad Terraform version: %s", workspace.TerraformVersion)
		}

		if workspace.WorkingDirectory != "terraform/test" {
			return fmt.Errorf("Bad working directory: %s", workspace.WorkingDirectory)
		}

		return nil
	}
}

func testAccCheckTFEWorkspaceAttributesUpdatedAddWorkingDirectory(
	workspace *tfe.Workspace) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if workspace.Name != "workspace-updated" {
			return fmt.Errorf("Bad name: %s", workspace.Name)
		}

		if workspace.WorkingDirectory != "terraform/test" {
			return fmt.Errorf("Today Bad working directory: %s", workspace.WorkingDirectory)
		}

		return nil
	}
}

func testAccCheckTFEWorkspaceAttributesUpdatedRemoveWorkingDirectory(
	workspace *tfe.Workspace) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if workspace.Name != "workspace-updated" {
			return fmt.Errorf("Bad name: %s", workspace.Name)
		}

		if workspace.WorkingDirectory != "" {
			return fmt.Errorf("Today Bad working directory: %s", workspace.WorkingDirectory)
		}

		return nil
	}
}

func testAccCheckTFEWorkspaceAttributesSSHKey(
	workspace *tfe.Workspace) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if workspace.SSHKey == nil {
			return fmt.Errorf("Bad SSH key: %v", workspace.SSHKey)
		}

		return nil
	}
}

func testAccCheckTFEWorkspaceUpdatedAddVCSRepoAttributes(
	workspace *tfe.Workspace) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if workspace.Name != "workspace-test-add-vcs-repo" {
			return fmt.Errorf("Bad name: %s", workspace.Name)
		}

		if workspace.VCSRepo == nil {
			return fmt.Errorf("Bad VCS repo: %v", workspace.VCSRepo)
		}

		if workspace.VCSRepo.Branch != "" {
			return fmt.Errorf("Bad VCS repo branch: %v", workspace.VCSRepo.Branch)
		}

		if workspace.VCSRepo.Identifier != GITHUB_WORKSPACE_IDENTIFIER {
			return fmt.Errorf("Bad VCS repo identifier: %v", workspace.VCSRepo.Identifier)
		}

		if workspace.VCSRepo.IngressSubmodules != false {
			return fmt.Errorf("Bad VCS repo ingress submodules: %v", workspace.VCSRepo.IngressSubmodules)
		}

		return nil
	}
}

func testAccCheckTFEWorkspaceUpdatedUpdateVCSRepoBranchAttributes(
	workspace *tfe.Workspace) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if workspace.Name != "workspace-test-update-vcs-repo-branch" {
			return fmt.Errorf("Bad name: %s", workspace.Name)
		}

		if workspace.VCSRepo == nil {
			return fmt.Errorf("Bad VCS repo: %v", workspace.VCSRepo)
		}

		if workspace.VCSRepo.Branch != GITHUB_WORKSPACE_BRANCH {
			return fmt.Errorf("Bad VCS repo branch: %v", workspace.VCSRepo.Branch)
		}

		if workspace.VCSRepo.Identifier != GITHUB_WORKSPACE_IDENTIFIER {
			return fmt.Errorf("Bad VCS repo identifier: %v", workspace.VCSRepo.Identifier)
		}

		if workspace.VCSRepo.IngressSubmodules != false {
			return fmt.Errorf("Bad VCS repo ingress submodules: %v", workspace.VCSRepo.IngressSubmodules)
		}

		return nil
	}
}

func testAccCheckTFEWorkspaceUpdatedRemoveVCSRepoAttributes(
	workspace *tfe.Workspace) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if workspace.Name != "workspace-test-remove-vcs-repo" {
			return fmt.Errorf("Bad name: %s", workspace.Name)
		}

		if workspace.VCSRepo != nil {
			return fmt.Errorf("Bad VCS repo: %v", workspace.VCSRepo)
		}

		return nil
	}
}

func testAccCheckTFEWorkspaceDestroy(s *terraform.State) error {
	tfeClient := testAccProvider.Meta().(*tfe.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_workspace" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := tfeClient.Workspaces.ReadByID(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Workspace %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

const testAccTFEWorkspace_basic = `
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = "${tfe_organization.foobar.id}"
  auto_apply   = true
}`

const testAccTFEWorkspace_basicFileTriggersOff = `
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name                  = "workspace-test"
  organization          = "${tfe_organization.foobar.id}"
  auto_apply            = true
  file_triggers_enabled = false
}`

const testAccTFEWorkspace_basicSpeculativeOff = `
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name                  = "workspace-test"
  organization          = "${tfe_organization.foobar.id}"
  auto_apply            = true
  speculative_enabled = false
}`

const testAccTFEWorkspace_monorepo = `
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name                  = "workspace-monorepo"
  organization          = "${tfe_organization.foobar.id}"
  file_triggers_enabled = true
  trigger_prefixes      = ["/modules", "/shared"]
  working_directory     = "/db"
}`

const testAccTFEWorkspace_renamed = `
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name         = "renamed-out-of-band"
  organization = "${tfe_organization.foobar.id}"
  auto_apply   = true
}`

const testAccTFEWorkspace_update = `
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name                  = "workspace-updated"
  organization          = "${tfe_organization.foobar.id}"
  auto_apply            = false
  file_triggers_enabled = true
  queue_all_runs        = false
  terraform_version     = "0.11.1"
  trigger_prefixes      = ["/modules", "/shared"]
  working_directory     = "terraform/test"
  operations            = false
}`

const testAccTFEWorkspace_updateAddWorkingDirectory = `
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name                  = "workspace-updated"
  organization          = "${tfe_organization.foobar.id}"
  auto_apply            = false
  working_directory     = "terraform/test"
}`

const testAccTFEWorkspace_updateRemoveWorkingDirectory = `
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name                  = "workspace-updated"
  organization          = "${tfe_organization.foobar.id}"
  auto_apply            = false
}`

const testAccTFEWorkspace_sshKey = `
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_ssh_key" "foobar" {
  name         = "ssh-key-test"
  organization = "${tfe_organization.foobar.id}"
  key          = "SSH-KEY-CONTENT"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = "${tfe_organization.foobar.id}"
  auto_apply   = true
  ssh_key_id   = "${tfe_ssh_key.foobar.id}"
}`

const testAccTFEWorkspace_noSSHKey = `
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_ssh_key" "foobar" {
  name         = "ssh-key-test"
  organization = "${tfe_organization.foobar.id}"
  key          = "SSH-KEY-CONTENT"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = "${tfe_organization.foobar.id}"
  auto_apply   = true
}`

const testAccTFEWorkspace_triggerPrefixes = `
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name                  = "workspace"
  organization          = "${tfe_organization.foobar.id}"
  trigger_prefixes      = ["/modules", "/shared"]
}`

const testAccTFEWorkspace_updateEmptyTriggerPrefixes = `
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}
resource "tfe_workspace" "foobar" {
  name                  = "workspace-test"
  organization          = "${tfe_organization.foobar.id}"
  auto_apply            = true
}`

func testAccTFEWorkspace_updateAddVCSRepo(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_oauth_client" "test" {
  organization     = "${tfe_organization.foobar.id}"
  api_url          = "https://api.github.com"
  http_url         = "https://github.com"
  oauth_token      = "%s"
  service_provider = "github"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test-add-vcs-repo"
  organization = "${tfe_organization.foobar.id}"
  auto_apply   = true
  vcs_repo {
    identifier     = "%s"
    oauth_token_id = "${tfe_oauth_client.test.oauth_token_id}"
  }
}
`,
		rInt,
		GITHUB_TOKEN,
		GITHUB_WORKSPACE_IDENTIFIER,
	)
}

func testAccTFEWorkspace_updateUpdateVCSRepoBranch(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_oauth_client" "test" {
  organization     = "${tfe_organization.foobar.id}"
  api_url          = "https://api.github.com"
  http_url         = "https://github.com"
  oauth_token      = "%s"
  service_provider = "github"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test-update-vcs-repo-branch"
  organization = "${tfe_organization.foobar.id}"
  auto_apply   = true
  vcs_repo {
    identifier     = "%s"
    oauth_token_id = "${tfe_oauth_client.test.oauth_token_id}"
    branch         = "%s"
  }
}
`,
		rInt,
		GITHUB_TOKEN,
		GITHUB_WORKSPACE_IDENTIFIER,
		GITHUB_WORKSPACE_BRANCH,
	)
}

const testAccTFEWorkspace_updateRemoveVCSRepo = `
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test-remove-vcs-repo"
  organization = "${tfe_organization.foobar.id}"
  auto_apply   = true
}
`
