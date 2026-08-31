// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"math/rand"
	"testing"
	"time"

	"fmt"
	"os"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccTFEDataRetentionPolicy_basic(t *testing.T) {
	skipIfCloud(t)

	policy := &tfe.DataRetentionPolicyChoice{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEDataRetentionPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEDataRetentionPolicy_basic(rInt, 42),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEDataRetentionPolicyExists("tfe_data_retention_policy.foobar", policy),
					resource.TestCheckResourceAttr(
						"tfe_data_retention_policy.foobar", "delete_older_than.days", "42"),
				),
			},
			{
				Config: testAccTFEDataRetentionPolicy_basic(rInt, 1337),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEDataRetentionPolicyExists("tfe_data_retention_policy.foobar", policy),
					resource.TestCheckResourceAttr(
						"tfe_data_retention_policy.foobar", "delete_older_than.days", "1337"),
				),
			},
			{
				ResourceName:      "tfe_data_retention_policy.foobar",
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("tst-terraform-%d/workspace-test", rInt),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTFEDataRetentionPolicy_dontdelete_basic(t *testing.T) {
	skipIfCloud(t)

	policy := &tfe.DataRetentionPolicyChoice{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEDataRetentionPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEDataRetentionPolicy_dontdelete_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEDataRetentionPolicyExists("tfe_data_retention_policy.foobar", policy),
					resource.TestCheckResourceAttrSet("tfe_data_retention_policy.foobar", "dont_delete.%"),
				),
			},
			{
				ResourceName:      "tfe_data_retention_policy.foobar",
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("tst-terraform-%d/workspace-test", rInt),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTFEDataRetentionPolicy_explicit_organization(t *testing.T) {
	skipIfCloud(t)

	policy := &tfe.DataRetentionPolicyChoice{}
	orgName, _ := setupDefaultOrganization(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEDataRetentionPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEDataRetentionPolicy_explicit_organization(orgName, 42),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEDataRetentionPolicyExists("tfe_data_retention_policy.foobar", policy),
					resource.TestCheckResourceAttr(
						"tfe_data_retention_policy.foobar", "delete_older_than.days", "42"),
				),
			},
			{
				Config: testAccTFEDataRetentionPolicy_explicit_organization(orgName, 1337),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEDataRetentionPolicyExists("tfe_data_retention_policy.foobar", policy),
					resource.TestCheckResourceAttr(
						"tfe_data_retention_policy.foobar", "delete_older_than.days", "1337"),
					resource.TestCheckResourceAttr(
						"tfe_data_retention_policy.foobar", "organization", orgName),
				),
			},
			{
				ResourceName:      "tfe_data_retention_policy.foobar",
				ImportState:       true,
				ImportStateId:     orgName,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTFEDataRetentionPolicy_update_type(t *testing.T) {
	skipIfCloud(t)

	policy := &tfe.DataRetentionPolicyChoice{}
	defaultOrgName, _ := setupDefaultOrganization(t)

	os.Setenv("TFE_ORGANIZATION", defaultOrgName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEDataRetentionPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEDataRetentionPolicy_implicit_organization(42),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEDataRetentionPolicyExists("tfe_data_retention_policy.foobar", policy),
					resource.TestCheckResourceAttr(
						"tfe_data_retention_policy.foobar", "delete_older_than.days", "42"),
				),
			},
			{
				Config: testAccTFEDataRetentionPolicy_dontDelete_implicit_organization(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEDataRetentionPolicyExists("tfe_data_retention_policy.foobar", policy),
					resource.TestCheckResourceAttr(
						"tfe_data_retention_policy.foobar", "organization", defaultOrgName),
				),
			},
			{
				Config: testAccTFEDataRetentionPolicy_implicit_organization(42),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEDataRetentionPolicyExists("tfe_data_retention_policy.foobar", policy),
					resource.TestCheckResourceAttr(
						"tfe_data_retention_policy.foobar", "delete_older_than.days", "42"),
				),
			},
		},
	})
}

