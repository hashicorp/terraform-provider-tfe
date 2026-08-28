// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/go-tfe/v2/api/models"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

const testResourceName = "tfe_saml_settings.foobar"

// FLAKE ALERT: SAML settings are a singleton resource shared by the entire TFE
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

// TestAccTFESAMLSettings_omnibus test suite is skipped in the CI, and will only run in TFE Nightly workflow
// Should this test name ever change, you will also need to update the regex in ci.yml
func TestAccTFESAMLSettings_writeOnly(t *testing.T) {
	s := tfe.AdminSAMLSetting{
		IDPCert:        "testIDPCertBasic",
		SLOEndpointURL: "https://foobar.com/slo_endpoint_url",
		SSOEndpointURL: "https://foobar.com/sso_endpoint_url",
		PrivateKey:     "TestPrivateKeyFull",
	}
	resource.Test(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(version.Must(version.NewVersion("1.11.0"))),
		},
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFESAMLSettings_writeOnly(s),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(testResourceName, "debug", "false"),
					resource.TestCheckResourceAttr(testResourceName, "authn_requests_signed", "false"),
					resource.TestCheckResourceAttr(testResourceName, "want_assertions_signed", "false"),
					resource.TestCheckResourceAttr(testResourceName, "team_management_enabled", "false"),
					resource.TestCheckResourceAttr(testResourceName, "idp_cert", s.IDPCert),
					resource.TestCheckResourceAttr(testResourceName, "slo_endpoint_url", s.SLOEndpointURL),
					resource.TestCheckResourceAttr(testResourceName, "sso_endpoint_url", s.SSOEndpointURL),
					resource.TestCheckResourceAttr(testResourceName, "attr_username", samlDefaultAttrUsername),
					resource.TestCheckResourceAttr(testResourceName, "attr_site_admin", samlDefaultAttrSiteAdmin),
					resource.TestCheckResourceAttr(testResourceName, "attr_groups", samlDefaultAttrGroups),
					resource.TestCheckResourceAttr(testResourceName, "site_admin_role", samlDefaultSiteAdminRole),
					resource.TestCheckResourceAttr(testResourceName, "sso_api_token_session_timeout", strconv.Itoa(int(samlDefaultSSOAPITokenSessionTimeoutSeconds))),
					resource.TestCheckResourceAttrSet(testResourceName, "acs_consumer_url"),
					resource.TestCheckResourceAttrSet(testResourceName, "metadata_url"),
					resource.TestCheckResourceAttr(testResourceName, "signature_signing_method", samlSignatureMethodSHA256),
					resource.TestCheckResourceAttr(testResourceName, "signature_digest_method", samlSignatureMethodSHA256),
					resource.TestCheckNoResourceAttr(
						testResourceName, "private_key_wo"),
					resource.TestCheckResourceAttr(testResourceName, "private_key_wo_version", "1"),
					resource.TestCheckResourceAttr(testResourceName, "provider_type", string(tfe.SAMLProviderTypeUnknown)),
				),
			},
		},
	})
}
func TestAccTFESAMLSettings_writeOnlyValidation(t *testing.T) {
	resource.Test(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(version.Must(version.NewVersion("1.11.0"))),
		},
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFESAMLSettings_privateKeyAndPrivateKeyWO(),
				ExpectError: regexp.MustCompile(`Attribute "private_key_wo" cannot be specified when "private_key" is\s+specified`),
			},
			{
				Config:      testAccTFESAMLSettings_privateKeyWOMissingVersion(),
				ExpectError: regexp.MustCompile(`Attribute "private_key_wo_version" must be specified when "private_key_wo" is\s+specified`),
			},
			{
				Config:      testAccTFESAMLSettings_versionMissingPrivateKeyWO(),
				ExpectError: regexp.MustCompile(`Attribute "private_key_wo" must be specified when "private_key_wo_version" is\s+specified`),
			},
			{
				Config:      testAccTFESAMLSettings_privateKeyVersionConflict(),
				ExpectError: regexp.MustCompile(`Attribute "private_key" cannot be specified when "private_key_wo_version" is\s+specified`),
			},
			{
				Config:      testAccTFESAMLSettings_samlProviderTypeInvalidValues(),
				ExpectError: regexp.MustCompile(`(?s)Attribute provider_type value must be one of: \[.*\]`),
			},
		},
	})
}

