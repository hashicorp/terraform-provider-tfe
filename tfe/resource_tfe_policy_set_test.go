package tfe

import (
	"fmt"
	"testing"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccTFEPolicySet_basic(t *testing.T) {
	policySet := &tfe.PolicySet{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySet_basic,
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
	policySet := &tfe.PolicySet{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySet_basic,
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
				Config: testAccTFEPolicySet_populated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetPopulated(policySet),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "name", "terraform-populated"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "global", "false"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "policy_ids.#", "1"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "workspace_external_ids.#", "1"),
				),
			},
		},
	})
}

func TestAccTFEPolicySet_updateWorkspaceIDs(t *testing.T) {
	policySet := &tfe.PolicySet{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySet_basic,
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
				Config: testAccTFEPolicySet_populated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetPopulated(policySet),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "name", "terraform-populated"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "global", "false"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "policy_ids.#", "1"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "workspace_external_ids.#", "1"),
				),
			},
			{
				Config: testAccTFEPolicySet_populatedWorkspaceIDs,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetPopulated(policySet),
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
	policySet := &tfe.PolicySet{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySet_basic,
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
				Config: testAccTFEPolicySet_empty,
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
	policySet := &tfe.PolicySet{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySet_populated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetPopulated(policySet),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "name", "terraform-populated"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "global", "false"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "policy_ids.#", "1"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "workspace_external_ids.#", "1"),
				),
			},

			{
				Config: testAccTFEPolicySet_updatePopulated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetPopulatedUpdated(policySet),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "name", "terraform-populated-updated"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "global", "false"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "policy_ids.#", "1"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "workspace_external_ids.#", "1"),
				),
			},
		},
	})
}

func TestAccTFEPolicySet_updatePopulatedWorkspaceIDs(t *testing.T) {
	policySet := &tfe.PolicySet{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySet_populatedWorkspaceIDs,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetPopulated(policySet),
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
				Config: testAccTFEPolicySet_updatePopulatedWorkspaceIDs,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetPopulatedUpdated(policySet),
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

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySet_populated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetPopulated(policySet),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "name", "terraform-populated"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "global", "false"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "policy_ids.#", "1"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "workspace_external_ids.#", "1"),
				),
			},

			{
				Config: testAccTFEPolicySet_global,
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

func TestAccTFEPolicySet_updateWorkspaceIDsToGlobal(t *testing.T) {
	policySet := &tfe.PolicySet{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySet_populatedWorkspaceIDs,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetPopulated(policySet),
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
				Config: testAccTFEPolicySet_global,
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
	policySet := &tfe.PolicySet{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySet_global,
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
				Config: testAccTFEPolicySet_populated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetPopulated(policySet),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "name", "terraform-populated"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "global", "false"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "policy_ids.#", "1"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "workspace_external_ids.#", "1"),
				),
			},
		},
	})
}

func TestAccTFEPolicySet_updateGlobalToWorkspaceIDs(t *testing.T) {
	policySet := &tfe.PolicySet{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySet_global,
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
				Config: testAccTFEPolicySet_populatedWorkspaceIDs,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetPopulated(policySet),
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
	policySet := &tfe.PolicySet{}

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
				Config: testAccTFEPolicySet_vcs,
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

func TestAccTFEPolicySet_updateToVcs(t *testing.T) {
	policySet := &tfe.PolicySet{}

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
				Config: testAccTFEPolicySet_vcs,
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
			{
				Config: testAccTFEPolicySet_updateVCSRepoBranch,
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
						"tfe_policy_set.foobar", "vcs_repo.0.branch", "test"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "vcs_repo.0.ingress_submodules", "true"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "policies_path", GITHUB_POLICY_SET_PATH),
				),
			},
		},
	})
}

