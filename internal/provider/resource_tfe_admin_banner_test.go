// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	tfe "github.com/hashicorp/go-tfe"
	tfev2 "github.com/hashicorp/go-tfe/v2"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const testAdminBannerResourceName = "tfe_admin_banner.foobar"

// FLAKE ALERT: The admin banner is a singleton resource shared by the entire
// TFE instance, and any test touching it is at high risk to flake.
// In order for these tests to be safe, the following requirements MUST be met:
//  1. All test cases for this resource must run within a SINGLE test func, using
//     t.Run to separate the individual test cases.
//  2. The inner sub-tests must not call t.Parallel.
//
// If these tests are split into multiple test funcs and they get allocated to
// different test runner partitions in CI, then they will inevitably flake, as
// tests running concurrently in different containers will be competing to set
// the same shared global state in the TFE instance.
func TestAccTFEAdminBanner_omnibus(t *testing.T) {
	skipIfCloud(t)

	t.Run("basic banner with defaults, update, and import", func(t *testing.T) {
		var firstID string
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			CheckDestroy:             testAccCheckTFEAdminBannerDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccTFEAdminBanner_basic("Maintenance window", "Planned <strong>maintenance</strong>."),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckTFEAdminBannerRecordID(&firstID),
						resource.TestCheckResourceAttrSet(testAdminBannerResourceName, "id"),
						resource.TestCheckResourceAttr(testAdminBannerResourceName, "title", "Maintenance window"),
						resource.TestCheckResourceAttr(testAdminBannerResourceName, "body", "Planned <strong>maintenance</strong>."),
						resource.TestCheckResourceAttr(testAdminBannerResourceName, "style", "info"),
						resource.TestCheckResourceAttr(testAdminBannerResourceName, "audience", "authenticated_only"),
						resource.TestCheckResourceAttrSet(testAdminBannerResourceName, "published_at"),
						resource.TestCheckNoResourceAttr(testAdminBannerResourceName, "scheduled_publish_at"),
						resource.TestCheckNoResourceAttr(testAdminBannerResourceName, "timezone"),
					),
				},
				{
					// Updating any attribute updates the banner in place.
					Config: testAccTFEAdminBanner_full("Critical outage", "Service degraded.", "critical", "authenticated_only"),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckTFEAdminBannerIDEquals(&firstID),
						resource.TestCheckResourceAttr(testAdminBannerResourceName, "title", "Critical outage"),
						resource.TestCheckResourceAttr(testAdminBannerResourceName, "body", "Service degraded."),
						resource.TestCheckResourceAttr(testAdminBannerResourceName, "style", "critical"),
						resource.TestCheckResourceAttr(testAdminBannerResourceName, "audience", "authenticated_only"),
					),
				},
				{
					ResourceName:      testAdminBannerResourceName,
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})

	t.Run("scheduled banner can be added and removed", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			CheckDestroy:             testAccCheckTFEAdminBannerDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccTFEAdminBanner_scheduled("2099-01-01T10:00:00Z", "2099-01-01T12:00:00Z", "America/New_York"),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(testAdminBannerResourceName, "scheduled_publish_at", "2099-01-01T10:00:00Z"),
						resource.TestCheckResourceAttr(testAdminBannerResourceName, "scheduled_expire_at", "2099-01-01T12:00:00Z"),
						resource.TestCheckResourceAttr(testAdminBannerResourceName, "timezone", "America/New_York"),
					),
				},
				{
					Config: testAccTFEAdminBanner_basic("Scheduled", "No longer scheduled."),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckNoResourceAttr(testAdminBannerResourceName, "scheduled_publish_at"),
						resource.TestCheckNoResourceAttr(testAdminBannerResourceName, "scheduled_expire_at"),
						resource.TestCheckNoResourceAttr(testAdminBannerResourceName, "timezone"),
					),
				},
			},
		})
	})

	t.Run("invalid configurations are rejected at plan time", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			Steps: []resource.TestStep{
				{
					Config:      testAccTFEAdminBanner_partialSchedule(),
					PlanOnly:    true,
					ExpectError: regexp.MustCompile(`(?s)timezone.*scheduled_expire_at|scheduled_expire_at.*timezone`),
				},
				{
					Config:      testAccTFEAdminBanner_scheduled("2099-01-01T12:00:00Z", "2099-01-01T10:00:00Z", "UTC"),
					PlanOnly:    true,
					ExpectError: regexp.MustCompile(`scheduled_expire_at must be after\s+scheduled_publish_at`),
				},
				{
					Config:      testAccTFEAdminBanner_full("Bad style", "Body", "fatal", "all_users"),
					PlanOnly:    true,
					ExpectError: regexp.MustCompile(`style`),
				},
				{
					Config:      testAccTFEAdminBanner_full("Bad audience", "Body", "info", "nobody"),
					PlanOnly:    true,
					ExpectError: regexp.MustCompile(`audience`),
				},
				{
					// id is Computed-only; declaring it in config must be rejected at plan.
					Config:      testAccTFEAdminBanner_withID(),
					PlanOnly:    true,
					ExpectError: regexp.MustCompile(`read.only|Invalid Configuration`),
				},
			},
		})
	})

	t.Run("out-of-band deactivation removes resource from state", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			CheckDestroy:             testAccCheckTFEAdminBannerDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccTFEAdminBanner_basic("Drift", "Drift test."),
					Check:  resource.TestCheckResourceAttrSet(testAdminBannerResourceName, "id"),
				},
				{
					// Deactivate the banner out-of-band by triggering a destroy-and-recreate
					// via a refresh; the provider should remove it from state, and the
					// next plan should propose to create it again.
					RefreshState:       true,
					ExpectNonEmptyPlan: true,
					PreConfig: func() {
						client, err := getClientV2UsingEnv()
						if err != nil {
							return
						}
						banners, err := client.API.Admin().Banners().Get(context.Background(), nil)
						if err != nil || banners == nil {
							return
						}
						for _, b := range banners.GetData() {
							_ = client.API.Admin().Banners().ById(valueOrZero(b.GetId())).Delete(context.Background(), nil)
						}
					},
				},
			},
		})
	})

	t.Run("a second banner fails with a conflict", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			CheckDestroy:             testAccCheckTFEAdminBannerDestroy,
			Steps: []resource.TestStep{
				{
					Config:      testAccTFEAdminBanner_two(),
					ExpectError: regexp.MustCompile(`409 Conflict`),
				},
			},
		})
	})
}

