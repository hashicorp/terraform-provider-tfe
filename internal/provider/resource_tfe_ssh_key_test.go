// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccTFESSHKey_basic(t *testing.T) {
	sshKey := &tfe.SSHKey{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFESSHKeyDestroy,
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
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFESSHKeyDestroy,
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
