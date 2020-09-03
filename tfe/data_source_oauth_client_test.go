package tfe

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccTFEOAuthClientDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOAuthClientDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"tfe_oauth_client.test", "api_url",
						"data.tfe_oauth_client.client", "api_url"),
					resource.TestCheckResourceAttrPair(
						"tfe_oauth_client.test", "http_url",
						"data.tfe_oauth_client.client", "http_url"),
					resource.TestCheckResourceAttrPair(
						"tfe_oauth_client.test", "oauth_token_id",
						"data.tfe_oauth_client.client", "token_id"),
				),
			},
		},
	})
}

func testAccTFEOAuthClientDataSourceConfig() string {
	return fmt.Sprintf(`
	resource "tfe_organization" "foobar" {
		name  = "tst-terraform"
		email = "admin@company.com"
	  }
	  
	  resource "tfe_oauth_client" "test" {
		organization     = "${tfe_organization.foobar.id}"
		api_url          = "https://api.github.com"
		http_url         = "https://github.com"
		oauth_token      = "%s"
		service_provider = "github"
	  }	

	  data "tfe_oauth_client" "client" {
		  oauth_client_id = "${tfe_oauth_client.test.id}"
	  }
	`, GITHUB_TOKEN)
}
