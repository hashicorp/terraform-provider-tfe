// Copyright IBM Corp. 2024, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// exampleExceptions is the parsed representation of examples/error_exceptions.json.
//
// To exempt example live tests at the directory level, add the resource directory
// name (for example, "tfe_agent_pool") to examples_as_ci_exempt_directories in
// examples/error_exceptions.json. Use that only when every example in the
// directory is intentionally skipped in CI.
//
// For narrower cases that cut across otherwise testable directories, prefer
// targeted content-based checks in skipReason() instead of adding more directory
// exemptions.
type exampleExceptions struct {
	ExamplesAsCIExemptDirectories []string `json:"examples_as_ci_exempt_directories"`
}

// loadExampleExceptions reads and parses examples/error_exceptions.json.
// The path is relative to the package root (two levels up from internal/provider).
func loadExampleExceptions(t *testing.T) exampleExceptions {
	t.Helper()
	data, err := os.ReadFile("../../examples/error_exceptions.json")
	if err != nil {
		t.Fatalf("failed to read error_exceptions.json: %s", err)
	}
	var exc exampleExceptions
	if err := json.Unmarshal(data, &exc); err != nil {
		t.Fatalf("failed to parse error_exceptions.json: %s", err)
	}
	return exc
}

// isDirExempt reports whether the given example file's resource directory is
// in the examples_as_ci_exempt_directories list.
func isDirExempt(path string, exc exampleExceptions) bool {
	dir := filepath.Base(filepath.Dir(path))
	for _, d := range exc.ExamplesAsCIExemptDirectories {
		if dir == d {
			return true
		}
	}
	return false
}

// Compiled regexes used by skipReason.
var (
	// backendBlockRe matches a terraform{} block at the start of a line — these
	// are backend/cloud configurations that cannot appear in test configs.
	backendBlockRe = regexp.MustCompile(`(?m)^\s*terraform\s*\{`)

	// memberUserIDRe matches references to tfe_organization_membership.*.user_id
	// or data.tfe_organization_membership.*.user_id. These are always empty until
	// an invitation is manually accepted, so any config containing them will fail.
	memberUserIDRe = regexp.MustCompile(`tfe_organization_membership\.\w+\.user_id`)

	// externalServiceURLRe detects the placeholder URL used in run-task examples.
	// It is replaced by RUN_TASKS_URL at substitution time; if it survives to this
	// check that means RUN_TASKS_URL was not set.
	externalServiceURLRe = regexp.MustCompile(`https://external\.service\.com`)

	// adminProviderRe matches an explicit provider "tfe" block in the config.
	// Admin examples declare a second tfe provider alias with an admin token;
	// the test harness only sets up one provider, so these cannot run.
	adminProviderRe = regexp.MustCompile(`(?m)^\s*provider\s+"tfe"\s*\{`)

	// samlSettingsRe detects configs that create tfe_saml_settings, which
	// requires a live SAML IdP and enterprise-only admin access.
	samlSettingsRe = regexp.MustCompile(`resource\s+"tfe_saml_settings"`)

	// teamMemberUsernameRe detects hardcoded usernames in tfe_team_member /
	// tfe_team_members that reference pre-existing users not present in CI.
	teamMemberUsernameRe = regexp.MustCompile(`(?m)^\s*usernames?\s*=`)
)

