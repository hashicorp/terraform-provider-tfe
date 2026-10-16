// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	tfev2 "github.com/hashicorp/go-tfe/v2"
	"github.com/hashicorp/go-tfe/v2/api/models"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testTFPolicyVersionResourceName = "tfe_tfpolicy_version.foobar"

type testTFPolicyBuild struct {
	arch string
	sha  string
	url  string
}

func (b testTFPolicyBuild) downloadURL() string {
	if b.url != "" {
		return b.url
	}
	return testTFPolicyPluginURL(b.arch)
}

// TestAccTFETFPolicyVersion_basic creates single-architecture releases without amd64, and a
// release with every supported architecture.
func TestAccTFETFPolicyVersion_basic(t *testing.T) {
	skipUnlessTFPolicyVersionsSupported(t)

	for name, archs := range map[string][]string{
		"amd64 only":        {"amd64"},
		"arm64 only":        {"arm64"},
		"s390x only":        {"s390x"},
		"all architectures": {"amd64", "arm64", "s390x"},
	} {
		t.Run(name, func(t *testing.T) {
			version := genSafeRandomTFPolicyVersion()
			builds := make([]testTFPolicyBuild, 0, len(archs))
			checks := []resource.TestCheckFunc{
				testAccCheckTFETFPolicyVersionAPI(testTFPolicyVersionResourceName, version, false, archs...),
				resource.TestMatchResourceAttr(testTFPolicyVersionResourceName, "id", regexp.MustCompile(`^tool-.+`)),
				resource.TestCheckResourceAttr(testTFPolicyVersionResourceName, "version", version),
				resource.TestCheckResourceAttr(testTFPolicyVersionResourceName, "official", "false"),
				resource.TestCheckResourceAttr(testTFPolicyVersionResourceName, "enabled", "true"),
				resource.TestCheckResourceAttr(testTFPolicyVersionResourceName, "beta", "false"),
				resource.TestCheckResourceAttr(testTFPolicyVersionResourceName, "deprecated", "false"),
				resource.TestCheckNoResourceAttr(testTFPolicyVersionResourceName, "deprecated_reason"),
				resource.TestCheckNoResourceAttr(testTFPolicyVersionResourceName, "url"),
				resource.TestCheckNoResourceAttr(testTFPolicyVersionResourceName, "sha"),
				resource.TestCheckResourceAttr(testTFPolicyVersionResourceName, "archs.#", fmt.Sprint(len(archs))),
			}
			for _, arch := range archs {
				build := testTFPolicyBuild{arch: arch, sha: genSha(t, "tfpolicy", arch)}
				builds = append(builds, build)
				checks = append(checks, testAccCheckTFETFPolicyVersionBuild(build))
			}

			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { testAccPreCheck(t) },
				ProtoV6ProviderFactories: testAccMuxedProviders,
				CheckDestroy:             testAccCheckTFETFPolicyVersionDestroy,
				Steps: []resource.TestStep{
					{
						Config: testAccTFETFPolicyVersionConfig(version, "", builds...),
						Check:  resource.ComposeAggregateTestCheckFunc(checks...),
					},
				},
			})
		})
	}
}

