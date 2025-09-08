// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

// test that agent pool needs execution mode and vice versa

func TestAccTFEProjectSettings_DefaultExecutionMode(t *testing.T) {
	project := &tfe.Project{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEProjectSettings_empty(rInt),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectUnknownValue("tfe_project_settings.foobar_settings", tfjsonpath.New("default_execution_mode")),
						plancheck.ExpectUnknownValue("tfe_project_settings.foobar_settings", tfjsonpath.New("default_agent_pool_id")),
						plancheck.ExpectUnknownValue("tfe_project_settings.foobar_settings", tfjsonpath.New("overwrites")),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEProjectExists(
						"tfe_project.foobar", project),

					func(s *terraform.State) error {
						if project.Name != "project_settings_test" {
							return fmt.Errorf("Bad name: %s", project.Name)
						}
						return nil
					},

					resource.TestCheckResourceAttr(
						"tfe_project.foobar", "name", "project_settings_test"),
					// the default execution mode should be the org default which is remote
					resource.TestCheckResourceAttr(
						"tfe_project_settings.foobar_settings", "default_execution_mode", "remote"),
					resource.TestCheckNoResourceAttr(
						"tfe_project_settings.foobar_settings", "default_agent_pool_id"),
					resource.TestCheckResourceAttr(
						"tfe_project_settings.foobar_settings", "overwrites.0.default_execution_mode", "false"),
					resource.TestCheckResourceAttr(
						"tfe_project_settings.foobar_settings", "overwrites.0.default_agent_pool_id", "false"),
				),
			},
			{
				Config: testAccTFEProjectSettings_executionMode(rInt, "remote"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectKnownValue("tfe_project_settings.foobar_settings",
							tfjsonpath.New("default_execution_mode"),
							knownvalue.StringExact("remote")),
						plancheck.ExpectKnownValue("tfe_project_settings.foobar_settings",
							tfjsonpath.New("default_agent_pool_id"),
							knownvalue.Null()),
						plancheck.ExpectKnownValue("tfe_project_settings.foobar_settings",
							tfjsonpath.New("overwrites").AtSliceIndex(0).AtMapKey("default_execution_mode"),
							knownvalue.Bool(true)),
						plancheck.ExpectKnownValue("tfe_project_settings.foobar_settings",
							tfjsonpath.New("overwrites").AtSliceIndex(0).AtMapKey("default_agent_pool_id"),
							knownvalue.Bool(true)),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"tfe_project_settings.foobar_settings", "default_execution_mode", "remote"),
					resource.TestCheckNoResourceAttr(
						"tfe_project_settings.foobar_settings", "default_agent_pool_id"),
					resource.TestCheckResourceAttr(
						"tfe_project_settings.foobar_settings", "overwrites.0.default_execution_mode", "true"),
					resource.TestCheckResourceAttr(
						"tfe_project_settings.foobar_settings", "overwrites.0.default_agent_pool_id", "true"),
				),
			},
			{
				Config: testAccTFEProjectSettings_executionMode(rInt, "local"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectKnownValue("tfe_project_settings.foobar_settings",
							tfjsonpath.New("default_execution_mode"),
							knownvalue.StringExact("local")),
						plancheck.ExpectKnownValue("tfe_project_settings.foobar_settings",
							tfjsonpath.New("overwrites").AtSliceIndex(0).AtMapKey("default_execution_mode"),
							knownvalue.Bool(true)),
						plancheck.ExpectKnownValue("tfe_project_settings.foobar_settings",
							tfjsonpath.New("overwrites").AtSliceIndex(0).AtMapKey("default_agent_pool_id"),
							knownvalue.Bool(true)),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"tfe_project_settings.foobar_settings", "default_execution_mode", "local"),
					resource.TestCheckNoResourceAttr(
						"tfe_project_settings.foobar_settings", "default_agent_pool_id"),
					resource.TestCheckResourceAttr(
						"tfe_project_settings.foobar_settings", "overwrites.0.default_execution_mode", "true"),
					resource.TestCheckResourceAttr(
						"tfe_project_settings.foobar_settings", "overwrites.0.default_agent_pool_id", "true"),
				),
			},
			{
				Config: testAccTFEProjectSettings_executionMode(rInt, "agent"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectKnownValue("tfe_project_settings.foobar_settings",
							tfjsonpath.New("default_execution_mode"),
							knownvalue.StringExact("agent")),
						plancheck.ExpectKnownValue("tfe_project_settings.foobar_settings",
							tfjsonpath.New("default_agent_pool_id"),
							knownvalue.StringFunc(func(v string) error {
								if strings.HasPrefix(v, "apool-") {
									return nil
								}
								return fmt.Errorf("expected an agent pool id, got %s", v)
							})),
						plancheck.ExpectKnownValue("tfe_project_settings.foobar_settings",
							tfjsonpath.New("overwrites").AtSliceIndex(0).AtMapKey("default_execution_mode"),
							knownvalue.Bool(true)),
						plancheck.ExpectKnownValue("tfe_project_settings.foobar_settings",
							tfjsonpath.New("overwrites").AtSliceIndex(0).AtMapKey("default_agent_pool_id"),
							knownvalue.Bool(true)),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"tfe_project_settings.foobar_settings", "default_execution_mode", "agent"),
					resource.TestCheckResourceAttrSet(
						"tfe_project_settings.foobar_settings", "default_agent_pool_id"),
					resource.TestCheckResourceAttr(
						"tfe_project_settings.foobar_settings", "overwrites.0.default_execution_mode", "true"),
					resource.TestCheckResourceAttr(
						"tfe_project_settings.foobar_settings", "overwrites.0.default_agent_pool_id", "true"),
				),
			},
			{
				Config: testAccTFEProjectSettings_empty(rInt),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectUnknownValue("tfe_project_settings.foobar_settings", tfjsonpath.New("default_execution_mode")),
						plancheck.ExpectUnknownValue("tfe_project_settings.foobar_settings", tfjsonpath.New("default_agent_pool_id")),
						plancheck.ExpectKnownValue("tfe_project_settings.foobar_settings",
							tfjsonpath.New("overwrites").AtSliceIndex(0).AtMapKey("default_execution_mode"),
							knownvalue.Bool(false)),
						plancheck.ExpectKnownValue("tfe_project_settings.foobar_settings",
							tfjsonpath.New("overwrites").AtSliceIndex(0).AtMapKey("default_agent_pool_id"),
							knownvalue.Bool(false)),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"tfe_project_settings.foobar_settings", "default_execution_mode", "remote"),
					resource.TestCheckNoResourceAttr(
						"tfe_project_settings.foobar_settings", "default_agent_pool_id"),
					resource.TestCheckResourceAttr(
						"tfe_project_settings.foobar_settings", "overwrites.0.default_execution_mode", "false"),
					resource.TestCheckResourceAttr(
						"tfe_project_settings.foobar_settings", "overwrites.0.default_agent_pool_id", "false"),
				),
			},
		},
	})
}