func TestAccTFEDataRetentionPolicy_implicit_organization(t *testing.T) {
	skipIfCloud(t)

	policy := &tfe.DataRetentionPolicyChoice{}
	defaultOrgName, _ := setupDefaultOrganization(t)

	os.Setenv("TFE_ORGANIZATION", defaultOrgName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEDataRetentionPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEDataRetentionPolicy_implicit_organization(42),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEDataRetentionPolicyExists("tfe_data_retention_policy.foobar", policy),
					resource.TestCheckResourceAttr(
						"tfe_data_retention_policy.foobar", "delete_older_than.days", "42"),
				),
			},
			{
				Config: testAccTFEDataRetentionPolicy_implicit_organization(1337),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEDataRetentionPolicyExists("tfe_data_retention_policy.foobar", policy),
					resource.TestCheckResourceAttr(
						"tfe_data_retention_policy.foobar", "organization", defaultOrgName),
				),
			},
			{
				ResourceName:      "tfe_data_retention_policy.foobar",
				ImportState:       true,
				ImportStateId:     defaultOrgName,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTFEDataRetentionPolicy_dontdelete_organization_level(t *testing.T) {
	skipIfCloud(t)

	policy := &tfe.DataRetentionPolicyChoice{}
	defaultOrgName, _ := setupDefaultOrganization(t)

	os.Setenv("TFE_ORGANIZATION", defaultOrgName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEDataRetentionPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEDataRetentionPolicy_dontDelete_implicit_organization(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEDataRetentionPolicyExists("tfe_data_retention_policy.foobar", policy),
					resource.TestCheckResourceAttr("tfe_data_retention_policy.foobar", "organization", defaultOrgName),
					resource.TestCheckResourceAttrSet("tfe_data_retention_policy.foobar", "dont_delete.%"),
				),
			},
			{
				Config: testAccTFEDataRetentionPolicy_dontDelete_explicit_organization(defaultOrgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEDataRetentionPolicyExists("tfe_data_retention_policy.foobar", policy),
					resource.TestCheckResourceAttr("tfe_data_retention_policy.foobar", "organization", defaultOrgName),
					resource.TestCheckResourceAttrSet("tfe_data_retention_policy.foobar", "dont_delete.%"),
				),
			},
			{
				ResourceName:      "tfe_data_retention_policy.foobar",
				ImportState:       true,
				ImportStateId:     defaultOrgName,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccTFEDataRetentionPolicy_basic(rInt int, deleteOlderThan int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_data_retention_policy" "foobar" {
  workspace_id = tfe_workspace.foobar.id
	
  delete_older_than {
    days = %d
  }
}`, rInt, deleteOlderThan)
}

func testAccTFEDataRetentionPolicy_dontdelete_basic(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_data_retention_policy" "foobar" {
  workspace_id = tfe_workspace.foobar.id
	
  dont_delete {}
}`, rInt)
}

func testAccTFEDataRetentionPolicy_explicit_organization(organization string, deleteOlderThan int) string {
	return fmt.Sprintf(`
resource "tfe_data_retention_policy" "foobar" {
  organization = "%s"
  delete_older_than {
    days = %d
  }
}`, organization, deleteOlderThan)
}

func testAccTFEDataRetentionPolicy_implicit_organization(deleteOlderThan int) string {
	return fmt.Sprintf(`
resource "tfe_data_retention_policy" "foobar" {
  delete_older_than {
    days = %d
  }
}`, deleteOlderThan)
}

func testAccTFEDataRetentionPolicy_dontDelete_explicit_organization(orgName string) string {
	return fmt.Sprintf(`
resource "tfe_data_retention_policy" "foobar" {
  organization = "%s"
  dont_delete {}
}`, orgName)
}

func testAccTFEDataRetentionPolicy_dontDelete_implicit_organization() string {
	return `
resource "tfe_data_retention_policy" "foobar" {
  dont_delete {}
}`
}

func TestAccTFEDataRetentionPolicy_keep_latest_count(t *testing.T) {
	skipIfCloud(t)

	policy := &tfe.DataRetentionPolicyChoice{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEDataRetentionPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEDataRetentionPolicy_keepLatestCount(rInt, 30, 5, 10, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEDataRetentionPolicyExists("tfe_data_retention_policy.foobar", policy),
					resource.TestCheckNoResourceAttr("tfe_data_retention_policy.foobar", "delete_older_than.days"),
					resource.TestCheckResourceAttr("tfe_data_retention_policy.foobar", "delete_older_than.state_versions_keep_latest_count", "5"),
					resource.TestCheckResourceAttr("tfe_data_retention_policy.foobar", "delete_older_than.configuration_versions_keep_latest_count", "10"),
					resource.TestCheckResourceAttr("tfe_data_retention_policy.foobar", "delete_older_than.run_data_keep_latest_count", "3"),
				),
			},
			{
				ResourceName:      "tfe_data_retention_policy.foobar",
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("tst-terraform-%d/workspace-test", rInt),
				ImportStateVerify: true,
			},
			{
				Config: testAccTFEDataRetentionPolicy_keepLatestCount(rInt, 30, 5, 10, 3),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr("tfe_data_retention_policy.foobar", "delete_older_than.days"),
					resource.TestCheckResourceAttr("tfe_data_retention_policy.foobar", "delete_older_than.state_versions_keep_latest_count", "5"),
				),
			},
		},
	})
}

