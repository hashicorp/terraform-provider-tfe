package tfe

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccTFEOAuthClientsDataSource_basic(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			if GITHUB_TOKEN == "" {
				t.Skip("Please set GITHUB_TOKEN to run this test")
			}
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{Destroy: false,
				PreventPostDestroyRefresh: true,
				Config:                    testAccTFEOAuthClientsDataSourceConfig(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.tfe_oauth_clients.foobar", "oauth_clients.#", "2"),
					resource.TestCheckResourceAttr(
						"data.tfe_oauth_clients.foobar", "oauth_clients.0.service_provider", "github"),
					resource.TestCheckResourceAttr(
						"data.tfe_oauth_clients.foobar", "oauth_clients.1.service_provider", "github"),
				),
			},
		},
	})
}

func testAccTFEOAuthClientsDataSourceConfig(rInt int) string {
	return fmt.Sprintf(`
	resource "tfe_organization" "foobar" {
		name  = "tst-terraform-%d"
		email = "admin@company.com"
	
	}
	
	resource "tfe_oauth_client" "foo" {
		organization     = tfe_organization.foobar.id
		api_url          = "https://api.github.com"
		http_url         = "https://github.com"
		oauth_token      = "%s"
		service_provider = "github"
	}
	
	resource "tfe_oauth_client" "bar" {
		organization     = tfe_organization.foobar.id
		api_url          = "https://api.github.com"
		http_url         = "https://github.com"
		oauth_token      = "%s"
		service_provider = "github"
	}
	
	data "tfe_oauth_clients" "foobar" {
		organization = tfe_organization.foobar.id
		depends_on   = [tfe_organization.foobar, tfe_oauth_client.foo, tfe_oauth_client.bar]
	}
	`, rInt, GITHUB_TOKEN, GITHUB_TOKEN)
}
