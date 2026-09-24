// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
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

func TestNewADOServiceOAuthClientEnvelopeWithADOOrgName(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceTFEOAuthClient().Schema, map[string]interface{}{
		"ado_org_name":        "my-company",
		"agent_pool_id":       "apool-123",
		"api_url":             "https://dev.azure.com",
		"http_url":            "https://dev.azure.com",
		"key":                 "key",
		"name":                "ado",
		"oauth_token":         "token",
		"organization_scoped": true,
		"service_provider":    "ado_services",
	})

	envelope := newADOServiceOAuthClientEnvelope(d, true)
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
		"api_url":             "https://dev.azure.com",
		"http_url":            "https://dev.azure.com",
		"oauth_token":         "token",
		"organization_scoped": true,
		"service_provider":    "ado_services",
	})

	if _, err := client.API.Organizations().ByOrganization_name("my-org").OauthClients().Post(ctx, newADOServiceOAuthClientEnvelope(d, true), nil); err != nil {
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
		if value, present := payload.Data.Attributes["oauth-token-string"]; present {
			t.Errorf("expected oauth-token-string to be omitted, got %#v", value)
		}

		w.Header().Set("Content-Type", "application/vnd.api+json")
		fmt.Fprint(w, `{"data":{"id":"oc-123","type":"oauth-clients","attributes":{"ado-org-name":null}}}`)
	}))
	d := schema.TestResourceDataRaw(t, resourceTFEOAuthClient().Schema, map[string]interface{}{
		"organization_scoped": true,
		"service_provider":    "ado_services",
	})
	d.SetId("oc-123")

	if _, err := client.API.OauthClients().ByOauth_client_id(d.Id()).Patch(ctx, newADOServiceOAuthClientEnvelope(d, false), nil); err != nil {
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

func TestReadOAuthClientADOOrgNameIncludesAPIErrorDetail(t *testing.T) {
	client := testTfeClientV2(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprint(w, `{"errors":[{"status":"422","detail":"The PAT cannot access this Azure DevOps organization"}]}`)
	}))

	_, err := readOAuthClientADOOrgName(ConfiguredClient{ClientV2: client}, "oc-123")
	if err == nil || !strings.Contains(err.Error(), "The PAT cannot access this Azure DevOps organization") {
		t.Fatalf("expected API error detail, got %v", err)
	}
}

func testOAuthClientConfiguredClient(t *testing.T, handler http.Handler) ConfiguredClient {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v2/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	mux.Handle("/", handler)
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	client, err := tfe.NewClient(&tfe.Config{
		Address: server.URL,
		Token:   "not-a-token",
	})
	if err != nil {
		t.Fatal(err)
	}
	return ConfiguredClient{Client: client, ClientV2: testTfeClientV2(t, mux)}
}

func testOAuthClientResponse(serviceProvider, adoOrgName string, list bool) string {
	data := fmt.Sprintf(`{
		"id":"oc-123","type":"oauth-clients",
		"attributes":{
			"name":"my-client",
			"service-provider":%q,
			"ado-org-name":%q,
			"api-url":"https://dev.azure.com",
			"http-url":"https://dev.azure.com",
			"organization-scoped":true
		},
		"relationships":{
			"organization":{"data":{"id":"my-org","type":"organizations"}},
			"oauth-tokens":{"data":[{"id":"ot-123","type":"oauth-tokens"}]}
		}
	}`, serviceProvider, adoOrgName)
	if list {
		data = "[" + data + "]"
	}
	return fmt.Sprintf(`{"data":%s,"included":[
		{"id":"my-org","type":"organizations","attributes":{"name":"my-org"}}
	]}`, data)
}