func TestAccTFESAMLSettings_omnibus(t *testing.T) {
	t.Run("basic SAML settings resource", func(t *testing.T) {
		s := tfe.AdminSAMLSetting{
			IDPCert:        "testIDPCertBasic",
			SLOEndpointURL: "https://foobar.com/slo_endpoint_url",
			SSOEndpointURL: "https://foobar.com/sso_endpoint_url",
		}
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			CheckDestroy:             testAccTFESAMLSettingsDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccTFESAMLSettings_basic(s),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(testResourceName, "enabled", "true"),
						resource.TestCheckResourceAttr(testResourceName, "debug", "false"),
						resource.TestCheckResourceAttr(testResourceName, "authn_requests_signed", "false"),
						resource.TestCheckResourceAttr(testResourceName, "want_assertions_signed", "false"),
						resource.TestCheckResourceAttr(testResourceName, "team_management_enabled", "false"),
						resource.TestCheckResourceAttr(testResourceName, "idp_cert", s.IDPCert),
						resource.TestCheckResourceAttr(testResourceName, "slo_endpoint_url", s.SLOEndpointURL),
						resource.TestCheckResourceAttr(testResourceName, "sso_endpoint_url", s.SSOEndpointURL),
						resource.TestCheckResourceAttr(testResourceName, "attr_username", samlDefaultAttrUsername),
						resource.TestCheckResourceAttr(testResourceName, "attr_site_admin", samlDefaultAttrSiteAdmin),
						resource.TestCheckResourceAttr(testResourceName, "attr_groups", samlDefaultAttrGroups),
						resource.TestCheckResourceAttr(testResourceName, "site_admin_role", samlDefaultSiteAdminRole),
						resource.TestCheckResourceAttr(testResourceName, "attr_site_auditor", samlDefaultAttrSiteAuditor),
						resource.TestCheckResourceAttr(testResourceName, "site_auditor_role", samlDefaultSiteAuditorRole),
						resource.TestCheckResourceAttr(testResourceName, "sso_api_token_session_timeout", strconv.Itoa(int(samlDefaultSSOAPITokenSessionTimeoutSeconds))),
						resource.TestCheckResourceAttrSet(testResourceName, "acs_consumer_url"),
						resource.TestCheckResourceAttrSet(testResourceName, "metadata_url"),
						resource.TestCheckResourceAttr(testResourceName, "signature_signing_method", samlSignatureMethodSHA256),
						resource.TestCheckResourceAttr(testResourceName, "signature_digest_method", samlSignatureMethodSHA256),
						resource.TestCheckResourceAttr(testResourceName, "provider_type", string(tfe.SAMLProviderTypeUnknown)),
					),
				},
			},
		})
	})

	t.Run("full SAML settings resource", func(t *testing.T) {
		s := tfe.AdminSAMLSetting{
			IDPCert:                   "testIDPCertFull",
			SLOEndpointURL:            "https://foobar.com/slo_endpoint_url",
			SSOEndpointURL:            "https://foobar.com/sso_endpoint_url",
			Debug:                     true,
			AuthnRequestsSigned:       true,
			WantAssertionsSigned:      true,
			TeamManagementEnabled:     false,
			AttrUsername:              "Foo" + samlDefaultAttrUsername,
			AttrSiteAdmin:             "Foo" + samlDefaultAttrSiteAdmin,
			AttrGroups:                "Foo" + samlDefaultAttrGroups,
			SiteAdminRole:             "foo-" + samlDefaultSiteAdminRole,
			SSOAPITokenSessionTimeout: 1101100,
			Certificate:               "TestCertificateFull",
			PrivateKey:                "TestPrivateKeyFull",
			SignatureSigningMethod:    samlSignatureMethodSHA1,
			SignatureDigestMethod:     samlSignatureMethodSHA256,
			ProviderType:              tfe.SAMLProviderTypeOkta,
		}
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			CheckDestroy:             testAccTFESAMLSettingsDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccTFESAMLSettings_full(s),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(testResourceName, "enabled", "true"),
						resource.TestCheckResourceAttr(testResourceName, "debug", strconv.FormatBool(s.Debug)),
						resource.TestCheckResourceAttr(testResourceName, "authn_requests_signed", strconv.FormatBool(s.AuthnRequestsSigned)),
						resource.TestCheckResourceAttr(testResourceName, "want_assertions_signed", strconv.FormatBool(s.WantAssertionsSigned)),
						resource.TestCheckResourceAttr(testResourceName, "team_management_enabled", strconv.FormatBool(s.TeamManagementEnabled)),
						resource.TestCheckResourceAttr(testResourceName, "idp_cert", s.IDPCert),
						resource.TestCheckResourceAttr(testResourceName, "slo_endpoint_url", s.SLOEndpointURL),
						resource.TestCheckResourceAttr(testResourceName, "sso_endpoint_url", s.SSOEndpointURL),
						resource.TestCheckResourceAttr(testResourceName, "attr_username", s.AttrUsername),
						resource.TestCheckResourceAttr(testResourceName, "attr_site_admin", s.AttrSiteAdmin),
						resource.TestCheckResourceAttr(testResourceName, "attr_groups", s.AttrGroups),
						resource.TestCheckResourceAttr(testResourceName, "site_admin_role", s.SiteAdminRole),
						resource.TestCheckResourceAttr(testResourceName, "sso_api_token_session_timeout", strconv.Itoa(s.SSOAPITokenSessionTimeout)),
						resource.TestCheckResourceAttrSet(testResourceName, "acs_consumer_url"),
						resource.TestCheckResourceAttrSet(testResourceName, "metadata_url"),
						resource.TestCheckResourceAttr(testResourceName, "signature_signing_method", s.SignatureSigningMethod),
						resource.TestCheckResourceAttr(testResourceName, "signature_digest_method", s.SignatureDigestMethod),
						resource.TestCheckResourceAttr(testResourceName, "provider_type", string(tfe.SAMLProviderTypeOkta)),
					),
				},
			},
		})
	})

	t.Run("SAML settings update", func(t *testing.T) {
		s := tfe.AdminSAMLSetting{
			IDPCert:        "testIDPCertUpdateInit",
			SLOEndpointURL: "https://foobar.com/slo_endpoint_url",
			SSOEndpointURL: "https://foobar.com/sso_endpoint_url",
		}
		updatedSetting := tfe.AdminSAMLSetting{
			IDPCert:                   "testIDPCertUpdateInit",
			SLOEndpointURL:            "https://foobar-updated.com/slo_endpoint_url",
			SSOEndpointURL:            "https://foobar-updated.com/sso_endpoint_url",
			Debug:                     true,
			AuthnRequestsSigned:       true,
			WantAssertionsSigned:      true,
			TeamManagementEnabled:     false,
			AttrUsername:              "FooUpdate" + samlDefaultAttrUsername,
			AttrSiteAdmin:             "FooUpdate" + samlDefaultAttrSiteAdmin,
			AttrGroups:                "FooUpdate" + samlDefaultAttrGroups,
			SiteAdminRole:             "foo-update-" + samlDefaultSiteAdminRole,
			SSOAPITokenSessionTimeout: 1234567,
			Certificate:               "TestCertificateUpdate",
			PrivateKey:                "TestPrivateKeyUpdate",
			SignatureSigningMethod:    samlSignatureMethodSHA1,
			SignatureDigestMethod:     samlSignatureMethodSHA256,
			ProviderType:              tfe.SAMLProviderTypeEntra,
		}

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			CheckDestroy:             testAccTFESAMLSettingsDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccTFESAMLSettings_basic(s),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(testResourceName, "enabled", "true"),
						resource.TestCheckResourceAttr(testResourceName, "debug", "false"),
						resource.TestCheckResourceAttr(testResourceName, "authn_requests_signed", "false"),
						resource.TestCheckResourceAttr(testResourceName, "want_assertions_signed", "false"),
						resource.TestCheckResourceAttr(testResourceName, "team_management_enabled", "false"),
						resource.TestCheckResourceAttr(testResourceName, "idp_cert", s.IDPCert),
						resource.TestCheckResourceAttr(testResourceName, "slo_endpoint_url", s.SLOEndpointURL),
						resource.TestCheckResourceAttr(testResourceName, "sso_endpoint_url", s.SSOEndpointURL),
						resource.TestCheckResourceAttr(testResourceName, "attr_username", samlDefaultAttrUsername),
						resource.TestCheckResourceAttr(testResourceName, "attr_site_admin", samlDefaultAttrSiteAdmin),
						resource.TestCheckResourceAttr(testResourceName, "attr_groups", samlDefaultAttrGroups),
						resource.TestCheckResourceAttr(testResourceName, "site_admin_role", samlDefaultSiteAdminRole),
						resource.TestCheckResourceAttr(testResourceName, "sso_api_token_session_timeout", strconv.Itoa(int(samlDefaultSSOAPITokenSessionTimeoutSeconds))),
						resource.TestCheckResourceAttrSet(testResourceName, "acs_consumer_url"),
						resource.TestCheckResourceAttrSet(testResourceName, "metadata_url"),
						resource.TestCheckResourceAttr(testResourceName, "signature_signing_method", samlSignatureMethodSHA256),
						resource.TestCheckResourceAttr(testResourceName, "signature_digest_method", samlSignatureMethodSHA256),
						resource.TestCheckResourceAttr(testResourceName, "provider_type", string(tfe.SAMLProviderTypeUnknown)),
					),
				},
				{
					Config: testAccTFESAMLSettings_full(updatedSetting),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(testResourceName, "enabled", "true"),
						resource.TestCheckResourceAttr(testResourceName, "debug", strconv.FormatBool(updatedSetting.Debug)),
						resource.TestCheckResourceAttr(testResourceName, "authn_requests_signed", strconv.FormatBool(updatedSetting.AuthnRequestsSigned)),
						resource.TestCheckResourceAttr(testResourceName, "want_assertions_signed", strconv.FormatBool(updatedSetting.WantAssertionsSigned)),
						resource.TestCheckResourceAttr(testResourceName, "team_management_enabled", strconv.FormatBool(updatedSetting.TeamManagementEnabled)),
						resource.TestCheckResourceAttr(testResourceName, "idp_cert", updatedSetting.IDPCert),
						resource.TestCheckResourceAttr(testResourceName, "slo_endpoint_url", updatedSetting.SLOEndpointURL),
						resource.TestCheckResourceAttr(testResourceName, "sso_endpoint_url", updatedSetting.SSOEndpointURL),
						resource.TestCheckResourceAttr(testResourceName, "attr_username", updatedSetting.AttrUsername),
						resource.TestCheckResourceAttr(testResourceName, "attr_site_admin", updatedSetting.AttrSiteAdmin),
						resource.TestCheckResourceAttr(testResourceName, "attr_groups", updatedSetting.AttrGroups),
						resource.TestCheckResourceAttr(testResourceName, "site_admin_role", updatedSetting.SiteAdminRole),
						resource.TestCheckResourceAttr(testResourceName, "sso_api_token_session_timeout", strconv.Itoa(updatedSetting.SSOAPITokenSessionTimeout)),
						resource.TestCheckResourceAttrSet(testResourceName, "acs_consumer_url"),
						resource.TestCheckResourceAttrSet(testResourceName, "metadata_url"),
						resource.TestCheckResourceAttr(testResourceName, "signature_signing_method", updatedSetting.SignatureSigningMethod),
						resource.TestCheckResourceAttr(testResourceName, "signature_digest_method", updatedSetting.SignatureDigestMethod),
						resource.TestCheckResourceAttr(testResourceName, "provider_type", string(tfe.SAMLProviderTypeEntra)),
					),
				},
			},
		})
	})

	// Site Auditor SAML provisioning requires TFE minTFEVersionSiteAuditor or
	// later. Against an older release the provider fails this subtest with an
	// explicit minimum-version error rather than a confusing inconsistent-result
	// error, which is the behaviour we want to surface.
	t.Run("SAML settings with Site Auditor", func(t *testing.T) {
		attrSiteAuditor := "SiteAuditorAttr"
		siteAuditorRole := "site-auditors-custom"
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			CheckDestroy:             testAccTFESAMLSettingsDestroy,
			Steps: []resource.TestStep{
				{
					// Explicitly configured Site Auditor attributes round-trip.
					Config: testAccTFESAMLSettings_siteAuditor(attrSiteAuditor, siteAuditorRole),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(testResourceName, "attr_site_auditor", attrSiteAuditor),
						resource.TestCheckResourceAttr(testResourceName, "site_auditor_role", siteAuditorRole),
						// The data source reports the same values.
						resource.TestCheckResourceAttr("data.tfe_saml_settings.foobar", "attr_site_auditor", attrSiteAuditor),
						resource.TestCheckResourceAttr("data.tfe_saml_settings.foobar", "site_auditor_role", siteAuditorRole),
					),
				},
				{
					// Updating just one of the pair leaves the other intact.
					Config: testAccTFESAMLSettings_siteAuditor(attrSiteAuditor, "site-auditors-updated"),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(testResourceName, "attr_site_auditor", attrSiteAuditor),
						resource.TestCheckResourceAttr(testResourceName, "site_auditor_role", "site-auditors-updated"),
					),
				},
				{
					// Dropping the attributes from config falls back to the
					// schema defaults rather than clearing them server-side.
					Config: testAccTFESAMLSettings_basic(tfe.AdminSAMLSetting{
						IDPCert:        "testIDPCertSiteAuditor",
						SLOEndpointURL: "https://foobar.com/slo_endpoint_url",
						SSOEndpointURL: "https://foobar.com/sso_endpoint_url",
					}),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(testResourceName, "attr_site_auditor", samlDefaultAttrSiteAuditor),
						resource.TestCheckResourceAttr(testResourceName, "site_auditor_role", samlDefaultSiteAuditorRole),
					),
				},
			},
		})
	})

	t.Run("SAML settings import", func(t *testing.T) {
		idpCert := "testIDPCertImport"
		slo := "https://foobar-import.com/slo_endpoint_url"
		sso := "https://foobar-import.com/sso_endpoint_url"
		s := tfe.AdminSAMLSetting{
			IDPCert:        idpCert,
			SLOEndpointURL: slo,
			SSOEndpointURL: sso,
		}
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			CheckDestroy:             testAccTFESAMLSettingsDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccTFESAMLSettings_basic(s),
				},
				{
					ResourceName: testResourceName,
					ImportState:  true,
					ImportStateCheck: func(s []*terraform.InstanceState) error {
						if len(s) != 1 {
							return fmt.Errorf("expected 1 state: %+v", s)
						}
						rs := s[0]
						if rs.Attributes["private_key"] != "" {
							return fmt.Errorf("expected private_key attribute to not be set, received: %s", rs.Attributes["private_key"])
						}
						if rs.Attributes["idp_cert"] != idpCert {
							return fmt.Errorf("expected idp_cert attribute to be equal to %s, received: %s", idpCert, rs.Attributes["idp_cert"])
						}
						if rs.Attributes["slo_endpoint_url"] != slo {
							return fmt.Errorf("expected slo_endpoint_url attribute to be equal to %s, received: %s", slo, rs.Attributes["slo_endpoint_url"])
						}
						if rs.Attributes["sso_endpoint_url"] != sso {
							return fmt.Errorf("expected sso_endpoint_url attribute to be equal to %s, received: %s", sso, rs.Attributes["sso_endpoint_url"])
						}

						if rs.Attributes["provider_type"] != string(tfe.SAMLProviderTypeUnknown) {
							return fmt.Errorf("expected provider_type attribute to be equal to %s, received: %s", tfe.SAMLProviderTypeUnknown, rs.Attributes["provider_type"])
						}
						return nil
					},
				},
			},
		})
	})
}

