// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccTFEPolicySet_basic(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	policySet := &tfe.PolicySet{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySet_basic(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetAttributes(policySet),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "name", "tst-terraform"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "description", "Policy Set"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "global", "false"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "policy_ids.#", "1"),
				),
			},
		},
	})
}

func TestAccTFEPolicySetOPA_basic(t *testing.T) {
	skipUnlessBeta(t)
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	policySet := &tfe.PolicySet{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySetOPA_basic(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetAttributes(policySet),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "name", "tst-terraform"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "kind", "opa"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "overridable", "false"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "description", "Policy Set"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "global", "false"),
				),
			},
		},
	})
}

func TestAccTFEPolicySet_updateOverridable(t *testing.T) {
	skipUnlessBeta(t)
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	policySet := &tfe.PolicySet{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySetOPA_basic(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetAttributes(policySet),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "name", "tst-terraform"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "description", "Policy Set"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "kind", "opa"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "global", "false"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "overridable", "true"),
				),
			},

			{
				Config: testAccTFEPolicySetOPA_overridable(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.foobar", policySet),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "name", "tst-terraform"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "global", "false"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "kind", "opa"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "workspace_ids.#", "1"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "overridable", "false"),
				),
			},
		},
	})
}

func TestAccTFEPolicySet_update(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	policySet := &tfe.PolicySet{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySet_basic(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetAttributes(policySet),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "name", "tst-terraform"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "description", "Policy Set"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "global", "false"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "policy_ids.#", "1"),
				),
			},

			{
				Config: testAccTFEPolicySet_populated(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetPopulated(policySet, org.Name),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "name", "terraform-populated"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "global", "false"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "kind", "sentinel"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "policy_ids.#", "1"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "workspace_ids.#", "1"),
				),
			},
		},
	})
}

func TestAccTFEPolicySet_updateEmpty(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	policySet := &tfe.PolicySet{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySet_basic(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetAttributes(policySet),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "name", "tst-terraform"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "description", "Policy Set"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "global", "false"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "policy_ids.#", "1"),
				),
			},

			{
				Config: testAccTFEPolicySet_empty(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetAttributes(policySet),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "name", "tst-terraform"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "description", "Policy Set"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "global", "false"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "policy_ids.#", "0"),
				),
			},
		},
	})
}

func TestAccTFEPolicySet_updatePopulated(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	policySet := &tfe.PolicySet{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySet_populated(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetPopulated(policySet, org.Name),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "name", "terraform-populated"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "global", "false"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "policy_ids.#", "1"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "workspace_ids.#", "1"),
				),
			},

			{
				Config: testAccTFEPolicySet_updatePopulated(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetPopulatedUpdated(policySet, org.Name),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "name", "terraform-populated-updated"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "global", "false"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "kind", "sentinel"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "policy_ids.#", "1"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "workspace_ids.#", "1"),
				),
			},
		},
	})
}

func TestAccTFEPolicySet_updateToGlobal(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	policySet := &tfe.PolicySet{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySet_populated(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetPopulated(policySet, org.Name),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "name", "terraform-populated"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "global", "false"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "policy_ids.#", "1"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "workspace_ids.#", "1"),
				),
			},

			{
				Config: testAccTFEPolicySet_global(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetGlobal(policySet),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "name", "terraform-global"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "global", "true"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "kind", "sentinel"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "policy_ids.#", "1"),
				),
			},
		},
	})
}

func TestAccTFEPolicySet_updateToWorkspace(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	policySet := &tfe.PolicySet{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySet_global(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetGlobal(policySet),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "name", "terraform-global"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "global", "true"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "policy_ids.#", "1"),
				),
			},

			{
				Config: testAccTFEPolicySet_populated(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetPopulated(policySet, org.Name),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "name", "terraform-populated"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "global", "false"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "policy_ids.#", "1"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "workspace_ids.#", "1"),
				),
			},
		},
	})
}

