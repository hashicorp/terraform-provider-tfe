// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"errors"
	"fmt"
	"regexp"
	"testing"

	tfev2 "github.com/hashicorp/go-tfe/v2"
	"github.com/hashicorp/go-tfe/v2/api/models"
	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-tfe/internal/provider/customtypes"
)

const testIDPCertResourceName = "tfe_saml_idp_certificate.foobar"

// FLAKE ALERT: IdP certificates belong to the SAML settings singleton shared by
// the entire TFE instance. All cases must stay in this single func, use t.Run,
// and never call t.Parallel. They also must not run concurrently with
// TestAccTFESAMLSettings_omnibus.
//
// Skipped in CI via the "TestAccTFESAML" skip pattern in ci.yml; runs in the
// TFE Nightly workflow.
//
// TFE only allows deleting the last trusted certificate while SAML is
// disabled, so these cases run with SAML disabled.
func TestAccTFESAMLIDPCertificate_omnibus(t *testing.T) {
	skipIfCloud(t)

	client := testAccSAMLIDPCertClient(t)
	supported, err := client.MeetsMinRemoteTFEVersion(minTFEVersionSAMLIDPCertificates)
	if err != nil {
		t.Fatalf("failed to check Terraform Enterprise version: %v", err)
	}

	t.Run("older Terraform Enterprise returns a minimum-version error", func(t *testing.T) {
		if supported {
			t.Skipf("Terraform Enterprise %s supports SAML IdP certificates", client.RemoteTFEVersion())
		}
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			Steps: []resource.TestStep{
				{
					Config:      testAccTFESAMLIDPCertificate_basic("too-old", generateSelfSignedCertPEM(t)),
					ExpectError: regexp.MustCompile(`Terraform Enterprise version does not support SAML IdP\s+certificates`),
				},
			},
		})
	})

	if !supported {
		t.Skipf("remaining cases require Terraform Enterprise %s or later, got %s", minTFEVersionSAMLIDPCertificates, client.RemoteTFEVersion())
	}

	t.Run("basic, update and import", func(t *testing.T) {
		cert := generateSelfSignedCertPEM(t)
		renewed := generateSelfSignedCertPEM(t)
		sameID := statecheck.CompareValue(compare.ValuesSame())
		sameCreatedAt := statecheck.CompareValue(compare.ValuesSame())
		newFingerprint := statecheck.CompareValue(compare.ValuesDiffer())
		var certID string

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			CheckDestroy:             testAccTFESAMLIDPCertificateDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccTFESAMLIDPCertificate_basic("fooidp-us-east", cert),
					Check: resource.ComposeTestCheckFunc(
						resource.TestMatchResourceAttr(testIDPCertResourceName, "id", regexp.MustCompile(`^idpc-`)),
						resource.TestCheckResourceAttr(testIDPCertResourceName, "display_name", "fooidp-us-east"),
						resource.TestCheckResourceAttr(testIDPCertResourceName, "cert", cert),
						resource.TestCheckResourceAttr(testIDPCertResourceName, "cert_role", "managed"),
						resource.TestCheckResourceAttrSet(testIDPCertResourceName, "fingerprint"),
						resource.TestCheckResourceAttrSet(testIDPCertResourceName, "expires_at"),
						resource.TestCheckResourceAttrSet(testIDPCertResourceName, "created_at"),
					),
					ConfigStateChecks: []statecheck.StateCheck{
						sameID.AddStateValue(testIDPCertResourceName, tfjsonpath.New("id")),
						sameCreatedAt.AddStateValue(testIDPCertResourceName, tfjsonpath.New("created_at")),
					},
				},
				{
					// Rename in place.
					Config: testAccTFESAMLIDPCertificate_basic("fooidp-us-east-renamed", cert),
					ConfigPlanChecks: resource.ConfigPlanChecks{
						PreApply: []plancheck.PlanCheck{
							plancheck.ExpectResourceAction(testIDPCertResourceName, plancheck.ResourceActionUpdate),
						},
					},
					Check: resource.TestCheckResourceAttr(testIDPCertResourceName, "display_name", "fooidp-us-east-renamed"),
					ConfigStateChecks: []statecheck.StateCheck{
						newFingerprint.AddStateValue(testIDPCertResourceName, tfjsonpath.New("fingerprint")),
					},
				},
				{
					// Rename and replace the body in one apply: same certificate
					// object, new derived attributes.
					Config: testAccTFESAMLIDPCertificate_basic("fooidp-us-east-rotated", renewed),
					ConfigPlanChecks: resource.ConfigPlanChecks{
						PreApply: []plancheck.PlanCheck{
							plancheck.ExpectResourceAction(testIDPCertResourceName, plancheck.ResourceActionUpdate),
							plancheck.ExpectUnknownValue(testIDPCertResourceName, tfjsonpath.New("fingerprint")),
						},
					},
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(testIDPCertResourceName, "display_name", "fooidp-us-east-rotated"),
						resource.TestCheckResourceAttr(testIDPCertResourceName, "cert", renewed),
						resource.TestCheckResourceAttr(testIDPCertResourceName, "cert_role", "managed"),
						resource.TestCheckResourceAttrWith(testIDPCertResourceName, "id", func(v string) error {
							certID = v
							return nil
						}),
					),
					ConfigStateChecks: []statecheck.StateCheck{
						sameID.AddStateValue(testIDPCertResourceName, tfjsonpath.New("id")),
						sameCreatedAt.AddStateValue(testIDPCertResourceName, tfjsonpath.New("created_at")),
						newFingerprint.AddStateValue(testIDPCertResourceName, tfjsonpath.New("fingerprint")),
					},
				},
				{
					ResourceName:      testIDPCertResourceName,
					ImportState:       true,
					ImportStateVerify: true,
					// The backend re-wraps the body; semantic equality is
					// covered by the reformatting case.
					ImportStateVerifyIgnore: []string{"cert"},
				},
				{
					ResourceName:  testIDPCertResourceName,
					ImportState:   true,
					ImportStateId: "not-a-certificate-id",
					ExpectError:   regexp.MustCompile(`Invalid SAML IdP certificate ID`),
				},
				{
					// Deleted outside Terraform: Read drops it and the plan recreates it.
					PreConfig: func() {
						if err := client.ClientV2.API.Admin().SamlSettings().IdpCertificates().ByExternal_id(certID).Delete(ctx, nil); err != nil {
							t.Fatalf("failed to delete SAML IdP certificate %s out of band: %v", certID, apiErrorDetail(err))
						}
					},
					Config: testAccTFESAMLIDPCertificate_basic("fooidp-us-east-rotated", renewed),
					ConfigPlanChecks: resource.ConfigPlanChecks{
						PreApply: []plancheck.PlanCheck{
							plancheck.ExpectResourceAction(testIDPCertResourceName, plancheck.ResourceActionCreate),
						},
					},
				},
			},
		})
	})

	t.Run("multiple certificates coexist but bodies must be unique", func(t *testing.T) {
		certA := generateSelfSignedCertPEM(t)
		certB := generateSelfSignedCertPEM(t)

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			CheckDestroy:             testAccTFESAMLIDPCertificateDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccTFESAMLIDPCertificate_pair(certA, certB),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("tfe_saml_idp_certificate.a", "display_name", "fooidp-a"),
						resource.TestCheckResourceAttr("tfe_saml_idp_certificate.a", "cert", certA),
						resource.TestCheckResourceAttr("tfe_saml_idp_certificate.b", "display_name", "fooidp-b"),
						resource.TestCheckResourceAttr("tfe_saml_idp_certificate.b", "cert", certB),
						testAccCheckResourceAttrsDiffer("tfe_saml_idp_certificate.a", "tfe_saml_idp_certificate.b", "id"),
						testAccCheckResourceAttrsDiffer("tfe_saml_idp_certificate.a", "tfe_saml_idp_certificate.b", "fingerprint"),
					),
				},
				{
					Config:      testAccTFESAMLIDPCertificate_pair(certA, certB) + testAccTFESAMLIDPCertificate_named("dupe", "duplicate", certA),
					ExpectError: regexp.MustCompile(`Error creating SAML IdP certificate`),
				},
			},
		})
	})

	t.Run("invalid certificate is rejected", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			CheckDestroy:             testAccTFESAMLIDPCertificateDestroy,
			Steps: []resource.TestStep{
				{
					Config:      testAccTFESAMLIDPCertificate_basic("garbage", "not-a-certificate"),
					ExpectError: regexp.MustCompile(`Error creating SAML IdP certificate`),
				},
			},
		})
	})

	t.Run("last certificate cannot be deleted while SAML is enabled", func(t *testing.T) {
		cert := generateSelfSignedCertPEM(t)
		t.Cleanup(func() { testAccSetSAMLEnabled(t, client, false) })

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			CheckDestroy:             testAccTFESAMLIDPCertificateDestroy,
			Steps: []resource.TestStep{
				{
					PreConfig: func() { testAccDeleteAllIDPCerts(t, client) },
					Config:    testAccTFESAMLIDPCertificate_basic("only", cert),
				},
				{
					PreConfig:   func() { testAccSetSAMLEnabled(t, client, true) },
					Config:      testAccTFESAMLIDPCertificate_empty(),
					ExpectError: regexp.MustCompile(`Error deleting SAML IdP certificate`),
				},
				{
					// Once SAML is disabled the same delete succeeds.
					PreConfig: func() { testAccSetSAMLEnabled(t, client, false) },
					Config:    testAccTFESAMLIDPCertificate_empty(),
				},
			},
		})
	})

	t.Run("reformatting does not drift", func(t *testing.T) {
		body := testIDPCertBody(t)
		wrapped76 := wrapPEM(t, body, 76)
		sameID := statecheck.CompareValue(compare.ValuesSame())
		sameFingerprint := statecheck.CompareValue(compare.ValuesSame())

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			CheckDestroy:             testAccTFESAMLIDPCertificateDestroy,
			Steps: []resource.TestStep{
				{
					// Backend re-wraps to 64; state keeps our 76 and the
					// post-apply plan is empty.
					Config: testAccTFESAMLIDPCertificate_basic("reformat", wrapped76),
					Check:  resource.TestCheckResourceAttr(testIDPCertResourceName, "cert", wrapped76),
					ConfigStateChecks: []statecheck.StateCheck{
						sameID.AddStateValue(testIDPCertResourceName, tfjsonpath.New("id")),
						sameFingerprint.AddStateValue(testIDPCertResourceName, tfjsonpath.New("fingerprint")),
					},
				},
				{
					// Same body without armor. The config change plans an
					// update, but it is the same certificate: the ID and
					// fingerprint must not change.
					Config: testAccTFESAMLIDPCertificate_basic("reformat", body),
					Check:  resource.TestCheckResourceAttr(testIDPCertResourceName, "cert", body),
					ConfigStateChecks: []statecheck.StateCheck{
						sameID.AddStateValue(testIDPCertResourceName, tfjsonpath.New("id")),
						sameFingerprint.AddStateValue(testIDPCertResourceName, tfjsonpath.New("fingerprint")),
					},
				},
			},
		})
	})

	t.Run("import legacy certificate", func(t *testing.T) {
		idpCert := testIDPCertBody(t)
		other := generateSelfSignedCertPEM(t)
		sameFingerprint := statecheck.CompareValue(compare.ValuesSame())
		t.Cleanup(func() { testAccDeleteLegacyIDPCerts(t, client) })

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			CheckDestroy:             testAccTFESAMLIDPCertificateDestroy,
			Steps: []resource.TestStep{
				{
					// Setting idp_cert creates the legacy_primary certificate.
					// SAML stays disabled so the certificate can be deleted on
					// destroy without tripping the last-certificate guard.
					PreConfig:          func() { testAccSeedLegacyIDPCert(t, client, idpCert) },
					Config:             testAccTFESAMLIDPCertificate_basic("legacy", idpCert),
					ResourceName:       testIDPCertResourceName,
					ImportState:        true,
					ImportStatePersist: true,
					ImportStateIdFunc:  testAccTFESAMLIDPCertificateIDByRole(client, "legacy_primary"),
					ImportStateCheck:   testAccCheckImportedCertRole("legacy_primary"),
				},
				{
					// Renaming a legacy certificate is allowed.
					Config: testAccTFESAMLIDPCertificate_basic("legacy-renamed", idpCert),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(testIDPCertResourceName, "display_name", "legacy-renamed"),
						resource.TestCheckResourceAttr(testIDPCertResourceName, "cert_role", "legacy_primary"),
					),
					ConfigStateChecks: []statecheck.StateCheck{
						sameFingerprint.AddStateValue(testIDPCertResourceName, tfjsonpath.New("fingerprint")),
					},
				},
				{
					// Re-wrapping the same body is not a body change, so the
					// legacy guard must not fire.
					Config: testAccTFESAMLIDPCertificate_basic("legacy-renamed", wrapPEM(t, idpCert, 76)),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(testIDPCertResourceName, "cert", wrapPEM(t, idpCert, 76)),
						resource.TestCheckResourceAttr(testIDPCertResourceName, "cert_role", "legacy_primary"),
					),
					ConfigStateChecks: []statecheck.StateCheck{
						sameFingerprint.AddStateValue(testIDPCertResourceName, tfjsonpath.New("fingerprint")),
					},
				},
				{
					// Replacing its body is rejected at plan time.
					Config:      testAccTFESAMLIDPCertificate_basic("legacy-renamed", other),
					ExpectError: regexp.MustCompile(`Cannot change the body of a legacy SAML IdP certificate`),
				},
			},
		})
	})
}