func testAccTFESAMLSettingsDestroy(_ *terraform.State) error {
	s, err := testAccConfiguredClient.Client.Admin.Settings.SAML.Read(ctx)
	if err != nil {
		return fmt.Errorf("failed to read SAML Settings: %w", err)
	}
	if s.Enabled {
		return errors.New("SAML settings are still enabled")
	}
	if s.Debug {
		return errors.New("SAML settings debug is set to true")
	}
	if s.AuthnRequestsSigned {
		return errors.New("SAML settings AuthnRequestsSigned is set to true")
	}
	if s.WantAssertionsSigned {
		return errors.New("SAML settings WantAssertionsSigned is set to true")
	}
	if s.TeamManagementEnabled {
		return errors.New("SAML settings TeamManagementEnabled is set to true")
	}
	if s.IDPCert != "" {
		return fmt.Errorf("SAML settings IDPCert is not empty: `%s`", s.IDPCert)
	}
	if s.SLOEndpointURL != "" {
		return fmt.Errorf("SAML settings SLOEndpointURL is not empty: `%s`", s.SLOEndpointURL)
	}
	if s.SSOEndpointURL != "" {
		return fmt.Errorf("SAML settings SSOEndpointURL is not empty: `%s`", s.SSOEndpointURL)
	}
	if s.Certificate != "" {
		return fmt.Errorf("SAML settings Certificate is not empty: `%s`", s.Certificate)
	}
	if s.PrivateKey != "" {
		return errors.New("SAML settings PrivateKey is not empty")
	}
	if s.AttrUsername != samlDefaultAttrUsername {
		return fmt.Errorf("SAML settings AttrUsername is not `%s`", samlDefaultAttrUsername)
	}
	if s.AttrSiteAdmin != samlDefaultAttrSiteAdmin {
		return fmt.Errorf("SAML settings AttrSiteAdmin is not `%s`", samlDefaultAttrSiteAdmin)
	}
	if s.AttrGroups != samlDefaultAttrGroups {
		return fmt.Errorf("SAML settings AttrGroups is not `%s`", samlDefaultAttrGroups)
	}
	if s.SiteAdminRole != samlDefaultSiteAdminRole {
		return fmt.Errorf("SAML settings SiteAdminRole is not `%s`", samlDefaultSiteAdminRole)
	}
	if s.SignatureSigningMethod != samlSignatureMethodSHA256 {
		return fmt.Errorf("SAML settings SignatureSigningMethod is not `%s`", samlSignatureMethodSHA256)
	}
	if s.SignatureDigestMethod != samlSignatureMethodSHA256 {
		return fmt.Errorf("SAML settings SignatureDigestMethod is not `%s`", samlSignatureMethodSHA256)
	}
	if s.SSOAPITokenSessionTimeout != int(samlDefaultSSOAPITokenSessionTimeoutSeconds) {
		return fmt.Errorf("SAML settings SignatureDigestMethod is not `%d`", samlDefaultSSOAPITokenSessionTimeoutSeconds)
	}
	if s.ProviderType != tfe.SAMLProviderTypeUnknown {
		return fmt.Errorf("SAML settings ProviderType is not `%s`", tfe.SAMLProviderTypeUnknown)
	}
	return nil
}