func TestAccTFEPolicySet_vcs(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	policySet := &tfe.PolicySet{}

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
				t.Skip("Please set GITHUB_POLICY_SET_PATH to run this test")
			}
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySet_vcs(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetAttributes(policySet),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "name", "tst-terraform"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "description", "Policy Set"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "global", "false"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "vcs_repo.0.identifier", envGithubPolicySetIdentifier),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "vcs_repo.0.branch", "main"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "vcs_repo.0.ingress_submodules", "true"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "policies_path", envGithubPolicySetPath),
				),
			},
		},
	})
}

func TestAccTFEPolicySet_GithubApp(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	policySet := &tfe.PolicySet{}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccGHAInstallationPreCheck(t)
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
				t.Skip("Please set GITHUB_POLICY_SET_PATH to run this test")
			}
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySet_GithubApp(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetAttributes(policySet),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "name", "tst-terraform"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "description", "Policy Set"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "global", "false"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "vcs_repo.0.identifier", envGithubPolicySetIdentifier),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "vcs_repo.0.branch", "main"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "vcs_repo.0.ingress_submodules", "true"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "policies_path", envGithubPolicySetPath),
				),
			},
		},
	})
}

func TestAccTFEPolicySet_updateVCSBranch(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	policySet := &tfe.PolicySet{}

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
				t.Skip("Please set GITHUB_POLICY_SET_PATH to run this test")
			}
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySet_vcs(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetAttributes(policySet),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "name", "tst-terraform"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "description", "Policy Set"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "global", "false"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "vcs_repo.0.identifier", envGithubPolicySetIdentifier),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "vcs_repo.0.branch", "main"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "vcs_repo.0.ingress_submodules", "true"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "policies_path", envGithubPolicySetPath),
				),
			},

			{
				Config: testAccTFEPolicySet_updateVCSBranch(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetAttributes(policySet),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "name", "tst-terraform"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "description", "Policy Set"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "global", "false"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "vcs_repo.0.identifier", envGithubPolicySetIdentifier),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "vcs_repo.0.branch", envGithubPolicySetBranch),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "vcs_repo.0.ingress_submodules", "true"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "policies_path", envGithubPolicySetPath),
				),
			},
		},
	})
}

func TestAccTFEPolicySet_versionedSlug(t *testing.T) {
	skipIfUnitTest(t)

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	policySet := &tfe.PolicySet{}
	checksum, err := hashPolicies(testFixtureVersionFiles)
	if err != nil {
		t.Fatalf("Unable to generate checksum for policies %v", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySet_versionSlug(org.Name, testFixtureVersionFiles),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetAttributes(policySet),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "name", "tst-terraform"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "description", "Policy Set"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "slug.%", "2"),
					resource.TestCheckResourceAttrSet(
						"tfe_policy_set.foobar", "slug.source_path"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "slug.source_path", testFixtureVersionFiles),
					resource.TestCheckResourceAttrSet(
						"tfe_policy_set.foobar", "slug.id"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "slug.id", checksum),
				),
			},
		},
	})
}

func TestAccTFEPolicySet_versionedSlugUpdate(t *testing.T) {
	skipIfUnitTest(t)

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	policySet := &tfe.PolicySet{}

	originalChecksum, err := hashPolicies(testFixtureVersionFiles)
	if err != nil {
		t.Fatalf("Unable to generate checksum for policies %v", err)
	}

	newFile := fmt.Sprintf("%s/newfile.test.sentinel", testFixtureVersionFiles)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySet_versionSlug(org.Name, testFixtureVersionFiles),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetAttributes(policySet),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "name", "tst-terraform"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "description", "Policy Set"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "global", "false"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "slug.id", originalChecksum),
				),
			},
			{
				PreConfig: func() {
					err = os.WriteFile(newFile, []byte("main = rule { true }"), 0o755)
					if err != nil {
						t.Fatalf("error writing to file %s", newFile)
					}
					t.Cleanup(func() {
						os.Remove(newFile)
					})
				},
				Config: testAccTFEPolicySet_versionSlug(org.Name, testFixtureVersionFiles),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.foobar", policySet),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "name", "tst-terraform"),
					testAccCheckTFEPolicySetVersionValidateChecksum("tfe_policy_set.foobar", testFixtureVersionFiles),
				),
			},
		},
	})
}

