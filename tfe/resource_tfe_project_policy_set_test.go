// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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

func TestAccTFEProjectPolicySet_basic(t *testing.T) {
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
		CheckDestroy: testAccCheckTFEProjectPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEProjectPolicySet_basic(org.Name, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEProjectPolicySetExists(
						"tfe_project_policy_set.test"),
				),
			},
			{
				ResourceName:      "tfe_project_policy_set.test",
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s/tst-terraform-%d/tst-policy-set-%d", org.Name, rInt, rInt),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTFEProjectPolicySet_incorrectImportSyntax(t *testing.T) {
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
				Config: testAccTFEProjectPolicySet_basic(org.Name, rInt),
			},
			{
				ResourceName:  "tfe_project_policy_set.test",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s/tst-terraform-%d", org.Name, rInt),
				ExpectError:   regexp.MustCompile(`Error: invalid project policy set input format`),
			},
		},
	})
}

func testAccCheckTFEProjectPolicySetExists(n string) resource.TestCheckFunc {
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

		projectID := rs.Primary.Attributes["project_id"]
		if projectID == "" {
			return fmt.Errorf("No project id set")
		}

		policySet, err := config.Client.PolicySets.ReadWithOptions(ctx, policySetID, &tfe.PolicySetReadOptions{
			Include: []tfe.PolicySetIncludeOpt{tfe.PolicySetProjects},
		})
		if err != nil {
			return fmt.Errorf("error reading polciy set %s: %w", policySetID, err)
		}
		for _, project := range policySet.Projects {
			if project.ID == projectID {
				return nil
			}
		}

		return fmt.Errorf("Project (%s) is not attached to policy set (%s).", projectID, policySetID)
	}
}

func testAccCheckTFEProjectPolicySetDestroy(s *terraform.State) error {
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

func testAccTFEProjectPolicySet_basic(orgName string, rInt int) string {
	return fmt.Sprintf(`
	resource "tfe_project" "test" {
		name         = "tst-terraform-%d"
		organization = "%s"
	}

	resource "tfe_policy_set" "test" {
		name         = "tst-policy-set-%d"
		description  = "Policy Set"
		organization = "%s"
	}

	resource "tfe_project_policy_set" "test" {
		policy_set_id = tfe_policy_set.test.id
		project_id  = tfe_project.test.id
	}`, rInt, orgName, rInt, orgName)
}
