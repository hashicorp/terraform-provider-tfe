package tfe

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	testFixturePolicySetVersionFiles = "test-fixtures/policy-set-version"
)

func TestAccTFEPolicySetVersionFiles_basic(t *testing.T) {
	expectedHash, err := hashPolicies(testFixturePolicySetVersionFiles)
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySetVersionFilesConfig_basic(testFixturePolicySetVersionFiles),
				Check: resource.ComposeAggregateTestCheckFunc(
					// check data attrs
					resource.TestCheckResourceAttr("data.tfe_policy_set_version_files.policy", "source", testFixturePolicySetVersionFiles),
					resource.TestCheckResourceAttr("data.tfe_policy_set_version_files.policy", "output_sha", expectedHash),
				),
			},
		},
	})
}

func testAccTFEPolicySetVersionFilesConfig_basic(source string) string {
	return fmt.Sprintf(`
data "tfe_policy_set_version_files" "policy" {
  source = "%s"
}
`, source)
}