func TestAccTFEPolicySet_versionedNoConflicts(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFEPolicySet_versionsConflict(org.Name, testFixtureVersionFiles),
				ExpectError: regexp.MustCompile(`Conflicting configuration`),
			},
		},
	})
}

func testAccCheckTFEPolicySetVersionValidateChecksum(n string, sourcePath string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		newChecksum, err := hashPolicies(sourcePath)
		if err != nil {
			return fmt.Errorf("unable to generate checksum for policies %w", err)
		}

		if rs.Primary.Attributes["slug.id"] != newChecksum {
			return fmt.Errorf("the new checksum for the policies contents did not match")
		}

		return nil
	}
}

func TestAccTFEPolicySet_invalidName(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFEPolicySet_invalidName(org.Name),
				ExpectError: regexp.MustCompile(`can only include letters, numbers, -, and _.`),
			},
		},
	})
}

func TestAccTFEPolicySetImport(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySet_populated(org.Name),
			},

			{
				ResourceName:      "tfe_policy_set.foobar",
				ImportState:       true,
				ImportStateVerify: true,
				// Note: We ignore the optional fields below, since the old API endpoints send empty values
				// and the results may vary depending on the API version
				ImportStateVerifyIgnore: []string{"kind", "overridable"},
			},
		},
	})
}

func testAccCheckTFEPolicySetExists(n string, policySet *tfe.PolicySet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(ConfiguredClient)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		ps, err := config.Client.PolicySets.Read(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}

		if ps.ID != rs.Primary.ID {
			return fmt.Errorf("PolicySet not found")
		}

		*policySet = *ps

		return nil
	}
}

func testAccCheckTFEPolicySetAttributes(policySet *tfe.PolicySet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if policySet.Name != "tst-terraform" {
			return fmt.Errorf("Bad name: %s", policySet.Name)
		}

		if policySet.Description != "Policy Set" {
			return fmt.Errorf("Bad description: %s", policySet.Description)
		}

		if policySet.Global {
			return fmt.Errorf("Bad value for global: %v", policySet.Global)
		}

		return nil
	}
}

func testAccCheckTFEPolicySetPopulated(policySet *tfe.PolicySet, orgName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(ConfiguredClient)

		if policySet.Name != "terraform-populated" {
			return fmt.Errorf("Bad name: %s", policySet.Name)
		}

		if policySet.Global {
			return fmt.Errorf("Bad value for global: %v", policySet.Global)
		}

		if len(policySet.Policies) != 1 {
			return fmt.Errorf("Wrong number of policies: %v", len(policySet.Policies))
		}

		policyID := policySet.Policies[0].ID
		policy, _ := config.Client.Policies.Read(ctx, policyID)
		if policy.Name != "policy-foo" {
			return fmt.Errorf("Wrong member policy: %v", policy.Name)
		}

		if len(policySet.Workspaces) != 1 {
			return fmt.Errorf("Wrong number of workspaces: %v", len(policySet.Workspaces))
		}

		workspaceID := policySet.Workspaces[0].ID
		workspace, _ := config.Client.Workspaces.Read(ctx, orgName, "workspace-foo")
		if workspace.ID != workspaceID {
			return fmt.Errorf("Wrong member workspace: %v", workspace.Name)
		}

		return nil
	}
}

func testAccCheckTFEPolicySetPopulatedUpdated(policySet *tfe.PolicySet, orgName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(ConfiguredClient)

		if policySet.Name != "terraform-populated-updated" {
			return fmt.Errorf("Bad name: %s", policySet.Name)
		}

		if policySet.Global {
			return fmt.Errorf("Bad value for global: %v", policySet.Global)
		}

		if len(policySet.Policies) != 1 {
			return fmt.Errorf("Wrong number of policies: %v", len(policySet.Policies))
		}

		policyID := policySet.Policies[0].ID
		policy, _ := config.Client.Policies.Read(ctx, policyID)
		if policy.Name != "policy-bar" {
			return fmt.Errorf("Wrong member policy: %v", policy.Name)
		}

		if len(policySet.Workspaces) != 1 {
			return fmt.Errorf("Wrong number of workspaces: %v", len(policySet.Workspaces))
		}

		workspaceID := policySet.Workspaces[0].ID
		workspace, _ := config.Client.Workspaces.Read(ctx, orgName, "workspace-bar")
		if workspace.ID != workspaceID {
			return fmt.Errorf("Wrong member workspace: %v", workspace.Name)
		}

		return nil
	}
}

