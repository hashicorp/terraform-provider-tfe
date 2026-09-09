// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"testing"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestTFEOAuthClientADOOrgNameValidation(t *testing.T) {
	validate := resourceTFEOAuthClient().Schema["ado_org_name"].ValidateFunc
	tests := map[string]bool{
		"":            true,
		"a":           true,
		"my-company":  true,
		"1-company-2": true,
		"-my-company": false,
		"my-company-": false,
		"my_company":  false,
		"my company":  false,
		"abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwx":  true,
		"abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxy": false,
	}

	for value, valid := range tests {
		t.Run(value, func(t *testing.T) {
			_, errors := validate(value, "ado_org_name")
			if valid && len(errors) != 0 {
				t.Fatalf("expected %q to be valid, got %v", value, errors)
			}
			if !valid && len(errors) == 0 {
				t.Fatalf("expected %q to be invalid", value)
			}
		})
	}
}

func TestNewOAuthClientEnvelopeWithADOOrgName(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceTFEOAuthClient().Schema, map[string]interface{}{
		"ado_org_name":        "my-company",
		"agent_pool_id":       "apool-123",
		"api_url":             "https://app.vssps.visualstudio.com",
		"http_url":            "https://dev.azure.com",
		"key":                 "key",
		"name":                "ado",
		"oauth_token":         "token",
		"organization_scoped": true,
		"service_provider":    "ado_services",
	})

	envelope := newOAuthClientEnvelope(d, true)
	client := envelope.GetData()
	attrs := client.GetAttributes()

	if got := valueOrZero(attrs.GetAdoOrgName()); got != "my-company" {
		t.Fatalf("expected ado_org_name my-company, got %q", got)
	}
	if got := attrs.GetAdditionalData()["oauth-token-string"]; got != "token" {
		t.Fatalf("expected oauth-token-string token, got %#v", got)
	}
	if got := valueOrZero(attrs.GetServiceProvider()); got != "ado_services" {
		t.Fatalf("expected service_provider ado_services, got %q", got)
	}
	agentPool := client.GetRelationships().GetAgentPool().GetData()
	if got := valueOrZero(agentPool.GetId()); got != "apool-123" {
		t.Fatalf("expected agent_pool_id apool-123, got %q", got)
	}
}

func TestOAuthClientEnvelopeSerializesADOOrgName(t *testing.T) {
	client := testTfeClientV2(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v2/organizations/my-org/oauth-clients" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}

		var payload struct {
			Data struct {
				Attributes map[string]interface{} `json:"attributes"`
			} `json:"data"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		if got := payload.Data.Attributes["ado-org-name"]; got != "my-company" {
			t.Errorf("expected ado-org-name my-company, got %#v", got)
		}
		if got := payload.Data.Attributes["oauth-token-string"]; got != "token" {
			t.Errorf("expected oauth-token-string token, got %#v", got)
		}

		w.Header().Set("Content-Type", "application/vnd.api+json")
		fmt.Fprint(w, `{"data":{"id":"oc-123","type":"oauth-clients","attributes":{"ado-org-name":"my-company"}}}`)
	}))
	d := schema.TestResourceDataRaw(t, resourceTFEOAuthClient().Schema, map[string]interface{}{
		"ado_org_name":        "my-company",
		"api_url":             "https://app.vssps.visualstudio.com",
		"http_url":            "https://dev.azure.com",
		"oauth_token":         "token",
		"organization_scoped": true,
		"service_provider":    "ado_services",
	})

	if _, err := client.API.Organizations().ByOrganization_name("my-org").OauthClients().Post(ctx, newOAuthClientEnvelope(d, true), nil); err != nil {
		t.Fatal(err)
	}
}

func TestOAuthClientEnvelopeSerializesClearedADOOrgName(t *testing.T) {
	client := testTfeClientV2(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch || r.URL.Path != "/api/v2/oauth-clients/oc-123" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}

		var payload struct {
			Data struct {
				Attributes map[string]interface{} `json:"attributes"`
			} `json:"data"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		value, present := payload.Data.Attributes["ado-org-name"]
		if !present || value != nil {
			t.Errorf("expected ado-org-name to be null, got %#v", value)
		}

		w.Header().Set("Content-Type", "application/vnd.api+json")
		fmt.Fprint(w, `{"data":{"id":"oc-123","type":"oauth-clients","attributes":{"ado-org-name":null}}}`)
	}))
	d := schema.TestResourceDataRaw(t, resourceTFEOAuthClient().Schema, map[string]interface{}{
		"organization_scoped": true,
		"service_provider":    "ado_services",
	})
	d.SetId("oc-123")

	if _, err := client.API.OauthClients().ByOauth_client_id(d.Id()).Patch(ctx, newOAuthClientEnvelope(d, false), nil); err != nil {
		t.Fatal(err)
	}
}

func TestReadOAuthClientADOOrgName(t *testing.T) {
	client := testTfeClientV2(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/oauth-clients/oc-123" {
			t.Fatalf("unexpected request path %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/vnd.api+json")
		fmt.Fprint(w, `{"data":{"id":"oc-123","type":"oauth-clients","attributes":{"ado-org-name":"my-company"}}}`)
	}))

	got, err := readOAuthClientADOOrgName(ConfiguredClient{ClientV2: client}, "oc-123")
	if err != nil {
		t.Fatal(err)
	}
	if got != "my-company" {
		t.Fatalf("expected ado_org_name my-company, got %q", got)
	}
}

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
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEOAuthClientDestroy,
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

