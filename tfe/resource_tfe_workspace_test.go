package tfe

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"regexp"
	"strconv"
	"testing"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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
				Config: testAccTFEWorkspace_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					testAccCheckTFEWorkspaceAttributes(workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "name", "workspace-test"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "description", "My favorite workspace!"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "allow_destroy_plan", "false"),
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
						"tfe_workspace.foobar", "structured_run_output_enabled", "true"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "tag_names.0", "fav"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "tag_names.1", "test"),
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
				Config:             testAccTFEWorkspace_basic(rInt),
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
				Config: testAccTFEWorkspace_monorepo(rInt),
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
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEWorkspace_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					testAccCheckTFEWorkspaceAttributes(workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "name", "workspace-test"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "description", "My favorite workspace!"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "allow_destroy_plan", "false"),
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
				PreConfig: testAccCheckTFEWorkspaceRename(orgName),
				Config:    testAccTFEWorkspace_renamed(rInt),
				PlanOnly:  true,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					testAccCheckTFEWorkspaceAttributes(workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "name", "workspace-test"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "description", "My favorite workspace!"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "allow_destroy_plan", "false"),
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
				Config: testAccTFEWorkspace_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					testAccCheckTFEWorkspaceAttributes(workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "name", "workspace-test"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "allow_destroy_plan", "false"),
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
				Config: testAccTFEWorkspace_update(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					testAccCheckTFEWorkspaceAttributesUpdated(workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "name", "workspace-updated"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "allow_destroy_plan", "true"),
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
				Config: testAccTFEWorkspace_basic(rInt),
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
				Config: testAccTFEWorkspace_updateAddWorkingDirectory(rInt),
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
				Config: testAccTFEWorkspace_updateRemoveWorkingDirectory(rInt),
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
				Config: testAccTFEWorkspace_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "file_triggers_enabled", "true"),
				),
			},

			{
				Config: testAccTFEWorkspace_basicFileTriggersOff(rInt),
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
				Config: testAccTFEWorkspace_triggerPrefixes(rInt),
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
				Config: testAccTFEWorkspace_updateEmptyTriggerPrefixes(rInt),
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

func TestAccTFEWorkspace_overwriteTriggerPatternsWithPrefixes(t *testing.T) {
	workspace := &tfe.Workspace{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEWorkspace_triggerPatterns(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "trigger_patterns.#", "2"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "trigger_prefixes.#", "0"),
				),
			},
			{
				Config: testAccTFEWorkspace_triggerPrefixes(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "trigger_prefixes.#", "2"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "trigger_prefixes.0", "/modules"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "trigger_prefixes.1", "/shared"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "trigger_patterns.#", "0"),
				),
			},
			{
				Config: testAccTFEWorkspace_updateEmptyTriggerPrefixes(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					testAccCheckTFEWorkspaceAttributes(workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "trigger_prefixes.#", "0"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "trigger_patterns.#", "0"),
				),
			},
		},
	})
}