func TestAccTFEPolicySetImport(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySet_populatedWorkspaceIDs,
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

func testAccCheckTFEPolicySetPopulated(policySet *tfe.PolicySet) resource.TestCheckFunc {
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
		workspace, _ := tfeClient.Workspaces.Read(ctx, "tst-terraform", "workspace-foo")
		if workspace.ID != workspaceID {
			return fmt.Errorf("Wrong member workspace: %v", workspace.Name)
		}

		return nil
	}
}

func testAccCheckTFEPolicySetPopulatedUpdated(policySet *tfe.PolicySet) resource.TestCheckFunc {
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
		workspace, _ := tfeClient.Workspaces.Read(ctx, "tst-terraform", "workspace-bar")
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

const testAccTFEPolicySet_basic = `
resource "tfe_organization" "foobar" {
  name  = "tst-terraform"
  email = "admin@company.com"
}

resource "tfe_sentinel_policy" "foo" {
  name         = "policy-foo"
  policy       = "main = rule { true }"
  organization = "${tfe_organization.foobar.id}"
}

resource "tfe_policy_set" "foobar" {
  name         = "tst-terraform"
  description  = "Policy Set"
  organization = "${tfe_organization.foobar.id}"
  policy_ids   = ["${tfe_sentinel_policy.foo.id}"]
}`

const testAccTFEPolicySet_empty = `
resource "tfe_organization" "foobar" {
  name  = "tst-terraform"
  email = "admin@company.com"
}
 resource "tfe_policy_set" "foobar" {
  name         = "tst-terraform"
  description  = "Policy Set"
  organization = "${tfe_organization.foobar.id}"
}`

const testAccTFEPolicySet_populated = `
resource "tfe_organization" "foobar" {
  name  = "tst-terraform"
  email = "admin@company.com"
}

resource "tfe_sentinel_policy" "foo" {
  name         = "policy-foo"
  policy       = "main = rule { true }"
  organization = "${tfe_organization.foobar.id}"
}

resource "tfe_workspace" "foo" {
  name         = "workspace-foo"
  organization = "${tfe_organization.foobar.id}"
}

resource "tfe_policy_set" "foobar" {
  name                   = "terraform-populated"
  organization           = "${tfe_organization.foobar.id}"
  policy_ids             = ["${tfe_sentinel_policy.foo.id}"]
  workspace_external_ids = ["${tfe_workspace.foo.id}"]
}`

const testAccTFEPolicySet_populatedWorkspaceIDs = `
resource "tfe_organization" "foobar" {
  name  = "tst-terraform"
  email = "admin@company.com"
}

resource "tfe_sentinel_policy" "foo" {
  name         = "policy-foo"
  policy       = "main = rule { true }"
  organization = "${tfe_organization.foobar.id}"
}

resource "tfe_workspace" "foo" {
  name         = "workspace-foo"
  organization = "${tfe_organization.foobar.id}"
}

resource "tfe_policy_set" "foobar" {
  name          = "terraform-populated"
  organization  = "${tfe_organization.foobar.id}"
  policy_ids    = ["${tfe_sentinel_policy.foo.id}"]
  workspace_ids = ["${tfe_workspace.foo.id}"]
}`

const testAccTFEPolicySet_updatePopulated = `
resource "tfe_organization" "foobar" {
  name  = "tst-terraform"
  email = "admin@company.com"
}

resource "tfe_sentinel_policy" "foo" {
  name         = "policy-foo"
  policy       = "main = rule { true }"
  organization = "${tfe_organization.foobar.id}"
}

resource "tfe_sentinel_policy" "bar" {
  name         = "policy-bar"
  policy       = "main = rule { false }"
  organization = "${tfe_organization.foobar.id}"
}

resource "tfe_workspace" "foo" {
  name         = "workspace-foo"
  organization = "${tfe_organization.foobar.id}"
}

resource "tfe_workspace" "bar" {
  name         = "workspace-bar"
  organization = "${tfe_organization.foobar.id}"
}

resource "tfe_policy_set" "foobar" {
  name                   = "terraform-populated-updated"
  organization           = "${tfe_organization.foobar.id}"
  policy_ids             = ["${tfe_sentinel_policy.bar.id}"]
  workspace_external_ids = ["${tfe_workspace.bar.id}"]
}`

const testAccTFEPolicySet_updatePopulatedWorkspaceIDs = `
resource "tfe_organization" "foobar" {
  name  = "tst-terraform"
  email = "admin@company.com"
}

resource "tfe_sentinel_policy" "foo" {
  name         = "policy-foo"
  policy       = "main = rule { true }"
  organization = "${tfe_organization.foobar.id}"
}

resource "tfe_sentinel_policy" "bar" {
  name         = "policy-bar"
  policy       = "main = rule { false }"
  organization = "${tfe_organization.foobar.id}"
}

resource "tfe_workspace" "foo" {
  name         = "workspace-foo"
  organization = "${tfe_organization.foobar.id}"
}

resource "tfe_workspace" "bar" {
  name         = "workspace-bar"
  organization = "${tfe_organization.foobar.id}"
}

resource "tfe_policy_set" "foobar" {
  name          = "terraform-populated-updated"
  organization  = "${tfe_organization.foobar.id}"
  policy_ids    = ["${tfe_sentinel_policy.bar.id}"]
  workspace_ids = ["${tfe_workspace.bar.id}"]
}`

const testAccTFEPolicySet_global = `
resource "tfe_organization" "foobar" {
  name  = "tst-terraform"
  email = "admin@company.com"
}

resource "tfe_sentinel_policy" "foo" {
  name         = "policy-foo"
  policy       = "main = rule { true }"
  organization = "${tfe_organization.foobar.id}"
}

resource "tfe_workspace" "foo" {
  name         = "workspace-foo"
  organization = "${tfe_organization.foobar.id}"
}

resource "tfe_policy_set" "foobar" {
  name         = "terraform-global"
  organization = "${tfe_organization.foobar.id}"
  global       = true
  policy_ids   = ["${tfe_sentinel_policy.foo.id}"]
}`

var testAccTFEPolicySet_vcs = fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform"
  email = "admin@company.com"
}

