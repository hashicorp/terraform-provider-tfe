package tfe

import (
	"fmt"
	"testing"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccTFEPolicySetCreate_basic(t *testing.T) {
	policySet := &tfe.PolicySet{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccTFEPolicySet_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.my_set", policySet),
					testAccCheckTFEPolicySetAttributes(policySet),
					resource.TestCheckResourceAttr("tfe_policy_set.my_set", "name", "my-set"),
					resource.TestCheckResourceAttr("tfe_policy_set.my_set", "description", "A set for some of my Sentinel policies"),
					resource.TestCheckResourceAttr("tfe_policy_set.my_set", "global", "false"),
					resource.TestCheckResourceAttr("tfe_policy_set.my_set", "policy_ids.#", "0"),
					resource.TestCheckResourceAttr("tfe_policy_set.my_set", "workspace_external_ids.#", "0"),
				),
			},
		},
	})
}

func TestAccTFEPolicySetCreate_populated(t *testing.T) {
	policySet := &tfe.PolicySet{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccTFEPolicySet_populated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.my_set", policySet),
					testAccCheckTFEPolicySetPopulated(policySet),
					resource.TestCheckResourceAttr("tfe_policy_set.my_set", "name", "populated-set"),
					resource.TestCheckResourceAttr("tfe_policy_set.my_set", "description", "A set populated with some policies"),
					resource.TestCheckResourceAttr("tfe_policy_set.my_set", "global", "false"),
					resource.TestCheckResourceAttr("tfe_policy_set.my_set", "policy_ids.#", "1"),
					resource.TestCheckResourceAttr("tfe_policy_set.my_set", "workspace_external_ids.#", "1"),
				),
			},
		},
	})
}

func TestAccTFEPolicySetUpdate_basic(t *testing.T) {
	policySet := &tfe.PolicySet{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccTFEPolicySet_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.my_set", policySet),
					testAccCheckTFEPolicySetAttributes(policySet),
					resource.TestCheckResourceAttr("tfe_policy_set.my_set", "name", "my-set"),
					resource.TestCheckResourceAttr("tfe_policy_set.my_set", "description", "A set for some of my Sentinel policies"),
					resource.TestCheckResourceAttr("tfe_policy_set.my_set", "global", "false"),
					resource.TestCheckResourceAttr("tfe_policy_set.my_set", "policy_ids.#", "0"),
					resource.TestCheckResourceAttr("tfe_policy_set.my_set", "workspace_external_ids.#", "0"),
				),
			},

			resource.TestStep{
				Config: testAccTFEPolicySet_populated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.my_set", policySet),
					testAccCheckTFEPolicySetPopulated(policySet),
					resource.TestCheckResourceAttr("tfe_policy_set.my_set", "name", "populated-set"),
					resource.TestCheckResourceAttr("tfe_policy_set.my_set", "description", "A set populated with some policies"),
					resource.TestCheckResourceAttr("tfe_policy_set.my_set", "global", "false"),
					resource.TestCheckResourceAttr("tfe_policy_set.my_set", "policy_ids.#", "1"),
					resource.TestCheckResourceAttr("tfe_policy_set.my_set", "workspace_external_ids.#", "1"),
				),
			},
		},
	})
}

func TestAccTFEPolicySetUpdate_populated(t *testing.T) {
	policySet := &tfe.PolicySet{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccTFEPolicySet_populated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.my_set", policySet),
					testAccCheckTFEPolicySetPopulated(policySet),
					resource.TestCheckResourceAttr("tfe_policy_set.my_set", "name", "populated-set"),
					resource.TestCheckResourceAttr("tfe_policy_set.my_set", "description", "A set populated with some policies"),
					resource.TestCheckResourceAttr("tfe_policy_set.my_set", "global", "false"),
					resource.TestCheckResourceAttr("tfe_policy_set.my_set", "policy_ids.#", "1"),
					resource.TestCheckResourceAttr("tfe_policy_set.my_set", "workspace_external_ids.#", "1"),
				),
			},

			resource.TestStep{
				Config: testAccTFEPolicySet_newMembers,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.my_set", policySet),
					testAccCheckTFEPolicySetNewMembers(policySet),
					resource.TestCheckResourceAttr("tfe_policy_set.my_set", "name", "populated-set"),
					resource.TestCheckResourceAttr("tfe_policy_set.my_set", "description", "A set populated with some policies"),
					resource.TestCheckResourceAttr("tfe_policy_set.my_set", "global", "false"),
					resource.TestCheckResourceAttr("tfe_policy_set.my_set", "policy_ids.#", "1"),
					resource.TestCheckResourceAttr("tfe_policy_set.my_set", "workspace_external_ids.#", "1"),
				),
			},
		},
	})
}