func TestAccTFEDataRetentionPolicy_keep_latest_count_absent_fields_are_null(t *testing.T) {
	skipIfCloud(t)

	policy := &tfe.DataRetentionPolicyChoice{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEDataRetentionPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEDataRetentionPolicy_basic(rInt, 42),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEDataRetentionPolicyExists("tfe_data_retention_policy.foobar", policy),
					resource.TestCheckNoResourceAttr("tfe_data_retention_policy.foobar", "delete_older_than.state_versions_keep_latest_count"),
					resource.TestCheckNoResourceAttr("tfe_data_retention_policy.foobar", "delete_older_than.configuration_versions_keep_latest_count"),
					resource.TestCheckNoResourceAttr("tfe_data_retention_policy.foobar", "delete_older_than.run_data_keep_latest_count"),
				),
			},
		},
	})
}

func TestAccTFEDataRetentionPolicy_keep_latest_count_replacement(t *testing.T) {
	skipIfCloud(t)

	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEDataRetentionPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEDataRetentionPolicy_keepLatestCount(rInt, 30, 5, 10, 3),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tfe_data_retention_policy.foobar", "delete_older_than.state_versions_keep_latest_count", "5"),
				),
			},
			{
				Config: testAccTFEDataRetentionPolicy_keepLatestCount(rInt, 30, 7, 10, 3),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tfe_data_retention_policy.foobar", "delete_older_than.state_versions_keep_latest_count", "7"),
					resource.TestCheckResourceAttr("tfe_data_retention_policy.foobar", "delete_older_than.configuration_versions_keep_latest_count", "10"),
					resource.TestCheckResourceAttr("tfe_data_retention_policy.foobar", "delete_older_than.run_data_keep_latest_count", "3"),
				),
			},
		},
	})
}