func testAccTFESAMLSettings_basic(s tfe.AdminSAMLSetting) string {
	return fmt.Sprintf(`
resource "tfe_saml_settings" "foobar" {
  idp_cert         = "%s"
  slo_endpoint_url = "%s"
  sso_endpoint_url = "%s"
}`, s.IDPCert, s.SLOEndpointURL, s.SSOEndpointURL)
}

func testAccTFESAMLSettings_full(s tfe.AdminSAMLSetting) string {
	return fmt.Sprintf(`
resource "tfe_saml_settings" "foobar" {
  idp_cert         				= "%s"
  slo_endpoint_url 				= "%s"
  sso_endpoint_url 				= "%s"
  debug 		   				= %t
  authn_requests_signed 		= %t
  want_assertions_signed 		= %t
  team_management_enabled 		= %t
  attr_username 				= "%s"
  attr_site_admin 				= "%s"
  attr_groups 					= "%s"
  site_admin_role 				= "%s"
  sso_api_token_session_timeout = %d
  certificate 					= "%s"
  private_key 					= "%s"
  signature_signing_method 		= "%s"
  signature_digest_method 		= "%s"
  provider_type                 = "%s"
}`, s.IDPCert, s.SLOEndpointURL, s.SSOEndpointURL, s.Debug, s.AuthnRequestsSigned, s.WantAssertionsSigned, s.TeamManagementEnabled, s.AttrUsername, s.AttrSiteAdmin, s.AttrGroups, s.SiteAdminRole, s.SSOAPITokenSessionTimeout, s.Certificate, s.PrivateKey, s.SignatureSigningMethod, s.SignatureDigestMethod, s.ProviderType)
}

