package tfe

import (
	"fmt"
	"testing"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccTFERunTrigger_basic(t *testing.T) {
	runTrigger := &tfe.RunTrigger{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFERunTriggerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFERunTrigger_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFERunTriggerExists(
						"tfe_run_trigger.foobar", runTrigger),
					testAccCheckTFERunTriggerAttributes(runTrigger),
				),
			},
		},
	})
}

func TestAccTFERunTriggerImport(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFERunTriggerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFERunTrigger_basic,
			},

			{
				ResourceName:      "tfe_run_trigger.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckTFERunTriggerExists(n string, runTrigger *tfe.RunTrigger) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		tfeClient := testAccProvider.Meta().(*tfe.Client)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		rt, err := tfeClient.RunTriggers.Read(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}

		*runTrigger = *rt

		return nil
	}
}

func testAccCheckTFERunTriggerAttributes(runTrigger *tfe.RunTrigger) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		tfeClient := testAccProvider.Meta().(*tfe.Client)

		workspaceID := runTrigger.Workspace.ID
		workspace, _ := tfeClient.Workspaces.Read(ctx, "tst-terraform", "workspace-test")
		if workspace.ID != workspaceID {
			return fmt.Errorf("Wrong workspace: %v", workspace.ID)
		}

		sourceableID := runTrigger.Sourceable.ID
		sourceable, _ := tfeClient.Workspaces.Read(ctx, "tst-terraform", "sourceable-test")
		if sourceable.ID != sourceableID {
			return fmt.Errorf("Wrong sourceable: %v", sourceable.ID)
		}

		return nil
	}
}

func testAccCheckTFERunTriggerDestroy(s *terraform.State) error {
	tfeClient := testAccProvider.Meta().(*tfe.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_run_trigger" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := tfeClient.RunTriggers.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Notification configuration %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

const testAccTFERunTrigger_basic = `
resource "tfe_organization" "foobar" {
  name  = "tst-terraform"
  email = "admin@company.com"
}

resource "tfe_workspace" "workspace" {
  name         = "workspace-test"
  organization = "${tfe_organization.foobar.id}"
}

resource "tfe_workspace" "sourceable" {
  name         = "sourceable-test"
  organization = "${tfe_organization.foobar.id}"
}

resource "tfe_run_trigger" "foobar" {
  workspace_external_id = "${tfe_workspace.workspace.external_id}"
  sourceable_id         = "${tfe_workspace.sourceable.external_id}"
}`
