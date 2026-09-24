// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const testAdminGeneralResourceName = "tfe_admin_general_settings.foobar"

// FLAKE ALERT: General settings are a singleton resource shared by the entire
// TFE instance, and any test touching them is at high risk to flake.
// In order for these tests to be safe, the following requirements MUST be met:
//  1. All test cases for this resource must run within a SINGLE test func, using
//     t.Run to separate the individual test cases.
//  2. The inner sub-tests must not call t.Parallel.
//
// If these tests are split into multiple test funcs and they get allocated to
// different test runner partitions in CI, then they will inevitably flake, as
// tests running concurrently in different containers will be competing to set
// the same shared global state in the TFE instance.

func TestAccTFEAdminGeneralSettings_omnibus(t *testing.T) {
	skipIfCloud(t)
	if os.Getenv("TF_ACC") == "" {
		t.Skip("acceptance tests require TF_ACC")
	}

	// Capture the value that exists before the test manages the settings so the
	// destroy check can assert it is restored. The shared test client is only
	// configured once a step runs, so build one directly from the environment.
	preTestClient, err := getClientV2UsingEnv()
	if err != nil {
		t.Fatalf("failed to build pre-test client: %s", err)
	}
	original, err := readAdminGeneralSettings(ctx, preTestClient)
	if err != nil {
		t.Fatalf("failed to read pre-test General Settings: %s", err)
	}

	// Always put the instance back to its pre-test values, even if a step
	// fails and the framework's post-test destroy never runs. Without this a
	// failed run leaves modified settings on the shared instance.
	t.Cleanup(func() {
		if _, err := patchAdminGeneralSettings(ctx, preTestClient, ptr(original.EnablePublicRegistration)); err != nil {
			t.Errorf("failed to restore pre-test General Settings: %s", err)
		}
	})

	t.Run("create with attribute omitted preserves existing server value", func(t *testing.T) {
		// Explicitly ensure the server value is true before creating the resource with the attribute omitted.
		if _, err := patchAdminGeneralSettings(ctx, preTestClient, ptr(true)); err != nil {
			t.Fatalf("failed to set pre-test general settings: %s", err)
		}
		t.Cleanup(func() {
			if _, err := patchAdminGeneralSettings(ctx, preTestClient, ptr(original.EnablePublicRegistration)); err != nil {
				t.Fatalf("failed to restore pre-test general settings: %s", err)
			}
		})

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			CheckDestroy:             testAccTFEAdminGeneralSettingsRestored(&adminGeneralSettings{EnablePublicRegistration: true}),
			Steps: []resource.TestStep{
				{
					Config: testAccTFEAdminGeneralSettingsOmitted(),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet(testAdminGeneralResourceName, "id"),
						resource.TestCheckResourceAttr(testAdminGeneralResourceName, "enable_public_registration", "true"),
					),
				},
			},
		})
	})

	t.Run("create, read, update, and restore general settings", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			CheckDestroy:             testAccTFEAdminGeneralSettingsRestored(original),
			Steps: []resource.TestStep{
				// Create with public registration enabled.
				{
					Config: testAccTFEAdminGeneralSettings(true),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet(testAdminGeneralResourceName, "id"),
						resource.TestCheckResourceAttr(testAdminGeneralResourceName, "enable_public_registration", "true"),
					),
				},
				// Re-applying the same config must be a no-op.
				{
					Config:   testAccTFEAdminGeneralSettings(true),
					PlanOnly: true,
				},
				// Update: disable public registration.
				{
					Config: testAccTFEAdminGeneralSettings(false),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(testAdminGeneralResourceName, "enable_public_registration", "false"),
					),
				},
				// When the attribute is omitted, the current server-side value is
				// preserved rather than being reset to a default. Because the prior
				// step disabled registration, the value must remain false with no
				// planned change.
				{
					Config:   testAccTFEAdminGeneralSettingsOmitted(),
					PlanOnly: true,
				},
				// Import.
				{
					ResourceName:      testAdminGeneralResourceName,
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})
}

func testAccTFEAdminGeneralSettings(enablePublicRegistration bool) string {
	return fmt.Sprintf(`
resource "tfe_admin_general_settings" "foobar" {
  enable_public_registration = %t
}`, enablePublicRegistration)
}

// testAccTFEAdminGeneralSettingsOmitted omits enable_public_registration so the
// current server-side value is preserved (Optional + Computed with no default).
func testAccTFEAdminGeneralSettingsOmitted() string {
	return `
resource "tfe_admin_general_settings" "foobar" {
}`
}

func testAccTFEAdminGeneralSettingsRestored(original *adminGeneralSettings) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		settings, err := readAdminGeneralSettings(ctx, testAccConfiguredClient.ClientV2)
		if err != nil {
			return fmt.Errorf("failed to read General Settings: %w", err)
		}

		// General settings cannot be deleted, only restored. Confirm the value
		// was returned to whatever existed before the test managed it.
		if settings.EnablePublicRegistration != original.EnablePublicRegistration {
			return fmt.Errorf("enable_public_registration not restored: got %v, want %v",
				settings.EnablePublicRegistration, original.EnablePublicRegistration)
		}

		return nil
	}
}

// fakePrivateState is an in-memory stand-in for the framework's private state,
// used to exercise the pre-management value round-trip without a full provider
// lifecycle.
type fakePrivateState struct {
	store map[string][]byte
}

func newFakePrivateState() *fakePrivateState {
	return &fakePrivateState{store: map[string][]byte{}}
}

func (f *fakePrivateState) SetKey(_ context.Context, key string, value []byte) diag.Diagnostics {
	f.store[key] = value
	return nil
}

func (f *fakePrivateState) GetKey(_ context.Context, key string) ([]byte, diag.Diagnostics) {
	return f.store[key], nil
}

// TestOriginalPublicRegistration_RoundTrip verifies that a recorded value is
// read back exactly, for both true and false, so delete restores what create
// observed.
func TestOriginalPublicRegistration_RoundTrip(t *testing.T) {
	ctx := context.Background()

	for _, original := range []bool{true, false} {
		private := newFakePrivateState()

		if diags := setOriginalPublicRegistration(ctx, private, original); diags.HasError() {
			t.Fatalf("setOriginalPublicRegistration returned diagnostics: %s", diags)
		}

		got, ok, diags := getOriginalPublicRegistration(ctx, private)
		if diags.HasError() {
			t.Fatalf("getOriginalPublicRegistration returned diagnostics: %s", diags)
		}
		if !ok {
			t.Fatalf("expected a recorded value for original=%v", original)
		}
		if got != original {
			t.Fatalf("expected restored value %v, got %v", original, got)
		}
	}
}

// TestOriginalPublicRegistration_NotRecorded verifies that when no value was
// ever recorded (for example an imported resource), the reader reports absence
// so delete leaves the server-side setting untouched.
func TestOriginalPublicRegistration_NotRecorded(t *testing.T) {
	ctx := context.Background()
	private := newFakePrivateState()

	_, ok, diags := getOriginalPublicRegistration(ctx, private)
	if diags.HasError() {
		t.Fatalf("getOriginalPublicRegistration returned diagnostics: %s", diags)
	}
	if ok {
		t.Fatal("expected no recorded value to be reported when nothing was stored")
	}
}