func testAccCheckTFEPolicySetGlobal(policySet *tfe.PolicySet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(ConfiguredClient)

		if policySet.Name != "terraform-global" {
			return fmt.Errorf("Bad name: %s", policySet.Name)
		}

		if !policySet.Global {
			return fmt.Errorf("Bad value for global: %v", policySet.Global)
		}

		if len(policySet.Policies) != 1 {
			return fmt.Errorf("Wrong number of policies: %v", len(policySet.Policies))
		}

		policyID := policySet.Policies[0].ID
		policy, _ := config.Client.Policies.Read(ctx, policyID)
		if policy.Name != "policy-foo" {
			return fmt.Errorf("Wrong member policy: %v", policy.Name)
		}

		// No workspaces are returned for global policy sets
		if len(policySet.Workspaces) != 0 {
			return fmt.Errorf("Wrong number of workspaces: %v", len(policySet.Workspaces))
		}

		return nil
	}
}

func testAccCheckTFEPolicySetDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(ConfiguredClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_policy_set" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := config.Client.PolicySets.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Sentinel policy %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccTFEPolicySet_basic(organization string) string {
	return fmt.Sprintf(`
resource "tfe_sentinel_policy" "foo" {
  name         = "policy-foo"
  policy       = "main = rule { true }"
  organization = "%s"
}

resource "tfe_policy_set" "foobar" {
  name         = "tst-terraform"
  description  = "Policy Set"
  organization = "%s"
  policy_ids   = [tfe_sentinel_policy.foo.id]
}`, organization, organization)
}

func testAccTFEPolicySetOPA_basic(organization string) string {
	return fmt.Sprintf(`
resource "tfe_policy_set" "foobar" {
  name         = "tst-terraform"
  description  = "Policy Set"
  organization = "%s"
  kind         = "opa"
  overridable = "true"
}`, organization)
}

func testAccTFEPolicySet_empty(organization string) string {
	return fmt.Sprintf(`
 resource "tfe_policy_set" "foobar" {
  name         = "tst-terraform"
  description  = "Policy Set"
  organization = "%s"
}`, organization)
}

func testAccTFEPolicySet_populated(organization string) string {
	return fmt.Sprintf(`
locals {
    organization_name = "%s"
}

resource "tfe_sentinel_policy" "foo" {
  name         = "policy-foo"
  policy       = "main = rule { true }"
  organization = local.organization_name
}

resource "tfe_workspace" "foo" {
  name         = "workspace-foo"
  organization = local.organization_name
}

resource "tfe_policy_set" "foobar" {
  name          = "terraform-populated"
  organization = local.organization_name
  policy_ids    = [tfe_sentinel_policy.foo.id]
  workspace_ids = [tfe_workspace.foo.id]
}`, organization)
}

func testAccTFEPolicySetOPA_overridable(organization string) string {
	return fmt.Sprintf(`
locals {
    organization_name = "%s"
}

resource "tfe_workspace" "foo" {
  name         = "workspace-foo"
  organization = local.organization_name
}

resource "tfe_policy_set" "foobar" {
  name          = "tst-terraform"
  organization = local.organization_name
  workspace_ids = [tfe_workspace.foo.id]
  overridable = "false"
  kind = "opa"
}`, organization)
}

func testAccTFEPolicySet_updatePopulated(organization string) string {
	return fmt.Sprintf(`
locals {
    organization_name = "%s"
}

resource "tfe_sentinel_policy" "foo" {
  name         = "policy-foo"
  policy       = "main = rule { true }"
  organization = local.organization_name
}

resource "tfe_sentinel_policy" "bar" {
  name         = "policy-bar"
  policy       = "main = rule { false }"
  organization = local.organization_name
}

resource "tfe_workspace" "foo" {
  name         = "workspace-foo"
  organization = local.organization_name
}

resource "tfe_workspace" "bar" {
  name         = "workspace-bar"
  organization = local.organization_name
}

resource "tfe_policy_set" "foobar" {
  name          = "terraform-populated-updated"
  organization = local.organization_name
  policy_ids    = [tfe_sentinel_policy.bar.id]
  workspace_ids = [tfe_workspace.bar.id]
}`, organization)
}