// This test suite contains large number of tests in order to build confidence
// in the fix for https://github.com/hashicorp/terraform-provider-tfe/issues/552
// TODO: remove or trim once the fix is battle tested
func TestAccTFEWorkspace_permutation_test_suite(t *testing.T) {
	t.Run("file triggers enabled is false", func(t *testing.T) {
		t.Run("and trigger prefixes are set and empty", func(t *testing.T) {
			workspace := &tfe.Workspace{}
			rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
			resource.Test(t, resource.TestCase{
				PreCheck:     func() { testAccPreCheck(t) },
				Providers:    testAccProviders,
				CheckDestroy: testAccCheckTFEWorkspaceDestroy,
				Steps: []resource.TestStep{
					{
						Config: testAccTFEWorkspace_triggersConfigurationGenerator(
							rInt,
							true, true,
							false, "",
							true, `["/pattern1", "/pattern2"]`,
						),
						Check: resource.ComposeTestCheckFunc(
							testAccCheckTFEWorkspaceExists(
								"tfe_workspace.foobar", workspace),
							resource.TestCheckResourceAttr(
								"tfe_workspace.foobar", "trigger_patterns.#", "2"),
						),
					},
					{
						Config: testAccTFEWorkspace_triggersConfigurationGenerator(
							rInt,
							true, false,
							true, "[]",
							false, "",
						),
						Check: resource.ComposeTestCheckFunc(
							testAccCheckTFEWorkspaceExists(
								"tfe_workspace.foobar", workspace),
							resource.TestCheckResourceAttr(
								"tfe_workspace.foobar", "trigger_prefixes.#", "0"),
							resource.TestCheckResourceAttr(
								"tfe_workspace.foobar", "trigger_patterns.#", "0"),
						),
					},
				},
			})
		})
		t.Run("and trigger prefixes are populated", func(t *testing.T) {
			workspace := &tfe.Workspace{}
			rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
			resource.Test(t, resource.TestCase{
				PreCheck:     func() { testAccPreCheck(t) },
				Providers:    testAccProviders,
				CheckDestroy: testAccCheckTFEWorkspaceDestroy,
				Steps: []resource.TestStep{
					{
						Config: testAccTFEWorkspace_triggersConfigurationGenerator(
							rInt,
							true, true,
							false, "",
							true, `["/pattern1", "/pattern2"]`,
						),
						Check: resource.ComposeTestCheckFunc(
							testAccCheckTFEWorkspaceExists(
								"tfe_workspace.foobar", workspace),
							resource.TestCheckResourceAttr(
								"tfe_workspace.foobar", "trigger_patterns.#", "2"),
						),
					},
					{
						Config: testAccTFEWorkspace_triggersConfigurationGenerator(
							rInt,
							true, false,
							true, `["/prefix1"]`,
							false, "",
						),
						Check: resource.ComposeTestCheckFunc(
							testAccCheckTFEWorkspaceExists(
								"tfe_workspace.foobar", workspace),
							resource.TestCheckResourceAttr(
								"tfe_workspace.foobar", "trigger_prefixes.#", "1"),
							resource.TestCheckResourceAttr(
								"tfe_workspace.foobar", "trigger_prefixes.0", "/prefix1"),
							resource.TestCheckResourceAttr(
								"tfe_workspace.foobar", "trigger_patterns.#", "0"),
						),
					},
				},
			})
		})
		t.Run("and trigger patterns are set and empty", func(t *testing.T) {
			workspace := &tfe.Workspace{}
			rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
			resource.Test(t, resource.TestCase{
				PreCheck:     func() { testAccPreCheck(t) },
				Providers:    testAccProviders,
				CheckDestroy: testAccCheckTFEWorkspaceDestroy,
				Steps: []resource.TestStep{
					{
						Config: testAccTFEWorkspace_triggersConfigurationGenerator(
							rInt,
							true, false,
							true, `["omar"]`,
							false, "",
						),
						Check: resource.ComposeTestCheckFunc(
							testAccCheckTFEWorkspaceExists(
								"tfe_workspace.foobar", workspace),
							resource.TestCheckResourceAttr(
								"tfe_workspace.foobar", "trigger_prefixes.#", "1"),
							resource.TestCheckResourceAttr(
								"tfe_workspace.foobar", "trigger_patterns.#", "0"),
						),
					},
					{
						Config: testAccTFEWorkspace_triggersConfigurationGenerator(
							rInt,
							true, false,
							false, "",
							true, "[]",
						),
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr(
								"tfe_workspace.foobar", "trigger_prefixes.#", "0"),
							resource.TestCheckResourceAttr(
								"tfe_workspace.foobar", "trigger_patterns.#", "0"),
						),
					},
				},
			})
		})
		t.Run("and trigger patterns are populated", func(t *testing.T) {
			workspace := &tfe.Workspace{}
			rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
			resource.Test(t, resource.TestCase{
				PreCheck:     func() { testAccPreCheck(t) },
				Providers:    testAccProviders,
				CheckDestroy: testAccCheckTFEWorkspaceDestroy,
				Steps: []resource.TestStep{
					{
						Config: testAccTFEWorkspace_triggersConfigurationGenerator(
							rInt,
							true, false,
							true, `["prefix"]`,
							false, "",
						),
						Check: resource.ComposeTestCheckFunc(
							testAccCheckTFEWorkspaceExists(
								"tfe_workspace.foobar", workspace),
							resource.TestCheckResourceAttr(
								"tfe_workspace.foobar", "trigger_prefixes.#", "1"),
							resource.TestCheckResourceAttr(
								"tfe_workspace.foobar", "trigger_patterns.#", "0"),
						),
					},
					{
						Config: testAccTFEWorkspace_triggersConfigurationGenerator(
							rInt,
							true, false,
							false, "",
							true, `["pattern"]`,
						),
						ExpectError: regexp.MustCompile(`'trigger_patterns' cannot be populated when 'file_triggers_enabled' is set to false.`),
					},
				},
			})
		})
	})

	t.Run("file triggers enabled is true", func(t *testing.T) {
		t.Run("and trigger prefixes are set and empty", func(t *testing.T) {
			workspace := &tfe.Workspace{}
			rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
			resource.Test(t, resource.TestCase{
				PreCheck:     func() { testAccPreCheck(t) },
				Providers:    testAccProviders,
				CheckDestroy: testAccCheckTFEWorkspaceDestroy,
				Steps: []resource.TestStep{
					{
						Config: testAccTFEWorkspace_triggersConfigurationGenerator(
							rInt,
							true, true,
							false, "",
							true, `["/pattern1", "/pattern2"]`,
						),
						Check: resource.ComposeTestCheckFunc(
							testAccCheckTFEWorkspaceExists(
								"tfe_workspace.foobar", workspace),
							resource.TestCheckResourceAttr(
								"tfe_workspace.foobar", "trigger_patterns.#", "2"),
						),
					},
					{
						Config: testAccTFEWorkspace_triggersConfigurationGenerator(
							rInt,
							true, true,
							true, "[]",
							false, "",
						),
						Check: resource.ComposeTestCheckFunc(
							testAccCheckTFEWorkspaceExists(
								"tfe_workspace.foobar", workspace),
							resource.TestCheckResourceAttr(
								"tfe_workspace.foobar", "trigger_prefixes.#", "0"),
							resource.TestCheckResourceAttr(
								"tfe_workspace.foobar", "trigger_patterns.#", "0"),
						),
					},
				},
			})
		})
		t.Run("and trigger prefixes are populated", func(t *testing.T) {
			workspace := &tfe.Workspace{}
			rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
			resource.Test(t, resource.TestCase{
				PreCheck:     func() { testAccPreCheck(t) },
				Providers:    testAccProviders,
				CheckDestroy: testAccCheckTFEWorkspaceDestroy,
				Steps: []resource.TestStep{
					{
						Config: testAccTFEWorkspace_triggersConfigurationGenerator(
							rInt,
							true, true,
							false, "",
							true, `["/pattern1", "/pattern2"]`,
						),
						Check: resource.ComposeTestCheckFunc(
							testAccCheckTFEWorkspaceExists(
								"tfe_workspace.foobar", workspace),
							resource.TestCheckResourceAttr(
								"tfe_workspace.foobar", "trigger_patterns.#", "2"),
						),
					},
					{
						Config: testAccTFEWorkspace_triggersConfigurationGenerator(
							rInt,
							true, true,
							true, `["/prefix1"]`,
							false, "",
						),
						Check: resource.ComposeTestCheckFunc(
							testAccCheckTFEWorkspaceExists(
								"tfe_workspace.foobar", workspace),
							resource.TestCheckResourceAttr(
								"tfe_workspace.foobar", "trigger_prefixes.#", "1"),
							resource.TestCheckResourceAttr(
								"tfe_workspace.foobar", "trigger_prefixes.0", "/prefix1"),
							resource.TestCheckResourceAttr(
								"tfe_workspace.foobar", "trigger_patterns.#", "0"),
						),
					},
				},
			})
		})
		t.Run("and trigger patterns are set and empty", func(t *testing.T) {
			workspace := &tfe.Workspace{}
			rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
			resource.Test(t, resource.TestCase{
				PreCheck:     func() { testAccPreCheck(t) },
				Providers:    testAccProviders,
				CheckDestroy: testAccCheckTFEWorkspaceDestroy,
				Steps: []resource.TestStep{
					{
						Config: testAccTFEWorkspace_triggersConfigurationGenerator(
							rInt,
							true, false,
							true, `["/prefix1", "/prefix2"]`,
							false, "",
						),
						Check: resource.ComposeTestCheckFunc(
							testAccCheckTFEWorkspaceExists(
								"tfe_workspace.foobar", workspace),
							resource.TestCheckResourceAttr(
								"tfe_workspace.foobar", "trigger_prefixes.#", "2"),
						),
					},
					{
						Config: testAccTFEWorkspace_triggersConfigurationGenerator(
							rInt,
							true, true,
							false, "",
							true, `["/patterns1", "/patterns2", "/patterns3"]`,
						),
						Check: resource.ComposeTestCheckFunc(
							testAccCheckTFEWorkspaceExists(
								"tfe_workspace.foobar", workspace),
							resource.TestCheckResourceAttr(
								"tfe_workspace.foobar", "trigger_patterns.#", "3"),
							resource.TestCheckResourceAttr(
								"tfe_workspace.foobar", "trigger_patterns.0", "/patterns1"),
							resource.TestCheckResourceAttr(
								"tfe_workspace.foobar", "trigger_patterns.1", "/patterns2"),
							resource.TestCheckResourceAttr(
								"tfe_workspace.foobar", "trigger_patterns.2", "/patterns3"),
						),
					},
				},
			})
		})
		t.Run("and trigger patterns are populated", func(t *testing.T) {
			workspace := &tfe.Workspace{}
			rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
			resource.Test(t, resource.TestCase{
				PreCheck:     func() { testAccPreCheck(t) },
				Providers:    testAccProviders,
				CheckDestroy: testAccCheckTFEWorkspaceDestroy,
				Steps: []resource.TestStep{
					{
						Config: testAccTFEWorkspace_triggersConfigurationGenerator(
							rInt,
							true, false,
							true, `["prefix"]`,
							false, "",
						),
						Check: resource.ComposeTestCheckFunc(
							testAccCheckTFEWorkspaceExists(
								"tfe_workspace.foobar", workspace),
							resource.TestCheckResourceAttr(
								"tfe_workspace.foobar", "trigger_prefixes.#", "1"),
							resource.TestCheckResourceAttr(
								"tfe_workspace.foobar", "trigger_patterns.#", "0"),
						),
					},
					{
						Config: testAccTFEWorkspace_triggersConfigurationGenerator(
							rInt,
							true, true,
							false, "",
							true, `["pattern-x/**/*", "**/pattern-y/*", "pattern-z"]`,
						),
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr(
								"tfe_workspace.foobar", "trigger_prefixes.#", "0"),
							resource.TestCheckResourceAttr(
								"tfe_workspace.foobar", "trigger_patterns.#", "3"),
							resource.TestCheckResourceAttr(
								"tfe_workspace.foobar", "trigger_patterns.0", "pattern-x/**/*"),
							resource.TestCheckResourceAttr(
								"tfe_workspace.foobar", "trigger_patterns.1", "**/pattern-y/*"),
							resource.TestCheckResourceAttr(
								"tfe_workspace.foobar", "trigger_patterns.2", "pattern-z"),
						),
					},
				},
			})
		})
		t.Run("and both trigger prefixes and patterns are populated", func(t *testing.T) {
			rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
			resource.Test(t, resource.TestCase{
				PreCheck:     func() { testAccPreCheck(t) },
				Providers:    testAccProviders,
				CheckDestroy: testAccCheckTFEWorkspaceDestroy,
				Steps: []resource.TestStep{
					{
						Config: testAccTFEWorkspace_triggersConfigurationGenerator(
							rInt,
							true, true,
							true, `["prefix"]`,
							true, `["pattern"]`,
						),
						ExpectError: regexp.MustCompile(`Conflicting configuration`),
					},
				},
			})
		})
		t.Run("and both trigger prefixes and patterns are empty", func(t *testing.T) {
			rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
			resource.Test(t, resource.TestCase{
				PreCheck:     func() { testAccPreCheck(t) },
				Providers:    testAccProviders,
				CheckDestroy: testAccCheckTFEWorkspaceDestroy,
				Steps: []resource.TestStep{
					{
						Config: testAccTFEWorkspace_triggersConfigurationGenerator(
							rInt,
							true, true,
							true, "[]",
							true, "[]",
						),
						ExpectError: regexp.MustCompile(`Conflicting configuration`),
					},
				},
			})
		})

		t.Run("change trigger prefixes to trigger patterns and vice-versa", func(t *testing.T) {
			workspace := &tfe.Workspace{}
			rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
			resource.Test(t, resource.TestCase{
				PreCheck:     func() { testAccPreCheck(t) },
				Providers:    testAccProviders,
				CheckDestroy: testAccCheckTFEWorkspaceDestroy,
				Steps: []resource.TestStep{
					{
						Config: testAccTFEWorkspace_triggersConfigurationGenerator(
							rInt,
							true, false,
							true, `["/prefix1", "/prefix2"]`,
							false, "",
						),
						Check: resource.ComposeTestCheckFunc(
							testAccCheckTFEWorkspaceExists(
								"tfe_workspace.foobar", workspace),
							resource.TestCheckResourceAttr(
								"tfe_workspace.foobar", "trigger_prefixes.#", "2"),
							resource.TestCheckResourceAttr(
								"tfe_workspace.foobar", "trigger_patterns.#", "0"),
						),
					},
					{
						Config: testAccTFEWorkspace_triggersConfigurationGenerator(
							rInt,
							false, true,
							false, "",
							false, "",
						),
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr(
								"tfe_workspace.foobar", "trigger_prefixes.#", "0"),
							resource.TestCheckResourceAttr(
								"tfe_workspace.foobar", "trigger_patterns.#", "0"),
						),
					},
					{
						Config: testAccTFEWorkspace_triggersConfigurationGenerator(
							rInt,
							true, true,
							false, "",
							true, `["/pattern1", "/pattern2"]`,
						),
						Check: resource.ComposeTestCheckFunc(
							testAccCheckTFEWorkspaceExists(
								"tfe_workspace.foobar", workspace),
							resource.TestCheckResourceAttr(
								"tfe_workspace.foobar", "trigger_prefixes.#", "0"),
							resource.TestCheckResourceAttr(
								"tfe_workspace.foobar", "trigger_patterns.#", "2"),
						),
					},
					{
						Config: testAccTFEWorkspace_triggersConfigurationGenerator(
							rInt,
							true, true,
							true, `["/prefix1", "/prefix2", "another_one"]`,
							false, "",
						),
						Check: resource.ComposeTestCheckFunc(
							testAccCheckTFEWorkspaceExists(
								"tfe_workspace.foobar", workspace),
							resource.TestCheckResourceAttr(
								"tfe_workspace.foobar", "trigger_prefixes.#", "3"),
							resource.TestCheckResourceAttr(
								"tfe_workspace.foobar", "trigger_patterns.#", "0"),
						),
					},
					{
						Config: testAccTFEWorkspace_triggersConfigurationGenerator(
							rInt,
							true, true,
							true, `["/prefix1"]`,
							false, "[]",
						),
						Check: resource.ComposeTestCheckFunc(
							testAccCheckTFEWorkspaceExists(
								"tfe_workspace.foobar", workspace),
							resource.TestCheckResourceAttr(
								"tfe_workspace.foobar", "trigger_prefixes.#", "1"),
							resource.TestCheckResourceAttr(
								"tfe_workspace.foobar", "trigger_patterns.#", "0"),
						),
					},
					{
						Config: testAccTFEWorkspace_triggersConfigurationGenerator(
							rInt,
							true, true,
							true, "[]",
							false, "",
						),
						Check: resource.ComposeTestCheckFunc(
							testAccCheckTFEWorkspaceExists(
								"tfe_workspace.foobar", workspace),
							resource.TestCheckResourceAttr(
								"tfe_workspace.foobar", "trigger_prefixes.#", "0"),
							resource.TestCheckResourceAttr(
								"tfe_workspace.foobar", "trigger_patterns.#", "0"),
						),
					},
				},
			})
		})
	})
}

