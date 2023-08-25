// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

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
			if envGithubToken == "" {
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

func TestAccTFEOAuthClient_rsaKeys(t *testing.T) {
	oc := &tfe.OAuthClient{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEOAuthClientDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOAuthClient_rsaKeys(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOAuthClientExists("tfe_oauth_client.foobar", oc),
					testAccCheckTFEOAuthClientAttributes(oc),
					resource.TestCheckResourceAttr(
						"tfe_oauth_client.foobar", "api_url", "https://bbs.example.com"),
					resource.TestCheckResourceAttr(
						"tfe_oauth_client.foobar", "http_url", "https://bbs.example.com"),
					resource.TestCheckResourceAttr(
						"tfe_oauth_client.foobar", "service_provider", "bitbucket_server"),
					resource.TestCheckResourceAttr(
						"tfe_oauth_client.foobar", "key", "1e4843e138b0d44911a50d15e0f7cee4"),
					resource.TestCheckResourceAttr(
						"tfe_oauth_client.foobar", "rsa_public_key", "-----BEGIN PUBLIC KEY-----\nVGm9w0J8t6gWe745gW6E9NHJGiDKehh58bAtjO0wPvFg5l8Ea9s+PpAvP4wCZWDS\nhwIDAQAB\n-----END PUBLIC KEY-----\n"),
				),
			},
		},
	})
}

func testAccCheckTFEOAuthClientExists(
	n string, oc *tfe.OAuthClient) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(ConfiguredClient)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		client, err := config.Client.OAuthClients.Read(ctx, rs.Primary.ID)
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
		if oc.ServiceProvider == tfe.ServiceProviderGithub && oc.APIURL != "https://api.github.com" {
			return fmt.Errorf("Bad API URL: %s", oc.APIURL)
		}

		if oc.ServiceProvider == tfe.ServiceProviderGithub && oc.HTTPURL != "https://github.com" {
			return fmt.Errorf("Bad HTTP URL: %s", oc.HTTPURL)
		}

		return nil
	}
}

func testAccCheckTFEOAuthClientDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(ConfiguredClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_oauth_client" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := config.Client.OAuthClients.Read(ctx, rs.Primary.ID)
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
}`, rInt, envGithubToken)
}

func testAccTFEOAuthClient_rsaKeys(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_oauth_client" "foobar" {
  organization     = tfe_organization.foobar.id
	name 						 = "foobar_oauth"
  api_url          = "https://bbs.example.com"
  http_url         = "https://bbs.example.com"
  service_provider = "bitbucket_server"
  key       			 = "1e4843e138b0d44911a50d15e0f7cee4"
  secret           = <<EOT
-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAoKizy4xbN6qZFAwIJV24liz/vYBSvR3SjEiUzhpp0uMAmICN
-----END RSA PRIVATE KEY-----
EOT
  rsa_public_key   = <<EOT
-----BEGIN PUBLIC KEY-----
VGm9w0J8t6gWe745gW6E9NHJGiDKehh58bAtjO0wPvFg5l8Ea9s+PpAvP4wCZWDS
hwIDAQAB
-----END PUBLIC KEY-----
EOT
}`, rInt)
}
