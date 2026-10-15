// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const testAdminCustomizationResourceName = "tfe_admin_customization_settings.foobar"

// FLAKE ALERT: Customization settings are a singleton resource shared by the
// entire TFE instance, and any test touching them is at high risk to flake.
// In order for these tests to be safe, the following requirements MUST be met:
//  1. All test cases for this resource must run within a SINGLE test func, using
//     t.Run to separate the individual test cases.
//  2. The inner sub-tests must not call t.Parallel.
//
// If these tests are split into multiple test funcs and they get allocated to
// different test runner partitions in CI, then they will inevitably flake, as
// tests running concurrently in different containers will be competing to set
// the same shared global state in the TFE instance.

func TestAccTFEAdminCustomizationSettings_omnibus(t *testing.T) {
	skipIfCloud(t)
	if os.Getenv("TF_ACC") == "" {
		t.Skip("acceptance tests require TF_ACC")
	}

	// Capture the values that exist before the test manages the settings so the
	// destroy check can assert they are restored (not wiped to empty). The
	// shared test client is only configured once a step runs, so build one
	// directly from the environment here.
	preTestClient, err := getClientV2UsingEnv()
	if err != nil {
		t.Fatalf("failed to build pre-test client: %s", err)
	}
	original, err := readAdminCustomizationSettings(ctx, preTestClient)
	if err != nil {
		t.Fatalf("failed to read pre-test Customization Settings: %s", err)
	}

	// If a previous run aborted before destroy, the instance still holds this
	// test's fixture values. Capturing those as "original" would make every
	// later run restore fixtures instead of the real settings, so stop here and
	// ask for a manual reset.
	if field, value, polluted := customizationHasTestFixtures(original); polluted {
		t.Fatalf("Terraform Enterprise customization settings still contain test fixture data from an earlier "+
			"aborted run (%s = %q). Reset them in the Admin UI (or via the API) to their real values, then rerun.", field, value)
	}

	// Always put the instance back to its pre-test values, even if a step
	// fails and the framework's post-test destroy never runs. Without this a
	// failed run leaves fixture values on the shared instance.
	t.Cleanup(func() {
		if _, err := patchAdminCustomizationSettings(ctx, preTestClient, adminCustomizationSettingsUpdateFrom(original)); err != nil {
			t.Errorf("failed to restore pre-test Customization Settings: %s", err)
		}
	})

	t.Run("create, read, update, and restore customization settings", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			CheckDestroy:             testAccTFEAdminCustomizationSettingsRestored(original),
			Steps: []resource.TestStep{
				// Create with the full set of Phase 1 attributes.
				{
					Config: testAccTFEAdminCustomizationSettings_full(),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet(testAdminCustomizationResourceName, "id"),
						resource.TestCheckResourceAttr(testAdminCustomizationResourceName, "support_email", "support@example.com"),
						// support_url_address is a hostname (no scheme); when unset it is computed from TFE.
						resource.TestCheckResourceAttrSet(testAdminCustomizationResourceName, "support_url_address"),
						resource.TestCheckResourceAttr(testAdminCustomizationResourceName, "internal_support_url", "https://internal.example.com/support"),
						resource.TestCheckResourceAttr(testAdminCustomizationResourceName, "internal_documentation_url", "https://internal.example.com/docs"),
						resource.TestCheckResourceAttr(testAdminCustomizationResourceName, "internal_tutorials_url", "https://internal.example.com/tutorials"),
						resource.TestCheckResourceAttr(testAdminCustomizationResourceName, "login_help", "<p>login help</p>"),
						resource.TestCheckResourceAttr(testAdminCustomizationResourceName, "footer", "<p>footer</p>"),
						resource.TestCheckResourceAttr(testAdminCustomizationResourceName, "error", "<p>error</p>"),
						resource.TestCheckResourceAttr(testAdminCustomizationResourceName, "new_user", "<p>welcome</p>"),
						resource.TestCheckResourceAttr(testAdminCustomizationResourceName, "avatar_sources.%", "1"),
						resource.TestCheckResourceAttr(testAdminCustomizationResourceName, "avatar_sources.gravatar", "true"),
					),
				},
				// Re-applying the same config must be a no-op (guards against
				// perpetual diffs on the Optional+Computed attributes and the
				// avatar_sources map projection).
				{
					Config:   testAccTFEAdminCustomizationSettings_full(),
					PlanOnly: true,
				},
				// Update several fields, including clearing an internal URL override
				// and toggling the avatar source to disabled.
				{
					Config: testAccTFEAdminCustomizationSettings_updated(),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(testAdminCustomizationResourceName, "support_email", "help@example.com"),
						resource.TestCheckResourceAttr(testAdminCustomizationResourceName, "internal_support_url", "https://internal.example.com/new-support"),
						// Cleared override is restored to an empty string.
						resource.TestCheckResourceAttr(testAdminCustomizationResourceName, "internal_documentation_url", ""),
						resource.TestCheckResourceAttr(testAdminCustomizationResourceName, "avatar_sources.gravatar", "false"),
					),
				},
				// A mailto: support URL is accepted.
				{
					Config: testAccTFEAdminCustomizationSettings_mailtoSupport(),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(testAdminCustomizationResourceName, "internal_support_url", "mailto:support@example.com"),
					),
				},
				// Import.
				{
					ResourceName:      testAdminCustomizationResourceName,
					ImportState:       true,
					ImportStateVerify: true,
				},
				// Invalid support URL scheme is rejected at plan time.
				{
					Config:      testAccTFEAdminCustomizationSettings_invalidScheme(),
					ExpectError: regexp.MustCompile(`Invalid URL scheme`),
				},
				// A mailto: scheme is not valid for the documentation URL.
				{
					Config:      testAccTFEAdminCustomizationSettings_mailtoDocumentation(),
					ExpectError: regexp.MustCompile(`Invalid URL scheme`),
				},
				// An unknown avatar source key is rejected at plan time. Validation
				// is delegated to Terraform Enterprise (via go-tfe's generated,
				// nightly-refreshed API contract) rather than a hardcoded list.
				{
					Config: testAccTFEAdminCustomizationSettings_invalidAvatarSource(),
					// Terraform wraps diagnostics, so allow any whitespace between words.
					ExpectError: regexp.MustCompile(`is\s+not\s+a\s+recognized\s+avatar\s+source`),
				},
				// Finish on a valid config so the post-test destroy has a
				// plannable configuration. A plan-time validation error in the
				// preceding steps leaves an unplannable config in place, which
				// would otherwise fail the framework's automatic destroy.
				{
					Config: testAccTFEAdminCustomizationSettings_mailtoSupport(),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(testAdminCustomizationResourceName, "internal_support_url", "mailto:support@example.com"),
					),
				},
			},
		})
	})
}