func testAccTFESAMLSettings_siteAuditor(attrSiteAuditor, siteAuditorRole string) string {
	return fmt.Sprintf(`
resource "tfe_saml_settings" "foobar" {
  idp_cert          = "testIDPCertSiteAuditor"
  slo_endpoint_url  = "https://foobar.com/slo_endpoint_url"
  sso_endpoint_url  = "https://foobar.com/sso_endpoint_url"
  attr_site_auditor = "%s"
  site_auditor_role = "%s"
}

data "tfe_saml_settings" "foobar" {
  depends_on = [tfe_saml_settings.foobar]
}`, attrSiteAuditor, siteAuditorRole)
}

func testAccTFESAMLSettings_writeOnly(s tfe.AdminSAMLSetting) string {
	return fmt.Sprintf(`
resource "tfe_saml_settings" "foobar" {
  idp_cert                 = "%s"
  slo_endpoint_url         = "%s"
  sso_endpoint_url         = "%s"
  private_key_wo           = "%s"
  private_key_wo_version   = 1
}`, s.IDPCert, s.SLOEndpointURL, s.SSOEndpointURL, s.PrivateKey)
}

func testAccTFESAMLSettings_privateKeyAndPrivateKeyWO() string {
	return `
resource "tfe_saml_settings" "foobar" {
  idp_cert               = "testIDPCert"
  slo_endpoint_url       = "https://foobar.com/slo"
  sso_endpoint_url       = "https://foobar.com/sso"
  private_key            = "some-key"
  private_key_wo         = "some-key"
  private_key_wo_version = 1
}`
}

