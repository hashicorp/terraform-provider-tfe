// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"math/rand"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// Known Tool Versions that will exist in the TFE/TFC instance that's
// being used for acceptance testing. Once the /api/v2/tool-versions endpoint
// is available in go-tfe, we can retrieve this dynamically.
const KnownOPAToolVersion = "0.44.0"
const KnownSentinelToolVersion = "0.22.1"

func TestAccTFEPolicySetDataSource_basic(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySetDataSourceConfig_basic(org.Name, rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tfe_policy_set.bar", "id"),
					resource.TestCheckResourceAttr(
						"data.tfe_policy_set.bar", "name", fmt.Sprintf("tst-policy-set-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.tfe_policy_set.bar", "description", "Policy Set"),
					resource.TestCheckResourceAttr(
						"data.tfe_policy_set.bar", "global", "false"),
					resource.TestCheckResourceAttr(
						"data.tfe_policy_set.bar", "organization", org.Name),
					resource.TestCheckResourceAttr(
						"data.tfe_policy_set.bar", "policy_ids.#", "1"),
					resource.TestCheckResourceAttr(
						"data.tfe_policy_set.bar", "agent_enabled", "false"),
					resource.TestCheckResourceAttr(
						"data.tfe_policy_set.bar", "workspace_ids.#", "1"),
					resource.TestCheckResourceAttr(
						"data.tfe_policy_set.bar", "excluded_workspace_ids.#", "1"),
					resource.TestCheckResourceAttr(
						"data.tfe_policy_set.bar", "project_ids.#", "1"),
					resource.TestCheckResourceAttr(
						"data.tfe_policy_set.bar", "vcs_repo.#", "0"),
				),
			},
		},
	},
	)
}

func TestAccTFEPolicySetDataSource_pinnedPolicyRuntimeVersion(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckTFESentinelVersionDestroy,
		ProtoV5ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySetDataSourceConfig_pinnedPolicyRuntimeVersion(org.Name, rInt, KnownSentinelToolVersion),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tfe_policy_set.bar", "id"),
					resource.TestCheckResourceAttr(
						"data.tfe_policy_set.bar", "name", fmt.Sprintf("tst-policy-set-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.tfe_policy_set.bar", "description", "Policy Set"),
					resource.TestCheckResourceAttr(
						"data.tfe_policy_set.bar", "global", "false"),
					resource.TestCheckResourceAttr(
						"data.tfe_policy_set.bar", "organization", org.Name),
					resource.TestCheckResourceAttr(
						"data.tfe_policy_set.bar", "policy_ids.#", "1"),
					resource.TestCheckResourceAttr(
						"data.tfe_policy_set.bar", "agent_enabled", "true"),
					resource.TestCheckResourceAttr(
						"data.tfe_policy_set.bar", "policy_tool_version", KnownSentinelToolVersion),
					resource.TestCheckResourceAttr(
						"data.tfe_policy_set.bar", "workspace_ids.#", "1"),
					resource.TestCheckResourceAttr(
						"data.tfe_policy_set.bar", "excluded_workspace_ids.#", "1"),
					resource.TestCheckResourceAttr(
						"data.tfe_policy_set.bar", "project_ids.#", "1"),
					resource.TestCheckResourceAttr(
						"data.tfe_policy_set.bar", "vcs_repo.#", "0"),
				),
			},
		},
	},
	)
}

func TestAccTFEPolicySetDataSourceOPA_basic(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckTFEOPAVersionDestroy,
		ProtoV5ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySetDataSourceConfigOPA_basic(org.Name, rInt, KnownOPAToolVersion),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tfe_policy_set.bar", "id"),
					resource.TestCheckResourceAttr(
						"data.tfe_policy_set.bar", "name", fmt.Sprintf("tst-policy-set-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.tfe_policy_set.bar", "description", "Policy Set"),
					resource.TestCheckResourceAttr(
						"data.tfe_policy_set.bar", "global", "false"),
					resource.TestCheckResourceAttr(
						"data.tfe_policy_set.bar", "organization", org.Name),
					resource.TestCheckResourceAttr(
						"data.tfe_policy_set.bar", "kind", "opa"),
					resource.TestCheckResourceAttr(
						"data.tfe_policy_set.bar", "agent_enabled", "true"),
					resource.TestCheckResourceAttr(
						"data.tfe_policy_set.bar", "overridable", "true"),
					resource.TestCheckResourceAttr(
						"data.tfe_policy_set.bar", "workspace_ids.#", "1"),
					resource.TestCheckResourceAttr(
						"data.tfe_policy_set.bar", "excluded_workspace_ids.#", "1"),
					resource.TestCheckResourceAttr(
						"data.tfe_policy_set.bar", "project_ids.#", "1"),
					resource.TestCheckResourceAttr(
						"data.tfe_policy_set.bar", "vcs_repo.#", "0"),
				),
			},
		},
	},
	)
}

