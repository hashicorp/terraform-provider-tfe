// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"math/rand"
	"regexp"
	"testing"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccTFEOrganizationRunTaskGlobalSettings_validateSchemaAttributeUrl(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			// enforcement_level
			{
				Config:      testAccTFEOrganizationRunTaskGlobalSettings_parameters("", `["pre_plan"]`),
				ExpectError: regexp.MustCompile(`Attribute enforcement_level value must be one of: \[.*\]`),
			},
			{
				Config:      testAccTFEOrganizationRunTaskGlobalSettings_parameters("bad name", `["pre_plan"]`),
				ExpectError: regexp.MustCompile(`Attribute enforcement_level value must be one of: \[.*\]`),
			},
			// stages
			{
				Config:      testAccTFEOrganizationRunTaskGlobalSettings_parameters(string(tfe.Mandatory), `[]`),
				ExpectError: regexp.MustCompile(`Attribute stages list must contain at least 1 elements.*`),
			},
			{
				Config:      testAccTFEOrganizationRunTaskGlobalSettings_parameters(string(tfe.Mandatory), `["pre_plan","BADWOLF","post_plan"]`),
				ExpectError: regexp.MustCompile(`Attribute stages\[1\] value must be.*`),
			},
			{
				Config:      testAccTFEOrganizationRunTaskGlobalSettings_parameters(string(tfe.Mandatory), `["pre_plan","pre_plan","pre_plan"]`),
				ExpectError: regexp.MustCompile(`Error: Duplicate List Value`),
			},
		},
	})
}

func TestAccTFEOrganizationRunTaskGlobalSettings_create(t *testing.T) {
	skipUnlessRunTasksDefined(t)

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEOrganizationRunTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrganizationRunTaskGlobalSettings_basic(org.Name, rInt, runTasksURL(), runTasksHMACKey()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOrganizationRunTaskGlobalEnabled("tfe_organization_run_task.foobar", true),
					resource.TestCheckResourceAttr("tfe_organization_run_task_global_settings.sut", "enabled", "true"),
					resource.TestCheckResourceAttr("tfe_organization_run_task_global_settings.sut", "enforcement_level", "mandatory"),
					resource.TestCheckResourceAttr("tfe_organization_run_task_global_settings.sut", "stages.#", "1"),
					resource.TestCheckResourceAttr("tfe_organization_run_task_global_settings.sut", "stages.0", "post_plan"),
				),
			},
			{
				Config: testAccTFEOrganizationRunTaskGlobalSettings_update(org.Name, rInt, runTasksURL(), runTasksHMACKey()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOrganizationRunTaskGlobalEnabled("tfe_organization_run_task.foobar", false),
					resource.TestCheckResourceAttr("tfe_organization_run_task_global_settings.sut", "enabled", "false"),
					resource.TestCheckResourceAttr("tfe_organization_run_task_global_settings.sut", "enforcement_level", "advisory"),
					resource.TestCheckResourceAttr("tfe_organization_run_task_global_settings.sut", "stages.#", "2"),
					resource.TestCheckResourceAttr("tfe_organization_run_task_global_settings.sut", "stages.0", "pre_plan"),
					resource.TestCheckResourceAttr("tfe_organization_run_task_global_settings.sut", "stages.1", "post_plan"),
				),
			},
		},
	})
}

func TestAccTFEOrganizationRunTaskGlobalSettings_createUnsupported(t *testing.T) {
	skipUnlessRunTasksDefined(t)

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createTrialOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEOrganizationRunTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFEOrganizationRunTaskGlobalSettings_basic(org.Name, rInt, runTasksURL(), runTasksHMACKey()),
				ExpectError: regexp.MustCompile(`Error: Organization does not support global run tasks`),
			},
		},
	})
}