func testAccCheckTFEAdminBannerRecordID(id *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[testAdminBannerResourceName]
		if !ok {
			return fmt.Errorf("resource %s not found in state", testAdminBannerResourceName)
		}
		*id = rs.Primary.ID
		return nil
	}
}

// testAccCheckTFEAdminBannerIDEquals fails if the banner was replaced rather
// than updated in place.
func testAccCheckTFEAdminBannerIDEquals(id *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[testAdminBannerResourceName]
		if !ok {
			return fmt.Errorf("resource %s not found in state", testAdminBannerResourceName)
		}
		if rs.Primary.ID != *id {
			return fmt.Errorf("banner was replaced: id changed from %s to %s", *id, rs.Primary.ID)
		}
		return nil
	}
}

func testAccCheckTFEAdminBannerDestroy(s *terraform.State) error {
	client, err := getClientV2UsingEnv()
	if err != nil {
		return err
	}

	banners, err := client.API.Admin().Banners().Get(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("error listing admin banners: %w", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_admin_banner" {
			continue
		}
		for _, b := range banners.GetData() {
			if valueOrZero(b.GetId()) == rs.Primary.ID {
				return fmt.Errorf("admin banner %s still active", rs.Primary.ID)
			}
		}
	}
	return nil
}

func testAccTFEAdminBanner_basic(title, body string) string {
	return fmt.Sprintf(`
resource "tfe_admin_banner" "foobar" {
  title = %q
  body  = %q
}
`, title, body)
}

func testAccTFEAdminBanner_full(title, body, style, audience string) string {
	return fmt.Sprintf(`
resource "tfe_admin_banner" "foobar" {
  title    = %q
  body     = %q
  style    = %q
  audience = %q
}
`, title, body, style, audience)
}

func testAccTFEAdminBanner_withID() string {
	return `
resource "tfe_admin_banner" "foobar" {
  id    = "some-id"
  title = "Test"
  body  = "Test."
}
`
}