func testAccTFESAMLSettings_privateKeyWOMissingVersion() string {
	return `
resource "tfe_saml_settings" "foobar" {
  idp_cert         = "testIDPCert"
  slo_endpoint_url = "https://foobar.com/slo"
  sso_endpoint_url = "https://foobar.com/sso"
  private_key_wo   = "some-key"
}`
}

func testAccTFESAMLSettings_versionMissingPrivateKeyWO() string {
	return `
resource "tfe_saml_settings" "foobar" {
  idp_cert               = "testIDPCert"
  slo_endpoint_url       = "https://foobar.com/slo"
  sso_endpoint_url       = "https://foobar.com/sso"
  private_key_wo_version = 1
}`
}

func testAccTFESAMLSettings_privateKeyVersionConflict() string {
	return `
resource "tfe_saml_settings" "foobar" {
  idp_cert               = "testIDPCert"
  slo_endpoint_url       = "https://foobar.com/slo"
  sso_endpoint_url       = "https://foobar.com/sso"
  private_key            = "some-key"
  private_key_wo_version = 1
}`
}

func testAccTFESAMLSettings_samlProviderTypeInvalidValues() string {
	return `
resource "tfe_saml_settings" "foobar" {
  idp_cert         = "testIDPCert"
  slo_endpoint_url = "https://foobar.com/slo"
  sso_endpoint_url = "https://foobar.com/sso"
  provider_type    = "foo"
}`
}

