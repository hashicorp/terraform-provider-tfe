// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"math/rand"
	"regexp"
	"testing"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccTFEProjectOAuthClient_basic(t *testing.T) {
	skipUnlessBeta(t)
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createOrganization(t, tfeClient, tfe.OrganizationCreateOptions{
		Name:  tfe.String(fmt.Sprintf("tst-terraform-%d", rInt)),
		Email: tfe.String(fmt.Sprintf("%s@tfe.local", randomString(t))),
	})
	t.Cleanup(orgCleanup)

	// Make a project
	project := createProject(t, tfeClient, org.Name, tfe.ProjectCreateOptions{
		Name: randomString(t),
	})

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEProjectOAuthClientDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEProjectOAuthClient_basic(org.Name, project.ID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEProjectOAuthClientExists(
						"tfe_project_oauth_client.test"),
				),
			},
			{
				ResourceName:      "tfe_project_oauth_client.test",
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s/%s/oauth_client_test", org.Name, project.ID),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTFEProjectOAuthClient_incorrectImportSyntax(t *testing.T) {
	skipUnlessBeta(t)
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createOrganization(t, tfeClient, tfe.OrganizationCreateOptions{
		Name:  tfe.String(fmt.Sprintf("tst-terraform-%d", rInt)),
		Email: tfe.String(fmt.Sprintf("%s@tfe.local", randomString(t))),
	})
	t.Cleanup(orgCleanup)

	// Make a project
	project := createProject(t, tfeClient, org.Name, tfe.ProjectCreateOptions{
		Name: randomString(t),
	})

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEProjectOAuthClient_basic(org.Name, project.ID),
			},
			{
				ResourceName:  "tfe_project_oauth_client.test",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s/tst-terraform-%d", org.Name, rInt),
				ExpectError:   regexp.MustCompile(`Error: invalid project oauth client input format`),
			},
		},
	})
}

func testAccCheckTFEProjectOAuthClientExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(ConfiguredClient)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		id := rs.Primary.ID
		if id == "" {
			return fmt.Errorf("no ID is set")
		}

		oauthClientID := rs.Primary.Attributes["oauth_client_id"]
		if oauthClientID == "" {
			return fmt.Errorf("no oauth client id set")
		}

		projectID := rs.Primary.Attributes["project_id"]
		if projectID == "" {
			return fmt.Errorf("no project id set")
		}

		oauthClient, err := config.Client.OAuthClients.ReadWithOptions(ctx, oauthClientID, &tfe.OAuthClientReadOptions{
			Include: []tfe.OAuthClientIncludeOpt{tfe.OauthClientProjects},
		})
		if err != nil {
			return fmt.Errorf("error reading oauth client %s: %w", oauthClientID, err)
		}
		for _, project := range oauthClient.Projects {
			if project.ID == projectID {
				return nil
			}
		}

		return fmt.Errorf("project (%s) is not attached to oauth client (%s).", projectID, oauthClientID)
	}
}

func testAccCheckTFEProjectOAuthClientDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(ConfiguredClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_oauth_client" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no instance ID is set")
		}

		_, err := config.Client.OAuthClients.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("oauth client %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccTFEProjectOAuthClient_base(orgName string) string {
	return fmt.Sprintf(`
		resource "tfe_oauth_client" "test" {			
			name = "oauth_client_test"			
			organization     = "%s"
			api_url          = "https://api.github.com"
			http_url         = "https://github.com"
			oauth_token      = "%s"
			service_provider = "github"
		}
	`, orgName, envGithubToken)
}

func testAccTFEProjectOAuthClient_basic(orgName string, prjID string) string {
	return testAccTFEProjectOAuthClient_base(orgName) + fmt.Sprintf(`
		resource "tfe_project_oauth_client" "test" {
		  oauth_client_id = tfe_oauth_client.test.id
		  project_id    = "%s"
		}
	`, prjID)
}