func TestAccTFETFPolicyVersion_update(t *testing.T) {
	skipUnlessTFPolicyVersionsSupported(t)

	version := genSafeRandomTFPolicyVersion()
	renamed := version + "-rc1"
	amd64 := testTFPolicyBuild{arch: "amd64", sha: genSha(t, "tfpolicy", "amd64")}
	amd64Rebuilt := testTFPolicyBuild{arch: "amd64", sha: genSha(t, "tfpolicy", "amd64-rebuilt"), url: "https://example.com/tfpolicy-plugin/rebuilt/tfpolicy-plugin_linux_amd64.zip"}
	arm64 := testTFPolicyBuild{arch: "arm64", sha: genSha(t, "tfpolicy", "arm64")}
	s390x := testTFPolicyBuild{arch: "s390x", sha: genSha(t, "tfpolicy", "s390x")}
	sameID := statecheck.CompareValue(compare.ValuesSame())
	idPath := tfjsonpath.New("id")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFETFPolicyVersionDestroy,
		Steps: []resource.TestStep{
			{
				Config:            testAccTFETFPolicyVersionConfig(version, "", amd64),
				ConfigStateChecks: []statecheck.StateCheck{sameID.AddStateValue(testTFPolicyVersionResourceName, idPath)},
			},
			{
				// Rename, change flags, rebuild amd64 with a new URL and checksum, and add arm64 in place.
				Config: testAccTFETFPolicyVersionConfig(renamed, "  beta = true\n  enabled = false", amd64Rebuilt, arm64),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(testTFPolicyVersionResourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{sameID.AddStateValue(testTFPolicyVersionResourceName, idPath)},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTFETFPolicyVersionAPI(testTFPolicyVersionResourceName, renamed, false, "amd64", "arm64"),
					resource.TestCheckResourceAttr(testTFPolicyVersionResourceName, "version", renamed),
					resource.TestCheckResourceAttr(testTFPolicyVersionResourceName, "beta", "true"),
					resource.TestCheckResourceAttr(testTFPolicyVersionResourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(testTFPolicyVersionResourceName, "archs.#", "2"),
					testAccCheckTFETFPolicyVersionBuild(amd64Rebuilt),
					testAccCheckTFETFPolicyVersionBuild(arm64),
				),
			},
			{
				// archs replaces the full set.
				Config:            testAccTFETFPolicyVersionConfig(renamed, "  beta = true\n  enabled = false", s390x),
				ConfigStateChecks: []statecheck.StateCheck{sameID.AddStateValue(testTFPolicyVersionResourceName, idPath)},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTFETFPolicyVersionAPI(testTFPolicyVersionResourceName, renamed, false, "s390x"),
					resource.TestCheckResourceAttr(testTFPolicyVersionResourceName, "archs.#", "1"),
					testAccCheckTFETFPolicyVersionBuild(s390x),
				),
			},
			{
				// Flag-only change with unchanged archs keeps the existing builds.
				Config:            testAccTFETFPolicyVersionConfig(renamed, "", s390x),
				ConfigStateChecks: []statecheck.StateCheck{sameID.AddStateValue(testTFPolicyVersionResourceName, idPath)},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTFETFPolicyVersionAPI(testTFPolicyVersionResourceName, renamed, false, "s390x"),
					resource.TestCheckResourceAttr(testTFPolicyVersionResourceName, "beta", "false"),
					resource.TestCheckResourceAttr(testTFPolicyVersionResourceName, "enabled", "true"),
					testAccCheckTFETFPolicyVersionBuild(s390x),
				),
			},
		},
	})
}

