// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/echoprovider"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func TestAccOrganizationTokenEphemeralResource_basic(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

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
				Config: testAccOrganizationTokenEphemeralResourceConfig_basic(org.Name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("echo.this", tfjsonpath.New("data"), knownvalue.StringExact(org.Name)),
				},
				RefreshState: false,
			},
		},
	})
}

func TestAccOrganizationTokenEphemeralResource_expiredAt(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

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
				Config: testAccOrganizationTokenEphemeralResourceConfig_expiredAt(org.Name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("echo.this", tfjsonpath.New("data"), knownvalue.StringExact("2100-01-01T00:00:00Z")),
				},
				RefreshState: false,
			},
		},
	})
}

func testAccOrganizationTokenEphemeralResourceConfig_basic(orgName string) string {
	return fmt.Sprintf(`
ephemeral "tfe_organization_token" "this" {
  organization = "%s"
}

provider "echo" {
	data = ephemeral.tfe_organization_token.this.organization
}

resource "echo" "this" {}
`, orgName)
}

func testAccOrganizationTokenEphemeralResourceConfig_expiredAt(orgName string) string {
	return fmt.Sprintf(`
ephemeral "tfe_organization_token" "this" {
  organization = "%s"
	expired_at = "2100-01-01T00:00:00Z"
}

provider "echo" {
	data = ephemeral.tfe_organization_token.this.expired_at
}

resource "echo" "this" {}
`, orgName)
}
