// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const testResourceName = "tfe_saml_settings.foobar"

func TestAccTFESAMLSettings_basic(t *testing.T) {
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
}

func TestAccTFESAMLSettings_full(t *testing.T) {
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
}

func TestAccTFESAMLSettings_update(t *testing.T) {
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
}

func TestAccTFESAMLSettings_import(t *testing.T) {
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
}

func testAccTFESAMLSettingsDestroy(_ *terraform.State) error {
	s, err := testAccProvider.Meta().(ConfiguredClient).Client.Admin.Settings.SAML.Read(ctx)
	if err != nil {
		return fmt.Errorf("failed to read SAML Settings: %v", err)
	}
	if s.Enabled {
		return fmt.Errorf("SAML settings are still enabled")
	}
	if s.Debug {
		return fmt.Errorf("SAML settings debug is set to true")
	}
	if s.AuthnRequestsSigned {
		return fmt.Errorf("SAML settings AuthnRequestsSigned is set to true")
	}
	if s.WantAssertionsSigned {
		return fmt.Errorf("SAML settings WantAssertionsSigned is set to true")
	}
	if s.TeamManagementEnabled {
		return fmt.Errorf("SAML settings TeamManagementEnabled is set to true")
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
		return fmt.Errorf("SAML settings PrivateKey is not empty")
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
