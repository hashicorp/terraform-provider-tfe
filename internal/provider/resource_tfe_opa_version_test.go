// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"testing"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccTFEOPAVersion_basic(t *testing.T) {
	skipIfCloud(t)

	opaVersion := &tfe.AdminOPAVersion{}
	sha := genOPASha(t, "secret", "data")
	version := genSafeRandomOPAVersion()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEOPAVersionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOPAVersion_basic(version, sha),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOPAVersionExists("tfe_opa_version.foobar", opaVersion),
					testAccCheckTFEOPAVersionAttributesBasic(opaVersion, version, sha),
					resource.TestCheckResourceAttr(
						"tfe_opa_version.foobar", "version", version),
					resource.TestCheckResourceAttr(
						"tfe_opa_version.foobar", "url", "https://www.hashicorp.com"),
					resource.TestCheckResourceAttr(
						"tfe_opa_version.foobar", "sha", sha),
				),
			},
		},
	})
}

func TestAccTFEOPAVersion_import(t *testing.T) {
	skipIfCloud(t)

	sha := genOPASha(t, "secret", "data")
	version := genSafeRandomOPAVersion()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEOPAVersionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOPAVersion_basic(version, sha),
			},
			{
				ResourceName:      "tfe_opa_version.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "tfe_opa_version.foobar",
				ImportState:       true,
				ImportStateId:     version,
				ImportStateVerify: true,
			},
		},
	})
}
func TestAccTFEOPAVersion_full(t *testing.T) {
	skipIfCloud(t)

	opaVersion := &tfe.AdminOPAVersion{}
	sha := genOPASha(t, "secret", "data")
	version := genSafeRandomOPAVersion()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEOPAVersionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOPAVersion_full(version, sha),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOPAVersionExists("tfe_opa_version.foobar", opaVersion),
					testAccCheckTFEOPAVersionAttributesFull(opaVersion, version, sha),
					resource.TestCheckResourceAttr(
						"tfe_opa_version.foobar", "version", version),
					resource.TestCheckResourceAttr(
						"tfe_opa_version.foobar", "url", "https://www.hashicorp.com"),
					resource.TestCheckResourceAttr(
						"tfe_opa_version.foobar", "sha", sha),
					resource.TestCheckResourceAttr(
						"tfe_opa_version.foobar", "official", "false"),
					resource.TestCheckResourceAttr(
						"tfe_opa_version.foobar", "enabled", "true"),
					resource.TestCheckResourceAttr(
						"tfe_opa_version.foobar", "beta", "true"),
					resource.TestCheckResourceAttr(
						"tfe_opa_version.foobar", "deprecated", "true"),
					resource.TestCheckResourceAttr(
						"tfe_opa_version.foobar", "deprecated_reason", "foobar"),
				),
			},
		},
	})
}

func TestAccTFEOPAVersion_archs(t *testing.T) {
	skipIfCloud(t)

	opaVersion := &tfe.AdminOPAVersion{}
	amd64Sha := genOPASha(t, "secret", "data")
	arm64Sha := genOPASha(t, "another-secret", "data")
	version := genSafeRandomOPAVersion()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEOPAVersionDestroy,
		Steps: []resource.TestStep{
			{
				Config: resourceTFEOPAVersion_archs(version, amd64Sha, arm64Sha),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOPAVersionExists("tfe_opa_version.foobar", opaVersion),
					testAccCheckTFEOPAVersionAttributesArchs(opaVersion, version, amd64Sha, arm64Sha),
					resource.TestCheckResourceAttr(
						"tfe_opa_version.foobar", "version", version),
					resource.TestCheckResourceAttr(
						"tfe_opa_version.foobar", "url", "https://www.hashicorp.com"),
					resource.TestCheckResourceAttr(
						"tfe_opa_version.foobar", "sha", amd64Sha),
					resource.TestCheckResourceAttr(
						"tfe_opa_version.foobar", "archs.0.url", "https://www.hashicorp.com"),
					resource.TestCheckResourceAttr(
						"tfe_opa_version.foobar", "archs.0.sha", amd64Sha),
					resource.TestCheckResourceAttr(
						"tfe_opa_version.foobar", "archs.0.os", "linux"),
					resource.TestCheckResourceAttr(
						"tfe_opa_version.foobar", "archs.0.arch", "amd64"),
					resource.TestCheckResourceAttr(
						"tfe_opa_version.foobar", "archs.1.url", "https://www.hashicorp.com"),
					resource.TestCheckResourceAttr(
						"tfe_opa_version.foobar", "archs.1.sha", arm64Sha),
					resource.TestCheckResourceAttr(
						"tfe_opa_version.foobar", "archs.1.os", "linux"),
					resource.TestCheckResourceAttr(
						"tfe_opa_version.foobar", "archs.1.arch", "arm64"),
				),
			},
		},
	})
}

func testAccCheckTFEOPAVersionDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_opa_version" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := testAccConfiguredClient.Client.Admin.OPAVersions.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("OPA version %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckTFEOPAVersionExists(n string, opaVersion *tfe.AdminOPAVersion) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		v, err := testAccConfiguredClient.Client.Admin.OPAVersions.Read(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}

		if v.ID != rs.Primary.ID {
			return fmt.Errorf("OPA version not found")
		}

		*opaVersion = *v

		return nil
	}
}

