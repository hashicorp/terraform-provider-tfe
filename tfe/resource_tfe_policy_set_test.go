package tfe

import (
	"fmt"
	"testing"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
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
						"tfe_policy_set.foobar", "name", "terraform-test"),
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
						"tfe_policy_set.foobar", "name", "terraform-test"),
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
						"tfe_policy_set.foobar", "workspace_ids.#", "1"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "workspace_ids.294081462", "terraform-test/workspace-foo"),
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
						"tfe_policy_set.foobar", "workspace_ids.#", "1"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "workspace_ids.294081462", "terraform-test/workspace-foo"),
				),
			},

			{
				Config: testAccTFEPolicySet_updatePopulated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.foobar", policySet),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "name", "terraform-populated-updated"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "global", "false"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "policy_ids.#", "1"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "workspace_ids.#", "1"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "workspace_ids.1796703735", "terraform-test/workspace-bar"),
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
						"tfe_policy_set.foobar", "workspace_ids.#", "1"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "workspace_ids.294081462", "terraform-test/workspace-foo"),
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
						"tfe_policy_set.foobar", "workspace_ids.#", "1"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "workspace_ids.294081462", "terraform-test/workspace-foo"),
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
				Config: testAccTFEPolicySet_basic,
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
		if policySet.Name != "terraform-test" {
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

		policy, _ := tfeClient.Policies.Read(ctx, policySet.Policies[0].ID)
		if policy.Name != "policy-foo" {
			return fmt.Errorf("Wrong member policy: %v", policy.Name)
		}

		if len(policySet.Workspaces) != 1 {
			return fmt.Errorf("Wrong number of workspaces: %v", len(policySet.Workspaces))
		}

		workspace, _ := tfeClient.Workspaces.Read(ctx, "terraform-test", "workspace-foo")
		if workspace.ID != policySet.Workspaces[0].ID {
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

		policy, _ := tfeClient.Policies.Read(ctx, policySet.Policies[0].ID)
		if policy.Name != "policy-foo" {
			return fmt.Errorf("Wrong member policy: %v", policy.Name)
		}

		// Even though the terraform config should have 0 workspaces, the API will return
		// workspaces for global policy sets. This list would be the same as listing the
		// workspaces for the organization itself.
		if len(policySet.Workspaces) != 1 {
			return fmt.Errorf("Wrong number of workspaces: %v", len(policySet.Workspaces))
		}

		workspace, _ := tfeClient.Workspaces.Read(ctx, "terraform-test", "workspace-foo")
		if workspace.ID != policySet.Workspaces[0].ID {
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
  name  = "terraform-test"
  email = "admin@company.com"
}

resource "tfe_sentinel_policy" "foo" {
  name         = "policy-foo"
  policy       = "main = rule { true }"
  organization = "${tfe_organization.foobar.id}"
}

resource "tfe_policy_set" "foobar" {
  name         = "terraform-test"
  description  = "Policy Set"
  organization = "${tfe_organization.foobar.id}"
  policy_ids   = ["${tfe_sentinel_policy.foo.id}"]
}`

const testAccTFEPolicySet_populated = `
resource "tfe_organization" "foobar" {
  name  = "terraform-test"
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
  name  = "terraform-test"
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
  name  = "terraform-test"
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
