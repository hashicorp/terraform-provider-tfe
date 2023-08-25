// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccTFESSHKeyDataSource_basic(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFESSHKeyDataSourceConfig(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.tfe_ssh_key.foobar", "name", fmt.Sprintf("ssh-key-test-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.tfe_ssh_key.foobar", "organization", orgName),
					resource.TestCheckResourceAttrSet("data.tfe_ssh_key.foobar", "id"),
				),
			},
		},
	})
}

func testAccTFESSHKeyDataSourceConfig(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_ssh_key" "foobar" {
  name         = "ssh-key-test-%d"
  organization = tfe_organization.foobar.id
  key          = "SSH-KEY-CONTENT"
}

data "tfe_ssh_key" "foobar" {
  name         = tfe_ssh_key.foobar.name
  organization = tfe_ssh_key.foobar.organization
}`, rInt, rInt)
}