func TestAccTFEPolicySetDataSource_vcs(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)

			if envGithubToken == "" {
				t.Skip("Please set GITHUB_TOKEN to run this test")
			}
			if envGithubPolicySetIdentifier == "" {
				t.Skip("Please set GITHUB_POLICY_SET_IDENTIFIER to run this test")
			}
			if envGithubPolicySetBranch == "" {
				t.Skip("Please set GITHUB_POLICY_SET_BRANCH to run this test")
			}
			if envGithubPolicySetPath == "" {
				t.Skip("Please set GITHUB_POLIY_SET_PATH to run this test")
			}
		},
		ProtoV5ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySetDataSourceConfig_vcs(org.Name, rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tfe_policy_set.bar", "id"),
					resource.TestCheckResourceAttr(
						"data.tfe_policy_set.bar", "name", fmt.Sprintf("tst-policy-set-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.tfe_policy_set.bar", "description", "Policy Set"),
					resource.TestCheckResourceAttr(
						"data.tfe_policy_set.bar", "global", "false"),
					resource.TestCheckResourceAttr(
						"data.tfe_policy_set.bar", "kind", "sentinel"),
					resource.TestCheckResourceAttr(
						"data.tfe_policy_set.bar", "agent_enabled", "false"),
					resource.TestCheckResourceAttr(
						"data.tfe_policy_set.bar", "organization", org.Name),
					resource.TestCheckResourceAttr(
						"data.tfe_policy_set.bar", "policy_ids.#", "0"),
					resource.TestCheckResourceAttr(
						"data.tfe_policy_set.bar", "workspace_ids.#", "0"),
					resource.TestCheckResourceAttr(
						"data.tfe_policy_set.bar", "excluded_workspace_ids.#", "0"),
					resource.TestCheckResourceAttr(
						"data.tfe_policy_set.bar", "project_ids.#", "0"),
					resource.TestCheckResourceAttr(
						"data.tfe_policy_set.bar", "vcs_repo.#", "1"),
				),
			},
		},
	},
	)
}

func TestAccTFEPolicySetDataSource_notFound(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFEPolicySetDataSourceConfig_notFound(rInt),
				ExpectError: regexp.MustCompile(`Error: could not find policy set`),
			},
		},
	},
	)
}

func testAccTFEPolicySetDataSourceConfig_basic(organization string, rInt int) string {
	return fmt.Sprintf(`
locals {
  organization_name = "%s"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-foo-%d"
  organization = local.organization_name
}

resource "tfe_project" "foobar" {
  name         = "project-foo-%d"
  organization = local.organization_name
}

resource "tfe_sentinel_policy" "foo" {
  name         = "policy-foo"
  policy       = "main = rule { true }"
  organization = local.organization_name
}

resource "tfe_policy_set" "foobar" {
  name         = "tst-policy-set-%d"
  description  = "Policy Set"
  organization = local.organization_name
  policy_ids   = [tfe_sentinel_policy.foo.id]
  workspace_ids = [tfe_workspace.foobar.id]
}

resource "tfe_project_policy_set" "foobar" {
	policy_set_id = tfe_policy_set.foobar.id
	project_id = tfe_project.foobar.id
}

resource "tfe_workspace_policy_set_exclusion" "foobar" {
	policy_set_id = tfe_policy_set.foobar.id
	workspace_id = tfe_workspace.foobar.id
}

data "tfe_policy_set" "bar" {
  name = tfe_policy_set.foobar.name
  organization = local.organization_name
  depends_on=[tfe_policy_set.foobar, tfe_project_policy_set.foobar, tfe_workspace_policy_set_exclusion.foobar]
}`, organization, rInt, rInt, rInt)
}

