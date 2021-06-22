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
	expectedChecksum, err := hashPolicies(testFixturePolicySetVersionFiles)
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
					resource.TestCheckResourceAttr("data.tfe_policy_set_version_files.policy", "source_path", testFixturePolicySetVersionFiles),
					resource.TestCheckResourceAttr("data.tfe_policy_set_version_files.policy", "checksum", expectedChecksum),
				),
			},
		},
	})
}

func testAccTFEPolicySetVersionFilesConfig_basic(sourcePath string) string {
	return fmt.Sprintf(`
data "tfe_version_files" "policy" {
  source_path = "%s"
}
`, sourcePath)
}