func TestAccTFEOrganizationRunTaskGlobalSettings_import(t *testing.T) {
	skipUnlessRunTasksDefined(t)

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFETeamAccessDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrganizationRunTaskGlobalSettings_basic(org.Name, rInt, runTasksURL(), runTasksHMACKey()),
			},
			{
				ResourceName:      "tfe_organization_run_task_global_settings.sut",
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s/foobar-task-%d", org.Name, rInt),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTFEOrganizationRunTaskGlobalSettings_Read(t *testing.T) {
	skipUnlessRunTasksDefined(t)

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)
	key := runTasksHMACKey()
	task := createRunTask(t, tfeClient, org.Name, tfe.RunTaskCreateOptions{
		Name:    fmt.Sprintf("tst-task-%s", randomString(t)),
		URL:     runTasksURL(),
		HMACKey: &key,
	})

	org_tf := fmt.Sprintf(`data "tfe_organization" "orgtask" { name = %q }`, org.Name)

	create_settings_tf := fmt.Sprintf(`
		%s
		resource "tfe_organization_run_task_global_settings" "sut" {
			task_id = %q

			enabled           = true
			enforcement_level = "mandatory"
			stages            = ["post_plan"]
		}
		`, org_tf, task.ID)

	delete_task_settings := func() {
		_, err := tfeClient.RunTasks.Update(ctx, task.ID, tfe.RunTaskUpdateOptions{
			Global: &tfe.GlobalRunTaskOptions{
				Enabled: tfe.Bool(false),
			},
		})
		if err != nil {
			t.Fatalf("Error updating task: %s", err)
		}
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEOrganizationRunTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: create_settings_tf,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tfe_organization_run_task_global_settings.sut", "enabled", "true"),
				),
			},
			{
				// Delete the created run task settings and ensure we can re-create it
				PreConfig: delete_task_settings,
				Config:    create_settings_tf,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tfe_organization_run_task_global_settings.sut", "enabled", "true"),
				),
			},
			{
				// Delete the created run task settings and ensure we can ignore it if we no longer need to manage it
				PreConfig: delete_task_settings,
				Config:    org_tf,
				Check: resource.ComposeTestCheckFunc(
					testCheckResourceNotExist("tfe_organization_run_task_global_settings.sut"),
				),
			},
		},
	})
}

func testAccCheckTFEOrganizationRunTaskGlobalEnabled(resourceName string, expectedEnabled bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}
		rt, err := testAccConfiguredClient.Client.RunTasks.Read(ctx, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error reading Run Task: %w", err)
		}

		if rt == nil {
			return fmt.Errorf("Organization Run Task not found")
		}

		if rt.Global == nil {
			return fmt.Errorf("Organization Run Task exists but does not support global run tasks")
		}

		if rt.Global.Enabled != expectedEnabled {
			return fmt.Errorf("Task expected a global enabled value of %t, got %t", expectedEnabled, rt.Global.Enabled)
		}

		return nil
	}
}

func testAccTFEOrganizationRunTaskGlobalSettings_basic(orgName string, rInt int, runTaskURL, runTaskHMACKey string) string {
	return fmt.Sprintf(`
resource "tfe_organization_run_task" "foobar" {
	organization = "%s"
	url          = "%s"
	name         = "foobar-task-%d"
	enabled      = false
	hmac_key     = "%s"
}

resource "tfe_organization_run_task_global_settings" "sut" {
  task_id = tfe_organization_run_task.foobar.id

  enabled           = true
  enforcement_level = "mandatory"
  stages            = ["post_plan"]
}
`, orgName, runTaskURL, rInt, runTaskHMACKey)
}

func testAccTFEOrganizationRunTaskGlobalSettings_parameters(enforceLevel, stages string) string {
	return fmt.Sprintf(`
resource "tfe_organization_run_task" "foobar" {
	organization = "foo"
	url          = "http://somewhere.local"
	name         = "task_name"
	enabled      = false
	hmac_key     = "something"
}

resource "tfe_organization_run_task_global_settings" "sut" {
  task_id = tfe_organization_run_task.foobar.id

  enabled           = true
  enforcement_level = "%s"
  stages            = %s
}
`, enforceLevel, stages)
}

func testAccTFEOrganizationRunTaskGlobalSettings_update(orgName string, rInt int, runTaskURL, runTaskHMACKey string) string {
	return fmt.Sprintf(`
	resource "tfe_organization_run_task" "foobar" {
		organization = "%s"
		url          = "%s"
		name         = "foobar-task-%d-new"
		enabled      = true
		hmac_key     = "%s"
		description  = "a description"
	}

	resource "tfe_organization_run_task_global_settings" "sut" {
		task_id = tfe_organization_run_task.foobar.id

		enabled           = false
		enforcement_level = "advisory"
		stages            = ["pre_plan", "post_plan"]
	}
`, orgName, runTaskURL, rInt, runTaskHMACKey)
}
