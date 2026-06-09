// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestAccTFESCIMGroupMapping_omnibus is the single test function for all SCIM
// group mapping acceptance tests.
//
// FLAKE ALERT: SCIM settings are a singleton resource shared by the entire TFE
// instance. Every sub-test here enables SCIM (via an inline tfe_scim_settings
// block) as a prerequisite. Keeping all cases in one function with no
// t.Parallel call prevents concurrent tests from racing over that singleton.
//
// FLAKE ALERT (dual-singleton): This suite also contends with
// resource_tfe_saml_settings_test.go for the SAML singleton. Both singletons
// must be treated as exclusive resources: do not run SCIM and SAML acceptance
// tests concurrently.
//
// Keep this test name matching the SCIM acceptance-test prefix used by the
// skip regex in ci.yml (currently TestAccTFESCIM), or update that regex.
func TestAccTFESCIMGroupMapping_omnibus(t *testing.T) {
	skipIfCloud(t)

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	t.Run("full lifecycle: create, update paused, import", func(t *testing.T) {
		org, cleanupOrg := createScimGroupMappingOrganization(t, tfeClient)
		t.Cleanup(cleanupOrg)

		rand := randomString(t)
		teamName := "tf-acc-scim-map-" + rand
		tokenDescription := "scim group mapping lifecycle " + rand
		groupName := "tf-acc-scim-map-group-" + rand

		var scimToken string
		requireToken := func() {
			if scimToken == "" {
				t.Fatal("captured SCIM token value is empty")
			}
		}

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			CheckDestroy:             testAccTFESCIMGroupMappingDestroy,
			Steps: []resource.TestStep{
				// Enable SCIM and grab a token for the SCIM API.
				{
					Config: testAccTFESCIMGroupMapping_setup(org.Name, teamName, tokenDescription),
					Check:  captureSCIMTokenValue("tfe_scim_token.this", &scimToken),
				},
				// Create a SCIM group out-of-band, then map it to the team. Its
				// scim_group_id is resolved from the name by the tfe_scim_group
				// data source.
				{
					PreConfig: func() {
						requireToken()
						createSCIMGroup(t, groupName, scimToken)
					},
					Config: testAccTFESCIMGroupMapping_basic(org.Name, teamName, tokenDescription, groupName),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrPair(
							"tfe_scim_group_mapping.test", "team_id",
							"tfe_team.test", "id",
						),
						// scim_group_id matches the ID resolved by the data source.
						resource.TestCheckResourceAttrPair(
							"tfe_scim_group_mapping.test", "scim_group_id",
							"data.tfe_scim_group.test", "id",
						),
						resource.TestCheckResourceAttr("tfe_scim_group_mapping.test", "paused", "false"),
						// id mirrors team_id
						resource.TestCheckResourceAttrPair(
							"tfe_scim_group_mapping.test", "id",
							"tfe_team.test", "id",
						),
					),
				},
				// Re-apply must be a no-op.
				{
					Config:   testAccTFESCIMGroupMapping_basic(org.Name, teamName, tokenDescription, groupName),
					PlanOnly: true,
				},
				// Pause the mapping (in-place update).
				{
					Config: testAccTFESCIMGroupMapping_paused(org.Name, teamName, tokenDescription, groupName, true),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("tfe_scim_group_mapping.test", "paused", "true"),
						resource.TestCheckResourceAttrPair(
							"tfe_scim_group_mapping.test", "scim_group_id",
							"data.tfe_scim_group.test", "id",
						),
					),
				},
				// Unpause again (in-place update).
				{
					Config: testAccTFESCIMGroupMapping_basic(org.Name, teamName, tokenDescription, groupName),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("tfe_scim_group_mapping.test", "paused", "false"),
					),
				},
				// Import by team ID; Read resolves scim_group_id by group name.
				{
					ResourceName:      "tfe_scim_group_mapping.test",
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})

	t.Run("create starts paused", func(t *testing.T) {
		org, cleanupOrg := createScimGroupMappingOrganization(t, tfeClient)
		t.Cleanup(cleanupOrg)

		rand := randomString(t)
		teamName := "tf-acc-scim-map-paused-" + rand
		tokenDescription := "scim group mapping create paused " + rand
		groupName := "tf-acc-scim-map-paused-group-" + rand

		var scimToken string
		requireToken := func() {
			if scimToken == "" {
				t.Fatal("captured SCIM token value is empty")
			}
		}

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			CheckDestroy:             testAccTFESCIMGroupMappingDestroy,
			Steps: []resource.TestStep{
				// Enable SCIM and grab a token for the SCIM API.
				{
					Config: testAccTFESCIMGroupMapping_setup(org.Name, teamName, tokenDescription),
					Check:  captureSCIMTokenValue("tfe_scim_token.this", &scimToken),
				},
				// Create the SCIM group out-of-band, then map it to the team
				// with paused set to true on creation. Create always starts
				// unpaused, so the provider pauses it in a follow-up update.
				{
					PreConfig: func() {
						requireToken()
						createSCIMGroup(t, groupName, scimToken)
					},
					Config: testAccTFESCIMGroupMapping_paused(org.Name, teamName, tokenDescription, groupName, true),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("tfe_scim_group_mapping.test", "paused", "true"),
						resource.TestCheckResourceAttrPair(
							"tfe_scim_group_mapping.test", "scim_group_id",
							"data.tfe_scim_group.test", "id",
						),
					),
				},
				// Re-apply must be a no-op.
				{
					Config:   testAccTFESCIMGroupMapping_paused(org.Name, teamName, tokenDescription, groupName, true),
					PlanOnly: true,
				},
			},
		})
	})

	t.Run("out-of-band drift is detected and re-created", func(t *testing.T) {
		org, cleanupOrg := createScimGroupMappingOrganization(t, tfeClient)
		t.Cleanup(cleanupOrg)

		rand := randomString(t)
		teamName := "tf-acc-scim-map-drift-" + rand
		tokenDescription := "scim group mapping drift " + rand
		groupName := "tf-acc-scim-map-drift-group-" + rand

		var scimToken string
		var teamID string
		requireToken := func() {
			if scimToken == "" {
				t.Fatal("captured SCIM token value is empty")
			}
		}

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			CheckDestroy:             testAccTFESCIMGroupMappingDestroy,
			Steps: []resource.TestStep{
				// Enable SCIM and grab a token for the SCIM API.
				{
					Config: testAccTFESCIMGroupMapping_setup(org.Name, teamName, tokenDescription),
					Check:  captureSCIMTokenValue("tfe_scim_token.this", &scimToken),
				},
				// Create a SCIM group out-of-band and map it to the team.
				{
					PreConfig: func() {
						requireToken()
						createSCIMGroup(t, groupName, scimToken)
					},
					Config: testAccTFESCIMGroupMapping_basic(org.Name, teamName, tokenDescription, groupName),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrPair(
							"tfe_scim_group_mapping.test", "scim_group_id",
							"data.tfe_scim_group.test", "id",
						),
						func(s *terraform.State) error {
							rs, ok := s.RootModule().Resources["tfe_scim_group_mapping.test"]
							if !ok {
								return fmt.Errorf("tfe_scim_group_mapping.test not found in state")
							}
							teamID = rs.Primary.ID
							return nil
						},
					),
				},
				// Unlink the mapping out-of-band; Read clears state and the
				// re-apply re-creates it.
				{
					PreConfig: func() {
						if err := tfeClient.Admin.Settings.SCIM.SCIMGroupMappings.Delete(ctx, teamID); err != nil {
							t.Fatalf("delete SCIM group mapping out-of-band: %v", err)
						}
					},
					Config: testAccTFESCIMGroupMapping_basic(org.Name, teamName, tokenDescription, groupName),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrPair(
							"tfe_scim_group_mapping.test", "scim_group_id",
							"data.tfe_scim_group.test", "id",
						),
						resource.TestCheckResourceAttr("tfe_scim_group_mapping.test", "paused", "false"),
					),
				},
			},
		})
	})

	t.Run("team_id change: old team is unlinked, new team is linked", func(t *testing.T) {
		org, cleanupOrg := createScimGroupMappingOrganization(t, tfeClient)
		t.Cleanup(cleanupOrg)

		rand := randomString(t)
		teamAName := "tf-acc-scim-map-team-a-" + rand
		teamBName := "tf-acc-scim-map-team-b-" + rand
		tokenDescription := "scim group mapping team swap " + rand
		groupName := "tf-acc-scim-map-swap-group-" + rand

		var scimToken string
		var teamAID string
		requireToken := func() {
			if scimToken == "" {
				t.Fatal("captured SCIM token value is empty")
			}
		}

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			CheckDestroy:             testAccTFESCIMGroupMappingDestroy,
			Steps: []resource.TestStep{
				// Enable SCIM and grab a token.
				{
					Config: testAccTFESCIMGroupMapping_setupTwoTeams(org.Name, teamAName, teamBName, tokenDescription),
					Check:  captureSCIMTokenValue("tfe_scim_token.this", &scimToken),
				},
				// Create the SCIM group out-of-band and map it to Team A.
				{
					PreConfig: func() {
						requireToken()
						createSCIMGroup(t, groupName, scimToken)
					},
					Config: testAccTFESCIMGroupMapping_twoTeamsMapped(org.Name, teamAName, teamBName, tokenDescription, groupName, "tfe_team.team_a"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrPair(
							"tfe_scim_group_mapping.test", "team_id",
							"tfe_team.team_a", "id",
						),
						resource.TestCheckResourceAttrPair(
							"tfe_scim_group_mapping.test", "scim_group_id",
							"data.tfe_scim_group.test", "id",
						),
						// Capture Team A's ID to verify it is unlinked in the next step.
						func(s *terraform.State) error {
							rs, ok := s.RootModule().Resources["tfe_team.team_a"]
							if !ok {
								return fmt.Errorf("tfe_team.team_a not found in state")
							}
							teamAID = rs.Primary.ID
							return nil
						},
					),
				},
				// Re-point the mapping to Team B. The plan should be a replace
				// (RequiresReplace on team_id); then verify Team B is linked and
				// Team A is unlinked via the Teams API.
				{
					Config: testAccTFESCIMGroupMapping_twoTeamsMapped(org.Name, teamAName, teamBName, tokenDescription, groupName, "tfe_team.team_b"),
					ConfigPlanChecks: resource.ConfigPlanChecks{
						PreApply: []plancheck.PlanCheck{
							plancheck.ExpectResourceAction(
								"tfe_scim_group_mapping.test",
								plancheck.ResourceActionDestroyBeforeCreate,
							),
						},
					},
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrPair(
							"tfe_scim_group_mapping.test", "team_id",
							"tfe_team.team_b", "id",
						),
						resource.TestCheckResourceAttrPair(
							"tfe_scim_group_mapping.test", "scim_group_id",
							"data.tfe_scim_group.test", "id",
						),
						// Verify Team A is no longer SCIM-linked.
						func(_ *terraform.State) error {
							if teamAID == "" {
								return fmt.Errorf("teamAID was not captured from previous step")
							}
							team, err := tfeClient.Teams.Read(ctx, teamAID)
							if err != nil {
								return fmt.Errorf("reading team A (%s): %w", teamAID, err)
							}
							if team.SCIMLinked != nil && *team.SCIMLinked {
								return fmt.Errorf("team A (%s) is still SCIM-linked after team_id was changed to team B", teamAID)
							}
							return nil
						},
					),
				},
			},
		})
	})

	t.Run("scim_group_id change: mapping is re-created with new SCIM group", func(t *testing.T) {
		org, cleanupOrg := createScimGroupMappingOrganization(t, tfeClient)
		t.Cleanup(cleanupOrg)

		rand := randomString(t)
		teamName := "tf-acc-scim-map-grpswap-" + rand
		tokenDescription := "scim group mapping group swap " + rand
		groupName1 := "tf-acc-scim-map-grpswap-1-" + rand
		groupName2 := "tf-acc-scim-map-grpswap-2-" + rand

		var scimToken string
		requireToken := func() {
			if scimToken == "" {
				t.Fatal("captured SCIM token value is empty")
			}
		}

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			CheckDestroy:             testAccTFESCIMGroupMappingDestroy,
			Steps: []resource.TestStep{
				// Enable SCIM and grab a token.
				{
					Config: testAccTFESCIMGroupMapping_setup(org.Name, teamName, tokenDescription),
					Check:  captureSCIMTokenValue("tfe_scim_token.this", &scimToken),
				},
				// Create both SCIM groups up-front and map the team to group 1.
				{
					PreConfig: func() {
						requireToken()
						createSCIMGroup(t, groupName1, scimToken)
						createSCIMGroup(t, groupName2, scimToken)
					},
					Config: testAccTFESCIMGroupMapping_basic(org.Name, teamName, tokenDescription, groupName1),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrPair(
							"tfe_scim_group_mapping.test", "scim_group_id",
							"data.tfe_scim_group.test", "id",
						),
						resource.TestCheckResourceAttr("tfe_scim_group_mapping.test", "paused", "false"),
					),
				},
				// Switch the mapping to group 2. The plan should be a replace
				// (RequiresReplace on scim_group_id); then verify the mapping
				// now points at group 2.
				{
					Config: testAccTFESCIMGroupMapping_basic(org.Name, teamName, tokenDescription, groupName2),
					ConfigPlanChecks: resource.ConfigPlanChecks{
						PreApply: []plancheck.PlanCheck{
							plancheck.ExpectResourceAction(
								"tfe_scim_group_mapping.test",
								plancheck.ResourceActionDestroyBeforeCreate,
							),
						},
					},
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrPair(
							"tfe_scim_group_mapping.test", "scim_group_id",
							"data.tfe_scim_group.test", "id",
						),
						resource.TestCheckResourceAttr("tfe_scim_group_mapping.test", "paused", "false"),
					),
				},
			},
		})
	})

	t.Run("validation: empty config arguments", func(t *testing.T) {
		lengthErr := regexp.MustCompile(`(?s)Invalid Attribute Value Length|at least 1`)

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			Steps: []resource.TestStep{
				{
					Config:      testAccTFESCIMGroupMappingNoArgs(),
					ExpectError: regexp.MustCompile(`(?s)Missing required argument|argument "team_id" is required|argument "scim_group_id" is required`),
					PlanOnly:    true,
				},
				{
					Config:      testAccTFESCIMGroupMappingEmptyTeamID(),
					ExpectError: lengthErr,
					PlanOnly:    true,
				},
				{
					Config:      testAccTFESCIMGroupMappingEmptySCIMGroupID(),
					ExpectError: lengthErr,
					PlanOnly:    true,
				},
			},
		})
	})
}