// testAccSAMLIDPCertClient builds a client from the environment, since
// testAccConfiguredClient is only set once a test step configures the provider.
func testAccSAMLIDPCertClient(t *testing.T) ConfiguredClient {
	t.Helper()
	pc, err := getProviderClientUsingEnv()
	if err != nil {
		t.Fatalf("failed to build client: %v", err)
	}
	return ConfiguredClient{Client: pc.TfeClient, ClientV2: pc.TFEClientV2}
}

func testAccSetSAMLEnabled(t *testing.T, c ConfiguredClient, enabled bool) {
	t.Helper()
	env := samlSettingsEnvelopeForTest(func(a *models.AdminSamlSettings_attributes) {
		a.SetEnabled(ptr(enabled))
		if enabled {
			a.SetSsoEndpointUrl(ptr("https://foobar.com/sso"))
			a.SetSloEndpointUrl(ptr("https://foobar.com/slo"))
		}
	})
	if _, err := c.ClientV2.API.Admin().SamlSettings().Patch(ctx, env, nil); err != nil {
		t.Fatalf("failed to set SAML enabled=%t: %v", enabled, apiErrorDetail(err))
	}
}

func testAccSetIDPCert(t *testing.T, c ConfiguredClient, cert string) {
	t.Helper()
	env := samlSettingsEnvelopeForTest(func(a *models.AdminSamlSettings_attributes) {
		a.SetIdpCert(ptr(cert))
	})
	if _, err := c.ClientV2.API.Admin().SamlSettings().Patch(ctx, env, nil); err != nil {
		t.Fatalf("failed to set idp_cert: %v", apiErrorDetail(err))
	}
}

