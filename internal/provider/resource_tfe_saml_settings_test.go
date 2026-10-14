// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/go-tfe/v2/api/models"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-tfe/internal/provider/customtypes"
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
		IDPCert:        testIDPCertBody(t),
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
			IDPCert:        testIDPCertBody(t),
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
			IDPCert:                   testIDPCertBody(t),
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
		idpCert := testIDPCertBody(t)
		s := tfe.AdminSAMLSetting{
			IDPCert:        idpCert,
			SLOEndpointURL: "https://foobar.com/slo_endpoint_url",
			SSOEndpointURL: "https://foobar.com/sso_endpoint_url",
		}
		updatedSetting := tfe.AdminSAMLSetting{
			IDPCert:                   idpCert,
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
		idpCert := testIDPCertBody(t)
		attrSiteAuditor := "SiteAuditorAttr"
		siteAuditorRole := "site-auditors-custom"
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			CheckDestroy:             testAccTFESAMLSettingsDestroy,
			Steps: []resource.TestStep{
				{
					// Explicitly configured Site Auditor attributes round-trip.
					Config: testAccTFESAMLSettings_siteAuditor(idpCert, attrSiteAuditor, siteAuditorRole),
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
					Config: testAccTFESAMLSettings_siteAuditor(idpCert, attrSiteAuditor, "site-auditors-updated"),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(testResourceName, "attr_site_auditor", attrSiteAuditor),
						resource.TestCheckResourceAttr(testResourceName, "site_auditor_role", "site-auditors-updated"),
					),
				},
				{
					// Dropping the attributes from config falls back to the
					// schema defaults rather than clearing them server-side.
					Config: testAccTFESAMLSettings_basic(tfe.AdminSAMLSetting{
						IDPCert:        idpCert,
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
		idpCert := testIDPCertBody(t)
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
						// Import takes the cert as TFE returns it: wrapped PEM from
						// 2.1.0, the raw body before that.
						got := rs.Attributes["idp_cert"]
						if got != idpCert && got != wrapPEM(t, idpCert, 64) {
							return fmt.Errorf("expected idp_cert attribute to be equal to %q or %q, received: %q", idpCert, wrapPEM(t, idpCert, 64), got)
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

	t.Run("idp_cert reformatting does not drift", func(t *testing.T) {
		// The backend always stores certs wrapped at 64, so feeding it any
		// other formatting must still leave state alone.
		body := testIDPCertBody(t)
		wrapped76 := wrapPEM(t, body, 76)

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			Steps: []resource.TestStep{
				{
					// Backend re-wraps to 64; state keeps our 76.
					Config: testAccTFESAMLSettings_idpCert(wrapped76),
					Check:  resource.TestCheckResourceAttr(testResourceName, "idp_cert", wrapped76),
				},
				{
					// Same cert without the armor isn't a change.
					Config: testAccTFESAMLSettings_idpCert(body),
					ConfigPlanChecks: resource.ConfigPlanChecks{
						PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
					},
					Check: resource.TestCheckResourceAttr(testResourceName, "idp_cert", wrapped76),
				},
			},
		})
	})

	t.Run("removing idp_cert from config keeps the certificate", func(t *testing.T) {
		client := testAccSkipBeforeSAMLIDPCertificates(t)
		idpCert := testIDPCertBody(t)
		// Destroy keeps the cert, so remove it for later tests.
		t.Cleanup(func() { testAccDeleteLegacyIDPCerts(t, client) })
		s := tfe.AdminSAMLSetting{
			IDPCert:        idpCert,
			SLOEndpointURL: "https://foobar.com/slo_endpoint_url",
			SSOEndpointURL: "https://foobar.com/sso_endpoint_url",
		}
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			CheckDestroy: resource.ComposeTestCheckFunc(
				testAccTFESAMLSettingsDestroy,
				testAccCheckSAMLIDPCertKept(idpCert),
			),
			Steps: []resource.TestStep{
				{
					Config: testAccTFESAMLSettings_basic(s),
					Check:  resource.TestCheckResourceAttr(testResourceName, "idp_cert", idpCert),
				},
				{
					// Dropping idp_cert from config isn't a change.
					Config: testAccTFESAMLSettings_noIDPCert("https://foobar.com/slo_endpoint_url"),
					ConfigPlanChecks: resource.ConfigPlanChecks{
						PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
					},
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(testResourceName, "idp_cert", idpCert),
						testAccCheckSAMLIDPCertKept(idpCert),
					),
				},
				{
					// Other updates leave the cert alone.
					Config: testAccTFESAMLSettings_noIDPCert("https://foobar-updated.com/slo_endpoint_url"),
					ConfigPlanChecks: resource.ConfigPlanChecks{
						PreApply: []plancheck.PlanCheck{
							plancheck.ExpectResourceAction(testResourceName, plancheck.ResourceActionUpdate),
							plancheck.ExpectUnknownValue(testResourceName, tfjsonpath.New("idp_cert")),
						},
					},
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(testResourceName, "slo_endpoint_url", "https://foobar-updated.com/slo_endpoint_url"),
						testAccCheckSAMLIDPCertKept(idpCert),
					),
				},
				{
					// State has TFE's PEM format; the same cert in another format is not a change.
					Config: testAccTFESAMLSettings_idpCertSLO(idpCert, "https://foobar-updated.com/slo_endpoint_url"),
					ConfigPlanChecks: resource.ConfigPlanChecks{
						PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
					},
					Check: testAccCheckSAMLIDPCertKept(idpCert),
				},
				{
					Config: testAccTFESAMLSettings_idpCertSLO(idpCert, "https://foobar-updated.com/slo_endpoint_url"),
					ConfigPlanChecks: resource.ConfigPlanChecks{
						PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
					},
				},
			},
		})
	})

	t.Run("unchanged idp_cert is not re-sent", func(t *testing.T) {
		client := testAccSkipBeforeSAMLIDPCertificates(t)
		idpCert := testIDPCertBody(t)
		// Without legacy_old, TFE rejects re-sending the current cert.
		t.Cleanup(func() { testAccDeleteLegacyIDPCerts(t, client) })
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			CheckDestroy:             testAccTFESAMLSettingsDestroy,
			Steps: []resource.TestStep{
				{
					PreConfig: func() { testAccDeleteLegacyIDPCerts(t, client) },
					Config:    testAccTFESAMLSettings_idpCertSLO(idpCert, "https://foobar.com/slo_endpoint_url"),
				},
				{
					// Update another attribute.
					Config: testAccTFESAMLSettings_idpCertSLO(idpCert, "https://foobar-updated.com/slo_endpoint_url"),
					Check:  resource.TestCheckResourceAttr(testResourceName, "slo_endpoint_url", "https://foobar-updated.com/slo_endpoint_url"),
				},
				{
					// Destroy keeps the cert.
					Config:  testAccTFESAMLSettings_idpCertSLO(idpCert, "https://foobar-updated.com/slo_endpoint_url"),
					Destroy: true,
				},
				{
					Config: testAccTFESAMLSettings_idpCertSLO(idpCert, "https://foobar-updated.com/slo_endpoint_url"),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(testResourceName, "enabled", "true"),
						testAccCheckSAMLIDPCertKept(idpCert),
					),
				},
			},
		})
	})

	t.Run("SAML settings with managed certificates and no idp_cert", func(t *testing.T) {
		client := testAccSkipBeforeSAMLIDPCertificates(t)
		certA := generateSelfSignedCertPEM(t)
		certB := generateSelfSignedCertPEM(t)

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			CheckDestroy: resource.ComposeTestCheckFunc(
				testAccTFESAMLSettingsDestroy,
				testAccTFESAMLIDPCertificateDestroy,
			),
			Steps: []resource.TestStep{
				{
					// Only managed certs should be trusted.
					PreConfig: func() { testAccDeleteLegacyIDPCerts(t, client) },
					Config:    testAccTFESAMLSettings_managedCerts(certA, certB),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(testResourceName, "enabled", "true"),
						resource.TestCheckResourceAttr(testResourceName, "idp_cert", ""),
						resource.TestCheckResourceAttr("tfe_saml_idp_certificate.primary", "cert_role", samlIDPCertRoleManaged),
						resource.TestCheckResourceAttr("tfe_saml_idp_certificate.failover", "cert_role", samlIDPCertRoleManaged),
					),
				},
			},
		})
	})

	t.Run("legacy certificate deleted in the same apply as an update", func(t *testing.T) {
		client := testAccSkipBeforeSAMLIDPCertificates(t)
		legacy := testIDPCertBody(t)
		managed := generateSelfSignedCertPEM(t)
		t.Cleanup(func() { testAccDeleteLegacyIDPCerts(t, client) })

		// legacy depends on the settings, so it's deleted before they're updated.
		withLegacy := testAccTFESAMLSettings_managedAndLegacy(managed, legacy, "https://foobar.com/slo_endpoint_url")
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			CheckDestroy: resource.ComposeTestCheckFunc(
				testAccTFESAMLSettingsDestroy,
				testAccTFESAMLIDPCertificateDestroy,
			),
			Steps: []resource.TestStep{
				{
					PreConfig: func() { testAccSeedLegacyIDPCert(t, client, legacy) },
					Config:    testAccTFESAMLSettings_managedOnly(managed, "https://foobar.com/slo_endpoint_url"),
					Check:     resource.TestCheckResourceAttrSet(testResourceName, "idp_cert"),
				},
				{
					Config:             withLegacy,
					ResourceName:       "tfe_saml_idp_certificate.legacy",
					ImportState:        true,
					ImportStatePersist: true,
					ImportStateIdFunc:  testAccTFESAMLIDPCertificateIDByRole(client, "legacy_primary"),
				},
				{
					Config: withLegacy,
				},
				{
					// No stale idp_cert when the legacy cert goes in the same apply.
					Config: testAccTFESAMLSettings_managedOnly(managed, "https://foobar-updated.com/slo_endpoint_url"),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(testResourceName, "slo_endpoint_url", "https://foobar-updated.com/slo_endpoint_url"),
						resource.TestCheckResourceAttr(testResourceName, "idp_cert", ""),
					),
				},
			},
		})
	})

	t.Run("older Terraform Enterprise requires idp_cert", func(t *testing.T) {
		testAccSkipFromSAMLIDPCertificates(t)
		idpCert := testIDPCertBody(t)
		missing := regexp.MustCompile(`idp_cert is required on Terraform Enterprise`)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			CheckDestroy:             testAccTFESAMLSettingsDestroy,
			Steps: []resource.TestStep{
				{
					// Missing idp_cert is a plan-time error.
					Config:      testAccTFESAMLSettings_noIDPCert("https://foobar.com/slo_endpoint_url"),
					PlanOnly:    true,
					ExpectError: missing,
				},
				{
					Config: testAccTFESAMLSettings_idpCert(idpCert),
				},
				{
					Config:      testAccTFESAMLSettings_noIDPCert("https://foobar-updated.com/slo_endpoint_url"),
					PlanOnly:    true,
					ExpectError: missing,
				},
			},
		})
	})

	t.Run("destroy disables SAML and keeps the certificate", func(t *testing.T) {
		client := testAccSkipBeforeSAMLIDPCertificates(t)
		idpCert := testIDPCertBody(t)
		// Destroy keeps the cert, so remove it for later tests.
		t.Cleanup(func() { testAccDeleteLegacyIDPCerts(t, client) })
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			CheckDestroy: resource.ComposeTestCheckFunc(
				testAccTFESAMLSettingsDestroy,
				testAccCheckSAMLIDPCertKept(idpCert),
			),
			Steps: []resource.TestStep{
				{
					Config: testAccTFESAMLSettings_idpCert(idpCert),
					Check:  resource.TestCheckResourceAttr(testResourceName, "enabled", "true"),
				},
			},
		})
	})

	t.Run("destroy clears idp_cert on older Terraform Enterprise", func(t *testing.T) {
		testAccSkipFromSAMLIDPCertificates(t)
		idpCert := testIDPCertBody(t)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			CheckDestroy: resource.ComposeTestCheckFunc(
				testAccTFESAMLSettingsDestroy,
				testAccCheckSAMLIDPCertCleared,
			),
			Steps: []resource.TestStep{
				{
					Config: testAccTFESAMLSettings_idpCert(idpCert),
					Check:  resource.TestCheckResourceAttr(testResourceName, "enabled", "true"),
				},
			},
		})
	})

	t.Run("empty idp_cert is rejected", func(t *testing.T) {
		blank := regexp.MustCompile(`Invalid idp_cert`)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			Steps: []resource.TestStep{
				{
					Config:      testAccTFESAMLSettings_idpCert(""),
					PlanOnly:    true,
					ExpectError: blank,
				},
				{
					Config:      testAccTFESAMLSettings_idpCert(" \n\t "),
					PlanOnly:    true,
					ExpectError: blank,
				},
				{
					Config:      testAccTFESAMLSettings_idpCert("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----"),
					PlanOnly:    true,
					ExpectError: blank,
				},
			},
		})
	})

	t.Run("enabling SAML without any certificate is rejected", func(t *testing.T) {
		client := testAccSkipBeforeSAMLIDPCertificates(t)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			CheckDestroy:             testAccTFESAMLSettingsDestroy,
			Steps: []resource.TestStep{
				{
					PreConfig:   func() { testAccDeleteAllIDPCerts(t, client) },
					Config:      testAccTFESAMLSettings_noIDPCert("https://foobar.com/slo_endpoint_url"),
					ExpectError: regexp.MustCompile(`Error creating SAML Settings`),
				},
			},
		})
	})
}