func testAccTFEWorkspace_triggersConfigurationGenerator(
	rInt int,
	fileTriggersEnabledPresent bool,
	fileTriggersEnabledValue bool,
	triggerPrefixesPresent bool,
	triggerPrefixesValue string,
	triggerPatternsPresent bool,
	triggerPatternsValue string,
) string {
	result := fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d-ff-on"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
 name                  = "workspace"
 organization          = tfe_organization.foobar.id
`, rInt)

	if fileTriggersEnabledPresent {
		result += fmt.Sprintf(`
file_triggers_enabled = %v
	`, fileTriggersEnabledValue)
	}
	if triggerPrefixesPresent {
		result += fmt.Sprintf(`
trigger_prefixes = %s
	`, triggerPrefixesValue)
	}
	if triggerPatternsPresent {
		result += fmt.Sprintf(`
trigger_patterns = %s
	`, triggerPatternsValue)
	}
	result += "}"
	return result
}

func TestAccTFEWorkspace_updateTriggerPatterns(t *testing.T) {
	workspace := &tfe.Workspace{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEWorkspaceDestroy,
		Steps: []resource.TestStep{
			// Create trigger prefixes first so we can verify they are being removed if we introduce trigger patterns
			{
				Config: testAccTFEWorkspace_triggerPrefixes(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "trigger_prefixes.#", "2"),
				),
			},
			// Overwrite prefixes with patterns
			{
				Config: testAccTFEWorkspace_triggerPatterns(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "trigger_patterns.#", "2"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "trigger_patterns.0", "/modules/**/*"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "trigger_patterns.1", "/**/networking/*"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "trigger_prefixes.#", "0"),
				),
			},
			// Second update
			{
				Config: testAccTFEWorkspace_updateTriggerPatterns(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "trigger_patterns.#", "3"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "trigger_patterns.0", "/**/networking/*"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "trigger_patterns.1", "/another_module/*/test/*"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "trigger_patterns.2", "/**/resources/**/*"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "trigger_prefixes.#", "0"),
				),
			},
			{
				Config: testAccTFEWorkspace_updateEmptyTriggerPatterns(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists("tfe_workspace.foobar", workspace),
					testAccCheckTFEWorkspaceAttributes(workspace),
					resource.TestCheckResourceAttr("tfe_workspace.foobar", "trigger_patterns.#", "0"),
					resource.TestCheckResourceAttr("tfe_workspace.foobar", "trigger_prefixes.#", "0"),
				),
			},
		},
	})
}

func TestAccTFEWorkspace_patternsAndPrefixesConflicting(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFEWorkspace_prefixesAndPatternsConflicting(rInt),
				ExpectError: regexp.MustCompile(`Conflicting configuration`),
			},
		},
	})
}

func TestAccTFEWorkspace_changeTags(t *testing.T) {
	workspace := &tfe.Workspace{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				// create with 2 tags
				Config: testAccTFEWorkspace_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "tag_names.#", "2"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "tag_names.0", "fav"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "tag_names.1", "test"),
				),
			},
			{
				// remove 1
				Config: testAccTFEWorkspace_basicRemoveTag(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "tag_names.#", "1"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "tag_names.0", "prod"),
				),
			},
			{
				// add 1
				Config: testAccTFEWorkspace_basicChangeTags(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "tag_names.#", "2"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "tag_names.0", "fav"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "tag_names.1", "prod"),
				),
			},
			{
				// remove 1 again
				Config: testAccTFEWorkspace_basicRemoveTag(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "tag_names.#", "1"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "tag_names.0", "prod"),
				),
			},
			{
				// change unrelated attr
				Config: testAccTFEWorkspace_basicRemoveTagAlt(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "tag_names.#", "1"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "tag_names.0", "prod"),
				),
			},
			{
				// remove 1, add 2
				Config: testAccTFEWorkspace_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "tag_names.#", "2"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "tag_names.0", "fav"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "tag_names.1", "test"),
				),
			},
			{
				// remove all
				Config: testAccTFEWorkspace_basicNoTags(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "tag_names.#", "0"),
				),
			},
			{
				// add 2
				Config: testAccTFEWorkspace_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "tag_names.#", "2"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "tag_names.0", "fav"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "tag_names.1", "test"),
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
				Config: testAccTFEWorkspace_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "speculative_enabled", "true"),
				),
			},

			{
				Config: testAccTFEWorkspace_basicSpeculativeOff(rInt),
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

func TestAccTFEWorkspace_structuredRunOutputDisabled(t *testing.T) {
	workspace := &tfe.Workspace{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEWorkspace_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "structured_run_output_enabled", "true"),
				),
			},

			{
				Config: testAccTFEWorkspace_updateStructuredRunOutput(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "structured_run_output_enabled", "false"),
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
			testAccGithubPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEWorkspace_basic(rInt),
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
						"tfe_workspace.foobar", "description", "workspace-test-add-vcs-repo"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "vcs_repo.0.identifier", GITHUB_WORKSPACE_IDENTIFIER),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "vcs_repo.0.branch", ""),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "vcs_repo.0.ingress_submodules", "false"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "vcs_repo.0.tags_regex", ""),
				),
			},
			{
				Config: testAccTFEWorkspace_updateUpdateVCSRepoBranch(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists("tfe_workspace.foobar", workspace),
					testAccCheckTFEWorkspaceUpdatedUpdateVCSRepoBranchAttributes(workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "description", "workspace-test-update-vcs-repo-branch"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "vcs_repo.0.identifier", GITHUB_WORKSPACE_IDENTIFIER),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "vcs_repo.0.branch", GITHUB_WORKSPACE_BRANCH),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "vcs_repo.0.ingress_submodules", "false"),
				),
			},
			{
				Config: testAccTFEWorkspace_updateRemoveVCSRepo(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists("tfe_workspace.foobar", workspace),
					testAccCheckTFEWorkspaceUpdatedRemoveVCSRepoAttributes(workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "description", "workspace-test-remove-vcs-repo"),
					resource.TestCheckNoResourceAttr("tfe_workspace.foobar", "vcs_repo"),
				),
			},
		},
	})
}

func TestAccTFEWorkspace_updateVCSRepoTagsRegex(t *testing.T) {
	workspace := &tfe.Workspace{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccGithubPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEWorkspace_updateUpdateVCSRepoTagsRegex(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists("tfe_workspace.foobar", workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "description", "workspace-test-update-vcs-repo-tags-regex"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "file_triggers_enabled", "false"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "vcs_repo.0.identifier", GITHUB_WORKSPACE_IDENTIFIER),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "vcs_repo.0.branch", GITHUB_WORKSPACE_BRANCH),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "vcs_repo.0.ingress_submodules", "false"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "vcs_repo.0.tags_regex", `\d+.\d+.\d+`),
				),
			},
			{
				Config: testAccTFEWorkspace_updateRemoveVCSRepoTagsRegex(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists("tfe_workspace.foobar", workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "description", "workspace-test-update-vcs-repo-tags-regex"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "file_triggers_enabled", "false"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "vcs_repo.0.identifier", GITHUB_WORKSPACE_IDENTIFIER),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "vcs_repo.0.branch", GITHUB_WORKSPACE_BRANCH),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "vcs_repo.0.ingress_submodules", "false"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "vcs_repo.0.tags_regex", ""),
				),
			},
		},
	})
}

func TestAccTFEWorkspace_updateVCSRepoChangeTagRegexToTriggerPattern(t *testing.T) {
	workspace := &tfe.Workspace{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccGithubPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEWorkspace_updateUpdateVCSRepoTagsRegex(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists("tfe_workspace.foobar", workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "description", "workspace-test-update-vcs-repo-tags-regex"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "file_triggers_enabled", "false"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "vcs_repo.0.identifier", GITHUB_WORKSPACE_IDENTIFIER),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "vcs_repo.0.branch", GITHUB_WORKSPACE_BRANCH),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "vcs_repo.0.ingress_submodules", "false"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "vcs_repo.0.tags_regex", `\d+.\d+.\d+`),
				),
			},
			{
				Config: testAccTFEWorkspace_updateToTriggerPatternsFromTagsRegex(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists("tfe_workspace.foobar", workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "description", "workspace-test-update-vcs-repo-tags-regex"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "file_triggers_enabled", "true"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "vcs_repo.0.identifier", GITHUB_WORKSPACE_IDENTIFIER),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "vcs_repo.0.branch", GITHUB_WORKSPACE_BRANCH),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "vcs_repo.0.ingress_submodules", "false"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "vcs_repo.0.tags_regex", ""),
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
				Config: testAccTFEWorkspace_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					testAccCheckTFEWorkspaceAttributes(workspace),
				),
			},

			{
				Config: testAccTFEWorkspace_sshKey(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					testAccCheckTFEWorkspaceAttributesSSHKey(workspace),
					resource.TestCheckResourceAttrSet(
						"tfe_workspace.foobar", "ssh_key_id"),
				),
			},

			{
				Config: testAccTFEWorkspace_noSSHKey(rInt),
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
				Config: testAccTFEWorkspace_basic(rInt),
			},

			{
				ResourceName:      "tfe_workspace.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "tfe_workspace.foobar",
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("tst-terraform-%d/workspace-test", rInt),
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
			testAccGithubPreCheck(t)
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
						"tfe_workspace.foobar", "description", "workspace-test-update-vcs-repo-branch"),
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

func TestAccTFEWorkspace_operationsAndExecutionModeInteroperability(t *testing.T) {
	skipIfFreeOnly(t)
	skipIfEnterprise(t)

	workspace := &tfe.Workspace{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEWorkspace_operationsTrue(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "operations", "true"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "execution_mode", "remote"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "agent_pool_id", ""),
				),
			},
			{
				Config: testAccTFEWorkspace_executionModeLocal(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "operations", "false"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "execution_mode", "local"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "agent_pool_id", ""),
				),
			},
			{
				Config: testAccTFEWorkspace_operationsFalse(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "operations", "false"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "execution_mode", "local"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "agent_pool_id", ""),
				),
			},
			{
				Config: testAccTFEWorkspace_executionModeRemote(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "operations", "true"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "execution_mode", "remote"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "agent_pool_id", ""),
				),
			},
			{
				Config: testAccTFEWorkspace_executionModeAgent(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "operations", "true"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "execution_mode", "agent"),
					resource.TestCheckResourceAttrSet(
						"tfe_workspace.foobar", "agent_pool_id"),
				),
			},
		},
	})
}

func TestAccTFEWorkspace_globalRemoteState(t *testing.T) {
	workspace := &tfe.Workspace{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEWorkspace_globalRemoteStateFalse(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					testAccCheckTFEWorkspaceAttributes(workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "name", "workspace-test"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "global_remote_state", "false"),
				),
			},
			{
				Config: testAccTFEWorkspace_globalRemoteStateTrue(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					testAccCheckTFEWorkspaceAttributes(workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "name", "workspace-test"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "global_remote_state", "true"),
				),
			},
		},
	})
}

func TestAccTFEWorkspace_alterRemoteStateConsumers(t *testing.T) {
	workspace := &tfe.Workspace{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEWorkspace_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					resource.TestCheckResourceAttr("tfe_workspace.foobar", "global_remote_state", "true"),
				),
			},
			{
				Config: testAccTFEWorkspace_OneRemoteStateConsumer(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					resource.TestCheckResourceAttr("tfe_workspace.foobar", "global_remote_state", "false"),
					testAccCheckTFEWorkspaceHasRemoteConsumers("tfe_workspace.foobar", []string{"tfe_workspace.foobar_one"}),
				),
			},
			{
				Config: testAccTFEWorkspace_TwoRemoteStateConsumers(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					resource.TestCheckResourceAttr("tfe_workspace.foobar", "global_remote_state", "false"),
					testAccCheckTFEWorkspaceHasRemoteConsumers("tfe_workspace.foobar", []string{"tfe_workspace.foobar_one", "tfe_workspace.foobar_two"}),
				),
			},
			{
				Config: testAccTFEWorkspace_OneRemoteStateConsumer(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					resource.TestCheckResourceAttr("tfe_workspace.foobar", "global_remote_state", "false"),
					testAccCheckTFEWorkspaceHasRemoteConsumers("tfe_workspace.foobar", []string{"tfe_workspace.foobar_one"}),
				),
			},
			{
				Config: testAccTFEWorkspace_globalRemoteStateTrue(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					resource.TestCheckResourceAttr("tfe_workspace.foobar", "global_remote_state", "true"),
					testAccCheckTFEWorkspaceHasRemoteConsumers("tfe_workspace.foobar", []string{}),
				),
			},
		},
	})
}

func TestAccTFEWorkspace_createWithRemoteStateConsumers(t *testing.T) {
	workspace := &tfe.Workspace{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEWorkspace_TwoRemoteStateConsumers(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					resource.TestCheckResourceAttr("tfe_workspace.foobar", "global_remote_state", "false"),
					testAccCheckTFEWorkspaceHasRemoteConsumers("tfe_workspace.foobar", []string{"tfe_workspace.foobar_one", "tfe_workspace.foobar_two"}),
				),
			},
		},
	})
}

// Test pagination works for remote state consumers. Adding over 100 consumers should result in a
// subsequent empty plan if pagination works correctly. The client fetches the maximum results per
// page (100) by default.
func TestAccTFEWorkspace_paginatedRemoteStateConsumers(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEWorkspace_OverAPageOfRemoteStateConsumers(rInt),
				Check:  resource.TestCheckResourceAttr("tfe_workspace.foobar", "remote_state_consumer_ids.#", "105"),
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
			return fmt.Errorf("Could not delete %s: %w", n, err)
		}

		// Read the workspace again using the lower level resource reader
		// which will trigger the panic
		rd := &schema.ResourceData{}
		rd.SetId(rs.Primary.ID)

		err = resourceTFEWorkspaceRead(rd, testAccProvider.Meta())
		if err != nil && err != tfe.ErrResourceNotFound {
			return fmt.Errorf("Could not re-read resource directly: %w", err)
		}

		return nil
	}
}

func testAccCheckTFEWorkspaceHasRemoteConsumers(ws string, wsConsumers []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rsWorkspace, ok := s.RootModule().Resources[ws]
		if !ok {
			return fmt.Errorf("Not found: %s", ws)
		}
		numConsumersStr := rsWorkspace.Primary.Attributes["remote_state_consumer_ids.#"]
		numConsumers, _ := strconv.Atoi(numConsumersStr)

		remoteConsumerMap := map[string]struct{}{}
		for i := 0; i < numConsumers; i++ {
			consumer := rsWorkspace.Primary.Attributes[fmt.Sprintf("remote_state_consumer_ids.%d", i)]
			remoteConsumerMap[consumer] = struct{}{}
		}

		for _, consumer := range wsConsumers {
			remoteConsumer, ok := s.RootModule().Resources[consumer]
			if !ok {
				return fmt.Errorf("Not found: %s", consumer)
			}
			consumerID := remoteConsumer.Primary.Attributes["id"]
			if _, hasConsumer := remoteConsumerMap[consumerID]; !hasConsumer {
				return fmt.Errorf("The Workspace %s does not appear to be a remote state consumer for %s", rsWorkspace.Primary.ID, consumerID)
			}
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

		if workspace.ExecutionMode != "remote" {
			return fmt.Errorf("Bad execution mode: %s", workspace.ExecutionMode)
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

		if workspace.ExecutionMode != "local" {
			return fmt.Errorf("Bad execution mode: %s", workspace.ExecutionMode)
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
		if workspace.Description != "workspace-test-add-vcs-repo" {
			return fmt.Errorf("Bad description: %s", workspace.Name)
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
		if workspace.Description != "workspace-test-update-vcs-repo-branch" {
			return fmt.Errorf("Bad description: %s", workspace.Name)
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
		if workspace.Description != "workspace-test-remove-vcs-repo" {
			return fmt.Errorf("Bad description: %s", workspace.Name)
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

func testAccTFEWorkspace_basic(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name               = "workspace-test"
  organization       = tfe_organization.foobar.id
  description        = "My favorite workspace!"
  allow_destroy_plan = false
  auto_apply         = true
  tag_names          = ["fav", "test"]
}`, rInt)
}