// testAccDeleteIDPCerts disables SAML and deletes every certificate matching keep == false.
func testAccDeleteIDPCerts(t *testing.T, c ConfiguredClient, keep func(role string) bool) {
	t.Helper()
	testAccSetSAMLEnabled(t, c, false)

	certs := c.ClientV2.API.Admin().SamlSettings().IdpCertificates()
	list, err := certs.Get(ctx, nil)
	if err != nil {
		t.Fatalf("failed to list SAML IdP certificates: %v", apiErrorDetail(err))
	}
	for _, cert := range list.GetData() {
		if cert.GetAttributes() == nil || keep(enumStringOrEmpty(cert.GetAttributes().GetCertRole())) {
			continue
		}
		id := valueOrZero(cert.GetId())
		if err := certs.ByExternal_id(id).Delete(ctx, nil); err != nil && !errors.Is(err, tfev2.ErrNotFound) {
			t.Fatalf("failed to delete SAML IdP certificate %s: %v", id, apiErrorDetail(err))
		}
	}
}

func testAccDeleteLegacyIDPCerts(t *testing.T, c ConfiguredClient) {
	t.Helper()
	testAccDeleteIDPCerts(t, c, func(role string) bool { return role == samlIDPCertRoleManaged })
}

func testAccDeleteAllIDPCerts(t *testing.T, c ConfiguredClient) {
	t.Helper()
	testAccDeleteIDPCerts(t, c, func(string) bool { return false })
}