// testAccCheckSAMLIDPCertCleared checks destroy cleared idp_cert (before 2.1.0).
func testAccCheckSAMLIDPCertCleared(_ *terraform.State) error {
	s, err := testAccConfiguredClient.Client.Admin.Settings.SAML.Read(ctx)
	if err != nil {
		return fmt.Errorf("failed to read SAML Settings: %w", err)
	}
	if s.IDPCert != "" {
		return fmt.Errorf("expected idp_cert to be cleared on destroy, got %q", s.IDPCert)
	}
	return nil
}

// testAccCheckSAMLIDPCertKept checks the instance still has cert (2.1.0+).
func testAccCheckSAMLIDPCertKept(cert string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		s, err := testAccConfiguredClient.Client.Admin.Settings.SAML.Read(ctx)
		if err != nil {
			return fmt.Errorf("failed to read SAML Settings: %w", err)
		}
		same, diags := customtypes.NewPEMCertificateValue(s.IDPCert).
			StringSemanticEquals(ctx, customtypes.NewPEMCertificateValue(cert))
		if diags.HasError() {
			return fmt.Errorf("failed to compare idp_cert: %v", diags)
		}
		if !same {
			return fmt.Errorf("expected the IdP certificate to be kept on the instance, got %q", s.IDPCert)
		}
		return nil
	}
}

