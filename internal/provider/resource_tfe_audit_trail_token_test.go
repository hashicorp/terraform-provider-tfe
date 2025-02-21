// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"regexp"
	"testing"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccTFEAuditTrailToken_basic(t *testing.T) {
	token := &tfe.OrganizationToken{}

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}
	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEAuditTrailTokenDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEAuditTrailToken_basic(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEAuditTrailTokenExists(
						"tfe_audit_trail_token.foobar", token),
					resource.TestCheckResourceAttr(
						"tfe_audit_trail_token.foobar", "organization", org.Name),
				),
			},
		},
	})
}

func TestAccTFEAuditTrailToken_existsWithoutForce(t *testing.T) {
	token := &tfe.OrganizationToken{}

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}
	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEAuditTrailTokenDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEAuditTrailToken_basic(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEAuditTrailTokenExists(
						"tfe_audit_trail_token.foobar", token),
					resource.TestCheckResourceAttr(
						"tfe_audit_trail_token.foobar", "organization", org.Name),
				),
			},

			{
				Config:      testAccTFEAuditTrailToken_existsWithoutForce(org.Name),
				ExpectError: regexp.MustCompile(`token already exists`),
			},
		},
	})
}

func TestAccTFEAuditTrailToken_existsWithForce(t *testing.T) {
	token := &tfe.OrganizationToken{}

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}
	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEAuditTrailTokenDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEAuditTrailToken_basic(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEAuditTrailTokenExists(
						"tfe_audit_trail_token.foobar", token),
					resource.TestCheckResourceAttr(
						"tfe_audit_trail_token.foobar", "organization", org.Name),
				),
			},

			{
				Config: testAccTFEAuditTrailToken_existsWithForce(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEAuditTrailTokenExists(
						"tfe_audit_trail_token.regenerated", token),
					resource.TestCheckResourceAttr(
						"tfe_audit_trail_token.regenerated", "organization", org.Name),
				),
			},
		},
	})
}

func TestAccTFEAuditTrailToken_withValidExpiry(t *testing.T) {
	token := &tfe.OrganizationToken{}

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}
	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	expiredAt := "2051-04-11T23:15:59Z"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEAuditTrailTokenDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEAuditTrailToken_withValidExpiry(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEAuditTrailTokenExists(
						"tfe_audit_trail_token.expiry", token),
					resource.TestCheckResourceAttr(
						"tfe_audit_trail_token.expiry", "expired_at", expiredAt),
				),
			},
		},
	})
}

func TestAccTFEAuditTrailToken_withInvalidExpiry(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}
	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEAuditTrailTokenDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFEAuditTrailToken_withInvalidExpiry(org.Name),
				ExpectError: regexp.MustCompile(`Invalid RFC3339 String Value`),
			},
		},
	})
}

func TestAccTFEAuditTrailToken_import(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}
	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEAuditTrailTokenDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEAuditTrailToken_basic(org.Name),
			},

			{
				ResourceName:            "tfe_audit_trail_token.foobar",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"token"},
			},
		},
	})
}

func testAccCheckTFEAuditTrailTokenExists(
	n string, token *tfe.OrganizationToken) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(ConfiguredClient)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}
		auditTrailTokenType := tfe.AuditTrailToken
		readOptions := tfe.OrganizationTokenReadOptions{
			TokenType: &auditTrailTokenType,
		}
		ot, err := config.Client.OrganizationTokens.ReadWithOptions(ctx, rs.Primary.ID, readOptions)
		if err != nil {
			return err
		}

		if ot == nil {
			return fmt.Errorf("Audit trail token not found")
		}

		*token = *ot

		return nil
	}
}

func testAccCheckTFEAuditTrailTokenDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(ConfiguredClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_audit_trail_token" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}
		auditTrailTokenType := tfe.AuditTrailToken
		readOptions := tfe.OrganizationTokenReadOptions{
			TokenType: &auditTrailTokenType,
		}
		_, err := config.Client.OrganizationTokens.ReadWithOptions(ctx, rs.Primary.ID, readOptions)
		if err == nil {
			return fmt.Errorf("Audit trail token %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccTFEAuditTrailToken_basic(orgName string) string {
	return fmt.Sprintf(`
resource "tfe_audit_trail_token" "foobar" {
  organization = "%s"
}`, orgName)
}

// NOTE: This config is invalid because you cannot manage multiple tokens for
// one org. It is expected to always error.
func testAccTFEAuditTrailToken_existsWithoutForce(orgName string) string {
	return fmt.Sprintf(`
resource "tfe_audit_trail_token" "foobar" {
  organization = "%s"
}

resource "tfe_audit_trail_token" "error" {
  organization = "%s"
}`, orgName, orgName)
}

// NOTE: This config is invalid because you cannot manage multiple tokens for
// one org. It can run without error _once_ due to the presence of
// force_regenerate, but is expected to error on any subsequent run.
func testAccTFEAuditTrailToken_existsWithForce(orgName string) string {
	return fmt.Sprintf(`
resource "tfe_audit_trail_token" "foobar" {
  organization = "%s"
}

resource "tfe_audit_trail_token" "regenerated" {
  organization     = "%s"
  force_regenerate = true
}`, orgName, orgName)
}

func testAccTFEAuditTrailToken_withBlankExpiry(orgName string) string {
	return fmt.Sprintf(`
resource "tfe_audit_trail_token" "foobar" {
  organization = "%s"
  expired_at = ""
}`, orgName)
}

func testAccTFEAuditTrailToken_withValidExpiry(orgName string) string {
	return fmt.Sprintf(`
resource "tfe_audit_trail_token" "expiry" {
  organization  = "%s"
  expired_at 	= "2051-04-11T23:15:59Z"
}`, orgName)
}

func testAccTFEAuditTrailToken_withInvalidExpiry(orgName string) string {
	return fmt.Sprintf(`
resource "tfe_audit_trail_token" "expiry" {
  organization  = "%s"
  expired_at 	= "2000-04-11"
}`, orgName)
}