func testAccTFEAdminCustomizationSettings_full() string {
	return `
resource "tfe_admin_customization_settings" "foobar" {
  support_email              = "support@example.com"
  internal_support_url       = "https://internal.example.com/support"
  internal_documentation_url = "https://internal.example.com/docs"
  internal_tutorials_url     = "https://internal.example.com/tutorials"
  login_help                 = "<p>login help</p>"
  footer                     = "<p>footer</p>"
  error                      = "<p>error</p>"
  new_user                   = "<p>welcome</p>"

  avatar_sources = {
    gravatar = true
  }
}`
}

func testAccTFEAdminCustomizationSettings_updated() string {
	return `
resource "tfe_admin_customization_settings" "foobar" {
  support_email              = "help@example.com"
  internal_support_url       = "https://internal.example.com/new-support"
  internal_documentation_url = ""
  internal_tutorials_url     = "https://internal.example.com/tutorials"
  login_help                 = "<p>login help</p>"
  footer                     = "<p>footer</p>"
  error                      = "<p>error</p>"
  new_user                   = "<p>welcome</p>"

  avatar_sources = {
    gravatar = false
  }
}`
}

func testAccTFEAdminCustomizationSettings_mailtoSupport() string {
	return `
resource "tfe_admin_customization_settings" "foobar" {
  support_email        = "help@example.com"
  internal_support_url = "mailto:support@example.com"

  avatar_sources = {
    gravatar = false
  }
}`
}

