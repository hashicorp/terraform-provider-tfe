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

func TestAccTFETerraformVersion_basic(t *testing.T) {
	skipIfCloud(t)

	tfVersion := &tfe.AdminTerraformVersion{}
	sha := genSha(t, "secret", "data")
	version := genSafeRandomTerraformVersion()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFETerraformVersionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETerraformVersion_basic(version, sha),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETerraformVersionExists("tfe_terraform_version.foobar", tfVersion),
					testAccCheckTFETerraformVersionAttributesBasic(tfVersion, version, sha),
					resource.TestCheckResourceAttr(
						"tfe_terraform_version.foobar", "version", version),
					resource.TestCheckResourceAttr(
						"tfe_terraform_version.foobar", "url", "https://www.hashicorp.com"),
					resource.TestCheckResourceAttr(
						"tfe_terraform_version.foobar", "sha", sha),
				),
			},
		},
	})
}

func TestAccTFETerraformVersion_archs(t *testing.T) {
	skipIfCloud(t)

	tfVersion := &tfe.AdminTerraformVersion{}
	sha := genSha(t, "secret", "data")
	version := genSafeRandomTerraformVersion()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFETerraformVersionDestroy,
		Steps: []resource.TestStep{
			{
				Config: resourceTFETerraformVersion_archs(version, sha),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETerraformVersionExists("tfe_terraform_version.foobar", tfVersion),
					testAccCheckTFETerraformVersionAttributesArchs(tfVersion, version, sha),
					resource.TestCheckResourceAttr(
						"tfe_terraform_version.foobar", "version", version),
					resource.TestCheckResourceAttr(
						"tfe_terraform_version.foobar", "archs.0.url", "https://www.hashicorp.com"),
					resource.TestCheckResourceAttr(
						"tfe_terraform_version.foobar", "archs.0.sha", sha),
					resource.TestCheckResourceAttr(
						"tfe_terraform_version.foobar", "archs.0.os", "linux"),
					resource.TestCheckResourceAttr(
						"tfe_terraform_version.foobar", "archs.0.arch", "arm64"),
				),
			},
		},
	})
}

func TestAccTFETerraformVersion_import(t *testing.T) {
	skipIfCloud(t)

	sha := genSha(t, "secret", "data")
	version := genSafeRandomTerraformVersion()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFETerraformVersionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETerraformVersion_basic(version, sha),
			},
			{
				ResourceName:      "tfe_terraform_version.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "tfe_terraform_version.foobar",
				ImportState:       true,
				ImportStateId:     version,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTFETerraformVersion_full(t *testing.T) {
	skipIfCloud(t)

	tfVersion := &tfe.AdminTerraformVersion{}
	sha := genSha(t, "secret", "data")
	version := genSafeRandomTerraformVersion()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFETerraformVersionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETerraformVersion_full(version, sha),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETerraformVersionExists("tfe_terraform_version.foobar", tfVersion),
					testAccCheckTFETerraformVersionAttributesFull(tfVersion, version, sha),
					resource.TestCheckResourceAttr(
						"tfe_terraform_version.foobar", "version", version),
					resource.TestCheckResourceAttr(
						"tfe_terraform_version.foobar", "url", "https://www.hashicorp.com"),
					resource.TestCheckResourceAttr(
						"tfe_terraform_version.foobar", "sha", sha),
					resource.TestCheckResourceAttr(
						"tfe_terraform_version.foobar", "official", "false"),
					resource.TestCheckResourceAttr(
						"tfe_terraform_version.foobar", "enabled", "true"),
					resource.TestCheckResourceAttr(
						"tfe_terraform_version.foobar", "beta", "true"),
					resource.TestCheckResourceAttr(
						"tfe_terraform_version.foobar", "deprecated", "true"),
					resource.TestCheckResourceAttr(
						"tfe_terraform_version.foobar", "deprecated_reason", "foobar"),
				),
			},
		},
	})
}

func testAccCheckTFETerraformVersionDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_terraform_version" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := testAccConfiguredClient.Client.Admin.TerraformVersions.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Terraform version %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckTFETerraformVersionExists(n string, tfVersion *tfe.AdminTerraformVersion) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		v, err := testAccConfiguredClient.Client.Admin.TerraformVersions.Read(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}

		if v.ID != rs.Primary.ID {
			return fmt.Errorf("Terraform version not found")
		}

		*tfVersion = *v

		return nil
	}
}

func testAccCheckTFETerraformVersionAttributesBasic(tfVersion *tfe.AdminTerraformVersion, version string, sha string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if tfVersion.URL != "https://www.hashicorp.com" {
			return fmt.Errorf("Bad URL: %s", tfVersion.URL)
		}

		if tfVersion.Version != version {
			return fmt.Errorf("Bad version: %s", tfVersion.Version)
		}

		if tfVersion.Sha != sha {
			return fmt.Errorf("Bad value for Sha: %v", tfVersion.Sha)
		}

		return nil
	}
}