func TestTFEOAuthClientReadADOOrgNameByServiceProvider(t *testing.T) {
	for _, serviceProvider := range []string{"github", "gitlab_hosted", "bitbucket_hosted", "ado_services"} {
		for _, lookup := range []string{"resource", "id", "name", "service_provider"} {
			t.Run(serviceProvider+"/"+lookup, func(t *testing.T) {
				var mu sync.Mutex
				requests := 0
				config := testOAuthClientConfiguredClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					mu.Lock()
					defer mu.Unlock()
					requests++
					list := r.URL.Path == "/api/v2/organizations/my-org/oauth-clients"
					if r.Method != http.MethodGet || (!list && r.URL.Path != "/api/v2/oauth-clients/oc-123") {
						t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
						http.Error(w, "unexpected request", http.StatusBadRequest)
						return
					}
					w.Header().Set("Content-Type", "application/vnd.api+json")
					fmt.Fprint(w, testOAuthClientResponse(serviceProvider, "my-company", list))
				}))

				res := dataSourceTFEOAuthClient()
				read := dataSourceTFEOAuthClientRead
				values := map[string]interface{}{"organization": "my-org"}
				switch lookup {
				case "resource":
					res = resourceTFEOAuthClient()
					read = resourceTFEOAuthClientRead
				case "id":
					values["oauth_client_id"] = "oc-123"
				case "name":
					values["name"] = "my-client"
				case "service_provider":
					values["service_provider"] = serviceProvider
				}
				d := schema.TestResourceDataRaw(t, res.Schema, values)
				d.SetId("oc-123")
				if err := d.Set("ado_org_name", "stale-company"); err != nil {
					t.Fatal(err)
				}
				if err := read(d, config); err != nil {
					t.Fatal(err)
				}
				wantName, wantRequests := "", 1
				if serviceProvider == "ado_services" {
					wantName, wantRequests = "my-company", 2
				}
				if got := d.Get("ado_org_name").(string); got != wantName {
					t.Errorf("expected ado_org_name %q, got %q", wantName, got)
				}
				mu.Lock()
				defer mu.Unlock()
				if requests != wantRequests {
					t.Errorf("expected %d API requests, got %d", wantRequests, requests)
				}
			})
		}
	}
}