// createScimGroupMappingOrganization creates a plain organization for the SCIM
// group mapping tests. createBusinessOrganization fails on a real TFE instance
// (its subscription upgrade is unsupported there), so we use createOrganization
// directly instead.
func createScimGroupMappingOrganization(t *testing.T, client *tfe.Client) (*tfe.Organization, func()) {
	return createOrganization(t, client, tfe.OrganizationCreateOptions{
		Name:  tfe.String("tst-" + randomString(t)),
		Email: tfe.String(fmt.Sprintf("%s@tfe.local", randomString(t))),
	})
}

// testAccTFESCIMGroupMappingDestroy verifies all tfe_scim_group_mapping
// resources have been unlinked from their teams.
func testAccTFESCIMGroupMappingDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_scim_group_mapping" {
			continue
		}

		// The id is the team id.
		team, err := testAccConfiguredClient.Client.Teams.Read(ctx, rs.Primary.ID)
		if err != nil {
			// The team being gone means the mapping is gone too.
			if errors.Is(err, tfe.ErrResourceNotFound) {
				continue
			}
			return fmt.Errorf("unexpected error checking team %s: %w", rs.Primary.ID, err)
		}
		if team.SCIMLinked != nil && *team.SCIMLinked {
			return fmt.Errorf("team %s is still linked to a SCIM group", rs.Primary.ID)
		}
	}
	return nil
}