func testAccTFESAMLSettings_idpCertSLO(cert, slo string) string {
	return fmt.Sprintf(`
resource "tfe_saml_settings" "foobar" {
  idp_cert         = %q
  slo_endpoint_url = %q
  sso_endpoint_url = "https://foobar.com/sso_endpoint_url"
}`, cert, slo)
}

func testAccTFESAMLSettings_noIDPCert(slo string) string {
	return fmt.Sprintf(`
resource "tfe_saml_settings" "foobar" {
  slo_endpoint_url = %q
  sso_endpoint_url = "https://foobar.com/sso_endpoint_url"
}`, slo)
}

// testAccSkipFromSAMLIDPCertificates skips on TFE >= 2.1.0.
func testAccSkipFromSAMLIDPCertificates(t *testing.T) {
	t.Helper()
	client := testAccSAMLIDPCertClient(t)
	supported, err := client.MeetsMinRemoteTFEVersion(minTFEVersionSAMLIDPCertificates)
	if err != nil {
		t.Fatalf("failed to check Terraform Enterprise version: %v", err)
	}
	if supported {
		t.Skipf("requires Terraform Enterprise earlier than %s, got %s", minTFEVersionSAMLIDPCertificates, client.RemoteTFEVersion())
	}
}

// testAccSkipBeforeSAMLIDPCertificates skips on TFE < 2.1.0.
func testAccSkipBeforeSAMLIDPCertificates(t *testing.T) ConfiguredClient {
	t.Helper()
	client := testAccSAMLIDPCertClient(t)
	supported, err := client.MeetsMinRemoteTFEVersion(minTFEVersionSAMLIDPCertificates)
	if err != nil {
		t.Fatalf("failed to check Terraform Enterprise version: %v", err)
	}
	if !supported {
		t.Skipf("requires Terraform Enterprise %s or later, got %s", minTFEVersionSAMLIDPCertificates, client.RemoteTFEVersion())
	}
	return client
}

