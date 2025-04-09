// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"math/rand"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/terraform"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTFEProject_basic(t *testing.T) {
	project := &tfe.Project{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEProjectDestroy,
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
						"tfe_project.foobar", "description", "project description"),
					resource.TestCheckResourceAttr(
						"tfe_project.foobar", "organization", fmt.Sprintf("tst-terraform-%d", rInt)),
				),
			},
		},
	})
}

func TestAccTFEProject_invalidName(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFEProject_invalidNameChar(rInt),
				ExpectError: regexp.MustCompile(`can only include letters, numbers, spaces, -, and _.`),
			},
			{
				Config:      testAccTFEProject_invalidNameLen(rInt),
				ExpectError: regexp.MustCompile(`string length must be between 3 and 40`),
			},
		},
	})
}

func TestAccTFEProject_update(t *testing.T) {
	project := &tfe.Project{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEProjectDestroy,
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
						"tfe_project.foobar", "description", "project description"),
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
					resource.TestCheckResourceAttr(
						"tfe_project.foobar", "description", "project description updated"),
				),
			},
			{
				Config: testAccTFEProject_updateRemoveBindings(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEProjectExists(
						"tfe_project.foobar", project),
					testAccCheckTFEProjectAttributesUpdated(project),
					resource.TestCheckResourceAttr(
						"tfe_project.foobar", "name", "project updated"),
					resource.TestCheckResourceAttr(
						"tfe_project.foobar", "description", "project description updated"),
				),
			},
		},
	})
}

func TestAccTFEProject_ignoreAdditionalTags(t *testing.T) {
	project := &tfe.Project{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEProject_ignoreAdditionalTags(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEProjectExists(
						"tfe_project.foobar", project),
					testAccCheckTFEProjectAttributes(project),
					resource.TestCheckResourceAttr(
						"tfe_project.foobar", "name", "projecttest"),
					resource.TestCheckResourceAttr(
						"tfe_project.foobar", "description", "project description"),
					resource.TestCheckResourceAttr(
						"tfe_project.foobar", "tags.%", "2"),
				),
			},
			{
				Config: testAccTFEProject_ignoreAdditionalTags(rInt),
				PreConfig: func() {
					organization := fmt.Sprintf("tst-terraform-%d", rInt)
					projects, err := testAccConfiguredClient.Client.Projects.List(ctx, organization, &tfe.ProjectListOptions{Name: "projecttest"})
					if err != nil {
						t.Fatalf("failed reading projecttest: %v", err)
					}
					if len(projects.Items) == 0 {
						t.Fatalf("expected to find projecttest, for %s", organization)
					}

					_, err = testAccConfiguredClient.Client.Projects.AddTagBindings(ctx, projects.Items[0].ID, tfe.ProjectAddTagBindingsOptions{
						TagBindings: []*tfe.TagBinding{{
							Key:   "additional",
							Value: "tag",
						}},
					})
					if err != nil {
						t.Fatalf("failed adding tag binding via API call: %v", err)
					}
				},
				PlanOnly: true,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEProjectExists(
						"tfe_project.foobar", project),
					testAccCheckTFEProjectAttributes(project),
					resource.TestCheckResourceAttr(
						"tfe_project.foobar", "name", "projecttest"),
					resource.TestCheckResourceAttr(
						"tfe_project.foobar", "description", "project description"),
					resource.TestCheckResourceAttr(
						"tfe_project.foobar", "tags.%", "2"),
				),
			},
		},
	})
}