// testAccTFESCIMGroupMapping_setup enables SAML + SCIM, creates a SCIM token,
// and provisions an organization team that a SCIM group can be mapped to.
func testAccTFESCIMGroupMapping_setup(orgName, teamName, tokenDescription string) string {
	return fmt.Sprintf(`
%s

data "tfe_organization" "org" {
  name = "%s"
}

resource "tfe_team" "test" {
  name         = "%s"
  organization = data.tfe_organization.org.name
}

resource "tfe_scim_settings" "enable_scim" {
  depends_on = [tfe_saml_settings.enable_saml]
}

resource "tfe_scim_token" "this" {
  description = "%s"
  depends_on  = [tfe_scim_settings.enable_scim]
}
`, testAccTFESCIMSettings_enableSAMLWithProviderType(scimTestSAMLSetting), orgName, teamName, tokenDescription)
}

// testAccTFESCIMGroupMapping_basic maps the team to the SCIM group named
// groupName (which must already exist) with paused set to false.
func testAccTFESCIMGroupMapping_basic(orgName, teamName, tokenDescription, groupName string) string {
	return testAccTFESCIMGroupMapping_paused(orgName, teamName, tokenDescription, groupName, false)
}

// testAccTFESCIMGroupMapping_paused is like the basic config but sets an
// explicit paused state.
func testAccTFESCIMGroupMapping_paused(orgName, teamName, tokenDescription, groupName string, paused bool) string {
	return fmt.Sprintf(`
%s

data "tfe_scim_group" "test" {
  name       = "%s"
  depends_on = [tfe_scim_token.this]
}

resource "tfe_scim_group_mapping" "test" {
  team_id       = tfe_team.test.id
  scim_group_id = data.tfe_scim_group.test.id
  paused        = %t
}
`, testAccTFESCIMGroupMapping_setup(orgName, teamName, tokenDescription), groupName, paused)
}