func testAccTFESAMLSettings_managedOnly(cert, slo string) string {
	return testAccTFESAMLIDPCertificate_named("managed", "fooidp-managed", cert) + fmt.Sprintf(`
resource "tfe_saml_settings" "foobar" {
  slo_endpoint_url = %q
  sso_endpoint_url = "https://foobar.com/sso_endpoint_url"

  depends_on = [tfe_saml_idp_certificate.managed]
}`, slo)
}

func testAccTFESAMLSettings_managedAndLegacy(managed, legacy, slo string) string {
	return testAccTFESAMLSettings_managedOnly(managed, slo) + fmt.Sprintf(`
resource "tfe_saml_idp_certificate" "legacy" {
  display_name = "fooidp-legacy"
  cert         = %q

  depends_on = [tfe_saml_settings.foobar]
}`, legacy)
}

func testAccTFESAMLSettings_managedCerts(certA, certB string) string {
	return testAccTFESAMLIDPCertificate_named("primary", "fooidp-primary", certA) +
		testAccTFESAMLIDPCertificate_named("failover", "fooidp-failover", certB) + `
resource "tfe_saml_settings" "foobar" {
  slo_endpoint_url = "https://foobar.com/slo_endpoint_url"
  sso_endpoint_url = "https://foobar.com/sso_endpoint_url"

  depends_on = [
    tfe_saml_idp_certificate.primary,
    tfe_saml_idp_certificate.failover,
  ]
}`
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
	// IDPCert depends on the release, see testAccCheckSAMLIDPCertKept and testAccCheckSAMLIDPCertCleared.
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

func testAccTFESAMLSettings_siteAuditor(idpCert, attrSiteAuditor, siteAuditorRole string) string {
	return fmt.Sprintf(`
resource "tfe_saml_settings" "foobar" {
  idp_cert          = %q
  slo_endpoint_url  = "https://foobar.com/slo_endpoint_url"
  sso_endpoint_url  = "https://foobar.com/sso_endpoint_url"
  attr_site_auditor = "%s"
  site_auditor_role = "%s"
}

data "tfe_saml_settings" "foobar" {
  depends_on = [tfe_saml_settings.foobar]
}`, idpCert, attrSiteAuditor, siteAuditorRole)
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

// TestSAMLSettingsSessionTimeoutRange pins that an out-of-range session timeout
// is rejected rather than clamped.
//
// Clamping looked safe but silently changed the practitioner's value. Terraform
// Enterprise validates this attribute for presence only, so it accepts and
// echoes back the clamped number — Terraform then fails the apply with
// "Provider produced inconsistent result after apply", because the plan still
// holds the original. The schema validator catches these at plan time; this
// guard also proves the narrowing in range for gosec.
func TestSAMLSettingsSessionTimeoutRange(t *testing.T) {
	for _, tc := range []struct {
		name    string
		in      int64
		want    int32
		wantErr bool
	}{
		{name: "the schema default is in range", in: samlDefaultSSOAPITokenSessionTimeoutSeconds, want: 1209600},
		{name: "zero is in range", in: 0, want: 0},
		{name: "the upper bound is in range", in: math.MaxInt32, want: math.MaxInt32},
		{name: "the lower bound is in range", in: math.MinInt32, want: math.MinInt32},
		{name: "above int32 is rejected, not clamped", in: math.MaxInt32 + 1, wantErr: true},
		{name: "below int32 is rejected, not clamped", in: math.MinInt32 - 1, wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := int32SessionTimeout(tc.in)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("int32SessionTimeout(%d) = %d with no error; an out-of-range value must be rejected, since clamping it would break the apply", tc.in, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("int32SessionTimeout(%d): %v", tc.in, err)
			}
			if got != tc.want {
				t.Errorf("int32SessionTimeout(%d) = %d, want %d", tc.in, got, tc.want)
			}
		})
	}
}

// The same certificate written differently must compare equal, so the
// configured formatting survives in state and plans converge.
func TestSAMLSettingsIDPCertSemanticEquality(t *testing.T) {
	armored := generateSelfSignedCertPEM(t)
	other := generateSelfSignedCertPEM(t)

	// Built with plain string ops, not the code under test.
	lines := strings.Split(strings.TrimSpace(armored), "\n")
	base64Only := strings.Join(lines[1:len(lines)-1], "")
	oneLine := strings.ReplaceAll(armored, "\n", "")

	for _, tc := range []struct {
		name   string
		server string
		config string
		want   bool
	}{
		{"same cert wrapped at 76 instead of 64", armored, wrapPEM(t, base64Only, 76), true},
		{"same cert on a single line", armored, oneLine, true},
		{"config omits the PEM armor", armored, base64Only, true},
		{"a genuinely different certificate", armored, other, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// The framework invokes this on the server value with the plan
			// value as the argument; null and unknown are short-circuited
			// before it is reached.
			got, diags := customtypes.NewPEMCertificateValue(tc.server).
				StringSemanticEquals(t.Context(), customtypes.NewPEMCertificateValue(tc.config))
			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags)
			}
			if got != tc.want {
				t.Errorf("semantic equality = %v, want %v", got, tc.want)
			}
		})
	}
}

