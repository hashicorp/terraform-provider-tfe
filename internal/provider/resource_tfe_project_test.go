// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccTFEProject_basic(t *testing.T) {
	project := &tfe.Project{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEProject_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEProjectExists(
						"tfe_project.foobar", project),
					testAccCheckTFEProjectAttributes(project),
					resource.TestCheckResourceAttr(
						"tfe_project.foobar", "name", "projecttest"),
					resource.TestCheckResourceAttr(
						"tfe_project.foobar", "organization", fmt.Sprintf("tst-terraform-%d", rInt)),
				),
			},
		},
	})
}

func TestAccTFEProject_update(t *testing.T) {
	project := &tfe.Project{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEProject_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEProjectExists(
						"tfe_project.foobar", project),
					testAccCheckTFEProjectAttributes(project),
					resource.TestCheckResourceAttr(
						"tfe_project.foobar", "name", "projecttest"),
				),
			},
			{
				Config: testAccTFEProject_update(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEProjectExists(
						"tfe_project.foobar", project),
					testAccCheckTFEProjectAttributesUpdated(project),
					resource.TestCheckResourceAttr(
						"tfe_project.foobar", "name", "project updated"),
				),
			},
		},
	})
}

func TestAccTFEProject_import(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	project := &tfe.Project{}
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEProject_basic(rInt),
				Check: testAccCheckTFEProjectExists(
					"tfe_project.foobar", project),
			},

			{
				ResourceName:      "tfe_project.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "tfe_project.foobar",
				ImportState:       true,
				ImportStateId:     project.ID,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccTFEProject_update(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_project" "foobar" {
  organization = tfe_organization.foobar.name
  name = "project updated"
}`, rInt)
}

func testAccTFEProject_basic(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_project" "foobar" {
  organization = tfe_organization.foobar.name
  name = "projecttest"
}`, rInt)
}

func testAccCheckTFEProjectDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(ConfiguredClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_project" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := config.Client.Projects.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Project %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckTFEProjectExists(n string, project *tfe.Project) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(ConfiguredClient)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		p, err := config.Client.Projects.Read(ctx, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("unable to read project with ID %s", project.ID)
		}

		*project = *p

		return nil
	}
}

func testAccCheckTFEProjectAttributes(
	project *tfe.Project) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if project.Name != "projecttest" {
			return fmt.Errorf("Bad name: %s", project.Name)
		}

		return nil
	}
}

func testAccCheckTFEProjectAttributesUpdated(
	project *tfe.Project) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if project.Name != "project updated" {
			return fmt.Errorf("Bad name: %s", project.Name)
		}

		return nil
	}
}