func testAccTFEAdminCustomizationSettings_invalidScheme() string {
	return `
resource "tfe_admin_customization_settings" "foobar" {
  internal_support_url = "ftp://internal.example.com/support"
}`
}

func testAccTFEAdminCustomizationSettings_mailtoDocumentation() string {
	return `
resource "tfe_admin_customization_settings" "foobar" {
  internal_documentation_url = "mailto:docs@example.com"
}`
}

func testAccTFEAdminCustomizationSettings_invalidAvatarSource() string {
	return `
resource "tfe_admin_customization_settings" "foobar" {
  avatar_sources = {
    not-a-real-source = true
  }
}`
}

// customizationHasTestFixtures reports whether the settings contain any value
// written by this file's test configurations.
func customizationHasTestFixtures(s *adminCustomizationSettings) (string, string, bool) {
	fields := []struct{ name, value string }{
		{"support_email", s.SupportEmail},
		{"internal_support_url", s.InternalSupportURL},
		{"internal_documentation_url", s.InternalDocumentationURL},
		{"internal_tutorials_url", s.InternalTutorialsURL},
		{"login_help", s.LoginHelp},
		{"footer", s.Footer},
		{"error", s.Error},
		{"new_user", s.NewUser},
	}
	fixtures := map[string]bool{
		"support@example.com":                      true,
		"help@example.com":                         true,
		"https://internal.example.com/support":     true,
		"https://internal.example.com/new-support": true,
		"https://internal.example.com/docs":        true,
		"https://internal.example.com/tutorials":   true,
		"mailto:support@example.com":               true,
		"<p>login help</p>":                        true,
		"<p>footer</p>":                            true,
		"<p>error</p>":                             true,
		"<p>welcome</p>":                           true,
	}
	for _, f := range fields {
		if fixtures[f.value] {
			return f.name, f.value, true
		}
	}
	return "", "", false
}

func testAccTFEAdminCustomizationSettingsRestored(original *adminCustomizationSettings) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		settings, err := readAdminCustomizationSettings(ctx, testAccConfiguredClient.ClientV2)
		if err != nil {
			return fmt.Errorf("failed to read Customization Settings: %w", err)
		}

		// Customization settings cannot be deleted, only restored. Confirm the
		// text overrides were returned to the values that existed before the
		// test managed them, rather than wiped to empty.
		switch {
		case settings.SupportEmail != original.SupportEmail:
			return fmt.Errorf("support_email not restored: got %q, want %q", settings.SupportEmail, original.SupportEmail)
		case settings.SupportURLAddress != original.SupportURLAddress:
			return fmt.Errorf("support_url_address not restored: got %q, want %q", settings.SupportURLAddress, original.SupportURLAddress)
		case settings.InternalSupportURL != original.InternalSupportURL:
			return fmt.Errorf("internal_support_url not restored: got %q, want %q", settings.InternalSupportURL, original.InternalSupportURL)
		case settings.InternalDocumentationURL != original.InternalDocumentationURL:
			return fmt.Errorf("internal_documentation_url not restored: got %q, want %q", settings.InternalDocumentationURL, original.InternalDocumentationURL)
		case settings.InternalTutorialsURL != original.InternalTutorialsURL:
			return fmt.Errorf("internal_tutorials_url not restored: got %q, want %q", settings.InternalTutorialsURL, original.InternalTutorialsURL)
		case settings.LoginHelp != original.LoginHelp:
			return fmt.Errorf("login_help not restored: got %q, want %q", settings.LoginHelp, original.LoginHelp)
		case settings.Footer != original.Footer:
			return fmt.Errorf("footer not restored: got %q, want %q", settings.Footer, original.Footer)
		case settings.Error != original.Error:
			return fmt.Errorf("error not restored: got %q, want %q", settings.Error, original.Error)
		case settings.NewUser != original.NewUser:
			return fmt.Errorf("new_user not restored: got %q, want %q", settings.NewUser, original.NewUser)
		}

		for _, want := range original.AvatarSources {
			found := false
			for _, got := range settings.AvatarSources {
				if got.Source == want.Source {
					found = true
					if got.Enabled != want.Enabled {
						return fmt.Errorf("avatar source %q enabled status not restored: got %v, want %v", want.Source, got.Enabled, want.Enabled)
					}
					break
				}
			}
			if !found {
				return fmt.Errorf("avatar source %q missing after restore", want.Source)
			}
		}

		return nil
	}
}

