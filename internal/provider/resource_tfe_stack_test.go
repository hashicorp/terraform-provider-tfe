// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccTFEStackResource_basic(t *testing.T) {
	skipUnlessBeta(t)

	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEStackResourceConfig(orgName, envGithubToken, "brandonc/pet-nulls-stack"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tfe_stack.foobar", "id"),
					resource.TestCheckResourceAttrSet("tfe_stack.foobar", "project_id"),
					resource.TestCheckResourceAttr("tfe_stack.foobar", "name", "example-stack"),
					resource.TestCheckResourceAttr("tfe_stack.foobar", "description", "Just an ordinary stack"),
					resource.TestCheckResourceAttr("tfe_stack.foobar", "vcs_repo.identifier", "brandonc/pet-nulls-stack"),
					resource.TestCheckResourceAttrSet("tfe_stack.foobar", "vcs_repo.oauth_token_id"),
					resource.TestCheckResourceAttrSet("tfe_stack.foobar", "created_at"),
					resource.TestCheckResourceAttrSet("tfe_stack.foobar", "updated_at"),
				),
			},
			{
				ResourceName:      "tfe_stack.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccTFEStackResourceConfig(orgName, ghToken, ghRepoIdentifier string) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "%s"
  email = "admin@tfe.local"
}

resource "tfe_project" "example" {
	name         = "example"
	organization = tfe_organization.foobar.name
}

resource "tfe_oauth_client" "foobar" {
  organization     = tfe_organization.foobar.name
  api_url          = "https://api.github.com"
  http_url         = "https://github.com"
  oauth_token      = "%s"
  service_provider = "github"
}

resource "tfe_stack" "foobar" {
	name        = "example-stack"
	description = "Just an ordinary stack"
  project_id  = tfe_project.example.id

	vcs_repo {
    identifier         = "%s"
    oauth_token_id     = tfe_oauth_client.foobar.oauth_token_id
  }
}
`, orgName, ghToken, ghRepoIdentifier)
}
