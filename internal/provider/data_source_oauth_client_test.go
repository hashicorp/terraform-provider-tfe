// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"math/rand"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func testAccTFEOAuthClientDataSourcePreCheck(t *testing.T) {
	testAccPreCheck(t)
	if envGithubToken == "" {
		t.Skip("Please set GITHUB_TOKEN to run this test")
	}
}

func TestAccTFEOAuthClientDataSource_findByID(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccTFEOAuthClientDataSourcePreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOAuthClientDataSourceConfig_findByID(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"tfe_oauth_client.test", "api_url",
						"data.tfe_oauth_client.client", "api_url"),
					resource.TestCheckResourceAttrPair(
						"tfe_oauth_client.test", "http_url",
						"data.tfe_oauth_client.client", "http_url"),
					resource.TestCheckResourceAttrPair(
						"tfe_oauth_client.test", "service_provider",
						"data.tfe_oauth_client.client", "service_provider"),
					resource.TestCheckResourceAttrPair(
						"tfe_oauth_client.test", "oauth_token_id",
						"data.tfe_oauth_client.client", "oauth_token_id"),
				),
			},
		},
	})
}

func TestAccTFEOAuthClientDataSource_findByName(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccTFEOAuthClientDataSourcePreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOAuthClientDataSourceConfig_findByName(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"tfe_oauth_client.test", "api_url",
						"data.tfe_oauth_client.client", "api_url"),
					resource.TestCheckResourceAttrPair(
						"tfe_oauth_client.test", "http_url",
						"data.tfe_oauth_client.client", "http_url"),
					resource.TestCheckResourceAttrPair(
						"tfe_oauth_client.test", "service_provider",
						"data.tfe_oauth_client.client", "service_provider"),
					resource.TestCheckResourceAttrPair(
						"tfe_oauth_client.test", "oauth_token_id",
						"data.tfe_oauth_client.client", "oauth_token_id"),
				),
			},
		},
	})
}

func TestAccTFEOAuthClientDataSource_findByServiceProvider(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccTFEOAuthClientDataSourcePreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOAuthClientDataSourceConfig_findByServiceProvider(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"tfe_oauth_client.test", "api_url",
						"data.tfe_oauth_client.client", "api_url"),
					resource.TestCheckResourceAttrPair(
						"tfe_oauth_client.test", "http_url",
						"data.tfe_oauth_client.client", "http_url"),
					resource.TestCheckResourceAttrPair(
						"tfe_oauth_client.test", "service_provider",
						"data.tfe_oauth_client.client", "service_provider"),
					resource.TestCheckResourceAttrPair(
						"tfe_oauth_client.test", "oauth_token_id",
						"data.tfe_oauth_client.client", "oauth_token_id"),
				),
			},
		},
	})
}

func TestAccTFEOAuthClientDataSource_missingParameters(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccTFEOAuthClientDataSourcePreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFEOAuthClientDataSourceConfig_missingParameters(rInt),
				ExpectError: regexp.MustCompile("one of `name,oauth_client_id,service_provider` must"),
			},
		},
	})
}

func TestAccTFEOAuthClientDataSource_missingOrgWithName(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccTFEOAuthClientDataSourcePreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFEOAuthClientDataSourceConfig_missingOrgWithName(rInt),
				ExpectError: regexp.MustCompile("all of `name,organization` must"),
			},
		},
	})
}

func TestAccTFEOAuthClientDataSource_missingOrgWithServiceProvider(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccTFEOAuthClientDataSourcePreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFEOAuthClientDataSourceConfig_missingOrgWithServiceProvider(rInt),
				ExpectError: regexp.MustCompile("all of `organization,service_provider` must be"),
			},
		},
	})
}

func TestAccTFEOAuthClientDataSource_sameName(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccTFEOAuthClientDataSourcePreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFEOAuthClientDataSourceConfig_sameName(rInt),
				ExpectError: regexp.MustCompile("too many OAuthClients were found to match the given parameters"),
			},
		},
	})
}

func TestAccTFEOAuthClientDataSource_noName(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccTFEOAuthClientDataSourcePreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFEOAuthClientDataSourceConfig_noName(rInt),
				ExpectError: regexp.MustCompile("no OAuthClients found matching the given parameters"),
			},
		},
	})
}

func TestAccTFEOAuthClientDataSource_sameServiceProvider(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccTFEOAuthClientDataSourcePreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFEOAuthClientDataSourceConfig_sameServiceProvider(rInt),
				ExpectError: regexp.MustCompile("too many OAuthClients were found to match the given parameters"),
			},
		},
	})
}

func testAccTFEOAuthClientDataSourceConfig_findByID(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
	name  = "tst-terraform-%d"
	email = "admin@company.com"
}
resource "tfe_oauth_client" "test" {
	organization     = tfe_organization.foobar.name
	api_url          = "https://api.github.com"
	http_url         = "https://github.com"
	oauth_token      = "%s"
	service_provider = "github"
}
data "tfe_oauth_client" "client" {
	oauth_client_id = tfe_oauth_client.test.id
}
`, rInt, envGithubToken)
}

func testAccTFEOAuthClientDataSourceConfig_findByName(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
	name  = "tst-terraform-%d"
	email = "admin@company.com"
}
resource "tfe_oauth_client" "test" {
	organization     = tfe_organization.foobar.name
	api_url          = "https://api.github.com"
	http_url         = "https://github.com"
	name             = "tst-github-%d"
	oauth_token      = "%s"
	service_provider = "github"
}
data "tfe_oauth_client" "client" {
    organization = "tst-terraform-%d"
	name         = "tst-github-%d"
	depends_on = [tfe_oauth_client.test]
}
`, rInt, rInt, envGithubToken, rInt, rInt)
}