// samlSettingsEnvelopeForTest builds a minimal admin SAML settings response.
// Attributes left unset return nil from their getters, which is exactly how a
// Terraform Enterprise release that predates an attribute behaves.
func samlSettingsEnvelopeForTest(mutate func(*models.AdminSamlSettings_attributes)) models.AdminSamlSettingsEnvelopeable {
	attrs := models.NewAdminSamlSettings_attributes()
	if mutate != nil {
		mutate(attrs)
	}
	return samlSettingsEnvelope(attrs)
}

// TestSAMLSettingsPrivateKeyStateConsistency guards the invariant that the
// private_key value written to state always matches the planned value.
//
// Update reuses one field for two jobs: a null tells the request builder to
// omit private-key from the PATCH (leaving the stored key alone), but that
// sentinel must never become the state value. When it did, Terraform rejected
// the apply with "Provider produced inconsistent result after apply", surfaced
// opaquely because private_key is sensitive.
func TestSAMLSettingsPrivateKeyStateConsistency(t *testing.T) {
	for _, tc := range []struct {
		name     string
		planned  types.String
		expected string
	}{
		{"key absent from config, so plan holds the schema default", types.StringValue(""), ""},
		{"key present and unchanged", types.StringValue("KEY"), "KEY"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			plan := modelTFESAMLSettings{PrivateKey: tc.planned}
			state := modelTFESAMLSettings{PrivateKey: tc.planned}
			config := modelTFESAMLSettings{}
			r := &resourceTFESAMLSettings{}

			// Mirrors Update: capture the planned value for state, then null the
			// working copy when the key is unchanged.
			stateKey := plan.PrivateKey
			if pk := r.determinePrivateKeyForUpdate(plan, state, config); pk != nil {
				plan.PrivateKey = types.StringValue(*pk)
				stateKey = plan.PrivateKey
			} else {
				plan.PrivateKey = types.StringNull()
			}
			if !plan.PrivateKey.IsNull() {
				t.Fatalf("precondition: expected the request copy to be nulled for an unchanged key, got %s", plan.PrivateKey)
			}

			result, err := modelFromV2SAMLSettings(samlSettingsEnvelopeForTest(nil), stateKey, types.Int64Null(), plan)
			if err != nil {
				t.Fatalf("modelFromV2SAMLSettings: %v", err)
			}
			if result.PrivateKey.IsNull() {
				t.Errorf("private_key written to state as null while the plan held %q; Terraform would reject this apply", tc.expected)
			}
			if got := result.PrivateKey.ValueString(); got != tc.expected {
				t.Errorf("private_key in state = %q, want %q (the planned value)", got, tc.expected)
			}
		})
	}
}

