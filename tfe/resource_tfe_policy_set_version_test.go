package tfe

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"testing"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccTFEPolicySetVersion_basic(t *testing.T) {
	skipIfFreeOnly(t)

	policySet := &tfe.PolicySet{}
	policySetVersion := &tfe.PolicySetVersion{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	checksum, err := hashPolicies(testFixturePolicySetVersionFiles)
	if err != nil {
		t.Fatalf("Unable to generate checksum for policies %v", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySetVersion_basic(rInt, testFixturePolicySetVersionFiles),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetAttributes(policySet),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "name", "tst-terraform"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "description", "Policy Set"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "global", "false"),
					testAccCheckTFEPolicySetVersionExists("tfe_policy_set_version.foobar", policySetVersion),
					resource.TestCheckResourceAttr(
						"tfe_policy_set_version.foobar", "status", string(tfe.PolicySetVersionReady)),
					resource.TestCheckResourceAttr(
						"tfe_policy_set_version.foobar", "policies_path_contents_checksum", checksum),
				),
			},
		},
	})
}

func TestAccTFEPolicySetVersion_recreate(t *testing.T) {
	skipIfFreeOnly(t)

	policySet := &tfe.PolicySet{}
	policySetVersion := &tfe.PolicySetVersion{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	originalChecksum, err := hashPolicies(testFixturePolicySetVersionFiles)
	if err != nil {
		t.Fatalf("Unable to generate checksum for policies %v", err)
	}

	newFile := fmt.Sprintf("%s/newfile.test.sentinel", testFixturePolicySetVersionFiles)
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
				Config: testAccTFEPolicySetVersion_basic(rInt, testFixturePolicySetVersionFiles),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetAttributes(policySet),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "name", "tst-terraform"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "description", "Policy Set"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "global", "false"),
					testAccCheckTFEPolicySetVersionExists("tfe_policy_set_version.foobar", policySetVersion),
					resource.TestCheckResourceAttr(
						"tfe_policy_set_version.foobar", "status", string(tfe.PolicySetVersionReady)),
					resource.TestCheckResourceAttr(
						"tfe_policy_set_version.foobar", "policies_path_contents_checksum", originalChecksum),
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
				Config: testAccTFEPolicySetVersion_basic(rInt, testFixturePolicySetVersionFiles),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("tfe_policy_set.foobar", policySet),
					resource.TestCheckResourceAttr(
						"tfe_policy_set.foobar", "name", "tst-terraform"),
					testAccCheckTFEPolicySetVersionExists("tfe_policy_set_version.foobar", policySetVersion),
					resource.TestCheckResourceAttr(
						"tfe_policy_set_version.foobar", "status", string(tfe.PolicySetVersionReady)),
					testAccCheckTFEPolicySetVersionValidateChecksum("tfe_policy_set_version.foobar", testFixturePolicySetVersionFiles),
				),
			},
		},
	})
}

func testAccCheckTFEPolicySetVersionExists(n string, policySetVersion *tfe.PolicySetVersion) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		tfeClient := testAccProvider.Meta().(*tfe.Client)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		psv, err := tfeClient.PolicySetVersions.Read(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}

		if psv.ID != rs.Primary.ID {
			return fmt.Errorf("PolicySetVersion not found")
		}

		*policySetVersion = *psv

		return nil
	}
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

		if rs.Primary.Attributes["policies_path_contents_checksum"] != newChecksum {
			return fmt.Errorf("The new checksum for the policies contents did not match")
		}

		return nil
	}
}

func testAccTFEPolicySetVersion_basic(rInt int, sourcePath string) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_policy_set" "foobar" {
  name         = "tst-terraform"
  description  = "Policy Set"
  organization = tfe_organization.foobar.id
}

data "tfe_policy_set_version_files" "policy" {
  source_path = "%s"
}

resource "tfe_policy_set_version" "foobar" {
  policy_set_id = tfe_policy_set.foobar.id
  policies_path_contents_checksum = data.tfe_policy_set_version_files.policy.output_sha
  policies_path = data.tfe_policy_set_version_files.policy.source_path
}`, rInt, sourcePath)
}
