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

func TestAccTFEWorkspacePolicySet_basic(t *testing.T) {
	skipIfFreeOnly(t)

	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEWorkspacePolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEWorkspacePolicySet_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspacePolicySetExists(
						"tfe_workspace_policy_set.test"),
				),
			},
			{
				ResourceName:      "tfe_workspace_policy_set.test",
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("tst-terraform-%d/tst-terraform-%d/tst-policy-set-%d", rInt, rInt, rInt),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTFEWorkspacePolicySet_incorrectImportSyntax(t *testing.T) {
	skipIfFreeOnly(t)

	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEWorkspacePolicySet_basic(rInt),
			},
			{
				ResourceName:  "tfe_workspace_policy_set.test",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("tst-terraform-%d/tst-terraform-%d", rInt, rInt),
				ExpectError:   regexp.MustCompile(`Error: invalid workspace policy set input format`),
			},
		},
	})
}

func testAccCheckTFEWorkspacePolicySetExists(n string) resource.TestCheckFunc {
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

		policySetID := rs.Primary.Attributes["policy_set_id"]
		if policySetID == "" {
			return fmt.Errorf("No policy set id set")
		}

		workspaceID := rs.Primary.Attributes["workspace_id"]
		if workspaceID == "" {
			return fmt.Errorf("No workspace id set")
		}

		policySet, err := tfeClient.PolicySets.ReadWithOptions(ctx, policySetID, &tfe.PolicySetReadOptions{
			Include: []tfe.PolicySetIncludeOpt{tfe.PolicySetWorkspaces},
		})
		if err != nil {
			return fmt.Errorf("error reading polciy set %s: %w", policySetID, err)
		}
		for _, workspace := range policySet.Workspaces {
			if workspace.ID == workspaceID {
				return nil
			}
		}

		return fmt.Errorf("Workspace (%s) is not attached to policy set (%s).", workspaceID, policySetID)
	}
}

func testAccCheckTFEWorkspacePolicySetDestroy(s *terraform.State) error {
	tfeClient := testAccProvider.Meta().(*tfe.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_policy_set" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := tfeClient.PolicySets.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Policy Set %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccTFEWorkspacePolicySet_basic(rInt int) string {
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

	resource "tfe_policy_set" "test" {
		name         = "tst-policy-set-%d"
		description  = "Policy Set"
		organization = tfe_organization.test.id
	}

	resource "tfe_workspace_policy_set" "test" {
		policy_set_id = tfe_policy_set.test.id
		workspace_id  = tfe_workspace.test.id
	}`, rInt, rInt, rInt)
}