func testAccTFEWorkspace_basicChangeTags(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name               = "workspace-test"
  organization       = tfe_organization.foobar.id
  auto_apply         = true
  tag_names          = ["fav", "prod"]
}`, rInt)
}

func testAccTFEWorkspace_basicNoTags(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name               = "workspace-test"
  organization       = tfe_organization.foobar.id
  auto_apply         = true
  tag_names          = []
}`, rInt)
}

func testAccTFEWorkspace_basicRemoveTag(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name               = "workspace-test"
  organization       = tfe_organization.foobar.id
  auto_apply         = true
  tag_names          = ["prod"]
}`, rInt)
}

func testAccTFEWorkspace_basicRemoveTagAlt(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name               = "workspace-test"
  organization       = tfe_organization.foobar.id
  auto_apply         = false
  tag_names          = ["prod"]
}`, rInt)
}

func testAccTFEWorkspace_basicFileTriggersOff(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name                  = "workspace-test"
  organization          = tfe_organization.foobar.id
  auto_apply            = true
  file_triggers_enabled = false
}`, rInt)
}

func testAccTFEWorkspace_operationsTrue(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
  operations = true
}`, rInt)
}

func testAccTFEWorkspace_operationsFalse(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
  operations = false
}`, rInt)
}

func testAccTFEWorkspace_executionModeRemote(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
  execution_mode = "remote"
}`, rInt)
}

