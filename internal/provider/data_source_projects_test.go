// Copyright IBM Corp. 2018, 2025
// SPDX-License-IDentifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"testing"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestListOrganizationProjectsPagination(t *testing.T) {
	const orgName = "hashicorp"
	project := func(id, name string) string {
		return fmt.Sprintf(`{
			"id": %q,
			"type": "projects",
			"attributes": {"name": %q, "description": ""},
			"relationships": {"organization": {"data": {"id": %q, "type": "organizations"}}}
		}`, id, name, orgName)
	}
	pages := map[string]string{
		"1": fmt.Sprintf(`{"data":[%s],"meta":{"pagination":{"current-page":1,"next-page":2,"page-size":100,"total-count":2,"total-pages":2}}}`, project("prj-1", "first")),
		"2": fmt.Sprintf(`{"data":[%s],"meta":{"pagination":{"current-page":2,"next-page":null,"page-size":100,"total-count":2,"total-pages":2}}}`, project("prj-2", "second")),
	}

	var requestedPages []string
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v2/organizations/"+orgName+"/projects", func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page[number]")
		requestedPages = append(requestedPages, page)
		if pageSize := r.URL.Query().Get("page[size]"); pageSize != "100" {
			t.Errorf("page[size] = %q, want 100", pageSize)
		}
		body, ok := pages[page]
		if !ok {
			http.Error(w, `{"errors":[{"status":"404","title":"not found"}]}`, http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/vnd.api+json")
		fmt.Fprint(w, body)
	})

	projects, err := listOrganizationProjects(context.Background(), testTfeClientV2(t, mux), orgName)
	if err != nil {
		t.Fatalf("list projects: %v", err)
	}
	if want := []string{"1", "2"}; !reflect.DeepEqual(requestedPages, want) {
		t.Fatalf("requested pages = %v, want %v", requestedPages, want)
	}
	if len(projects) != 2 {
		t.Fatalf("project count = %d, want 2", len(projects))
	}
	if got := []string{valueOrZero(projects[0].GetId()), valueOrZero(projects[1].GetId())}; !reflect.DeepEqual(got, []string{"prj-1", "prj-2"}) {
		t.Fatalf("project IDs = %v, want [prj-1 prj-2]", got)
	}
}

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
		ProtoV6ProviderFactories: testAccMuxedProviders,
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
		ProtoV6ProviderFactories: testAccMuxedProviders,
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
