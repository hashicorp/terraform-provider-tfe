// Copyright IBM Corp. 2018, 2025
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

func TestAccTFERunTrigger_basic(t *testing.T) {
	var runTrigger models.RunTriggersable
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFERunTriggerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFERunTrigger_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFERunTriggerExists(
						"tfe_run_trigger.foobar", &runTrigger),
					testAccCheckTFERunTriggerAttributes(&runTrigger, orgName),
				),
			},
		},
	})
}

func TestAccTFERunTrigger_many(t *testing.T) {
	checks := make([]resource.TestCheckFunc, 10)
	for i := range checks {
		checks[i] = resource.TestCheckResourceAttrSet(fmt.Sprintf("tfe_run_trigger.foobar.%d", i), "id")
	}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFERunTriggerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFERunTrigger_many(rInt),
				Check:  resource.ComposeTestCheckFunc(checks...),
			},
		},
	})
}

func TestAccTFERunTriggerImport(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFERunTriggerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFERunTrigger_basic(rInt),
			},
			{
				ResourceName:      "tfe_run_trigger.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckTFERunTriggerExists(n string, runTrigger *models.RunTriggersable) resource.TestCheckFunc { //nolint:gocritic // runTrigger is an output param the caller reads after Check runs; RunTriggersable must stay addressable to be settable
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		rtEnvelope, err := testAccConfiguredClient.ClientV2.API.RunTriggers().ById(rs.Primary.ID).Get(ctx, nil)
		if err != nil {
			return err
		}
		if rtEnvelope == nil || rtEnvelope.GetData() == nil {
			return fmt.Errorf("Run trigger not found")
		}

		*runTrigger = rtEnvelope.GetData()

		return nil
	}
}

func testAccCheckTFERunTriggerAttributes(runTrigger *models.RunTriggersable, orgName string) resource.TestCheckFunc { //nolint:gocritic // runTrigger is populated by the paired Exists check at test-execution time; must stay a pointer so this reads that value, not a stale copy captured at construction time
	return func(s *terraform.State) error {
		relationships := (*runTrigger).GetRelationships()

		workspaceID := runTriggerWorkspaceID(relationships)
		workspace, _ := testAccConfiguredClient.Client.Workspaces.Read(ctx, orgName, "workspace-test")
		if workspace.ID != workspaceID {
			return fmt.Errorf("Wrong workspace: %v", workspace.ID)
		}

		sourceableID := runTriggerSourceableID(relationships)
		sourceable, _ := testAccConfiguredClient.Client.Workspaces.Read(ctx, orgName, "sourceable-test")
		if sourceable.ID != sourceableID {
			return fmt.Errorf("Wrong sourceable: %v", sourceable.ID)
		}

		return nil
	}
}

func testAccCheckTFERunTriggerDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_run_trigger" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := testAccConfiguredClient.ClientV2.API.RunTriggers().ById(rs.Primary.ID).Get(ctx, nil)
		if err == nil {
			return fmt.Errorf("Run trigger %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccTFERunTrigger_basic(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "workspace" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_workspace" "sourceable" {
  name         = "sourceable-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_run_trigger" "foobar" {
  workspace_id  = tfe_workspace.workspace.id
  sourceable_id = tfe_workspace.sourceable.id
}`, rInt)
}

func testAccTFERunTrigger_many(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "workspace" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_workspace" "sourceable" {
  count = 10

  name         = "sourceable-test-${count.index}"
  organization = tfe_organization.foobar.id
}

resource "tfe_run_trigger" "foobar" {
  count = 10

  workspace_id  = tfe_workspace.workspace.id
  sourceable_id = tfe_workspace.sourceable[count.index].id
}`, rInt)
}
