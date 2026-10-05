// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/go-tfe/v2/api/models"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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
			},
		})
	})

	t.Run("SCIM settings site admin group", func(t *testing.T) {
		var siteAdminGroupID string
		var siteAdminGroupName string
		var siteAdminGroupBID string
		var siteAdminGroupBName string

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
						tokenName := "tf-acc-test-scim-token-" + randomString(t)
						token, err := testAccConfiguredClient.Client.Admin.Settings.SCIM.Tokens.Create(
							context.Background(), tokenName,
						)
						if err != nil {
							t.Fatalf("create SCIM token: %v", err)
						}
						t.Cleanup(func() {
							_ = testAccConfiguredClient.Client.Admin.Settings.SCIM.Tokens.Delete(context.Background(), token.ID)
						})

						// No explicit group cleanup: disabling SCIM (CheckDestroy) removes all groups from the backend.
						siteAdminGroupName = "tf-acc-site-admins-" + randomString(t)
						siteAdminGroupID = createSCIMGroup(t, siteAdminGroupName, token.Token)
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
						resource.TestCheckResourceAttrPtr(
							"tfe_scim_settings.enable_scim",
							"site_admin_group_display_name",
							&siteAdminGroupName,
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
				// Switch from group A to group B (non-null → non-null).
				{
					PreConfig: func() {
						tokenName := "tf-acc-test-scim-token-b-" + randomString(t)
						token, err := testAccConfiguredClient.Client.Admin.Settings.SCIM.Tokens.Create(
							context.Background(), tokenName,
						)
						if err != nil {
							t.Fatalf("create SCIM token for group B: %v", err)
						}
						t.Cleanup(func() {
							_ = testAccConfiguredClient.Client.Admin.Settings.SCIM.Tokens.Delete(context.Background(), token.ID)
						})
						// No explicit group cleanup: disabling SCIM (CheckDestroy) removes all groups from the backend.
						siteAdminGroupBName = "tf-acc-site-admins-b-" + randomString(t)
						siteAdminGroupBID = createSCIMGroup(t, siteAdminGroupBName, token.Token)
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
						resource.TestCheckResourceAttrPtr(
							"tfe_scim_settings.enable_scim",
							"site_admin_group_display_name",
							&siteAdminGroupBName,
						),
					),
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

	// Linking a SCIM group to the Site Auditor role requires TFE
	// minTFEVersionSiteAuditor or later. Against an older release the provider
	// fails this subtest with an explicit minimum-version error rather than
	// silently doing nothing, which is the behaviour we want to surface.
	t.Run("SCIM settings site auditor group", func(t *testing.T) {
		var siteAuditorGroupID string
		var siteAuditorGroupName string

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			CheckDestroy:             testAccTFESCIMSettingsDestroy,
			Steps: []resource.TestStep{
				// Enable SCIM with no site auditor group linked.
				{
					Config: testAccTFESCIMSettings_enable(),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("tfe_scim_settings.enable_scim", "enabled", "true"),
						resource.TestCheckResourceAttr("tfe_scim_settings.enable_scim", "site_auditor_group_scim_id", ""),
						resource.TestCheckResourceAttr("tfe_scim_settings.enable_scim", "site_auditor_group_display_name", ""),
					),
				},
				// Create a SCIM group out-of-band and link it via TF_VAR.
				{
					PreConfig: func() {
						tokenName := "tf-acc-test-scim-token-auditor-" + randomString(t)
						token, err := testAccConfiguredClient.Client.Admin.Settings.SCIM.Tokens.Create(
							context.Background(), tokenName,
						)
						if err != nil {
							t.Fatalf("create SCIM token for the site auditor group: %v", err)
						}
						t.Cleanup(func() {
							_ = testAccConfiguredClient.Client.Admin.Settings.SCIM.Tokens.Delete(context.Background(), token.ID)
						})

						// No explicit group cleanup: disabling SCIM (CheckDestroy) removes all groups from the backend.
						siteAuditorGroupName = "tf-acc-site-auditors-" + randomString(t)
						siteAuditorGroupID = createSCIMGroup(t, siteAuditorGroupName, token.Token)
						t.Setenv("TF_VAR_site_auditor_group_scim_id", siteAuditorGroupID)
					},
					Config: testAccTFESCIMSettings_withSiteAuditorGroup(),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("tfe_scim_settings.enable_scim", "enabled", "true"),
						resource.TestCheckResourceAttrPtr(
							"tfe_scim_settings.enable_scim",
							"site_auditor_group_scim_id",
							&siteAuditorGroupID,
						),
						resource.TestCheckResourceAttrPtr(
							"tfe_scim_settings.enable_scim",
							"site_auditor_group_display_name",
							&siteAuditorGroupName,
						),
						// Linking the auditor group must not disturb the admin group.
						resource.TestCheckResourceAttr("tfe_scim_settings.enable_scim", "site_admin_group_scim_id", ""),
						// The data source reports the same values.
						resource.TestCheckResourceAttrPtr(
							"data.tfe_scim_settings.foobar",
							"site_auditor_group_scim_id",
							&siteAuditorGroupID,
						),
						resource.TestCheckResourceAttrPtr(
							"data.tfe_scim_settings.foobar",
							"site_auditor_group_display_name",
							&siteAuditorGroupName,
						),
					),
				},
				// Re-apply same config: should be a no-op (no perpetual diff).
				{
					Config:   testAccTFESCIMSettings_withSiteAuditorGroup(),
					PlanOnly: true,
				},
				// Import round-trips the linked group through state.
				{
					ResourceName:      "tfe_scim_settings.enable_scim",
					ImportState:       true,
					ImportStateId:     "scim",
					ImportStateVerify: true,
				},
				// Clear the site auditor group by setting it to "".
				{
					Config: testAccTFESCIMSettings_clearSiteAuditorGroup(),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("tfe_scim_settings.enable_scim", "enabled", "true"),
						resource.TestCheckResourceAttr("tfe_scim_settings.enable_scim", "site_auditor_group_scim_id", ""),
						resource.TestCheckResourceAttr("tfe_scim_settings.enable_scim", "site_auditor_group_display_name", ""),
					),
				},
				// Re-link the same group, so the next step can prove that
				// omitting the argument (not just emptying it) unlinks.
				{
					Config: testAccTFESCIMSettings_withSiteAuditorGroup(),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrPtr(
							"tfe_scim_settings.enable_scim",
							"site_auditor_group_scim_id",
							&siteAuditorGroupID,
						),
					),
				},
				// Omitting site_auditor_group_scim_id reverts to the default (""), unlinking the group.
				{
					Config: testAccTFESCIMSettings_enable(),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("tfe_scim_settings.enable_scim", "enabled", "true"),
						resource.TestCheckResourceAttr("tfe_scim_settings.enable_scim", "site_auditor_group_scim_id", ""),
						resource.TestCheckResourceAttr("tfe_scim_settings.enable_scim", "site_auditor_group_display_name", ""),
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

	t.Run("destroy when SCIM already disabled out-of-band", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			CheckDestroy:             testAccTFESCIMSettingsDestroy,
			Steps: []resource.TestStep{
				// Enable SCIM.
				{
					Config: testAccTFESCIMSettings_enable(),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("tfe_scim_settings.enable_scim", "enabled", "true"),
					),
				},
				// Disable SCIM out-of-band, then refresh: Read should remove the resource
				// from state without error, and the subsequent destroy should be a no-op.
				{
					PreConfig: func() {
						if err := testAccConfiguredClient.Client.Admin.Settings.SCIM.Delete(ctx); err != nil {
							t.Fatalf("disable SCIM out-of-band: %v", err)
						}
					},
					RefreshState:       true,
					ExpectNonEmptyPlan: true, // config still wants the resource; Terraform plans to re-create it
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

	// The go-tfe v1 struct read above predates the Site Auditor attributes, so
	// the auditor group needs a second read through the v2 client. On releases
	// older than minTFEVersionSiteAuditor the attribute is simply absent, which
	// valueOrZero reports as "" — the same as unlinked, so this check is safe
	// to run unconditionally.
	env, err := testAccConfiguredClient.ClientV2.API.Admin().ScimSettings().Get(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to read SCIM Settings via go-tfe v2: %s", apiErrorDetail(err))
	}
	if env == nil || env.GetData() == nil || env.GetData().GetAttributes() == nil {
		return errors.New("SCIM settings response did not contain any data")
	}
	if valueOrZero(env.GetData().GetAttributes().GetSiteAuditorGroupScimId()) != "" {
		return errors.New("SCIM Settings still have site auditor group linked")
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

func testAccTFESCIMSettings_withSiteAuditorGroup() string {
	return fmt.Sprintf(`
%s

variable "site_auditor_group_scim_id" {
	type = string
}
resource "tfe_scim_settings" "enable_scim" {
	site_auditor_group_scim_id = var.site_auditor_group_scim_id
	depends_on                 = [tfe_saml_settings.enable_saml]
}

data "tfe_scim_settings" "foobar" {
	depends_on = [tfe_scim_settings.enable_scim]
}
`, testAccTFESCIMSettings_enableSAMLWithProviderType(scimTestSAMLSetting))
}

func testAccTFESCIMSettings_clearSiteAuditorGroup() string {
	return fmt.Sprintf(`
%s

resource "tfe_scim_settings" "enable_scim" {
	site_auditor_group_scim_id = ""
	depends_on                 = [tfe_saml_settings.enable_saml]
}
`, testAccTFESCIMSettings_enableSAMLWithProviderType(scimTestSAMLSetting))
}

// scimSettingsEnvelopeForTest builds a minimal admin SCIM settings response.
// Attributes left unset return nil from their getters, which is how both an
// unlinked group and a Terraform Enterprise release predating the attribute
// arrive at the provider.
func scimSettingsEnvelopeForTest(mutate func(*models.AdminScimSettings_attributes)) models.AdminScimSettingsEnvelopeable {
	attrs := models.NewAdminScimSettings_attributes()
	if mutate != nil {
		mutate(attrs)
	}
	return scimSettingsEnvelope(attrs)
}

// TestSCIMSettingsSiteAuditorAbsentAttributes pins what the provider records
// when the server sends no Site Auditor attributes. Two different situations
// produce that response and both must land on "" rather than null: an unlinked
// group (Terraform Enterprise delegates the attribute to the group with
// allow_nil, so it serializes as JSON null) and a release older than
// minTFEVersionSiteAuditor (the attribute does not exist at all). "" is the
// schema default, so plan and state agree; null would show up as a spurious
// diff on an untouched resource.
func TestSCIMSettingsSiteAuditorAbsentAttributes(t *testing.T) {
	result, err := modelFromV2SCIMSettings(scimSettingsEnvelopeForTest(nil))
	if err != nil {
		t.Fatalf("modelFromV2SCIMSettings: %v", err)
	}
	if result.SiteAuditorGroupSCIMID.IsNull() || result.SiteAuditorGroupDisplayName.IsNull() {
		t.Fatal("Site Auditor attributes recorded as null; this produces a spurious diff against the \"\" schema default")
	}
	if got := result.SiteAuditorGroupSCIMID.ValueString(); got != "" {
		t.Errorf("site_auditor_group_scim_id = %q, want \"\"", got)
	}
	if got := result.SiteAuditorGroupDisplayName.ValueString(); got != "" {
		t.Errorf("site_auditor_group_display_name = %q, want \"\"", got)
	}
}

// TestSCIMSettingsSiteAuditorServerValues covers the supported-release path,
// where both attributes come back populated.
func TestSCIMSettingsSiteAuditorServerValues(t *testing.T) {
	env := scimSettingsEnvelopeForTest(func(a *models.AdminScimSettings_attributes) {
		a.SetEnabled(ptr(true))
		a.SetSiteAuditorGroupScimId(ptr("scim-group-auditors"))
		a.SetSiteAuditorGroupDisplayName(ptr("Site Auditors"))
	})
	result, err := modelFromV2SCIMSettings(env)
	if err != nil {
		t.Fatalf("modelFromV2SCIMSettings: %v", err)
	}
	if !result.Enabled.ValueBool() {
		t.Error("enabled = false, want true")
	}
	if got := result.SiteAuditorGroupSCIMID.ValueString(); got != "scim-group-auditors" {
		t.Errorf("site_auditor_group_scim_id = %q, want scim-group-auditors", got)
	}
	if got := result.SiteAuditorGroupDisplayName.ValueString(); got != "Site Auditors" {
		t.Errorf("site_auditor_group_display_name = %q, want \"Site Auditors\"", got)
	}
}

// TestSCIMSettingsGroupSCIMIDForRequest pins the wire value for each planned
// state of a group ID. The empty string matters: the generated client omits a
// nil pointer from the request body entirely, and Terraform Enterprise reads an
// absent key as "leave this link alone" rather than "unlink". Sending "" is
// what actually unlinks a group.
func TestSCIMSettingsGroupSCIMIDForRequest(t *testing.T) {
	for _, tc := range []struct {
		name    string
		planned types.String
		want    string
		wantErr bool
	}{
		{name: "a linked group is sent as its ID", planned: types.StringValue("scim-group-1"), want: "scim-group-1"},
		{name: "an empty value unlinks", planned: types.StringValue(""), want: ""},
		{name: "a null value unlinks rather than being omitted", planned: types.StringNull(), want: ""},
		{name: "an unresolved value is rejected", planned: types.StringUnknown(), wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := groupSCIMIDForRequest(tc.planned, "site_auditor_group_scim_id")
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected an error for an unknown value, got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("groupSCIMIDForRequest: %v", err)
			}
			if got == nil {
				t.Fatal("group ID sent as a nil pointer; the generated client omits it from the request body, which leaves the existing link in place")
			}
			if *got != tc.want {
				t.Errorf("group ID on the wire = %q, want %q", *got, tc.want)
			}
		})
	}
}

// TestSCIMSettingsEmptyResponse checks that a response carrying no data is
// reported as an error instead of being silently written to state as a set of
// zero values, which would look like SCIM had been disabled.
func TestSCIMSettingsEmptyResponse(t *testing.T) {
	if _, err := modelFromV2SCIMSettings(models.NewAdminScimSettingsEnvelope()); err == nil {
		t.Fatal("expected an error for a response with no data, got none")
	}
}
