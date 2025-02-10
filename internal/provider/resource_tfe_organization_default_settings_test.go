// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"errors"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccTFEOrganizationDefaultSettings_remote(t *testing.T) {
	org := &tfe.Organization{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEOrganizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrganizationDefaultSettings_remote(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOrganizationExists(
						"tfe_organization.foobar", org),
					testAccCheckTFEOrganizationDefaultSettings(org, "remote"),
				),
			},
		},
	})
}

func TestAccTFEOrganizationDefaultSettings_local(t *testing.T) {
	org := &tfe.Organization{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEOrganizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrganizationDefaultSettings_local(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOrganizationExists(
						"tfe_organization.foobar", org),
					testAccCheckTFEOrganizationDefaultSettings(org, "local"),
				),
			},
		},
	})
}

func TestAccTFEOrganizationDefaultSettings_agent(t *testing.T) {
	org := &tfe.Organization{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEOrganizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrganizationDefaultSettings_agent(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOrganizationExists(
						"tfe_organization.foobar", org),
					testAccCheckTFEOrganizationDefaultSettings(org, "agent"),
					testAccCheckTFEOrganizationDefaultAgentPoolIDExists(org),
				),
			},
		},
	})
}

func TestAccTFEOrganizationDefaultSettings_project(t *testing.T) {
	org := &tfe.Organization{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEOrganizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrganizationDefaultSettings_project(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOrganizationExists(
						"tfe_organization.foobar", org),
					testAccCheckTFEOrganizationDefaultProjectIDExists(org),
				),
			},
		},
	})
}

func TestAccTFEOrganizationDefaultSettings_update(t *testing.T) {
	org := &tfe.Organization{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEOrganizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrganizationDefaultSettings_remote(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOrganizationExists(
						"tfe_organization.foobar", org),
					testAccCheckTFEOrganizationDefaultSettings(org, "remote"),
				),
			},
			{
				Config: testAccTFEOrganizationDefaultSettings_agent(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOrganizationExists(
						"tfe_organization.foobar", org),
					testAccCheckTFEOrganizationDefaultSettings(org, "agent"),
					testAccCheckTFEOrganizationDefaultAgentPoolIDExists(org),
				),
			},
			{
				Config: testAccTFEOrganizationDefaultSettings_project(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOrganizationExists(
						"tfe_organization.foobar", org),
					testAccCheckTFEOrganizationDefaultProjectIDExists(org),
				),
			},
			{
				Config: testAccTFEOrganizationDefaultSettings_local(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOrganizationExists(
						"tfe_organization.foobar", org),
					testAccCheckTFEOrganizationDefaultSettings(org, "local"),
				),
			},
			{
				Config: testAccTFEOrganizationDefaultSettings_remote(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOrganizationExists(
						"tfe_organization.foobar", org),
					testAccCheckTFEOrganizationDefaultSettings(org, "remote"),
				),
			},
		},
	})
}

func TestAccTFEOrganizationDefaultSettings_import(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEOrganizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrganizationDefaultSettings_remote(rInt),
			},

			{
				ResourceName:      "tfe_organization_default_settings.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckTFEOrganizationDefaultSettings(org *tfe.Organization, expectedExecutionMode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if org.DefaultExecutionMode != expectedExecutionMode {
			return fmt.Errorf("default Execution Mode did not match, expected: %s, but was: %s", expectedExecutionMode, org.DefaultExecutionMode)
		}

		return nil
	}
}

func testAccCheckTFEOrganizationDefaultAgentPoolIDExists(org *tfe.Organization) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if org.DefaultAgentPool == nil {
			return errors.New("default agent pool was not set")
		}

		return nil
	}
}

func testAccCheckTFEOrganizationDefaultProjectIDExists(org *tfe.Organization) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if org.DefaultProject == nil {
			return errors.New("default project was not set")
		}

		return nil
	}
}

func testAccTFEOrganizationDefaultSettings_remote(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_organization_default_settings" "foobar" {
  organization = tfe_organization.foobar.name
  default_execution_mode = "remote"
}`, rInt)
}

func testAccTFEOrganizationDefaultSettings_local(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_organization_default_settings" "foobar" {
  organization = tfe_organization.foobar.name
  default_execution_mode = "local"
}`, rInt)
}

func testAccTFEOrganizationDefaultSettings_agent(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_agent_pool" "foobar" {
  name = "agent-pool-test"
  organization = tfe_organization.foobar.name
}

resource "tfe_organization_default_settings" "foobar" {
  organization = tfe_organization.foobar.name
  default_execution_mode = "agent"
  default_agent_pool_id = tfe_agent_pool.foobar.id
}`, rInt)
}

func testAccTFEOrganizationDefaultSettings_project(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_agent_pool" "foobar" {
  name = "agent-pool-test"
  organization = tfe_organization.foobar.name
}

resource "tfe_project" "foobar" {
  name = "project-test"
  organization = tfe_organization.foobar.name
}

resource "tfe_organization_default_settings" "foobar" {
  organization       = tfe_organization.foobar.name
  default_execution_mode = "agent"
  default_agent_pool_id = tfe_agent_pool.foobar.id
  default_project_id = tfe_project.foobar.id
}`, rInt)
}