// TestSAMLSettingsPrivateKeyNullNeverReachesState pins the guard in
// modelFromV2SAMLSettings directly: a null private_key argument is the
// "omit from the request" sentinel and must not be copied into state.
// types.String.String() renders null as "<null>", so a length check here looks
// like it filters nulls out but never does.
func TestSAMLSettingsPrivateKeyNullNeverReachesState(t *testing.T) {
	result, err := modelFromV2SAMLSettings(samlSettingsEnvelopeForTest(nil), types.StringNull(), types.Int64Null(), modelTFESAMLSettings{})
	if err != nil {
		t.Fatalf("modelFromV2SAMLSettings: %v", err)
	}
	if result.PrivateKey.IsNull() {
		t.Fatal("a null private_key sentinel was copied into state; Terraform would reject the apply as an inconsistent result")
	}
	if got := result.PrivateKey.ValueString(); got != "" {
		t.Errorf("private_key in state = %q, want \"\"", got)
	}
}

// TestSAMLSettingsSiteAuditorFallback covers what the provider records when the
// server omits the Site Auditor attributes, which every Terraform Enterprise
// release before minTFEVersionSiteAuditor does.
func TestSAMLSettingsSiteAuditorFallback(t *testing.T) {
	t.Run("falls back to the prior value so plan and state agree", func(t *testing.T) {
		prior := modelTFESAMLSettings{
			AttrSiteAuditor: types.StringValue("CustomAuditor"),
			SiteAuditorRole: types.StringValue("custom-auditors"),
		}
		result, err := modelFromV2SAMLSettings(samlSettingsEnvelopeForTest(nil), types.StringValue(""), types.Int64Null(), prior)
		if err != nil {
			t.Fatalf("modelFromV2SAMLSettings: %v", err)
		}
		if got := result.AttrSiteAuditor.ValueString(); got != "CustomAuditor" {
			t.Errorf("attr_site_auditor = %q, want the prior value CustomAuditor", got)
		}
		if got := result.SiteAuditorRole.ValueString(); got != "custom-auditors" {
			t.Errorf("site_auditor_role = %q, want the prior value custom-auditors", got)
		}
	})

	t.Run("falls back to the schema default when there is no prior value", func(t *testing.T) {
		// Import, and the first refresh after upgrading from a provider whose
		// state predates these attributes, both land here. Returning null would
		// show a spurious null -> default diff on an untouched resource.
		result, err := modelFromV2SAMLSettings(samlSettingsEnvelopeForTest(nil), types.StringValue(""), types.Int64Null(), modelTFESAMLSettings{})
		if err != nil {
			t.Fatalf("modelFromV2SAMLSettings: %v", err)
		}
		if result.AttrSiteAuditor.IsNull() || result.SiteAuditorRole.IsNull() {
			t.Fatal("Site Auditor attributes recorded as null with no prior value; this produces a spurious diff")
		}
		if got := result.AttrSiteAuditor.ValueString(); got != samlDefaultAttrSiteAuditor {
			t.Errorf("attr_site_auditor = %q, want %q", got, samlDefaultAttrSiteAuditor)
		}
		if got := result.SiteAuditorRole.ValueString(); got != samlDefaultSiteAuditorRole {
			t.Errorf("site_auditor_role = %q, want %q", got, samlDefaultSiteAuditorRole)
		}
	})

	t.Run("prefers the server value when the release returns it", func(t *testing.T) {
		env := samlSettingsEnvelopeForTest(func(a *models.AdminSamlSettings_attributes) {
			a.SetAttrSiteAuditor(ptr("ServerAuditor"))
			a.SetSiteAuditorRole(ptr("server-auditors"))
		})
		prior := modelTFESAMLSettings{AttrSiteAuditor: types.StringValue("Stale")}
		result, err := modelFromV2SAMLSettings(env, types.StringValue(""), types.Int64Null(), prior)
		if err != nil {
			t.Fatalf("modelFromV2SAMLSettings: %v", err)
		}
		if got := result.AttrSiteAuditor.ValueString(); got != "ServerAuditor" {
			t.Errorf("attr_site_auditor = %q, want the server value ServerAuditor", got)
		}
		if got := result.SiteAuditorRole.ValueString(); got != "server-auditors" {
			t.Errorf("site_auditor_role = %q, want the server value server-auditors", got)
		}
	})
}