func testAccTFEDataRetentionPolicy_keepLatestCount(rInt int, stateVersionsDays, stateKeep, cvKeep, runKeep int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_data_retention_policy" "foobar" {
  workspace_id = tfe_workspace.foobar.id

  delete_older_than {
    delete_state_versions                    = true
    delete_configuration_versions            = true
    delete_run_data_and_logs                 = true
    state_versions_delete_after_n_days       = %d
    state_versions_keep_latest_count         = %d
    configuration_versions_keep_latest_count = %d
    run_data_keep_latest_count               = %d
  }
}`, rInt, stateVersionsDays, stateKeep, cvKeep, runKeep)
}

func TestAccTFEDataRetentionPolicy_per_artifact_windows(t *testing.T) {
	skipIfCloud(t)

	policy := &tfe.DataRetentionPolicyChoice{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEDataRetentionPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEDataRetentionPolicy_perArtifactWindows(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEDataRetentionPolicyExists("tfe_data_retention_policy.foobar", policy),
					resource.TestCheckResourceAttr("tfe_data_retention_policy.foobar", "delete_older_than.delete_state_versions", "true"),
					resource.TestCheckResourceAttr("tfe_data_retention_policy.foobar", "delete_older_than.delete_configuration_versions", "true"),
					resource.TestCheckResourceAttr("tfe_data_retention_policy.foobar", "delete_older_than.delete_run_data_and_logs", "false"),
					resource.TestCheckResourceAttr("tfe_data_retention_policy.foobar", "delete_older_than.state_versions_delete_after_n_days", "30"),
					resource.TestCheckResourceAttr("tfe_data_retention_policy.foobar", "delete_older_than.configuration_versions_delete_after_n_days", "60"),
					resource.TestCheckNoResourceAttr("tfe_data_retention_policy.foobar", "delete_older_than.run_data_and_logs_delete_after_n_days"),
				),
			},
			{
				ResourceName:      "tfe_data_retention_policy.foobar",
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("tst-terraform-%d/workspace-test", rInt),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTFEDataRetentionPolicy_per_artifact_with_keep_latest(t *testing.T) {
	skipIfCloud(t)

	policy := &tfe.DataRetentionPolicyChoice{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEDataRetentionPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEDataRetentionPolicy_perArtifactWithKeepLatest(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEDataRetentionPolicyExists("tfe_data_retention_policy.foobar", policy),
					resource.TestCheckResourceAttr("tfe_data_retention_policy.foobar", "delete_older_than.delete_state_versions", "true"),
					resource.TestCheckResourceAttr("tfe_data_retention_policy.foobar", "delete_older_than.delete_run_data_and_logs", "true"),
					resource.TestCheckResourceAttr("tfe_data_retention_policy.foobar", "delete_older_than.state_versions_delete_after_n_days", "14"),
					resource.TestCheckResourceAttr("tfe_data_retention_policy.foobar", "delete_older_than.state_versions_keep_latest_count", "5"),
					resource.TestCheckResourceAttr("tfe_data_retention_policy.foobar", "delete_older_than.run_data_keep_latest_count", "3"),
				),
			},
		},
	})
}

func testAccTFEDataRetentionPolicy_perArtifactWindows(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_data_retention_policy" "foobar" {
  workspace_id = tfe_workspace.foobar.id

  delete_older_than {
    delete_state_versions                      = true
    delete_configuration_versions              = true
    delete_run_data_and_logs                   = false
    state_versions_delete_after_n_days         = 30
    configuration_versions_delete_after_n_days = 60
  }
}`, rInt)
}