func TestAccTFETFPolicyVersion_deprecation(t *testing.T) {
	skipUnlessTFPolicyVersionsSupported(t)

	version := genSafeRandomTFPolicyVersion()
	renamed := version + "-rc1"
	amd64 := testTFPolicyBuild{arch: "amd64", sha: genSha(t, "tfpolicy", "amd64")}
	reasonPath := tfjsonpath.New("deprecated_reason")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFETFPolicyVersionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETFPolicyVersionConfig(version, "  deprecated = true\n  deprecated_reason = \"first\"", amd64),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTFETFPolicyVersionAPI(testTFPolicyVersionResourceName, version, true, "amd64"),
					resource.TestCheckResourceAttr(testTFPolicyVersionResourceName, "deprecated", "true"),
					resource.TestCheckResourceAttr(testTFPolicyVersionResourceName, "deprecated_reason", "first"),
				),
			},
			{
				Config: testAccTFETFPolicyVersionConfig(version, "  deprecated = true\n  deprecated_reason = \"second\"", amd64),
				Check:  resource.TestCheckResourceAttr(testTFPolicyVersionResourceName, "deprecated_reason", "second"),
			},
			{
				// Undeprecating with an omitted reason plans a null reason.
				Config: testAccTFETFPolicyVersionConfig(version, "  deprecated = false", amd64),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectKnownValue(testTFPolicyVersionResourceName, reasonPath, knownvalue.Null()),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTFETFPolicyVersionAPI(testTFPolicyVersionResourceName, version, false, "amd64"),
					resource.TestCheckNoResourceAttr(testTFPolicyVersionResourceName, "deprecated_reason"),
				),
			},
			{
				// Deprecating an existing version with an explicit reason.
				Config: testAccTFETFPolicyVersionConfig(version, "  deprecated = true\n  deprecated_reason = \"third\"", amd64),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTFETFPolicyVersionAPI(testTFPolicyVersionResourceName, version, true, "amd64"),
					resource.TestCheckResourceAttr(testTFPolicyVersionResourceName, "deprecated_reason", "third"),
				),
			},
			{
				// Changing another field on a deprecated version keeps the reason.
				Config: testAccTFETFPolicyVersionConfig(version, "  enabled = false\n  deprecated = true\n  deprecated_reason = \"third\"", amd64),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(testTFPolicyVersionResourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(testTFPolicyVersionResourceName, reasonPath, knownvalue.StringExact("third")),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testTFPolicyVersionResourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(testTFPolicyVersionResourceName, "deprecated_reason", "third"),
				),
			},
			{
				// An explicit null reason behaves like an omitted one.
				Config: testAccTFETFPolicyVersionConfig(version, "  deprecated = false\n  deprecated_reason = null", amd64),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectKnownValue(testTFPolicyVersionResourceName, reasonPath, knownvalue.Null()),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTFETFPolicyVersionAPI(testTFPolicyVersionResourceName, version, false, "amd64"),
					resource.TestCheckNoResourceAttr(testTFPolicyVersionResourceName, "deprecated_reason"),
				),
			},
			{
				// Renaming while deprecating sends the configured reason, not the server default.
				Config: testAccTFETFPolicyVersionConfig(renamed, "  deprecated = true\n  deprecated_reason = \"fourth\"", amd64),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTFETFPolicyVersionAPI(testTFPolicyVersionResourceName, renamed, true, "amd64"),
					resource.TestCheckResourceAttr(testTFPolicyVersionResourceName, "version", renamed),
					resource.TestCheckResourceAttr(testTFPolicyVersionResourceName, "deprecated_reason", "fourth"),
				),
			},
			{
				// An empty reason is accepted when not deprecated and stays stable.
				Config: testAccTFETFPolicyVersionConfig(renamed, "  deprecated = false\n  deprecated_reason = \"\"", amd64),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTFETFPolicyVersionAPI(testTFPolicyVersionResourceName, renamed, false, "amd64"),
					resource.TestCheckResourceAttr(testTFPolicyVersionResourceName, "deprecated_reason", ""),
				),
			},
		},
	})
}

func TestAccTFETFPolicyVersion_import(t *testing.T) {
	skipUnlessTFPolicyVersionsSupported(t)

	version := genSafeRandomTFPolicyVersion()
	amd64 := testTFPolicyBuild{arch: "amd64", sha: genSha(t, "tfpolicy", "amd64")}
	arm64 := testTFPolicyBuild{arch: "arm64", sha: genSha(t, "tfpolicy", "arm64")}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFETFPolicyVersionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETFPolicyVersionConfig(version, "  deprecated = true\n  deprecated_reason = \"imported\"", amd64, arm64),
			},
			{
				ResourceName:      testTFPolicyVersionResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      testTFPolicyVersionResourceName,
				ImportState:       true,
				ImportStateId:     version,
				ImportStateVerify: true,
			},
			{
				ResourceName:  testTFPolicyVersionResourceName,
				ImportState:   true,
				ImportStateId: "tool-",
				ExpectError:   regexp.MustCompile(`Invalid Terraform Policy Plugin Version Import`),
			},
			{
				ResourceName:  testTFPolicyVersionResourceName,
				ImportState:   true,
				ImportStateId: version + "-missing",
				ExpectError:   regexp.MustCompile(`was not\s+found`),
			},
		},
	})
}