func TestAccTFEPolicySetUpdate_global(t *testing.T) {
	policySet := &tfe.PolicySet{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccTFEPolicySet_populated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.my_set", policySet),
					testAccCheckTFEPolicySetPopulated(policySet),
					resource.TestCheckResourceAttr("tfe_policy_set.my_set", "name", "populated-set"),
					resource.TestCheckResourceAttr("tfe_policy_set.my_set", "description", "A set populated with some policies"),
					resource.TestCheckResourceAttr("tfe_policy_set.my_set", "global", "false"),
					resource.TestCheckResourceAttr("tfe_policy_set.my_set", "policy_ids.#", "1"),
					resource.TestCheckResourceAttr("tfe_policy_set.my_set", "workspace_external_ids.#", "1"),
				),
			},

			resource.TestStep{
				Config: testAccTFEPolicySet_global,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.my_set", policySet),
					testAccCheckTFEPolicySetGlobal(policySet),
					resource.TestCheckResourceAttr("tfe_policy_set.my_set", "name", "global-set"),
					resource.TestCheckResourceAttr("tfe_policy_set.my_set", "description", "A global set populated with some policies"),
					resource.TestCheckResourceAttr("tfe_policy_set.my_set", "global", "true"),
					resource.TestCheckResourceAttr("tfe_policy_set.my_set", "policy_ids.#", "1"),
					resource.TestCheckResourceAttr("tfe_policy_set.my_set", "workspace_external_ids.#", "0"),
				),
			},
		},
	})
}

func TestAccTFEPolicySetUpdate_workspaceSwap(t *testing.T) {
	policySet := &tfe.PolicySet{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccTFEPolicySet_global,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.my_set", policySet),
					testAccCheckTFEPolicySetGlobal(policySet),
					resource.TestCheckResourceAttr("tfe_policy_set.my_set", "name", "global-set"),
					resource.TestCheckResourceAttr("tfe_policy_set.my_set", "description", "A global set populated with some policies"),
					resource.TestCheckResourceAttr("tfe_policy_set.my_set", "global", "true"),
					resource.TestCheckResourceAttr("tfe_policy_set.my_set", "policy_ids.#", "1"),
					resource.TestCheckResourceAttr("tfe_policy_set.my_set", "workspace_external_ids.#", "0"),
				),
			},

			resource.TestStep{
				Config: testAccTFEPolicySet_globalUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.my_set", policySet),
					testAccCheckTFEPolicySetGlobalUpdate(policySet),
					resource.TestCheckResourceAttr("tfe_policy_set.my_set", "name", "populated-set"),
					resource.TestCheckResourceAttr("tfe_policy_set.my_set", "description", "A set populated with some policies"),
					resource.TestCheckResourceAttr("tfe_policy_set.my_set", "global", "false"),
					resource.TestCheckResourceAttr("tfe_policy_set.my_set", "policy_ids.#", "1"),
					resource.TestCheckResourceAttr("tfe_policy_set.my_set", "workspace_external_ids.#", "1"),
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
			resource.TestStep{
				Config: testAccTFEPolicySet_basic,
			},

			resource.TestStep{
				ResourceName:      "tfe_policy_set.my_set",
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
		if policySet.Name != "my-set" {
			return fmt.Errorf("Bad name: %s", policySet.Name)
		}

		if policySet.Description != "A set for some of my Sentinel policies" {
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

		if policySet.Name != "populated-set" {
			return fmt.Errorf("Bad name: %s", policySet.Name)
		}

		if policySet.Description != "A set populated with some policies" {
			return fmt.Errorf("Bad description: %s", policySet.Description)
		}

		if policySet.Global {
			return fmt.Errorf("Bad value for global: %v", policySet.Global)
		}

		if len(policySet.Policies) != 1 {
			return fmt.Errorf("Wrong number of policies: %v", len(policySet.Policies))
		}

		policyID := policySet.Policies[0].ID
		policy, _ := tfeClient.Policies.Read(ctx, policyID)
		if policy.Name != "policy-one" {
			return fmt.Errorf("Wrong member policy: %v", policy.Name)
		}

		if len(policySet.Workspaces) != 1 {
			return fmt.Errorf("Wrong number of workspaces: %v", len(policySet.Workspaces))
		}

		workspaceID := policySet.Workspaces[0].ID
		workspace, _ := tfeClient.Workspaces.Read(ctx, "terraform-test", "workspace-one")
		if workspace.ID != workspaceID {
			return fmt.Errorf("Wrong member workspace: %v", workspace.Name)
		}

		return nil
	}
}

func testAccCheckTFEPolicySetNewMembers(policySet *tfe.PolicySet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		tfeClient := testAccProvider.Meta().(*tfe.Client)

		if policySet.Name != "populated-set" {
			return fmt.Errorf("Bad name: %s", policySet.Name)
		}

		if policySet.Description != "A set populated with some policies" {
			return fmt.Errorf("Bad description: %s", policySet.Description)
		}

		if policySet.Global {
			return fmt.Errorf("Bad value for global: %v", policySet.Global)
		}

		if len(policySet.Policies) != 1 {
			return fmt.Errorf("Wrong number of policies: %v", len(policySet.Policies))
		}

		policyID := policySet.Policies[0].ID
		policy, _ := tfeClient.Policies.Read(ctx, policyID)
		if policy.Name != "policy-two" {
			return fmt.Errorf("Wrong member policy: %v", policy.Name)
		}

		if len(policySet.Workspaces) != 1 {
			return fmt.Errorf("Wrong number of workspaces: %v", len(policySet.Workspaces))
		}

		workspaceID := policySet.Workspaces[0].ID
		workspace, _ := tfeClient.Workspaces.Read(ctx, "terraform-test", "workspace-two")
		if workspace.ID != workspaceID {
			return fmt.Errorf("Wrong member workspace: %v", workspace.Name)
		}

		return nil
	}
}

