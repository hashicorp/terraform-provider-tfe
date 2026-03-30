package provider

import (
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"testing"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccTFEProjectPolicySetExclusion_basic(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	param := modelProjectPolicySetExclusionParameter{}

	t.Cleanup(orgCleanup)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccProjectPolicySetExclusionSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEProjectPolicySetExclusion_basic(org.Name, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEProjectPolicySetExclusionExists(
						"tfe_project_policy_set_exclusion.test", &param),
				),
			},
			{
				ResourceName:      "tfe_project_policy_set_exclusion.test",
				ImportState:       true,
				ImportStateIdFunc: testImportIdFunc(),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTFEProjectPolicySetExclusion_incorrectImportSyntax(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEProjectPolicySetExclusion_basic(org.Name, rInt),
			},
			{
				ResourceName:  "tfe_project_policy_set_exclusion.test",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("policy_set_test/%s", "tst-terraform-project-"+fmt.Sprint(rInt)),
				ExpectError:   regexp.MustCompile(`Error: Project Not Found During Import`),
			},
		},
	})
}

func testAccCheckTFEProjectPolicySetExclusionExists(resourceName string, param *modelProjectPolicySetExclusionParameter) resource.TestCheckFunc {
	return func(s *terraform.State) (err error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		policySetID := rs.Primary.Attributes["policy_set_id"]
		if policySetID == "" {
			return fmt.Errorf("policy_set_id is not set")
		}
		projectID := rs.Primary.Attributes["project_id"]
		if projectID == "" {
			return fmt.Errorf("project_id is not set")
		}

		policySet, err := testAccConfiguredClient.Client.PolicySets.ReadWithOptions(ctx, policySetID, &tfe.PolicySetReadOptions{
			Include: []tfe.PolicySetIncludeOpt{tfe.PolicySetProjectExclusions},
		})

		if err != nil {
			return fmt.Errorf("error reading policy set with id %s: %w", policySetID, err)
		}

		for _, projectExclusion := range policySet.ProjectExclusions {
			if projectExclusion.ID == projectID {
				param.ID = types.StringValue(rs.Primary.ID)
				param.PolicySetID = types.StringValue(policySetID)
				param.ProjectID = types.StringValue(projectID)

				return nil
			}
		}

		return fmt.Errorf("project with id %s is not excluded from policy set with id %s", projectID, policySetID)
	}
}

func testAccProjectPolicySetExclusionSetDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_project_policy_set_exclusion" {
			continue
		}

		policySetID := rs.Primary.Attributes["policy_set_id"]
		projectID := rs.Primary.Attributes["project_id"]

		policySet, err := testAccConfiguredClient.Client.PolicySets.ReadWithOptions(ctx, policySetID, &tfe.PolicySetReadOptions{
			Include: []tfe.PolicySetIncludeOpt{tfe.PolicySetProjectExclusions},
		})
		if err != nil {
			if errors.Is(err, tfe.ErrResourceNotFound) {
				continue
			}
			return fmt.Errorf("error reading policy set with id %s: %w", policySetID, err)
		}

		for _, projectExclusion := range policySet.ProjectExclusions {
			if projectExclusion.ID == projectID {
				return fmt.Errorf("project with id %s is still excluded from policy set with id %s", projectID, policySetID)
			}
		}
	}
	return nil
}

func testAccTFEProjectPolicySetExclusion_basic(orgName string, rInt int) string {
	return fmt.Sprintf(`
		resource "tfe_project" "test_project" {
			name         = "tst-terraform-%d"
			organization = "%s"
		}
		
		resource "tfe_policy_set" "test_policy_set" {
			name         = "tst-terraform-policy-set-%d"
			description  = "a test policy set"
			global       = true
			organization = "%s"
		}
			
		resource "tfe_project_policy_set_exclusion" "test" {
			project_id    = tfe_project.test_project.id
			policy_set_id = tfe_policy_set.test_policy_set.id
		}
	`, rInt, orgName, rInt, orgName)
}

func testImportIdFunc() resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		res := s.RootModule().Resources
		if res == nil {
			return "", fmt.Errorf("resource not found in state: %s", "tfe_project_policy_set_exclusion.test")
		}

		policySet := res["tfe_policy_set.test_policy_set"]
		project := res["tfe_project.test_project"]

		return fmt.Sprintf("%s/%s", project.Primary.ID, policySet.Primary.ID), nil
	}
}