func TestAccTFETFPolicyVersion_recreatedWhenDeletedOutsideTerraform(t *testing.T) {
	skipUnlessTFPolicyVersionsSupported(t)

	version := genSafeRandomTFPolicyVersion()
	amd64 := testTFPolicyBuild{arch: "amd64", sha: genSha(t, "tfpolicy", "amd64")}
	newID := statecheck.CompareValue(compare.ValuesDiffer())
	var id string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFETFPolicyVersionDestroy,
		Steps: []resource.TestStep{
			{
				Config:            testAccTFETFPolicyVersionConfig(version, "", amd64),
				ConfigStateChecks: []statecheck.StateCheck{newID.AddStateValue(testTFPolicyVersionResourceName, tfjsonpath.New("id"))},
				Check: func(s *terraform.State) error {
					id = s.RootModule().Resources[testTFPolicyVersionResourceName].Primary.ID
					return nil
				},
			},
			{
				PreConfig: func() {
					if err := testAccConfiguredClient.ClientV2.API.Admin().TfpolicyVersions().ById(id).Delete(ctx, nil); err != nil {
						t.Fatalf("failed to delete Terraform policy plugin version %s: %s", id, apiErrorDetail(err))
					}
				},
				Config:            testAccTFETFPolicyVersionConfig(version, "", amd64),
				ConfigStateChecks: []statecheck.StateCheck{newID.AddStateValue(testTFPolicyVersionResourceName, tfjsonpath.New("id"))},
				Check:             testAccCheckTFETFPolicyVersionAPI(testTFPolicyVersionResourceName, version, false, "amd64"),
			},
		},
	})
}

func TestAccTFETFPolicyVersion_archsChangedOutsideTerraform(t *testing.T) {
	skipUnlessTFPolicyVersionsSupported(t)

	version := genSafeRandomTFPolicyVersion()
	amd64 := testTFPolicyBuild{arch: "amd64", sha: genSha(t, "tfpolicy", "amd64")}
	arm64 := testTFPolicyBuild{arch: "arm64", sha: genSha(t, "tfpolicy", "arm64")}
	s390x := testTFPolicyBuild{arch: "s390x", sha: genSha(t, "tfpolicy", "s390x")}
	sameID := statecheck.CompareValue(compare.ValuesSame())
	var id string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFETFPolicyVersionDestroy,
		Steps: []resource.TestStep{
			{
				Config:            testAccTFETFPolicyVersionConfig(version, "", amd64, arm64),
				ConfigStateChecks: []statecheck.StateCheck{sameID.AddStateValue(testTFPolicyVersionResourceName, tfjsonpath.New("id"))},
				Check: func(s *terraform.State) error {
					id = s.RootModule().Resources[testTFPolicyVersionResourceName].Primary.ID
					return nil
				},
			},
			{
				PreConfig: func() { testAccReplaceTFPolicyVersionArchs(t, id, s390x) },
				Config:    testAccTFETFPolicyVersionConfig(version, "", amd64, arm64),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(testTFPolicyVersionResourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{sameID.AddStateValue(testTFPolicyVersionResourceName, tfjsonpath.New("id"))},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTFETFPolicyVersionAPI(testTFPolicyVersionResourceName, version, false, "amd64", "arm64"),
					resource.TestCheckResourceAttr(testTFPolicyVersionResourceName, "archs.#", "2"),
					testAccCheckTFETFPolicyVersionBuild(amd64),
					testAccCheckTFETFPolicyVersionBuild(arm64),
				),
			},
		},
	})
}