func testAccTFEPolicySet_global(organization string) string {
	return fmt.Sprintf(`
locals {
    organization_name = "%s"
}

resource "tfe_sentinel_policy" "foo" {
  name         = "policy-foo"
  policy       = "main = rule { true }"
  organization = local.organization_name
}

resource "tfe_workspace" "foo" {
  name         = "workspace-foo"
  organization = local.organization_name
}

resource "tfe_policy_set" "foobar" {
  name         = "terraform-global"
  organization = local.organization_name
  global       = true
  policy_ids   = [tfe_sentinel_policy.foo.id]
}`, organization)
}

func testAccTFEPolicySet_vcs(organization string) string {
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
  name         = "tst-terraform"
  description  = "Policy Set"
  organization     = local.organization_name
  vcs_repo {
    identifier         = "%s"
    branch             = "main"
    ingress_submodules = true
    oauth_token_id     = tfe_oauth_client.test.oauth_token_id
  }

  policies_path = "%s"
}
`, organization,
		envGithubToken,
		envGithubPolicySetIdentifier,
		envGithubPolicySetPath,
	)
}

func testAccTFEPolicySet_GithubApp(organization string) string {
	return fmt.Sprintf(`
locals {
    organization_name = "%s"
}

resource "tfe_policy_set" "foobar" {
  name         = "tst-terraform"
  description  = "Policy Set"
  organization     = local.organization_name
  vcs_repo {
    identifier         = "%s"
    branch             = "main"
    ingress_submodules = true
    github_app_installation_id = "%s"
  }

  policies_path = "%s"
}
`, organization,
		envGithubPolicySetIdentifier,
		envGithubAppInstallationID,
		envGithubPolicySetPath,
	)
}

func testAccTFEPolicySet_updateVCSBranch(organization string) string {
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
  name         = "tst-terraform"
  description  = "Policy Set"
  organization     = local.organization_name
  vcs_repo {
    identifier         = "%s"
    branch             = "%s"
    ingress_submodules = true
    oauth_token_id     = tfe_oauth_client.test.oauth_token_id
  }

  policies_path = "%s"
}
`, organization,
		envGithubToken,
		envGithubPolicySetIdentifier,
		envGithubPolicySetBranch,
		envGithubPolicySetPath,
	)
}

func testAccTFEPolicySet_invalidName(organization string) string {
	return fmt.Sprintf(`
locals {
    organization_name = "%s"
}

resource "tfe_sentinel_policy" "foo" {
  name         = "policy-foo"
  policy       = "main = rule { true }"
  organization = local.organization_name
}

resource "tfe_policy_set" "foobar" {
  name         = "not the right format"
  description  = "Policy Set"
  organization = local.organization_name
  policy_ids   = [tfe_sentinel_policy.foo.id]
}`, organization)
}

func testAccTFEPolicySet_versionSlug(organization string, sourcePath string) string {
	return fmt.Sprintf(`
data "tfe_slug" "policy" {
  source_path = "%s"
}

resource "tfe_policy_set" "foobar" {
  name         = "tst-terraform"
  description  = "Policy Set"
  organization = "%s"
  slug         = data.tfe_slug.policy
}`, sourcePath, organization)
}

func testAccTFEPolicySet_versionsConflict(organization string, sourcePath string) string {
	return fmt.Sprintf(`
data "tfe_slug" "policy" {
  source_path = "%s"
}

resource "tfe_policy_set" "foobar" {
  name         = "tst-terraform"
  description  = "Policy Set"
  organization = "%s"
	slug = data.tfe_slug.policy
  vcs_repo {
    identifier         = "foo"
    branch             = "foo"
    ingress_submodules = true
    oauth_token_id     = "id"
  }
} `, sourcePath, organization)
}