// TestAvatarSourcesToMap_ProjectsManagedKeysOnly is the key regression test for
// the forward-compatibility requirement: when Terraform Enterprise extends its
// avatar-source catalog with sources the practitioner does not manage, those
// extra sources must NOT be written into state, so no perpetual drift or
// "inconsistent result after apply" error can occur.
func TestAvatarSourcesToMap_ProjectsManagedKeysOnly(t *testing.T) {
	ctx := context.Background()

	// The API returns a catalog larger than what the practitioner manages.
	apiSources := []adminAvatarSource{
		{Source: "gravatar", Enabled: true},
		{Source: "saml", Enabled: true}, // hypothetical future source
	}

	// The practitioner only manages "gravatar".
	prior, diags := types.MapValueFrom(ctx, types.BoolType, map[string]bool{"gravatar": true})
	if diags.HasError() {
		t.Fatalf("failed to build prior map: %s", diags)
	}

	result, diags := avatarSourcesToMap(ctx, apiSources, prior)
	if diags.HasError() {
		t.Fatalf("avatarSourcesToMap returned diagnostics: %s", diags)
	}

	elements := result.Elements()
	if len(elements) != 1 {
		t.Fatalf("expected only the managed key to be projected, got %d keys: %v", len(elements), elements)
	}
	if _, ok := elements["gravatar"]; !ok {
		t.Fatalf("expected managed key 'gravatar' to be present, got: %v", elements)
	}
	if _, ok := elements["saml"]; ok {
		t.Fatalf("unmanaged key 'saml' must not be projected into state")
	}
}

// TestAvatarSourcesToMap_NilPriorProjectsFullCatalog verifies that on import
// (no prior config) the full catalog reported by the API is projected.
func TestAvatarSourcesToMap_NilPriorProjectsFullCatalog(t *testing.T) {
	ctx := context.Background()

	apiSources := []adminAvatarSource{
		{Source: "gravatar", Enabled: true},
	}

	result, diags := avatarSourcesToMap(ctx, apiSources, types.MapNull(types.BoolType))
	if diags.HasError() {
		t.Fatalf("avatarSourcesToMap returned diagnostics: %s", diags)
	}

	elements := result.Elements()
	if len(elements) != 1 {
		t.Fatalf("expected full catalog to be projected, got %d keys: %v", len(elements), elements)
	}
	if _, ok := elements["gravatar"]; !ok {
		t.Fatalf("expected 'gravatar' to be present, got: %v", elements)
	}
}

// TestAvatarSourcesToMap_ManagedKeyMissingFromAPIDisabled verifies that a source
// the practitioner manages but which the API does not report is treated as
// disabled rather than dropped, keeping the config and state consistent.
func TestAvatarSourcesToMap_ManagedKeyMissingFromAPIDisabled(t *testing.T) {
	ctx := context.Background()

	apiSources := []adminAvatarSource{} // API reports nothing enabled

	prior, diags := types.MapValueFrom(ctx, types.BoolType, map[string]bool{"gravatar": true})
	if diags.HasError() {
		t.Fatalf("failed to build prior map: %s", diags)
	}

	result, diags := avatarSourcesToMap(ctx, apiSources, prior)
	if diags.HasError() {
		t.Fatalf("avatarSourcesToMap returned diagnostics: %s", diags)
	}

	var got map[string]bool
	result.ElementsAs(ctx, &got, false)
	if enabled, ok := got["gravatar"]; !ok || enabled {
		t.Fatalf("expected managed key 'gravatar' to be present and disabled, got: %v", got)
	}
}