func TestAccTFETFPolicyVersion_validation(t *testing.T) {
	skipUnlessTFPolicyVersionsSupported(t)

	version := genSafeRandomTFPolicyVersion()
	amd64 := testTFPolicyBuild{arch: "amd64", sha: genSha(t, "tfpolicy", "amd64")}
	sha := amd64.sha
	rawBuild := func(url, sha, os, arch string) string {
		return fmt.Sprintf(`
resource "tfe_tfpolicy_version" "foobar" {
  version = %q
  archs = [{
    url  = %q
    sha  = %q
    os   = %q
    arch = %q
  }]
}`, version, url, sha, os, arch)
	}
	url := testTFPolicyPluginURL("amd64")

	cases := []struct {
		name   string
		config string
		err    string
	}{
		{"missing archs", fmt.Sprintf("resource \"tfe_tfpolicy_version\" \"foobar\" {\n  version = %q\n}", version), `"archs" is\s+required`},
		{"empty archs", fmt.Sprintf("resource \"tfe_tfpolicy_version\" \"foobar\" {\n  version = %q\n  archs = []\n}", version), `at least 1`},
		{"duplicate os and arch", testAccTFETFPolicyVersionConfig(version, "", amd64, testTFPolicyBuild{arch: "amd64", sha: genSha(t, "tfpolicy", "other")}), `Duplicate Architecture`},
		{"non-http url", rawBuild("ftp://example.com/tfpolicy-plugin/tfpolicy-plugin_linux_amd64.zip", sha, "linux", "amd64"), `Invalid Attribute Value`},
		{"uppercase sha", rawBuild(url, strings.ToUpper(sha), "linux", "amd64"), `must be a SHA-256\s+checksum`},
		{"short sha", rawBuild(url, sha[:63], "linux", "amd64"), `must be a SHA-256\s+checksum`},
		{"unsupported os", rawBuild(url, sha, "darwin", "amd64"), `value must be one of`},
		{"unsupported arch", rawBuild(url, sha, "linux", "386"), `value must be one of`},
		{"reason when not deprecated", testAccTFETFPolicyVersionConfig(version, "  deprecated_reason = \"old\"", amd64), `Invalid Deprecation Reason`},
		{"blank reason when deprecated", testAccTFETFPolicyVersionConfig(version, "  deprecated = true\n  deprecated_reason = \"  \"", amd64), `Invalid Deprecation Reason`},
		{"missing reason when deprecated", testAccTFETFPolicyVersionConfig(version, "  deprecated = true", amd64), `Invalid Deprecation Reason`},
		{"top-level url", testAccTFETFPolicyVersionConfig(version, fmt.Sprintf("  url = %q", url), amd64), `argument named "url" is not\s+expected`},
		{"top-level sha", testAccTFETFPolicyVersionConfig(version, fmt.Sprintf("  sha = %q", sha), amd64), `argument named "sha" is not\s+expected`},
	}

	steps := make([]resource.TestStep, 0, len(cases))
	for _, tc := range cases {
		steps = append(steps, resource.TestStep{
			Config:      tc.config,
			PlanOnly:    true,
			ExpectError: regexp.MustCompile(tc.err),
		})
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps:                    steps,
	})
}

func TestTFPolicyVersionAttributes(t *testing.T) {
	ctx := context.Background()

	t.Run("maps plan to request", func(t *testing.T) {
		plan := testTFPolicyVersionPlan(t, "amd64", "arm64", "s390x")
		plan.Deprecated = types.BoolValue(true)
		plan.DeprecatedReason = types.StringValue("old")

		attributes, diags := tfpolicyVersionAttributes(ctx, plan)
		require.False(t, diags.HasError(), "%v", diags)
		assert.Equal(t, "1.0.0", valueOrZero(attributes.GetVersion()))
		assert.False(t, valueOrZero(attributes.GetOfficial()))
		assert.True(t, valueOrZero(attributes.GetEnabled()))
		assert.False(t, valueOrZero(attributes.GetBeta()))
		assert.True(t, valueOrZero(attributes.GetDeprecated()))
		assert.Equal(t, "old", valueOrZero(attributes.GetDeprecatedReason()))

		var archs []string
		for _, build := range attributes.GetArchs() {
			assert.Equal(t, models.LINUX_ADMINTFPOLICYVERSIONARCHITECTURE_OS, *build.GetOs())
			assert.Equal(t, testTFPolicyPluginURL(build.GetArch().String()), valueOrZero(build.GetUrl()))
			assert.Equal(t, strings.Repeat("a", 64), valueOrZero(build.GetSha()))
			archs = append(archs, build.GetArch().String())
		}
		assert.ElementsMatch(t, []string{"amd64", "arm64", "s390x"}, archs)
	})

	t.Run("omits a null reason", func(t *testing.T) {
		attributes, diags := tfpolicyVersionAttributes(ctx, testTFPolicyVersionPlan(t, "amd64"))
		require.False(t, diags.HasError(), "%v", diags)
		assert.Nil(t, attributes.GetDeprecatedReason())
	})

	t.Run("rejects invalid plan", func(t *testing.T) {
		other := testToolArchitecture("linux", "amd64")
		other.Sha = types.StringValue(strings.Repeat("b", 64))
		archs, diags := toolArchitecturesToSet(ctx, []ToolArchitecture{testToolArchitecture("linux", "amd64"), other})
		require.False(t, diags.HasError(), "%v", diags)
		plan := testTFPolicyVersionPlan(t, "amd64")
		plan.Archs = archs
		plan.Version = types.StringUnknown()
		plan.Beta = types.BoolUnknown()

		attributes, diags := tfpolicyVersionAttributes(ctx, plan)
		assert.Nil(t, attributes)
		var summaries []string
		for _, d := range diags.Errors() {
			summaries = append(summaries, d.Summary())
		}
		assert.ElementsMatch(t, []string{"Invalid Version", "Invalid Release Flag", "Duplicate Architecture"}, summaries)
	})
}

