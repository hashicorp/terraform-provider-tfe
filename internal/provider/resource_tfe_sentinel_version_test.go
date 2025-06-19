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

func TestAccTFESentinelVersion_basic(t *testing.T) {
	skipIfCloud(t)

	sentinelVersion := &tfe.AdminSentinelVersion{}
	sha := genSentinelSha(t, "secret", "data")
	version := genSafeRandomSentinelVersion()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFESentinelVersionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFESentinelVersion_basic(version, sha),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFESentinelVersionExists("tfe_sentinel_version.foobar", sentinelVersion),
					testAccCheckTFESentinelVersionAttributesBasic(sentinelVersion, version, sha),
					resource.TestCheckResourceAttr(
						"tfe_sentinel_version.foobar", "version", version),
					resource.TestCheckResourceAttr(
						"tfe_sentinel_version.foobar", "url", "https://www.hashicorp.com"),
					resource.TestCheckResourceAttr(
						"tfe_sentinel_version.foobar", "sha", sha),
				),
			},
		},
	})
}

func TestAccTFESentinelVersion_archs(t *testing.T) {
	skipIfCloud(t)

	sentinelVersion := &tfe.AdminSentinelVersion{}
	sha := genSentinelSha(t, "secret", "data")
	version := genSafeRandomSentinelVersion()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFESentinelVersionDestroy,
		Steps: []resource.TestStep{
			{
				Config: resourceTFESentinelVersion_archs(version, sha),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFESentinelVersionExists("tfe_sentinel_version.foobar", sentinelVersion),
					testAccCheckTFESentinelVersionAttributeArchs(sentinelVersion, version, sha),
					resource.TestCheckResourceAttr(
						"tfe_sentinel_version.foobar", "version", version),
					resource.TestCheckResourceAttr(
						"tfe_sentinel_version.foobar", "url", "https://www.hashicorp.com"),
					resource.TestCheckResourceAttr(
						"tfe_sentinel_version.foobar", "sha", sha),
				),
			},
		},
	})
}

func TestAccTFESentinelVersion_import(t *testing.T) {
	skipIfCloud(t)

	sha := genSentinelSha(t, "secret", "data")
	version := genSafeRandomSentinelVersion()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFESentinelVersionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFESentinelVersion_basic(version, sha),
			},
			{
				ResourceName:      "tfe_sentinel_version.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "tfe_sentinel_version.foobar",
				ImportState:       true,
				ImportStateId:     version,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTFESentinelVersion_full(t *testing.T) {
	skipIfCloud(t)

	sentinelVersion := &tfe.AdminSentinelVersion{}
	sha := genSentinelSha(t, "secret", "data")
	version := genSafeRandomSentinelVersion()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFESentinelVersionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFESentinelVersion_full(version, sha),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFESentinelVersionExists("tfe_sentinel_version.foobar", sentinelVersion),
					testAccCheckTFESentinelVersionAttributesFull(sentinelVersion, version, sha),
					resource.TestCheckResourceAttr(
						"tfe_sentinel_version.foobar", "version", version),
					resource.TestCheckResourceAttr(
						"tfe_sentinel_version.foobar", "url", "https://www.hashicorp.com"),
					resource.TestCheckResourceAttr(
						"tfe_sentinel_version.foobar", "sha", sha),
					resource.TestCheckResourceAttr(
						"tfe_sentinel_version.foobar", "official", "false"),
					resource.TestCheckResourceAttr(
						"tfe_sentinel_version.foobar", "enabled", "true"),
					resource.TestCheckResourceAttr(
						"tfe_sentinel_version.foobar", "beta", "true"),
					resource.TestCheckResourceAttr(
						"tfe_sentinel_version.foobar", "deprecated", "true"),
					resource.TestCheckResourceAttr(
						"tfe_sentinel_version.foobar", "deprecated_reason", "foobar"),
				),
			},
		},
	})
}

func testAccCheckTFESentinelVersionDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_sentinel_version" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := testAccConfiguredClient.Client.Admin.SentinelVersions.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Sentinel version %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckTFESentinelVersionExists(n string, sentinelVersion *tfe.AdminSentinelVersion) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		v, err := testAccConfiguredClient.Client.Admin.SentinelVersions.Read(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}

		if v.ID != rs.Primary.ID {
			return fmt.Errorf("Sentinel version not found")
		}

		*sentinelVersion = *v

		return nil
	}
}

func testAccCheckTFESentinelVersionAttributesBasic(sentinelVersion *tfe.AdminSentinelVersion, version string, sha string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if sentinelVersion.URL != "https://www.hashicorp.com" {
			return fmt.Errorf("Bad URL: %s", sentinelVersion.URL)
		}

		if sentinelVersion.Version != version {
			return fmt.Errorf("Bad version: %s", sentinelVersion.Version)
		}

		if sentinelVersion.SHA != sha {
			return fmt.Errorf("Bad value for Sha: %v", sentinelVersion.SHA)
		}

		return nil
	}
}