// TestAvatarSourcesFromMap_StableOrder verifies keys are sorted so the request
// payload is deterministic.
func TestAvatarSourcesFromMap_StableOrder(t *testing.T) {
	ctx := context.Background()

	in, diags := types.MapValueFrom(ctx, types.BoolType, map[string]bool{
		"gravatar": true,
		"saml":     false,
	})
	if diags.HasError() {
		t.Fatalf("failed to build map: %s", diags)
	}

	sources, diags := avatarSourcesFromMap(ctx, in)
	if diags.HasError() {
		t.Fatalf("avatarSourcesFromMap returned diagnostics: %s", diags)
	}

	if len(sources) != 2 {
		t.Fatalf("expected 2 sources, got %d", len(sources))
	}
	if sources[0].Source != "gravatar" || sources[1].Source != "saml" {
		t.Fatalf("expected sorted order [gravatar, saml], got [%s, %s]", sources[0].Source, sources[1].Source)
	}
}

// TestMergeAvatarSources_PreservesUnmanaged verifies that sources the
// practitioner does not manage keep their server-side state, because Terraform
// Enterprise disables any source omitted from a PATCH.
func TestMergeAvatarSources_PreservesUnmanaged(t *testing.T) {
	current := []adminAvatarSource{
		{Source: "gravatar", Enabled: true},
		{Source: "future", Enabled: true},
	}
	configured := []adminAvatarSource{{Source: "gravatar", Enabled: false}}

	got := mergeAvatarSources(current, configured)
	want := []adminAvatarSource{
		{Source: "gravatar", Enabled: false},
		{Source: "future", Enabled: true},
	}
	if len(got) != len(want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("expected %v, got %v", want, got)
		}
	}
}

// TestMergeAvatarSources_AppendsUnreported verifies configured sources that the
// server did not report are still sent.
func TestMergeAvatarSources_AppendsUnreported(t *testing.T) {
	got := mergeAvatarSources(nil, []adminAvatarSource{{Source: "gravatar", Enabled: true}})
	if len(got) != 1 || got[0].Source != "gravatar" || !got[0].Enabled {
		t.Fatalf("expected configured source to be sent, got %v", got)
	}
}

func TestSupportURLAddressRegexp(t *testing.T) {
	for _, v := range []string{"support.example.com", "support.example.com/help", "host:8443"} {
		if !noURLSchemeRegexp.MatchString(v) {
			t.Errorf("expected %q to be accepted", v)
		}
	}
	for _, v := range []string{"https://support.example.com", "mailto:a@b.com", "has space.com"} {
		if noURLSchemeRegexp.MatchString(v) {
			t.Errorf("expected %q to be rejected", v)
		}
	}
}

func TestSupportEmailRegexp(t *testing.T) {
	for _, v := range []string{
		"support@example.com",
		"support@localhost",
		"admin@internal-host",
		"first.last+tag@sub.domain.co",
	} {
		if !supportEmailRegexp.MatchString(v) {
			t.Errorf("expected %q to be accepted", v)
		}
	}
	for _, v := range []string{
		"plainaddress",
		"@missingusername.com",
		"username@",
		"has space@example.com",
	} {
		if supportEmailRegexp.MatchString(v) {
			t.Errorf("expected %q to be rejected", v)
		}
	}
}