func TestTFPolicyVersionModel(t *testing.T) {
	ctx := context.Background()

	t.Run("maps response to state", func(t *testing.T) {
		state, err := tfpolicyVersionModel(ctx, testTFPolicyVersionResponse("tool-1", nil), "tool-1")
		require.NoError(t, err)
		assert.Equal(t, "tool-1", state.ID.ValueString())
		assert.Equal(t, "1.0.0", state.Version.ValueString())
		assert.True(t, state.Enabled.ValueBool())
		assert.False(t, state.Deprecated.ValueBool())
		assert.True(t, state.DeprecatedReason.IsNull())

		archs, diags := toolArchitecturesFromSet(ctx, state.Archs)
		require.False(t, diags.HasError(), "%v", diags)
		assert.ElementsMatch(t, []ToolArchitecture{testToolArchitecture("linux", "amd64"), testToolArchitecture("linux", "arm64")}, archs)
	})

	t.Run("ignores reason when not deprecated", func(t *testing.T) {
		state, err := tfpolicyVersionModel(ctx, testTFPolicyVersionResponse("tool-1", func(a *models.AdminTfpolicyVersions_attributes) {
			a.SetDeprecatedReason(ptr("stale"))
		}), "")
		require.NoError(t, err)
		assert.True(t, state.DeprecatedReason.IsNull())
	})

	t.Run("keeps reason when deprecated", func(t *testing.T) {
		state, err := tfpolicyVersionModel(ctx, testTFPolicyVersionResponse("tool-1", func(a *models.AdminTfpolicyVersions_attributes) {
			a.SetDeprecated(ptr(true))
			a.SetDeprecatedReason(ptr("old"))
		}), "")
		require.NoError(t, err)
		assert.Equal(t, "old", state.DeprecatedReason.ValueString())
	})

	invalid := map[string]struct {
		id     string
		expect string
		mutate func(*models.AdminTfpolicyVersions_attributes)
	}{
		"unexpected ID":           {"tool-2", "tool-1", nil},
		"non-tool ID":             {"ws-1", "", nil},
		"missing flag":            {"tool-1", "", func(a *models.AdminTfpolicyVersions_attributes) { a.SetBeta(nil) }},
		"deprecated, no reason":   {"tool-1", "", func(a *models.AdminTfpolicyVersions_attributes) { a.SetDeprecated(ptr(true)) }},
		"incomplete build":        {"tool-1", "", func(a *models.AdminTfpolicyVersions_attributes) { a.GetArchs()[0].SetSha(nil) }},
		"no builds":               {"tool-1", "", func(a *models.AdminTfpolicyVersions_attributes) { a.SetArchs(nil) }},
		"duplicate builds":        {"tool-1", "", func(a *models.AdminTfpolicyVersions_attributes) { a.SetArchs(append(a.GetArchs(), a.GetArchs()[0])) }},
		"invalid sha in response": {"tool-1", "", func(a *models.AdminTfpolicyVersions_attributes) { a.GetArchs()[0].SetSha(ptr("bad")) }},
	}
	for name, tc := range invalid {
		t.Run("rejects "+name, func(t *testing.T) {
			_, err := tfpolicyVersionModel(ctx, testTFPolicyVersionResponse(tc.id, tc.mutate), tc.expect)
			assert.Error(t, err)
		})
	}

	t.Run("rejects missing data", func(t *testing.T) {
		_, err := tfpolicyVersionModel(ctx, models.NewAdminTfpolicyVersionsEnvelope(), "")
		assert.Error(t, err)
	})
}

