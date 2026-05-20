// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-tfe/internal/client"
)

var scimTestSAMLSetting = tfe.AdminSAMLSetting{
	IDPCert:        "testIDPCertBasic",
	SLOEndpointURL: "https://foobar.com/slo_endpoint_url",
	SSOEndpointURL: "https://foobar.com/sso_endpoint_url",
	ProviderType:   tfe.SAMLProviderTypeOkta,
}

// FLAKE ALERT: SCIM settings are a singleton resource shared by the entire TFE
// instance, and any test touching them is at high risk to flake.
// In order for these tests to be safe, the following requirements MUST be met:
//  1. All test cases for this resource must run within a SINGLE test func, using
//     t.Run to separate the individual test cases.
//  2. The inner sub-tests must not call t.Parallel.
//
// If these tests are split into multiple test funcs and they get allocated to
// different test runner partitions in CI, then they will inevitably flake, as
// tests running concurrently in different containers will be competing to set
// the same shared global state in the TFE instance.
//
// FLAKE ALERT (dual-singleton): Every SCIM test inlines a tfe_saml_settings
// block, so this suite also contends with resource_tfe_saml_settings_test.go
// for the SAML singleton. Both singletons must be treated as exclusive
// resources: do not run SCIM and SAML acceptance tests concurrently.

