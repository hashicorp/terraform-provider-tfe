package tfe

import (
	"fmt"
	"testing"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccTFESSHKey_basic(t *testing.T) {
	sshKey := &tfe.SSHKey{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFESSHKeyDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccTFESSHKey_basic,
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

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFESSHKeyDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccTFESSHKey_basic,
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

			resource.TestStep{
				Config: testAccTFESSHKey_update,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFESSHKeyExists(
						"tfe_ssh_key.foobar", sshKey),
					testAccCheckTFESSHKeyAttributesUpdated(sshKey),
					resource.TestCheckResourceAttr(
						"tfe_ssh_key.foobar", "name", "ssh-key-updated"),
					resource.TestCheckResourceAttr(
						"tfe_ssh_key.foobar", "key", "UPDATED-SSH-KEY-CONTENT"),
				),
			},
		},
	})
}

func TestAccTFESSHKey_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFESSHKeyDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccTFESSHKey_basic,
			},

			resource.TestStep{
				ResourceName:            "tfe_ssh_key.foobar",
				ImportState:             true,
				ImportStateIdPrefix:     "terraform-test/",
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"key"},
			},
		},
	})
}

func testAccCheckTFESSHKeyExists(
	n string, sshKey *tfe.SSHKey) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		tfeClient := testAccProvider.Meta().(*tfe.Client)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		sk, err := tfeClient.SSHKeys.Read(ctx, rs.Primary.ID)
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
	tfeClient := testAccProvider.Meta().(*tfe.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_ssh_key" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := tfeClient.SSHKeys.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("SSH key %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

const testAccTFESSHKey_basic = `
resource "tfe_organization" "foobar" {
  name = "terraform-test"
  email = "admin@company.com"
}

resource "tfe_ssh_key" "foobar" {
  name = "ssh-key-test"
  organization = "${tfe_organization.foobar.id}"
  key = "SSH-KEY-CONTENT"
}`

const testAccTFESSHKey_update = `
resource "tfe_organization" "foobar" {
  name = "terraform-test"
  email = "admin@company.com"
}

resource "tfe_ssh_key" "foobar" {
  name = "ssh-key-updated"
  organization = "${tfe_organization.foobar.id}"
  key = "UPDATED-SSH-KEY-CONTENT"
}`
