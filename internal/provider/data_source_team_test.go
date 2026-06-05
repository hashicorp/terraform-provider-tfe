// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccTFETeamDataSource_basic(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamDataSourceConfig_basic(rInt, org.Name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.tfe_team.foobar", "name", fmt.Sprintf("team-test-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.tfe_team.foobar", "organization", org.Name),
					resource.TestCheckResourceAttrSet("data.tfe_team.foobar", "id"),
				),
			},
		},
	})
}

func TestAccTFETeamDataSource_ssoTeamId(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	testSsoTeamID := fmt.Sprintf("sso-team-id-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamDataSourceConfig_ssoTeamId(rInt, org.Name, testSsoTeamID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.tfe_team.sso_team", "name", fmt.Sprintf("team-test-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.tfe_team.sso_team", "organization", org.Name),
					resource.TestCheckResourceAttrSet("data.tfe_team.sso_team", "id"),
					resource.TestCheckResourceAttr(
						"data.tfe_team.sso_team", "sso_team_id", testSsoTeamID),
				),
			},
		},
	})
}

func testAccTFETeamDataSourceConfig_basic(rInt int, organization string) string {
	return fmt.Sprintf(`
resource "tfe_team" "foobar" {
  name         = "team-test-%d"
  organization = "%s"
}

data "tfe_team" "foobar" {
  name         = tfe_team.foobar.name
  organization = "%s"
}`, rInt, organization, organization)
}

func testAccTFETeamDataSourceConfig_ssoTeamId(rInt int, organization string, ssoTeamID string) string {
	return fmt.Sprintf(`
resource "tfe_team" "sso_team" {
  name         = "team-test-%d"
  organization = "%s"
  sso_team_id  = "%s"
}

data "tfe_team" "sso_team" {
  name         = tfe_team.sso_team.name
  organization = tfe_team.sso_team.organization
}`, rInt, organization, ssoTeamID)
}

// TestAccTFETeamDataSource_scim validates SCIM attributes on the data source
// across three stages: no SCIM, SCIM enabled without a group link, and SCIM
// enabled with a group linked.
//
// FLAKE ALERT: SCIM settings are a singleton resource shared by the entire TFE
// instance. Running all SCIM cases inside one function — without calling
// t.Parallel in any sub-test — prevents concurrent tests from racing over the
// same singleton state.
//
// Should this test name ever change, you will also need to update the regex in ci.yml.
func TestAccTFESCIMTeamDataSource_omnibus(t *testing.T) {
	skipIfCloud(t)

	t.Run("SCIM attributes across full lifecycle", func(t *testing.T) {
		teamName := "tf-acc-scim-team-" + randomString(t)
		org := os.Getenv("TFE_ORGANIZATION")

		// teamID is captured from Terraform state after step 1 so that step 3
		// PreConfig can call linkSCIMGroupToTeam with the correct external ID.
		var teamID string

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			Steps: []resource.TestStep{
				// Step 1: No SCIM at all. Verify the data source reads successfully
				// and does not panic when SCIM fields are absent from the API response.
				{
					Config: testAccTFETeamDataSourceConfig_noSCIM(teamName, org),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("data.tfe_team.test", "id"),
						resource.TestCheckResourceAttr("data.tfe_team.test", "name", teamName),
						// Capture the team external ID for use in step 3.
						func(s *terraform.State) error {
							rs, ok := s.RootModule().Resources["tfe_team.test"]
							if !ok {
								return fmt.Errorf("tfe_team.test not found in state")
							}
							teamID = rs.Primary.ID
							return nil
						},
					),
				},
				// Step 2: SCIM enabled, team not yet linked to any SCIM group.
				// The API now returns SCIM fields (scim_enabled? is true), so we
				// assert their unlinked zero values.
				{
					Config: testAccTFETeamDataSourceConfig_scimEnabled(teamName, org),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.tfe_team.test", "scim_linked", "false"),
						resource.TestCheckNoResourceAttr("data.tfe_team.test", "scim_group_name"),
						resource.TestCheckResourceAttr("data.tfe_team.test", "scim_sync_paused", "false"),
						resource.TestCheckNoResourceAttr("data.tfe_team.test", "scim_updated_at"),
					),
				},
				// Step 3: Create a SCIM group out-of-band, link it to the team,
				// then verify the populated SCIM attributes on the data source.
				{
					PreConfig: func() {
						tokenName := "tf-acc-scim-token-" + randomString(t)
						token, err := testAccConfiguredClient.Client.Admin.Settings.SCIM.Tokens.Create(
							context.Background(), tokenName,
						)
						if err != nil {
							t.Fatalf("create SCIM token: %v", err)
						}
						t.Cleanup(func() {
							_ = testAccConfiguredClient.Client.Admin.Settings.SCIM.Tokens.Delete(
								context.Background(), token.ID)
						})

						// No explicit group cleanup: disabling SCIM (CheckDestroy) removes all groups.
						scimGroupID := createSCIMGroup(t, teamName, token.Token)
						linkSCIMGroupToTeam(t, teamID, scimGroupID)
					},
					Config: testAccTFETeamDataSourceConfig_scimEnabled(teamName, org),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.tfe_team.test", "scim_linked", "true"),
						resource.TestCheckResourceAttr("data.tfe_team.test", "scim_group_name", teamName),
						resource.TestCheckResourceAttr("data.tfe_team.test", "scim_sync_paused", "false"),
						resource.TestCheckResourceAttrSet("data.tfe_team.test", "scim_updated_at"),
					),
				},
			},
		})
	})
}

// testAccTFETeamDataSourceConfig_noSCIM returns a config with a team and its
// data source, with no SCIM resources present.
func testAccTFETeamDataSourceConfig_noSCIM(teamName, org string) string {
	return fmt.Sprintf(`
resource "tfe_team" "test" {
  name         = "%s"
  organization = "%s"
}

data "tfe_team" "test" {
  name         = tfe_team.test.name
  organization = tfe_team.test.organization
  depends_on   = [tfe_team.test]
}
`, teamName, org)
}

// testAccTFETeamDataSourceConfig_scimEnabled returns a config with SAML + SCIM
// enabled alongside the team and data source. Used for both the "SCIM on, not
// linked" and "SCIM on, linked" steps so the config is identical between them.
func testAccTFETeamDataSourceConfig_scimEnabled(teamName, org string) string {
	return fmt.Sprintf(`
%s

resource "tfe_scim_settings" "enable_scim" {
  depends_on = [tfe_saml_settings.enable_saml]
}

resource "tfe_team" "test" {
  name         = "%s"
  organization = "%s"
}

data "tfe_team" "test" {
  name         = tfe_team.test.name
  organization = tfe_team.test.organization
  depends_on   = [tfe_scim_settings.enable_scim]
}
`, testAccTFESCIMSettings_enableSAMLWithProviderType(scimTestSAMLSetting), teamName, org)
}