func testAccTFESAMLSettings_idpCert(cert string) string {
	return fmt.Sprintf(`
resource "tfe_saml_settings" "foobar" {
  idp_cert         = %q
  slo_endpoint_url = "https://foobar.com/slo_endpoint_url"
  sso_endpoint_url = "https://foobar.com/sso_endpoint_url"
}`, cert)
}

func TestSAMLSettingsUpdateAttrsIDPCert(t *testing.T) {
	m := modelTFESAMLSettings{
		IDPCert:                   customtypes.NewPEMCertificateValue("cert-body"),
		SSOAPITokenSessionTimeout: types.Int64Value(samlDefaultSSOAPITokenSessionTimeoutSeconds),
	}

	attrs, err := samlSettingsUpdateAttrs(m, false, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := attrs.GetIdpCert(); got != nil {
		t.Errorf("expected idp_cert to be omitted when not configured, got %q", *got)
	}

	attrs, err = samlSettingsUpdateAttrs(m, false, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := attrs.GetIdpCert(); got == nil || *got != "cert-body" {
		t.Errorf("expected idp_cert %q to be sent when configured, got %v", "cert-body", got)
	}
}

func TestPEMBodyNotEmptyValidator(t *testing.T) {
	cases := map[string]struct {
		value   customtypes.PEMCertificateValue
		wantErr bool
	}{
		"null":       {customtypes.NewPEMCertificateNull(), false},
		"body":       {customtypes.NewPEMCertificateValue("MIIBbody"), false},
		"empty":      {customtypes.NewPEMCertificateValue(""), true},
		"whitespace": {customtypes.NewPEMCertificateValue(" \n\t "), true},
		"armor only": {customtypes.NewPEMCertificateValue("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----"), true},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			resp := &validator.StringResponse{}
			pemBodyNotEmptyValidator{}.ValidateString(ctx, validator.StringRequest{ConfigValue: tc.value.StringValue}, resp)
			if got := resp.Diagnostics.HasError(); got != tc.wantErr {
				t.Fatalf("expected error %t, got %v", tc.wantErr, resp.Diagnostics)
			}
		})
	}
}

func TestPEMKeepStateIfEquivalent(t *testing.T) {
	body := "MIIBbodyAAAA"
	wrapped := "-----BEGIN CERTIFICATE-----\n" + body + "\n-----END CERTIFICATE-----\n"
	cases := map[string]struct {
		config, state types.String
		want          types.String
	}{
		"equivalent keeps state": {types.StringValue(body), types.StringValue(wrapped), types.StringValue(wrapped)},
		"different keeps plan":   {types.StringValue("MIIBother"), types.StringValue(wrapped), types.StringValue("MIIBother")},
		"null config":            {types.StringNull(), types.StringValue(wrapped), types.StringNull()},
		"null state":             {types.StringValue(body), types.StringNull(), types.StringValue(body)},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			resp := &planmodifier.StringResponse{PlanValue: tc.config}
			pemKeepStateIfEquivalent{}.PlanModifyString(ctx, planmodifier.StringRequest{ConfigValue: tc.config, StateValue: tc.state, PlanValue: tc.config}, resp)
			if !resp.PlanValue.Equal(tc.want) {
				t.Fatalf("expected %v, got %v", tc.want, resp.PlanValue)
			}
		})
	}
}