func testAccTFEWorkspace_executionModeLocal(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
  execution_mode = "local"
}`, rInt)
}

func testAccTFEWorkspace_executionModeAgent(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_agent_pool" "foobar" {
  name = "agent-pool-test"
  organization = tfe_organization.foobar.name
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
  execution_mode = "agent"
  agent_pool_id = tfe_agent_pool.foobar.id
}`, rInt)
}

func testAccTFEWorkspace_basicSpeculativeOff(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name                  = "workspace-test"
  organization          = tfe_organization.foobar.id
  auto_apply            = true
  speculative_enabled = false
}`, rInt)
}

func testAccTFEWorkspace_monorepo(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name                  = "workspace-monorepo"
  organization          = tfe_organization.foobar.id
  file_triggers_enabled = true
  trigger_prefixes      = ["/modules", "/shared"]
  working_directory     = "/db"
}`, rInt)
}

func testAccTFEWorkspace_renamed(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name               = "renamed-out-of-band"
  organization       = tfe_organization.foobar.id
  description        = "My favorite workspace!"
  allow_destroy_plan = false
  auto_apply         = true
}`, rInt)
}

func testAccTFEWorkspace_update(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name                  = "workspace-updated"
  organization          = tfe_organization.foobar.id
  allow_destroy_plan    = true
  auto_apply            = false
  file_triggers_enabled = true
  queue_all_runs        = false
  terraform_version     = "0.11.1"
  trigger_prefixes      = ["/modules", "/shared"]
  working_directory     = "terraform/test"
  operations            = false
}`, rInt)
}

