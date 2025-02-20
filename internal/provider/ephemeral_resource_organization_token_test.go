// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/echoprovider"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccOrganizationTokenEphemeralResource_basic(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"echo": echoprovider.NewProviderServer(),
		},
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationTokenEphemeralResourceConfig(org.Name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("echo.this", tfjsonpath.New("data"), knownvalue.StringRegexp(regexp.MustCompile(`^[a-zA-Z0-9]+\.atlasv1\.[a-zA-Z0-9]+$`))),
				},
			},
		},
	})
}

func testAccOrganizationTokenEphemeralResourceConfig(orgName string) string {
	return fmt.Sprintf(`
ephemeral "tfe_organization_token" "this" {
  organization = "%s"
}

provider "echo" {
	data = ephemeral.tfe_organization_token.this.token
}

resource "echo" "this" {}
`, orgName)
}