// skipReason returns a human-readable reason and true if the config should be
// skipped, or ("", false) if it should be run. Detection is purely content-based
// — no static allowlist is consulted.
func skipReason(cfg string) (string, bool) {
	// Empty config after all substitutions (e.g. tfe_organization/resource.tf
	// which becomes empty after the shared fixture strips its only block).
	if strings.TrimSpace(cfg) == "" {
		return "config is empty after substitutions", true
	}
	// Explicit provider "tfe" block: admin examples declare a second tfe provider
	// alias; the test harness configures only one provider instance.
	if adminProviderRe.MatchString(cfg) {
		return "declares explicit provider block (admin token alias not available in test harness)", true
	}
	// VCS / OAuth dependency: requires a live GitHub/GitLab token not available in CI.
	if strings.Contains(cfg, `"tfe_oauth_client"`) ||
		strings.Contains(cfg, "oauth_token_id") {
		return "requires tfe_oauth_client (VCS token unavailable in CI)", true
	}
	// GitHub App installation lookup requires a configured GitHub App token not available in CI.
	if strings.Contains(cfg, `data "tfe_github_app_installation"`) {
		return "requires tfe_github_app_installation (GitHub App installation unavailable in CI)", true
	}
	// file() call: references a local file path that does not exist in CI.
	if strings.Contains(cfg, "file(") {
		return "uses file() with a path not present in CI", true
	}
	// terraform{} backend/cloud block: structurally invalid inside a test config.
	if backendBlockRe.MatchString(cfg) {
		return "contains terraform{} backend block, invalid in test context", true
	}
	// data "tfe_workspace" with no backing workspace resource in the same config.
	if strings.Contains(cfg, `data "tfe_workspace"`) &&
		!strings.Contains(cfg, `resource "tfe_workspace"`) {
		return "uses data.tfe_workspace with no backing workspace resource in config", true
	}
	// tfe_stack reference: requires ENABLE_BETA + special TFE build.
	if strings.Contains(cfg, `"tfe_stack"`) {
		return "references tfe_stack (requires ENABLE_BETA)", true
	}
	// data "tfe_slug": requires a local policy directory that does not exist in CI.
	if strings.Contains(cfg, `data "tfe_slug"`) {
		return "uses data.tfe_slug with a local path not present in CI", true
	}
	// tfe_saml_settings: requires a live SAML IdP and enterprise admin access.
	if samlSettingsRe.MatchString(cfg) {
		return "creates tfe_saml_settings (requires enterprise SAML IdP)", true
	}
	// tfe_organization_membership.*.user_id or data.tfe_organization_membership.*.user_id:
	// always empty until the invitation is manually accepted.
	if memberUserIDRe.MatchString(cfg) {
		return "references tfe_organization_membership.*.user_id (empty until invite accepted)", true
	}
	// tfe_team_member / tfe_team_members with hardcoded usernames: those users do
	// not exist in the CI test org.
	if (strings.Contains(cfg, `"tfe_team_member"`) || strings.Contains(cfg, `"tfe_team_members"`)) &&
		teamMemberUsernameRe.MatchString(cfg) {
		return "references hardcoded usernames not present in CI test org", true
	}
	// time_rotating: requires the external HashiCorp time provider.
	if strings.Contains(cfg, "time_rotating") {
		return "requires external time provider", true
	}
	// write-only HMAC key examples require a realistic key value; synthetic injected values are rejected.
	if strings.Contains(cfg, "hmac_key_wo") {
		return "requires a valid HMAC key value not available in CI", true
	}
	// External service URL still present: RUN_TASKS_URL was not set in this environment.
	if externalServiceURLRe.MatchString(cfg) {
		return "uses https://external.service.com and RUN_TASKS_URL is not set", true
	}
	// no_code_module with version_pin: requires a published module version.
	if strings.Contains(cfg, "version_pin") {
		return "uses version_pin which requires a published module version", true
	}
	// Notification configuration examples are not currently accepted by the CI environment/API.
	if strings.Contains(cfg, `"tfe_notification_configuration"`) ||
		strings.Contains(cfg, `"tfe_project_notification_configuration"`) ||
		strings.Contains(cfg, `"tfe_team_notification_configuration"`) {
		return "uses notification configuration resources not currently supported in CI", true
	}
	// Agent pool examples outside exempt directories still require agent pool capability unavailable in CI.
	if strings.Contains(cfg, `"tfe_agent_pool_allowed_projects"`) ||
		strings.Contains(cfg, `"tfe_agent_pool_allowed_workspaces"`) ||
		strings.Contains(cfg, `"tfe_agent_pool_excluded_workspaces"`) {
		return "uses agent pool resources unavailable in CI", true
	}
	// data "tfe_agent_pool" resolves during the pre-apply plan (its arguments are
	// literal and it has no dependency on the harness's injected agent-pool
	// fixture), so the lookup runs before that pool is created and returns 404.
	if strings.Contains(cfg, `data "tfe_agent_pool"`) {
		return "uses data.tfe_agent_pool, which is read at plan time before the injected agent-pool fixture exists (404)", true
	}
	// Examples using for_each with indexed resource references are not supported by the test harness.
	if strings.Contains(cfg, "for_each") &&
		(strings.Contains(cfg, `tfe_workspace.test[`) || strings.Contains(cfg, `tfe_organization_membership.`)) {
		return "uses for_each with indexed resource references unsupported by the test harness", true
	}
	// Standalone data-source workspace settings example relies on pre-existing workspaces not guaranteed in CI.
	if strings.Contains(cfg, `data "tfe_workspace"`) && strings.Contains(cfg, `"tfe_workspace_settings"`) {
		return "uses data.tfe_workspace lookup without guaranteed backing workspace in CI", true
	}
	// Tag policy set examples require tag lookup behavior that is not reliable in CI.
	if strings.Contains(cfg, `"tfe_tag_policy_set"`) || strings.Contains(cfg, `"tfe_tag_policy_set_exclusion"`) {
		return "uses tag policy set resources that require tag lookup behavior unavailable in CI", true
	}
	return "", false
}