func testAccCheckTFESentinelVersionAttributesFull(sentinelVersion *tfe.AdminSentinelVersion, version string, sha string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if sentinelVersion.URL != "https://www.hashicorp.com" {
			return fmt.Errorf("Bad URL: %s", sentinelVersion.URL)
		}

		if sentinelVersion.Version != version {
			return fmt.Errorf("Bad version: %s", sentinelVersion.Version)
		}

		if sentinelVersion.SHA != sha {
			return fmt.Errorf("Bad value for Sha: %v", sentinelVersion.SHA)
		}

		if sentinelVersion.Official != false {
			return fmt.Errorf("Bad value for official: %t", sentinelVersion.Official)
		}

		if sentinelVersion.Enabled != true {
			return fmt.Errorf("Bad value for enabled: %t", sentinelVersion.Enabled)
		}

		if sentinelVersion.Beta != true {
			return fmt.Errorf("Bad value for beta: %t", sentinelVersion.Beta)
		}

		if sentinelVersion.Deprecated != true {
			return fmt.Errorf("Bad value for deprecated: %t", sentinelVersion.Deprecated)
		}

		if *sentinelVersion.DeprecatedReason != "foobar" {
			return fmt.Errorf("Bad value for deprecated_reason: %s", *sentinelVersion.DeprecatedReason)
		}

		return nil
	}
}

func testAccCheckTFESentinelVersionAttributeArchs(sentinelVersion *tfe.AdminSentinelVersion, version string, sha string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if sentinelVersion.Version != version {
			return fmt.Errorf("bad version: %s", sentinelVersion.Version)
		}

		if sentinelVersion.Official != false {
			return fmt.Errorf("bad value for official: %t", sentinelVersion.Official)
		}

		if sentinelVersion.Enabled != true {
			return fmt.Errorf("bad value for enabled: %t", sentinelVersion.Enabled)
		}

		if len(sentinelVersion.Archs) != 1 {
			return fmt.Errorf("Eexpected 1 arch, got %d", len(sentinelVersion.Archs))
		}

		arch := sentinelVersion.Archs[0]
		if arch.URL != "https://www.hashicorp.com" {
			return fmt.Errorf("bad value for URL: %s", arch.URL)
		}

		if arch.Sha != sha {
			return fmt.Errorf("bad value for Sha: %v", arch.Sha)
		}

		if arch.OS != "linux" {
			return fmt.Errorf("bad value for OS: %s", arch.OS)
		}

		if arch.Arch != "amd64" {
			return fmt.Errorf("bad value for Arch: %s", arch.Arch)
		}

		return nil
	}
}

func testAccTFESentinelVersion_basic(version string, sha string) string {
	return fmt.Sprintf(`
resource "tfe_sentinel_version" "foobar" {
  version = "%s"
  url = "https://www.hashicorp.com"
  sha = "%s"
}`, version, sha)
}

func testAccTFESentinelVersion_full(version string, sha string) string {
	return fmt.Sprintf(`
resource "tfe_sentinel_version" "foobar" {
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

func resourceTFESentinelVersion_archs(version string, sha string) string {
	return fmt.Sprintf(`
resource "tfe_sentinel_version" "foobar" {
  version = "%s"
  official = false
  enabled = true
  archs {
      url = "https://www.hashicorp.com"
 	  sha = "%s"
	  os = "linux"
	  arch = "amd64"
	    }
}`, version, sha)
}

// Helper functions
func genSentinelSha(t *testing.T, secret, data string) string {
	h := hmac.New(sha256.New, []byte(secret))
	_, err := h.Write([]byte(data))
	if err != nil {
		t.Fatalf("error writing hmac: %s", err)
	}

	sha := hex.EncodeToString(h.Sum(nil))
	return sha
}

// genSafeRandomSentinelVersion returns a random version number of the form
// `0.0.<RANDOM>`, which HCP Terraform won't ever select as the latest available
// Sentinel. (At the time of writing, a fresh HCP Terraform instance will include
// official Sentinels 0.22.1 and higher.) This is necessary because newly created
// workspaces default to the latest available version, and there's nothing
// preventing unrelated processes from creating workspaces during these tests.
func genSafeRandomSentinelVersion() string {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	// Avoid colliding with an official Sentinel version. Highest was
	// 0.24.0, so add a little padding and call it good.
	for rInt < 20 {
		rInt = rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	}
	return fmt.Sprintf("0.0.%d", rInt)
}