func TestAccTFEProject_tagBindings(t *testing.T) {
	project := &tfe.Project{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEProject_basicTagBindings(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEProjectExists(
						"tfe_project.foobar", project),
					testAccCheckTFEProjectAttributes(project),
					resource.TestCheckResourceAttr(
						"tfe_project.foobar", "name", "projecttest"),
					resource.TestCheckResourceAttr(
						"tfe_project.foobar", "description", "project description"),
					resource.TestCheckResourceAttr(
						"tfe_project.foobar", "organization", fmt.Sprintf("tst-terraform-%d", rInt)),
					resource.TestCheckResourceAttr(
						"tfe_project.foobar", "tags.%", "2"),
					resource.TestCheckResourceAttr(
						"tfe_project.foobar", "tags.keyA", "valueA"),
					resource.TestCheckResourceAttr(
						"tfe_project.foobar", "tags.keyB", "valueB"),
				),
			},
			{
				Config: testAccTFEProject_basicTagBindingsAddOne(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEProjectExists(
						"tfe_project.foobar", project),
					testAccCheckTFEProjectAttributes(project),
					resource.TestCheckResourceAttr(
						"tfe_project.foobar", "name", "projecttest"),
					resource.TestCheckResourceAttr(
						"tfe_project.foobar", "description", "project description"),
					resource.TestCheckResourceAttr(
						"tfe_project.foobar", "organization", fmt.Sprintf("tst-terraform-%d", rInt)),
					resource.TestCheckResourceAttr(
						"tfe_project.foobar", "tags.%", "3"),
					resource.TestCheckResourceAttr(
						"tfe_project.foobar", "tags.keyA", "valueA"),
					resource.TestCheckResourceAttr(
						"tfe_project.foobar", "tags.keyB", "valueB"),
					resource.TestCheckResourceAttr(
						"tfe_project.foobar", "tags.keyC", "valueC"),
				),
			},
			{
				Config: testAccTFEProject_basicTagBindingsRemoveAll(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEProjectExists(
						"tfe_project.foobar", project),
					testAccCheckTFEProjectAttributes(project),
					resource.TestCheckResourceAttr(
						"tfe_project.foobar", "name", "projecttest"),
					resource.TestCheckResourceAttr(
						"tfe_project.foobar", "description", "project description"),
					resource.TestCheckResourceAttr(
						"tfe_project.foobar", "organization", fmt.Sprintf("tst-terraform-%d", rInt)),
					resource.TestCheckResourceAttr(
						"tfe_project.foobar", "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccTFEProject_import(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	project := &tfe.Project{}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEProjectDestroy,
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

func TestAccTFEProject_withAutoDestroy(t *testing.T) {
	project := &tfe.Project{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEProject_basicWithAutoDestroy(rInt, "3d"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEProjectExists(
						"tfe_project.foobar", project),
					testAccCheckTFEProjectAttributes(project),
					resource.TestCheckResourceAttr(
						"tfe_project.foobar", "auto_destroy_activity_duration", "3d"),
				),
			},
			{
				Config:      testAccTFEProject_basicWithAutoDestroy(rInt, "10m"),
				ExpectError: regexp.MustCompile("must be 1-4 digits followed by"),
			},
			{
				Config: testAccTFEProject_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEProjectExists(
						"tfe_project.foobar", project),
					testAccCheckTFEProjectAttributes(project),
					resource.TestCheckNoResourceAttr("tfe_project.foobar", "auto_destroy_activity_duration"),
				),
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
  description = "project description updated"
}`, rInt)
}

func testAccTFEProject_updateRemoveBindings(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_project" "foobar" {
  organization = tfe_organization.foobar.name
  name = "project updated"
  description = "project description updated"
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
  description = "project description"
}`, rInt)
}

func testAccTFEProject_basicTagBindings(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_project" "foobar" {
  organization = tfe_organization.foobar.name
  name = "projecttest"
  description = "project description"
  tags = {
	  keyA = "valueA"
	  keyB = "valueB"
  }
}`, rInt)
}

func testAccTFEProject_basicTagBindingsAddOne(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_project" "foobar" {
  organization = tfe_organization.foobar.name
  name = "projecttest"
  description = "project description"
  tags = {
	  keyA = "valueA"
	  keyB = "valueB"
	  keyC = "valueC"
  }
}`, rInt)
}

func testAccTFEProject_basicTagBindingsRemoveAll(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_project" "foobar" {
  organization = tfe_organization.foobar.name
  name = "projecttest"
  description = "project description"
  tags = {}
}`, rInt)
}

func testAccTFEProject_ignoreAdditionalTags(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_project" "foobar" {
  organization = tfe_organization.foobar.name
  name = "projecttest"
  description = "project description"
  tags = {
	  keyA = "valueA"
	  keyB = "valueB"
  }
  ignore_additional_tags = true
}`, rInt)
}

func testAccTFEProject_invalidNameChar(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_project" "foobar" {
  organization = tfe_organization.foobar.name
  name = "invalidchar#"
}`, rInt)
}
func testAccTFEProject_invalidNameLen(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_project" "foobar" {
  organization = tfe_organization.foobar.name
  name = "aa"
}`, rInt)
}

func testAccTFEProject_basicWithAutoDestroy(rInt int, duration string) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_project" "foobar" {
  organization = tfe_organization.foobar.name
  name = "projecttest"
  auto_destroy_activity_duration = "%s"
}`, rInt, duration)
}

func testAccCheckTFEProjectDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_project" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := testAccConfiguredClient.Client.Projects.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Project %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckTFEProjectExists(n string, project *tfe.Project) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		p, err := testAccConfiguredClient.Client.Projects.Read(ctx, rs.Primary.ID)
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