func testTFPolicyVersionPlan(t *testing.T, archs ...string) modelAdminTFPolicyVersion {
	t.Helper()
	builds := make([]ToolArchitecture, 0, len(archs))
	for _, arch := range archs {
		builds = append(builds, testToolArchitecture("linux", arch))
	}
	set, diags := toolArchitecturesToSet(context.Background(), builds)
	require.False(t, diags.HasError(), "%v", diags)
	return modelAdminTFPolicyVersion{
		Version:          types.StringValue("1.0.0"),
		Official:         types.BoolValue(false),
		Enabled:          types.BoolValue(true),
		Beta:             types.BoolValue(false),
		Deprecated:       types.BoolValue(false),
		DeprecatedReason: types.StringNull(),
		Archs:            set,
	}
}

func testTFPolicyVersionResponse(id string, mutate func(*models.AdminTfpolicyVersions_attributes)) models.AdminTfpolicyVersionsEnvelopeable {
	attributes := models.NewAdminTfpolicyVersions_attributes()
	attributes.SetVersion(ptr("1.0.0"))
	attributes.SetOfficial(ptr(false))
	attributes.SetEnabled(ptr(true))
	attributes.SetBeta(ptr(false))
	attributes.SetDeprecated(ptr(false))
	var builds []models.AdminTfpolicyVersionArchitectureable
	for _, arch := range []models.AdminTfpolicyVersionArchitecture_arch{
		models.AMD64_ADMINTFPOLICYVERSIONARCHITECTURE_ARCH, models.ARM64_ADMINTFPOLICYVERSIONARCHITECTURE_ARCH,
	} {
		build := models.NewAdminTfpolicyVersionArchitecture()
		build.SetOs(ptr(models.LINUX_ADMINTFPOLICYVERSIONARCHITECTURE_OS))
		build.SetArch(ptr(arch))
		build.SetUrl(ptr(testTFPolicyPluginURL(arch.String())))
		build.SetSha(ptr(strings.Repeat("a", 64)))
		builds = append(builds, build)
	}
	attributes.SetArchs(builds)
	if mutate != nil {
		mutate(attributes)
	}
	data := models.NewAdminTfpolicyVersions()
	data.SetId(ptr(id))
	data.SetTypeEscaped(ptr(models.TFPOLICYVERSIONS_ADMINTFPOLICYVERSIONS_TYPE))
	data.SetAttributes(attributes)
	envelope := models.NewAdminTfpolicyVersionsEnvelope()
	envelope.SetData(data)
	return envelope
}

// skipUnlessTFPolicyVersionsSupported skips unless running acceptance tests against
// Terraform Enterprise with the admin Terraform policy plugin versions API.
func skipUnlessTFPolicyVersionsSupported(t *testing.T) {
	t.Helper()
	skipIfCloud(t)
	skipIfUnitTest(t)

	pc, err := getProviderClientUsingEnv()
	if err != nil {
		t.Fatalf("failed to build client: %v", err)
	}
	client := ConfiguredClient{Client: pc.TfeClient, ClientV2: pc.TFEClientV2, Hostname: pc.Hostname}
	supported, err := client.MeetsMinRemoteTFEVersion(minTFEVersionTFPolicyVersions)
	if err != nil {
		t.Fatalf("failed to check Terraform Enterprise version: %v", err)
	}
	if !supported {
		t.Skipf("Terraform policy plugin versions require Terraform Enterprise %s or later, got %s",
			minTFEVersionTFPolicyVersions, client.RemoteTFEVersion())
	}
}

func testAccTFETFPolicyVersionConfig(version, extra string, builds ...testTFPolicyBuild) string {
	var archs strings.Builder
	for _, build := range builds {
		fmt.Fprintf(&archs, `
    {
      url  = %q
      sha  = %q
      os   = "linux"
      arch = %q
    },`, build.downloadURL(), build.sha, build.arch)
	}
	return fmt.Sprintf(`
resource "tfe_tfpolicy_version" "foobar" {
  version = %q
%s
  archs = [%s
  ]
}`, version, extra, archs.String())
}