// TestOriginalCustomization_RoundTrip verifies that the pre-management snapshot
// captured on create is read back exactly, so delete restores what create
// observed rather than wiping the settings to empty.
func TestOriginalCustomization_RoundTrip(t *testing.T) {
	ctx := context.Background()
	private := newFakePrivateState()

	source := &adminCustomizationSettings{
		SupportEmail:             "pre@example.com",
		SupportURLAddress:        "support.example.com",
		LoginHelp:                "<p>pre login</p>",
		Footer:                   "<p>pre footer</p>",
		Error:                    "<p>pre error</p>",
		NewUser:                  "<p>pre new user</p>",
		InternalSupportURL:       "https://pre.example.com/support",
		InternalDocumentationURL: "https://pre.example.com/docs",
		InternalTutorialsURL:     "https://pre.example.com/tutorials",
		AvatarSources:            []adminAvatarSource{{Source: "gravatar", Enabled: true}},
	}

	avatarMap, _ := types.MapValueFrom(ctx, types.BoolType, map[string]bool{"gravatar": true})
	config := modelTFEAdminCustomizationSettings{
		SupportEmail:             types.StringValue("configured@example.com"),
		SupportURLAddress:        types.StringValue("support.example.com"),
		LoginHelp:                types.StringValue("<p>login</p>"),
		Footer:                   types.StringValue("<p>footer</p>"),
		Error:                    types.StringValue("<p>error</p>"),
		NewUser:                  types.StringValue("<p>new</p>"),
		InternalSupportURL:       types.StringValue("https://support.example.com"),
		InternalDocumentationURL: types.StringValue("https://docs.example.com"),
		InternalTutorialsURL:     types.StringValue("https://tutorials.example.com"),
		AvatarSources:            avatarMap,
	}

	if diags := setOriginalCustomization(ctx, private, source, config); diags.HasError() {
		t.Fatalf("setOriginalCustomization returned diagnostics: %s", diags)
	}

	got, ok, diags := getOriginalCustomization(ctx, private)
	if diags.HasError() {
		t.Fatalf("getOriginalCustomization returned diagnostics: %s", diags)
	}
	if !ok {
		t.Fatal("expected a recorded snapshot")
	}

	if *got.SupportEmail != source.SupportEmail ||
		*got.SupportURLAddress != source.SupportURLAddress ||
		*got.LoginHelp != source.LoginHelp ||
		*got.Footer != source.Footer ||
		*got.Error != source.Error ||
		*got.NewUser != source.NewUser ||
		*got.InternalSupportURL != source.InternalSupportURL ||
		*got.InternalDocumentationURL != source.InternalDocumentationURL ||
		*got.InternalTutorialsURL != source.InternalTutorialsURL {
		t.Fatalf("restored text overrides do not match original: %+v", got)
	}
	if len(got.AvatarSources) != 1 || got.AvatarSources[0].Source != "gravatar" || !got.AvatarSources[0].Enabled {
		t.Fatalf("restored avatar sources do not match original: %+v", got.AvatarSources)
	}
}

// TestOriginalCustomization_PartialConfig verifies that unconfigured fields
// are not recorded in the snapshot, so they will not be restored on delete.
func TestOriginalCustomization_PartialConfig(t *testing.T) {
	ctx := context.Background()
	private := newFakePrivateState()

	source := &adminCustomizationSettings{
		SupportEmail:             "pre@example.com",
		Footer:                   "<p>pre footer</p>",
		InternalSupportURL:       "https://pre.example.com/support",
		InternalDocumentationURL: "https://pre.example.com/docs",
		AvatarSources: []adminAvatarSource{
			{Source: "gravatar", Enabled: true},
			{Source: "other", Enabled: false},
		},
	}

	// Configuration manages footer and an empty avatar_sources map.
	emptyMap, _ := types.MapValueFrom(ctx, types.BoolType, map[string]bool{})
	config := modelTFEAdminCustomizationSettings{
		Footer:        types.StringValue("<p>new footer</p>"),
		AvatarSources: emptyMap,
	}

	if diags := setOriginalCustomization(ctx, private, source, config); diags.HasError() {
		t.Fatalf("setOriginalCustomization returned diagnostics: %s", diags)
	}

	got, ok, diags := getOriginalCustomization(ctx, private)
	if diags.HasError() {
		t.Fatalf("getOriginalCustomization returned diagnostics: %s", diags)
	}
	if !ok {
		t.Fatal("expected a recorded snapshot")
	}

	if got.Footer == nil || *got.Footer != "<p>pre footer</p>" {
		t.Fatalf("expected recorded footer, got %+v", got.Footer)
	}
	if got.SupportEmail != nil {
		t.Fatalf("expected unconfigured SupportEmail to not be recorded, got %+v", got.SupportEmail)
	}
	if got.InternalSupportURL != nil {
		t.Fatalf("expected unconfigured InternalSupportURL to not be recorded, got %+v", got.InternalSupportURL)
	}
	if len(got.AvatarSources) != 0 {
		t.Fatalf("expected empty avatar_sources in snapshot when no keys configured, got %+v", got.AvatarSources)
	}
}

