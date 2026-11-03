// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestLegacyCallbacksHonorCancellation(t *testing.T) {
	for _, tc := range []struct {
		name     string
		resource *schema.Resource
		values   map[string]interface{}
	}{
		{"agent pool", dataSourceTFEAgentPool(), map[string]interface{}{"name": "pool", "organization": "org"}},
		{"github installation", dataSourceTFEGHAInstallation(), map[string]interface{}{"name": "app"}},
		{"oauth client", dataSourceTFEOAuthClient(), map[string]interface{}{"oauth_client_id": "oc-test"}},
		{"admin organization", resourceTFEAdminOrganizationSettings(), map[string]interface{}{"organization": "org"}},
		{"allowed projects", resourceTFEAgentPoolAllowedProjects(), map[string]interface{}{"agent_pool_id": "ap-test"}},
		{"allowed workspaces", resourceTFEAgentPoolAllowedWorkspaces(), map[string]interface{}{"agent_pool_id": "ap-test"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.resource.ReadContext == nil {
				t.Fatal("resource must expose a context-aware read callback")
			}
			var requests atomic.Int32
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests.Add(1)
				http.Error(w, "unexpected request", http.StatusBadRequest)
			})
			server := httptest.NewServer(handler)
			defer server.Close()
			client, err := tfe.NewClient(&tfe.Config{Address: server.URL, Token: "test-token"})
			if err != nil {
				t.Fatal(err)
			}
			config := ConfiguredClient{Client: client, ClientV2: testTfeClientV2(t, handler), Organization: "org"}
			// Client construction performs a discovery request before callbacks run.
			requests.Store(0)
			for name, callback := range map[string]func(context.Context, *schema.ResourceData, interface{}) diag.Diagnostics{"create": tc.resource.CreateContext, "read": tc.resource.ReadContext, "update": tc.resource.UpdateContext, "delete": tc.resource.DeleteContext} {
				if callback == nil {
					continue
				}
				// Deleting admin settings only drops local state and performs no request.
				if tc.name == "admin organization" && name == "delete" {
					continue
				}
				t.Run(name, func(t *testing.T) {
					d := schema.TestResourceDataRaw(t, tc.resource.Schema, tc.values)
					d.SetId("test-id")
					ctx, cancel := context.WithCancel(context.Background())
					cancel()
					diags := callback(ctx, d, config)
					if !diags.HasError() || !strings.Contains(fmt.Sprint(diags), "context canceled") {
						t.Fatalf("expected cancellation diagnostic, got %v", diags)
					}
					if got := requests.Load(); got != 0 {
						t.Fatalf("canceled operation sent %d HTTP requests", got)
					}
				})
			}
		})
	}
}
