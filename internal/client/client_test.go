// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/hashicorp/go-tfe"
	svchost "github.com/hashicorp/terraform-svchost"
	"github.com/hashicorp/terraform-svchost/disco"
)

// testToken has to be used against the fake server when making an API call, otherwise
// a 404 error is returned.
var testToken = "test-token-1234567890"

// testDefaultRequestHandlers is a map of request handlers intended to be used in a request
// multiplexer for a test server. A caller may use testServer to start a server with
// this base set of routes.
var testDefaultRequestHandlers = map[string]func(http.ResponseWriter, *http.Request){
	// Respond to service discovery calls.
	"/.well-known/terraform.json": func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{
	"tfe.v2": "/api/v2/",
	"tfe.v2.1": "/api/v2/",
	"tfe.v2.2": "/api/v2/"
}`)
	},

	// Respond to pings to get the API version header.
	"/api/v2/ping": func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("TFP-API-Version", "2.5")
	},

	"/api/v2/organizations": func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("TFP-API-Version", "2.5")

		if r.Header["Authorization"][0] != fmt.Sprintf("Bearer %s", testToken) {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Write([]byte(`{"data": []}`))
	},
}

func testServer(t *testing.T) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()
	for route, handler := range testDefaultRequestHandlers {
		mux.HandleFunc(route, handler)
	}

	return httptest.NewTLSServer(mux)
}

func Test_GetClient(t *testing.T) {
	srv := testServer(t)
	t.Cleanup(func() {
		srv.Close()
	})

	serverURL, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatalf("Unexpected error when parsing testServer URL: %q", err)
	}

	cliConfig, err := os.CreateTemp("", "cliconfig")
	if err != nil {
		t.Fatalf("Failed to create temp CLI config: %s", err)
	}
	t.Cleanup(func() {
		os.Remove(cliConfig.Name())
	})

	fmt.Fprintf(cliConfig, `
credentials "%s" {
	token = "%s"
}`, serverURL.Host, testToken)

	cases := map[string]struct {
		env               map[string]string
		hostname          string
		token             string
		expectMissingAuth bool
		expectTokenSource tokenSource
	}{
		"everything from env": {
			env: map[string]string{
				"TFE_HOSTNAME": serverURL.Host,
				"TFE_TOKEN":    testToken,
			},
			expectTokenSource: environmentVariable,
		},
		"token from env": {
			env: map[string]string{
				"TFE_HOSTNAME": serverURL.Host,
				"TFE_TOKEN":    "",
			},
			token:             testToken,
			expectTokenSource: providerArgument,
		},
		"everything from provider config": {
			env: map[string]string{
				"TFE_HOSTNAME": "",
				"TFE_TOKEN":    "",
			},
			hostname:          serverURL.Host,
			token:             testToken,
			expectTokenSource: providerArgument,
		},
		"token missing": {
			env: map[string]string{
				"TFE_HOSTNAME": "",
				"TFE_TOKEN":    "",
			},
			hostname:          serverURL.Host,
			expectMissingAuth: true,
		},
		"token from CLI config": {
			env: map[string]string{
				"TFE_TOKEN":          "",
				"TF_CLI_CONFIG_FILE": cliConfig.Name(),
			},
			hostname:          serverURL.Host,
			expectTokenSource: credentialFiles,
		},
	}

	for _, c := range cases {
		for k, v := range c.env {
			t.Setenv(k, v)
		}
		// Must always skip SSL verification for this test server
		providerClient, err := GetClient(c.hostname, c.token, true)

		if c.expectMissingAuth {
			if !errors.Is(err, ErrMissingAuthToken) {
				t.Errorf("Expected ErrMissingAuthToken, got %v", err)
			}
			continue
		}

		if err != nil {
			t.Errorf("Unexpected error when getting client: %q", err)
		}

		client := providerClient.TfeClient
		if client == nil {
			t.Fatal("Unexpected client was nil")
		}

		if providerClient.Hostname != serverURL.Host {
			t.Fatalf("Expected hostname %q, got %q", serverURL.Host, providerClient.Hostname)
		}

		tokenSource := providerClient.tokenSource
		if tokenSource != c.expectTokenSource {
			t.Fatalf("Expected token source %d, got %d", c.expectTokenSource, tokenSource)
		}

		_, err = client.Organizations.List(context.Background(), &tfe.OrganizationListOptions{})
		if err != nil {
			t.Errorf("Unexpected error from using client: %q", err)
		}
	}
}

func TestClient_sendAuthenticationWarning(t *testing.T) {
	// This tests that the SendAuthenticationWarning function returns true when the
	// token source is credentialFiles and the TFE_AGENT_VERSION env var is set
	cases := map[string]struct {
		tokenSource                   tokenSource
		tfcAgentVersionEnvVariableSet bool
		expectResult                  bool
	}{
		"token from credentials files and TFC_AGENT_VERSION is set": {
			tokenSource:                   credentialFiles,
			tfcAgentVersionEnvVariableSet: true,
			expectResult:                  true,
		},
		"token from credentials files but TFC_AGENT_VERSION not set": {
			tokenSource:                   credentialFiles,
			tfcAgentVersionEnvVariableSet: false,
			expectResult:                  false,
		},
		"TFC_AGENT_VERSION is set but token not from credentials files": {
			tokenSource:                   providerArgument,
			tfcAgentVersionEnvVariableSet: true,
			expectResult:                  false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if tc.tfcAgentVersionEnvVariableSet {
				t.Setenv("TFC_AGENT_VERSION", "1.0")
			}

			providerClient := ProviderClient{
				TfeClient:   nil,
				tokenSource: tc.tokenSource,
			}

			result := providerClient.SendAuthenticationWarning()
			if result != tc.expectResult {
				t.Fatalf("%s: SendAuthenticationWarning() expected result: %t, got %t", name, tc.expectResult, result)
			}
		})
	}
}

func TestClient_canonicalHostname(t *testing.T) {
	const proxy = "http://127.0.0.1:40591/api/v2/"
	proxiedAPI := map[string]interface{}{
		"tfe.v2":   proxy,
		"tfe.v2.1": proxy,
		"tfe.v2.2": proxy,
	}
	mergeServices := func(base, extra map[string]interface{}) map[string]interface{} {
		services := map[string]interface{}{}
		for k, v := range base {
			services[k] = v
		}
		for k, v := range extra {
			services[k] = v
		}
		return services
	}
	registry := func(host string) map[string]interface{} {
		return map[string]interface{}{
			"modules.v1":   "https://" + host + "/api/registry/v1/modules/",
			"providers.v1": "https://" + host + "/api/registry/v1/providers/",
		}
	}

	cases := map[string]struct {
		configured string
		services   map[string]interface{}
		expected   string
	}{
		"configured host with proxied API": {
			configured: "app.terraform.io",
			services:   mergeServices(proxiedAPI, registry("app.terraform.io")),
			expected:   "app.terraform.io",
		},
		"configured host is normalized": {
			configured: "TFE.Example.com:443",
			services:   mergeServices(proxiedAPI, nil),
			expected:   "tfe.example.com",
		},
		"configured host keeps non-default port": {
			configured: "tfe.example.com:8443",
			services:   mergeServices(proxiedAPI, nil),
			expected:   "tfe.example.com:8443",
		},
		"generic hostname with proxied API (HYOK)": {
			configured: genericHostname,
			services:   mergeServices(proxiedAPI, registry("tfe.example.com")),
			expected:   "tfe.example.com",
		},
		"generic hostname without proxy": {
			configured: genericHostname,
			services:   registry("app.terraform.io"),
			expected:   "app.terraform.io",
		},
		"generic hostname keeps non-default registry port": {
			configured: genericHostname,
			services:   mergeServices(proxiedAPI, registry("tfe.example.com:8443")),
			expected:   "tfe.example.com:8443",
		},
		"generic hostname with only providers.v1": {
			configured: genericHostname,
			services: mergeServices(proxiedAPI, map[string]interface{}{
				"providers.v1": "https://tfe.example.com/api/registry/v1/providers/",
			}),
			expected: "tfe.example.com",
		},
		"generic hostname ignores loopback registry": {
			configured: genericHostname,
			services:   mergeServices(proxiedAPI, registry("127.0.0.1:40591")),
			expected:   genericHostname,
		},
		"generic hostname without registry services": {
			configured: genericHostname,
			services:   proxiedAPI,
			expected:   genericHostname,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			configured, err := svchost.ForComparison(tc.configured)
			if err != nil {
				t.Fatalf("invalid hostname %q: %s", tc.configured, err)
			}

			services := disco.New()
			services.ForceHostServices(configured, tc.services)
			host, err := services.Discover(configured)
			if err != nil {
				t.Fatalf("unexpected discovery error: %s", err)
			}

			if got := canonicalHostname(configured, host); got != tc.expected {
				t.Fatalf("expected hostname %q, got %q", tc.expected, got)
			}
		})
	}
}
