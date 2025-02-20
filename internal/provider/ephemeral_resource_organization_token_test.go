// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/echoprovider"
)

func TestAccOrganizationTokenEphemeralResource_basic(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	result, err := tfeClient.OrganizationTokens.Create(context.Background(), org.Name)
	if err != nil {
		t.Fatal(err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"echo": echoprovider.NewProviderServer(),
		},
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationTokenEphemeralResourceConfig(org.Name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("echo.data", "token", result.Token),
				),
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
