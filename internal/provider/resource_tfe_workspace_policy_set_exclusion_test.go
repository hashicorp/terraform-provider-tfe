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

func TestAccTFEWorkspacePolicySetExclusion_basic(t *testing.T) {
	skipUnlessBeta(t)
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
		CheckDestroy: testAccCheckTFEWorkspacePolicySetExclusionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEWorkspacePolicySetExclusion_basic(org.Name, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspacePolicySetExclusionExists(
						"tfe_workspace_policy_set_exclusion.test"),
				),
			},
			{
				ResourceName:      "tfe_workspace_policy_set_exclusion.test",
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s/tst-terraform-%d/tst-policy-set-%d", org.Name, rInt, rInt),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTFEWorkspacePolicySetExclusion_incorrectImportSyntax(t *testing.T) {
	skipUnlessBeta(t)
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
				Config: testAccTFEWorkspacePolicySetExclusion_basic(org.Name, rInt),
			},
			{
				ResourceName:  "tfe_workspace_policy_set_exclusion.test",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s/tst-terraform-%d", org.Name, rInt),
				ExpectError:   regexp.MustCompile(`Error: invalid excluded workspace policy set input format`),
			},
		},
	})
}

func testAccCheckTFEWorkspacePolicySetExclusionExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(ConfiguredClient)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		id := rs.Primary.ID
		if id == "" {
			return fmt.Errorf("no ID is set")
		}

		policySetID := rs.Primary.Attributes["policy_set_id"]
		if policySetID == "" {
			return fmt.Errorf("no policy set id set")
		}

		workspaceExclusionsID := rs.Primary.Attributes["workspace_id"]
		if workspaceExclusionsID == "" {
			return fmt.Errorf("no excluded workspace id set")
		}

		policySet, err := config.Client.PolicySets.ReadWithOptions(ctx, policySetID, &tfe.PolicySetReadOptions{
			Include: []tfe.PolicySetIncludeOpt{tfe.PolicySetWorkspaces},
		})
		if err != nil {
			return fmt.Errorf("error reading polciy set %s: %w", policySetID, err)
		}
		for _, workspaceExclusion := range policySet.WorkspaceExclusions {
			if workspaceExclusion.ID == workspaceExclusionsID {
				return nil
			}
		}

		return fmt.Errorf("excluded workspace (%s) is not added to policy set (%s)", workspaceExclusionsID, policySetID)
	}
}

func testAccCheckTFEWorkspacePolicySetExclusionDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(ConfiguredClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_policy_set" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no instance ID is set")
		}

		_, err := config.Client.PolicySets.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("policy Set %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccTFEWorkspacePolicySetExclusion_basic(orgName string, rInt int) string {
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

	resource "tfe_workspace_policy_set_exclusion" "test" {
		policy_set_id = tfe_policy_set.test.id
		workspace_id  = tfe_workspace.test.id
	}`, rInt, orgName, rInt, orgName)
}
