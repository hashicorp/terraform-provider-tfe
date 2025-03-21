// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"math/rand"
	"regexp"
	"testing"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccTFESSHKey_basic(t *testing.T) {
	sshKey := &tfe.SSHKey{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFESSHKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFESSHKey_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFESSHKeyExists(
						"tfe_ssh_key.foobar", sshKey),
					testAccCheckTFESSHKeyAttributes(sshKey),
					resource.TestCheckResourceAttr(
						"tfe_ssh_key.foobar", "name", "ssh-key-test"),
					resource.TestCheckResourceAttr(
						"tfe_ssh_key.foobar", "key", "SSH-KEY-CONTENT"),
				),
			},
		},
	})
}

func TestAccTFESSHKey_update(t *testing.T) {
	sshKey := &tfe.SSHKey{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFESSHKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFESSHKey_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFESSHKeyExists(
						"tfe_ssh_key.foobar", sshKey),
					testAccCheckTFESSHKeyAttributes(sshKey),
					resource.TestCheckResourceAttr(
						"tfe_ssh_key.foobar", "name", "ssh-key-test"),
					resource.TestCheckResourceAttr(
						"tfe_ssh_key.foobar", "key", "SSH-KEY-CONTENT"),
				),
			},

			{
				Config: testAccTFESSHKey_update(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFESSHKeyExists(
						"tfe_ssh_key.foobar", sshKey),
					testAccCheckTFESSHKeyAttributesUpdated(sshKey),
					resource.TestCheckResourceAttr(
						"tfe_ssh_key.foobar", "name", "ssh-key-updated"),
					resource.TestCheckResourceAttr(
						"tfe_ssh_key.foobar", "key", "SSH-KEY-CONTENT"),
				),
			},
		},
	})
}

func TestAccTFESSHKey_keyWO(t *testing.T) {
	sshKey := &tfe.SSHKey{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	// Create the value comparer so we can add state values to it during the test steps
	compareValuesDiffer := statecheck.CompareValue(compare.ValuesDiffer())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEOrganizationRunTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFESSHKey_keyAndKeyWO(rInt),
				ExpectError: regexp.MustCompile(`Attribute "key_wo" cannot be specified when "key" is specified`),
			},
			{
				Config: testAccTFESSHKey_keyWO(rInt, "SSH-KEY-CONTENT"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFESSHKeyExists("tfe_ssh_key.foobar", sshKey),
					testAccCheckTFESSHKeyAttributes(sshKey),
					resource.TestCheckResourceAttr("tfe_ssh_key.foobar", "name", "ssh-key-test"),
					resource.TestCheckNoResourceAttr("tfe_ssh_key.foobar", "key"),
					resource.TestCheckNoResourceAttr("tfe_ssh_key.foobar", "key_wo"),
				),
				// Register the id with the value comparer so we can assert that the
				// resource has been replaced in the next step.
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesDiffer.AddStateValue(
						"tfe_ssh_key.foobar", tfjsonpath.New("id"),
					),
				},
			},
			{
				Config: testAccTFESSHKey_keyWO(rInt, "SSH-KEY-CONTENT-UPDATED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFESSHKeyExists("tfe_ssh_key.foobar", sshKey),
					testAccCheckTFESSHKeyAttributes(sshKey),
					resource.TestCheckResourceAttr("tfe_ssh_key.foobar", "name", "ssh-key-test"),
					resource.TestCheckNoResourceAttr("tfe_ssh_key.foobar", "key"),
					resource.TestCheckNoResourceAttr("tfe_ssh_key.foobar", "key_wo"),
				),
				// Register the id with the value comparer so we can assert that the
				// resource has been replaced in the next step.
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesDiffer.AddStateValue(
						"tfe_ssh_key.foobar", tfjsonpath.New("id"),
					),
				},
			},
			{
				Config: testAccTFESSHKey_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFESSHKeyExists("tfe_ssh_key.foobar", sshKey),
					testAccCheckTFESSHKeyAttributes(sshKey),
					resource.TestCheckResourceAttr("tfe_ssh_key.foobar", "name", "ssh-key-test"),
					resource.TestCheckResourceAttr("tfe_ssh_key.foobar", "key", "SSH-KEY-CONTENT"),
				),
				// Ensure that the resource has been replaced
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesDiffer.AddStateValue(
						"tfe_ssh_key.foobar", tfjsonpath.New("id"),
					),
				},
			},
		},
	})
}

func testAccCheckTFESSHKeyExists(
	n string, sshKey *tfe.SSHKey) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(ConfiguredClient)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		sk, err := config.Client.SSHKeys.Read(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}

		if sk == nil {
			return fmt.Errorf("SSH key not found")
		}

		*sshKey = *sk

		return nil
	}
}

func testAccCheckTFESSHKeyAttributes(
	sshKey *tfe.SSHKey) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if sshKey.Name != "ssh-key-test" {
			return fmt.Errorf("Bad name: %s", sshKey.Name)
		}
		return nil
	}
}

func testAccCheckTFESSHKeyAttributesUpdated(
	sshKey *tfe.SSHKey) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if sshKey.Name != "ssh-key-updated" {
			return fmt.Errorf("Bad name: %s", sshKey.Name)
		}
		return nil
	}
}

func testAccCheckTFESSHKeyDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(ConfiguredClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_ssh_key" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := config.Client.SSHKeys.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("SSH key %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccTFESSHKey_basic(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_ssh_key" "foobar" {
  name         = "ssh-key-test"
  organization = tfe_organization.foobar.id
  key          = "SSH-KEY-CONTENT"
}`, rInt)
}

func testAccTFESSHKey_update(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_ssh_key" "foobar" {
  name         = "ssh-key-updated"
  organization = tfe_organization.foobar.id
  key          = "SSH-KEY-CONTENT"
}`, rInt)
}

func testAccTFESSHKey_keyWO(rInt int, key string) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_ssh_key" "foobar" {
  name         = "ssh-key-test"
  organization = tfe_organization.foobar.id
  key_wo       = "%s"
}`, rInt, key)
}

func testAccTFESSHKey_keyAndKeyWO(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_ssh_key" "foobar" {
  name         = "ssh-key-test"
  organization = tfe_organization.foobar.id
  key          = "SSH-KEY-CONTENT"
  key_wo       = "SSH-KEY-CONTENT"
}`, rInt)
}