func testAccTFEPolicySetDataSourceConfig_pinnedPolicyRuntimeVersion(organization string, rInt int, version string) string {
	return fmt.Sprintf(`
locals {
  organization_name = "%s"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-foo-%d"
  organization = local.organization_name
}

resource "tfe_project" "foobar" {
  name         = "project-foo-%d"
  organization = local.organization_name
}

resource "tfe_sentinel_policy" "foo" {
  name         = "policy-foo"
  policy       = "main = rule { true }"
  organization = local.organization_name
}

resource "tfe_policy_set" "foobar" {
  name         = "tst-policy-set-%d"
  description  = "Policy Set"
  organization = local.organization_name
  agent_enabled = true
  policy_tool_version = "%s"
  policy_ids   = [tfe_sentinel_policy.foo.id]
  workspace_ids = [tfe_workspace.foobar.id]
}

resource "tfe_project_policy_set" "foobar" {
	policy_set_id = tfe_policy_set.foobar.id
	project_id = tfe_project.foobar.id
}

resource "tfe_workspace_policy_set_exclusion" "foobar" {
	policy_set_id = tfe_policy_set.foobar.id
	workspace_id = tfe_workspace.foobar.id
}

data "tfe_policy_set" "bar" {
  name = tfe_policy_set.foobar.name
  organization = local.organization_name
  depends_on=[tfe_policy_set.foobar, tfe_project_policy_set.foobar, tfe_workspace_policy_set_exclusion.foobar]
}`, organization, rInt, rInt, rInt, version)
}

func testAccTFEPolicySetDataSourceConfigOPA_basic(organization string, rInt int, version string) string {
	return fmt.Sprintf(`
locals {
  organization_name = "%s"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-foo-%d"
  organization = local.organization_name
}

resource "tfe_project" "foobar" {
  name         = "project-foo-%d"
  organization = local.organization_name
}

resource "tfe_policy_set" "foobar" {
  name         = "tst-policy-set-%d"
  description  = "Policy Set"
  organization = local.organization_name
  kind         = "opa"
  agent_enabled = true
  policy_tool_version = "%s"
  overridable  = true
  workspace_ids = [tfe_workspace.foobar.id]
}

resource "tfe_project_policy_set" "foobar" {
	policy_set_id = tfe_policy_set.foobar.id
	project_id = tfe_project.foobar.id
}

resource "tfe_workspace_policy_set_exclusion" "foobar" {
	policy_set_id = tfe_policy_set.foobar.id
	workspace_id = tfe_workspace.foobar.id
}

data "tfe_policy_set" "bar" {
  name = tfe_policy_set.foobar.name
  organization = local.organization_name
  kind = "opa"
  depends_on=[tfe_policy_set.foobar, tfe_project_policy_set.foobar, tfe_workspace_policy_set_exclusion.foobar]
}`, organization, rInt, rInt, rInt, version)
}

func testAccTFEPolicySetDataSourceConfig_vcs(organization string, rInt int) string {
	return fmt.Sprintf(`
locals {
    organization_name = "%s"
}

resource "tfe_oauth_client" "test" {
  organization     = local.organization_name
  api_url          = "https://api.github.com"
  http_url         = "https://github.com"
  oauth_token      = "%s"
  service_provider = "github"
}

resource "tfe_policy_set" "foobar" {
  name         = "tst-policy-set-%d"
  description  = "Policy Set"
  organization = local.organization_name
  vcs_repo {
	identifier         = "%s"
	branch             = "main"
	ingress_submodules = true
	oauth_token_id     = tfe_oauth_client.test.oauth_token_id
  }

  policies_path = "%s"
}

data "tfe_policy_set" "bar" {
  name         = tfe_policy_set.foobar.name
  organization = local.organization_name
}
`, organization,
		envGithubToken,
		rInt,
		envGithubPolicySetIdentifier,
		envGithubPolicySetPath,
	)
}

func testAccTFEPolicySetDataSourceConfig_notFound(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

data "tfe_policy_set" "not-found" {
  name = "does-not-exist"
  organization = tfe_organization.foobar.id
}`, rInt)
}