// TestAccTFESCIMSettings_omnibus test suite is skipped in the CI, and will only run in TFE Nightly workflow
// Should this test name ever change, you will also need to update the regex in ci.yml
func TestAccTFESCIMSettings_omnibus(t *testing.T) {
	skipIfCloud(t)

	t.Run("basic SCIM settings resource", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			CheckDestroy:             testAccTFESCIMSettingsDestroy,
			Steps: []resource.TestStep{
				// Enable SCIM with defaults.
				{
					Config: testAccTFESCIMSettings_enable(),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("tfe_scim_settings.enable_scim", "enabled", "true"),
						resource.TestCheckResourceAttr("tfe_scim_settings.enable_scim", "paused", "false"),
						resource.TestCheckResourceAttr("tfe_scim_settings.enable_scim", "site_admin_group_scim_id", ""),
						resource.TestCheckResourceAttr("tfe_scim_settings.enable_scim", "site_admin_group_display_name", ""),
					),
				},
				// Pause SCIM.
				{
					Config: testAccTFESCIMSettings_paused(),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("tfe_scim_settings.enable_scim", "enabled", "true"),
						resource.TestCheckResourceAttr("tfe_scim_settings.enable_scim", "paused", "true"),
						resource.TestCheckResourceAttr("tfe_scim_settings.enable_scim", "site_admin_group_scim_id", ""),
						resource.TestCheckResourceAttr("tfe_scim_settings.enable_scim", "site_admin_group_display_name", ""),
					),
				},
				// Omitting `paused` reverts to the default (false).
				{
					Config: testAccTFESCIMSettings_enable(),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("tfe_scim_settings.enable_scim", "enabled", "true"),
						resource.TestCheckResourceAttr("tfe_scim_settings.enable_scim", "paused", "false"),
					),
				},
				// Explicitly unpause.
				{
					Config: testAccTFESCIMSettings_unpaused(),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("tfe_scim_settings.enable_scim", "enabled", "true"),
						resource.TestCheckResourceAttr("tfe_scim_settings.enable_scim", "paused", "false"),
					),
				},
			},
		})
	})

	t.Run("SCIM settings site admin group", func(t *testing.T) {
		var siteAdminGroupID string
		var siteAdminGroupBID string

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			CheckDestroy:             testAccTFESCIMSettingsDestroy,
			Steps: []resource.TestStep{
				// Enable SCIM with no site admin group linked.
				{
					Config: testAccTFESCIMSettings_enable(),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("tfe_scim_settings.enable_scim", "enabled", "true"),
					),
				},
				// Create a SCIM group out-of-band and link it via TF_VAR.
				{
					PreConfig: func() {
						token, err := testAccConfiguredClient.Client.Admin.Settings.SCIM.Tokens.Create(
							context.Background(), "tf-acc-test-scim-token",
						)
						if err != nil {
							t.Fatalf("create SCIM token: %v", err)
						}

						siteAdminGroupID = createSCIMGroup(t, "tf-acc-site-admins", token.Token)
						t.Setenv("TF_VAR_site_admin_group_scim_id", siteAdminGroupID)
					},
					Config: testAccTFESCIMSettings_withSiteAdminGroup(),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("tfe_scim_settings.enable_scim", "enabled", "true"),
						resource.TestCheckResourceAttrPtr(
							"tfe_scim_settings.enable_scim",
							"site_admin_group_scim_id",
							&siteAdminGroupID,
						),
						resource.TestCheckResourceAttr(
							"tfe_scim_settings.enable_scim",
							"site_admin_group_display_name",
							"tf-acc-site-admins",
						),
					),
				},
				// Re-apply same config: should be a no-op (no perpetual diff).
				{
					Config:   testAccTFESCIMSettings_withSiteAdminGroup(),
					PlanOnly: true,
				},
				// Import round-trips the linked group through state.
				{
					ResourceName:      "tfe_scim_settings.enable_scim",
					ImportState:       true,
					ImportStateId:     "scim",
					ImportStateVerify: true,
				},
				// Clear the site admin group by setting it to "".
				{
					Config: testAccTFESCIMSettings_clearSiteAdminGroup(),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("tfe_scim_settings.enable_scim", "enabled", "true"),
						resource.TestCheckResourceAttr("tfe_scim_settings.enable_scim", "site_admin_group_scim_id", ""),
						resource.TestCheckResourceAttr("tfe_scim_settings.enable_scim", "site_admin_group_display_name", ""),
					),
				},
				// Re-link the same group (env var still set above).
				{
					Config: testAccTFESCIMSettings_withSiteAdminGroup(),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("tfe_scim_settings.enable_scim", "enabled", "true"),
						resource.TestCheckResourceAttrPtr(
							"tfe_scim_settings.enable_scim",
							"site_admin_group_scim_id",
							&siteAdminGroupID,
						),
					),
				},
				// Switch from group A to group B (non-null → non-null).
				{
					PreConfig: func() {
						token, err := testAccConfiguredClient.Client.Admin.Settings.SCIM.Tokens.Create(
							context.Background(), "tf-acc-test-scim-token-b",
						)
						if err != nil {
							t.Fatalf("create SCIM token for group B: %v", err)
						}
						siteAdminGroupBID = createSCIMGroup(t, "tf-acc-site-admins-b", token.Token)
						t.Setenv("TF_VAR_site_admin_group_b_scim_id", siteAdminGroupBID)
					},
					Config: testAccTFESCIMSettings_withSiteAdminGroupB(),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("tfe_scim_settings.enable_scim", "enabled", "true"),
						resource.TestCheckResourceAttrPtr(
							"tfe_scim_settings.enable_scim",
							"site_admin_group_scim_id",
							&siteAdminGroupBID,
						),
						resource.TestCheckResourceAttr(
							"tfe_scim_settings.enable_scim",
							"site_admin_group_display_name",
							"tf-acc-site-admins-b",
						),
					),
				},
				// Omitting site_admin_group_scim_id reverts to the default (""), unlinking the group.
				{
					Config: testAccTFESCIMSettings_enable(),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("tfe_scim_settings.enable_scim", "enabled", "true"),
						resource.TestCheckResourceAttr("tfe_scim_settings.enable_scim", "site_admin_group_scim_id", ""),
						resource.TestCheckResourceAttr("tfe_scim_settings.enable_scim", "site_admin_group_display_name", ""),
					),
				},
			},
		})
	})

	t.Run("SCIM settings import", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			CheckDestroy:             testAccTFESCIMSettingsDestroy,
			Steps: []resource.TestStep{
				// Enable SCIM.
				{
					Config: testAccTFESCIMSettings_enable(),
				},
				// Import by the fixed "scim" ID.
				{
					ResourceName:      "tfe_scim_settings.enable_scim",
					ImportState:       true,
					ImportStateId:     "scim",
					ImportStateVerify: true,
				},
			},
		})
	})

	t.Run("SCIM settings out-of-band drift", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			CheckDestroy:             testAccTFESCIMSettingsDestroy,
			Steps: []resource.TestStep{
				// Enable SCIM via Terraform.
				{
					Config: testAccTFESCIMSettings_enable(),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("tfe_scim_settings.enable_scim", "enabled", "true"),
					),
				},
				// Disable SCIM out-of-band (simulating an external change), then re-apply:
				// Read should detect the drift (resource absent) and Create should re-enable.
				{
					PreConfig: func() {
						if err := testAccConfiguredClient.Client.Admin.Settings.SCIM.Delete(ctx); err != nil {
							t.Fatalf("disable SCIM out-of-band: %v", err)
						}
					},
					Config: testAccTFESCIMSettings_enable(),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("tfe_scim_settings.enable_scim", "enabled", "true"),
					),
				},
			},
		})
	})
}