func testAccCheckTFETerraformVersionAttributesArchs(tfVersion *tfe.AdminTerraformVersion, version string, sha string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Check the version
		if tfVersion.Version != version {
			return fmt.Errorf("bad version: %s", tfVersion.Version)
		}

		// Check the archs attributes
		if len(tfVersion.Archs) == 0 {
			return fmt.Errorf("Archs is empty")
		}

		if tfVersion.Archs[0].URL != "https://www.hashicorp.com" {
			return fmt.Errorf("bad value for  URL: %s", tfVersion.Archs[0].URL)
		}

		if tfVersion.Archs[0].Sha != sha {
			return fmt.Errorf("bad value for Sha: %v", tfVersion.Archs[0].Sha)
		}

		if tfVersion.Archs[0].OS != "linux" {
			return fmt.Errorf("bad value for OS: %s", tfVersion.Archs[0].OS)
		}

		if tfVersion.Archs[0].Arch != "amd64" {
			return fmt.Errorf("bad value for Arch: %s", tfVersion.Archs[0].Arch)
		}

		// Check other attributes
		if tfVersion.Official != false {
			return fmt.Errorf("bad value for official: %t", tfVersion.Official)
		}

		if tfVersion.Enabled != true {
			return fmt.Errorf("bad value for enabled: %t", tfVersion.Enabled)
		}

		if tfVersion.Beta != true {
			return fmt.Errorf("bad value for beta: %t", tfVersion.Beta)
		}

		if tfVersion.Deprecated != true {
			return fmt.Errorf("bad value for deprecated: %t", tfVersion.Deprecated)
		}

		if *tfVersion.DeprecatedReason != "foobar" {
			return fmt.Errorf("bad value for deprecated_reason: %s", *tfVersion.DeprecatedReason)
		}

		return nil
	}
}

func testAccCheckTFETerraformVersionAttributesFull(tfVersion *tfe.AdminTerraformVersion, version string, sha string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if tfVersion.URL != "https://www.hashicorp.com" {
			return fmt.Errorf("Bad URL: %s", tfVersion.URL)
		}

		if tfVersion.Version != version {
			return fmt.Errorf("Bad version: %s", tfVersion.Version)
		}

		if tfVersion.Sha != sha {
			return fmt.Errorf("Bad value for Sha: %v", tfVersion.Sha)
		}

		if tfVersion.Official != false {
			return fmt.Errorf("Bad value for official: %t", tfVersion.Official)
		}

		if tfVersion.Enabled != true {
			return fmt.Errorf("Bad value for enabled: %t", tfVersion.Enabled)
		}

		if tfVersion.Beta != true {
			return fmt.Errorf("Bad value for beta: %t", tfVersion.Beta)
		}

		if tfVersion.Deprecated != true {
			return fmt.Errorf("Bad value for deprecated: %t", tfVersion.Deprecated)
		}

		if *tfVersion.DeprecatedReason != "foobar" {
			return fmt.Errorf("Bad value for deprecated_reason: %s", *tfVersion.DeprecatedReason)
		}

		return nil
	}
}

func testAccTFETerraformVersion_basic(version string, sha string) string {
	return fmt.Sprintf(`
resource "tfe_terraform_version" "foobar" {
  version = "%s"
  url = "https://www.hashicorp.com"
  sha = "%s"
}`, version, sha)
}

func resourceTFETerraformVersion_archs(version string, sha string) string {
	return fmt.Sprintf(`
resource "tfe_terraform_version" "foobar" {
  version = "%s"
  official = false
  enabled = true
  beta = true
  deprecated = true
  deprecated_reason = "foobar"
  archs = {
    url = "https://www.hashicorp.com"
	sha = "%s"
	os = "linux"
	arch = "amd64"
	  }
	}`, version, sha)
}

func testAccTFETerraformVersion_full(version string, sha string) string {
	return fmt.Sprintf(`
resource "tfe_terraform_version" "foobar" {
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

// Helper functions
func genSha(t *testing.T, secret, data string) string {
	h := hmac.New(sha256.New, []byte(secret))
	_, err := h.Write([]byte(data))
	if err != nil {
		t.Fatalf("error writing hmac: %s", err)
	}

	sha := hex.EncodeToString(h.Sum(nil))
	return sha
}

// genSafeRandomTerraformVersion returns a random version number of the form
// `1.0.<RANDOM>`, which HCP Terraform won't ever select as the latest available
// Terraform. (At the time of writing, a fresh HCP Terraform instance will include
// official Terraforms 1.2 and higher.) This is necessary because newly created
// workspaces default to the latest available version, and there's nothing
// preventing unrelated processes from creating workspaces during these tests.
func genSafeRandomTerraformVersion() string {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	// Avoid colliding with an official Terraform version. Highest 1.0 was
	// 1.0.11, so add a little padding and call it good.
	for rInt < 20 {
		rInt = rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	}
	return fmt.Sprintf("1.0.%d", rInt)
}