resource "tfe_oauth_client" "test" {
  organization     = "${tfe_organization.foobar.id}"
  api_url          = "https://api.github.com"
  http_url         = "https://github.com"
  oauth_token      = "%s"
  service_provider = "github"
}

resource "tfe_policy_set" "foobar" {
  name         = "tst-terraform"
  description  = "Policy Set"
  organization = "${tfe_organization.foobar.id}"
  vcs_repo {
    identifier         = "%s"
    branch             = "%s"
    ingress_submodules = true
    oauth_token_id     = "${tfe_oauth_client.test.oauth_token_id}"
  }

  policies_path = "%s"
}
`,
	GITHUB_TOKEN,
	GITHUB_POLICY_SET_IDENTIFIER,
	GITHUB_POLICY_SET_BRANCH,
	GITHUB_POLICY_SET_PATH,
)

var testAccTFEPolicySet_updateVCSRepoBranch = fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform"
  email = "admin@company.com"
}

resource "tfe_oauth_client" "test" {
  organization     = "${tfe_organization.foobar.id}"
  api_url          = "https://api.github.com"
  http_url         = "https://github.com"
  oauth_token      = "%s"
  service_provider = "github"
}

resource "tfe_policy_set" "foobar" {
  name         = "tst-terraform"
  description  = "Policy Set"
  organization = "${tfe_organization.foobar.id}"
  vcs_repo {
    identifier         = "%s"
    branch             = "test"
    ingress_submodules = true
    oauth_token_id     = "${tfe_oauth_client.test.oauth_token_id}"
  }

  policies_path = "%s"
}
`,
	GITHUB_TOKEN,
	GITHUB_POLICY_SET_IDENTIFIER,
	GITHUB_POLICY_SET_PATH,
)