func testAccTFESCIMSettingsDestroy(_ *terraform.State) error {
	s, err := testAccConfiguredClient.Client.Admin.Settings.SCIM.Read(ctx)
	if err != nil {
		return fmt.Errorf("failed to read SCIM Settings: %w", err)
	}
	if s.Enabled {
		return errors.New("SCIM Settings are still enabled")
	}
	if s.Paused {
		return errors.New("SCIM Settings are still paused")
	}
	if s.SiteAdminGroupSCIMID != "" {
		return errors.New("SCIM Settings still have site admin group linked")
	}
	return nil
}

// Similar to testAccTFESAMLSettings_basic in resource_tfe_saml_settings_test.go,
// duplicated here to keep the SCIM suite self-contained.
func testAccTFESCIMSettings_enableSAMLWithProviderType(a tfe.AdminSAMLSetting) string {
	return fmt.Sprintf(`
resource "tfe_saml_settings" "enable_saml" {
	idp_cert               = "%s"
	slo_endpoint_url       = "%s"
	sso_endpoint_url       = "%s"
	provider_type          = "%s"
}
`, a.IDPCert, a.SLOEndpointURL, a.SSOEndpointURL, a.ProviderType)
}

func testAccTFESCIMSettings_enable() string {
	return fmt.Sprintf(`

%s

resource "tfe_scim_settings" "enable_scim" {
    depends_on = [tfe_saml_settings.enable_saml]
}
`, testAccTFESCIMSettings_enableSAMLWithProviderType(scimTestSAMLSetting))
}

func testAccTFESCIMSettings_paused() string {
	return fmt.Sprintf(`

%s

resource "tfe_scim_settings" "enable_scim" {
	paused     = true
    depends_on = [tfe_saml_settings.enable_saml]
}
`, testAccTFESCIMSettings_enableSAMLWithProviderType(scimTestSAMLSetting))
}

func testAccTFESCIMSettings_unpaused() string {
	return fmt.Sprintf(`

%s

resource "tfe_scim_settings" "enable_scim" {
	paused     = false
    depends_on = [tfe_saml_settings.enable_saml]
}
`, testAccTFESCIMSettings_enableSAMLWithProviderType(scimTestSAMLSetting))
}

func testAccTFESCIMSettings_withSiteAdminGroup() string {
	return fmt.Sprintf(`
%s

variable "site_admin_group_scim_id" {
    type = string
}
resource "tfe_scim_settings" "enable_scim" {
	site_admin_group_scim_id = var.site_admin_group_scim_id
	depends_on               = [tfe_saml_settings.enable_saml]
}
`, testAccTFESCIMSettings_enableSAMLWithProviderType(scimTestSAMLSetting))
}

func testAccTFESCIMSettings_withSiteAdminGroupB() string {
	return fmt.Sprintf(`
%s

variable "site_admin_group_b_scim_id" {
	type = string
}
resource "tfe_scim_settings" "enable_scim" {
	site_admin_group_scim_id = var.site_admin_group_b_scim_id
	depends_on               = [tfe_saml_settings.enable_saml]
}
`, testAccTFESCIMSettings_enableSAMLWithProviderType(scimTestSAMLSetting))
}

func testAccTFESCIMSettings_clearSiteAdminGroup() string {
	return fmt.Sprintf(`
%s

resource "tfe_scim_settings" "enable_scim" {
    site_admin_group_scim_id = ""
    depends_on               = [tfe_saml_settings.enable_saml]
}
`, testAccTFESCIMSettings_enableSAMLWithProviderType(scimTestSAMLSetting))
}

// createSCIMGroup creates a SCIM group via the SCIM v2 API and returns its ID.
func createSCIMGroup(t *testing.T, displayName, scimToken string) string {
	t.Helper()

	hostname := os.Getenv("TFE_HOSTNAME")
	if hostname == "" {
		hostname = client.DefaultHostname
	}

	body, err := json.Marshal(map[string]any{
		"displayName": displayName,
		"schemas":     []string{"urn:ietf:params:scim:schemas:core:2.0:Group"},
	})
	if err != nil {
		t.Fatalf("marshal SCIM group request body: %v", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost,
		fmt.Sprintf("https://%s/scim/v2/Groups", hostname), bytes.NewReader(body))
	if err != nil {
		t.Fatalf("build SCIM group request: %v", err)
	}
	req.Header.Set("Content-Type", "application/scim+json")
	req.Header.Set("Authorization", "Bearer "+scimToken)

	httpClient := &http.Client{Timeout: 30 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		t.Fatalf("POST /scim/v2/Groups: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		errBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("create SCIM group: status %d body: %s", resp.StatusCode, errBody)
	}

	var res struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		t.Fatalf("decode SCIM group response: %v", err)
	}
	if res.ID == "" {
		t.Fatal("SCIM group response missing id")
	}
	return res.ID
}