// varDefaultRe matches a variable block that has no default value.
// Captures the variable name.
var varDefaultRe = regexp.MustCompile(`(?s)variable\s+"(\w+)"\s*\{[^}]*\}`)
var varHasDefault = regexp.MustCompile(`\bdefault\b`)

// extractVarsWithoutDefault returns variable names declared in the config
// that have no default value — these must be injected via ConfigVariables.
func extractVarsWithoutDefault(cfg string) []string {
	var names []string
	for _, match := range varDefaultRe.FindAllStringSubmatch(cfg, -1) {
		block := match[0]
		if !varHasDefault.MatchString(block) {
			names = append(names, match[1])
		}
	}
	return names
}

// reOrgResourceBlock matches a complete `resource "tfe_organization" "<label>" { ... }`
// block including its closing brace (simple single-level heuristic).
var reOrgResourceBlock = regexp.MustCompile(`(?s)resource "tfe_organization" "[^"]*" \{[^}]*\}`)

// reOrgRef matches tfe_organization.<label>.name or .id
var reOrgRef = regexp.MustCompile(`tfe_organization\.[a-zA-Z0-9_-]+\.(name|id)`)

// stripOrgBlocks removes any inline tfe_organization resource blocks from a
// config and replaces all .name/.id references with the literal org name.
// After this the shared fixture's tfe_organization.example is the only org
// block, pointing at the pre-created test org.
func stripOrgBlocks(cfg, orgName string) string {
	cfg = reOrgResourceBlock.ReplaceAllString(cfg, "")
	cfg = reOrgRef.ReplaceAllString(cfg, fmt.Sprintf("%q", orgName))
	return cfg
}