func testAccTFEAdminBanner_scheduled(publishAt, expireAt, timezone string) string {
	return fmt.Sprintf(`
resource "tfe_admin_banner" "foobar" {
  title                = "Scheduled"
  body                 = "Scheduled maintenance."
  scheduled_publish_at = %q
  scheduled_expire_at  = %q
  timezone             = %q
}
`, publishAt, expireAt, timezone)
}

func testAccTFEAdminBanner_partialSchedule() string {
	return `
resource "tfe_admin_banner" "foobar" {
  title                = "Partial"
  body                 = "Only one schedule attribute set."
  scheduled_publish_at = "2099-01-01T10:00:00Z"
}
`
}

func testAccTFEAdminBanner_two() string {
	return `
resource "tfe_admin_banner" "foobar" {
  title = "First"
  body  = "First banner."
}

resource "tfe_admin_banner" "second" {
  title      = "Second"
  body       = "Second banner."
  depends_on = [tfe_admin_banner.foobar]
}
`
}

// The tests below need no TFE instance; they exercise the model conversion and
// wire format against a fake API server.

func TestModelFromAdminBanner(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v2/admin/banners", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		fmt.Fprint(w, `{"data": [{
			"id": "ban-123",
			"type": "admin-banners",
			"attributes": {
				"title": "Maintenance",
				"body": "Down <b>soon</b>",
				"style": "warning",
				"audience": "authenticated_only",
				"scheduled-publish-at": "2099-01-01T10:00:00Z",
				"scheduled-expire-at": "2099-01-01T12:00:00Z",
				"timezone": "UTC",
				"published-at": "2099-01-01T10:00:00Z"
			}
		}]}`)
	})
	client := testTfeClientV2(t, mux)

	resp, err := client.API.Admin().Banners().Get(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.GetData()) != 1 {
		t.Fatalf("expected 1 banner, got %d", len(resp.GetData()))
	}

	got := modelFromAdminBanner(resp.GetData()[0])

	checks := []struct{ name, got, want string }{
		{"id", got.ID.ValueString(), "ban-123"},
		{"title", got.Title.ValueString(), "Maintenance"},
		{"body", got.Body.ValueString(), "Down <b>soon</b>"},
		{"style", got.Style.ValueString(), "warning"},
		{"audience", got.Audience.ValueString(), "authenticated_only"},
		{"timezone", got.Timezone.ValueString(), "UTC"},
		{"scheduled_publish_at", got.ScheduledPublishAt.ValueString(), "2099-01-01T10:00:00Z"},
		{"scheduled_expire_at", got.ScheduledExpireAt.ValueString(), "2099-01-01T12:00:00Z"},
		{"published_at", got.PublishedAt.ValueString(), "2099-01-01T10:00:00Z"},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s: got %q, want %q", c.name, c.got, c.want)
		}
	}
}

func TestModelFromAdminBanner_noSchedule(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v2/admin/banners", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		fmt.Fprint(w, `{"data": [{"id": "ban-1", "type": "admin-banners", "attributes": {"title": "t", "body": "b", "style": "info", "audience": "all_users", "published-at": "2026-01-01T00:00:00Z"}}]}`)
	})
	client := testTfeClientV2(t, mux)

	resp, err := client.API.Admin().Banners().Get(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := modelFromAdminBanner(resp.GetData()[0])
	if !got.ScheduledPublishAt.IsNull() || !got.ScheduledExpireAt.IsNull() || !got.Timezone.IsNull() {
		t.Errorf("expected schedule attributes to be null, got publish=%v expire=%v tz=%v",
			got.ScheduledPublishAt, got.ScheduledExpireAt, got.Timezone)
	}
	if got.PublishedAt.IsNull() {
		t.Error("expected published_at to be set")
	}
}

func TestBannerEnvelopeFromModel_requestBody(t *testing.T) {
	nullTime := timetypes.NewRFC3339Null()
	publishAt, diags := timetypes.NewRFC3339Value("2099-01-01T10:00:00Z")
	if diags.HasError() {
		t.Fatal(diags)
	}
	expireAt, diags := timetypes.NewRFC3339Value("2099-01-01T12:00:00Z")
	if diags.HasError() {
		t.Fatal(diags)
	}

	tests := map[string]struct {
		model         modelTFEAdminBanner
		clearSchedule bool
		wantAttrs     map[string]any
		absentAttrs   []string
	}{
		// style/audience are Optional+Computed with provider defaults (info /
		// authenticated_only). The plan always carries those default values — never
		// Unknown — so they are always sent on the wire.
		"minimal create with provider defaults sends style and audience": {
			model: modelTFEAdminBanner{
				Title:              types.StringValue("t"),
				Body:               types.StringValue("b"),
				Style:              types.StringValue("info"),
				Audience:           types.StringValue("authenticated_only"),
				ScheduledPublishAt: nullTime,
				ScheduledExpireAt:  nullTime,
				Timezone:           types.StringNull(),
			},
			wantAttrs: map[string]any{
				"title":    "t",
				"body":     "b",
				"style":    "info",
				"audience": "authenticated_only",
			},
			absentAttrs: []string{"scheduled-publish-at", "scheduled-expire-at", "timezone"},
		},
		"full create sends every attribute": {
			model: modelTFEAdminBanner{
				Title:              types.StringValue("t"),
				Body:               types.StringValue("b"),
				Style:              types.StringValue("critical"),
				Audience:           types.StringValue("authenticated_only"),
				ScheduledPublishAt: publishAt,
				ScheduledExpireAt:  expireAt,
				Timezone:           types.StringValue("UTC"),
			},
			wantAttrs: map[string]any{
				"title":                "t",
				"body":                 "b",
				"style":                "critical",
				"audience":             "authenticated_only",
				"scheduled-publish-at": "2099-01-01T10:00:00Z",
				"scheduled-expire-at":  "2099-01-01T12:00:00Z",
				"timezone":             "UTC",
			},
		},
		"clearing the schedule sends explicit nulls": {
			model: modelTFEAdminBanner{
				Title:              types.StringValue("t"),
				Body:               types.StringValue("b"),
				Style:              types.StringValue("info"),
				Audience:           types.StringValue("all_users"),
				ScheduledPublishAt: nullTime,
				ScheduledExpireAt:  nullTime,
				Timezone:           types.StringNull(),
			},
			clearSchedule: true,
			wantAttrs: map[string]any{
				"title":                "t",
				"body":                 "b",
				"scheduled-publish-at": nil,
				"scheduled-expire-at":  nil,
				"timezone":             nil,
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var gotBody []byte
			mux := http.NewServeMux()
			mux.HandleFunc("/api/v2/admin/banners/ban-1", func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPatch {
					t.Errorf("expected PATCH, got %s", r.Method)
				}
				gotBody, _ = io.ReadAll(r.Body)
				w.Header().Set("Content-Type", "application/vnd.api+json")
				fmt.Fprint(w, `{"data": {"id": "ban-1", "type": "admin-banners", "attributes": {"title": "t", "body": "b"}}}`)
			})
			client := testTfeClientV2(t, mux)

			envelope, d := bannerEnvelopeFromModel(tc.model, tc.clearSchedule)
			if d.HasError() {
				t.Fatalf("unexpected diagnostics: %v", d)
			}
			if _, err := client.API.Admin().Banners().ById("ban-1").Patch(context.Background(), envelope, nil); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			var sent struct {
				Data struct {
					Type       string         `json:"type"`
					Attributes map[string]any `json:"attributes"`
				} `json:"data"`
			}
			if err := json.Unmarshal(gotBody, &sent); err != nil {
				t.Fatalf("request body is not valid JSON: %v\n%s", err, gotBody)
			}
			if sent.Data.Type != "admin-banners" {
				t.Errorf("unexpected type: %q", sent.Data.Type)
			}
			for k, want := range tc.wantAttrs {
				got, present := sent.Data.Attributes[k]
				if !present {
					t.Errorf("attribute %q missing from request body: %s", k, gotBody)
					continue
				}
				if got != want {
					t.Errorf("attribute %q: got %v, want %v", k, got, want)
				}
			}
			for _, k := range tc.absentAttrs {
				if _, present := sent.Data.Attributes[k]; present {
					t.Errorf("attribute %q should not be sent: %s", k, gotBody)
				}
			}
		})
	}
}

func TestAdminBannerFindBanner(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v2/admin/banners", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		fmt.Fprint(w, `{"data": [{"id": "ban-1", "type": "admin-banners", "attributes": {"title": "t", "body": "b"}}]}`)
	})
	r := &resourceTFEAdminBanner{config: ConfiguredClient{ClientV2: testTfeClientV2(t, mux)}}

	if b, err := r.findBanner(context.Background(), "ban-1"); err != nil || b == nil {
		t.Errorf("expected to find ban-1, got banner=%v err=%v", b, err)
	}
	if b, err := r.findBanner(context.Background(), "ban-other"); err != nil || b != nil {
		t.Errorf("expected not to find ban-other, got banner=%v err=%v", b, err)
	}
}

// TestCheckBannerSupport_versionGuard exercises the version check logic used
// by checkBannerSupport. tfe_admin_banner uses MeetsMinRemoteTFEOnlyVersion,
// which calls checkTFEVersion internally (bypassing the Cloud pass-through of
// meetsMinTFEVersion). Cloud instances are rejected before the version check.
func TestCheckBannerSupport_versionGuard(t *testing.T) {
	// TFE version cases — exercise checkTFEVersion directly (same code path as
	// MeetsMinRemoteTFEOnlyVersion for non-Cloud instances).
	cases := []struct {
		name        string
		tfeVersion  string
		wantErr     bool
		errContains string
	}{
		{
			name:       "meeting version is allowed",
			tfeVersion: "2.1.0",
			wantErr:    false,
		},
		{
			name:       "newer version is allowed",
			tfeVersion: "2.5.0",
			wantErr:    false,
		},
		{
			name:        "older version is blocked",
			tfeVersion:  "2.0.1",
			wantErr:     true,
			errContains: "2.1.0",
		},
		{
			name:        "unknown version is blocked",
			tfeVersion:  "",
			wantErr:     true,
			errContains: "2.1.0",
		},
		{
			// Dev builds report a git SHA, not a semver string.
			name:       "dev build SHA is allowed",
			tfeVersion: "abc1234",
			wantErr:    false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			meets, err := checkTFEVersion(tc.tfeVersion, minTFEVersionAdminBanner)
			if err != nil {
				if !tc.wantErr {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if tc.wantErr && meets {
				t.Errorf("expected version guard to block %q, but it was allowed", tc.tfeVersion)
			}
			if !tc.wantErr && !meets {
				t.Errorf("expected version guard to allow %q, but it was blocked", tc.tfeVersion)
			}
			if tc.errContains != "" && meets {
				t.Errorf("expected error containing %q but got no error (meets=true)", tc.errContains)
			}
		})
	}

	// Cloud instances must be rejected — tfe_admin_banner is TFE-only.
	t.Run("cloud is blocked (TFE-only resource)", func(t *testing.T) {
		tfeClient, err := tfe.NewClient(&tfe.Config{Address: "https://app.terraform.io", Token: "test-token"})
		if err != nil {
			t.Fatalf("failed to create client: %v", err)
		}
		r := &resourceTFEAdminBanner{config: ConfiguredClient{Client: tfeClient}}
		var diags diag.Diagnostics
		if r.checkBannerSupport(&diags) {
			t.Error("expected checkBannerSupport to return false for HCP Terraform, but it returned true")
		}
		if !diags.HasError() {
			t.Error("expected an error diagnostic for HCP Terraform, got none")
		}
		for _, d := range diags {
			if d.Severity() == diag.SeverityError && strings.Contains(d.Detail(), "HCP Terraform") {
				return
			}
		}
		t.Errorf("expected error detail to mention 'HCP Terraform'; got: %v", diags)
	})
}

// newBannerHandlerTestResource returns a resource wired to a fake TFE that
// reports the given version on every response and serves mux.
func newBannerHandlerTestResource(t *testing.T, version string, mux *http.ServeMux) *resourceTFEAdminBanner {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("X-TFE-Current-Version", version)
		mux.ServeHTTP(w, req)
	}))
	t.Cleanup(server.Close)

	v1, err := tfe.NewClient(&tfe.Config{Address: server.URL, Token: "test-token"})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	v2, err := tfev2.NewClient(&tfev2.Config{Address: server.URL, Token: "test-token"})
	if err != nil {
		t.Fatalf("failed to create v2 client: %v", err)
	}
	return &resourceTFEAdminBanner{config: ConfiguredClient{Client: v1, ClientV2: v2}}
}

func bannerTestSchema(t *testing.T, r *resourceTFEAdminBanner) fwresource.SchemaResponse {
	t.Helper()
	var resp fwresource.SchemaResponse
	r.Schema(context.Background(), fwresource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("schema diagnostics: %v", resp.Diagnostics)
	}
	return resp
}

func bannerTestState(t *testing.T, r *resourceTFEAdminBanner, m modelTFEAdminBanner) tfsdk.State {
	t.Helper()
	s := bannerTestSchema(t, r).Schema
	state := tfsdk.State{Schema: s, Raw: tftypes.NewValue(s.Type().TerraformType(context.Background()), nil)}
	if diags := state.Set(context.Background(), &m); diags.HasError() {
		t.Fatalf("failed to build state: %v", diags)
	}
	return state
}

func bannerTestPlan(t *testing.T, r *resourceTFEAdminBanner, m modelTFEAdminBanner) tfsdk.Plan {
	t.Helper()
	s := bannerTestSchema(t, r).Schema
	plan := tfsdk.Plan{Schema: s, Raw: tftypes.NewValue(s.Type().TerraformType(context.Background()), nil)}
	if diags := plan.Set(context.Background(), &m); diags.HasError() {
		t.Fatalf("failed to build plan: %v", diags)
	}
	return plan
}

func bannerTestStateModel(t *testing.T, state tfsdk.State) modelTFEAdminBanner {
	t.Helper()
	var m modelTFEAdminBanner
	if diags := state.Get(context.Background(), &m); diags.HasError() {
		t.Fatalf("failed to read state: %v", diags)
	}
	return m
}

func testRFC3339(v string) timetypes.RFC3339 {
	return timetypes.NewRFC3339ValueMust(v)
}

// offsetScheduledPlan is a planned banner whose schedule uses a non-UTC offset
// and whose body contains markup that the server will strip.
func offsetScheduledPlan() modelTFEAdminBanner {
	return modelTFEAdminBanner{
		ID:                 types.StringUnknown(),
		Title:              types.StringValue("t"),
		Body:               types.StringValue("<script>x</script>ok"),
		Style:              types.StringValue("info"),
		Audience:           types.StringValue("authenticated_only"),
		ScheduledPublishAt: testRFC3339("2099-01-01T05:00:00-05:00"),
		ScheduledExpireAt:  testRFC3339("2099-01-01T07:00:00-05:00"),
		Timezone:           types.StringValue("America/New_York"),
		PublishedAt:        timetypes.NewRFC3339Unknown(),
	}
}

// utcBannerResponse is what TFE returns: sanitized body and UTC timestamps.
const utcBannerResponse = `{"data": {"id": "ban-1", "type": "admin-banners", "attributes": {
	"title": "t", "body": "ok", "style": "info", "audience": "authenticated_only",
	"scheduled-publish-at": "2099-01-01T10:00:00Z", "scheduled-expire-at": "2099-01-01T12:00:00Z",
	"timezone": "America/New_York", "published-at": "2026-01-01T00:00:00Z"}}}`

func assertBannerStateKeepsPlan(t *testing.T, got, plan modelTFEAdminBanner) {
	t.Helper()
	if !got.Body.Equal(plan.Body) {
		t.Errorf("body: got %q, want plan value %q", got.Body.ValueString(), plan.Body.ValueString())
	}
	if !got.ScheduledPublishAt.Equal(plan.ScheduledPublishAt) {
		t.Errorf("scheduled_publish_at: got %q, want plan value %q", got.ScheduledPublishAt.ValueString(), plan.ScheduledPublishAt.ValueString())
	}
	if !got.ScheduledExpireAt.Equal(plan.ScheduledExpireAt) {
		t.Errorf("scheduled_expire_at: got %q, want plan value %q", got.ScheduledExpireAt.ValueString(), plan.ScheduledExpireAt.ValueString())
	}
	if got.ID.ValueString() != "ban-1" {
		t.Errorf("id: got %q, want ban-1", got.ID.ValueString())
	}
	if got.PublishedAt.IsNull() || got.PublishedAt.IsUnknown() {
		t.Error("published_at must come from the server response")
	}
}

func TestAdminBannerCreate_handler(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v2/admin/banners", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		fmt.Fprint(w, utcBannerResponse)
	})
	r := newBannerHandlerTestResource(t, "2.1.0", mux)
	plan := offsetScheduledPlan()

	req := fwresource.CreateRequest{Plan: bannerTestPlan(t, r, plan)}
	resp := &fwresource.CreateResponse{State: bannerTestState(t, r, modelTFEAdminBanner{})}
	r.Create(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error: %v", resp.Diagnostics)
	}
	assertBannerStateKeepsPlan(t, bannerTestStateModel(t, resp.State), plan)

	if len(resp.Diagnostics.Warnings()) != 1 || !strings.Contains(resp.Diagnostics.Warnings()[0].Summary(), "modified the banner body") {
		t.Errorf("expected a single body-sanitization warning, got %v", resp.Diagnostics)
	}
}

func TestAdminBannerCreate_handler_unsupportedVersion(t *testing.T) {
	posted := false
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v2/admin/banners", func(w http.ResponseWriter, req *http.Request) { posted = true })
	r := newBannerHandlerTestResource(t, "2.0.0", mux)

	req := fwresource.CreateRequest{Plan: bannerTestPlan(t, r, offsetScheduledPlan())}
	resp := &fwresource.CreateResponse{State: bannerTestState(t, r, modelTFEAdminBanner{})}
	r.Create(context.Background(), req, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected Create to fail on an unsupported TFE version")
	}
	if posted {
		t.Error("Create must not call the API on an unsupported TFE version")
	}
}

func TestAdminBannerUpdate_handler(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v2/admin/banners/ban-1", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPatch {
			t.Errorf("unexpected method %s", req.Method)
		}
		w.Header().Set("Content-Type", "application/vnd.api+json")
		fmt.Fprint(w, utcBannerResponse)
	})
	r := newBannerHandlerTestResource(t, "2.1.0", mux)
	plan := offsetScheduledPlan()
	plan.ID = types.StringValue("ban-1")

	prior := plan
	prior.PublishedAt = testRFC3339("2026-01-01T00:00:00Z")

	req := fwresource.UpdateRequest{Plan: bannerTestPlan(t, r, plan), State: bannerTestState(t, r, prior)}
	resp := &fwresource.UpdateResponse{State: bannerTestState(t, r, prior)}
	r.Update(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error: %v", resp.Diagnostics)
	}
	assertBannerStateKeepsPlan(t, bannerTestStateModel(t, resp.State), plan)
}

func TestAdminBannerRead_handler(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v2/admin/banners", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		fmt.Fprintf(w, `{"data": [%s]}`, strings.TrimSuffix(strings.TrimPrefix(utcBannerResponse, `{"data": `), "}"))
	})
	r := newBannerHandlerTestResource(t, "2.1.0", mux)

	t.Run("equivalent instant keeps the configured offset", func(t *testing.T) {
		prior := offsetScheduledPlan()
		prior.ID = types.StringValue("ban-1")
		prior.PublishedAt = testRFC3339("2026-01-01T00:00:00Z")

		resp := &fwresource.ReadResponse{State: bannerTestState(t, r, prior)}
		r.Read(context.Background(), fwresource.ReadRequest{State: bannerTestState(t, r, prior)}, resp)

		if resp.Diagnostics.HasError() {
			t.Fatalf("unexpected error: %v", resp.Diagnostics)
		}
		got := bannerTestStateModel(t, resp.State)
		if !got.ScheduledPublishAt.Equal(prior.ScheduledPublishAt) || !got.ScheduledExpireAt.Equal(prior.ScheduledExpireAt) {
			t.Errorf("expected configured offsets to be kept, got publish=%q expire=%q",
				got.ScheduledPublishAt.ValueString(), got.ScheduledExpireAt.ValueString())
		}
		// body is refreshed from the server so sanitization surfaces as a diff.
		if got.Body.ValueString() != "ok" {
			t.Errorf("body: got %q, want server value %q", got.Body.ValueString(), "ok")
		}
	})

	t.Run("different instant takes the server value", func(t *testing.T) {
		prior := offsetScheduledPlan()
		prior.ID = types.StringValue("ban-1")
		prior.ScheduledPublishAt = testRFC3339("2099-01-01T06:00:00-05:00")
		prior.PublishedAt = testRFC3339("2026-01-01T00:00:00Z")

		resp := &fwresource.ReadResponse{State: bannerTestState(t, r, prior)}
		r.Read(context.Background(), fwresource.ReadRequest{State: bannerTestState(t, r, prior)}, resp)

		got := bannerTestStateModel(t, resp.State)
		if got.ScheduledPublishAt.ValueString() != "2099-01-01T10:00:00Z" {
			t.Errorf("scheduled_publish_at: got %q, want server value", got.ScheduledPublishAt.ValueString())
		}
	})
}

// TestAdminBannerDestroy_afterDowngrade covers the lifecycle `terraform destroy`
// performs against a TFE that was downgraded below the minimum version: refresh
// (Read) must not fail, and Delete must still reach the API, tolerating a 404.
func TestAdminBannerDestroy_afterDowngrade(t *testing.T) {
	listed, deleted := false, false
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v2/admin/banners", func(w http.ResponseWriter, req *http.Request) { listed = true })
	mux.HandleFunc("/api/v2/admin/banners/ban-1", func(w http.ResponseWriter, req *http.Request) {
		deleted = req.Method == http.MethodDelete
		w.WriteHeader(http.StatusNotFound)
	})
	r := newBannerHandlerTestResource(t, "2.0.0", mux)

	prior := offsetScheduledPlan()
	prior.ID = types.StringValue("ban-1")
	prior.PublishedAt = testRFC3339("2026-01-01T00:00:00Z")

	readResp := &fwresource.ReadResponse{State: bannerTestState(t, r, prior)}
	r.Read(context.Background(), fwresource.ReadRequest{State: bannerTestState(t, r, prior)}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("refresh must not fail on an unsupported version: %v", readResp.Diagnostics)
	}
	if len(readResp.Diagnostics.Warnings()) == 0 {
		t.Error("expected a warning explaining that state was not refreshed")
	}
	if listed {
		t.Error("Read must not call the API on an unsupported version")
	}
	if readResp.State.Raw.IsNull() {
		t.Fatal("Read must preserve state so destroy can proceed")
	}
	if !bannerTestStateModel(t, readResp.State).ID.Equal(prior.ID) {
		t.Error("Read must leave the existing state unchanged")
	}

	delResp := &fwresource.DeleteResponse{State: readResp.State}
	r.Delete(context.Background(), fwresource.DeleteRequest{State: readResp.State}, delResp)

	if delResp.Diagnostics.HasError() {
		t.Fatalf("Delete must tolerate 404 and skip the version guard: %v", delResp.Diagnostics)
	}
	if !deleted {
		t.Error("expected DELETE /api/v2/admin/banners/ban-1 to be called")
	}
}

func TestAdminBannerImportState_handler(t *testing.T) {
	listed := false
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v2/admin/banners", func(w http.ResponseWriter, req *http.Request) {
		listed = true
		w.Header().Set("Content-Type", "application/vnd.api+json")
		fmt.Fprintf(w, `{"data": [%s]}`, strings.TrimSuffix(strings.TrimPrefix(utcBannerResponse, `{"data": `), "}"))
	})

	t.Run("unsupported version fails without calling the API", func(t *testing.T) {
		listed = false
		r := newBannerHandlerTestResource(t, "2.0.0", mux)

		resp := &fwresource.ImportStateResponse{State: bannerTestState(t, r, modelTFEAdminBanner{})}
		r.ImportState(context.Background(), fwresource.ImportStateRequest{ID: "ban-1"}, resp)

		if !resp.Diagnostics.HasError() {
			t.Fatal("expected ImportState to fail on an unsupported TFE version")
		}
		if listed {
			t.Error("ImportState must not call the API on an unsupported TFE version")
		}
	})

	t.Run("supported version imports the banner", func(t *testing.T) {
		listed = false
		r := newBannerHandlerTestResource(t, "2.1.0", mux)

		resp := &fwresource.ImportStateResponse{State: bannerTestState(t, r, modelTFEAdminBanner{})}
		r.ImportState(context.Background(), fwresource.ImportStateRequest{ID: "ban-1"}, resp)

		if resp.Diagnostics.HasError() {
			t.Fatalf("unexpected error: %v", resp.Diagnostics)
		}
		if got := bannerTestStateModel(t, resp.State).ID.ValueString(); got != "ban-1" {
			t.Errorf("id: got %q, want ban-1", got)
		}
	})
}