func testAccTFEOAuthClientDataSourceConfig_findByServiceProvider(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
	name  = "tst-terraform-%d"
	email = "admin@company.com"
}
resource "tfe_oauth_client" "test" {
	organization     = tfe_organization.foobar.name
	api_url          = "https://api.github.com"
	http_url         = "https://github.com"
	oauth_token      = "%s"
	service_provider = "github"
}
data "tfe_oauth_client" "client" {
    organization = "tst-terraform-%d"
	service_provider = "github"
	depends_on = [tfe_oauth_client.test]
}
`, rInt, envGithubToken, rInt)
}

func testAccTFEOAuthClientDataSourceConfig_missingParameters(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
	name  = "tst-terraform-%d"
	email = "admin@company.com"
}
resource "tfe_oauth_client" "test" {
	organization     = tfe_organization.foobar.name
	api_url          = "https://api.github.com"
	http_url         = "https://github.com"
	oauth_token      = "%s"
	service_provider = "github"
}
data "tfe_oauth_client" "client" {
    organization = "tst-terraform-%d"
	depends_on = [tfe_oauth_client.test]
}
`, rInt, envGithubToken, rInt)
}

func testAccTFEOAuthClientDataSourceConfig_missingOrgWithName(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
	name  = "tst-terraform-%d"
	email = "admin@company.com"
}
resource "tfe_oauth_client" "test" {
	organization     = tfe_organization.foobar.name
	api_url          = "https://api.github.com"
	http_url         = "https://github.com"
	oauth_token      = "%s"
	service_provider = "github"
}
data "tfe_oauth_client" "client" {
	name = "github"
	depends_on = [tfe_oauth_client.test]
}
`, rInt, envGithubToken)
}

func testAccTFEOAuthClientDataSourceConfig_missingOrgWithServiceProvider(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
	name  = "tst-terraform-%d"
	email = "admin@company.com"
}
resource "tfe_oauth_client" "test" {
	organization     = tfe_organization.foobar.name
	api_url          = "https://api.github.com"
	http_url         = "https://github.com"
	oauth_token      = "%s"
	service_provider = "github"
}
data "tfe_oauth_client" "client" {
	service_provider = "github"
	depends_on = [tfe_oauth_client.test]
}
`, rInt, envGithubToken)
}

func testAccTFEOAuthClientDataSourceConfig_sameName(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
	name  = "tst-terraform-%d"
	email = "admin@company.com"
}
resource "tfe_oauth_client" "test1" {
	organization     = tfe_organization.foobar.name
	api_url          = "https://api.github.com"
	http_url         = "https://github.com"
	name             = "tst-github"
	oauth_token      = "%s"
	service_provider = "github"
}
resource "tfe_oauth_client" "test2" {
	organization     = tfe_organization.foobar.name
	api_url          = "https://api.github.com"
	http_url         = "https://github.com"
	name             = "tst-github"
	oauth_token      = "%s"
	service_provider = "github"
}
data "tfe_oauth_client" "client" {
	organization = tfe_organization.foobar.name
	name         = tfe_oauth_client.test1.name
	depends_on = [tfe_oauth_client.test1, tfe_oauth_client.test2]
}
`, rInt, envGithubToken, envGithubToken)
}

func testAccTFEOAuthClientDataSourceConfig_noName(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
	name  = "tst-terraform-%d"
	email = "admin@company.com"
}
resource "tfe_oauth_client" "test" {
	organization     = tfe_organization.foobar.name
	api_url          = "https://api.github.com"
	http_url         = "https://github.com"
	oauth_token      = "%s"
	service_provider = "github"
}
data "tfe_oauth_client" "client" {
	organization = tfe_organization.foobar.name
	name         = "tst-github"
	depends_on = [tfe_oauth_client.test]
}
`, rInt, envGithubToken)
}

func testAccTFEOAuthClientDataSourceConfig_sameServiceProvider(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
	name  = "tst-terraform-%d"
	email = "admin@company.com"
}
resource "tfe_oauth_client" "test1" {
	organization     = tfe_organization.foobar.name
	api_url          = "https://api.github.com"
	http_url         = "https://github.com"
	oauth_token      = "%s"
	service_provider = "github"
}
resource "tfe_oauth_client" "test2" {
	organization     = tfe_organization.foobar.name
	api_url          = "https://api.github.com"
	http_url         = "https://github.com"
	oauth_token      = "%s"
	service_provider = "github"
}
data "tfe_oauth_client" "client" {
    organization     = tfe_organization.foobar.name
	service_provider = "github"
	depends_on = [tfe_oauth_client.test1, tfe_oauth_client.test2]
}
`, rInt, envGithubToken, envGithubToken)
}
