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
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func TestAccTeamTokenEphemeralResource_basic(t *testing.T) {
	// The multiple-team-tokens flag is rolled out to prod but still needs to
	// evaluate true before these tests will pass in CI.
	skipUnlessBeta(t)

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
		PreCheck: func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"echo": echoprovider.NewProviderServer(),
		},
		Steps: []resource.TestStep{
			{
				Config: testAccTeamTokenEphemeralResourceConfig(org.Name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("echo.this", tfjsonpath.New("data").AtMapKey("team_id"), knownvalue.StringRegexp(regexp.MustCompile(`^team\-[a-zA-Z0-9]+$`))),
				},
			},
		},
	})
}

func TestAccTeamTokenEphemeralResource_expiredAt(t *testing.T) {
	// The multiple-team-tokens flag is rolled out to prod but still needs to
	// evaluate true before these tests will pass in CI.
	skipUnlessBeta(t)

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_10_0),
		},
		PreCheck: func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"echo": echoprovider.NewProviderServer(),
		},
		Steps: []resource.TestStep{
			{
				Config: testAccTeamTokenEphemeralResourceConfig_expiredAt(org.Name),
			},
		},
	})
}

func testAccTeamTokenEphemeralResourceConfig(orgName string) string {
	return fmt.Sprintf(`
resource "tfe_team" "this" {
  name         = "team-test"
  organization = "%s"
}

ephemeral "tfe_team_token" "this" {
  team_id = tfe_team.this.id
}

provider "echo" {
	data = ephemeral.tfe_team_token.this
}

resource "echo" "this" {}
`, orgName)
}

func testAccTeamTokenEphemeralResourceConfig_expiredAt(orgName string) string {
	return fmt.Sprintf(`
resource "tfe_team" "this" {
  name         = "team-test"
  organization = "%s"
}

ephemeral "tfe_team_token" "this" {
  team_id = tfe_team.this.id
	expired_at = "2100-01-01T00:00:00Z"
}
provider "echo" {
	data = ephemeral.tfe_team_token.this
}
resource "echo" "this" {}
`, orgName)
}