func testAccTFEWorkspace_updateAddWorkingDirectory(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name                  = "workspace-updated"
  organization          = tfe_organization.foobar.id
  auto_apply            = false
  working_directory     = "terraform/test"
}`, rInt)
}

func testAccTFEWorkspace_updateRemoveWorkingDirectory(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name                  = "workspace-updated"
  organization          = tfe_organization.foobar.id
  auto_apply            = false
}`, rInt)
}

func testAccTFEWorkspace_sshKey(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_ssh_key" "foobar" {
  name         = "ssh-key-test"
  organization = tfe_organization.foobar.id
  key          = "SSH-KEY-CONTENT"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
  auto_apply   = true
  ssh_key_id   = tfe_ssh_key.foobar.id
}`, rInt)
}

func testAccTFEWorkspace_noSSHKey(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_ssh_key" "foobar" {
  name         = "ssh-key-test"
  organization = tfe_organization.foobar.id
  key          = "SSH-KEY-CONTENT"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
  auto_apply   = true
}`, rInt)
}

func testAccTFEWorkspace_triggerPrefixes(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d-ff-on"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name                  = "workspace"
  organization          = tfe_organization.foobar.id
  trigger_prefixes      = ["/modules", "/shared"]
}`, rInt)
}

func testAccTFEWorkspace_updateEmptyTriggerPrefixes(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d-ff-on"
  email = "admin@company.com"
}
resource "tfe_workspace" "foobar" {
  name                  = "workspace-test"
  organization          = tfe_organization.foobar.id
  auto_apply            = true
  trigger_prefixes      = []
}`, rInt)
}