func testAccCheckTFEOPAVersionAttributesBasic(opaVersion *tfe.AdminOPAVersion, version string, sha string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if opaVersion.URL != "https://www.hashicorp.com" {
			return fmt.Errorf("Bad URL: %s", opaVersion.URL)
		}

		if opaVersion.Version != version {
			return fmt.Errorf("Bad version: %s", opaVersion.Version)
		}

		if opaVersion.SHA != sha {
			return fmt.Errorf("Bad value for Sha: %v", opaVersion.SHA)
		}

		return nil
	}
}

func testAccCheckTFEOPAVersionAttributesFull(opaVersion *tfe.AdminOPAVersion, version string, sha string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if opaVersion.URL != "https://www.hashicorp.com" {
			return fmt.Errorf("Bad URL: %s", opaVersion.URL)
		}

		if opaVersion.Version != version {
			return fmt.Errorf("Bad version: %s", opaVersion.Version)
		}

		if opaVersion.SHA != sha {
			return fmt.Errorf("Bad value for Sha: %v", opaVersion.SHA)
		}

		if opaVersion.Official != false {
			return fmt.Errorf("Bad value for official: %t", opaVersion.Official)
		}

		if opaVersion.Enabled != true {
			return fmt.Errorf("Bad value for enabled: %t", opaVersion.Enabled)
		}

		if opaVersion.Beta != true {
			return fmt.Errorf("Bad value for beta: %t", opaVersion.Beta)
		}

		if opaVersion.Deprecated != true {
			return fmt.Errorf("Bad value for deprecated: %t", opaVersion.Deprecated)
		}

		if *opaVersion.DeprecatedReason != "foobar" {
			return fmt.Errorf("Bad value for deprecated_reason: %s", *opaVersion.DeprecatedReason)
		}

		return nil
	}
}

func testAccCheckTFEOPAVersionAttributesArchs(opaVersion *tfe.AdminOPAVersion, version string, amd64Sha string, arm64Sha string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if opaVersion.Version != version {
			return fmt.Errorf("bad version: %s", opaVersion.Version)
		}
		if len(opaVersion.Archs) != 2 {
			return fmt.Errorf("expected 1 arch, got %d", len(opaVersion.Archs))
		}
		if opaVersion.Archs[0].OS != "linux" {
			return fmt.Errorf("bad value for OS: %s", opaVersion.Archs[0].OS)
		}
		if opaVersion.Archs[0].Arch != "amd64" {
			return fmt.Errorf("bad value for Arch: %s", opaVersion.Archs[0].Arch)
		}
		if opaVersion.Archs[0].URL != "https://www.hashicorp.com" {
			return fmt.Errorf("bad value for URL: %s", opaVersion.Archs[0].URL)
		}
		if opaVersion.Archs[0].Sha != amd64Sha {
			return fmt.Errorf("bad value for Sha: %v", opaVersion.Archs[0].Sha)
		}

		if opaVersion.Archs[1].OS != "linux" {
			return fmt.Errorf("bad value for OS: %s", opaVersion.Archs[1].OS)
		}
		if opaVersion.Archs[1].Arch != "arm64" {
			return fmt.Errorf("bad value for Arch: %s", opaVersion.Archs[1].Arch)
		}
		if opaVersion.Archs[1].URL != "https://www.hashicorp.com" {
			return fmt.Errorf("bad value for URL: %s", opaVersion.Archs[1].URL)
		}
		if opaVersion.Archs[1].Sha != arm64Sha {
			return fmt.Errorf("bad value for Sha: %v", opaVersion.Archs[1].Sha)
		}

		return nil
	}
}

func testAccTFEOPAVersion_basic(version string, sha string) string {
	return fmt.Sprintf(`
resource "tfe_opa_version" "foobar" {
  version = "%s"
  url = "https://www.hashicorp.com"
  sha = "%s"
}`, version, sha)
}

func testAccTFEOPAVersion_full(version string, sha string) string {
	return fmt.Sprintf(`
resource "tfe_opa_version" "foobar" {
  version = "%s"
  url = "https://www.hashicorp.com"
  sha = "%s"
  official = false
  enabled = true
  beta = true
  deprecated = true
  deprecated_reason = "foobar"
}`, version, sha)
}

func resourceTFEOPAVersion_archs(version string, amd64Sha string, arm64Sha string) string {
	return fmt.Sprintf(`
resource "tfe_opa_version" "foobar" {
  version = "%s"
  official = false
  enabled = true

  archs = [{
    os = "linux"
    arch = "amd64"
    url = "https://www.hashicorp.com"
    sha = "%s"
  },
  {
    os = "linux"
	arch = "arm64"
	url = "https://www.hashicorp.com"
	sha = "%s"}]
}
`, version, amd64Sha, arm64Sha)
}

// Helper functions
func genOPASha(t *testing.T, secret, data string) string {
	h := hmac.New(sha256.New, []byte(secret))
	_, err := h.Write([]byte(data))
	if err != nil {
		t.Fatalf("error writing hmac: %s", err)
	}

	sha := hex.EncodeToString(h.Sum(nil))
	return sha
}

// genSafeRandomOPAVersion returns a random version number of the form
// `0.0.<RANDOM>`, which HCP Terraform won't ever select as the latest available
// OPA. (At the time of writing, a fresh HCP Terraform instance will include
// official OPAs 0.44.0 and higher.) This is necessary because newly created
// workspaces default to the latest available version, and there's nothing
// preventing unrelated processes from creating workspaces during these tests.
func genSafeRandomOPAVersion() string {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	// Avoid colliding with an official OPA version. Highest was
	// 0.58.0, so add a little padding and call it good.
	for rInt < 20 {
		rInt = rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	}
	return fmt.Sprintf("0.0.%d", rInt)
}
