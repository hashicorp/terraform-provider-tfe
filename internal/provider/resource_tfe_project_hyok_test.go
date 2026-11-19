// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"testing"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/jsonapi"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// testAccHYOKProjectDefaultEnv returns an organization with HYOK configurations by
// workspace enabled and an available HYOK configuration in it, skipping the test
// when either is not set. The token must belong to an organization owner.
func testAccHYOKProjectDefaultEnv(t *testing.T) (orgName, hyokConfigID string) {
	t.Helper()
	skipUnlessHYOKEnabled(t)

	orgName = os.Getenv("HYOK_ORGANIZATION_NAME")
	hyokConfigID = os.Getenv("HYOK_AVAILABLE_CONFIGURATION_ID")
	if orgName == "" || hyokConfigID == "" {
		t.Skip("HYOK_ORGANIZATION_NAME and HYOK_AVAILABLE_CONFIGURATION_ID environment variables must be set to run this test")
	}
	return orgName, hyokConfigID
}

func TestAccTFEProject_defaultHYOKConfiguration(t *testing.T) {
	orgName, hyokConfigID := testAccHYOKProjectDefaultEnv(t)
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	resourceName := "tfe_project.foobar"

	var projectID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEProjectDestroy,
		Steps: []resource.TestStep{
			// Create with a default; the next plan must be empty.
			{
				Config: testAccTFEProject_defaultHYOK(orgName, rInt, hyokConfigID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "default_hyok_configuration_id", hyokConfigID),
					func(s *terraform.State) error {
						projectID = s.RootModule().Resources[resourceName].Primary.ID
						return nil
					},
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
			// Import reads the default back.
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Drift: clear the default outside Terraform; the plan restores it in place.
			{
				PreConfig: func() {
					_, err := testAccConfiguredClient.Client.Projects.Update(context.Background(), projectID, tfe.ProjectUpdateOptions{
						DefaultHYOKConfigurationID: jsonapi.NewNullNullableAttr[string](),
					})
					if err != nil {
						t.Fatalf("error clearing default HYOK configuration outside Terraform: %v", err)
					}
				},
				Config: testAccTFEProject_defaultHYOK(orgName, rInt, hyokConfigID),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.TestCheckResourceAttr(resourceName, "default_hyok_configuration_id", hyokConfigID),
			},
			// Removing the argument clears the default without replacing the project.
			{
				Config: testAccTFEProject_withoutDefaultHYOK(orgName, rInt),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(resourceName, "default_hyok_configuration_id"),
					resource.TestCheckResourceAttrWith(resourceName, "id", func(id string) error {
						if id != projectID {
							return fmt.Errorf("project was replaced: id %s, want %s", id, projectID)
						}
						return nil
					}),
				),
			},
		},
	})
}

func TestAccTFEProjectDataSource_defaultHYOKConfiguration(t *testing.T) {
	orgName, hyokConfigID := testAccHYOKProjectDefaultEnv(t)
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEProject_defaultHYOK(orgName, rInt, hyokConfigID) + `
data "tfe_project" "foobar" {
  name         = tfe_project.foobar.name
  organization = tfe_project.foobar.organization
}
`,
				Check: resource.TestCheckResourceAttr("data.tfe_project.foobar", "default_hyok_configuration_id", hyokConfigID),
			},
		},
	})
}

// TestAccTFEProject_defaultHYOKConfigurationFlagOff assumes that HYOK
// configurations by workspace is off by default for new organizations, so the
// server rejects the selection with a 422 before it looks up the ID.
func TestAccTFEProject_defaultHYOKConfigurationFlagOff(t *testing.T) {
	skipUnlessHYOKEnabled(t)

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}
	org, orgCleanup := createPremiumOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFEProject_defaultHYOK(org.Name, rInt, "hyokc-doesnotexist"),
				ExpectError: regexp.MustCompile(`HYOK configuration selection is not available for this\s+organization`),
			},
		},
	})
}

func testAccTFEProject_defaultHYOK(orgName string, rInt int, hyokConfigID string) string {
	return fmt.Sprintf(`
resource "tfe_project" "foobar" {
  organization                  = %q
  name                          = "projecttest-hyok-%d"
  default_hyok_configuration_id = %q
}
`, orgName, rInt, hyokConfigID)
}

func testAccTFEProject_withoutDefaultHYOK(orgName string, rInt int) string {
	return fmt.Sprintf(`
resource "tfe_project" "foobar" {
  organization = %q
  name         = "projecttest-hyok-%d"
}
`, orgName, rInt)
}
