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

func TestAccTFEVariableSetWorkspaceAttachment_basic(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEVariableSetWorkspaceAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEVariableSetWorkspaceAttachment_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEVariableSetWorkspaceAttachmentExists(
						"tfe_variable_set_workspace_attachment.test"),
				),
			},
			{
				ResourceName:        "tfe_variable_set_workspace_attachment.test",
				ImportState:         true,
				ImportStateIdPrefix: "",
				ImportStateVerify:   true,
			},
		},
	})
}

func testAccCheckTFEVariableSetWorkspaceAttachmentExists(
	n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		tfeClient := testAccProvider.Meta().(*tfe.Client)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		id := rs.Primary.ID

		if id == "" {
			return fmt.Errorf("No ID is set")
		}
		vSId, wId, err := DecodeVariableSetWorkspaceAttachment(id)

		vS, err := tfeClient.VariableSets.Read(ctx, vSId, &tfe.VariableSetReadOptions{
			Include: &[]tfe.VariableSetIncludeOpt{tfe.VariableSetWorkspaces},
		})
		if err != nil {
			return err
		}
		for _, workspace := range vS.Workspaces {
			if workspace.ID == wId {
				return nil
			}
		}

		return fmt.Errorf("Workspace (%s) is not attached to variable set (%s).", wId, vSId)
	}
}

func testAccCheckTFEVariableSetWorkspaceAttachmentDestroy(s *terraform.State) error {
	tfeClient := testAccProvider.Meta().(*tfe.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_variable_set" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := tfeClient.VariableSets.Read(ctx, rs.Primary.ID, nil)
		if err == nil {
			return fmt.Errorf("Variable Set %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccTFEVariableSetWorkspaceAttachment_base(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "test" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "test" {
  name         = "tst-terraform-%d"
  organization = tfe_organization.test.id
  auto_apply   = true
  tag_names    = ["test"]
}

resource "tfe_variable_set" "test" {
  name         = "variable_set_test"
  description  = "a test variable set"
  global       = false
  organization = tfe_organization.test.id
}
`, rInt, rInt)
}

func testAccTFEVariableSetWorkspaceAttachment_basic(rInt int) string {
	return testAccTFEVariableSetWorkspaceAttachment_base(rInt) + `
resource "tfe_variable_set_workspace_attachment" "test" {
  variable_set_id = tfe_variable_set.test.id
  workspace_id    = tfe_workspace.test.id
}`
}
