// Copyright IBM Corp. 2018, 2025
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
	"github.com/hashicorp/terraform-plugin-testing/tfversion"

	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const testSMTPResourceName = "tfe_smtp_settings.foobar"

// FLAKE ALERT: SMTP settings are a singleton resource shared by the entire TFE
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

// TestAccTFESMTPSettings_omnibus test suite is skipped in the CI, and will only run in TFE Nightly workflow
// Should this test name ever change, you will also need to update the regex in ci.yml

func TestAccTFESMTPSettings_omnibus(t *testing.T) {
	skipIfCloud(t)

	t.Run("basic SMTP settings without authentication", func(t *testing.T) {
		s := tfe.AdminSMTPSetting{
			Host:     "foobar.com",
			Port:     25,
			Sender:   "sender@foorbar.com",
			Auth:     "none",
		}
		resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccTFESMTPSettingsDestroy,
		Steps: []resource.TestStep{
				{
					Config: testAccTFESMTPSettings_AuthNone(s),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(testSMTPResourceName, "id", "smtp"),
						resource.TestCheckResourceAttr(testSMTPResourceName, "enabled", "false"),
						resource.TestCheckResourceAttr(testSMTPResourceName, "host", s.Host),
						resource.TestCheckResourceAttr(testSMTPResourceName, "port", strconv.Itoa(s.Port)),
						resource.TestCheckResourceAttr(testSMTPResourceName, "sender", s.Sender),
						resource.TestCheckResourceAttr(testSMTPResourceName, "auth", string(s.Auth)),
					),
				},
			},
		})
	})
}

func TestAccTFESMTPSettings_AuthNone(t *testing.T) {
	s := tfe.AdminSMTPSetting{
		Host:     "foobar.com",
		Port:     25,
		Sender:   "sender@foorbar.com",
		Auth:     "none",
	}

	resource.Test(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(version.Must(version.NewVersion("1.11.0"))),
		},
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFESMTPSettings_AuthNone(s),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testSMTPResourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(testSMTPResourceName, "host", s.Host),
					resource.TestCheckResourceAttr(testSMTPResourceName, "port", strconv.Itoa(s.Port)),
					resource.TestCheckResourceAttr(testSMTPResourceName, "auth", string(s.Auth)),
				),
			},
		},
	})
}
func TestAccTFESMTPSettings_AuthPlain_writeOnly(t *testing.T) {
	s := tfe.AdminSMTPSetting{
		Host:     "foobar.com",
		Port:     25,
		Sender:   "sender@foorbar.com",
		Auth:     "plain",
		Username: "foo",
	}
	password := randomString(t)
	resource.Test(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(version.Must(version.NewVersion("1.11.0"))),
		},
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFESMTPSettings_AuthPlainLogin_writeOnly(s, password),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testSMTPResourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(testSMTPResourceName, "host", s.Host),
					resource.TestCheckResourceAttr(testSMTPResourceName, "port", strconv.Itoa(s.Port)),
					resource.TestCheckResourceAttr(testSMTPResourceName, "username", s.Username),
					resource.TestCheckNoResourceAttr(testSMTPResourceName, "password_wo"),
					resource.TestCheckResourceAttr(testSMTPResourceName, "password_wo_version", "1"),
				),
			},
		},
	})
}

func TestAccTFESMTPSettings_AuthLogin_writeOnly(t *testing.T) {
	s := tfe.AdminSMTPSetting{
		Host:     "foobar.com",
		Port:     25,
		Sender:   "sender@foorbar.com",
		Auth:     "login",
		Username: "foo",
	}
	password := randomString(t)
	resource.Test(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(version.Must(version.NewVersion("1.11.0"))),
		},
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFESMTPSettings_AuthPlainLogin_writeOnly(s, password),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testSMTPResourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(testSMTPResourceName, "host", s.Host),
					resource.TestCheckResourceAttr(testSMTPResourceName, "port", strconv.Itoa(s.Port)),
					resource.TestCheckResourceAttr(testSMTPResourceName, "username", s.Username),
					resource.TestCheckNoResourceAttr(testSMTPResourceName, "password_wo"),
					resource.TestCheckResourceAttr(testSMTPResourceName, "password_wo_version", "1"),
				),
			},
		},
	})
}


