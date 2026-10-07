// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"reflect"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// FLAKE ALERT: IdP certificates belong to the SAML settings singleton shared by
// the entire TFE instance. All cases must stay in this single func, use t.Run,
// and never call t.Parallel. They also must not run concurrently with
// TestAccTFESAMLSettings_omnibus, TestAccTFESAMLIDPCertificate_omnibus or
// TestAccTFESAMLIDPCertificateDataSource_omnibus.
//
// Skipped in CI via the "TestAccTFESAML" skip pattern in ci.yml; runs in the
// TFE Nightly workflow.
func TestAccTFESAMLIDPCertificatesDataSource_omnibus(t *testing.T) {
	skipIfCloud(t)

	client := testAccSAMLIDPCertClient(t)
	supported, err := client.MeetsMinRemoteTFEVersion(minTFEVersionSAMLIDPCertificates)
	if err != nil {
		t.Fatalf("failed to check Terraform Enterprise version: %v", err)
	}

	// Validators run before any API call, so this case works on every version.
	t.Run("empty display_name_match is rejected", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			Steps: []resource.TestStep{
				{
					Config: `
data "tfe_saml_idp_certificates" "foobar" {
  display_name_match = ""
}`,
					ExpectError: regexp.MustCompile(`string length must be at least 1`),
				},
			},
		})
	})

	t.Run("older Terraform Enterprise returns a minimum-version error", func(t *testing.T) {
		if supported {
			t.Skipf("Terraform Enterprise %s supports SAML IdP certificates", client.RemoteTFEVersion())
		}
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			Steps: []resource.TestStep{
				{
					Config:      `data "tfe_saml_idp_certificates" "foobar" {}`,
					ExpectError: regexp.MustCompile(`Terraform Enterprise version does not support SAML IdP\s+certificates`),
				},
			},
		})
	})

	if !supported {
		t.Skipf("remaining cases require Terraform Enterprise %s or later, got %s", minTFEVersionSAMLIDPCertificates, client.RemoteTFEVersion())
	}

	t.Run("lists and filters managed and legacy certificates", func(t *testing.T) {
		legacyCert := testIDPCertBody(t)
		certA := generateSelfSignedCertPEM(t)
		certB := generateSelfSignedCertPEM(t)
		t.Cleanup(func() { testAccDeleteLegacyIDPCerts(t, client) })

		// Start from a known set: no certificates except one legacy_primary.
		testAccDeleteAllIDPCerts(t, client)
		testAccSetIDPCert(t, client, legacyCert)
		legacyID, err := testAccTFESAMLIDPCertificateIDByRole(client, "legacy_primary")(nil)
		if err != nil {
			t.Fatal(err)
		}

		const (
			all     = "data.tfe_saml_idp_certificates.all"
			matched = "data.tfe_saml_idp_certificates.matched"
			none    = "data.tfe_saml_idp_certificates.none"
		)

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			CheckDestroy:             testAccTFESAMLIDPCertificateDataSourcesDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccTFESAMLIDPCertificatesDataSource_full(certA, certB),
					Check: resource.ComposeTestCheckFunc(
						// No filter returns everything, legacy included.
						resource.TestCheckResourceAttr(all, "id", "saml"),
						resource.TestCheckResourceAttr(all, "certificates.#", "3"),
						resource.TestCheckTypeSetElemNestedAttrs(all, "certificates.*", map[string]string{"id": legacyID, "cert_role": "legacy_primary"}),
						resource.TestCheckTypeSetElemNestedAttrs(all, "certificates.*", map[string]string{"display_name": "Fooidp-US-East", "cert_role": "managed"}),
						resource.TestCheckTypeSetElemNestedAttrs(all, "certificates.*", map[string]string{"display_name": "fooidp-eu-west", "cert_role": "managed"}),
						// Case-insensitive substring match.
						resource.TestCheckResourceAttr(matched, "id", "saml"),
						resource.TestCheckResourceAttr(matched, "certificates.#", "1"),
						resource.TestCheckResourceAttrPair(matched, "certificates.0.id", "tfe_saml_idp_certificate.a", "id"),
						resource.TestCheckResourceAttr(matched, "certificates.0.display_name", "Fooidp-US-East"),
						// No match returns an empty list, not an error.
						resource.TestCheckResourceAttr(none, "certificates.#", "0"),
					),
				},
			},
		})
	})
}

func testAccTFESAMLIDPCertificatesDataSource_full(certA, certB string) string {
	return testAccTFESAMLIDPCertificate_named("a", "Fooidp-US-East", certA) +
		testAccTFESAMLIDPCertificate_named("b", "fooidp-eu-west", certB) + `
data "tfe_saml_idp_certificates" "all" {
  depends_on = [tfe_saml_idp_certificate.a, tfe_saml_idp_certificate.b]
}

data "tfe_saml_idp_certificates" "matched" {
  display_name_match = "us-EAST"
  depends_on         = [tfe_saml_idp_certificate.a, tfe_saml_idp_certificate.b]
}

data "tfe_saml_idp_certificates" "none" {
  display_name_match = "no-such-certificate"
  depends_on         = [tfe_saml_idp_certificate.a, tfe_saml_idp_certificate.b]
}`
}

func TestFilterSAMLIDPCertificatesByDisplayName(t *testing.T) {
	certs := []modelTFESAMLIDPCertificate{
		{DisplayName: types.StringValue("Fooidp-US-East")},
		{DisplayName: types.StringValue("fooidp-eu-west")},
		{DisplayName: types.StringValue("legacy")},
	}
	for _, tc := range []struct {
		match string
		want  []string
	}{
		{"", []string{"Fooidp-US-East", "fooidp-eu-west", "legacy"}},
		{"fooidp", []string{"Fooidp-US-East", "fooidp-eu-west"}},
		{"US-east", []string{"Fooidp-US-East"}},
		{"EU", []string{"fooidp-eu-west"}},
		{" ", []string{}},
		{"nothing", []string{}},
	} {
		got := filterSAMLIDPCertificatesByDisplayName(certs, tc.match)
		if got == nil {
			t.Errorf("match %q: expected an empty list, got nil", tc.match)
			continue
		}
		names := make([]string, 0, len(got))
		for _, c := range got {
			names = append(names, c.DisplayName.ValueString())
		}
		if !reflect.DeepEqual(names, tc.want) {
			t.Errorf("match %q: expected %v, got %v", tc.match, tc.want, names)
		}
	}
}