// TestAccTFEExamples applies every non-exempt example file as a real Terraform
// configuration against a live TFE instance, verifying that documented
// examples can be applied and destroyed without error.
//
// Guards on TF_ACC so it runs automatically in the existing acceptance test CI
// job alongside all other TestAcc* tests.
func TestAccTFEExamples(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("set TF_ACC=1 to run live example tests")
	}

	exc := loadExampleExceptions(t)

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatalf("failed to build TFE client: %s", err)
	}

	suiteRand := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Load the shared fixture once; substitutions are applied to it just like
	// any example file so that resource names resolve to the test org.
	sharedRaw, err := os.ReadFile("../../examples/shared/example.tf")
	if err != nil {
		t.Fatalf("failed to read examples/shared/example.tf: %s", err)
	}

	paths, err := filepath.Glob("../../examples/resources/*/resource*.tf")
	if err != nil {
		t.Fatalf("failed to glob example files: %s", err)
	}
	if len(paths) == 0 {
		t.Fatal("no example files found under examples/resources/")
	}

	for _, path := range paths {
		// Derive stable sub-test name: "<dir>/<filename>"
		name := filepath.Join(filepath.Base(filepath.Dir(path)), filepath.Base(path))

		t.Run(name, func(t *testing.T) {
			if isDirExempt(path, exc) {
				t.Skip("directory exempted in error_exceptions.json under examples_as_ci_exempt_directories")
				return
			}

			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("failed to read %s: %s", path, err)
			}

			exampleRand := suiteRand.Int()
			orgName := fmt.Sprintf("tst-ex-%d", exampleRand)

			_, orgCleanup := createOrganization(t, tfeClient,
				tfe.OrganizationCreateOptions{
					Name:  tfe.String(orgName),
					Email: tfe.String("admin@example.com"),
				})
			t.Cleanup(orgCleanup)

			providers := muxedProvidersWithDefaultOrganization(orgName)

			// Apply substitutions to both the shared fixture and the example.
			shared := applyExampleSubstitutions(string(sharedRaw), orgName, exampleRand)
			example := applyExampleSubstitutions(string(raw), orgName, exampleRand)

			// Combine: shared fixtures first, then the example.
			combined := shared + "\n" + example

			// Strip any inline tfe_organization blocks the example still contains
			// (pre-migration files, or examples where the org is the subject).
			combined = stripOrgBlocks(combined, orgName)

			if reason, skip := skipReason(combined); skip {
				t.Skipf("auto-skipped: %s", reason)
				return
			}

			// Inject values for any variable blocks that have no default
			// (pattern used by write-only field examples).
			vars := config.Variables{}
			for _, v := range extractVarsWithoutDefault(combined) {
				vars[v] = config.StringVariable("synthetic-test-value")
			}

			step := resource.TestStep{Config: combined}
			if len(vars) > 0 {
				step.ConfigVariables = vars
			}

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: providers,
				Steps:                    []resource.TestStep{step},
			})
		})
	}
}

// applyExampleSubstitutions replaces well-known placeholder strings in an
// example config with values valid for the live test environment.
func applyExampleSubstitutions(cfg, orgName string, rInt int) string {
	runTasksURL := os.Getenv("RUN_TASKS_URL")
	if runTasksURL == "" {
		// Leave the literal in place; skipReason will detect it and skip the test.
		runTasksURL = "https://external.service.com"
	}

	r := strings.NewReplacer(
		`"my-org-name"`, fmt.Sprintf("%q", orgName),
		`"my-example-org"`, fmt.Sprintf("%q", orgName),
		`"my-organization"`, fmt.Sprintf("%q", orgName),
		`"org-name"`, fmt.Sprintf("%q", orgName),
		`"organization name"`, fmt.Sprintf("%q", orgName),
		`"my-workspace-name"`, fmt.Sprintf(`"tst-ex-ws-%d"`, rInt),
		`"my-sourceable-workspace-name"`, fmt.Sprintf(`"tst-ex-ws2-%d"`, rInt),
		`"my-project-name"`, fmt.Sprintf(`"tst-ex-proj-%d"`, rInt),
		`"my-team-name"`, fmt.Sprintf(`"tst-ex-team-%d"`, rInt),
		`"my-admin-team"`, fmt.Sprintf(`"tst-ex-admin-team-%d"`, rInt),
		`"my-agent-pool-name"`, fmt.Sprintf(`"tst-ex-pool-%d"`, rInt),
		`"https://external.service.com"`, fmt.Sprintf("%q", runTasksURL),
	)
	return r.Replace(cfg)
}
