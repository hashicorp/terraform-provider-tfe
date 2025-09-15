// Copyright (c) HashiCorp, Inc.
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

func testAccCheckTFEDataRetentionPolicyExists(
	n string, policy *tfe.DataRetentionPolicyChoice) resource.TestCheckFunc {
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
			ws, err := testAccConfiguredClient.Client.Workspaces.ReadByID(ctx, wsID)
			if err != nil {
				return fmt.Errorf(
					"Error retrieving workspace %s: %w", wsID, err)
			}

			drp, err := testAccConfiguredClient.Client.Workspaces.ReadDataRetentionPolicyChoice(ctx, ws.ID)
			if err != nil {
				return fmt.Errorf(
					"Error retrieving data retention policy for workspace %s: %w", ws.ID, err)
			}

			*policy = *drp
		} else {
			orgName := rs.Primary.Attributes["organization"]

			drp, err := testAccConfiguredClient.Client.Organizations.ReadDataRetentionPolicyChoice(ctx, orgName)
			if err != nil {
				return fmt.Errorf(
					"Error retrieving data retention policy for organization %s: %w", orgName, err)
			}

			*policy = *drp
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

		dataRetentionPolicy, err := testAccConfiguredClient.Client.Workspaces.ReadDataRetentionPolicyChoice(ctx, rs.Primary.Attributes["workspace_id"])
		if err == nil {
			if dataRetentionPolicy.DataRetentionPolicyDeleteOlder != nil {
				return fmt.Errorf("data retention policy %s still exists", dataRetentionPolicy.DataRetentionPolicyDeleteOlder.ID)
			}
			if dataRetentionPolicy.DataRetentionPolicyDontDelete != nil {
				return fmt.Errorf("data retention policy %s still exists", dataRetentionPolicy.DataRetentionPolicyDontDelete.ID)
			}
			return fmt.Errorf("data retention policy likely exists but couldn't be serialized")
		}
	}

	return nil
}
