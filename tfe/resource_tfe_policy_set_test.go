package tfe

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"regexp"
	"testing"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccTFEPolicySet_basic(t *testing.T) {
	skipIfFreeOnly(t)

	policySet := &tfe.PolicySet{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySet_basic(rInt),
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

func TestAccTFEPolicySet_update(t *testing.T) {
	skipIfFreeOnly(t)

	policySet := &tfe.PolicySet{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySet_basic(rInt),
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
				Config: testAccTFEPolicySet_populated(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetPopulated(policySet, orgName),
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

func TestAccTFEPolicySet_updateEmpty(t *testing.T) {
	skipIfFreeOnly(t)

	policySet := &tfe.PolicySet{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySet_basic(rInt),
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
				Config: testAccTFEPolicySet_empty(rInt),
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
	skipIfFreeOnly(t)

	policySet := &tfe.PolicySet{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySet_populated(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetPopulated(policySet, orgName),
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
				Config: testAccTFEPolicySet_updatePopulated(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetPopulatedUpdated(policySet, orgName),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "name", "terraform-populated-updated"),
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

func TestAccTFEPolicySet_updateToGlobal(t *testing.T) {
	policySet := &tfe.PolicySet{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySet_populated(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetPopulated(policySet, orgName),
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
				Config: testAccTFEPolicySet_global(rInt),
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
		},
	})
}

func TestAccTFEPolicySet_updateToWorkspace(t *testing.T) {
	skipIfFreeOnly(t)

	policySet := &tfe.PolicySet{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySet_global(rInt),
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
				Config: testAccTFEPolicySet_populated(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetPopulated(policySet, orgName),
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
	skipIfFreeOnly(t)

	policySet := &tfe.PolicySet{}
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
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySet_vcs(rInt),
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
						"tfe_policy_set.foobar", "vcs_repo.0.identifier", GITHUB_POLICY_SET_IDENTIFIER),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "vcs_repo.0.branch", "main"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "vcs_repo.0.ingress_submodules", "true"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "policies_path", GITHUB_POLICY_SET_PATH),
				),
			},
		},
	})
}

func TestAccTFEPolicySet_updateVCSBranch(t *testing.T) {
	skipIfFreeOnly(t)

	policySet := &tfe.PolicySet{}
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
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySet_vcs(rInt),
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
						"tfe_policy_set.foobar", "vcs_repo.0.identifier", GITHUB_POLICY_SET_IDENTIFIER),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "vcs_repo.0.branch", "main"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "vcs_repo.0.ingress_submodules", "true"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "policies_path", GITHUB_POLICY_SET_PATH),
				),
			},

			{
				Config: testAccTFEPolicySet_updateVCSBranch(rInt),
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
						"tfe_policy_set.foobar", "vcs_repo.0.identifier", GITHUB_POLICY_SET_IDENTIFIER),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "vcs_repo.0.branch", GITHUB_POLICY_SET_BRANCH),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "vcs_repo.0.ingress_submodules", "true"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "policies_path", GITHUB_POLICY_SET_PATH),
				),
			},
		},
	})
}

func TestAccTFEPolicySet_versionedSlug(t *testing.T) {
	skipIfFreeOnly(t)

	policySet := &tfe.PolicySet{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
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
				Config: testAccTFEPolicySet_versionSlug(rInt, testFixtureVersionFiles),
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
	skipIfFreeOnly(t)

	policySet := &tfe.PolicySet{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	originalChecksum, err := hashPolicies(testFixtureVersionFiles)
	if err != nil {
		t.Fatalf("Unable to generate checksum for policies %v", err)
	}

	newFile := fmt.Sprintf("%s/newfile.test.sentinel", testFixtureVersionFiles)
	removeFile := func() {
		// This func is used below, that is why it is not an anonymous function.
		// It is used because if there is a test fatal (t.Fatal), then defer does
		// not get called. So we call this `removeFile` function both in the defer
		// and explicitly below.
		err := os.Remove(newFile)
		if err != nil {
			t.Fatalf("Error removing file %v", err)
		}
	}
	defer removeFile()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySet_versionSlug(rInt, testFixtureVersionFiles),
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
					err = ioutil.WriteFile(newFile, []byte("main = rule { true }"), 0755)
					if err != nil {
						// this function is called here as well as the defer because
						// when t.Fatal is called, it exits the program and ignores defers.
						removeFile()
						t.Fatalf("error writing to file %s", newFile)
					}
				},
				Config: testAccTFEPolicySet_versionSlug(rInt, testFixtureVersionFiles),
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
	skipIfFreeOnly(t)

	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFEPolicySet_versionsConflict(rInt, testFixtureVersionFiles),
				ExpectError: regexp.MustCompile(`Conflicting configuration arguments`),
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
			return fmt.Errorf("Unable to generate checksum for policies %v", err)
		}

		if rs.Primary.Attributes["slug.id"] != newChecksum {
			return fmt.Errorf("The new checksum for the policies contents did not match")
		}

		return nil
	}
}

