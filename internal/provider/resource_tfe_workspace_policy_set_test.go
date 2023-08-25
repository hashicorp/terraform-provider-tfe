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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccTFEWorkspacePolicySet_basic(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEWorkspacePolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEWorkspacePolicySet_basic(org.Name, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspacePolicySetExists(
						"tfe_workspace_policy_set.test"),
				),
			},
			{
				ResourceName:      "tfe_workspace_policy_set.test",
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s/tst-terraform-%d/tst-policy-set-%d", org.Name, rInt, rInt),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTFEWorkspacePolicySet_incorrectImportSyntax(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEWorkspacePolicySet_basic(org.Name, rInt),
			},
			{
				ResourceName:  "tfe_workspace_policy_set.test",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s/tst-terraform-%d", org.Name, rInt),
				ExpectError:   regexp.MustCompile(`Error: invalid workspace policy set input format`),
			},
		},
	})
}

func testAccCheckTFEWorkspacePolicySetExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(ConfiguredClient)

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

		policySet, err := config.Client.PolicySets.ReadWithOptions(ctx, policySetID, &tfe.PolicySetReadOptions{
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
	config := testAccProvider.Meta().(ConfiguredClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_policy_set" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := config.Client.PolicySets.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Policy Set %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccTFEWorkspacePolicySet_basic(orgName string, rInt int) string {
	return fmt.Sprintf(`
	resource "tfe_workspace" "test" {
		name         = "tst-terraform-%d"
		organization = "%s"
		auto_apply   = true
		tag_names    = ["test"]
	}

	resource "tfe_policy_set" "test" {
		name         = "tst-policy-set-%d"
		description  = "Policy Set"
		organization = "%s"
	}

	resource "tfe_workspace_policy_set" "test" {
		policy_set_id = tfe_policy_set.test.id
		workspace_id  = tfe_workspace.test.id
	}`, rInt, orgName, rInt, orgName)
}