func testAccTFEWorkspace_triggerPatterns(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d-ff-on"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name                  = "workspace"
  organization          = tfe_organization.foobar.id
  trigger_patterns      = ["/modules/**/*", "/**/networking/*"]
}`, rInt)
}

func testAccTFEWorkspace_updateTriggerPatterns(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d-ff-on"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name                  = "workspace"
  organization          = tfe_organization.foobar.id
  trigger_patterns      = ["/**/networking/*", "/another_module/*/test/*", "/**/resources/**/*"]
}`, rInt)
}

func testAccTFEWorkspace_updateEmptyTriggerPatterns(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d-ff-on"
  email = "admin@company.com"
}
resource "tfe_workspace" "foobar" {
  name                  = "workspace-test"
  organization          = tfe_organization.foobar.id
  auto_apply            = true
  trigger_patterns      = []
}`, rInt)
}

func testAccTFEWorkspace_prefixesAndPatternsConflicting(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d-ff-on"
  email = "admin@company.com"
}
resource "tfe_workspace" "foobar" {
  name                  = "workspace-test"
  organization          = tfe_organization.foobar.id
  trigger_prefixes      = []
  trigger_patterns      = []
}`, rInt)
}

func testAccTFEWorkspace_updateAddVCSRepo(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_oauth_client" "test" {
  organization     = tfe_organization.foobar.id
  api_url          = "https://api.github.com"
  http_url         = "https://github.com"
  oauth_token      = "%s"
  service_provider = "github"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  description  = "workspace-test-add-vcs-repo"
  organization = tfe_organization.foobar.id
  auto_apply   = true
  vcs_repo {
    identifier     = "%s"
    oauth_token_id = tfe_oauth_client.test.oauth_token_id
  }
}
`,
		rInt,
		GITHUB_TOKEN,
		GITHUB_WORKSPACE_IDENTIFIER,
	)
}

func testAccTFEWorkspace_updateUpdateVCSRepoBranchFileTriggersDisabled(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_oauth_client" "test" {
  organization     = tfe_organization.foobar.id
  api_url          = "https://api.github.com"
  http_url         = "https://github.com"
  oauth_token      = "%s"
  service_provider = "github"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  description  = "workspace-test-update-vcs-repo-branch"
  organization = tfe_organization.foobar.id
  auto_apply   = true
	## file_triggers_enabled = false
  vcs_repo {
    identifier     = "%s"
    oauth_token_id = tfe_oauth_client.test.oauth_token_id
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

func testAccTFEWorkspace_updateUpdateVCSRepoBranch(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_oauth_client" "test" {
  organization     = tfe_organization.foobar.id
  api_url          = "https://api.github.com"
  http_url         = "https://github.com"
  oauth_token      = "%s"
  service_provider = "github"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  description  = "workspace-test-update-vcs-repo-branch"
  organization = tfe_organization.foobar.id
  auto_apply   = true
  vcs_repo {
    identifier     = "%s"
    oauth_token_id = tfe_oauth_client.test.oauth_token_id
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

func testAccTFEWorkspace_updateRemoveVCSRepo(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  description  = "workspace-test-remove-vcs-repo"
  organization = tfe_organization.foobar.id
  auto_apply   = true
}
`, rInt)
}

