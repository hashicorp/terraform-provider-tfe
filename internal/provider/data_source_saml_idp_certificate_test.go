// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"

	tfev2 "github.com/hashicorp/go-tfe/v2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// FLAKE ALERT: IdP certificates belong to the SAML settings singleton shared by
// the entire TFE instance. All cases must stay in this single func, use t.Run,
// and never call t.Parallel. They also must not run concurrently with
// TestAccTFESAMLSettings_omnibus, TestAccTFESAMLIDPCertificate_omnibus or
// TestAccTFESAMLIDPCertificatesDataSource_omnibus.
//
// Skipped in CI via the "TestAccTFESAML" skip pattern in ci.yml; runs in the
// TFE Nightly workflow.
func TestAccTFESAMLIDPCertificateDataSource_omnibus(t *testing.T) {
	skipIfCloud(t)

	client := testAccSAMLIDPCertClient(t)
	supported, err := client.MeetsMinRemoteTFEVersion(minTFEVersionSAMLIDPCertificates)
	if err != nil {
		t.Fatalf("failed to check Terraform Enterprise version: %v", err)
	}

	// Validators run before any API call, so this case works on every version.
	t.Run("invalid id is rejected", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			Steps: []resource.TestStep{
				{
					Config:      testAccTFESAMLIDPCertificateDataSource_byID(`"not-an-idpc-id"`),
					ExpectError: regexp.MustCompile(`must be a SAML IdP certificate ID`),
				},
				{
					Config:      testAccTFESAMLIDPCertificateDataSource_byID(`"idpc-"`),
					ExpectError: regexp.MustCompile(`must be a SAML IdP certificate ID`),
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
					Config:      testAccTFESAMLIDPCertificateDataSource_byID(`"idpc-doesnotmatter"`),
					ExpectError: regexp.MustCompile(`Terraform Enterprise version does not support SAML IdP\s+certificates`),
				},
			},
		})
	})

	if !supported {
		t.Skipf("remaining cases require Terraform Enterprise %s or later, got %s", minTFEVersionSAMLIDPCertificates, client.RemoteTFEVersion())
	}

	t.Run("reads managed and legacy certificates", func(t *testing.T) {
		legacyCert := testIDPCertBody(t)
		cert := generateSelfSignedCertPEM(t)
		t.Cleanup(func() { testAccDeleteLegacyIDPCerts(t, client) })

		// Start from a known set: no certificates except one legacy_primary.
		testAccDeleteAllIDPCerts(t, client)
		testAccSetIDPCert(t, client, legacyCert)
		legacyID, err := testAccTFESAMLIDPCertificateIDByRole(client, "legacy_primary")(nil)
		if err != nil {
			t.Fatal(err)
		}

		const (
			managed = "data.tfe_saml_idp_certificate.managed"
			legacy  = "data.tfe_saml_idp_certificate.legacy"
		)

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			CheckDestroy:             testAccTFESAMLIDPCertificateDataSourcesDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccTFESAMLIDPCertificateDataSource_full(cert, legacyID),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrPair(managed, "id", "tfe_saml_idp_certificate.a", "id"),
						resource.TestCheckResourceAttr(managed, "display_name", "Fooidp-US-East"),
						resource.TestCheckResourceAttr(managed, "cert_role", "managed"),
						resource.TestCheckResourceAttrPair(managed, "fingerprint", "tfe_saml_idp_certificate.a", "fingerprint"),
						resource.TestCheckResourceAttrPair(managed, "expires_at", "tfe_saml_idp_certificate.a", "expires_at"),
						resource.TestCheckResourceAttrPair(managed, "created_at", "tfe_saml_idp_certificate.a", "created_at"),
						resource.TestCheckResourceAttrPair(managed, "issuer", "tfe_saml_idp_certificate.a", "issuer"),
						resource.TestCheckResourceAttrSet(managed, "cert"),
						resource.TestCheckResourceAttr(legacy, "id", legacyID),
						resource.TestCheckResourceAttr(legacy, "cert_role", "legacy_primary"),
						resource.TestCheckResourceAttrSet(legacy, "fingerprint"),
					),
				},
			},
		})
	})

	// Kept separate from the case above: a step that fails at plan time leaves
	// earlier resources in state, and the post-test destroy would then re-run
	// the failing lookup and leave them dangling.
	t.Run("errors on an unknown certificate ID", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			Steps: []resource.TestStep{
				{
					Config:      testAccTFESAMLIDPCertificateDataSource_byID(`"idpc-doesnotexist"`),
					ExpectError: regexp.MustCompile(`SAML IdP certificate not found`),
				},
			},
		})
	})
}

func testAccTFESAMLIDPCertificateDataSource_byID(id string) string {
	return fmt.Sprintf(`
data "tfe_saml_idp_certificate" "foobar" {
  id = %s
}`, id)
}

func testAccTFESAMLIDPCertificateDataSource_full(cert, legacyID string) string {
	return testAccTFESAMLIDPCertificate_named("a", "Fooidp-US-East", cert) + fmt.Sprintf(`
data "tfe_saml_idp_certificate" "managed" {
  id = tfe_saml_idp_certificate.a.id
}

data "tfe_saml_idp_certificate" "legacy" {
  id = %q
}`, legacyID)
}

// testAccTFESAMLIDPCertificateDataSourcesDestroy checks that the managed
// certificates are gone. Data sources share the tfe_saml_idp_certificate type
// name and can point at certificates the test doesn't own, like the seeded
// legacy one, so they are skipped. Shared by both data source tests.
func testAccTFESAMLIDPCertificateDataSourcesDestroy(s *terraform.State) error {
	for name, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_saml_idp_certificate" || strings.HasPrefix(name, "data.") {
			continue
		}
		_, err := testAccConfiguredClient.ClientV2.API.Admin().SamlSettings().IdpCertificates().ByExternal_id(rs.Primary.ID).Get(ctx, nil)
		if err == nil {
			return fmt.Errorf("SAML IdP certificate %s still exists", rs.Primary.ID)
		}
		if !errors.Is(err, tfev2.ErrNotFound) {
			return fmt.Errorf("failed to read SAML IdP certificate %s: %w", rs.Primary.ID, err)
		}
	}
	return nil
}