func TestAccTFEPolicySet_invalidName(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFEPolicySet_invalidName(rInt),
				ExpectError: regexp.MustCompile(`can only include letters, numbers, -, and _.`),
			},
		},
	})
}

func TestAccTFEPolicySetImport(t *testing.T) {
	skipIfFreeOnly(t)

	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySet_populated(rInt),
			},

			{
				ResourceName:      "tfe_policy_set.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckTFEPolicySetExists(n string, policySet *tfe.PolicySet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		tfeClient := testAccProvider.Meta().(*tfe.Client)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		ps, err := tfeClient.PolicySets.Read(ctx, rs.Primary.ID)
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
		tfeClient := testAccProvider.Meta().(*tfe.Client)

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
		policy, _ := tfeClient.Policies.Read(ctx, policyID)
		if policy.Name != "policy-foo" {
			return fmt.Errorf("Wrong member policy: %v", policy.Name)
		}

		if len(policySet.Workspaces) != 1 {
			return fmt.Errorf("Wrong number of workspaces: %v", len(policySet.Workspaces))
		}

		workspaceID := policySet.Workspaces[0].ID
		workspace, _ := tfeClient.Workspaces.Read(ctx, orgName, "workspace-foo")
		if workspace.ID != workspaceID {
			return fmt.Errorf("Wrong member workspace: %v", workspace.Name)
		}

		return nil
	}
}

func testAccCheckTFEPolicySetPopulatedUpdated(policySet *tfe.PolicySet, orgName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		tfeClient := testAccProvider.Meta().(*tfe.Client)

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
		policy, _ := tfeClient.Policies.Read(ctx, policyID)
		if policy.Name != "policy-bar" {
			return fmt.Errorf("Wrong member policy: %v", policy.Name)
		}

		if len(policySet.Workspaces) != 1 {
			return fmt.Errorf("Wrong number of workspaces: %v", len(policySet.Workspaces))
		}

		workspaceID := policySet.Workspaces[0].ID
		workspace, _ := tfeClient.Workspaces.Read(ctx, orgName, "workspace-bar")
		if workspace.ID != workspaceID {
			return fmt.Errorf("Wrong member workspace: %v", workspace.Name)
		}

		return nil
	}
}

func testAccCheckTFEPolicySetGlobal(policySet *tfe.PolicySet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		tfeClient := testAccProvider.Meta().(*tfe.Client)

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
		policy, _ := tfeClient.Policies.Read(ctx, policyID)
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
	tfeClient := testAccProvider.Meta().(*tfe.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_policy_set" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := tfeClient.PolicySets.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Sentinel policy %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccTFEPolicySet_basic(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_sentinel_policy" "foo" {
  name         = "policy-foo"
  policy       = "main = rule { true }"
  organization = tfe_organization.foobar.id
}

resource "tfe_policy_set" "foobar" {
  name         = "tst-terraform"
  description  = "Policy Set"
  organization = tfe_organization.foobar.id
  policy_ids   = [tfe_sentinel_policy.foo.id]
}`, rInt)
}

func testAccTFEPolicySet_empty(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}
 resource "tfe_policy_set" "foobar" {
  name         = "tst-terraform"
  description  = "Policy Set"
  organization = tfe_organization.foobar.id
}`, rInt)
}

func testAccTFEPolicySet_populated(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_sentinel_policy" "foo" {
  name         = "policy-foo"
  policy       = "main = rule { true }"
  organization = tfe_organization.foobar.id
}

