// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"github.com/hashicorp/go-tfe"
	"math/rand"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccTFEProjectsDataSource_basic(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}
	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)
	orgName := org.Name

	prj1 := createProject(t, tfeClient, org.Name, tfe.ProjectCreateOptions{
		Name: randomString(t),
	})
	prj2 := createProject(t, tfeClient, org.Name, tfe.ProjectCreateOptions{
		Name: randomString(t),
	})
	prj3 := createProject(t, tfeClient, org.Name, tfe.ProjectCreateOptions{
		Name: randomString(t),
	})
	prjNames := []string{"Default Project", prj1.Name, prj2.Name, prj3.Name}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEProjectsDataSourceConfig(orgName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.tfe_projects.all", "organization", orgName),
					resource.TestCheckResourceAttr(
						"data.tfe_projects.all", "projects.#", "4"),
					resource.TestCheckResourceAttrSet(
						"data.tfe_projects.all", "projects.0.id"),
					resource.TestCheckResourceAttrWith(
						"data.tfe_projects.all", "projects.0.name", func(value string) error {
							for _, name := range prjNames {
								if name == value {
									return nil
								}
							}
							return fmt.Errorf("Excepted project name %s to be in the list %v but not found. ", value, prjNames)
						}),
					resource.TestCheckResourceAttr(
						"data.tfe_projects.all", "projects.0.description", ""),
					resource.TestCheckResourceAttr(
						"data.tfe_projects.all", "projects.0.organization", orgName),
					resource.TestCheckResourceAttrSet(
						"data.tfe_projects.all", "projects.1.id"),
					resource.TestCheckResourceAttrWith(
						"data.tfe_projects.all", "projects.1.name", func(value string) error {
							for _, name := range prjNames {
								if name == value {
									return nil
								}
							}
							return fmt.Errorf("Excepted project name %s to be in the list %v but not found. ", value, prjNames)
						}),
					resource.TestCheckResourceAttr(
						"data.tfe_projects.all", "projects.1.organization", orgName),
					resource.TestCheckResourceAttrSet(
						"data.tfe_projects.all", "projects.2.id"),
					resource.TestCheckResourceAttrWith(
						"data.tfe_projects.all", "projects.2.name", func(value string) error {
							for _, name := range prjNames {
								if name == value {
									return nil
								}
							}
							return fmt.Errorf("Excepted project name %s to be in the list %v but not found. ", value, prjNames)
						}),
					resource.TestCheckResourceAttr(
						"data.tfe_projects.all", "projects.2.organization", orgName),
					resource.TestCheckResourceAttrSet(
						"data.tfe_projects.all", "projects.3.id"),
					resource.TestCheckResourceAttrWith(
						"data.tfe_projects.all", "projects.3.name", func(value string) error {
							for _, name := range prjNames {
								if name == value {
									return nil
								}
							}
							return fmt.Errorf("Excepted project name %s to be in the list %v but not found. ", value, prjNames)
						}),
					resource.TestCheckResourceAttr(
						"data.tfe_projects.all", "projects.3.organization", orgName),
				),
			},
		},
	})
}

func TestAccTFEProjectsDataSource_basicNoProjects(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEProjectsDataSourceConfig_noProjects(orgName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.tfe_projects.all", "organization", orgName),
					resource.TestCheckResourceAttr(
						"data.tfe_projects.all", "projects.#", "1"),
					resource.TestCheckResourceAttr(
						"data.tfe_projects.all", "projects.0.name", "Default Project"),
				),
			},
		},
	})
}

func testAccTFEProjectsDataSourceConfig(orgName string) string {
	return fmt.Sprintf(`
	data tfe_projects "all" {
	  organization = "%s"
	}
	`, orgName)
}

func testAccTFEProjectsDataSourceConfig_noProjects(orgName string) string {
	return fmt.Sprintf(`
resource "tfe_organization" "organization" {
  name  = "%s"
  email = "admin@tfe.local"
}
data tfe_projects "all" {
  organization = tfe_organization.organization.name
}
`, orgName)
}
