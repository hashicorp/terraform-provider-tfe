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
	}{
		"everything from env": {
			env: map[string]string{
				"TFE_HOSTNAME": serverURL.Host,
				"TFE_TOKEN":    testToken,
			},
		},
		"token from env": {
			env: map[string]string{
				"TFE_HOSTNAME": serverURL.Host,
				"TFE_TOKEN":    "",
			},
			token: testToken,
		},
		"everything from provider config": {
			env: map[string]string{
				"TFE_HOSTNAME": "",
				"TFE_TOKEN":    "",
			},
			hostname: serverURL.Host,
			token:    testToken,
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
			hostname: serverURL.Host,
		},
	}

	for _, c := range cases {
		for k, v := range c.env {
			t.Setenv(k, v)
		}
		// Must always skip SSL verification for this test server
		client, err := GetClient(c.hostname, c.token, true)
		if c.expectMissingAuth {
			if !errors.Is(err, ErrMissingAuthToken) {
				t.Errorf("Expected ErrMissingAuthToken, got %v", err)
			}
			continue
		}

		if err != nil {
			t.Errorf("Unexpected error when getting client: %q", err)
		}

		if client == nil {
			t.Fatal("Unexpected client was nil")
		}

		_, err = client.Organizations.List(context.Background(), &tfe.OrganizationListOptions{})
		if err != nil {
			t.Errorf("Unexpected error from using client: %q", err)
		}
	}
}