// testAccSeedLegacyIDPCert clears leftover legacy certificates, then sets
// idp_cert, which creates a legacy_primary certificate.
func testAccSeedLegacyIDPCert(t *testing.T, c ConfiguredClient, cert string) {
	t.Helper()
	testAccDeleteLegacyIDPCerts(t, c)
	testAccSetIDPCert(t, c, cert)
}

func testAccCheckImportedCertRole(role string) resource.ImportStateCheckFunc {
	return func(s []*terraform.InstanceState) error {
		if len(s) != 1 {
			return fmt.Errorf("expected 1 imported instance, got %d", len(s))
		}
		if got := s[0].Attributes["cert_role"]; got != role {
			return fmt.Errorf("expected cert_role %s, got %q", role, got)
		}
		return nil
	}
}

func testAccCheckResourceAttrsDiffer(a, b, attr string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ra, ok := s.RootModule().Resources[a]
		if !ok {
			return fmt.Errorf("resource %s not found", a)
		}
		rb, ok := s.RootModule().Resources[b]
		if !ok {
			return fmt.Errorf("resource %s not found", b)
		}
		if va, vb := ra.Primary.Attributes[attr], rb.Primary.Attributes[attr]; va == vb {
			return fmt.Errorf("expected %s to differ between %s and %s, both are %q", attr, a, b, va)
		}
		return nil
	}
}

func testAccTFESAMLIDPCertificateDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_saml_idp_certificate" {
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

func testAccTFESAMLIDPCertificateIDByRole(c ConfiguredClient, role string) resource.ImportStateIdFunc {
	return func(_ *terraform.State) (string, error) {
		list, err := c.ClientV2.API.Admin().SamlSettings().IdpCertificates().Get(ctx, nil)
		if err != nil {
			return "", fmt.Errorf("failed to list SAML IdP certificates: %w", err)
		}
		for _, cert := range list.GetData() {
			if cert.GetAttributes() != nil && enumStringOrEmpty(cert.GetAttributes().GetCertRole()) == role {
				return valueOrZero(cert.GetId()), nil
			}
		}
		return "", fmt.Errorf("no %s SAML IdP certificate found", role)
	}
}

func testAccTFESAMLIDPCertificate_basic(displayName, cert string) string {
	return testAccTFESAMLIDPCertificate_named("foobar", displayName, cert)
}

func testAccTFESAMLIDPCertificate_named(name, displayName, cert string) string {
	return fmt.Sprintf(`
resource "tfe_saml_idp_certificate" %q {
  display_name = %q
  cert         = %q
}`, name, displayName, cert)
}

func testAccTFESAMLIDPCertificate_pair(certA, certB string) string {
	return testAccTFESAMLIDPCertificate_named("a", "fooidp-a", certA) +
		testAccTFESAMLIDPCertificate_named("b", "fooidp-b", certB)
}

// testAccTFESAMLIDPCertificate_empty is a config with no certificates, used to
// destroy them while keeping the test step non-empty.
func testAccTFESAMLIDPCertificate_empty() string {
	return `locals {}`
}