func testAccTFEDataRetentionPolicy_perArtifactWithKeepLatest(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_data_retention_policy" "foobar" {
  workspace_id = tfe_workspace.foobar.id

  delete_older_than {
    delete_state_versions              = true
    delete_run_data_and_logs           = true
    state_versions_delete_after_n_days = 14
    state_versions_keep_latest_count   = 5
    run_data_keep_latest_count         = 3
  }
}`, rInt)
}

func testAccCheckTFEDataRetentionPolicyExists(
	n string, _ *tfe.DataRetentionPolicyChoice) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		wsID := rs.Primary.Attributes["workspace_id"]

		if wsID != "" {
			env, err := testAccConfiguredClient.ClientV2.API.Workspaces().ByWorkspace_id(wsID).Relationships().DataRetentionPolicy().Get(ctx, nil)
			if err != nil {
				return fmt.Errorf("Error retrieving data retention policy for workspace %s: %w", wsID, err)
			}
			if env.GetData() == nil {
				return fmt.Errorf("data retention policy not found for workspace %s", wsID)
			}
		} else {
			orgName := rs.Primary.Attributes["organization"]
			env, err := testAccConfiguredClient.ClientV2.API.Organizations().ByOrganization_name(orgName).Relationships().DataRetentionPolicy().Get(ctx, nil)
			if err != nil {
				return fmt.Errorf("Error retrieving data retention policy for organization %s: %w", orgName, err)
			}
			if env.GetData() == nil {
				return fmt.Errorf("data retention policy not found for organization %s", orgName)
			}
		}

		return nil
	}
}

func testAccCheckTFEDataRetentionPolicyDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_data_retention_policy" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		wsID := rs.Primary.Attributes["workspace_id"]
		orgName := rs.Primary.Attributes["organization"]

		if wsID != "" {
			dataRetentionPolicy, err := testAccConfiguredClient.Client.Workspaces.ReadDataRetentionPolicyChoice(ctx, wsID)
			if err == nil {
				if dataRetentionPolicy.DataRetentionPolicyDeleteOlder != nil {
					return fmt.Errorf("data retention policy %s still exists", dataRetentionPolicy.DataRetentionPolicyDeleteOlder.ID)
				}
				if dataRetentionPolicy.DataRetentionPolicyDontDelete != nil {
					return fmt.Errorf("data retention policy %s still exists", dataRetentionPolicy.DataRetentionPolicyDontDelete.ID)
				}
				return fmt.Errorf("data retention policy likely exists but couldn't be serialized")
			}
		} else if orgName != "" {
			dataRetentionPolicy, err := testAccConfiguredClient.Client.Organizations.ReadDataRetentionPolicyChoice(ctx, orgName)
			if err == nil && dataRetentionPolicy != nil {
				if dataRetentionPolicy.DataRetentionPolicyDeleteOlder != nil {
					return fmt.Errorf("data retention policy %s still exists", dataRetentionPolicy.DataRetentionPolicyDeleteOlder.ID)
				}
				if dataRetentionPolicy.DataRetentionPolicyDontDelete != nil {
					return fmt.Errorf("data retention policy %s still exists", dataRetentionPolicy.DataRetentionPolicyDontDelete.ID)
				}
			}
		}
	}

	return nil
}

func TestAccTFEDataRetentionPolicy_org_level_per_artifact_windows(t *testing.T) {
	skipIfCloud(t)

	policy := &tfe.DataRetentionPolicyChoice{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEDataRetentionPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEDataRetentionPolicy_orgLevelPerArtifactWindows(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEDataRetentionPolicyExists("tfe_data_retention_policy.foobar", policy),
					resource.TestCheckResourceAttr("tfe_data_retention_policy.foobar", "delete_older_than.delete_state_versions", "true"),
					resource.TestCheckResourceAttr("tfe_data_retention_policy.foobar", "delete_older_than.delete_configuration_versions", "true"),
					resource.TestCheckResourceAttr("tfe_data_retention_policy.foobar", "delete_older_than.state_versions_delete_after_n_days", "30"),
					resource.TestCheckResourceAttr("tfe_data_retention_policy.foobar", "delete_older_than.configuration_versions_delete_after_n_days", "60"),
					resource.TestCheckNoResourceAttr("tfe_data_retention_policy.foobar", "delete_older_than.days"),
				),
			},
			{
				ResourceName:      "tfe_data_retention_policy.foobar",
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("tst-terraform-%d", rInt),
				ImportStateVerify: true,
			},
		},
	})
}

func testAccTFEDataRetentionPolicy_orgLevelPerArtifactWindows(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_data_retention_policy" "foobar" {
  organization = tfe_organization.foobar.name

  delete_older_than {
    delete_state_versions                      = true
    delete_configuration_versions              = true
    state_versions_delete_after_n_days         = 30
    configuration_versions_delete_after_n_days = 60
  }
}`, rInt)
}

func TestAccTFEDataRetentionPolicy_update_granular_to_days(t *testing.T) {
	skipIfCloud(t)

	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEDataRetentionPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEDataRetentionPolicy_keepLatestCount(rInt, 30, 5, 10, 3),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr("tfe_data_retention_policy.foobar", "delete_older_than.days"),
					resource.TestCheckResourceAttr("tfe_data_retention_policy.foobar", "delete_older_than.state_versions_keep_latest_count", "5"),
				),
			},
			{
				Config: testAccTFEDataRetentionPolicy_basic(rInt, 42),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tfe_data_retention_policy.foobar", "delete_older_than.days", "42"),
					resource.TestCheckNoResourceAttr("tfe_data_retention_policy.foobar", "delete_older_than.state_versions_keep_latest_count"),
					resource.TestCheckNoResourceAttr("tfe_data_retention_policy.foobar", "delete_older_than.delete_state_versions"),
					resource.TestCheckNoResourceAttr("tfe_data_retention_policy.foobar", "delete_older_than.delete_configuration_versions"),
					resource.TestCheckNoResourceAttr("tfe_data_retention_policy.foobar", "delete_older_than.delete_run_data_and_logs"),
				),
			},
		},
	})
}
