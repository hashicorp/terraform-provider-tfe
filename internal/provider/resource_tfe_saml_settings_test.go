// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/go-version"
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
		ProtoV5ProviderFactories: testAccMuxedProviders,
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
				),
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
			ProtoV5ProviderFactories: testAccMuxedProviders,
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
		}
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV5ProviderFactories: testAccMuxedProviders,
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
		}

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV5ProviderFactories: testAccMuxedProviders,
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
			ProtoV5ProviderFactories: testAccMuxedProviders,
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
}`, s.IDPCert, s.SLOEndpointURL, s.SSOEndpointURL, s.Debug, s.AuthnRequestsSigned, s.WantAssertionsSigned, s.TeamManagementEnabled, s.AttrUsername, s.AttrSiteAdmin, s.AttrGroups, s.SiteAdminRole, s.SSOAPITokenSessionTimeout, s.Certificate, s.PrivateKey, s.SignatureSigningMethod, s.SignatureDigestMethod)
}

func testAccTFESAMLSettings_writeOnly(s tfe.AdminSAMLSetting) string {
	return fmt.Sprintf(`
resource "tfe_saml_settings" "foobar" {
  idp_cert         = "%s"
  slo_endpoint_url = "%s"
  sso_endpoint_url = "%s"
  private_key_wo 					= "%s"
}`, s.IDPCert, s.SLOEndpointURL, s.SSOEndpointURL, s.PrivateKey)
}