func testAccCheckTFEPolicySetGlobal(policySet *tfe.PolicySet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		tfeClient := testAccProvider.Meta().(*tfe.Client)

		if policySet.Name != "global-set" {
			return fmt.Errorf("Bad name: %s", policySet.Name)
		}

		if policySet.Description != "A global set populated with some policies" {
			return fmt.Errorf("Bad description: %s", policySet.Description)
		}

		if !policySet.Global {
			return fmt.Errorf("Bad value for global: %v", policySet.Global)
		}

		if len(policySet.Policies) != 1 {
			return fmt.Errorf("Wrong number of policies: %v", len(policySet.Policies))
		}

		policyID := policySet.Policies[0].ID
		policy, _ := tfeClient.Policies.Read(ctx, policyID)
		if policy.Name != "policy-one" {
			return fmt.Errorf("Wrong member policy: %v", policy.Name)
		}

		// Even though the terraform config should have 0 workspaces, the API will return
		// workspaces for global policy sets. This list would be the same as listing the
		// workspaces for the organization itself.
		if len(policySet.Workspaces) != 1 {
			return fmt.Errorf("Wrong number of workspaces: %v", len(policySet.Workspaces))
		}

		workspaceID := policySet.Workspaces[0].ID
		workspace, _ := tfeClient.Workspaces.Read(ctx, "terraform-test", "workspace-one")
		if workspace.ID != workspaceID {
			return fmt.Errorf("Wrong member workspace: %v", workspace.Name)
		}

		return nil
	}
}