func TestAccTFEOAuthClientWithOrganizationScoped_basic(t *testing.T) {
	oc := &tfe.OAuthClient{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			if envGithubToken == "" {
				t.Skip("Please set GITHUB_TOKEN to run this test")
			}
		},
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEOAuthClientDestroy,
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
					resource.TestCheckResourceAttr(
						"tfe_oauth_client.foobar", "organization_scoped", "true"),
				),
			},
		},
	})
}

func TestAccTFEOAuthClient_rsaKeys(t *testing.T) {
	oc := &tfe.OAuthClient{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEOAuthClientDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOAuthClient_rsaKeys(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOAuthClientExists("tfe_oauth_client.foobar", oc),
					testAccCheckTFEOAuthClientAttributes(oc),
					resource.TestCheckResourceAttr(
						"tfe_oauth_client.foobar", "api_url", "https://bbdc.example.com"),
					resource.TestCheckResourceAttr(
						"tfe_oauth_client.foobar", "http_url", "https://bbdc.example.com"),
					resource.TestCheckResourceAttr(
						"tfe_oauth_client.foobar", "service_provider", "bitbucket_data_center"),
					resource.TestCheckResourceAttr(
						"tfe_oauth_client.foobar", "key", "1e4843e138b0d44911a50d15e0f7cee4"),
					resource.TestCheckResourceAttr(
						"tfe_oauth_client.foobar", "rsa_public_key", "-----BEGIN PUBLIC KEY-----\nVGm9w0J8t6gWe745gW6E9NHJGiDKehh58bAtjO0wPvFg5l8Ea9s+PpAvP4wCZWDS\nhwIDAQAB\n-----END PUBLIC KEY-----\n"),
				),
			},
		},
	})
}

func TestAccTFEOAuthClient_agentPool(t *testing.T) {
	skipUnlessBeta(t)
	oc := &tfe.OAuthClient{}
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			if envGithubToken == "" {
				t.Skip("Please set GITHUB_TOKEN to run this test")
			}
		},
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEOAuthClientDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOAuthClient_agentPool(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOAuthClientExists("tfe_oauth_client.foobar", oc),
					testAccCheckTFEOAuthClientAttributes(oc),
					resource.TestCheckResourceAttr(
						"tfe_oauth_client.foobar", "service_provider", "github_enterprise"),
				),
			},
		},
	})
}

func TestAccTFEOAuthClient_updateOAuthTokenID(t *testing.T) {
	oc := &tfe.OAuthClient{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	var initialOAuthTokenID string

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			if envGithubToken == "" {
				t.Skip("Please set GITHUB_TOKEN to run this test")
			}

			if envGithubToken2 == "" {
				t.Skip("Please set GITHUB_TOKEN2 to run this test")
			}
		},
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEOAuthClientDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create with the initial oauth_token_id.
			{
				Config: testAccTFEOAuthClient_updateOAuthTokenID(rInt, envGithubToken),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOAuthClientExists("tfe_oauth_client.foobar", oc),
					resource.TestCheckResourceAttrSet("tfe_oauth_client.foobar", "oauth_token_id"),
					func(s *terraform.State) error {
						initialOAuthTokenID = oc.OAuthTokens[0].ID
						return nil
					},
				),
			},
			// Step 2: Update the oauth_token_id value.
			{
				Config: testAccTFEOAuthClient_updateOAuthTokenID(rInt, envGithubToken2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOAuthClientExists("tfe_oauth_client.foobar", oc),
					resource.TestCheckResourceAttrSet("tfe_oauth_client.foobar", "oauth_token_id"),
					func(s *terraform.State) error {
						if initialOAuthTokenID != oc.OAuthTokens[0].ID {
							return fmt.Errorf("oauth_token_id changed")
						}
						return nil
					},
				),
			},
			// Step 3: Run a plan-only step to ensure no changes.
			{
				Config:   testAccTFEOAuthClient_updateOAuthTokenID(rInt, envGithubToken2),
				PlanOnly: true,
			},
		},
	})
}

func testAccCheckTFEOAuthClientExists(
	n string, oc *tfe.OAuthClient) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		client, err := testAccConfiguredClient.Client.OAuthClients.Read(ctx, rs.Primary.ID)
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
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_oauth_client" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := testAccConfiguredClient.Client.OAuthClients.Read(ctx, rs.Primary.ID)
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
  organization_scoped = true
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
  api_url          = "https://bbdc.example.com"
  http_url         = "https://bbdc.example.com"
  service_provider = "bitbucket_data_center"
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

func testAccTFEOAuthClient_agentPool() string {
	return fmt.Sprintf(`
data "tfe_organization" "foobar" {
  name  = "xxx"
}

data "tfe_agent_pool" "foobar" {
  name = "xxx"
  organization = data.tfe_organization.foobar.name
}

resource "tfe_oauth_client" "foobar" {
  organization     = data.tfe_organization.foobar.name
  api_url          = "https://githubenterprise.xxx/api/v3"
  http_url         = "https://githubenterprise.xxx"
  oauth_token      = "%s"
  service_provider = "github_enterprise"
  agent_pool_id    = data.tfe_agent_pool.foobar.id
}`, envGithubToken)
}

func testAccTFEOAuthClient_updateOAuthTokenID(rInt int, oAuthToken string) string {
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
  organization_scoped = true
}`, rInt, oAuthToken)
}
