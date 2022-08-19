package tfe

import (
	"fmt"
	"math/rand"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccTFEPolicySetDataSource_basic(t *testing.T) {
	skipIfFreeOnly(t)

	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySetDataSourceConfig_basic(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tfe_policy_set.bar", "id"),
					resource.TestCheckResourceAttr(
						"data.tfe_policy_set.bar", "name", fmt.Sprintf("tst-policy-set-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.tfe_policy_set.bar", "description", "Policy Set"),
					resource.TestCheckResourceAttr(
						"data.tfe_policy_set.bar", "global", "false"),
					resource.TestCheckResourceAttr(
						"data.tfe_policy_set.bar", "organization", fmt.Sprintf("tst-terraform-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.tfe_policy_set.bar", "policy_ids.#", "1"),
					resource.TestCheckResourceAttr(
						"data.tfe_policy_set.bar", "workspace_ids.#", "1"),
					resource.TestCheckResourceAttr(
						"data.tfe_policy_set.bar", "vcs_repo.#", "0"),
				),
			},
		},
	},
	)
}

func TestAccTFEPolicySetDataSource_vcs(t *testing.T) {
	skipIfFreeOnly(t)

	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)

			if GITHUB_TOKEN == "" {
				t.Skip("Please set GITHUB_TOKEN to run this test")
			}
			if GITHUB_POLICY_SET_IDENTIFIER == "" {
				t.Skip("Please set GITHUB_POLICY_SET_IDENTIFIER to run this test")
			}
			if GITHUB_POLICY_SET_BRANCH == "" {
				t.Skip("Please set GITHUB_POLICY_SET_BRANCH to run this test")
			}
			if GITHUB_POLICY_SET_PATH == "" {
				t.Skip("Please set GITHUB_POLICY_SET_PATH to run this test")
			}
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySetDataSourceConfig_vcs(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tfe_policy_set.bar", "id"),
					resource.TestCheckResourceAttr(
						"data.tfe_policy_set.bar", "name", fmt.Sprintf("tst-policy-set-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.tfe_policy_set.bar", "description", "Policy Set"),
					resource.TestCheckResourceAttr(
						"data.tfe_policy_set.bar", "global", "false"),
					resource.TestCheckResourceAttr(
						"data.tfe_policy_set.bar", "organization", fmt.Sprintf("tst-terraform-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.tfe_policy_set.bar", "policy_ids.#", "0"),
					resource.TestCheckResourceAttr(
						"data.tfe_policy_set.bar", "workspace_ids.#", "0"),
					resource.TestCheckResourceAttr(
						"data.tfe_policy_set.bar", "vcs_repo.#", "1"),
				),
			},
		},
	},
	)
}

func TestAccTFEPolicySetDataSource_notFound(t *testing.T) {
	skipIfFreeOnly(t)

	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFEPolicySetDataSourceConfig_notFound(rInt),
				ExpectError: regexp.MustCompile(`Error: Could not find policy set`),
			},
		},
	},
	)
}

func testAccTFEPolicySetDataSourceConfig_basic(rInt int) string {
	return fmt.Sprintf(`
	resource "tfe_organization" "foobar" {
		name  = "tst-terraform-%d"
		email = "admin@company.com"
	}

	resource "tfe_workspace" "foobar" {
		name         = "workspace-foo-%d"
		organization = tfe_organization.foobar.id
	}

	resource "tfe_sentinel_policy" "foo" {
		name         = "policy-foo"
		policy       = "main = rule { true }"
		organization = tfe_organization.foobar.id
	}

	resource "tfe_policy_set" "foobar" {
		name         = "tst-policy-set-%d"
		description  = "Policy Set"
		organization = tfe_organization.foobar.id
		policy_ids   = [tfe_sentinel_policy.foo.id]
		workspace_ids = [tfe_workspace.foobar.id]
	}

  data "tfe_policy_set" "bar" {
		name = tfe_policy_set.foobar.name
		organization = tfe_organization.foobar.id
	}`, rInt, rInt, rInt)
}

func testAccTFEPolicySetDataSourceConfig_vcs(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_oauth_client" "test" {
  organization     = tfe_organization.foobar.id
  api_url          = "https://api.github.com"
  http_url         = "https://github.com"
  oauth_token      = "%s"
  service_provider = "github"
}

resource "tfe_policy_set" "foobar" {
  name         = "tst-policy-set-%d"
  description  = "Policy Set"
  organization = tfe_organization.foobar.id
	vcs_repo {
		identifier         = "%s"
		branch             = "main"
		ingress_submodules = true
		oauth_token_id     = tfe_oauth_client.test.oauth_token_id
	}

  policies_path = "%s"
}

data "tfe_policy_set" "bar" {
	name = tfe_policy_set.foobar.name
	organization = tfe_organization.foobar.id
}
`, rInt,
		GITHUB_TOKEN,
		rInt,
		GITHUB_POLICY_SET_IDENTIFIER,
		GITHUB_POLICY_SET_PATH,
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