// testAccTFESCIMGroupMapping_setupTwoTeams is like testAccTFESCIMGroupMapping_setup
// but provisions two teams (team_a and team_b) for the team_id-change scenario.
func testAccTFESCIMGroupMapping_setupTwoTeams(orgName, teamAName, teamBName, tokenDescription string) string {
	return fmt.Sprintf(`
%s

data "tfe_organization" "org" {
  name = "%s"
}

resource "tfe_team" "team_a" {
  name         = "%s"
  organization = data.tfe_organization.org.name
}

resource "tfe_team" "team_b" {
  name         = "%s"
  organization = data.tfe_organization.org.name
}

resource "tfe_scim_settings" "enable_scim" {
  depends_on = [tfe_saml_settings.enable_saml]
}

resource "tfe_scim_token" "this" {
  description = "%s"
  depends_on  = [tfe_scim_settings.enable_scim]
}
`, testAccTFESCIMSettings_enableSAMLWithProviderType(scimTestSAMLSetting), orgName, teamAName, teamBName, tokenDescription)
}

// testAccTFESCIMGroupMapping_twoTeamsMapped maps the SCIM group to the team
// referenced by teamRef (e.g. "tfe_team.team_a"). Both teams come from the
// embedded setup block, so one helper covers the initial mapping and the
// team_id swap.
func testAccTFESCIMGroupMapping_twoTeamsMapped(orgName, teamAName, teamBName, tokenDescription, groupName, teamRef string) string {
	return fmt.Sprintf(`
%s

data "tfe_scim_group" "test" {
  name       = "%s"
  depends_on = [tfe_scim_token.this]
}

resource "tfe_scim_group_mapping" "test" {
  team_id       = %s.id
  scim_group_id = data.tfe_scim_group.test.id
}
`, testAccTFESCIMGroupMapping_setupTwoTeams(orgName, teamAName, teamBName, tokenDescription), groupName, teamRef)
}

// testAccTFESCIMGroupMappingNoArgs returns a resource block with no arguments,
// which must fail because both team_id and scim_group_id are required.
func testAccTFESCIMGroupMappingNoArgs() string {
	return `resource "tfe_scim_group_mapping" "test" {}`
}

// testAccTFESCIMGroupMappingEmptyTeamID returns a resource block with an empty
// team_id, which must fail the LengthAtLeast(1) validator.
func testAccTFESCIMGroupMappingEmptyTeamID() string {
	return `
resource "tfe_scim_group_mapping" "test" {
  team_id       = ""
  scim_group_id = "some-group-id"
}
`
}

// testAccTFESCIMGroupMappingEmptySCIMGroupID returns a resource block with an
// empty scim_group_id, which must fail the LengthAtLeast(1) validator.
func testAccTFESCIMGroupMappingEmptySCIMGroupID() string {
	return `
resource "tfe_scim_group_mapping" "test" {
  team_id       = "some-team-id"
  scim_group_id = ""
}
`
}
