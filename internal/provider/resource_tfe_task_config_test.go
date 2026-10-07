// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/hashicorp/go-tfe/v2/api/models"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestModelFromTaskConfig(t *testing.T) {
	configID := "task-config-123"
	taskID := "task-123"
	ownerName := "example-org"
	enabled := true
	enforcementLevel := models.MANDATORY_TASKCONFIGS_ATTRIBUTES_ENFORCEMENTLEVEL

	attributes := models.NewTaskConfigs_attributes()
	attributes.SetGlobal(&enabled)
	attributes.SetEnforcementLevel(&enforcementLevel)
	attributes.SetAllowedStages([]models.TaskConfigs_attributes_allowedStages{
		models.PRE_PLAN_TASKCONFIGS_ATTRIBUTES_ALLOWEDSTAGES,
		models.POST_PLAN_TASKCONFIGS_ATTRIBUTES_ALLOWEDSTAGES,
	})

	ownerIdentifier := models.NewOrganizationsIdentifier()
	ownerIdentifier.SetId(&ownerName)
	owner := models.NewTaskConfigOwnerHasOne()
	owner.SetData(ownerIdentifier)
	taskIdentifier := models.NewTasksIdentifier()
	taskIdentifier.SetId(&taskID)
	task := models.NewTasksHasOne()
	task.SetData(taskIdentifier)
	relationships := models.NewTaskConfigs_relationships()
	relationships.SetOwner(owner)
	relationships.SetTask(task)

	config := models.NewTaskConfigs()
	config.SetId(&configID)
	config.SetAttributes(attributes)
	config.SetRelationships(relationships)

	result := modelFromTaskConfig(ctx, config, ownerName, taskID)
	if result.ID.ValueString() != configID {
		t.Fatalf("expected config ID %q, got %q", configID, result.ID.ValueString())
	}
	if result.Owner.ValueString() != ownerName || result.TaskID.ValueString() != taskID {
		t.Fatalf("unexpected owner/task relationship: owner=%q task_id=%q", result.Owner.ValueString(), result.TaskID.ValueString())
	}
	if !result.Enabled.ValueBool() || result.EnforcementLevel.ValueString() != "mandatory" {
		t.Fatalf("unexpected config attributes: enabled=%t enforcement_level=%q", result.Enabled.ValueBool(), result.EnforcementLevel.ValueString())
	}
	if stages := result.Stages.Elements(); len(stages) != 2 || stages[0].String() != `"pre_plan"` || stages[1].String() != `"post_plan"` {
		t.Fatalf("unexpected stages: %v", stages)
	}
}

func TestAccTFETaskConfig_createUpdateImport(t *testing.T) {
	skipUnlessRunTasksDefined(t)

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	config := testAccTFETaskConfig(org.Name, rInt, runTasksURL(), runTasksHMACKey(), true, "mandatory", `["post_plan"]`)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEOrganizationRunTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tfe_task_config.test", "owner", org.Name),
					resource.TestCheckResourceAttr("tfe_task_config.test", "enabled", "true"),
					resource.TestCheckResourceAttr("tfe_task_config.test", "enforcement_level", "mandatory"),
					resource.TestCheckResourceAttr("tfe_task_config.test", "stages.0", "post_plan"),
				),
			},
			{
				Config: testAccTFETaskConfig(org.Name, rInt, runTasksURL(), runTasksHMACKey(), true, "advisory", `["pre_plan","post_plan"]`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tfe_task_config.test", "enabled", "true"),
					resource.TestCheckResourceAttr("tfe_task_config.test", "enforcement_level", "advisory"),
					resource.TestCheckResourceAttr("tfe_task_config.test", "stages.#", "2"),
				),
			},
			{
				ResourceName:      "tfe_task_config.test",
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s/foobar-task-%d", org.Name, rInt),
				ImportStateVerify: true,
			},
			{
				Config: testAccTFETaskOnly(org.Name, rInt, runTasksURL(), runTasksHMACKey()),
				Check: resource.ComposeTestCheckFunc(
					testCheckResourceNotExist("tfe_task_config.test"),
					testAccCheckTFETaskConfigDisabled("tfe_organization_run_task.task"),
				),
			},
		},
	})
}

func testAccTFETaskConfig(owner string, rInt int, taskURL, hmacKey string, enabled bool, enforcementLevel, stages string) string {
	return fmt.Sprintf(`
resource "tfe_organization_run_task" "task" {
  organization = %q
  url          = %q
  name         = "foobar-task-%d"
  enabled      = false
  hmac_key     = %q
}

resource "tfe_task_config" "test" {
  owner            = %q
  task_id          = tfe_organization_run_task.task.id
  enabled          = %t
  enforcement_level = %q
  stages           = %s
}
`, owner, taskURL, rInt, hmacKey, owner, enabled, enforcementLevel, stages)
}

func testAccTFETaskOnly(owner string, rInt int, taskURL, hmacKey string) string {
	return fmt.Sprintf(`
resource "tfe_organization_run_task" "task" {
  organization = %q
  url          = %q
  name         = "foobar-task-%d"
  enabled      = false
  hmac_key     = %q
}
`, owner, taskURL, rInt, hmacKey)
}

func testAccCheckTFETaskConfigDisabled(taskResourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		task, ok := state.RootModule().Resources[taskResourceName]
		if !ok {
			return fmt.Errorf("task resource %q not found", taskResourceName)
		}
		owner := task.Primary.Attributes["organization"]
		if owner == "" {
			return fmt.Errorf("task resource %q has no owner", taskResourceName)
		}

		config, err := getOrganizationRunTaskConfig(ctx, testAccConfiguredClient.ClientV2, task.Primary.ID, owner)
		if err != nil {
			return fmt.Errorf("error reading task config after destroy: %w", err)
		}
		if config == nil || config.GetAttributes() == nil {
			return fmt.Errorf("task config was not found after destroy")
		}
		if valueOrZero(config.GetAttributes().GetGlobal()) {
			return fmt.Errorf("expected task config to be disabled after destroy")
		}
		return nil
	}
}