// TestOriginalCustomization_NotRecorded verifies that when no snapshot was ever
// recorded (for example an imported resource), the reader reports absence so
// delete leaves the server-side settings untouched.
func TestOriginalCustomization_NotRecorded(t *testing.T) {
	ctx := context.Background()
	private := newFakePrivateState()

	_, ok, diags := getOriginalCustomization(ctx, private)
	if diags.HasError() {
		t.Fatalf("getOriginalCustomization returned diagnostics: %s", diags)
	}
	if ok {
		t.Fatal("expected no recorded snapshot to be reported when nothing was stored")
	}
}

func TestAdminCustomizationSettings_checkMinTFEVersion(t *testing.T) {
	t.Run("HCP Terraform is rejected", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client, err := tfe.NewClient(&tfe.Config{Address: "https://app.terraform.io", Token: "test-token"})
		if err != nil {
			t.Fatalf("failed to create client: %v", err)
		}
		r := &resourceTFEAdminCustomizationSettings{
			config: ConfiguredClient{Client: client},
		}

		var diags diag.Diagnostics
		if r.checkMinTFEVersion(&diags) {
			t.Fatal("expected checkMinTFEVersion to return false for HCP Terraform")
		}
		if !diags.HasError() {
			t.Fatal("expected error diagnostic for HCP Terraform")
		}
		if !strings.Contains(diags.Errors()[0].Detail(), "cannot be used with HCP Terraform") {
			t.Fatalf("unexpected error message: %s", diags.Errors()[0].Detail())
		}
	})

	t.Run("older TFE version is rejected", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-TFE-Current-Version", "2.0.0")
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client, err := tfe.NewClient(&tfe.Config{Address: server.URL, Token: "test-token"})
		if err != nil {
			t.Fatalf("failed to create client: %v", err)
		}
		r := &resourceTFEAdminCustomizationSettings{
			config: ConfiguredClient{Client: client},
		}

		var diags diag.Diagnostics
		if r.checkMinTFEVersion(&diags) {
			t.Fatal("expected checkMinTFEVersion to return false for older TFE")
		}
		if !diags.HasError() {
			t.Fatal("expected error diagnostic for older TFE")
		}
		if !strings.Contains(diags.Errors()[0].Summary(), "Unsupported Terraform Enterprise version") {
			t.Fatalf("unexpected error summary: %s", diags.Errors()[0].Summary())
		}
	})

	t.Run("supported TFE version passes", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-TFE-Current-Version", "2.1.0")
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client, err := tfe.NewClient(&tfe.Config{Address: server.URL, Token: "test-token"})
		if err != nil {
			t.Fatalf("failed to create client: %v", err)
		}
		r := &resourceTFEAdminCustomizationSettings{
			config: ConfiguredClient{Client: client},
		}

		var diags diag.Diagnostics
		if !r.checkMinTFEVersion(&diags) {
			t.Fatalf("expected checkMinTFEVersion to return true, got errors: %s", diags)
		}
		if diags.HasError() {
			t.Fatalf("expected no errors, got: %s", diags)
		}
	})
}