func testAccCheckTFEPolicySetGlobalUpdate(policySet *tfe.PolicySet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		tfeClient := testAccProvider.Meta().(*tfe.Client)

		if policySet.Name != "populated-set" {
			return fmt.Errorf("Bad name: %s", policySet.Name)
		}

		if policySet.Description != "A set populated with some policies" {
			return fmt.Errorf("Bad description: %s", policySet.Description)
		}

		if policySet.Global {
			return fmt.Errorf("Bad value for global: %v", policySet.Global)
		}

		if len(policySet.Policies) != 1 {
			return fmt.Errorf("Wrong number of policies: %v", len(policySet.Policies))
		}

		policyID := policySet.Policies[0].ID
		policy, _ := tfeClient.Policies.Read(ctx, policyID)
		if policy.Name != "policy-one" {
			return fmt.Errorf("Wrong member policy: %v", policy.Name)
		}

		if len(policySet.Workspaces) != 1 {
			return fmt.Errorf("Wrong number of workspaces: %v", len(policySet.Workspaces))
		}

		workspaceID := policySet.Workspaces[0].ID
		workspace, _ := tfeClient.Workspaces.Read(ctx, "terraform-test", "workspace-two")
		if workspace.ID != workspaceID {
			return fmt.Errorf("Wrong member workspace: %v", workspace.Name)
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
  name = "terraform-test"
  email = "admin@company.com"
}

resource "tfe_policy_set" "my_set" {
  name = "my-set"
  description = "A set for some of my Sentinel policies"
  organization = "${tfe_organization.foobar.id}"
  policy_ids = []
}`

const testAccTFEPolicySet_populated = `
resource "tfe_organization" "foobar" {
  name = "terraform-test"
  email = "admin@company.com"
}

resource "tfe_sentinel_policy" "one" {
  name = "policy-one"
  policy = "main = rule { true }"
  organization = "${tfe_organization.foobar.id}"
}

resource "tfe_workspace" "one" {
  name = "workspace-one"
  organization = "${tfe_organization.foobar.id}"
}

resource "tfe_policy_set" "my_set" {
  name = "populated-set"
  description = "A set populated with some policies"
  organization = "${tfe_organization.foobar.id}"
  policy_ids = ["${tfe_sentinel_policy.one.id}"]
  workspace_external_ids = ["${tfe_workspace.one.external_id}"]
}`

const testAccTFEPolicySet_newMembers = `
resource "tfe_organization" "foobar" {
  name = "terraform-test"
  email = "admin@company.com"
}

resource "tfe_sentinel_policy" "one" {
  name = "policy-one"
  policy = "main = rule { true }"
  organization = "${tfe_organization.foobar.id}"
}

resource "tfe_sentinel_policy" "two" {
  name = "policy-two"
  policy = "main = rule { false }"
  organization = "${tfe_organization.foobar.id}"
}

resource "tfe_workspace" "two" {
  name = "workspace-two"
  organization = "${tfe_organization.foobar.id}"
}

resource "tfe_policy_set" "my_set" {
  name = "populated-set"
  description = "A set populated with some policies"
  organization = "${tfe_organization.foobar.id}"
  policy_ids = ["${tfe_sentinel_policy.two.id}"]
  workspace_external_ids = ["${tfe_workspace.two.external_id}"]
}`

const testAccTFEPolicySet_global = `
resource "tfe_organization" "foobar" {
  name = "terraform-test"
  email = "admin@company.com"
}

resource "tfe_sentinel_policy" "one" {
  name = "policy-one"
  policy = "main = rule { true }"
  organization = "${tfe_organization.foobar.id}"
}

resource "tfe_workspace" "one" {
  name = "workspace-one"
  organization = "${tfe_organization.foobar.id}"
}

resource "tfe_policy_set" "my_set" {
  name = "global-set"
  description = "A global set populated with some policies"
  organization = "${tfe_organization.foobar.id}"
  global = true
  policy_ids = ["${tfe_sentinel_policy.one.id}"]
}`

const testAccTFEPolicySet_globalUpdate = `
resource "tfe_organization" "foobar" {
  name = "terraform-test"
  email = "admin@company.com"
}

resource "tfe_sentinel_policy" "one" {
  name = "policy-one"
  policy = "main = rule { true }"
  organization = "${tfe_organization.foobar.id}"
}

resource "tfe_workspace" "one" {
  name = "workspace-one"
  organization = "${tfe_organization.foobar.id}"
}

resource "tfe_workspace" "two" {
  name = "workspace-two"
  organization = "${tfe_organization.foobar.id}"
}

resource "tfe_policy_set" "my_set" {
  name = "populated-set"
  description = "A set populated with some policies"
  organization = "${tfe_organization.foobar.id}"
  policy_ids = ["${tfe_sentinel_policy.one.id}"]
  workspace_external_ids = ["${tfe_workspace.two.external_id}"]
}`
