// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccTFEWorkspaceHYOKEnabled_basic(t *testing.T) {
	skipUnlessHYOKEnabled(t)

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createPremiumOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEWorkspaceHYOKEnabledConfig(org.Name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTFEWorkspaceHYOKEnabledExists("tfe_workspace_hyok_enabled.test"),
					resource.TestCheckResourceAttrSet("tfe_workspace_hyok_enabled.test", "id"),
					resource.TestCheckResourceAttrSet("tfe_workspace_hyok_enabled.test", "workspace_id"),
					resource.TestCheckResourceAttrPair(
						"tfe_workspace_hyok_enabled.test", "id",
						"tfe_workspace.test", "id",
					),
				),
			},
			{
				ResourceName:      "tfe_workspace_hyok_enabled.test",
				ImportState:       true,
				ImportStateIdFunc: testAccTFEWorkspaceHYOKEnabledImportID("tfe_workspace_hyok_enabled.test"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTFEWorkspaceHYOKEnabled_destroy(t *testing.T) {
	skipUnlessHYOKEnabled(t)

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createPremiumOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	var workspaceID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEWorkspaceHYOKEnabledConfig(org.Name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTFEWorkspaceHYOKEnabledExists("tfe_workspace_hyok_enabled.test"),
					testAccCaptureWorkspaceID("tfe_workspace_hyok_enabled.test", &workspaceID),
				),
			},
			{
				Config: testAccTFEWorkspaceOnlyConfig(org.Name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTFEWorkspaceHYOKStillEnabled(tfeClient, &workspaceID),
				),
			},
		},
	})
}

func testAccCheckTFEWorkspaceHYOKEnabledExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no instance ID is set")
		}

		ws, err := testAccConfiguredClient.Client.Workspaces.ReadByID(ctx, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error reading workspace %s: %w", rs.Primary.ID, err)
		}

		if ws.HYOKEnabled == nil || !*ws.HYOKEnabled {
			return fmt.Errorf("expected HYOK to be enabled on workspace %s, but it is not", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckTFEWorkspaceHYOKStillEnabled(client *tfe.Client, workspaceID *string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		if workspaceID == nil || *workspaceID == "" {
			return fmt.Errorf("workspaceID was not captured")
		}

		ws, err := client.Workspaces.ReadByID(ctx, *workspaceID)
		if err != nil {
			return fmt.Errorf("error reading workspace %s after destroy: %w", *workspaceID, err)
		}

		if ws.HYOKEnabled == nil || !*ws.HYOKEnabled {
			return fmt.Errorf("expected HYOK to still be enabled on workspace %s after resource destroy, but it is not", *workspaceID)
		}

		return nil
	}
}

func testAccCaptureWorkspaceID(n string, dst *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		*dst = rs.Primary.ID
		return nil
	}
}

func testAccTFEWorkspaceHYOKEnabledImportID(n string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return "", fmt.Errorf("not found: %s", n)
		}
		return rs.Primary.ID, nil
	}
}

func testAccTFEWorkspaceHYOKEnabledConfig(orgName string) string {
	return fmt.Sprintf(`
resource "tfe_workspace" "test" {
  organization = "%s"
  name         = "tfe-provider-test-workspace-hyok"
}

resource "tfe_workspace_hyok_enabled" "test" {
  workspace_id = tfe_workspace.test.id
}
`, orgName)
}

func testAccTFEWorkspaceOnlyConfig(orgName string) string {
	return fmt.Sprintf(`
resource "tfe_workspace" "test" {
  organization = "%s"
  name         = "tfe-provider-test-workspace-hyok"
}
`, orgName)
}
