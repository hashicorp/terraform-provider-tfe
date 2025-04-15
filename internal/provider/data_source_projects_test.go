// Copyright (c) HashiCorp, Inc.
// SPDX-License-IDentifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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
		Name: "project1",
	})
	prj2 := createProject(t, tfeClient, org.Name, tfe.ProjectCreateOptions{
		Name: "project2",
	})
	prj3 := createProject(t, tfeClient, org.Name, tfe.ProjectCreateOptions{
		Name: "project3",
	})

	prjNames := []string{"Default Project", prj1.Name, prj2.Name, prj3.Name}
	prjNameExists := func(value string) error {
		for _, name := range prjNames {
			if name == value {
				return nil
			}
		}
		return fmt.Errorf("Expected project name %s to be in the list %v but not found. ", value, prjNames)
	}

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
					resource.TestCheckResourceAttr(
						"data.tfe_projects.all", "projects.0.name", "Default Project"),
					resource.TestCheckResourceAttr(
						"data.tfe_projects.all", "projects.0.description", ""),
					resource.TestCheckResourceAttr(
						"data.tfe_projects.all", "projects.0.organization", orgName),
					resource.TestCheckResourceAttrSet(
						"data.tfe_projects.all", "projects.1.id"),
					resource.TestCheckResourceAttrWith(
						"data.tfe_projects.all", "projects.1.name", prjNameExists),
					resource.TestCheckResourceAttr(
						"data.tfe_projects.all", "projects.1.organization", orgName),
					resource.TestCheckResourceAttrSet(
						"data.tfe_projects.all", "projects.2.id"),
					resource.TestCheckResourceAttrWith(
						"data.tfe_projects.all", "projects.2.name", prjNameExists),
					resource.TestCheckResourceAttr(
						"data.tfe_projects.all", "projects.2.organization", orgName),
					resource.TestCheckResourceAttrSet(
						"data.tfe_projects.all", "projects.3.id"),
					resource.TestCheckResourceAttrWith(
						"data.tfe_projects.all", "projects.3.name", prjNameExists),
					resource.TestCheckResourceAttr(
						"data.tfe_projects.all", "projects.3.organization", orgName),
				),
			},
		},
	})
}

func TestAccTFEProjectsDataSource_basicNoProjects(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}
	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)
	orgName := org.Name

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