func testAccCheckTFETFPolicyVersionBuild(build testTFPolicyBuild) resource.TestCheckFunc {
	return resource.TestCheckTypeSetElemNestedAttrs(testTFPolicyVersionResourceName, "archs.*", map[string]string{
		"url":  build.downloadURL(),
		"sha":  build.sha,
		"os":   "linux",
		"arch": build.arch,
	})
}

// testAccCheckTFETFPolicyVersionAPI reads the version from the API and checks it matches.
func testAccCheckTFETFPolicyVersionAPI(name, version string, deprecated bool, archs ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}
		result, err := testAccConfiguredClient.ClientV2.API.Admin().TfpolicyVersions().ById(rs.Primary.ID).Get(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to read Terraform policy plugin version %s: %s", rs.Primary.ID, apiErrorDetail(err))
		}
		if result.GetData() == nil || result.GetData().GetAttributes() == nil {
			return fmt.Errorf("Terraform policy plugin version %s response is missing attributes", rs.Primary.ID)
		}
		a := result.GetData().GetAttributes()
		if got := valueOrZero(a.GetVersion()); got != version {
			return fmt.Errorf("expected version %q, got %q", version, got)
		}
		if got := valueOrZero(a.GetDeprecated()); got != deprecated {
			return fmt.Errorf("expected deprecated %t, got %t", deprecated, got)
		}
		got := make([]string, 0, len(a.GetArchs()))
		for _, build := range a.GetArchs() {
			got = append(got, build.GetArch().String())
		}
		slices.Sort(got)
		want := slices.Sorted(slices.Values(archs))
		if !slices.Equal(got, want) {
			return fmt.Errorf("expected archs %v, got %v", want, got)
		}
		return nil
	}
}

// testAccReplaceTFPolicyVersionArchs replaces a version's builds through the API to simulate drift.
func testAccReplaceTFPolicyVersionArchs(t *testing.T, id string, builds ...testTFPolicyBuild) {
	t.Helper()
	archs := make([]models.AdminTfpolicyVersionArchitectureable, 0, len(builds))
	for _, b := range builds {
		cpu := parseEnumPtr[models.AdminTfpolicyVersionArchitecture_arch](models.ParseAdminTfpolicyVersionArchitecture_arch, b.arch)
		if cpu == nil {
			t.Fatalf("unsupported test architecture %q", b.arch)
		}
		build := models.NewAdminTfpolicyVersionArchitecture()
		build.SetOs(ptr(models.LINUX_ADMINTFPOLICYVERSIONARCHITECTURE_OS))
		build.SetArch(cpu)
		build.SetUrl(ptr(b.downloadURL()))
		build.SetSha(ptr(b.sha))
		archs = append(archs, build)
	}
	attributes := models.NewAdminTfpolicyVersions_attributes()
	attributes.SetArchs(archs)
	data := models.NewAdminTfpolicyVersions()
	data.SetTypeEscaped(ptr(models.TFPOLICYVERSIONS_ADMINTFPOLICYVERSIONS_TYPE))
	data.SetAttributes(attributes)
	body := models.NewAdminTfpolicyVersionsEnvelope()
	body.SetData(data)
	if _, err := testAccConfiguredClient.ClientV2.API.Admin().TfpolicyVersions().ById(id).Patch(ctx, body, nil); err != nil {
		t.Fatalf("failed to replace builds for Terraform policy plugin version %s: %s", id, apiErrorDetail(err))
	}
}

func testAccCheckTFETFPolicyVersionDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_tfpolicy_version" {
			continue
		}
		_, err := testAccConfiguredClient.ClientV2.API.Admin().TfpolicyVersions().ById(rs.Primary.ID).Get(ctx, nil)
		if err == nil {
			return fmt.Errorf("Terraform policy plugin version %s still exists", rs.Primary.ID)
		}
		if !errors.Is(err, tfev2.ErrNotFound) {
			return fmt.Errorf("failed to read Terraform policy plugin version %s: %w", rs.Primary.ID, err)
		}
	}
	return nil
}

// genSafeRandomTFPolicyVersion returns a random version that will not collide with a real release.
func genSafeRandomTFPolicyVersion() string {
	return fmt.Sprintf("0.0.%d", rand.New(rand.NewSource(time.Now().UnixNano())).Int())
}
