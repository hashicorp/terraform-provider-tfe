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

func TestAccTFEOrganizationDefaultExecutionMode_remote(t *testing.T) {
	org := &tfe.Organization{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEOrganizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrganizationDefaultExecutionMode_remote(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOrganizationExists(
						"tfe_organization.foobar", org),
					testAccCheckTFEOrganizationDefaultExecutionMode(org, "remote"),
				),
			},
		},
	})
}

func TestAccTFEOrganizationDefaultExecutionMode_local(t *testing.T) {
	org := &tfe.Organization{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEOrganizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrganizationDefaultExecutionMode_local(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOrganizationExists(
						"tfe_organization.foobar", org),
					testAccCheckTFEOrganizationDefaultExecutionMode(org, "local"),
				),
			},
		},
	})
}

func TestAccTFEOrganizationDefaultExecutionMode_agent(t *testing.T) {
	org := &tfe.Organization{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEOrganizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrganizationDefaultExecutionMode_agent(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOrganizationExists(
						"tfe_organization.foobar", org),
					testAccCheckTFEOrganizationDefaultExecutionMode(org, "agent"),
					testAccCheckTFEOrganizationDefaultAgentPoolIDExists(org),
				),
			},
		},
	})
}

func TestAccTFEOrganizationDefaultExecutionMode_update(t *testing.T) {
	org := &tfe.Organization{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEOrganizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrganizationDefaultExecutionMode_remote(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOrganizationExists(
						"tfe_organization.foobar", org),
					testAccCheckTFEOrganizationDefaultExecutionMode(org, "remote"),
				),
			},
			{
				Config: testAccTFEOrganizationDefaultExecutionMode_agent(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOrganizationExists(
						"tfe_organization.foobar", org),
					testAccCheckTFEOrganizationDefaultExecutionMode(org, "agent"),
					testAccCheckTFEOrganizationDefaultAgentPoolIDExists(org),
				),
			},
			{
				Config: testAccTFEOrganizationDefaultExecutionMode_local(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOrganizationExists(
						"tfe_organization.foobar", org),
					testAccCheckTFEOrganizationDefaultExecutionMode(org, "local"),
				),
			},
			{
				Config: testAccTFEOrganizationDefaultExecutionMode_remote(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOrganizationExists(
						"tfe_organization.foobar", org),
					testAccCheckTFEOrganizationDefaultExecutionMode(org, "remote"),
				),
			},
		},
	})
}

func TestAccTFEOrganizationDefaultExecutionMode_import(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEOrganizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrganizationDefaultExecutionMode_remote(rInt),
			},

			{
				ResourceName:      "tfe_organization_default_execution_mode.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckTFEOrganizationDefaultExecutionMode(org *tfe.Organization, expectedExecutionMode string) resource.TestCheckFunc {
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

func testAccTFEOrganizationDefaultExecutionMode_remote(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_organization_default_execution_mode" "foobar" {
  organization = tfe_organization.foobar.name
  default_execution_mode = "remote"
}`, rInt)
}

func testAccTFEOrganizationDefaultExecutionMode_local(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_organization_default_execution_mode" "foobar" {
  organization = tfe_organization.foobar.name
  default_execution_mode = "local"
}`, rInt)
}

func testAccTFEOrganizationDefaultExecutionMode_agent(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_agent_pool" "foobar" {
  name = "agent-pool-test"
  organization = tfe_organization.foobar.name
}

resource "tfe_organization_default_execution_mode" "foobar" {
  organization = tfe_organization.foobar.name
  default_execution_mode = "agent"
  default_agent_pool_id = tfe_agent_pool.foobar.id
}`, rInt)
}