func TestAccTFEProjectSettingsImport(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	project := &tfe.Project{}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEProjectSettings_empty(rInt),
				Check: testAccCheckTFEProjectExists(
					"tfe_project.foobar", project),
			},

			{
				ResourceName:      "tfe_project_settings.foobar_settings",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "tfe_project_settings.foobar_settings",
				ImportState:       true,
				ImportStateId:     project.ID,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTFEProjectSettings_executionModeAgentPoolMismatch(t *testing.T) {
	project := &tfe.Project{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	// Verify that setting execution mode to agent requires and agent pool ID, and vice versa
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEProjectSettings_empty(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEProjectExists(
						"tfe_project.foobar", project),

					func(s *terraform.State) error {
						if project.Name != "project_settings_test" {
							return fmt.Errorf("Bad name: %s", project.Name)
						}
						return nil
					},
				),
			},
			{
				Config:      testAccTFEProjectSettings_executionModeWithoutAgentPool(rInt),
				ExpectError: regexp.MustCompile(`If default execution mode is \"agent\", \"default_agent_pool_id\" is required`),
			},
			{
				Config:      testAccTFEProjectSettings_agentPoolWithoutExecutionMode(rInt),
				ExpectError: regexp.MustCompile(`If default execution mode is not \"agent\", \"default_agent_pool_id\" must not be`),
			},
		},
	})
}

func testAccTFEProjectSettings_empty(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_agent_pool" "project_pool" {
  name         = "project-agent-pool"
  organization = tfe_organization.foobar.name
}

resource "tfe_project" "foobar" {
  organization = tfe_organization.foobar.name
  name = "project_settings_test"
  description = "project description"
}

resource "tfe_project_settings" "foobar_settings" {
	  project_id = tfe_project.foobar.id
}`, rInt)
}

func testAccTFEProjectSettings_executionMode(rInt int, executionMode string) string {
	agentPool := ""
	if executionMode == "agent" {
		agentPool = ` default_agent_pool_id = tfe_agent_pool.project_pool.id `
	}
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_agent_pool" "project_pool" {
  name         = "project-agent-pool"
  organization = tfe_organization.foobar.name
}

resource "tfe_project" "foobar" {
  organization = tfe_organization.foobar.name
  name = "projecttest"
  description = "project description"
}

resource "tfe_project_settings" "foobar_settings" {
	  project_id = tfe_project.foobar.id
	  default_execution_mode="%s"
	  %s
}`, rInt, executionMode, agentPool)
}

func testAccTFEProjectSettings_executionModeWithoutAgentPool(rInt int) string {
	return fmt.Sprintf(`resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_agent_pool" "project_pool" {
  name         = "project-agent-pool"
  organization = tfe_organization.foobar.name
}

resource "tfe_project" "foobar" {
  organization = tfe_organization.foobar.name
  name = "project_settings_test"
  description = "project description"
}

resource "tfe_project_settings" "foobar_settings" {
	  project_id = tfe_project.foobar.id
	  default_execution_mode="agent"
}`, rInt)
}

func testAccTFEProjectSettings_agentPoolWithoutExecutionMode(rInt int) string {
	return fmt.Sprintf(`resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_agent_pool" "project_pool" {
  name         = "project-agent-pool"
  organization = tfe_organization.foobar.name
}

resource "tfe_project" "foobar" {
  organization = tfe_organization.foobar.name
  name = "project_settings_test"
  description = "project description"
}

resource "tfe_project_settings" "foobar_settings" {
	  project_id = tfe_project.foobar.id
	  default_agent_pool_id = tfe_agent_pool.project_pool.id
}`, rInt)
}