func TestAccTFESMTPSettings_AuthPlain_writeOnly_update(t *testing.T) {
	s := tfe.AdminSMTPSetting{
		Host:     "foobar.com",
		Port:     25,
		Sender:   "sender@foorbar.com",
		Auth:     "plain",
		Username: "foo",
	}
	password1 := randomString(t)
	password2 := randomString(t)
	resource.Test(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(version.Must(version.NewVersion("1.11.0"))),
		},
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFESMTPSettings_AuthPlainLogin_writeOnly_version(s, password1, 1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testSMTPResourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(testSMTPResourceName, "host", s.Host),
					resource.TestCheckResourceAttr(testSMTPResourceName, "port", strconv.Itoa(s.Port)),
					resource.TestCheckResourceAttr(testSMTPResourceName, "username", s.Username),
					resource.TestCheckNoResourceAttr(testSMTPResourceName, "password_wo"),
					resource.TestCheckResourceAttr(testSMTPResourceName, "password_wo_version", "1"),
				),
			},
			{
				Config: testAccTFESMTPSettings_AuthPlainLogin_writeOnly_version(s, password2, 2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testSMTPResourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(testSMTPResourceName, "host", s.Host),
					resource.TestCheckResourceAttr(testSMTPResourceName, "port", strconv.Itoa(s.Port)),
					resource.TestCheckResourceAttr(testSMTPResourceName, "username", s.Username),
					resource.TestCheckNoResourceAttr(testSMTPResourceName, "password_wo"),
					resource.TestCheckResourceAttr(testSMTPResourceName, "password_wo_version", "2"),
				),
			},
		},
	})
}

func TestAccTFESMTPSettings_AuthPlain(t *testing.T) {
	s := tfe.AdminSMTPSetting{
		Host:     "foobar.com",
		Port:     25,
		Sender:   "sender@foorbar.com",
		Auth:     "plain",
		Username: "foo",
	}
	password := randomString(t)
	resource.Test(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(version.Must(version.NewVersion("1.11.0"))),
		},
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFESMTPSettings_AuthPlainLogin(s, password),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testSMTPResourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(testSMTPResourceName, "host", s.Host),
					resource.TestCheckResourceAttr(testSMTPResourceName, "port", strconv.Itoa(s.Port)),
					resource.TestCheckResourceAttr(testSMTPResourceName, "username", s.Username),
					resource.TestCheckResourceAttr(testSMTPResourceName, "password", password),

				),
			},
		},
	})
}

func TestAccTFESMTPSettings_AuthLogin(t *testing.T) {
	s := tfe.AdminSMTPSetting{
		Host:     "foobar.com",
		Port:     25,
		Sender:   "sender@foorbar.com",
		Auth:     "login",
		Username: "foo",
	}
	password := randomString(t)
	resource.Test(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(version.Must(version.NewVersion("1.11.0"))),
		},
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFESMTPSettings_AuthPlainLogin(s, password),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testSMTPResourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(testSMTPResourceName, "host", s.Host),
					resource.TestCheckResourceAttr(testSMTPResourceName, "port", strconv.Itoa(s.Port)),
					resource.TestCheckResourceAttr(testSMTPResourceName, "username", s.Username),
					resource.TestCheckResourceAttr(testSMTPResourceName, "password", password),
				),
			},
		},
	})
}

func testAccTFESMTPSettings_AuthPlainLogin_writeOnly(s tfe.AdminSMTPSetting, password string) string {
	return fmt.Sprintf(`
resource "tfe_smtp_settings" "foobar" {
  enabled               = false
  host                  = "%s"
  port                  = %d
  sender                = "%s"
  auth                  = "%s"
  username              = "%s"
  password_wo           = "%s"
  password_wo_version   = 1
}`, s.Host, s.Port, s.Sender, s.Auth, s.Username, password)
}

func testAccTFESMTPSettings_AuthPlainLogin_writeOnly_version(s tfe.AdminSMTPSetting, password string, version int) string {
	return fmt.Sprintf(`
resource "tfe_smtp_settings" "foobar" {
  enabled               = false
  host                  = "%s"
  port                  = %d
  sender                = "%s"
  auth                  = "%s"
  username              = "%s"
  password_wo           = "%s"
  password_wo_version   = %d
}`, s.Host, s.Port, s.Sender, s.Auth, s.Username, password, version)
}

func testAccTFESMTPSettings_AuthPlainLogin(s tfe.AdminSMTPSetting, password string) string {
	return fmt.Sprintf(`
resource "tfe_smtp_settings" "foobar" {
  enabled               = false
  host                  = "%s"
  port                  = %d
  sender                = "%s"
  auth                  = "%s"
  username              = "%s"
  password              = "%s"
}`, s.Host, s.Port, s.Sender, s.Auth, s.Username, password)
}
func testAccTFESMTPSettings_AuthNone(s tfe.AdminSMTPSetting) string {
	return fmt.Sprintf(`
resource "tfe_smtp_settings" "foobar" {
  enabled               = false
  host                  = "%s"
  port                  = %d
  sender                = "%s"
  auth                  = "%s"
}`, s.Host, s.Port, s.Sender, s.Auth)
}

func testAccTFESMTPSettingsDestroy(_ *terraform.State) error {
	settings, err := testAccConfiguredClient.Client.Admin.Settings.SMTP.Read(ctx)
	if err != nil {
		return fmt.Errorf("failed to read SMTP Settings: %w", err)
	}

	// SMTP settings cannot be deleted, only disabled
	// So we check if they are disabled after destroy
	if settings.Enabled {
		return errors.New("SMTP settings are still enabled")
	}

	return nil
}