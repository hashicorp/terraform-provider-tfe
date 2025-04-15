// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/echoprovider"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func TestAccOrganizationTokenEphemeralResource_basic(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_10_0),
		},
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"echo": echoprovider.NewProviderServer(),
		},
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationTokenEphemeralResourceConfig_basic(rInt),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("echo.this", tfjsonpath.New("data").AtMapKey("organization"), knownvalue.StringExact(orgName)),
				},
			},
		},
	})
}

func TestAccOrganizationTokenEphemeralResource_expiredAt(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_10_0),
		},
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"echo": echoprovider.NewProviderServer(),
		},
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationTokenEphemeralResourceConfig_expiredAt(rInt),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("echo.this", tfjsonpath.New("data").AtMapKey("expired_at"), knownvalue.StringExact("2100-01-01T00:00:00Z")),
				},
			},
		},
	})
}

func testAccOrganizationTokenEphemeralResourceConfig_basic(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "this" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

ephemeral "tfe_organization_token" "this" {
  organization = tfe_organization.this.id
}

provider "echo" {
	data = ephemeral.tfe_organization_token.this
}

resource "echo" "this" {}
`, rInt)
}

func testAccOrganizationTokenEphemeralResourceConfig_expiredAt(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "this" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

ephemeral "tfe_organization_token" "this" {
  organization = tfe_organization.this.id
	expired_at = "2100-01-01T00:00:00Z"
}

provider "echo" {
	data = ephemeral.tfe_organization_token.this
}

resource "echo" "this" {}
`, rInt)
}