resource "tfe_workspace" "foo" {
  name         = "workspace-foo"
  organization = tfe_organization.foobar.id
}

resource "tfe_policy_set" "foobar" {
  name          = "terraform-populated"
  organization  = tfe_organization.foobar.id
  policy_ids    = [tfe_sentinel_policy.foo.id]
  workspace_ids = [tfe_workspace.foo.id]
}`, rInt)
}

func testAccTFEPolicySet_updatePopulated(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_sentinel_policy" "foo" {
  name         = "policy-foo"
  policy       = "main = rule { true }"
  organization = tfe_organization.foobar.id
}

resource "tfe_sentinel_policy" "bar" {
  name         = "policy-bar"
  policy       = "main = rule { false }"
  organization = tfe_organization.foobar.id
}

resource "tfe_workspace" "foo" {
  name         = "workspace-foo"
  organization = tfe_organization.foobar.id
}

resource "tfe_workspace" "bar" {
  name         = "workspace-bar"
  organization = tfe_organization.foobar.id
}

resource "tfe_policy_set" "foobar" {
  name          = "terraform-populated-updated"
  organization  = tfe_organization.foobar.id
  policy_ids    = [tfe_sentinel_policy.bar.id]
  workspace_ids = [tfe_workspace.bar.id]
}`, rInt)
}

func testAccTFEPolicySet_global(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_sentinel_policy" "foo" {
  name         = "policy-foo"
  policy       = "main = rule { true }"
  organization = tfe_organization.foobar.id
}

resource "tfe_workspace" "foo" {
  name         = "workspace-foo"
  organization = tfe_organization.foobar.id
}

resource "tfe_policy_set" "foobar" {
  name         = "terraform-global"
  organization = tfe_organization.foobar.id
  global       = true
  policy_ids   = [tfe_sentinel_policy.foo.id]
}`, rInt)
}

func testAccTFEPolicySet_vcs(rInt int) string {
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
  name         = "tst-terraform"
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
`, rInt,
		GITHUB_TOKEN,
		GITHUB_POLICY_SET_IDENTIFIER,
		GITHUB_POLICY_SET_PATH,
	)
}

func testAccTFEPolicySet_updateVCSBranch(rInt int) string {
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
  name         = "tst-terraform"
  description  = "Policy Set"
  organization = tfe_organization.foobar.id
  vcs_repo {
    identifier         = "%s"
    branch             = "%s"
    ingress_submodules = true
    oauth_token_id     = tfe_oauth_client.test.oauth_token_id
  }

  policies_path = "%s"
}
`, rInt,
		GITHUB_TOKEN,
		GITHUB_POLICY_SET_IDENTIFIER,
		GITHUB_POLICY_SET_BRANCH,
		GITHUB_POLICY_SET_PATH,
	)
}

func testAccTFEPolicySet_invalidName(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_sentinel_policy" "foo" {
  name         = "policy-foo"
  policy       = "main = rule { true }"
  organization = tfe_organization.foobar.id
}

resource "tfe_policy_set" "foobar" {
  name         = "not the right format"
  description  = "Policy Set"
  organization = tfe_organization.foobar.id
  policy_ids   = [tfe_sentinel_policy.foo.id]
}`, rInt)
}

func testAccTFEPolicySet_versionSlug(rInt int, sourcePath string) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

data "tfe_slug" "policy" {
  source_path = "%s"
}

resource "tfe_policy_set" "foobar" {
  name         = "tst-terraform"
  description  = "Policy Set"
  organization = tfe_organization.foobar.id
	slug = data.tfe_slug.policy
}`, rInt, sourcePath)
}

func testAccTFEPolicySet_versionsConflict(rInt int, sourcePath string) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

data "tfe_slug" "policy" {
  source_path = "%s"
}

resource "tfe_policy_set" "foobar" {
  name         = "tst-terraform"
  description  = "Policy Set"
  organization = tfe_organization.foobar.id
	slug = data.tfe_slug.policy
  vcs_repo {
    identifier         = "foo"
    branch             = "foo"
    ingress_submodules = true
    oauth_token_id     = "id"
  }
} `, rInt, sourcePath)
}
