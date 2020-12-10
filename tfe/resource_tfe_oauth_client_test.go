package tfe

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccTFEOAuthClient_basic(t *testing.T) {
	oc := &tfe.OAuthClient{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			if GITHUB_TOKEN == "" {
				t.Skip("Please set GITHUB_TOKEN to run this test")
			}
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEOAuthClientDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOAuthClient_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOAuthClientExists("tfe_oauth_client.foobar", oc),
					testAccCheckTFEOAuthClientAttributes(oc),
					resource.TestCheckResourceAttr(
						"tfe_oauth_client.foobar", "api_url", "https://api.github.com"),
					resource.TestCheckResourceAttr(
						"tfe_oauth_client.foobar", "http_url", "https://github.com"),
					resource.TestCheckResourceAttr(
						"tfe_oauth_client.foobar", "service_provider", "github"),
				),
			},
		},
	})
}

func testAccCheckTFEOAuthClientExists(
	n string, oc *tfe.OAuthClient) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		tfeClient := testAccProvider.Meta().(*tfe.Client)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		client, err := tfeClient.OAuthClients.Read(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}

		if client.ID != rs.Primary.ID {
			return fmt.Errorf("OAuth client not found")
		}

		*oc = *client

		return nil
	}
}

func testAccCheckTFEOAuthClientAttributes(
	oc *tfe.OAuthClient) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if oc.APIURL != "https://api.github.com" {
			return fmt.Errorf("Bad API URL: %s", oc.APIURL)
		}

		if oc.HTTPURL != "https://github.com" {
			return fmt.Errorf("Bad HTTP URL: %s", oc.HTTPURL)
		}

		if oc.ServiceProvider != tfe.ServiceProviderGithub {
			return fmt.Errorf("Bad service provider: %s", oc.ServiceProvider)
		}

		return nil
	}
}

func testAccCheckTFEOAuthClientDestroy(s *terraform.State) error {
	tfeClient := testAccProvider.Meta().(*tfe.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_oauth_client" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := tfeClient.OAuthClients.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("OAuth client %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccTFEOAuthClient_basic(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_oauth_client" "foobar" {
  organization     = tfe_organization.foobar.id
  api_url          = "https://api.github.com"
  http_url         = "https://github.com"
  oauth_token      = "%s"
  service_provider = "github"
}`, rInt, GITHUB_TOKEN)
}