func TestTFEOAuthClientADOOrgNameLifecycle(t *testing.T) {
	var mu sync.Mutex
	var adoOrgName string
	var exists bool
	creates, updates, deletes := 0, 0, 0
	config := testOAuthClientConfiguredClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		w.Header().Set("Content-Type", "application/vnd.api+json")
		if r.Method == http.MethodPost {
			if r.URL.Path != "/api/v2/organizations/my-org/oauth-clients" || exists {
				t.Errorf("unexpected create request %s (exists: %t)", r.URL.Path, exists)
				http.Error(w, "unexpected create", http.StatusBadRequest)
				return
			}
		} else if r.URL.Path != "/api/v2/oauth-clients/oc-123" || !exists {
			http.Error(w, `{"errors":[{"status":"404"}]}`, http.StatusNotFound)
			return
		}
		switch r.Method {
		case http.MethodPost, http.MethodPatch:
			var payload struct {
				Data struct {
					Attributes map[string]interface{} `json:"attributes"`
				} `json:"data"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Errorf("invalid request body: %v", err)
				http.Error(w, "invalid request", http.StatusBadRequest)
				return
			}
			value, present := payload.Data.Attributes["ado-org-name"]
			if !present {
				t.Error("missing ado-org-name in write request")
				http.Error(w, "missing ado-org-name", http.StatusBadRequest)
				return
			}
			if value == nil {
				adoOrgName = ""
			} else {
				var ok bool
				adoOrgName, ok = value.(string)
				if !ok || adoOrgName == "" {
					t.Errorf("expected a nonempty ado-org-name or null, got %#v", value)
					http.Error(w, "invalid ado-org-name", http.StatusBadRequest)
					return
				}
			}
			if got := payload.Data.Attributes["oauth-token-string"]; got != "not-a-pat" {
				t.Errorf("expected oauth-token-string not-a-pat, got %#v", got)
			}
			if r.Method == http.MethodPost {
				if got := payload.Data.Attributes["service-provider"]; got != "ado_services" {
					t.Errorf("expected service-provider ado_services, got %#v", got)
				}
				exists = true
				creates++
				w.WriteHeader(http.StatusCreated)
			} else {
				updates++
			}
		case http.MethodGet:
		case http.MethodDelete:
			exists = false
			deletes++
			w.WriteHeader(http.StatusNoContent)
			return
		default:
			t.Errorf("unexpected request method %s", r.Method)
			http.Error(w, "unexpected method", http.StatusMethodNotAllowed)
			return
		}
		fmt.Fprint(w, testOAuthClientResponse("ado_services", adoOrgName, false))
	}))

	step := func(configuredName *string, expectedName string, wantUpdates int) resource.TestStep {
		attribute := ""
		if configuredName != nil {
			attribute = fmt.Sprintf("ado_org_name = %q", *configuredName)
		}
		return resource.TestStep{
			Config: fmt.Sprintf(`
resource "tfe_oauth_client" "test" {
  organization     = "my-org"
  api_url          = "https://dev.azure.com"
  http_url         = "https://dev.azure.com"
  service_provider = "ado_services"
  oauth_token      = "not-a-pat"
  %s
}

data "tfe_oauth_client" "test" {
  oauth_client_id = tfe_oauth_client.test.id
  depends_on     = [tfe_oauth_client.test]
}
`, attribute),
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("tfe_oauth_client.test", "ado_org_name", expectedName),
				resource.TestCheckResourceAttr("data.tfe_oauth_client.test", "ado_org_name", expectedName),
				resource.TestCheckResourceAttr("tfe_oauth_client.test", "oauth_token_id", "ot-123"),
				resource.TestCheckResourceAttr("data.tfe_oauth_client.test", "oauth_token_id", "ot-123"),
				func(_ *terraform.State) error {
					mu.Lock()
					defer mu.Unlock()
					if creates != 1 || updates != wantUpdates || adoOrgName != expectedName {
						return fmt.Errorf("expected one create, %d updates, and ado-org-name %q; got %d creates, %d updates, and %q",
							wantUpdates, expectedName, creates, updates, adoOrgName)
					}
					return nil
				},
			),
		}
	}
	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		ProtoV5ProviderFactories: map[string]func() (tfprotov5.ProviderServer, error){
			"tfe": func() (tfprotov5.ProviderServer, error) {
				p := Provider()
				p.ConfigureContextFunc = func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
					return config, nil
				}
				if err := p.InternalValidate(); err != nil {
					return nil, err
				}
				return p.GRPCProvider(), nil
			},
		},
		Steps: []resource.TestStep{
			step(ptr("my-company"), "my-company", 0),
			step(nil, "my-company", 0),
			step(ptr("other-company"), "other-company", 1),
			step(nil, "other-company", 1),
			step(ptr(""), "", 2),
		},
		CheckDestroy: func(_ *terraform.State) error {
			mu.Lock()
			defer mu.Unlock()
			if exists || deletes != 1 {
				return fmt.Errorf("expected the OAuth client to be deleted once; exists: %t, deletes: %d", exists, deletes)
			}
			return nil
		},
	})
}

func TestTFEOAuthClientEmptyADOOrgNameUsesLegacyUpdateForOtherProviders(t *testing.T) {
	var mu sync.Mutex
	var exists bool
	updates := 0
	config := testOAuthClientConfiguredClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		w.Header().Set("Content-Type", "application/vnd.api+json")

		if r.Method == http.MethodPost {
			exists = true
		} else if r.URL.Path != "/api/v2/oauth-clients/oc-123" || !exists {
			http.Error(w, `{"errors":[{"status":"404"}]}`, http.StatusNotFound)
			return
		}

		switch r.Method {
		case http.MethodPost, http.MethodPatch:
			var payload struct {
				Data struct {
					Attributes map[string]interface{} `json:"attributes"`
				} `json:"data"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Errorf("invalid request body: %v", err)
				http.Error(w, "invalid request", http.StatusBadRequest)
				return
			}
			if _, present := payload.Data.Attributes["ado-org-name"]; present {
				t.Error("expected non-ADO request to omit ado-org-name")
				http.Error(w, "unexpected ado-org-name", http.StatusBadRequest)
				return
			}
			if r.Method == http.MethodPatch {
				updates++
			} else {
				w.WriteHeader(http.StatusCreated)
			}
		case http.MethodGet:
		case http.MethodDelete:
			exists = false
			w.WriteHeader(http.StatusNoContent)
			return
		default:
			t.Errorf("unexpected request method %s", r.Method)
			http.Error(w, "unexpected method", http.StatusMethodNotAllowed)
			return
		}
		fmt.Fprint(w, `{
			"data":{
				"id":"oc-123",
				"type":"oauth-clients",
				"attributes":{
					"service-provider":"github",
					"api-url":"https://api.github.com",
					"http-url":"https://github.com",
					"organization-scoped":true
				},
				"relationships":{
					"organization":{"data":{"id":"my-org","type":"organizations"}},
					"oauth-tokens":{"data":[{"id":"ot-123","type":"oauth-tokens"}]}
				}
			},
			"included":[
				{"id":"my-org","type":"organizations","attributes":{"name":"my-org"}}
			]
		}`)
	}))

	providerFactories := map[string]func() (tfprotov5.ProviderServer, error){
		"tfe": func() (tfprotov5.ProviderServer, error) {
			p := Provider()
			p.ConfigureContextFunc = func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
				return config, nil
			}
			if err := p.InternalValidate(); err != nil {
				return nil, err
			}
			return p.GRPCProvider(), nil
		},
	}
	testConfig := func(token string) string {
		return fmt.Sprintf(`
resource "tfe_oauth_client" "test" {
  organization     = "my-org"
  api_url          = "https://api.github.com"
  http_url         = "https://github.com"
  service_provider = "github"
  oauth_token      = %q
  ado_org_name     = ""
}
`, token)
	}

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV5ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{Config: testConfig("token-one")},
			{
				Config: testConfig("token-two"),
				Check: func(_ *terraform.State) error {
					mu.Lock()
					defer mu.Unlock()
					if updates != 1 {
						return fmt.Errorf("expected one legacy update, got %d", updates)
					}
					return nil
				},
			},
		},
	})
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