func TestSAMLIDPCertificateModifyPlan(t *testing.T) {
	r := &resourceTFESAMLIDPCertificate{}
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	sch := schemaResp.Schema

	body := testIDPCertBody(t)
	original := wrapPEM(t, body, 64)
	reformatted := wrapPEM(t, body, 76)
	other := generateSelfSignedCertPEM(t)

	stateFor := func(role string) modelTFESAMLIDPCertificate {
		return modelTFESAMLIDPCertificate{
			ID:          types.StringValue("idpc-abc"),
			DisplayName: types.StringValue("idp"),
			Cert:        customtypes.NewPEMCertificateValue(original),
			Fingerprint: types.StringValue("AA:BB"),
			ExpiresAt:   types.StringValue("2030-01-01T00:00:00Z"),
			CreatedAt:   types.StringValue("2026-01-01T00:00:00Z"),
			Issuer:      types.StringValue("O=Foo"),
			CertRole:    types.StringValue(role),
		}
	}
	planFor := func(role string, cert customtypes.PEMCertificateValue) modelTFESAMLIDPCertificate {
		m := stateFor(role)
		m.DisplayName = types.StringValue("idp-renamed")
		m.Cert = cert
		m.Fingerprint = types.StringUnknown()
		m.ExpiresAt = types.StringUnknown()
		m.Issuer = types.StringUnknown()
		return m
	}
	unknownCert := customtypes.PEMCertificateValue{StringValue: types.StringUnknown()}

	for _, tc := range []struct {
		name            string
		role            string
		cert            customtypes.PEMCertificateValue
		wantErr         bool
		wantFingerprint types.String
	}{
		{"managed, new body", samlIDPCertRoleManaged, customtypes.NewPEMCertificateValue(other), false, types.StringUnknown()},
		{"managed, reformatted body keeps derived attributes", samlIDPCertRoleManaged, customtypes.NewPEMCertificateValue(reformatted), false, types.StringValue("AA:BB")},
		{"legacy_primary, reformatted body is allowed", "legacy_primary", customtypes.NewPEMCertificateValue(reformatted), false, types.StringValue("AA:BB")},
		{"legacy_primary, new body is rejected", "legacy_primary", customtypes.NewPEMCertificateValue(other), true, types.StringUnknown()},
		{"legacy_old, new body is rejected", "legacy_old", customtypes.NewPEMCertificateValue(other), true, types.StringUnknown()},
		{"unknown cert is left alone", "legacy_primary", unknownCert, false, types.StringUnknown()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			plan := tfsdk.Plan{Schema: sch}
			if diags := plan.Set(ctx, planFor(tc.role, tc.cert)); diags.HasError() {
				t.Fatalf("plan set: %v", diags)
			}
			state := tfsdk.State{Schema: sch}
			if diags := state.Set(ctx, stateFor(tc.role)); diags.HasError() {
				t.Fatalf("state set: %v", diags)
			}

			req := fwresource.ModifyPlanRequest{Plan: plan, State: state}
			resp := &fwresource.ModifyPlanResponse{Plan: plan}
			r.ModifyPlan(ctx, req, resp)

			if got := resp.Diagnostics.HasError(); got != tc.wantErr {
				t.Fatalf("expected error=%t, got diagnostics: %v", tc.wantErr, resp.Diagnostics)
			}
			var got modelTFESAMLIDPCertificate
			if diags := resp.Plan.Get(ctx, &got); diags.HasError() {
				t.Fatalf("plan get: %v", diags)
			}
			if !got.Fingerprint.Equal(tc.wantFingerprint) {
				t.Errorf("fingerprint: expected %s, got %s", tc.wantFingerprint, got.Fingerprint)
			}
		})
	}
}

func TestPreserveCertFormatting(t *testing.T) {
	body := testIDPCertBody(t)
	server := customtypes.NewPEMCertificateValue(wrapPEM(t, body, 64))
	reformatted := customtypes.NewPEMCertificateValue(wrapPEM(t, body, 76))
	other := customtypes.NewPEMCertificateValue(generateSelfSignedCertPEM(t))

	for _, tc := range []struct {
		name  string
		prior customtypes.PEMCertificateValue
		want  customtypes.PEMCertificateValue
	}{
		{"null prior uses server value", customtypes.NewPEMCertificateNull(), server},
		{"same body keeps prior formatting", reformatted, reformatted},
		{"different body uses server value", other, server},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := preserveCertFormatting(ctx, server, tc.prior); !got.Equal(tc.want) {
				t.Errorf("expected %q, got %q", tc.want.ValueString(), got.ValueString())
			}
		})
	}
}

func TestModelFromSAMLIDPCertificate(t *testing.T) {
	if _, err := modelFromSAMLIDPCertificate(nil); err == nil {
		t.Error("expected an error for nil data")
	}
	if _, err := modelFromSAMLIDPCertificate(models.NewSamlIdpCertificates()); err == nil {
		t.Error("expected an error for nil attributes")
	}
	if _, err := modelFromSAMLIDPCertificateEnvelope(nil); err == nil {
		t.Error("expected an error for a nil envelope")
	}

	attrs := models.NewSamlIdpCertificates_attributes()
	attrs.SetDisplayName(ptr("idp"))
	attrs.SetCert(ptr("CERT"))
	data := models.NewSamlIdpCertificates()
	data.SetId(ptr("idpc-abc"))
	data.SetAttributes(attrs)

	m, err := modelFromSAMLIDPCertificate(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.ID.ValueString() != "idpc-abc" || m.DisplayName.ValueString() != "idp" || m.Cert.ValueString() != "CERT" {
		t.Errorf("unexpected model: %+v", m)
	}
	if !m.ExpiresAt.IsNull() || !m.CreatedAt.IsNull() {
		t.Errorf("missing times must be null, got expires_at=%s created_at=%s", m.ExpiresAt, m.CreatedAt)
	}
}