func testAccTFEWorkspace_updateRemoveVCSTagsRegexRepo(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-tf-%d-git-tag-ff-on"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name         			= "workspace-test"
  description  			= "workspace-test-update-vcs-repo-tags-regex"
  organization = tfe_organization.foobar.id
  auto_apply   = true
}
`, rInt)
}

func testAccTFEWorkspace_updateUpdateVCSRepoTagsRegex(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-tf-%d-git-tag-ff-on"
  email = "admin@company.com"
}

resource "tfe_oauth_client" "test" {
  organization     = tfe_organization.foobar.id
  api_url          = "https://api.github.com"
  http_url         = "https://github.com"
  oauth_token      = "%s"
  service_provider = "github"
}

resource "tfe_workspace" "foobar" {
  name         			= "workspace-test"
  description  			= "workspace-test-update-vcs-repo-tags-regex"
  organization 			= tfe_organization.foobar.id
  auto_apply   			= true
  file_triggers_enabled = false
  vcs_repo {
    identifier     = "%s"
    oauth_token_id = tfe_oauth_client.test.oauth_token_id
    branch         = "%s"
	tags_regex     = "\\d+.\\d+.\\d+"
  }
}
`,
		rInt,
		GITHUB_TOKEN,
		GITHUB_WORKSPACE_IDENTIFIER,
		GITHUB_WORKSPACE_BRANCH,
	)
}

func testAccTFEWorkspace_updateRemoveVCSRepoTagsRegex(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-tf-%d-git-tag-ff-on"
  email = "admin@company.com"
}

resource "tfe_oauth_client" "test" {
  organization     = tfe_organization.foobar.id
  api_url          = "https://api.github.com"
  http_url         = "https://github.com"
  oauth_token      = "%s"
  service_provider = "github"
}

resource "tfe_workspace" "foobar" {
  name         			= "workspace-test"
  description  			= "workspace-test-update-vcs-repo-tags-regex"
  organization 			= tfe_organization.foobar.id
  auto_apply   			= true
  file_triggers_enabled = false
  vcs_repo {
    identifier     = "%s"
    oauth_token_id = tfe_oauth_client.test.oauth_token_id
    branch         = "%s"
	tags_regex     = ""
  }
}
`,
		rInt,
		GITHUB_TOKEN,
		GITHUB_WORKSPACE_IDENTIFIER,
		GITHUB_WORKSPACE_BRANCH,
	)
}

func testAccTFEWorkspace_updateToTriggerPatternsFromTagsRegex(rInt int) string {
	return fmt.Sprintf(`
	resource "tfe_organization" "foobar" {
		name  = "tst-tf-%d-git-tag-ff-on"
		email = "admin@company.com"
	}
	
	resource "tfe_oauth_client" "test" {
		organization     = tfe_organization.foobar.id
		api_url          = "https://api.github.com"
		http_url         = "https://github.com"
		oauth_token      = "%s"
		service_provider = "github"
	}
	
	resource "tfe_workspace" "foobar" {
		name         			= "workspace-test"
		description  			= "workspace-test-update-vcs-repo-tags-regex"
		organization 			= tfe_organization.foobar.id
		auto_apply   			= true
		file_triggers_enabled = true
		trigger_patterns = ["foo/**/*"]
		vcs_repo {
			identifier     = "%s"
			oauth_token_id = tfe_oauth_client.test.oauth_token_id
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

func testAccTFEWorkspace_globalRemoteStateFalse(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name                = "workspace-test"
  organization        = tfe_organization.foobar.id
  description         = "My favorite workspace!"
  allow_destroy_plan  = false
  auto_apply          = true
  global_remote_state = false
}`, rInt)
}

func testAccTFEWorkspace_globalRemoteStateTrue(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name                = "workspace-test"
  organization        = tfe_organization.foobar.id
  description         = "My favorite workspace!"
  allow_destroy_plan  = false
  auto_apply          = true
  global_remote_state = true
	remote_state_consumer_ids = []
}`, rInt)
}

func testAccTFEWorkspace_TwoRemoteStateConsumers(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name               = "workspace-test"
  organization       = tfe_organization.foobar.id
  allow_destroy_plan = false
  auto_apply = true
  global_remote_state = false
  remote_state_consumer_ids = [tfe_workspace.foobar_one.id, tfe_workspace.foobar_two.id]
}

resource "tfe_workspace" "foobar_one" {
  name               = "workspace-test-1"
  organization       = tfe_organization.foobar.id
}

resource "tfe_workspace" "foobar_two" {
  name               = "workspace-test-2"
  organization       = tfe_organization.foobar.id
}`, rInt)
}

func testAccTFEWorkspace_OneRemoteStateConsumer(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name               = "workspace-test"
  organization       = tfe_organization.foobar.id
  allow_destroy_plan = false
  auto_apply = true
  global_remote_state = false
  remote_state_consumer_ids = [tfe_workspace.foobar_one.id]
}

resource "tfe_workspace" "foobar_one" {
  name               = "workspace-test-1"
  organization       = tfe_organization.foobar.id
}

resource "tfe_workspace" "foobar_two" {
  name               = "workspace-test-2"
  organization       = tfe_organization.foobar.id
}`, rInt)
}

func testAccTFEWorkspace_OverAPageOfRemoteStateConsumers(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name               = "workspace-test"
  organization       = tfe_organization.foobar.id
  allow_destroy_plan = false
  auto_apply = true
  global_remote_state = false
  remote_state_consumer_ids = tfe_workspace.state_consumers[*].id
}

resource "tfe_workspace" "state_consumers" {
  count = 105

  name               = "remote-state-consumer-${count.index}"
  organization       = tfe_organization.foobar.id
}`, rInt)
}

func testAccTFEWorkspace_updateStructuredRunOutput(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name                          = "workspace-test"
  organization                  = tfe_organization.foobar.id
  auto_apply                    = true
  structured_run_output_enabled = false
}`, rInt)
}
