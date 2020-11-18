package tfe

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

var testPolicySetDir = ""

func TestAccTFEPolicySetVersion_basic(t *testing.T) {
	policySet := &tfe.PolicySet{}
	version := &tfe.PolicySetVersion{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEPolicySetVersionDestroy,
		Steps: []resource.TestStep{
			{
				SkipFunc: testAccTFEPolicySetVersion_skipfunc,
				Config:   testAccTFEPolicySetVersion_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetVersionExists(
						"tfe_policy_set_version.foobar", version),
					testAccCheckTFEPolicySetVersionAttributes(policySet, version),
					resource.TestCheckResourceAttr(
						"tfe_policy_set_version.foobar", "version", "1"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set_version.foobar", "directory", testPolicySetDir),
				),
			},
		},
	})
}

func TestAccTFEPolicySetVersion_update(t *testing.T) {
	policySet := &tfe.PolicySet{}
	version := &tfe.PolicySetVersion{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEPolicySetVersionDestroy,
		Steps: []resource.TestStep{
			{
				SkipFunc: testAccTFEPolicySetVersion_skipfunc,
				Config:   testAccTFEPolicySetVersion_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetVersionExists(
						"tfe_policy_set_version.foobar", version),
					testAccCheckTFEPolicySetVersionAttributes(policySet, version),
					resource.TestCheckResourceAttr(
						"tfe_policy_set_version.foobar", "version", "1"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set_version.foobar", "directory", testPolicySetDir),
				),
			},

			{
				SkipFunc: testAccTFEPolicySetVersion_skipfunc,
				Config:   testAccTFEPolicySetVersion_update(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetVersionExists(
						"tfe_policy_set_version.foobar", version),
					testAccCheckTFEPolicySetVersionAttributesUpdate(policySet, version),
					resource.TestCheckResourceAttr(
						"tfe_policy_set_version.foobar", "version", "2"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set_version.foobar", "directory", testPolicySetDir),
				),
			},
		},
	})
}

func TestAccTFEPolicySetVersion_import(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEPolicySetVersionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySetVersion_import(rInt),
			},

			{
				ResourceName: "tfe_policy_set_version.foobar",
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					resources := s.RootModule().Resources
					policySet := resources["tfe_policy_set.foobar"]
					version := resources["tfe_policy_set_version.foobar"]

					return fmt.Sprintf("%s/%s", policySet.Primary.ID, version.Primary.ID), nil
				},
				ImportStateVerify: false,
			},
		},
	})
}

func testAccCheckTFEPolicySetVersionExists(
	n string, version *tfe.PolicySetVersion) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		tfeClient := testAccProvider.Meta().(*tfe.Client)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		v, err := tfeClient.PolicySetVersions.Read(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}

		*version = *v

		return nil
	}
}

func testAccCheckTFEPolicySetVersionAttributes(policySet *tfe.PolicySet,
	version *tfe.PolicySetVersion) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if version.Data.Relationships.PolicySet.Data.ID != policySet.ID {
			return fmt.Errorf("Policy set and policy set version IDs do not match: %s, %s",
				version.Data.Relationships.PolicySet.Data.ID, policySet.ID)
		}

		return nil
	}
}

func testAccCheckTFEPolicySetVersionAttributesUpdate(policySet *tfe.PolicySet,
	version *tfe.PolicySetVersion) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if version.Data.Relationships.PolicySet.Data.ID != policySet.ID {
			return fmt.Errorf("Policy set and policy set version IDs do not match: %s, %s",
				version.Data.Relationships.PolicySet.Data.ID, policySet.ID)
		}

		return nil
	}
}

func testAccCheckTFEPolicySetVersionDestroy(s *terraform.State) error {
	tfeClient := testAccProvider.Meta().(*tfe.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_policy_set_version" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := tfeClient.PolicySetVersions.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Policy set version %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccTFEPolicySetVersion_basic(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_policy_set" "foobar" {
  name         = "policy-set-test"
  description  = "a new policy set"
  organization = "${tfe_organization.foobar.id}"
}

resource "tfe_policy_set_version" "foobar" {
  version          = "1"
  directory        = "%s"
  policy_set_id    = "${tfe_policy_set.foobar.id}"
}`, rInt, testPolicySetDir)
}

func testAccTFEPolicySetVersion_update(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_policy_set" "foobar" {
  name         = "policy-set-test"
  description  = "a new policy set"
  organization = "${tfe_organization.foobar.id}"
}

resource "tfe_policy_set_version" "foobar" {
  version          = "2"
  directory        = "%s"
  policy_set_id    = "${tfe_policy_set.foobar.id}"
}`, rInt, testPolicySetDir)
}

func testAccTFEPolicySetVersion_import(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_policy_set" "foobar" {
  name         = "policy-set-test"
  description  = "a new policy set"
  organization = "${tfe_organization.foobar.id}"
}

resource "tfe_policy_set_version" "foobar" {
  version          = "1"
  directory        = "."
  policy_set_id    = "${tfe_policy_set.foobar.id}"
}`, rInt)
}

func testAccTFEPolicySetVersion_skipfunc() (bool, error) {
	path, err := os.Getwd()
	if err != nil {
		fmt.Errorf("Could not get workding directory")
		return false, err
	}
	testPolicySetDir = path + "/../test-fixtures/policy-set-version"
	//fmt.Println("policy set directory: ", testPolicySetDir)
	return true, err
}
