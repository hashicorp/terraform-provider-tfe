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

func TestAccTeamTokenEphemeralResource_basic(t *testing.T) {
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
				Config: testAccTeamTokenEphemeralResourceConfig(org.Name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("echo.this", tfjsonpath.New("data"), knownvalue.StringRegexp(regexp.MustCompile(`^[a-zA-Z0-9]+\.atlasv1\.[a-zA-Z0-9]+$`))),
				},
			},
		},
	})
}

func testAccTeamTokenEphemeralResourceConfig(orgName string) string {
	return fmt.Sprintf(`
resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = "%s"
}

resource "tfe_team_token" "foobar" {
  team_id = tfe_team.foobar.id
}

provider "echo" {
  data = tfe_team_token.foobar.token
}

resource "echo" "this" {}
`, orgName)
}
